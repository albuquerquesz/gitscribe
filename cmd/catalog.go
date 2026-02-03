package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/albuquerquesz/gitscribe/internal/auth"
	"github.com/albuquerquesz/gitscribe/internal/catalog"
	appconfig "github.com/albuquerquesz/gitscribe/internal/config"
	"github.com/albuquerquesz/gitscribe/internal/secrets"
	"github.com/albuquerquesz/gitscribe/internal/style"
	"github.com/spf13/cobra"
)

var catalogCmd = &cobra.Command{
	Use:   "catalog",
	Short: "Manage AI model catalog",
	Long: `Browse and manage the AI model catalog.

The catalog contains model information for all supported providers:
- Anthropic (Claude)
- OpenAI (GPT)
- Groq (Llama, etc.)
- OpenRouter (aggregated models)
- Ollama (local models)

Models can be viewed from the static catalog or fetched dynamically
from provider APIs when available.`,
}

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Browse and enable AI models interactively",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runModelsInteractive()
	},
}

var catalogListCmd = &cobra.Command{
	Use:   "list [provider]",
	Short: "List available models",
	Long:  "List all models or filter by provider. Shows model ID, name, pricing, and capabilities.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := getCatalogManager()
		if err != nil {
			return err
		}

		showAll, _ := cmd.Flags().GetBool("all")
		showDetails, _ := cmd.Flags().GetBool("details")
		filterTier, _ := cmd.Flags().GetString("tier")
		filterCapability, _ := cmd.Flags().GetString("capability")

		if len(args) > 0 {
			provider := args[0]
			models, err := manager.GetModelsByProvider(provider)
			if err != nil {
				return fmt.Errorf("failed to get models for %s: %w", provider, err)
			}

			fmt.Printf("Models for %s:\n\n", provider)
			printModels(models, showDetails)
		} else {
			var models []catalog.Model

			if filterTier != "" || filterCapability != "" {
				opts := catalog.FilterOptions{}
				if filterTier != "" {
					opts.PricingTier = catalog.PricingTier(filterTier)
				}
				if filterCapability != "" {
					opts.Capability = catalog.Capability(filterCapability)
				}
				models = manager.FilterModels(opts)
				fmt.Printf("Filtered models (%d found):\n\n", len(models))
			} else if showAll {
				providers := manager.ListProviders()
				for _, p := range providers {
					pmodels, _ := manager.GetModelsByProvider(p)
					models = append(models, pmodels...)
				}
				fmt.Printf("All models (%d found):\n\n", len(models))
			} else {
				providers := manager.ListProviders()
				fmt.Println("Available providers:")
				fmt.Println()
				for _, p := range providers {
					config, _ := manager.GetProviderConfig(p)
					fetchStatus := "static only"
					if config.SupportsList {
						fetchStatus = "supports dynamic fetch"
					}
					fmt.Printf("  - %s (%s)\n", p, fetchStatus)
				}
				fmt.Println()
				fmt.Println("Use 'catalog list <provider>' to see models for a specific provider.")
				fmt.Println("Use 'catalog list --all' to see all models.")
				return nil
			}

			printModels(models, showDetails)
		}

		return nil
	},
}

var catalogShowCmd = &cobra.Command{
	Use:   "show <model-id>",
	Short: "Show detailed information about a model",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := getCatalogManager()
		if err != nil {
			return err
		}

		modelID := args[0]
		model, err := manager.GetModel(modelID)
		if err != nil {
			return fmt.Errorf("model not found: %s", modelID)
		}

		printModelDetails(model)
		return nil
	},
}

