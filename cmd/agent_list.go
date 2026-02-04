package cmd

import (
	"fmt"
	"strings"

	"github.com/albuquerquesz/gitscribe/internal/config"
	"github.com/albuquerquesz/gitscribe/internal/secrets"
	"github.com/spf13/cobra"
)

var agentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured agentss",
	RunE: func(cmd *cobra.Command, args []string) error {
		return listAllAgents()
	},
}

func init() {
	agentCmd.AddCommand(agentListCmd)
}

func listAllAgents() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("🤖 Configured AI Agents")
	fmt.Println(strings.Repeat("─", 50))

	keyMgr := secrets.NewAgentKeyManager()

	for _, agent := range cfg.Agents {
		defaultMarker := " "
		if cfg.Global.DefaultAgent == agent.Name {
			defaultMarker = "★"
		}

		statusIcon := "🔴"
		if agent.Enabled {
			statusIcon = "🟢"
		}

		keyStatus := "❌"
		if keyMgr.KeyExists(keyMgr.GetAgentKeyName(agent.Name)) {
			keyStatus = "✅"
		}

		fmt.Printf("%s %s %s\n", defaultMarker, statusIcon, agent.Name)
		fmt.Printf("   Provider: %s\n", agent.Provider)
		fmt.Printf("   Model: %s\n", agent.Model)
		fmt.Printf("   Priority: %d\n", agent.Priority)
		fmt.Printf("   API Key: %s\n", keyStatus)

		if agent.BaseURL != "" {
			fmt.Printf("   Base URL: %s\n", agent.BaseURL)
		}

		fmt.Println()
	}

	if len(cfg.Agents) == 0 {
		fmt.Println("No agents configured. Use 'gs agent add' to create one.")
	}

	return nil
}
