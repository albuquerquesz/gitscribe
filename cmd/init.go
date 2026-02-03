package cmd

import (
	"fmt"
	"time"

	"github.com/albuquerquesz/gitscribe/internal/config"
	"github.com/albuquerquesz/gitscribe/internal/secrets"
	"github.com/albuquerquesz/gitscribe/internal/style"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize GitScribe configuration",
	Long: `Initialize GitScribe with your preferred AI provider.

This command will:
1. Check for existing OpenCode authentication
2. Set up your default AI provider
3. Configure your first agent profile

Example:
  gs init`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInit()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit() error {
	fmt.Println("Welcome to GitScribe!")
	fmt.Println("Let's set up your AI provider.\n")

	auth, err := secrets.LoadOpenCodeAuth()
	if err == nil && auth != nil && len(auth) > 0 {
		return setupFromOpenCode(auth)
	}

	return setupManual()
}

func setupFromOpenCode(auth secrets.OpenCodeAuth) error {
	fmt.Println("Found existing OpenCode authentication:")

	providers := auth.ListProviders()
	for _, p := range providers {
		if key, ok := auth.GetAPIKey(p); ok {
			masked := maskKey(key)
			status := "ready"
			if auth.IsTokenExpired(p) {
				status = "expired"
			} else if auth.IsTokenExpiringSoon(p, 24*time.Hour) {
				status = "expiring soon"
			}
			fmt.Printf("  %s %s (%s) [%s]\n", style.SuccessIcon(), p, masked, status)
		}
	}
	fmt.Println()

	if !style.ConfirmAction("Use OpenCode keys for GitScribe?") {
		return setupManual()
	}

	preferredProvider := selectPreferredProvider(providers)
	return createDefaultAgent(preferredProvider, "opencode")
}

func setupManual() error {
	fmt.Println("No OpenCode authentication found.")
	fmt.Println("Please configure your API key manually.\n")

	providers := []string{"anthropic", "openai", "groq", "openrouter"}

	fmt.Println("Available providers:")
	for i, p := range providers {
		fmt.Printf("  %d. %s\n", i+1, p)
	}
	fmt.Println()

	fmt.Println("Run 'gs auth set-key -p <provider>' to configure your API key.")
	fmt.Println("Example: gs auth set-key -p anthropic")

	return nil
}

func selectPreferredProvider(providers []string) string {
	preferred := []string{"anthropic", "openai", "groq"}

	for _, p := range preferred {
		for _, available := range providers {
			if p == available {
				return p
			}
		}
	}

	if len(providers) > 0 {
		return providers[0]
	}
	return "anthropic"
}

func createDefaultAgent(provider string, source string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	modelMap := map[string]string{
		"anthropic":  "claude-3-5-sonnet-20241022",
		"openai":     "gpt-4o",
		"groq":       "llama-3.3-70b-versatile",
		"openrouter": "anthropic/claude-3.5-sonnet",
	}

	model := modelMap[provider]
	if model == "" {
		model = "default"
	}

	agentName := fmt.Sprintf("%s-default", provider)

	existingAgent, err := cfg.GetAgentByName(agentName)
	if err == nil {
		existingAgent.Enabled = true
		cfg.Global.DefaultAgent = agentName
	} else {
		newAgent := config.AgentProfile{
			Name:        agentName,
			Provider:    config.AgentProvider(provider),
			Model:       model,
			Temperature: 0.7,
			MaxTokens:   4096,
			Timeout:     60,
			Enabled:     true,
			Priority:    1,
		}
		cfg.AddAgent(newAgent)
		cfg.Global.DefaultAgent = agentName
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("\n%s Configuration complete!\n", style.SuccessIcon())
	fmt.Printf("Default agent: %s (using %s from %s)\n", agentName, model, source)
	fmt.Println("\nYou can now use 'gs commit' to generate commit messages.")
	fmt.Println("Use 'gs models' to change your default model.")

	return nil
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}
