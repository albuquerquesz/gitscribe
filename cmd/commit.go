package cmd

import (
	"fmt"

	"github.com/albuquerquesz/gitscribe/internal/ai"
	"github.com/albuquerquesz/gitscribe/internal/git"
	"github.com/albuquerquesz/gitscribe/internal/style"
	"github.com/albuquerquesz/gitscribe/internal/version"
	"github.com/spf13/cobra"
)

var msg, branch, commitAgent string

var commitCmd = &cobra.Command{
	Use:     "commit [files]",
	Aliases: []string{"cmt"},
	Args:    cobra.MinimumNArgs(0),
	Short:   "AI-powered git add, commit, and push",
	RunE: func(cmd *cobra.Command, args []string) error {
		return commit(args)
	},
}

func init() {
	commitCmd.Flags().StringVarP(&msg, "message", "m", "", "The commit message")
	commitCmd.Flags().StringVarP(&branch, "branch", "b", "", "The branch to push to")
	commitCmd.Flags().StringVarP(&commitAgent, "agent", "a", "", "The AI agent to use (overrides default)")

	rootCmd.AddCommand(commitCmd)
}

func commit(files []string) error {
	style.GetASCIIName()
	version.ShowUpdate(v)

	if len(files) == 0 {
		files = append(files, ".")
	}

	if err := git.StageFiles(files); err != nil {
		style.Error(err.Error())
		return err
	}
	style.Success("Files staged successfully!")

	if len(msg) == 0 {
		diff, err := git.GetStagedDiff()
		if err != nil {
			style.Error(err.Error())
			return err
		}

		if len(diff) == 0 {
			style.Warning("No changes found in stage. Nothing to commit.")
			return nil
		}

		var result string
		err = style.RunWithSpinner("Generating commit message...", func() error {
			var err error
			result, err = ai.SendPrompt(diff, commitAgent)
			return err
		})
		if err != nil {
			style.Error(fmt.Sprintf("Error generating message with AI: %v", err))
			return err
		}
		style.Success("Message generated!")
		msg = result
	}

	action, finalMsg := style.ShowCommitPrompt(msg)
	if action == "cancel" {
		fmt.Println()
		fmt.Println("Commit cancelled")
		return nil
	}
	msg = finalMsg

	if err := git.Commit(msg); err != nil {
		return err
	}
	style.Success("Commit successful!")

	targetBranch := branch
	if targetBranch == "" {
		current, err := git.GetCurrentBranch()
		if err != nil {
			style.Warning("Could not determine current branch. Please specify branch with -b.")
			return err
		}
		targetBranch = current
	}

	err := style.RunWithSpinner("Pushing to remote...", func() error {
		return git.Push(targetBranch)
	})
	if err != nil {
		return err
	}
	style.Success("All done!")

	return nil
}
