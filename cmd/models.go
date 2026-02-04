package cmd

import (
	"fmt"

	"github.com/albuquerquesz/gitscribe/internal/auth"
	"github.com/albuquerquesz/gitscribe/internal/catalog"
	appconfig "github.com/albuquerquesz/gitscribe/internal/config"
	"github.com/albuquerquesz/gitscribe/internal/secrets"
	"github.com/albuquerquesz/gitscribe/internal/style"
	"github.com/spf13/cobra"
)

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Browse and enable AI models",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runModelsInteractive()
	},
}

func init() {
	rootCmd.AddCommand(modelsCmd)
}

func runModelsInteractive() error {
	manager := catalog.NewCatalogManager(auth.LoadAPIKey)

	fmt.Println(style.TitleStyle.Render("\n AI Model Catalog"))

	selected, err := style.SelectModel(manager)
	if err != nil || selected == nil {
		return nil
	}

	return handleModelSelection(*selected, manager)
}

func handleModelSelection(m catalog.Model, manager *catalog.CatalogManager) error {
	apiKey, err := auth.LoadAPIKey(m.Provider)
	isAuthenticated := err == nil && apiKey != ""

	if !isAuthenticated {
		style.Info(fmt.Sprintf("Model %s from %s requires an API key.", m.Name, m.Provider))
		authProvider = m.Provider
		if err := runSetKey(); err != nil {
			return err
		}

		apiKey, err = auth.LoadAPIKey(m.Provider)
		if err != nil || apiKey == "" {
			return fmt.Errorf("API key was not stored correctly")
		}
	} else {
		style.Success(fmt.Sprintf("API key already configured for %s.", m.Provider))
	}

	cfg, err := appconfig.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	profileName := fmt.Sprintf("%s-%s", m.Provider, m.ID)
	if len(m.ID) > len(m.Provider) && m.ID[:len(m.Provider)] == m.Provider {
		profileName = m.ID
	}

	keyMgr := secrets.NewAgentKeyManager()
	keyringKey := keyMgr.GetAgentKeyName(profileName)

	pConfig, _ := manager.GetProviderConfig(m.Provider)

	existing, err := cfg.GetAgentByName(profileName)
	if err == nil {
		existing.Enabled = true
		existing.Model = m.ID
		existing.BaseURL = pConfig.BaseURL
		existing.KeyringKey = keyringKey
	} else {
		newAgent := appconfig.AgentProfile{
			Name:        profileName,
			Provider:    appconfig.AgentProvider(m.Provider),
			Model:       m.ID,
			BaseURL:     pConfig.BaseURL,
			Enabled:     true,
			Priority:    1,
			Temperature: 0.7,
			MaxTokens:   4096,
			Timeout:     30,
			KeyringKey:  keyringKey,
		}
		if err := cfg.AddAgent(newAgent); err != nil {
			return fmt.Errorf("failed to add agent: %w", err)
		}
	}

	if err := keyMgr.StoreAgentKey(profileName, apiKey); err != nil {
		style.Warning(fmt.Sprintf("Failed to link API key to agent: %v", err))
	}

	if err := cfg.SetDefaultAgent(profileName); err != nil {
		return fmt.Errorf("failed to set default agent: %w", err)
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	style.Success(fmt.Sprintf("Agent %s is now active and set as default!", profileName))
	return nil
}

func getCatalogManager() (*catalog.CatalogManager, error) {
	return catalog.NewCatalogManager(auth.LoadAPIKey), nil
}
