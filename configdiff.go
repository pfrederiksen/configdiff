// Package configdiff provides semantic, human-grade diffs for YAML/JSON/HCL configuration files.
//
// It parses configuration files into a normalized tree representation, applies customizable
// diff rules, and generates both machine-readable patches and human-friendly reports.
package configdiff

import (
	"fmt"

	"github.com/pfrederiksen/configdiff/diff"
	"github.com/pfrederiksen/configdiff/parse"
	"github.com/pfrederiksen/configdiff/tree"
)

// Re-export types from diff package for convenience.
type (
	// Options configures how diffs are computed.
	Options = diff.Options

	// Coercions defines rules for type coercion during comparison.
	Coercions = diff.Coercions

	// Change represents a single detected change.
	Change = diff.Change

	// ChangeType categorizes the kind of change.
	ChangeType = diff.ChangeType
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
	Patch Patch

	// Report is the human-friendly pretty report.
	Report string
}

// Patch represents a machine-readable set of operations.
type Patch struct {
	// Operations is the list of patch operations.
	Operations []Operation
}

// Operation is a single patch operation (JSON Patch-like).
type Operation struct {
	// Op is the operation type (add, remove, replace, move).
	Op string `json:"op"`

	// Path is the target path for the operation.
	Path string `json:"path"`

	// Value is the value for add/replace operations.
	Value interface{} `json:"value,omitempty"`

	// From is the source path for move operations.
	From string `json:"from,omitempty"`
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
	bTree, err := parse.Parse(b, parse.Format(bFormat))
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

	// Build result
	result := &Result{
		Changes: changes,
		Patch:   Patch{}, // TODO: implement patch generation
		Report:  "",      // TODO: implement report generation
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
