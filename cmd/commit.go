package cmd

import (
	"fmt"

	"github.com/albuquerquesz/gitscribe/internal/ai"
	"github.com/albuquerquesz/gitscribe/internal/git"
	"github.com/albuquerquesz/gitscribe/internal/style"
	"github.com/albuquerquesz/gitscribe/internal/version"
	"github.com/spf13/cobra"
)

var msg, branch string

var commitCmd = &cobra.Command{
	Use:   "cmt [files]",
	Args:  cobra.MinimumNArgs(0),
	Short: "AI-powered git add, commit, and push",
	RunE: func(cmd *cobra.Command, args []string) error {
		return commit(args)
	},
}

func init() {
	commitCmd.Flags().StringVarP(&msg, "message", "m", "", "The commit message")
	commitCmd.Flags().StringVarP(&branch, "branch", "b", "", "The branch to push to")

	rootCmd.AddCommand(commitCmd)
}

func commit(files []string) error {
	style.GetASCIIName()
	version.ShowUpdate(v)

	if len(files) == 0 {
		files = append(files, ".")
	}

	addSpinner := style.Spinner("Staging files...")
	if err := git.StageFiles(files); err != nil {
		addSpinner.Fail(err.Error())
		return err
	}
	addSpinner.Success("Files staged successfully!")

	if len(msg) == 0 {
		aiSpinner := style.Spinner("Analyzing changes and generating message with AI...")

		diff, err := git.GetStagedDiff()
		if err != nil {
			aiSpinner.Fail(err.Error())
			return err
		}

		if len(diff) == 0 {
			aiSpinner.Warning("No changes found in stage. Nothing to commit.")
			return nil
		}

		result, err := ai.SendPrompt(diff)
		if err != nil {
			aiSpinner.Fail(fmt.Sprintf("Error generating message with AI: %v", err))
			return err
		}
		msg = result
		aiSpinner.Success("Commit message generated!")
	}

	if !style.ConfirmAction(msg) {
		return fmt.Errorf("commit cancelled")
	}

	commitSpinner := style.Spinner("Committing...")
	if err := git.Commit(msg); err != nil {
		commitSpinner.Fail(err.Error())
		return err
	}
	commitSpinner.Success("Commit successful!")

	targetBranch := branch
	if targetBranch == "" {
		current, err := git.GetCurrentBranch()
		if err != nil {
			commitSpinner.Warning("Could not determine current branch. Please specify branch with -b.")
			return err
		}
		targetBranch = current
	}

	pushSpinner := style.Spinner(fmt.Sprintf("Pushing files into %s...", targetBranch))
	if err := git.Push(targetBranch); err != nil {
		pushSpinner.Fail(err.Error())
		return err
	}
	pushSpinner.Success("All done!")

	return nil
}