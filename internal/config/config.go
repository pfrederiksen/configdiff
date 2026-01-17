// Package config handles loading configuration from files.
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the configuration file structure.
type Config struct {
	// IgnorePaths is a list of paths to ignore in diffs.
	IgnorePaths []string `yaml:"ignore_paths"`

	// ArrayKeys maps paths to key fields for array-as-set behavior.
	ArrayKeys map[string]string `yaml:"array_keys"`

	// NumericStrings enables treating string numbers as numbers.
	NumericStrings bool `yaml:"numeric_strings"`

	// BoolStrings enables treating string booleans as booleans.
	BoolStrings bool `yaml:"bool_strings"`

	// StableOrder enables stable sorting of object keys and array elements.
	StableOrder bool `yaml:"stable_order"`

	// OutputFormat specifies the default output format (report/compact/json/patch).
	OutputFormat string `yaml:"output_format"`

	// MaxValueLength limits the displayed value length in reports.
	MaxValueLength int `yaml:"max_value_length"`

	// NoColor disables colored output.
	NoColor bool `yaml:"no_color"`
}

// Load attempts to load configuration from standard locations.
// It checks the following locations in order:
//   1. ./.configdiffrc
//   2. ./.configdiff.yaml
//   3. ~/.configdiffrc
//   4. ~/.configdiff.yaml
//
// Returns the first config file found, or an empty config if none exist.
func Load() (*Config, error) {
	locations := []string{
		".configdiffrc",
		".configdiff.yaml",
	}

	// Add home directory locations
	if home, err := os.UserHomeDir(); err == nil {
		locations = append(locations,
			filepath.Join(home, ".configdiffrc"),
			filepath.Join(home, ".configdiff.yaml"),
		)
	}

	// Try each location
	for _, path := range locations {
		if cfg, err := loadFile(path); err == nil {
			return cfg, nil
		}
	}

	// No config file found, return empty config
	return &Config{}, nil
}

// loadFile loads configuration from a specific file path.
func loadFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
