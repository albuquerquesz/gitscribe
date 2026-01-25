package cmd

import (
	"github.com/albqvictor1508/gitscribe/internal/store"
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
	if err != nil {
		return err
	}

	if len(key) == 0 || len(apiKey) == 0 {
	}

	return nil
}
