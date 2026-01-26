package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/albqvictor1508/gitscribe/internal/ai"
	"github.com/albqvictor1508/gitscribe/internal/store"
	"github.com/albqvictor1508/gitscribe/internal/style"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

var key string

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		return config()
	},
}

func init() {
	configCmd.Flags().StringVarP(&key, "key", "k", "", "salve")

	rootCmd.AddCommand(configCmd)
}

func isKeyNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, keyring.ErrNotFound) {
		return true
	}

	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "não existe") ||
		strings.Contains(errMsg, "not found") ||
		strings.Contains(errMsg, "does not exist") ||
		strings.Contains(errMsg, "no such")
}

func config() error {
	apiKey, err := store.Get()
	if err == nil && len(apiKey) > 0 && len(key) == 0 {
		maskedKey := style.StringMask(apiKey)

		fmt.Printf("API key already configured: %v\n", maskedKey)
		return nil
	}

	if err != nil && !isKeyNotFoundError(err) {
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
