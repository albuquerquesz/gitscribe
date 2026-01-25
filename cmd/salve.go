package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/albqvictor1508/gitscribe/internal/style"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var version = "1.0.2"

var msg, branch string

var commitCmd = &cobra.Command{
	Use:   "cmt [files]",
	Args:  cobra.MinimumNArgs(0),
	Short: "AI-powered git add, commit, and push",
	Run: func(cmd *cobra.Command, args []string) error {
		return commit(args)
	},
}

func init() {
	commitCmd.Flags().StringP("message", "m", &msg, "The commit message")
	commitCmd.Flags().StringP("branch", "b", &branch, "The branch to push to")

	rootCmd.AddCommand(commitCmd)
}

func commit(files []string) error {
	style.GetASCIIName()
	ShowUpdate(version)

	if len(files) == 0 {
		files = append(files, ".")
	}

	addSpinner := style.Spinner("Staging files...")
	for _, file := range files {
		addCmd := exec.Command("git", "add", file)
		addCmd.Stdout = io.Discard
		addCmd.Stderr = io.Discard
		if err := addCmd.Run(); err != nil {
			addSpinner.Fail(fmt.Sprintf("Failed to stage file %s: %v", file, err))
			os.Exit(1)
		}
	}
	addSpinner.Success("Files staged successfully!")

	if len(msg) == 0 {
		style.Spinner("Analyzing changes and generating message with AI...")

		var diffOutput bytes.Buffer
		diffCmd := exec.Command("git", "diff", "--staged")
		diffCmd.Stdout = &diffOutput
		diffCmd.Stderr = &diffOutput

		if err := diffCmd.Run(); err != nil {
			aiSpinner.Fail(fmt.Sprintf("Failed to get git diff: %s", diffOutput.String()))
			os.Exit(1)
		}

		if diffOutput.Len() == 0 {
			aiSpinner.Warning("No changes found in stage. Nothing to commit.")
			os.Exit(0)
		}

		context := fmt.Sprintf(
			"Analyze the following git diff and generate a commit message. "+
				"The message must follow the Conventional Commits standard. "+
				"Your response should contain *only* the commit message, without any additional text, explanations, or markdown formatting. "+
				"Focus on the primary purpose of the changes and be concise. "+
				"Do not include file names, line numbers, or the diff itself in the output. "+
				"Here is the diff:\n%v",
			diffOutput.String(),
		)

		result, err := internal.SendPrompt(context)
		if err != nil {
			aiSpinner.Fail(fmt.Sprintf("Error generating message with AI: %v", err))
			os.Exit(1)
		}
		message = result
		aiSpinner.Success("Commit message generated!")
	}

	if !style.ConfirmAction(message) {
		os.Exit(1)
	}

	commitSpinner, _ := pterm.DefaultSpinner.WithSequence("|", "/", "-", "\\ ").Start()
	commitSpinner.UpdateText("Committing...")

	var commitOutput bytes.Buffer
	commitCmd := exec.Command("git", "commit", "-m", message)
	commitCmd.Stdout = &commitOutput
	commitCmd.Stderr = &commitOutput
	if err := commitCmd.Run(); err != nil {
		commitSpinner.Fail(fmt.Sprintf("Error while committing: %s", commitOutput.String()))
		os.Exit(1)
	}
	commitSpinner.Success("Commit successful!")

	pushSpinner, _ := pterm.DefaultSpinner.WithSequence("|", "/", "-", "\\").Start()
	pushSpinner.UpdateText(fmt.Sprintf("pushing files into %s", branch))

	var pushOutput bytes.Buffer
	pushCmd := exec.Command("git", "push", "origin", branch)
	pushCmd.Stdout = &pushOutput
	pushCmd.Stderr = &pushOutput
	if err := pushCmd.Run(); err != nil {
		pushSpinner.Fail(fmt.Sprintf("Error while pushing: %s", pushOutput.String()))
		os.Exit(1)
	}
	pterm.Success.Println("All done!")
}
