package cmd

import (
	"fmt"
	"strings"
	"syscall"

	"github.com/albuquerquesz/gitscribe/internal/auth"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

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
	authSetKeyCmd.Flags().StringVarP(&authProvider, "provider", "p", "", "Provider to set the key for")
	authSetKeyCmd.MarkFlagRequired("provider")

	authCmd.AddCommand(authSetKeyCmd)
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
