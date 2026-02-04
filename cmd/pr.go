package cmd

import "github.com/spf13/cobra"

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "",
	RunE: func(cmd *cobra.Command, args []string) error {
		return realizePr()
	},
}

func init() {
	rootCmd.AddCommand(prCmd)
}

func realizePr() error {
	return nil
}
