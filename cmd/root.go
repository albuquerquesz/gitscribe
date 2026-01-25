package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// THIS I`LL BE REPLACED WHEN CLI BUILD
var v string = "v1.0.0"

var rootCmd = &cobra.Command{
	Use:     "gs",
	Version: v,
	Short:   "GitScribe: AI-powered commit messages",
	Long: `GitScribe (gs) helps you generate meaningful commit messages
using AI (Groq/Llama) and manages your workflow from staging to pushing.`,
}

func init() {
	rootCmd.SetVersionTemplate("GitScribe {{.Version}}\n")
}

func Exec() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
