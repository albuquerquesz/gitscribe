package cmd

import (
	"fmt"

	"github.com/albqvictor1508/gitscribe/internal/store"
	"github.com/albqvictor1508/gitscribe/internal/style"
	"github.com/spf13/cobra"
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
	fmt.Printf("API KEY: %v\n", apiKey)
	fmt.Printf("KEY: %v", key)

	if err != nil {
		fmt.Print(apiKey)
		fmt.Print(key)

		return err
	}

	return nil

	if len(key) == 0 || len(apiKey) == 0 {
		result, err := style.Prompt("Enter your GROQ API Key...")
		if err != nil {
			return err
		}

		store.Save(result)
	}

	return nil
}
