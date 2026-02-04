package cmd

import (
	"fmt"
	"slices"

	"github.com/albuquerquesz/gitscribe/internal/config"
	"github.com/albuquerquesz/gitscribe/internal/secrets"
	"github.com/albuquerquesz/gitscribe/internal/style"
	"github.com/spf13/cobra"
)

var (
	newAgentName     string
	newAgentProvider string
	newAgentModel    string
	newAgentKey      string
	newAgentBaseURL  string
)

var agentAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new agent profile",
	Example: `  gs agent add -n my-openai -p openai -m gpt-4
  gs agent add -n my-groq -p groq -m llama-3.3-70b-versatile -k gsk_xxx`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return addAgent()
	},
}

func init() {
	agentAddCmd.Flags().StringVarP(&newAgentName, "name", "n", "", "Agent profile name (required)")
	agentAddCmd.Flags().StringVarP(&newAgentProvider, "provider", "p", "", "Provider: openai, groq, claude, gemini, ollama (required)")
	agentAddCmd.Flags().StringVarP(&newAgentModel, "model", "m", "", "Model name (required)")
	agentAddCmd.Flags().StringVarP(&newAgentKey, "key", "k", "", "API key (will prompt if not provided)")
	agentAddCmd.Flags().StringVar(&newAgentBaseURL, "base-url", "", "Custom base URL (optional)")
	agentAddCmd.MarkFlagRequired("name")
	agentAddCmd.MarkFlagRequired("provider")
	agentAddCmd.MarkFlagRequired("model")

	agentCmd.AddCommand(agentAddCmd)
}

func addAgent() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	provider := config.AgentProvider(newAgentProvider)
	validProviders := []config.AgentProvider{
		config.ProviderOpenAI,
		config.ProviderGroq,
		config.ProviderClaude,
		config.ProviderGemini,
		config.ProviderOllama,
		config.ProviderOpenRouter,
		config.ProviderOpenCode,
	}

	valid := false
	if found := slices.Contains(validProviders, provider); found {
		valid = true
	}

	if !valid {
		return fmt.Errorf("invalid provider: %s", newAgentProvider)
	}

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

	if err := cfg.AddAgent(agent); err != nil {
		return err
	}

	keyMgr := secrets.NewAgentKeyManager()
	if err := keyMgr.StoreAgentKey(newAgentName, newAgentKey); err != nil {
		return fmt.Errorf("failed to store API key: %w", err)
	}

	if err := cfg.Save(); err != nil {
		keyMgr.DeleteAgentKey(newAgentName)
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Agent '%s' added successfully!\n", newAgentName)
	return nil
}
