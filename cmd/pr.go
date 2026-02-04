package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/albuquerquesz/gitscribe/internal/agents"
	"github.com/albuquerquesz/gitscribe/internal/config"
	"github.com/albuquerquesz/gitscribe/internal/git"
	"github.com/albuquerquesz/gitscribe/internal/router"
	"github.com/albuquerquesz/gitscribe/internal/style"
	"github.com/spf13/cobra"
)

var (
	prTitle, prBody, prTarget string
	prDraft                   bool
)

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "Create a pull request",
	RunE: func(cmd *cobra.Command, args []string) error {
		return realizePR()
	},
}

func init() {
	prCmd.Flags().StringVarP(&prTitle, "title", "t", "", "Pull request title")
	prCmd.Flags().StringVarP(&prBody, "body", "b", "", "Pull request body")
	prCmd.Flags().StringVar(&prTarget, "target", "", "Target branch (default: main/master)")
	prCmd.Flags().BoolVar(&prDraft, "draft", false, "Create as draft PR")

	rootCmd.AddCommand(prCmd)
}

func hasCommitsBetweenBranches(current, target string) (bool, error) {
	cmd := exec.Command("git", "log", "--oneline", fmt.Sprintf("%s..%s", target, current))
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(string(output))) > 0, nil
}

func realizePR() error {
	style.GetASCIIName()

	if err := git.IsInsideWorkTree(); err != nil {
		style.Error(err.Error())
		return err
	}

	branch, err := git.GetCurrentBranch()
	if err != nil {
		style.Error(fmt.Sprintf("Failed to get current branch: %v", err))
		return err
	}

	if branch == "main" {
		style.Error("Cannot create PR from main branch")
		return fmt.Errorf("cannot create PR from main branch")
	}

	targetBranch := prTarget
	if targetBranch == "" {
		targetBranch = detectDefaultBranch()
	}

	hasCommits, err := hasCommitsBetweenBranches(branch, targetBranch)
	if err != nil {
		style.Error(fmt.Sprintf("Failed to check commits: %v", err))
		return err
	}
	if !hasCommits {
		style.Warning(fmt.Sprintf("No commits between '%s' and '%s'", targetBranch, branch))
		style.Info("Make sure you have pushed your branch and have commits to merge")
		return fmt.Errorf("no commits to merge")
	}

	remoteURL, err := git.GetRemoteURL()
	if err != nil {
		style.Error(fmt.Sprintf("Failed to get remote URL: %v", err))
		return err
	}

	provider := git.DetectProvider(remoteURL)
	if provider == "" {
		style.Error("Could not detect git provider (GitHub or GitLab)")
		return fmt.Errorf("could not detect git provider")
	}

	var cli string
	switch provider {
	case "github":
		cli = "gh"
	case "gitlab":
		cli = "glab"
	}

	if err := generatePR(provider); err != nil {
		return err
	}

	if !git.IsCLIInstalled(cli) {
		style.Error(fmt.Sprintf("%s CLI is not installed. Please install it first:", cli))

		switch cli {
		case "gh":
			fmt.Println("  https://cli.github.com/")
		case "glab":
			fmt.Println("  https://glab.readthedocs.io/")
		}

		return fmt.Errorf("%s CLI not installed", cli)
	}

	err = style.RunWithSpinner(fmt.Sprintf("Pushing branch '%s' to remote...", branch), func() error {
		return git.Push(branch)
	})
	if err != nil {
		style.Error(fmt.Sprintf("Failed to push branch: %v", err))
		return err
	}
	style.Success("Branch pushed successfully!")

	// Verify branch exists on remote
	verifyCmd := exec.Command("git", "ls-remote", "--heads", "origin", branch)
	if _, err := verifyCmd.Output(); err != nil {
		style.Warning("Branch may not be available on remote yet")
	}

	if prTitle == "" || prBody == "" {
		style.Info("Generating PR title and body with AI...")
	}

	if prTitle == "" {
		style.Error("PR title cannot be empty")
		return fmt.Errorf("PR title is required")
	}

	// Debug: Print values before creating PR
	fmt.Printf("Debug: branch='%s', targetBranch='%s', prTitle='%s', prBody='%s'\n", branch, targetBranch, prTitle, prBody)

	if branch == "" {
		return fmt.Errorf("branch name is empty - cannot create PR")
	}

	style.Info(fmt.Sprintf("Creating %s PR from '%s' to '%s'...", provider, branch, targetBranch))

	// Get git working directory
	gitDirCmd := exec.Command("git", "rev-parse", "--show-toplevel")
	gitDirOutput, err := gitDirCmd.Output()
	if err != nil {
		style.Error("Failed to get git directory")
		return err
	}
	workDir := strings.TrimSpace(string(gitDirOutput))

	var createCmd *exec.Cmd

	switch provider {
	case "github":
		args := []string{"pr", "create", "--title", prTitle, "--body", prBody, "--base", targetBranch, "--head", branch}
		fmt.Printf("Debug: gh args = %v\n", args)
		if prDraft {
			args = append(args, "--draft")
		}
		createCmd = exec.Command("gh", args...)
		createCmd.Dir = workDir

	case "gitlab":
		args := []string{"mr", "create", "--title", prTitle, "--description", prBody, "--target-branch", targetBranch, "--source-branch", branch}
		if prDraft {
			args = append(args, "--draft")
		}
		createCmd = exec.Command("glab", args...)
		createCmd.Dir = workDir
	}

	createCmd.Stdout = os.Stdout
	createCmd.Stderr = os.Stderr

	err = style.RunWithSpinner("Creating pull request...", func() error {
		return createCmd.Run()
	})
	if err != nil {
		style.Error(fmt.Sprintf("Failed to create PR: %v", err))
		style.Info("Troubleshooting tips:")
		style.Info("  1. Ensure you've pushed your branch: git push origin " + branch)
		style.Info("  2. Check that you have commits to merge")
		style.Info("  3. Verify you have permission to create PRs in this repository")
		style.Info("  4. Try running manually: gh pr create --title \"...\" --body \"...\"")
		return err
	}

	style.Success(fmt.Sprintf("PR created successfully on %s!", provider))
	return nil
}

