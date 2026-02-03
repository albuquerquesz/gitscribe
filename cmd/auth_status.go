package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/albuquerquesz/gitscribe/internal/secrets"
	"github.com/spf13/cobra"
)

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status for providers",
	RunE: func(cmd *cobra.Command, args []string) error {
		return checkAuthStatus()
	},
}

func init() {
	authCmd.AddCommand(authStatusCmd)
}

func checkAuthStatus() error {
	providers := []string{"anthropic", "openai", "groq", "opencode", "openrouter", "ollama"}

	fmt.Println("Authentication Status:")
	fmt.Println(strings.Repeat("-", 60))

	for _, p := range providers {
		source, key := resolveKeyWithSource(p)
		if key != "" {
			masked := key[:4] + "..." + key[len(key)-4:]

			expiryInfo := ""
			if auth, err := secrets.LoadOpenCodeAuth(); err == nil && auth != nil {
				if auth.IsTokenExpired(p) {
					expiryInfo = " [EXPIRED]"
				} else if auth.IsTokenExpiringSoon(p, 24*time.Hour) {
					if expiry, ok := auth.GetTokenExpiry(p); ok {
						expiryInfo = fmt.Sprintf(" [expires in %s]", time.Until(expiry).Round(time.Hour))
					}
				}
			}

			fmt.Printf("%s: ✓ Configured (%s) [source: %s]%s\n", p, masked, source, expiryInfo)
		} else {
			fmt.Printf("%s: ✗ Not configured\n", p)
		}
	}

	return nil
}

func resolveKeyWithSource(provider string) (source, key string) {
	keyMgr := secrets.NewAgentKeyManager()

	if k, err := keyMgr.Retrieve(provider + "-api-key"); err == nil && k != "" {
		return "keyring", k
	}

	envVars := map[string]string{
		"anthropic":  "ANTHROPIC_API_KEY",
		"openai":     "OPENAI_API_KEY",
		"groq":       "GROQ_API_KEY",
		"opencode":   "OPENCODE_API_KEY",
		"openrouter": "OPENROUTER_API_KEY",
		"ollama":     "OLLAMA_API_KEY",
	}
	if envVar, ok := envVars[provider]; ok {
		if k := os.Getenv(envVar); k != "" {
			return "environment", k
		}
	}

	if auth, err := secrets.LoadOpenCodeAuth(); err == nil && auth != nil {
		if k, ok := auth.GetAPIKey(provider); ok {
			return "opencode", k
		}
	}

	return "", ""
}
