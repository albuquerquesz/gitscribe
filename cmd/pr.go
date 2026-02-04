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

	style.Info(fmt.Sprintf("Pushing branch '%s' to remote...", branch))
	if err := git.Push(branch); err != nil {
		style.Error(fmt.Sprintf("Failed to push branch: %v", err))
		return err
	}
	style.Success("Branch pushed successfully!")

	if prTitle == "" || prBody == "" {
		style.Info("Generating PR title and body with AI...")
	}

	targetBranch := prTarget
	if targetBranch == "" {
		targetBranch = detectDefaultBranch()
	}

	style.Info(fmt.Sprintf("Creating %s PR from '%s' to '%s'...", provider, branch, targetBranch))

	var createCmd *exec.Cmd

	switch provider {
	case "github":
		args := []string{"pr", "create", "--title", prTitle, "--body", prBody, "--base", targetBranch}
		if prDraft {
			args = append(args, "--draft")
		}
		createCmd = exec.Command("gh", args...)

	case "gitlab":
		args := []string{"mr", "create", "--title", prTitle, "--description", prBody, "--target-branch", targetBranch}
		if prDraft {
			args = append(args, "--draft")
		}
		createCmd = exec.Command("glab", args...)
	}

	createCmd.Stdout = os.Stdout
	createCmd.Stderr = os.Stderr

	if err := createCmd.Run(); err != nil {
		style.Error(fmt.Sprintf("Failed to create PR: %v", err))
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
	generatedContent, err := generatePRContent(commits, provider)
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
