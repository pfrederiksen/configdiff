// Package diff provides semantic diffing for normalized configuration trees.
package diff

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/pfrederiksen/configdiff/tree"
)

// Change represents a single detected change in the diff.
type Change struct {
	// Type is the kind of change.
	Type ChangeType

	// Path is the location of the change.
	Path string

	// OldValue is the previous value (nil for additions).
	OldValue *tree.Node

	// NewValue is the new value (nil for removals).
	NewValue *tree.Node

	// ArrayIndex is set for array element changes (optional).
	ArrayIndex int
}

// ChangeType categorizes the kind of change.
type ChangeType string

const (
	// ChangeTypeAdd indicates a value was added.
	ChangeTypeAdd ChangeType = "add"

	// ChangeTypeRemove indicates a value was removed.
	ChangeTypeRemove ChangeType = "remove"

	// ChangeTypeModify indicates a value was changed.
	ChangeTypeModify ChangeType = "modify"

	// ChangeTypeMove indicates a value was moved (array reordering).
	ChangeTypeMove ChangeType = "move"
)

// Options configures how diffs are computed.
type Options struct {
	// IgnorePaths specifies paths to ignore in the diff.
	// Supports glob-like patterns with wildcards (*).
	// Example: []string{"metadata.creationTimestamp", "status.*"}
	IgnorePaths []string

	// ArraySetKeys maps array paths to their key field names.
	// Arrays at these paths are treated as sets keyed by the specified field.
	// Example: map[string]string{"/spec/containers": "name"}
	ArraySetKeys map[string]string

	// Coercions configures type coercion rules.
	Coercions Coercions

	// StableOrder ensures deterministic ordering in output.
	StableOrder bool
}

// Coercions defines rules for type coercion during comparison.
type Coercions struct {
	// NumericStrings allows comparing string numbers with numeric values.
	// Example: "42" can equal 42
	NumericStrings bool

	// BoolStrings allows comparing string booleans with boolean values.
	// Example: "true" can equal true
	BoolStrings bool
}

// Diff compares two trees and returns the detected changes.
func Diff(a, b *tree.Node, opts Options) ([]Change, error) {
	d := &differ{
		opts:    opts,
		changes: make([]Change, 0),
	}

	d.diffNodes(a, b, "/")

	if opts.StableOrder {
		sort.Slice(d.changes, func(i, j int) bool {
			return d.changes[i].Path < d.changes[j].Path
		})
	}

	return d.changes, nil
}

// differ holds state during diff operation.
type differ struct {
	opts    Options
	changes []Change
}

// diffNodes compares two nodes at a given path.
func (d *differ) diffNodes(a, b *tree.Node, path string) {
	// Check if path should be ignored
	if d.shouldIgnore(path) {
		return
	}

	// Handle nil cases
	if a == nil && b == nil {
		return
	}
	if a == nil {
		d.addChange(Change{
			Type:     ChangeTypeAdd,
			Path:     path,
			NewValue: b,
		})
		return
	}
	if b == nil {
		d.addChange(Change{
			Type:     ChangeTypeRemove,
			Path:     path,
			OldValue: a,
		})
		return
	}

	// Try coercion if types differ
	if a.Kind != b.Kind {
		if d.canCoerce(a, b) {
			return // Values are equal after coercion
		}
		d.addChange(Change{
			Type:     ChangeTypeModify,
			Path:     path,
			OldValue: a,
			NewValue: b,
		})
		return
	}

	// Compare based on node kind
	switch a.Kind {
	case tree.KindNull:
		// Both null, no change
		return

	case tree.KindBool, tree.KindNumber, tree.KindString:
		if a.Value != b.Value {
			d.addChange(Change{
				Type:     ChangeTypeModify,
				Path:     path,
				OldValue: a,
				NewValue: b,
			})
		}

	case tree.KindObject:
		d.diffObjects(a, b, path)

	case tree.KindArray:
		d.diffArrays(a, b, path)
	}
}

// diffObjects compares two object nodes.
func (d *differ) diffObjects(a, b *tree.Node, path string) {
	allKeys := make(map[string]bool)
	for k := range a.Object {
		allKeys[k] = true
	}
	for k := range b.Object {
		allKeys[k] = true
	}

	keys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		keys = append(keys, k)
	}

	if d.opts.StableOrder {
		sort.Strings(keys)
	}

	for _, key := range keys {
		childPath := joinPath(path, key)
		aVal, aExists := a.Object[key]
		bVal, bExists := b.Object[key]

		if !aExists {
			d.diffNodes(nil, bVal, childPath)
		} else if !bExists {
			d.diffNodes(aVal, nil, childPath)
		} else {
			d.diffNodes(aVal, bVal, childPath)
		}
	}
}

// diffArrays compares two array nodes.
func (d *differ) diffArrays(a, b *tree.Node, path string) {
	// Check if this array should be treated as a set
	keyField, isSet := d.opts.ArraySetKeys[path]
	if isSet {
		d.diffArrayAsSet(a, b, path, keyField)
		return
	}

	// Positional array comparison
	maxLen := len(a.Array)
	if len(b.Array) > maxLen {
		maxLen = len(b.Array)
	}

	for i := 0; i < maxLen; i++ {
		childPath := fmt.Sprintf("%s[%d]", path, i)
		var aElem, bElem *tree.Node
		if i < len(a.Array) {
			aElem = a.Array[i]
		}
		if i < len(b.Array) {
			bElem = b.Array[i]
		}
		d.diffNodes(aElem, bElem, childPath)
	}
}

