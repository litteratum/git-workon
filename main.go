package main

import (
	"fmt"
	"log"
	"os"

	flag "github.com/spf13/pflag"
)

func main() {
	config := loadConfig()

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "You must pass a subcommand: start, done or config")
		os.Exit(1)
	}

	// common flags
	common := flag.NewFlagSet("common", flag.ExitOnError)
	directory := common.StringP("directory", "d", "", "working directory")

	// start command
	startCmd := newCmd("start", common)
	startCmd.AddFlagSet(common)
	startSource := startCmd.StringArrayP(
		"source",
		"s",
		[]string{},
		"git source to clone the project from. Pre-appended to the config's sources",
	)
	startEditor := startCmd.StringP("editor", "e", "", "editor to open the cloned project with")
	startOpen := startCmd.BoolP("open", "o", false, "open the project")

	// done command
	doneCmd := newCmd("done", common)
	doneCmd.AddFlagSet(common)
	doneForce := doneCmd.BoolP("force", "f", false, "force project removal")

	// config command
	configCmd := flag.NewFlagSet("config", flag.ExitOnError)

	switch os.Args[1] {
	case "start":
		startCmd.Parse(os.Args[2:])
		ensureDir(directory, config.Dir)
		wd := NewWorkingDir(*directory)
		err := wd.Start(
			startCmd.Args(),
			getSources(*startSource, config.Sources),
			getEditors(*startEditor, config.Editor),
			StartOpts{
				Open: *startOpen,
			},
		)
		if err != nil {
			exitWithError(err)
		}
	case "done":
		doneCmd.Parse(os.Args[2:])
		ensureDir(directory, config.Dir)
		wd := NewWorkingDir(*directory)
		err := wd.Done(
			doneCmd.Args(),
			DoneOpts{
				Force: *doneForce,
			},
		)
		if err != nil {
			exitWithError(err)
		}
	case "config":
		configCmd.Parse(os.Args[2:])
		config := loadConfig()
		fmt.Println(config)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1])
		fmt.Println("Available commands: start, done, config")
		os.Exit(1)
	}
}

func getEditors(cliEditor, configEditor string) []string {
	editors := []string{}
	if cliEditor != "" {
		editors = append(editors, cliEditor)
	}
	if configEditor != "" {
		editors = append(editors, configEditor)
	}

	envEditor, envEditorSet := os.LookupEnv("EDITOR")
	if envEditorSet {
		editors = append(editors, envEditor)
	}
	editors = append(editors, []string{"vim", "vi"}...)
	return editors
}

func getSources(cliSources, configSources []string) []string {
	sources := []string{}
	sources = append(sources, cliSources...)
	sources = append(sources, configSources...)
	if len(sources) == 0 {
		log.Fatalln("no GIT sources provided")
	}

	return sources
}

func newCmd(name string, common *flag.FlagSet) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	fs.AddFlagSet(common)
	return fs
}

func exitWithError(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
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
