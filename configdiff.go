// Package configdiff provides semantic, human-grade diffs for YAML/JSON/HCL configuration files.
//
// It parses configuration files into a normalized tree representation, applies customizable
// diff rules, and generates both machine-readable patches and human-friendly reports.
package configdiff

import (
	"fmt"

	"github.com/pfrederiksen/configdiff/diff"
	"github.com/pfrederiksen/configdiff/parse"
	"github.com/pfrederiksen/configdiff/patch"
	"github.com/pfrederiksen/configdiff/report"
	"github.com/pfrederiksen/configdiff/tree"
)

// Re-export types from diff and patch packages for convenience.
type (
	// Options configures how diffs are computed.
	Options = diff.Options

	// Coercions defines rules for type coercion during comparison.
	Coercions = diff.Coercions

	// Change represents a single detected change.
	Change = diff.Change

	// ChangeType categorizes the kind of change.
	ChangeType = diff.ChangeType

	// Patch represents a machine-readable set of operations.
	Patch = patch.Patch

	// Operation is a single patch operation (JSON Patch-like).
	Operation = patch.Operation
)

// Re-export change type constants.
const (
	// ChangeTypeAdd indicates a new value was added.
	ChangeTypeAdd = diff.ChangeTypeAdd

	// ChangeTypeRemove indicates a value was removed.
	ChangeTypeRemove = diff.ChangeTypeRemove

	// ChangeTypeModify indicates a value was changed.
	ChangeTypeModify = diff.ChangeTypeModify

	// ChangeTypeMove indicates a value was moved (array reordering).
	ChangeTypeMove = diff.ChangeTypeMove
)

// Result contains the output of a diff operation.
type Result struct {
	// Changes is the list of detected changes.
	Changes []Change

	// Patch is the machine-readable patch representation.
	Patch *Patch

	// Report is the human-friendly pretty report.
	Report string
}

// DiffBytes compares two configuration byte slices and returns the diff result.
//
// Supported formats: "yaml", "json", "hcl"
func DiffBytes(a []byte, aFormat string, b []byte, bFormat string, opts Options) (*Result, error) {
	// Parse format a
	aTree, err := parse.Parse(a, parse.Format(aFormat))
	if err != nil {
		return nil, fmt.Errorf("failed to parse format %s: %w", aFormat, err)
	}

	// Parse format b
	var bTree *tree.Node
	bTree, err = parse.Parse(b, parse.Format(bFormat))
	if err != nil {
		return nil, fmt.Errorf("failed to parse format %s: %w", bFormat, err)
	}

	return DiffTrees(aTree, bTree, opts)
}

// DiffTrees compares two normalized tree nodes and returns the diff result.
func DiffTrees(a, b *tree.Node, opts Options) (*Result, error) {
	// Compute the diff
	changes, err := diff.Diff(a, b, opts)
	if err != nil {
		return nil, fmt.Errorf("diff failed: %w", err)
	}

	// Generate patch from changes
	var patchObj *patch.Patch
	patchObj, err = patch.FromChanges(changes)
	if err != nil {
		return nil, fmt.Errorf("patch generation failed: %w", err)
	}

	// Generate pretty report
	reportText := report.GenerateDetailed(changes)

	// Build result
	result := &Result{
		Changes: changes,
		Patch:   patchObj,
		Report:  reportText,
	}

	return result, nil
}

// DiffYAML is a convenience function for comparing two YAML byte slices.
func DiffYAML(a, b []byte, opts Options) (*Result, error) {
	return DiffBytes(a, "yaml", b, "yaml", opts)
}

// DiffJSON is a convenience function for comparing two JSON byte slices.
func DiffJSON(a, b []byte, opts Options) (*Result, error) {
	return DiffBytes(a, "json", b, "json", opts)
}
