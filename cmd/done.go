package cmd

import (
	"github.com/litteratum/git-workon/internal/app"
	"github.com/spf13/cobra"
)

func buildDoneCommand() *cobra.Command {
	var (
		directory string
		force     bool
	)

	cmd := &cobra.Command{
		Use:   "done [<project>...]",
		Short: "Finish the project",
		Long:  `Remove the project(s) from the working directory`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config := app.LoadConfig()
			cache := app.NewCacheFromFile()
			ensureDir(&directory, config.Dir)
			wd := app.NewWorkingDir(directory, config, cache)
			return wd.Done(
				args,
				app.DoneOpts{Force: force},
			)
		},
		SilenceUsage: true,
	}

	cmd.Flags().StringVarP(&directory, "directory", "d", "", "working directory")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "force")

	return cmd
}
func init() {
	rootCmd.AddCommand(buildDoneCommand())
}
