package cmd

import (
	"github.com/spf13/cobra"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage AI agents and their API keys",
	Long:  "Configure, add, remove, and manage AI agent profiles and API keys",
}

func init() {
	rootCmd.AddCommand(agentCmd)
}
