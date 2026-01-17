package report

import (
	"fmt"
	"strings"

	"github.com/pfrederiksen/configdiff/diff"
)

// GenerateGitDiff creates output in git diff format.
// This is useful for git diff driver integration.
func GenerateGitDiff(changes []diff.Change, oldFile, newFile string) string {
	if len(changes) == 0 {
		return ""
	}

	var b strings.Builder
	
	// Git diff header
	b.WriteString(fmt.Sprintf("diff --configdiff a/%s b/%s\n", oldFile, newFile))
	b.WriteString(fmt.Sprintf("--- a/%s\n", oldFile))
	b.WriteString(fmt.Sprintf("+++ b/%s\n", newFile))
	
	// Group changes by path for better readability
	pathChanges := make(map[string][]diff.Change)
	var paths []string
	
	for _, change := range changes {
		// Extract base path (before array indices)
		basePath := strings.Split(change.Path, "[")[0]
		if pathChanges[basePath] == nil {
			paths = append(paths, basePath)
		}
		pathChanges[basePath] = append(pathChanges[basePath], change)
	}
	
	// Output changes grouped by path
	for _, basePath := range paths {
		b.WriteString(fmt.Sprintf("@@ %s @@\n", basePath))
		
		for _, change := range pathChanges[basePath] {
			switch change.Type {
			case diff.ChangeTypeAdd:
				val := formatValue(change.NewValue, 0)
				b.WriteString(fmt.Sprintf("+%s: %s\n", change.Path, val))
				
			case diff.ChangeTypeRemove:
				val := formatValue(change.OldValue, 0)
				b.WriteString(fmt.Sprintf("-%s: %s\n", change.Path, val))
				
			case diff.ChangeTypeModify:
				oldVal := formatValue(change.OldValue, 0)
				newVal := formatValue(change.NewValue, 0)
				b.WriteString(fmt.Sprintf("-%s: %s\n", change.Path, oldVal))
				b.WriteString(fmt.Sprintf("+%s: %s\n", change.Path, newVal))
				
			case diff.ChangeTypeMove:
				oldVal := formatValue(change.OldValue, 0)
				newVal := formatValue(change.NewValue, 0)
				b.WriteString(fmt.Sprintf("~%s: %s → %s\n", change.Path, oldVal, newVal))
			}
		}
	}
	
	return b.String()
}
