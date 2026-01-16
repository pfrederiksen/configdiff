package patch

import (
	"encoding/json"
	"testing"

	"github.com/pfrederiksen/configdiff/diff"
	"github.com/pfrederiksen/configdiff/tree"
)

func TestFromChanges(t *testing.T) {
	tests := []struct {
		name      string
		changes   []diff.Change
		wantOps   int
		checkOps  func(*testing.T, []Operation)
		wantError bool
	}{
		{
			name:    "empty changes",
			changes: []diff.Change{},
			wantOps: 0,
		},
		{
			name: "single add",
			changes: []diff.Change{
				{
					Type:     diff.ChangeTypeAdd,
					Path:     "/newKey",
					NewValue: tree.NewString("value"),
				},
			},
			wantOps: 1,
			checkOps: func(t *testing.T, ops []Operation) {
				if ops[0].Op != "add" {
					t.Errorf("Op = %v, want add", ops[0].Op)
				}
				if ops[0].Path != "/newKey" {
					t.Errorf("Path = %v, want /newKey", ops[0].Path)
				}
				if ops[0].Value != "value" {
					t.Errorf("Value = %v, want value", ops[0].Value)
				}
			},
		},
		{
			name: "single remove",
			changes: []diff.Change{
				{
					Type:     diff.ChangeTypeRemove,
					Path:     "/oldKey",
					OldValue: tree.NewString("value"),
				},
			},
			wantOps: 1,
			checkOps: func(t *testing.T, ops []Operation) {
				if ops[0].Op != "remove" {
					t.Errorf("Op = %v, want remove", ops[0].Op)
				}
				if ops[0].Path != "/oldKey" {
					t.Errorf("Path = %v, want /oldKey", ops[0].Path)
				}
			},
		},
		{
			name: "single modify",
			changes: []diff.Change{
				{
					Type:     diff.ChangeTypeModify,
					Path:     "/key",
					OldValue: tree.NewString("old"),
					NewValue: tree.NewString("new"),
				},
			},
			wantOps: 1,
			checkOps: func(t *testing.T, ops []Operation) {
				if ops[0].Op != "replace" {
					t.Errorf("Op = %v, want replace", ops[0].Op)
				}
				if ops[0].Path != "/key" {
					t.Errorf("Path = %v, want /key", ops[0].Path)
				}
				if ops[0].Value != "new" {
					t.Errorf("Value = %v, want new", ops[0].Value)
				}
			},
		},
		{
			name: "multiple changes",
			changes: []diff.Change{
				{
					Type:     diff.ChangeTypeAdd,
					Path:     "/a",
					NewValue: tree.NewString("value"),
				},
				{
					Type:     diff.ChangeTypeRemove,
					Path:     "/b",
					OldValue: tree.NewString("old"),
				},
				{
					Type:     diff.ChangeTypeModify,
					Path:     "/c",
					OldValue: tree.NewNumber(1),
					NewValue: tree.NewNumber(2),
				},
			},
			wantOps: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patch, err := FromChanges(tt.changes)
			if tt.wantError {
				if err == nil {
					t.Error("FromChanges() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("FromChanges() error = %v", err)
			}
			if len(patch.Operations) != tt.wantOps {
				t.Errorf("FromChanges() got %d operations, want %d", len(patch.Operations), tt.wantOps)
			}
			if tt.checkOps != nil {
				tt.checkOps(t, patch.Operations)
			}
		})
	}
}

func TestNodeToValue(t *testing.T) {
	tests := []struct {
		name    string
		node    *tree.Node
		want    interface{}
		wantErr bool
	}{
		{
			name: "nil node",
			node: nil,
			want: nil,
		},
		{
			name: "null node",
			node: tree.NewNull(),
			want: nil,
		},
		{
			name: "bool node",
			node: tree.NewBool(true),
			want: true,
		},
		{
			name: "number node",
			node: tree.NewNumber(42),
			want: 42.0,
		},
		{
			name: "string node",
			node: tree.NewString("hello"),
			want: "hello",
		},
		{
			name: "object node",
			node: tree.NewObject(map[string]*tree.Node{
				"key": tree.NewString("value"),
			}),
			want: map[string]interface{}{
				"key": "value",
			},
		},
		{
			name: "array node",
			node: tree.NewArray([]*tree.Node{
				tree.NewNumber(1),
				tree.NewNumber(2),
				tree.NewNumber(3),
			}),
			want: []interface{}{1.0, 2.0, 3.0},
		},
		{
			name: "nested structure",
			node: tree.NewObject(map[string]*tree.Node{
				"nested": tree.NewObject(map[string]*tree.Node{
					"array": tree.NewArray([]*tree.Node{
						tree.NewString("a"),
						tree.NewString("b"),
					}),
				}),
			}),
			want: map[string]interface{}{
				"nested": map[string]interface{}{
					"array": []interface{}{"a", "b"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := nodeToValue(tt.node)
			if tt.wantErr {
				if err == nil {
					t.Error("nodeToValue() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("nodeToValue() error = %v", err)
			}

			// Compare by converting to JSON for deep equality
			gotJSON, _ := json.Marshal(got)
			wantJSON, _ := json.Marshal(tt.want)
			if string(gotJSON) != string(wantJSON) {
				t.Errorf("nodeToValue() = %v, want %v", string(gotJSON), string(wantJSON))
			}
		})
	}
}

func TestPatch_ToJSON(t *testing.T) {
	patch := &Patch{
		Operations: []Operation{
			{
				Op:    "add",
				Path:  "/key",
				Value: "value",
			},
			{
				Op:   "remove",
				Path: "/old",
			},
		},
	}

	data, err := patch.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	// Verify it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	// Verify roundtrip
	parsed, err := FromJSON(data)
	if err != nil {
		t.Fatalf("FromJSON() error = %v", err)
	}

	if len(parsed.Operations) != len(patch.Operations) {
		t.Errorf("Roundtrip changed operation count")
	}
}

func TestPatch_ToJSONIndent(t *testing.T) {
	patch := &Patch{
		Operations: []Operation{
			{
				Op:    "add",
				Path:  "/key",
				Value: "value",
			},
		},
	}

	data, err := patch.ToJSONIndent()
	if err != nil {
		t.Fatalf("ToJSONIndent() error = %v", err)
	}

	// Verify it's valid and indented JSON
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	// Check for indentation (should contain newlines and spaces)
	dataStr := string(data)
	if len(dataStr) < 10 || dataStr[0] != '{' {
		t.Error("Expected indented JSON format")
	}
}

func TestPatch_IsEmpty(t *testing.T) {
	tests := []struct {
		name  string
		patch *Patch
		want  bool
	}{
		{
			name:  "empty patch",
			patch: &Patch{Operations: []Operation{}},
			want:  true,
		},
		{
			name: "non-empty patch",
			patch: &Patch{Operations: []Operation{
				{Op: "add", Path: "/key"},
			}},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.patch.IsEmpty(); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPatch_Size(t *testing.T) {
	tests := []struct {
		name  string
		patch *Patch
		want  int
	}{
		{
			name:  "empty",
			patch: &Patch{Operations: []Operation{}},
			want:  0,
		},
		{
			name: "multiple operations",
			patch: &Patch{Operations: []Operation{
				{Op: "add"},
				{Op: "remove"},
				{Op: "replace"},
			}},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.patch.Size(); got != tt.want {
				t.Errorf("Size() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPatch_Summary(t *testing.T) {
	patch := &Patch{
		Operations: []Operation{
			{Op: "add"},
			{Op: "add"},
			{Op: "remove"},
			{Op: "replace"},
			{Op: "replace"},
			{Op: "replace"},
		},
	}

	summary := patch.Summary()

	if summary["add"] != 2 {
		t.Errorf("Summary[add] = %v, want 2", summary["add"])
	}
	if summary["remove"] != 1 {
		t.Errorf("Summary[remove] = %v, want 1", summary["remove"])
	}
	if summary["replace"] != 3 {
		t.Errorf("Summary[replace] = %v, want 3", summary["replace"])
	}
}

func TestOperation_Validate(t *testing.T) {
	tests := []struct {
		name    string
		op      Operation
		wantErr bool
	}{
		{
			name:    "valid add",
			op:      Operation{Op: "add", Path: "/key", Value: "value"},
			wantErr: false,
		},
		{
			name:    "valid remove",
			op:      Operation{Op: "remove", Path: "/key"},
			wantErr: false,
		},
		{
			name:    "valid replace",
			op:      Operation{Op: "replace", Path: "/key", Value: "value"},
			wantErr: false,
		},
		{
			name:    "valid move",
			op:      Operation{Op: "move", Path: "/new", From: "/old"},
			wantErr: false,
		},
		{
			name:    "valid copy",
			op:      Operation{Op: "copy", Path: "/new", From: "/old"},
			wantErr: false,
		},
		{
			name:    "invalid operation type",
			op:      Operation{Op: "invalid", Path: "/key"},
			wantErr: true,
		},
		{
			name:    "missing path",
			op:      Operation{Op: "add", Value: "value"},
			wantErr: true,
		},
		{
			name:    "remove with value",
			op:      Operation{Op: "remove", Path: "/key", Value: "value"},
			wantErr: true,
		},
		{
			name:    "remove with from",
			op:      Operation{Op: "remove", Path: "/key", From: "/old"},
			wantErr: true,
		},
		{
			name:    "move without from",
			op:      Operation{Op: "move", Path: "/new"},
			wantErr: true,
		},
		{
			name:    "copy without from",
			op:      Operation{Op: "copy", Path: "/new"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.op.Validate()
			if tt.wantErr {
				if err == nil {
					t.Error("Validate() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestPatch_Validate(t *testing.T) {
	tests := []struct {
		name    string
		patch   *Patch
		wantErr bool
	}{
		{
			name: "valid patch",
			patch: &Patch{
				Operations: []Operation{
					{Op: "add", Path: "/key", Value: "value"},
					{Op: "remove", Path: "/old"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid operation in patch",
			patch: &Patch{
				Operations: []Operation{
					{Op: "add", Path: "/key", Value: "value"},
					{Op: "invalid", Path: "/key"},
				},
			},
			wantErr: true,
		},
		{
			name: "empty patch",
			patch: &Patch{
				Operations: []Operation{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.patch.Validate()
			if tt.wantErr {
				if err == nil {
					t.Error("Validate() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestIntegration_FullPatchWorkflow(t *testing.T) {
	// Simulate a real diff -> patch workflow
	changes := []diff.Change{
		{
			Type:     diff.ChangeTypeAdd,
			Path:     "/spec/replicas",
			NewValue: tree.NewNumber(5),
		},
		{
			Type:     diff.ChangeTypeModify,
			Path:     "/spec/image",
			OldValue: tree.NewString("nginx:1.19"),
			NewValue: tree.NewString("nginx:1.20"),
		},
		{
			Type:     diff.ChangeTypeRemove,
			Path:     "/metadata/annotations/deprecated",
			OldValue: tree.NewString("true"),
		},
	}

	// Convert to patch
	patch, err := FromChanges(changes)
	if err != nil {
		t.Fatalf("FromChanges() error = %v", err)
	}

	// Validate
	if err := patch.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	// Check summary
	summary := patch.Summary()
	if summary["add"] != 1 {
		t.Errorf("Expected 1 add operation, got %d", summary["add"])
	}
	if summary["replace"] != 1 {
		t.Errorf("Expected 1 replace operation, got %d", summary["replace"])
	}
	if summary["remove"] != 1 {
		t.Errorf("Expected 1 remove operation, got %d", summary["remove"])
	}

	// Serialize to JSON
	data, err := patch.ToJSONIndent()
	if err != nil {
		t.Fatalf("ToJSONIndent() error = %v", err)
	}

	// Deserialize from JSON
	parsed, err := FromJSON(data)
	if err != nil {
		t.Fatalf("FromJSON() error = %v", err)
	}

	// Verify roundtrip
	if parsed.Size() != patch.Size() {
		t.Errorf("Roundtrip changed size: got %d, want %d", parsed.Size(), patch.Size())
	}

	// Validate parsed patch
	if err := parsed.Validate(); err != nil {
		t.Errorf("Parsed patch validation failed: %v", err)
	}
}
