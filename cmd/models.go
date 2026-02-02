package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	appconfig "github.com/albuquerquesz/gitscribe/internal/config"
	"github.com/albuquerquesz/gitscribe/internal/models"
	"github.com/albuquerquesz/gitscribe/internal/secrets"
	"github.com/albuquerquesz/gitscribe/internal/tui"
)

// modelsCmd represents the models command
var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Browse and select AI models",
	Long: `Interactive model browser with configuration management.

Navigate through available AI models, configure providers, and set your default model.

Key bindings:
  enter     Select model as active
  c         Configure provider (OAuth2 or API key)
  r         Remove provider configuration
  /         Filter models
  ?         Toggle help
  q/esc     Quit`,
	RunE: runModels,
}

var (
	// Lipgloss styles for output
	successBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#04B575")).
			Padding(1, 2).
			MarginTop(1)

	successTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#04B575"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#808080"))
)

func runModels(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := appconfig.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize key manager
	keyMgr := secrets.NewAgentKeyManager()

	// Create TUI model
	model := tui.NewModel(cfg, keyMgr)

	// Run Bubble Tea program
	p := tea.NewProgram(model, tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	// Get final model state
	finalModel := m.(tui.Model)
	selected := finalModel.GetSelected()

	if selected == nil {
		// User quit without selecting
		return nil
	}

	// Check if provider is configured
	providerConfigured := false
	for _, agent := range cfg.Agents {
		if string(agent.Provider) == selected.Model.Provider && agent.Enabled {
			providerConfigured = true
			break
		}
	}

	if !providerConfigured {
		// Provider not configured - ask user what to do
		return handleUnconfiguredProvider(cfg, selected.Model)
	}

	// Provider is configured - just set as default
	return setModelAsDefault(cfg, selected.Model)
}

// handleUnconfiguredProvider handles when a user selects a model from an unconfigured provider
func handleUnconfiguredProvider(cfg *appconfig.Config, model models.ModelInfo) error {
	provider := models.Providers[model.Provider]

	var action string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(fmt.Sprintf("%s %s Not Configured", provider.Icon, provider.DisplayName)).
				Description(fmt.Sprintf("The model '%s' requires API access to %s.", model.Name, provider.DisplayName)).
				Options(
					huh.NewOption("Configure now (API Key)", "configure"),
					huh.NewOption("Cancel", "cancel"),
				).
				Value(&action),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	if action == "cancel" {
		return nil
	}

	// Configure the provider
	return configureProvider(cfg, model.Provider)
}

// configureProvider handles provider configuration
func configureProvider(cfg *appconfig.Config, providerKey string) error {
	provider := models.Providers[providerKey]

	// For now, use manual API key entry
	// OAuth2 can be added later for providers that support it
	var apiKey string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(fmt.Sprintf("Enter %s API Key", provider.DisplayName)).
				Description(fmt.Sprintf("Paste your API key for %s", provider.DisplayName)).
				EchoMode(huh.EchoModePassword).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("API key cannot be empty")
					}
					return nil
				}).
				Value(&apiKey),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	// Create agent profile
	agent := appconfig.AgentProfile{
		Name:        models.GenerateProfileName(providerKey, ""),
		Provider:    appconfig.AgentProvider(providerKey),
		Model:       "", // Will be set later when selecting specific model
		Enabled:     true,
		Priority:    1,
		Temperature: 0.7,
		MaxTokens:   4096,
		Timeout:     30,
		KeyringKey:  secrets.NewAgentKeyManager().GetAgentKeyName(models.GenerateProfileName(providerKey, "")),
	}

	// Add to config
	if err := cfg.AddAgent(agent); err != nil {
		return err
	}

	// Save API key to keyring
	keyMgr := secrets.NewAgentKeyManager()
	if err := keyMgr.StoreAgentKey(agent.Name, apiKey); err != nil {
		return fmt.Errorf("failed to store API key: %w", err)
	}

	// Save config
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("\n%s %s configured successfully!\n\n", provider.Icon, provider.DisplayName)

	return nil
}

// setModelAsDefault sets the selected model as the default agent
func setModelAsDefault(cfg *appconfig.Config, model models.ModelInfo) error {
	profileName := models.GenerateProfileName(model.Provider, model.ID)

	// Check if agent exists
	agent, err := cfg.GetAgentByName(profileName)
	if err != nil {
		// Create new agent profile for this specific model
		agent = &appconfig.AgentProfile{
			Name:        profileName,
			Provider:    appconfig.AgentProvider(model.Provider),
			Model:       model.ID,
			Enabled:     true,
			Priority:    1,
			Temperature: 0.7,
			MaxTokens:   model.MaxTokens,
			Timeout:     30,
			KeyringKey:  secrets.NewAgentKeyManager().GetAgentKeyName(profileName),
		}

		if err := cfg.AddAgent(*agent); err != nil {
			return err
		}
	} else {
		// Update existing agent with model
		agent.Model = model.ID
		agent.MaxTokens = model.MaxTokens
	}

	// Set as default
	if err := cfg.SetDefaultAgent(profileName); err != nil {
		return err
	}

	// Save config
	if err := cfg.Save(); err != nil {
		return err
	}

	// Display success message
	displaySuccessMessage(model, profileName)

	return nil
}

// displaySuccessMessage shows a styled success message
func displaySuccessMessage(model models.ModelInfo, profileName string) {
	provider := models.Providers[model.Provider]

	content := fmt.Sprintf(`%s Model Selected

Provider: %s %s
Model: %s
Profile: %s
Status: %s

%s Use with: gs cmt --agent %s`,
		successTitleStyle.Render("✓"),
		provider.Icon,
		provider.DisplayName,
		model.Name,
		profileName,
		successTitleStyle.Render("✓ Active"),
		infoStyle.Render("→"),
		profileName,
	)

	fmt.Println()
	fmt.Println(successBoxStyle.Render(content))
	fmt.Println()
}

func init() {
	rootCmd.AddCommand(modelsCmd)
}
