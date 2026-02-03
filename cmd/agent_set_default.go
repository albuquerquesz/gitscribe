package cmd

import (
	"fmt"

	"github.com/albuquerquesz/gitscribe/internal/config"
	"github.com/spf13/cobra"
)

var defaultAgentName string

var agentSetDefaultCmd = &cobra.Command{
	Use:   "set-default",
	Short: "Set the default agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		return setDefaultAgent(defaultAgentName)
	},
}

func init() {
	agentSetDefaultCmd.Flags().StringVarP(&defaultAgentName, "name", "n", "", "Agent name to set as default (required)")
	agentSetDefaultCmd.MarkFlagRequired("name")

	agentCmd.AddCommand(agentSetDefaultCmd)
}

func setDefaultAgent(name string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.SetDefaultAgent(name); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Default agent set to '%s'\n", name)
	return nil
}
