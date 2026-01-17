package cli

import (
	"fmt"
	"strings"

	"github.com/pfrederiksen/configdiff"
	"github.com/pfrederiksen/configdiff/internal/config"
)

// CLIOptions holds all CLI flag values
type CLIOptions struct {
	OldFile        string
	NewFile        string
	Format         string
	OldFormat      string
	NewFormat      string
	IgnorePaths    []string
	ArrayKeys      []string
	NumericStrings bool
	BoolStrings    bool
	StableOrder    bool
	OutputFormat   string
	NoColor        bool
	MaxValueLength int
	Quiet          bool
	ExitCode       bool
}

// ToLibraryOptions converts CLI options to configdiff library options
func (c *CLIOptions) ToLibraryOptions() (configdiff.Options, error) {
	// Parse array keys from "path=key" format
	arraySetKeys := make(map[string]string)
	for _, keySpec := range c.ArrayKeys {
		parts := strings.SplitN(keySpec, "=", 2)
		if len(parts) != 2 {
			return configdiff.Options{}, fmt.Errorf("invalid array-key format %q, expected path=key", keySpec)
		}
		path := parts[0]
		key := parts[1]

		// Ensure path starts with /
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}

		arraySetKeys[path] = key
	}

	return configdiff.Options{
		IgnorePaths:  c.IgnorePaths,
		ArraySetKeys: arraySetKeys,
		Coercions: configdiff.Coercions{
			NumericStrings: c.NumericStrings,
			BoolStrings:    c.BoolStrings,
		},
		StableOrder: c.StableOrder,
	}, nil
}

// GetOldFormat returns the format for the old file
func (c *CLIOptions) GetOldFormat() string {
	if c.OldFormat != "" {
		return c.OldFormat
	}
	return c.Format
}

// GetNewFormat returns the format for the new file
func (c *CLIOptions) GetNewFormat() string {
	if c.NewFormat != "" {
		return c.NewFormat
	}
	return c.Format
}

// ApplyConfigDefaults applies configuration file defaults to unset CLI options.
// CLI flags always take precedence over config file values.
func (c *CLIOptions) ApplyConfigDefaults(cfg *config.Config) {
	// Merge ignore paths (config file + CLI)
	if len(cfg.IgnorePaths) > 0 {
		// Create a map to deduplicate
		pathMap := make(map[string]bool)
		for _, p := range cfg.IgnorePaths {
			pathMap[p] = true
		}
		for _, p := range c.IgnorePaths {
			pathMap[p] = true
		}

		// Convert back to slice
		merged := make([]string, 0, len(pathMap))
		for p := range pathMap {
			merged = append(merged, p)
		}
		c.IgnorePaths = merged
	}

	// Merge array keys (config file + CLI)
	if len(cfg.ArrayKeys) > 0 {
		// Convert config map to CLI format (path=key)
		for path, key := range cfg.ArrayKeys {
			keySpec := fmt.Sprintf("%s=%s", path, key)
			c.ArrayKeys = append(c.ArrayKeys, keySpec)
		}
	}

	// Apply config defaults only if CLI flag wasn't set
	// For bool flags, we need to check if they were explicitly set
	// For now, we'll apply config if the CLI value is false (default)
	if !c.NumericStrings && cfg.NumericStrings {
		c.NumericStrings = cfg.NumericStrings
	}
	if !c.BoolStrings && cfg.BoolStrings {
		c.BoolStrings = cfg.BoolStrings
	}
	if !c.StableOrder && cfg.StableOrder {
		c.StableOrder = cfg.StableOrder
	}
	if !c.NoColor && cfg.NoColor {
		c.NoColor = cfg.NoColor
	}

	// Apply string defaults if not set
	if (c.OutputFormat == "" || c.OutputFormat == "report") && cfg.OutputFormat != "" {
		c.OutputFormat = cfg.OutputFormat
	}

	// Apply numeric defaults if not set
	if c.MaxValueLength == 0 && cfg.MaxValueLength > 0 {
		c.MaxValueLength = cfg.MaxValueLength
	}
}

// Validate validates the CLI options
func (c *CLIOptions) Validate() error {
	// Validate output format
	validFormats := map[string]bool{
		"report":  true,
		"compact": true,
		"json":    true,
		"patch":   true,
	}
	if !validFormats[c.OutputFormat] {
		return fmt.Errorf("invalid output format %q, must be one of: report, compact, json, patch", c.OutputFormat)
	}

	// Validate input format
	validInputFormats := map[string]bool{
		"auto": true,
		"yaml": true,
		"json": true,
		"hcl":  true,
	}
	if !validInputFormats[c.Format] {
		return fmt.Errorf("invalid format %q, must be one of: auto, yaml, json, hcl", c.Format)
	}
	if c.OldFormat != "" && !validInputFormats[c.OldFormat] {
		return fmt.Errorf("invalid old-format %q, must be one of: auto, yaml, json, hcl", c.OldFormat)
	}
	if c.NewFormat != "" && !validInputFormats[c.NewFormat] {
		return fmt.Errorf("invalid new-format %q, must be one of: auto, yaml, json, hcl", c.NewFormat)
	}

	return nil
}
