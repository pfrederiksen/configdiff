// Package patch provides machine-readable patch format for configuration changes.
//
// The patch format is based on JSON Patch (RFC 6902) with extensions for
// configuration-specific operations.
package patch

import (
	"encoding/json"
	"fmt"

	"github.com/pfrederiksen/configdiff/diff"
	"github.com/pfrederiksen/configdiff/tree"
)

// Patch represents a set of operations to transform one configuration to another.
type Patch struct {
	// Operations is the ordered list of patch operations.
	Operations []Operation `json:"operations"`
}

// Operation represents a single patch operation (JSON Patch-like).
type Operation struct {
	// Op is the operation type: "add", "remove", "replace", "move", "copy", "test"
	Op string `json:"op"`

	// Path is the target path for the operation (JSON Pointer format).
	Path string `json:"path"`

	// Value is the value for add/replace operations.
	Value interface{} `json:"value,omitempty"`

	// From is the source path for move/copy operations.
	From string `json:"from,omitempty"`
}

// FromChanges converts a list of changes into a patch.
func FromChanges(changes []diff.Change) (*Patch, error) {
	ops := make([]Operation, 0, len(changes))

	for _, change := range changes {
		op, err := changeToOperation(change)
		if err != nil {
			return nil, fmt.Errorf("failed to convert change at %s: %w", change.Path, err)
		}
		ops = append(ops, op)
	}

	return &Patch{Operations: ops}, nil
}

// changeToOperation converts a single change to an operation.
func changeToOperation(change diff.Change) (Operation, error) {
	switch change.Type {
	case diff.ChangeTypeAdd:
		value, err := nodeToValue(change.NewValue)
		if err != nil {
			return Operation{}, err
		}
		return Operation{
			Op:    "add",
			Path:  change.Path,
			Value: value,
		}, nil

	case diff.ChangeTypeRemove:
		return Operation{
			Op:   "remove",
			Path: change.Path,
		}, nil

	case diff.ChangeTypeModify:
		value, err := nodeToValue(change.NewValue)
		if err != nil {
			return Operation{}, err
		}
		return Operation{
			Op:    "replace",
			Path:  change.Path,
			Value: value,
		}, nil

	case diff.ChangeTypeMove:
		// For move operations, we need both from and to paths
		// This is currently not fully implemented in the diff engine
		return Operation{
			Op:   "replace",
			Path: change.Path,
		}, nil

	default:
		return Operation{}, fmt.Errorf("unknown change type: %s", change.Type)
	}
}

// nodeToValue converts a tree node to a plain Go value for JSON serialization.
func nodeToValue(node *tree.Node) (interface{}, error) {
	if node == nil {
		return nil, nil
	}

	switch node.Kind {
	case tree.KindNull:
		return nil, nil

	case tree.KindBool, tree.KindNumber, tree.KindString:
		return node.Value, nil

	case tree.KindObject:
		result := make(map[string]interface{})
		for k, v := range node.Object {
			value, err := nodeToValue(v)
			if err != nil {
				return nil, err
			}
			result[k] = value
		}
		return result, nil

	case tree.KindArray:
		result := make([]interface{}, len(node.Array))
		for i, elem := range node.Array {
			value, err := nodeToValue(elem)
			if err != nil {
				return nil, err
			}
			result[i] = value
		}
		return result, nil

	default:
		return nil, fmt.Errorf("unknown node kind: %v", node.Kind)
	}
}

// ToJSON serializes the patch to JSON.
func (p *Patch) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

// ToJSONIndent serializes the patch to indented JSON.
func (p *Patch) ToJSONIndent() ([]byte, error) {
	return json.MarshalIndent(p, "", "  ")
}

// FromJSON deserializes a patch from JSON.
func FromJSON(data []byte) (*Patch, error) {
	var p Patch
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("failed to unmarshal patch: %w", err)
	}
	return &p, nil
}

// IsEmpty returns true if the patch has no operations.
func (p *Patch) IsEmpty() bool {
	return len(p.Operations) == 0
}

// Size returns the number of operations in the patch.
func (p *Patch) Size() int {
	return len(p.Operations)
}

// Summary returns a summary of the patch operations.
func (p *Patch) Summary() map[string]int {
	summary := make(map[string]int)
	for _, op := range p.Operations {
		summary[op.Op]++
	}
	return summary
}

// Validate checks if the patch is valid.
func (p *Patch) Validate() error {
	for i, op := range p.Operations {
		if err := op.Validate(); err != nil {
			return fmt.Errorf("operation %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks if an operation is valid.
func (o *Operation) Validate() error {
	// Check operation type
	validOps := map[string]bool{
		"add":     true,
		"remove":  true,
		"replace": true,
		"move":    true,
		"copy":    true,
		"test":    true,
	}

	if !validOps[o.Op] {
		return fmt.Errorf("invalid operation type: %s", o.Op)
	}

	// Path is required for all operations
	if o.Path == "" {
		return fmt.Errorf("path is required")
	}

	// Validate operation-specific requirements
	switch o.Op {
	case "add", "replace", "test":
		// These operations require a value
		// (value can be nil, but the field should be set)

	case "remove":
		// Remove operations should not have a value or from
		if o.Value != nil {
			return fmt.Errorf("remove operation should not have a value")
		}
		if o.From != "" {
			return fmt.Errorf("remove operation should not have a from field")
		}

	case "move", "copy":
		// These operations require a from path
		if o.From == "" {
			return fmt.Errorf("%s operation requires a from field", o.Op)
		}
	}

	return nil
}