func generatePR(provider string) error {
	commits, err := getCommitLog()
	if err != nil {
		style.Error(fmt.Sprintf("Failed to get commit log: %v", err))
		return err
	}

	if len(commits) == 0 {
		style.Warning("No commits found to generate PR description")
		return nil
	}
	var generatedContent string
	err = style.RunWithSpinner("Generating PR description...", func() error {
		var err error
		generatedContent, err = generatePRContent(commits, provider)
		return err
	})
	if err != nil {
		style.Error(fmt.Sprintf("Failed to generate PR content: %v", err))
		return err
	}

	style.Success("PR content generated!")

	action, finalContent := style.ShowCommitPrompt(generatedContent)
	if action == "cancel" {
		style.Warning("PR creation cancelled")
		return nil
	}

	lines := strings.SplitN(finalContent, "\n", 2)
	if prTitle == "" && len(lines) > 0 {
		prTitle = strings.TrimSpace(lines[0])
	}
	if prBody == "" && len(lines) > 1 {
		prBody = strings.TrimSpace(lines[1])
	}

	return nil
}

func getCommitLog() (string, error) {
	target := detectDefaultBranch()
	cmd := exec.Command("git", "log", "--oneline", "-20", fmt.Sprintf("%s..HEAD", target))
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func detectDefaultBranch() string {
	for _, branch := range []string{"main", "master"} {
		cmd := exec.Command("git", "rev-parse", "--verify", branch)
		if err := cmd.Run(); err == nil {
			return branch
		}
	}
	return "main"
}

func generatePRContent(commits, provider string) (string, error) {
	cfg, err := config.Load()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	agent, err := cfg.GetDefaultAgent()
	if err != nil {
		return "", fmt.Errorf("no suitable agent found: %w", err)
	}

	r := router.NewRouter(cfg)

	prompt := fmt.Sprintf(
		"Generate a pull request title and body based on the following git commits. "+
			"The response should have the title on the first line, followed by a blank line, then the body. "+
			"The body should describe what changes were made and why. "+
			"For %s, use markdown formatting in the body. "+
			"Here are the commits:\n\n%s",
		provider, commits,
	)

	messages := []agents.Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	options := agents.RequestOptions{
		Temperature: 0.7,
	}

	resp, err := r.RouteRequest(context.Background(), agent.Name, messages, options)
	if err != nil {
		return "", fmt.Errorf("ai request failed: %w", err)
	}

	return resp.Content, nil
}