var catalogRefreshCmd = &cobra.Command{
	Use:   "refresh [provider]",
	Short: "Refresh model list from provider APIs",
	RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := getCatalogManager()
		if err != nil {
			return err
		}

		force, _ := cmd.Flags().GetBool("force")
		ctx := context.Background()

		if len(args) > 0 {
			provider := args[0]
			fmt.Printf("Refreshing %s...\n", provider)

			var err error
			if force {
				err = manager.ForceRefresh(ctx, provider)
			} else {
				err = manager.RefreshProvider(ctx, provider)
			}

			if err != nil {
				return fmt.Errorf("failed to refresh %s: %w", provider, err)
			}
			fmt.Printf("✓ %s refreshed successfully\n", provider)
		} else {
			fmt.Println("Refreshing all providers...")
			if err := manager.RefreshAll(ctx); err != nil {
				return fmt.Errorf("failed to refresh: %w", err)
			}
			fmt.Println("✓ All providers refreshed")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(catalogCmd)
	rootCmd.AddCommand(modelsCmd)

	catalogCmd.AddCommand(catalogListCmd)
	catalogCmd.AddCommand(catalogShowCmd)
	catalogCmd.AddCommand(catalogRefreshCmd)

	catalogListCmd.Flags().BoolP("all", "a", false, "Show all models from all providers")
	catalogListCmd.Flags().BoolP("details", "d", false, "Show detailed information")
}

func runModelsInteractive() error {
	manager, err := getCatalogManager()
	if err != nil {
		return err
	}

	fmt.Println(style.TitleStyle.Render("\n AI Model Catalog"))

	selected, err := style.SelectModel(manager)
	if err != nil {
		return nil
	}

	if selected == nil {
		return nil
	}

	return handleModelSelection(*selected, manager)
}

func handleModelSelection(m catalog.Model, manager *catalog.CatalogManager) error {
	apiKey, err := auth.LoadAPIKey(m.Provider)
	isAuthenticated := err == nil && apiKey != ""

	if !isAuthenticated {
		style.Info(fmt.Sprintf("Model %s from %s requires authentication.", m.Name, m.Provider))

		if m.Provider == "openai" || m.Provider == "anthropic" {
			if !style.ConfirmAction(fmt.Sprintf("Do you want to log in to %s via browser?", m.Provider)) {
				return nil
			}
			authProvider = m.Provider
			if err := runAuth(); err != nil {
				return err
			}
		} else {
			style.Info(fmt.Sprintf("Please provide an API key for %s", m.Provider))
			authProvider = m.Provider
			if err := runSetKey(); err != nil {
				return err
			}
		}
	} else {
		style.Success(fmt.Sprintf("Authentication found for %s.", m.Provider))
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

	existing, err := cfg.GetAgentByName(profileName)
	if err == nil {
		existing.Enabled = true
		existing.Model = m.ID
		existing.KeyringKey = keyringKey
		style.Info(fmt.Sprintf("Updated existing agent profile: %s", profileName))
	} else {
		newAgent := appconfig.AgentProfile{
			Name:        profileName,
			Provider:    appconfig.AgentProvider(m.Provider),
			Model:       m.ID,
			Enabled:     true,
			Priority:    1,
			Temperature: 0.7,
			MaxTokens:   m.MaxTokens,
			Timeout:     30,
			KeyringKey:  keyringKey,
		}
		if err := cfg.AddAgent(newAgent); err != nil {
			return fmt.Errorf("failed to add agent: %w", err)
		}
		style.Success(fmt.Sprintf("Created new agent profile: %s", profileName))
	}

	providerKey, err := auth.LoadAPIKey(m.Provider)
	if err == nil && providerKey != "" {
		if err := keyMgr.StoreAgentKey(profileName, providerKey); err != nil {
			style.Warning(fmt.Sprintf("Failed to link API key to agent: %v", err))
		}
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
	opts := catalog.ManagerOptions{
		CacheOptions: catalog.CacheOptions{
			CacheDuration:      24 * time.Hour,
			MinRefreshInterval: 1 * time.Hour,
		},
		APIKeyResolver: func(provider string) (string, error) {
			return auth.LoadAPIKey(provider)
		},
	}

	return catalog.NewCatalogManager(opts)
}

func printModels(models []catalog.Model, details bool) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if details {
		fmt.Fprintln(w, "ID\tName\tProvider\tTier\tContext\tPrice/1M\tCapabilities")
		fmt.Fprintln(w, "--\t----\t--------\t----\t-------\t--------\t------------")
	} else {
		fmt.Fprintln(w, "ID\tName\tTier\tContext")
		fmt.Fprintln(w, "--\t----\t----\t-------")
	}

	for _, m := range models {
		if !m.IsAvailable() {
			continue
		}

		if details {
			price := fmt.Sprintf("$%.2f", m.InputPrice)
			if m.InputPrice == 0 {
				price = "free"
			}
			caps := ""
			for i, c := range m.Capabilities {
				if i > 0 {
					caps += ","
				}
				caps += string(c)
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
				m.ID, m.Name, m.Provider, m.PricingTier, m.ContextWindow, price, caps)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\n",
				m.ID, m.Name, m.PricingTier, m.ContextWindow)
		}
	}

	w.Flush()
}

func printModelDetails(model *catalog.Model) {
	fmt.Printf("ID:          %s\n", model.ID)
	fmt.Printf("Name:        %s\n", model.Name)
	fmt.Printf("Provider:    %s\n", model.Provider)
	fmt.Printf("Status:      %s\n", model.Status)
	fmt.Printf("Description: %s\n", model.Description)
	fmt.Printf("\n")
	fmt.Printf("Context Window: %d tokens\n", model.ContextWindow)
	fmt.Printf("Max Output:     %d tokens\n", model.MaxTokens)
	fmt.Printf("Training Cutoff: %s\n", model.TrainingCutoff)
	fmt.Printf("\n")

	priceIn := "free"
	if model.InputPrice > 0 {
		priceIn = fmt.Sprintf("$%.2f", model.InputPrice)
	}
	priceOut := "free"
	if model.OutputPrice > 0 {
		priceOut = fmt.Sprintf("$%.2f", model.OutputPrice)
	}
	fmt.Printf("Pricing:\n")
	fmt.Printf("  Input:  %s per 1M tokens\n", priceIn)
	fmt.Printf("  Output: %s per 1M tokens\n", priceOut)
	fmt.Printf("  Tier:   %s\n", model.PricingTier)
	fmt.Printf("\n")

	fmt.Printf("Capabilities:\n")
	for _, cap := range model.Capabilities {
		fmt.Printf("  - %s\n", cap)
	}
	fmt.Printf("\n")

	if model.SupportsVision {
		fmt.Printf("✓ Supports Vision\n")
	}
	if model.SupportsToolUse {
		fmt.Printf("✓ Supports Tool Use\n")
	}

	if len(model.RecommendedFor) > 0 {
		fmt.Printf("\nRecommended for:\n")
		for _, rec := range model.RecommendedFor {
			fmt.Printf("  - %s\n", rec)
		}
	}

	if len(model.Aliases) > 0 {
		fmt.Printf("\nAliases:\n")
		for _, alias := range model.Aliases {
			fmt.Printf("  - %s\n", alias)
		}
	}
}

func formatDuration(d time.Duration) string {
	if d > 24*time.Hour {
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%dd", days)
	}
	if d > time.Hour {
		hours := int(d.Hours())
		return fmt.Sprintf("%dh", hours)
	}
	if d > time.Minute {
		mins := int(d.Minutes())
		return fmt.Sprintf("%dm", mins)
	}
	secs := int(d.Seconds())
	return fmt.Sprintf("%ds", secs)
}