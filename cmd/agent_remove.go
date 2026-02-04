package cmd

import (
	"fmt"
	"strings"

	"github.com/albuquerquesz/gitscribe/internal/config"
	"github.com/albuquerquesz/gitscribe/internal/secrets"
	"github.com/albuquerquesz/gitscribe/internal/style"
	"github.com/spf13/cobra"
)

var agentRemoveCmd = &cobra.Command{
	Use:     "remove [name]",
	Aliases: []string{"rm"},
	Short:   "Remove an agent profile",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return removeAgent(args[0])
	},
}

func init() {
	agentCmd.AddCommand(agentRemoveCmd)
}

func removeAgent(name string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	confirm, err := style.Prompt(fmt.Sprintf("Are you sure you want to remove agent '%s'? (yes/no): ", name))
	if err != nil {
		return err
	}

	if strings.ToLower(confirm) != "yes" {
		fmt.Println("Cancelled.")
		return nil
	}

	if err := cfg.RemoveAgent(name); err != nil {
		return err
	}

	keyMgr := secrets.NewAgentKeyManager()
	if err := keyMgr.DeleteAgentKey(name); err != nil {
		fmt.Printf("Warning: could not remove API key: %v\n", err)
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Agent '%s' removed successfully!\n", name)
	return nil
}
