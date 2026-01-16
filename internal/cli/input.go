package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pfrederiksen/configdiff/parse"
)

// InputSource represents a configuration input (file or stdin)
type InputSource struct {
	Path   string
	Data   []byte
	Format string
}

// ReadInput reads configuration data from a file or stdin
func ReadInput(path string, formatHint string) (*InputSource, error) {
	var data []byte
	var err error

	// Read from stdin or file
	if path == "-" {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("failed to read from stdin: %w", err)
		}
	} else {
		data, err = os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %q: %w", path, err)
		}
	}

	// Determine format
	format := formatHint
	if format == "" || format == "auto" {
		format = detectFormat(path, data)
		if format == "" {
			return nil, fmt.Errorf("unable to detect format for %q\nHint: Specify format explicitly with --format", path)
		}
	}

	return &InputSource{
		Path:   path,
		Data:   data,
		Format: format,
	}, nil
}

// detectFormat attempts to detect the configuration format
func detectFormat(path string, data []byte) string {
	// First, try to detect from file extension
	if path != "-" {
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".yaml", ".yml":
			return "yaml"
		case ".json":
			return "json"
		case ".hcl", ".tf":
			return "hcl"
		}
	}

	// Try to detect from content
	return detectFromContent(data)
}

// detectFromContent attempts to detect format from content
func detectFromContent(data []byte) string {
	// Trim leading whitespace
	trimmed := bytes.TrimLeft(data, " \t\n\r")
	if len(trimmed) == 0 {
		return ""
	}

	// JSON starts with { or [
	if trimmed[0] == '{' || trimmed[0] == '[' {
		return "json"
	}

	// Try parsing as YAML (most permissive)
	// YAML is the default fallback since it's the most common
	if _, err := parse.ParseYAML(data); err == nil {
		return "yaml"
	}

	return ""
}
