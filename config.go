package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/kirsle/configdir"
)

var ConfigDir = configdir.LocalConfig("git_workon")
var ConfigPath = filepath.Join(ConfigDir, "config.json")

type Config struct {
	Dir     string   `json:"dir"`
	Editor  string   `json:"editor"`
	Sources []string `json:"sources"`
}

func (c Config) String() string {
	// Marshal JSON
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal config: %s", err)
	}
	return string(data)
}

func NewDefaultConfig() Config {
	return Config{
		Dir:     "~/.workon",
		Editor:  "vi",
		Sources: []string{},
	}
}

func loadConfig() Config {
	err := configdir.MakePath(ConfigDir)
	if err != nil {
		log.Fatalf("failed to ensure the configuration directory exists: %s", err)
	}

	var config Config

	if _, err = os.Stat(ConfigPath); os.IsNotExist(err) {
		config := NewDefaultConfig()
		fh, err := os.Create(ConfigPath)
		if err != nil {
			log.Fatalf("failed to create the configuration file at %s: %s", ConfigPath, err)
		}
		defer fh.Close()

		encoder := json.NewEncoder(fh)
		err = encoder.Encode(&config)
		if err != nil {
			log.Fatalf("failed to encode config to %s: %s", ConfigPath, err)
		}

		log.Fatalf("missing configuration at %s", ConfigPath)
	} else {
		fh, err := os.Open(ConfigPath)
		if err != nil {
			log.Fatalf("failed to open configuration file at %s: %s", ConfigPath, err)
		}
		defer fh.Close()

		decoder := json.NewDecoder(fh)
		err = decoder.Decode(&config)

		if err != nil {
			log.Fatalf("failed to decode configuration from %s: %s", ConfigPath, err)
		}
	}

	if strings.HasPrefix(config.Dir, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("failed to get home directory: %s", err)
		}
		config.Dir = strings.Replace(config.Dir, "~", homeDir, 1)
	}

	return config
}
