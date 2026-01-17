package report

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/pfrederiksen/configdiff/diff"
)

// GenerateSideBySide creates a side-by-side comparison view.
func GenerateSideBySide(changes []diff.Change, opts Options) string {
	if len(changes) == 0 {
		return "No changes detected.\n"
	}

	// Save and restore color state
	originalNoColor := color.NoColor
	defer func() { color.NoColor = originalNoColor }()
	if opts.NoColor {
		color.NoColor = true
	}

	var b strings.Builder
	summary := summarizeChanges(changes)
	
	// Header
	b.WriteString("Summary: ")
	b.WriteString(formatSummary(summary, opts))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", 80))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("%-38s | %-38s\n", "Old Value", "New Value"))
	b.WriteString(strings.Repeat("─", 80))
	b.WriteString("\n")
	
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	
	for _, change := range changes {
		path := change.Path
		if len(path) > 76 {
			path = "..." + path[len(path)-73:]
		}
		
		b.WriteString(fmt.Sprintf("%s\n", path))
		
		switch change.Type {
		case diff.ChangeTypeAdd:
			oldVal := "(none)"
			newVal := formatValue(change.NewValue, opts.MaxValueLength)
			if !opts.NoColor {
				newVal = green(newVal)
			}
			b.WriteString(fmt.Sprintf("  %-36s | %s\n", oldVal, newVal))
			
		case diff.ChangeTypeRemove:
			oldVal := formatValue(change.OldValue, opts.MaxValueLength)
			if !opts.NoColor {
				oldVal = red(oldVal)
			}
			newVal := "(removed)"
			b.WriteString(fmt.Sprintf("  %-36s | %s\n", oldVal, newVal))
			
		case diff.ChangeTypeModify:
			oldVal := formatValue(change.OldValue, opts.MaxValueLength)
			newVal := formatValue(change.NewValue, opts.MaxValueLength)
			if !opts.NoColor {
				oldVal = yellow(oldVal)
				newVal = yellow(newVal)
			}
			b.WriteString(fmt.Sprintf("  %-36s | %s\n", oldVal, newVal))
			
		case diff.ChangeTypeMove:
			oldVal := formatValue(change.OldValue, opts.MaxValueLength)
			newVal := formatValue(change.NewValue, opts.MaxValueLength)
			b.WriteString(fmt.Sprintf("  %-36s → %s\n", oldVal, newVal))
		}
		
		b.WriteString("\n")
	}
	
	return b.String()
}
