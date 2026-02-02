package cmd

import (
	"errors"
	"fmt"

	"github.com/albuquerquesz/gitscribe/internal/ai"
	"github.com/albuquerquesz/gitscribe/internal/store"
	"github.com/albuquerquesz/gitscribe/internal/style"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

var key string

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfig()
	},
}

func init() {
	configCmd.Flags().StringVarP(&key, "key", "k", "", "salve")

	rootCmd.AddCommand(configCmd)
}

func runConfig() error {
	apiKey, err := store.Get()
	if err == nil && len(apiKey) > 0 && len(key) == 0 {
		maskedKey := style.StringMask(apiKey)

		fmt.Printf("API key already configured: %v\n", maskedKey)
		return nil
	}

	if err != nil && errors.Is(err, keyring.ErrNotFound) {
		return fmt.Errorf("error to get api key: %w", err)
	}

	if len(key) == 0 {
		res, err := style.Prompt("Enter your GROQ API Key...")
		if err != nil {
			return err
		}
		key = res
	}

	valid := ai.ValidateToken(key)
	if !valid {
		return fmt.Errorf("invalid api key")
	}

	if err := store.Save(key); err != nil {
		return err
	}

	fmt.Println("API Key saved successfully!")
	return nil
}
