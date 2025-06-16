package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gw",
	Short: "Tool for managing GIT projects",
	Long: `Easily clone projects from predefined sources.
Safely remove projects from the working directory (the tool will ensure you
have not left anything unpublished).

gw go <project> --open
gw done <project>
`,
}

func ensureDir(directory *string, configDir string) {
	if *directory == "" {
		*directory = configDir
	}

	err := os.Mkdir(*directory, 0755)
	if err != nil && !os.IsExist(err) {
		log.Fatalf("failed to create the directory \"%s\": %s", *directory, err)
	}
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
