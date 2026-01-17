package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "configdiff-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory for test
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	t.Run("no config file", func(t *testing.T) {
		cfg, err := Load()
		if err != nil {
			t.Errorf("Load() error = %v, want nil", err)
		}
		if cfg == nil {
			t.Error("Load() returned nil config")
		}
	})

	t.Run("load .configdiffrc", func(t *testing.T) {
		configContent := `ignore_paths:
  - /test/path
  - /another/path
array_keys:
  /containers: name
numeric_strings: true
bool_strings: false
stable_order: true
output_format: compact
max_value_length: 50
no_color: true
`
		if err := os.WriteFile(".configdiffrc", []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}
		defer os.Remove(".configdiffrc")

		cfg, err := Load()
		if err != nil {
			t.Errorf("Load() error = %v, want nil", err)
		}

		// Verify loaded values
		if len(cfg.IgnorePaths) != 2 {
			t.Errorf("IgnorePaths length = %d, want 2", len(cfg.IgnorePaths))
		}
		if len(cfg.ArrayKeys) != 1 {
			t.Errorf("ArrayKeys length = %d, want 1", len(cfg.ArrayKeys))
		}
		if !cfg.NumericStrings {
			t.Error("NumericStrings = false, want true")
		}
		if cfg.BoolStrings {
			t.Error("BoolStrings = true, want false")
		}
		if !cfg.StableOrder {
			t.Error("StableOrder = false, want true")
		}
		if cfg.OutputFormat != "compact" {
			t.Errorf("OutputFormat = %q, want %q", cfg.OutputFormat, "compact")
		}
		if cfg.MaxValueLength != 50 {
			t.Errorf("MaxValueLength = %d, want 50", cfg.MaxValueLength)
		}
		if !cfg.NoColor {
			t.Error("NoColor = false, want true")
		}
	})

	t.Run("load .configdiff.yaml", func(t *testing.T) {
		configContent := `ignore_paths:
  - /yaml/path
output_format: json
`
		if err := os.WriteFile(".configdiff.yaml", []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}
		defer os.Remove(".configdiff.yaml")

		cfg, err := Load()
		if err != nil {
			t.Errorf("Load() error = %v, want nil", err)
		}

		if len(cfg.IgnorePaths) != 1 {
			t.Errorf("IgnorePaths length = %d, want 1", len(cfg.IgnorePaths))
		}
		if cfg.OutputFormat != "json" {
			t.Errorf("OutputFormat = %q, want %q", cfg.OutputFormat, "json")
		}
	})

	t.Run("priority: .configdiffrc over .configdiff.yaml", func(t *testing.T) {
		// Create both files
		rcContent := `output_format: compact`
		yamlContent := `output_format: json`

		if err := os.WriteFile(".configdiffrc", []byte(rcContent), 0644); err != nil {
			t.Fatalf("Failed to write .configdiffrc: %v", err)
		}
		defer os.Remove(".configdiffrc")

		if err := os.WriteFile(".configdiff.yaml", []byte(yamlContent), 0644); err != nil {
			t.Fatalf("Failed to write .configdiff.yaml: %v", err)
		}
		defer os.Remove(".configdiff.yaml")

		cfg, err := Load()
		if err != nil {
			t.Errorf("Load() error = %v, want nil", err)
		}

		// Should load .configdiffrc (first in priority)
		if cfg.OutputFormat != "compact" {
			t.Errorf("OutputFormat = %q, want %q (from .configdiffrc)", cfg.OutputFormat, "compact")
		}
	})
}

func TestLoadFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "configdiff-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("valid config", func(t *testing.T) {
		path := filepath.Join(tmpDir, "valid.yaml")
		content := `ignore_paths: [/path1, /path2]`

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		cfg, err := loadFile(path)
		if err != nil {
			t.Errorf("loadFile() error = %v, want nil", err)
		}
		if len(cfg.IgnorePaths) != 2 {
			t.Errorf("IgnorePaths length = %d, want 2", len(cfg.IgnorePaths))
		}
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := loadFile("/nonexistent/file.yaml")
		if err == nil {
			t.Error("loadFile() error = nil, want error")
		}
	})

	t.Run("invalid yaml", func(t *testing.T) {
		path := filepath.Join(tmpDir, "invalid.yaml")
		content := `invalid: yaml: content: [[[`

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		_, err := loadFile(path)
		if err == nil {
			t.Error("loadFile() error = nil, want YAML parse error")
		}
	})
}
