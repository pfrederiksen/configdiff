package cli

import (
	"encoding/json"
	"fmt"

	"github.com/pfrederiksen/configdiff"
	"github.com/pfrederiksen/configdiff/report"
)

// OutputOptions controls how output is formatted
type OutputOptions struct {
	Format         string
	NoColor        bool
	MaxValueLength int
}

// FormatOutput formats the diff result according to the specified options
func FormatOutput(result *configdiff.Result, opts OutputOptions) (string, error) {
	switch opts.Format {
	case "report":
		// Detailed report with values
		return report.Generate(result.Changes, report.Options{
			Compact:        false,
			ShowValues:     true,
			MaxValueLength: opts.MaxValueLength,
			NoColor:        opts.NoColor,
		}), nil

	case "compact":
		// Compact report (paths only)
		return report.Generate(result.Changes, report.Options{
			Compact:    true,
			ShowValues: false,
			NoColor:    opts.NoColor,
		}), nil

	case "json":
		// JSON serialized changes
		data, err := json.MarshalIndent(result.Changes, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal changes to JSON: %w", err)
		}
		return string(data), nil

	case "patch":
		// JSON Patch format
		data, err := result.Patch.ToJSONIndent()
		if err != nil {
			return "", fmt.Errorf("failed to marshal patch to JSON: %w", err)
		}
		return string(data), nil

	default:
		return "", fmt.Errorf("unsupported output format: %s", opts.Format)
	}
}

// HasChanges returns true if there are any changes in the result
func HasChanges(result *configdiff.Result) bool {
	return len(result.Changes) > 0
}
