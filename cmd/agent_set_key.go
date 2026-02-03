package cmd

import (
	"fmt"

	"github.com/albuquerquesz/gitscribe/internal/config"
	"github.com/albuquerquesz/gitscribe/internal/secrets"
	"github.com/albuquerquesz/gitscribe/internal/style"
	"github.com/spf13/cobra"
)

var agentSetKeyCmd = &cobra.Command{
	Use:   "set-key [name]",
	Short: "Set or update API key for an agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return setAgentKey(args[0])
	},
}

func init() {
	agentCmd.AddCommand(agentSetKeyCmd)
}

func setAgentKey(name string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	agent, err := cfg.GetAgentByName(name)
	if err != nil {
		return err
	}

	prompt := fmt.Sprintf("Enter new API key for %s (%s):", name, agent.Provider)
	newKey, err := style.Prompt(prompt)
	if err != nil {
		return err
	}

	if newKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	keyMgr := secrets.NewAgentKeyManager()
	if err := keyMgr.StoreAgentKey(name, newKey); err != nil {
		return fmt.Errorf("failed to store API key: %w", err)
	}

	fmt.Printf("✅ API key updated for agent '%s'\n", name)
	return nil
}
