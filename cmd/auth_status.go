package cmd

import (
	"fmt"

	"github.com/albuquerquesz/gitscribe/internal/auth"
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