package cmd

import (
	"errors"
	"fmt"

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

func config() error {
	apiKey, err := store.Get()
	if err == nil && len(apiKey) > 0 && len(key) == 0 {

		fmt.Printf("your API key is %v !!", apiKey)
		return nil
	}

	if !errors.Is(err, keyring.ErrNotFound) {
		fmt.Print("salve")
		return fmt.Errorf("error to get api key: %w", err)
	}

	if len(key) == 0 {
		result, err := style.Prompt("Enter your GROQ API Key...")
		if err != nil {
			return err
		}

		fmt.Print(result)
		// store.Save(result)
	}

	return nil
}
