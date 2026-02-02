package cmd

import (
	"fmt"
	"strings"

	"github.com/albuquerquesz/gitscribe/internal/config"
	"github.com/albuquerquesz/gitscribe/internal/secrets"
	"github.com/albuquerquesz/gitscribe/internal/style"
	"github.com/spf13/cobra"
)

// agentCmd manages agent profiles and API keys
var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage AI agents and their API keys",
	Long:  "Configure, add, remove, and manage AI agent profiles and API keys",
}

var (
	// Flags for agent add
	newAgentName     string
	newAgentProvider string
	newAgentModel    string
	newAgentKey      string
	newAgentBaseURL  string

	// Flags for agent set-default
	defaultAgentName string
)

func init() {
	// Add subcommands
	agentCmd.AddCommand(agentAddCmd)
	agentCmd.AddCommand(agentRemoveCmd)
	agentCmd.AddCommand(agentSetDefaultCmd)
	agentCmd.AddCommand(agentSetKeyCmd)
	agentCmd.AddCommand(agentListCmd)

	// Add flags
	agentAddCmd.Flags().StringVarP(&newAgentName, "name", "n", "", "Agent profile name (required)")
	agentAddCmd.Flags().StringVarP(&newAgentProvider, "provider", "p", "", "Provider: openai, groq, claude, gemini, ollama (required)")
	agentAddCmd.Flags().StringVarP(&newAgentModel, "model", "m", "", "Model name (required)")
	agentAddCmd.Flags().StringVarP(&newAgentKey, "key", "k", "", "API key (will prompt if not provided)")
	agentAddCmd.Flags().StringVar(&newAgentBaseURL, "base-url", "", "Custom base URL (optional)")
	agentAddCmd.MarkFlagRequired("name")
	agentAddCmd.MarkFlagRequired("provider")
	agentAddCmd.MarkFlagRequired("model")

	agentSetDefaultCmd.Flags().StringVarP(&defaultAgentName, "name", "n", "", "Agent name to set as default (required)")
	agentSetDefaultCmd.MarkFlagRequired("name")

	rootCmd.AddCommand(agentCmd)
}

var agentAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new agent profile",
	Example: `  gs agent add -n my-openai -p openai -m gpt-4
  gs agent add -n my-groq -p groq -m llama-3.3-70b-versatile -k gsk_xxx`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return addAgent()
	},
}

var agentRemoveCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove an agent profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return removeAgent(args[0])
	},
}

var agentSetDefaultCmd = &cobra.Command{
	Use:   "set-default",
	Short: "Set the default agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		return setDefaultAgent(defaultAgentName)
	},
}

var agentSetKeyCmd = &cobra.Command{
	Use:   "set-key [name]",
	Short: "Set or update API key for an agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return setAgentKey(args[0])
	},
}

var agentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		return listAllAgents()
	},
}

func addAgent() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate provider
	provider := config.AgentProvider(newAgentProvider)
	validProviders := []config.AgentProvider{
		config.ProviderOpenAI,
		config.ProviderGroq,
		config.ProviderClaude,
		config.ProviderGemini,
		config.ProviderOllama,
		config.ProviderOpenRouter,
	}

	valid := false
	for _, p := range validProviders {
		if p == provider {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid provider: %s", newAgentProvider)
	}

	// Prompt for API key if not provided
	if newAgentKey == "" {
		prompt := fmt.Sprintf("Enter API key for %s (%s):", newAgentName, provider)
		key, err := style.Prompt(prompt)
		if err != nil {
			return err
		}
		newAgentKey = key
	}

	if newAgentKey == "" {
		return fmt.Errorf("API key is required")
	}

	// Create agent profile
	agent := config.AgentProfile{
		Name:        newAgentName,
		Provider:    provider,
		Model:       newAgentModel,
		BaseURL:     newAgentBaseURL,
		Enabled:     true,
		Priority:    1,
		Temperature: 0.7,
		MaxTokens:   2048,
		Timeout:     30,
		KeyringKey:  secrets.NewAgentKeyManager().GetAgentKeyName(newAgentName),
	}

	// Add to config
	if err := cfg.AddAgent(agent); err != nil {
		return err
	}

	// Save API key to keyring
	keyMgr := secrets.NewAgentKeyManager()
	if err := keyMgr.StoreAgentKey(newAgentName, newAgentKey); err != nil {
		return fmt.Errorf("failed to store API key: %w", err)
	}

	// Save config
	if err := cfg.Save(); err != nil {
		// Cleanup key if config save fails
		keyMgr.DeleteAgentKey(newAgentName)
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Agent '%s' added successfully!\n", newAgentName)
	fmt.Printf("   Provider: %s\n", provider)
	fmt.Printf("   Model: %s\n", newAgentModel)

	return nil
}

func removeAgent(name string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Confirm removal
	confirm, err := style.Prompt(fmt.Sprintf("Are you sure you want to remove agent '%s'? (yes/no): ", name))
	if err != nil {
		return err
	}

	if strings.ToLower(confirm) != "yes" {
		fmt.Println("Cancelled.")
		return nil
	}

	// Remove from config
	if err := cfg.RemoveAgent(name); err != nil {
		return err
	}

	// Remove API key from keyring
	keyMgr := secrets.NewAgentKeyManager()
	if err := keyMgr.DeleteAgentKey(name); err != nil {
		// Log but don't fail - key might not exist
		fmt.Printf("Warning: could not remove API key: %v\n", err)
	}

	// Save config
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Agent '%s' removed successfully!\n", name)
	return nil
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

func setAgentKey(name string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Verify agent exists
	agent, err := cfg.GetAgentByName(name)
	if err != nil {
		return err
	}

	// Prompt for new key
	prompt := fmt.Sprintf("Enter new API key for %s (%s):", name, agent.Provider)
	newKey, err := style.Prompt(prompt)
	if err != nil {
		return err
	}

	if newKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	// Update key in keyring
	keyMgr := secrets.NewAgentKeyManager()
	if err := keyMgr.StoreAgentKey(name, newKey); err != nil {
		return fmt.Errorf("failed to store API key: %w", err)
	}

	fmt.Printf("✅ API key updated for agent '%s'\n", name)
	return nil
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