// diffArrayAsSet compares arrays as sets keyed by a field.
func (d *differ) diffArrayAsSet(a, b *tree.Node, path, keyField string) {
	// Build maps of elements by key
	aMap := make(map[string]*tree.Node)
	bMap := make(map[string]*tree.Node)

	for _, elem := range a.Array {
		if key := d.extractKey(elem, keyField); key != "" {
			aMap[key] = elem
		}
	}

	for _, elem := range b.Array {
		if key := d.extractKey(elem, keyField); key != "" {
			bMap[key] = elem
		}
	}

	// Find all unique keys
	allKeys := make(map[string]bool)
	for k := range aMap {
		allKeys[k] = true
	}
	for k := range bMap {
		allKeys[k] = true
	}

	keys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		keys = append(keys, k)
	}

	if d.opts.StableOrder {
		sort.Strings(keys)
	}

	// Compare elements by key
	for _, key := range keys {
		childPath := fmt.Sprintf("%s[%s=%s]", path, keyField, key)
		aElem, aExists := aMap[key]
		bElem, bExists := bMap[key]

		if !aExists {
			d.diffNodes(nil, bElem, childPath)
		} else if !bExists {
			d.diffNodes(aElem, nil, childPath)
		} else {
			d.diffNodes(aElem, bElem, childPath)
		}
	}
}

// extractKey extracts the key field value from an object node.
func (d *differ) extractKey(node *tree.Node, keyField string) string {
	if node.Kind != tree.KindObject {
		return ""
	}
	keyNode, exists := node.Object[keyField]
	if !exists || keyNode.Kind != tree.KindString {
		return ""
	}
	return keyNode.Value.(string)
}

// canCoerce checks if two nodes can be considered equal via coercion.
func (d *differ) canCoerce(a, b *tree.Node) bool {
	// Numeric string coercion
	if d.opts.Coercions.NumericStrings {
		if a.Kind == tree.KindString && b.Kind == tree.KindNumber {
			if num, err := strconv.ParseFloat(a.Value.(string), 64); err == nil {
				return num == b.Value.(float64)
			}
		}
		if a.Kind == tree.KindNumber && b.Kind == tree.KindString {
			if num, err := strconv.ParseFloat(b.Value.(string), 64); err == nil {
				return a.Value.(float64) == num
			}
		}
	}

	// Bool string coercion
	if d.opts.Coercions.BoolStrings {
		if a.Kind == tree.KindString && b.Kind == tree.KindBool {
			if a.Value.(string) == "true" {
				return b.Value.(bool) == true
			}
			if a.Value.(string) == "false" {
				return b.Value.(bool) == false
			}
		}
		if a.Kind == tree.KindBool && b.Kind == tree.KindString {
			if b.Value.(string) == "true" {
				return a.Value.(bool) == true
			}
			if b.Value.(string) == "false" {
				return a.Value.(bool) == false
			}
		}
	}

	return false
}

// shouldIgnore checks if a path should be ignored.
func (d *differ) shouldIgnore(path string) bool {
	for _, pattern := range d.opts.IgnorePaths {
		if matchPath(path, pattern) {
			return true
		}
	}
	return false
}

// matchPath checks if a path matches a pattern (supports * wildcard).
func matchPath(path, pattern string) bool {
	// Simple implementation: support * as wildcard
	// Convert pattern to segments
	pathSegments := strings.Split(strings.Trim(path, "/"), "/")
	patternSegments := strings.Split(strings.Trim(pattern, "/"), "/")

	// If pattern has no wildcards, do exact match
	if !strings.Contains(pattern, "*") {
		return "/" + strings.Join(pathSegments, "/") == "/" + strings.Join(patternSegments, "/")
	}

	// Check if pattern matches
	return matchSegments(pathSegments, patternSegments)
}

// matchSegments checks if path segments match pattern segments.
func matchSegments(pathSegs, patternSegs []string) bool {
	if len(patternSegs) == 0 {
		return len(pathSegs) == 0
	}

	if len(pathSegs) == 0 {
		// Check if all remaining pattern segments are wildcards
		for _, p := range patternSegs {
			if p != "*" {
				return false
			}
		}
		return true
	}

	pattern := patternSegs[0]
	if pattern == "*" {
		// Wildcard can match:
		// 1. Nothing (move to next pattern segment)
		// 2. One or more path segments (consume path segments)
		if len(patternSegs) == 1 {
			// Last pattern segment is *, matches everything remaining
			return true
		}
		// Try matching the rest of the pattern at this position or later
		return matchSegments(pathSegs[1:], patternSegs[1:]) ||
			matchSegments(pathSegs[1:], patternSegs)
	}

	// Check if first segments match
	if pathSegs[0] != pattern {
		return false
	}

	return matchSegments(pathSegs[1:], patternSegs[1:])
}

// addChange adds a change to the list.
func (d *differ) addChange(c Change) {
	d.changes = append(d.changes, c)
}

// joinPath joins path segments.
func joinPath(base, key string) string {
	if base == "/" {
		return "/" + key
	}
	return base + "/" + key
}
