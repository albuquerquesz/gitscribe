package cmd

import (
	"fmt"
	"strings"
	"syscall"

	"github.com/albuquerquesz/gitscribe/internal/auth"
	appconfig "github.com/albuquerquesz/gitscribe/internal/config"
	"github.com/albuquerquesz/gitscribe/internal/secrets"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	authProvider string
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication keys",
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

var authSetKeyCmd = &cobra.Command{
	Use:   "set-key",
	Short: "Set an API key for a provider",
	Example: `  gs auth set-key --provider groq
  gs auth set-key --provider openai`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSetKey()
	},
}

func init() {
	authLogoutCmd.Flags().StringVarP(&authProvider, "provider", "p", "anthropic", "Provider to logout from")
	
	authSetKeyCmd.Flags().StringVarP(&authProvider, "provider", "p", "", "Provider to set the key for")
	authSetKeyCmd.MarkFlagRequired("provider")

	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authSetKeyCmd)
	rootCmd.AddCommand(authCmd)
}

func runSetKey() error {
	fmt.Printf("Enter API key for %s: ", authProvider)
	
	byteKey, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println()

	apiKey := strings.TrimSpace(string(byteKey))
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	if err := auth.StoreAPIKey(authProvider, apiKey); err != nil {
		return fmt.Errorf("failed to store API key: %w", err)
	}

	if err := updateAgentProfile(authProvider, apiKey); err != nil {
		fmt.Printf("Warning: Could not update agent profile: %v\n", err)
	}

	fmt.Printf("✓ API key for %s stored successfully in system keyring\n", authProvider)
	return nil
}

func checkAuthStatus() error {
	providers := []string{"anthropic", "openai", "groq", "opencode"}
	
	for _, p := range providers {
		apiKey, err := auth.LoadAPIKey(p)
		if err != nil || apiKey == "" {
			fmt.Printf("%s: ✗ No API key stored\n", p)
			continue
		}
		masked := apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
		fmt.Printf("%s: ✓ API key stored (%s)\n", p, masked)
	}

	return nil
}

func logout() error {
	if err := auth.DeleteAPIKey(authProvider); err != nil {
		fmt.Printf("Warning: Could not delete API key: %v\n", err)
	}

	fmt.Printf("✓ Logged out from %s\n", authProvider)
	return nil
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