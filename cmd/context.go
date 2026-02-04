package cmd

import (
	"github.com/spf13/cobra"
)

var contextCmd = &cobra.Command{
	Use:     "context",
	Short:   "",
	Aliases: []string{"ctx"},
}

func init() {
	rootCmd.AddCommand(contextCmd)
}
