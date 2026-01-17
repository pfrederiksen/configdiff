package report

import (
	"fmt"
	"strings"

	"github.com/pfrederiksen/configdiff/diff"
)

// GenerateStat creates a statistics summary similar to git diff --stat.
func GenerateStat(changes []diff.Change) string {
	if len(changes) == 0 {
		return "No changes detected.\n"
	}

	summary := summarizeChanges(changes)
	
	var b strings.Builder
	
	// Count affected paths
	paths := make(map[string]*pathStat)
	for _, change := range changes {
		path := change.Path
		if paths[path] == nil {
			paths[path] = &pathStat{}
		}
		
		switch change.Type {
		case diff.ChangeTypeAdd:
			paths[path].additions++
		case diff.ChangeTypeRemove:
			paths[path].deletions++
		case diff.ChangeTypeModify:
			paths[path].modifications++
		case diff.ChangeTypeMove:
			paths[path].moves++
		}
	}
	
	// Sort paths for stable output
	sortedPaths := make([]string, 0, len(paths))
	for path := range paths {
		sortedPaths = append(sortedPaths, path)
	}
	
	// Simple sort
	for i := 0; i < len(sortedPaths); i++ {
		for j := i + 1; j < len(sortedPaths); j++ {
			if sortedPaths[i] > sortedPaths[j] {
				sortedPaths[i], sortedPaths[j] = sortedPaths[j], sortedPaths[i]
			}
		}
	}
	
	// Print per-path statistics
	maxPathLen := 0
	for _, path := range sortedPaths {
		if len(path) > maxPathLen {
			maxPathLen = len(path)
		}
	}
	
	if maxPathLen > 60 {
		maxPathLen = 60
	}
	
	for _, path := range sortedPaths {
		stat := paths[path]
		displayPath := path
		if len(displayPath) > 60 {
			displayPath = "..." + displayPath[len(displayPath)-57:]
		}
		
		// Calculate total changes for this path
		total := stat.additions + stat.deletions + stat.modifications + stat.moves
		
		// Create visual bar
		barWidth := 40
		var bar string
		if total > 0 {
			plusCount := (stat.additions * barWidth) / total
			minusCount := (stat.deletions * barWidth) / total
			modCount := (stat.modifications * barWidth) / total
			
			bar = strings.Repeat("+", plusCount) + 
			      strings.Repeat("-", minusCount) + 
			      strings.Repeat("~", modCount)
			
			if len(bar) > barWidth {
				bar = bar[:barWidth]
			}
		}
		
		fmt.Fprintf(&b, " %-*s | %s\n", maxPathLen, displayPath, bar)
	}
	
	// Print summary
	b.WriteString(fmt.Sprintf(" %d paths changed", len(paths)))
	if summary.Added > 0 {
		b.WriteString(fmt.Sprintf(", %d additions(+)", summary.Added))
	}
	if summary.Removed > 0 {
		b.WriteString(fmt.Sprintf(", %d deletions(-)", summary.Removed))
	}
	if summary.Modified > 0 {
		b.WriteString(fmt.Sprintf(", %d modifications(~)", summary.Modified))
	}
	if summary.Moved > 0 {
		b.WriteString(fmt.Sprintf(", %d moves(→)", summary.Moved))
	}
	b.WriteString("\n")
	
	return b.String()
}

type pathStat struct {
	additions     int
	deletions     int
	modifications int
	moves         int
}
