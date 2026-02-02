package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/albuquerquesz/gitscribe/internal/auth"
	appconfig "github.com/albuquerquesz/gitscribe/internal/config"
	"github.com/albuquerquesz/gitscribe/internal/providers"
	"github.com/spf13/cobra"
)

var (
	authProvider  string
	authPort      int
	authNoBrowser bool
	authTimeout   time.Duration
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with AI providers using OAuth2",
	Long: `Authenticate with AI providers using OAuth2 PKCE flow.

This command will:
1. Start a local HTTP server to receive the OAuth callback
2. Open your browser to the provider's authorization page
3. Exchange the authorization code for tokens
4. Generate an API key for the provider
5. Securely store the tokens and API key

Supported providers:
- anthropic (Anthropic/Claude)
- openai (OpenAI)

Example:
  gitscribe auth --provider anthropic
  gitscribe auth --provider openai
  gitscribe auth --provider anthropic --port 9090 --no-browser`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAuth()
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status for providers",
	RunE: func(cmd *cobra.Command, args []string) error {
		return checkAuthStatus()
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout and remove stored credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		return logout()
	},
}

func init() {
	authCmd.Flags().StringVarP(&authProvider, "provider", "p", "anthropic", "OAuth provider (anthropic)")
	authCmd.Flags().IntVar(&authPort, "port", 8085, "Local port for OAuth callback server")
	authCmd.Flags().BoolVar(&authNoBrowser, "no-browser", false, "Don't open browser automatically")
	authCmd.Flags().DurationVar(&authTimeout, "timeout", 5*time.Minute, "OAuth flow timeout")

	authLogoutCmd.Flags().StringVarP(&authProvider, "provider", "p", "anthropic", "Provider to logout from")

	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)
	rootCmd.AddCommand(authCmd)
}

func runAuth() error {
	var provider auth.Provider

	switch authProvider {
	case "anthropic", "claude":
		provider = providers.NewAnthropicProvider()
	case "openai":
		provider = providers.NewOpenAIProvider()
	default:
		return fmt.Errorf("unsupported provider: %s", authProvider)
	}

	// Check if already authenticated
	isAuth, err := auth.IsAuthenticated(provider.Name())
	if err != nil {
		fmt.Printf("Warning: Could not check authentication status: %v\n", err)
	}

	if isAuth {
		fmt.Printf("Already authenticated with %s. Use 'auth logout' first to re-authenticate.\n", provider.Name())
		return nil
	}

	fmt.Printf("Authenticating with %s...\n", provider.Name())
	fmt.Println("Scopes requested:", provider.Scopes())

	// Setup flow configuration
	flowConfig := &auth.FlowConfig{
		Provider:    provider,
		RedirectURL: fmt.Sprintf("http://localhost:%d/callback", authPort),
		Port:        authPort,
		Timeout:     authTimeout,
		OpenBrowser: !authNoBrowser,
	}

	// Run the OAuth flow
	flow := auth.NewFlow(flowConfig)
	ctx := context.Background()

	tokens, apiKey, err := flow.Run(ctx)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Store the tokens securely
	storage, err := auth.NewTokenStorage()
	if err != nil {
		return fmt.Errorf("failed to initialize token storage: %w", err)
	}

	if err := storage.SaveToken(provider.Name(), tokens); err != nil {
		return fmt.Errorf("failed to save tokens: %w", err)
	}

	// Store the API key in keyring
	if err := auth.StoreAPIKey(provider.Name(), apiKey); err != nil {
		return fmt.Errorf("failed to store API key: %w", err)
	}

	// Update agent profile to use the new API key
	if err := updateAgentProfile(provider.Name(), apiKey); err != nil {
		fmt.Printf("Warning: Could not update agent profile: %v\n", err)
	}

	fmt.Printf("\n✓ Successfully authenticated with %s\n", provider.Name())
	fmt.Printf("✓ API key generated and stored securely\n")
	fmt.Printf("✓ Tokens stored in OS keyring\n")

	return nil
}

func checkAuthStatus() error {
	// Check Anthropic
	isAuth, err := auth.IsAuthenticated("anthropic")
	if err != nil {
		fmt.Printf("anthropic: Error checking status: %v\n", err)
	} else if isAuth {
		fmt.Printf("anthropic: ✓ Authenticated\n")
	} else {
		fmt.Printf("anthropic: ✗ Not authenticated\n")
	}

	// Check for stored API keys
	if apiKey, err := auth.LoadAPIKey("anthropic"); err == nil && apiKey != "" {
		masked := apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
		fmt.Printf("anthropic: ✓ API key stored (%s)\n", masked)
	} else {
		fmt.Printf("anthropic: ✗ No API key stored\n")
	}

	return nil
}

func logout() error {
	storage, err := auth.NewTokenStorage()
	if err != nil {
		return fmt.Errorf("failed to initialize token storage: %w", err)
	}

	// Delete tokens
	if err := storage.DeleteToken(authProvider); err != nil {
		fmt.Printf("Warning: Could not delete tokens: %v\n", err)
	}

	// Delete API key
	if err := auth.DeleteAPIKey(authProvider); err != nil {
		fmt.Printf("Warning: Could not delete API key: %v\n", err)
	}

	fmt.Printf("✓ Logged out from %s\n", authProvider)
	return nil
}

func updateAgentProfile(providerName, apiKey string) error {
	// Load current config
	cfg, err := appconfig.Load()
	if err != nil {
		return err
	}

	// Find the agent profile for this provider and update it
	keyringKey := fmt.Sprintf("%s-oauth-api-key", providerName)

	for i := range cfg.Agents {
		if string(cfg.Agents[i].Provider) == providerName {
			cfg.Agents[i].KeyringKey = keyringKey
			cfg.Agents[i].Enabled = true

			// Save the API key to keyring with the correct key
			if err := auth.StoreAPIKey(keyringKey, apiKey); err != nil {
				return fmt.Errorf("failed to store API key for agent: %w", err)
			}
		}
	}

	return cfg.Save()
}
