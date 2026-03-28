package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

// Config holds the parsed contents of ~/.ghostmanrc.
type Config struct {
	Collections []string `yaml:"collections"`
	DefaultEnv  string   `yaml:"default_env"`
}

// Load reads ~/.ghostmanrc and returns the parsed Config.
// If the file does not exist, a zero-value Config and nil error are returned.
func Load() (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, fmt.Errorf("home dir: %w", err)
	}
	return LoadFrom(filepath.Join(home, ".config", "ghostman", "config.yaml"))
}

// LoadFrom reads the config file at the given path and returns the parsed Config.
// If the file does not exist, a zero-value Config and nil error are returned.
func LoadFrom(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("config.yaml: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("config.yaml: %w", err)
	}

	return cfg, nil
}
