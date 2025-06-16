package cmd

import (
	"github.com/litteratum/git-workon/internal/app"
	"github.com/spf13/cobra"
)

func buildGoCommand() *cobra.Command {
	var (
		open      bool
		directory string
		sources   []string
		editor    string
	)

	cmd := &cobra.Command{
		Use:   "go <project>...",
		Short: "Start the project",
		Long: `Clone the project (if needed) into the working directory.
Sources from the configuration are used but may be extended by -s/--source.
Sources are resolved in the following order:
	* Defined by -s/--source
	* Cached for the project
	* Sources from the configuration

Use -o/--open to open the project in the configured editor.
Override the editor using -e/--editor.
	`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := app.LoadConfig()
			cache := app.NewCacheFromFile()
			ensureDir(&directory, config.Dir)
			wd := app.NewWorkingDir(directory, config, cache)
			return wd.Go(
				args,
				sources,
				editor,
				app.GoOpts{
					Open: open,
				},
			)
		},
		SilenceUsage: true,
	}

	cmd.Flags().BoolVarP(&open, "open", "o", false, "open the project in the configured editor")
	cmd.Flags().StringVarP(&directory, "directory", "d", "", "working directory")
	cmd.Flags().StringSliceVarP(&sources, "source", "s", []string{}, "additional sources")
	cmd.Flags().StringVarP(&editor, "editor", "e", "", "editor to use")

	return cmd
}

func init() {
	rootCmd.AddCommand(buildGoCommand())
}
