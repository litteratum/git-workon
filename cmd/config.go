package cmd

import (
	"fmt"

	"github.com/litteratum/git-workon/internal/app"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show the current config",
	Run: func(cmd *cobra.Command, args []string) {
		config := app.LoadConfig()
		fmt.Println(config)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
