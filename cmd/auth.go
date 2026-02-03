package cmd

import (
	"fmt"

	appconfig "github.com/albuquerquesz/gitscribe/internal/config"
	"github.com/albuquerquesz/gitscribe/internal/secrets"
	"github.com/spf13/cobra"
)

var authProvider string

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication keys",
}

func init() {
	rootCmd.AddCommand(authCmd)
}

func updateAgentProfile(providerName, apiKey string) error {
	cfg, err := appconfig.Load()
	if err != nil {
		return err
	}

	keyringKey := fmt.Sprintf("%s-api-key", providerName)

	for i := range cfg.Agents {
		if string(cfg.Agents[i].Provider) == providerName {
			cfg.Agents[i].KeyringKey = keyringKey
			cfg.Agents[i].Enabled = true

			keyMgr := secrets.NewAgentKeyManager()
			if err := keyMgr.StoreAgentKey(cfg.Agents[i].Name, apiKey); err != nil {
				return fmt.Errorf("failed to store API key for agent: %w", err)
			}
		}
	}

	return cfg.Save()
}
