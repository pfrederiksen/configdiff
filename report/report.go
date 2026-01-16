// Package report provides human-friendly reporting for configuration diffs.
package report

import (
	"fmt"
	"strings"

	"github.com/pfrederiksen/configdiff/diff"
	"github.com/pfrederiksen/configdiff/tree"
)

// Options configures report generation.
type Options struct {
	// Compact reduces whitespace in the output.
	Compact bool

	// ShowValues includes before/after values in the report.
	ShowValues bool

	// MaxValueLength limits the length of displayed values.
	// Values longer than this are truncated. 0 means no limit.
	MaxValueLength int

	// ContextLines shows N lines of context around changes (not implemented yet).
	ContextLines int
}

// DefaultOptions returns sensible defaults for report generation.
func DefaultOptions() Options {
	return Options{
		Compact:        false,
		ShowValues:     true,
		MaxValueLength: 80,
		ContextLines:   0,
	}
}

// Generate creates a human-friendly report from changes.
func Generate(changes []diff.Change, opts Options) string {
	if len(changes) == 0 {
		return "No changes detected.\n"
	}

	var b strings.Builder

	// Write summary
	summary := summarizeChanges(changes)
	b.WriteString(formatSummary(summary))

	if !opts.Compact {
		b.WriteString("\n")
	}

	// Write detailed changes
	b.WriteString("Changes:\n")
	for i, change := range changes {
		b.WriteString(formatChange(change, opts))
		if !opts.Compact && i < len(changes)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// Summary holds statistics about changes.
type Summary struct {
	Total    int
	Added    int
	Removed  int
	Modified int
	Moved    int
}

// summarizeChanges counts changes by type.
func summarizeChanges(changes []diff.Change) Summary {
	var s Summary
	s.Total = len(changes)

	for _, change := range changes {
		switch change.Type {
		case diff.ChangeTypeAdd:
			s.Added++
		case diff.ChangeTypeRemove:
			s.Removed++
		case diff.ChangeTypeModify:
			s.Modified++
		case diff.ChangeTypeMove:
			s.Moved++
		}
	}

	return s
}

// formatSummary creates a summary header.
func formatSummary(s Summary) string {
	parts := make([]string, 0, 4)

	if s.Added > 0 {
		parts = append(parts, fmt.Sprintf("+%d added", s.Added))
	}
	if s.Removed > 0 {
		parts = append(parts, fmt.Sprintf("-%d removed", s.Removed))
	}
	if s.Modified > 0 {
		parts = append(parts, fmt.Sprintf("~%d modified", s.Modified))
	}
	if s.Moved > 0 {
		parts = append(parts, fmt.Sprintf("↔%d moved", s.Moved))
	}

	summary := strings.Join(parts, ", ")
	return fmt.Sprintf("Summary: %s (%d total)\n", summary, s.Total)
}

// formatChange creates a formatted string for a single change.
func formatChange(change diff.Change, opts Options) string {
	var b strings.Builder

	// Change type symbol and path
	symbol := getChangeSymbol(change.Type)
	b.WriteString(fmt.Sprintf("  %s %s", symbol, change.Path))

	// Add values if requested
	if opts.ShowValues {
		switch change.Type {
		case diff.ChangeTypeAdd:
			val := formatValue(change.NewValue, opts.MaxValueLength)
			b.WriteString(fmt.Sprintf(" = %s", val))

		case diff.ChangeTypeRemove:
			val := formatValue(change.OldValue, opts.MaxValueLength)
			b.WriteString(fmt.Sprintf(" (was: %s)", val))

		case diff.ChangeTypeModify:
			oldVal := formatValue(change.OldValue, opts.MaxValueLength)
			newVal := formatValue(change.NewValue, opts.MaxValueLength)
			b.WriteString(fmt.Sprintf(": %s → %s", oldVal, newVal))
		}
	}

	b.WriteString("\n")
	return b.String()
}

// getChangeSymbol returns a symbol for each change type.
func getChangeSymbol(ct diff.ChangeType) string {
	switch ct {
	case diff.ChangeTypeAdd:
		return "+"
	case diff.ChangeTypeRemove:
		return "-"
	case diff.ChangeTypeModify:
		return "~"
	case diff.ChangeTypeMove:
		return "↔"
	default:
		return "?"
	}
}

// formatValue converts a node value to a display string.
func formatValue(node *tree.Node, maxLen int) string {
	if node == nil {
		return "<nil>"
	}

	var val string

	switch node.Kind {
	case tree.KindNull:
		val = "null"

	case tree.KindBool:
		val = fmt.Sprintf("%v", node.Value)

	case tree.KindNumber:
		// Format numbers nicely
		if f, ok := node.Value.(float64); ok {
			// Check if it's a whole number
			if f == float64(int64(f)) {
				val = fmt.Sprintf("%d", int64(f))
			} else {
				val = fmt.Sprintf("%g", f)
			}
		} else {
			val = fmt.Sprintf("%v", node.Value)
		}

	case tree.KindString:
		val = fmt.Sprintf("%q", node.Value)

	case tree.KindObject:
		val = fmt.Sprintf("{...} (%d keys)", len(node.Object))

	case tree.KindArray:
		val = fmt.Sprintf("[...] (%d items)", len(node.Array))

	default:
		val = fmt.Sprintf("<%s>", node.Kind)
	}

	// Truncate if needed
	if maxLen > 0 && len(val) > maxLen {
		val = val[:maxLen-3] + "..."
	}

	return val
}

// GenerateCompact is a convenience function for compact reports.
func GenerateCompact(changes []diff.Change) string {
	opts := DefaultOptions()
	opts.Compact = true
	opts.ShowValues = false
	return Generate(changes, opts)
}

// GenerateDetailed is a convenience function for detailed reports.
func GenerateDetailed(changes []diff.Change) string {
	opts := DefaultOptions()
	opts.Compact = false
	opts.ShowValues = true
	return Generate(changes, opts)
}
