package cmd

import (
	"fmt"

	"github.com/albuquerquesz/gitscribe/internal/auth"
	"github.com/spf13/cobra"
)

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout and remove stored credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		return logout()
	},
}

func init() {
	authLogoutCmd.Flags().StringVarP(&authProvider, "provider", "p", "anthropic", "Provider to logout from")

	authCmd.AddCommand(authLogoutCmd)
}

func logout() error {
	if err := auth.DeleteAPIKey(authProvider); err != nil {
		fmt.Printf("Warning: Could not delete API key: %v\n", err)
	}

	fmt.Printf("✓ Logged out from %s\n", authProvider)
	return nil
}
