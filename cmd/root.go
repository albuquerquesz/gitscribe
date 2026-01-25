package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var version string = "dev"

var rootCmd = &cobra.Command{
	Use:     "gs",
	Version: version,
	Short:   "GitScribe: AI-powered commit messages",
	Long: `GitScribe (gs) helps you generate meaningful commit messages
using AI (Groq/Llama) and manages your workflow from staging to pushing.`,
}

func Exec() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
