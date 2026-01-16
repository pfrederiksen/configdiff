package diff

import (
	"testing"

	"github.com/pfrederiksen/configdiff/tree"
)

func TestDiff_Scalars(t *testing.T) {
	tests := []struct {
		name       string
		a          *tree.Node
		b          *tree.Node
		opts       Options
		wantCount  int
		wantType   ChangeType
		checkFirst func(*testing.T, Change)
	}{
		{
			name:      "no change - null",
			a:         tree.NewNull(),
			b:         tree.NewNull(),
			wantCount: 0,
		},
		{
			name:      "no change - bool",
			a:         tree.NewBool(true),
			b:         tree.NewBool(true),
			wantCount: 0,
		},
		{
			name:      "no change - number",
			a:         tree.NewNumber(42),
			b:         tree.NewNumber(42),
			wantCount: 0,
		},
		{
			name:      "no change - string",
			a:         tree.NewString("hello"),
			b:         tree.NewString("hello"),
			wantCount: 0,
		},
		{
			name:      "modify - bool changed",
			a:         tree.NewBool(true),
			b:         tree.NewBool(false),
			wantCount: 1,
			wantType:  ChangeTypeModify,
		},
		{
			name:      "modify - number changed",
			a:         tree.NewNumber(42),
			b:         tree.NewNumber(43),
			wantCount: 1,
			wantType:  ChangeTypeModify,
		},
		{
			name:      "modify - string changed",
			a:         tree.NewString("old"),
			b:         tree.NewString("new"),
			wantCount: 1,
			wantType:  ChangeTypeModify,
		},
		{
			name:      "add - nil to value",
			a:         nil,
			b:         tree.NewString("added"),
			wantCount: 1,
			wantType:  ChangeTypeAdd,
		},
		{
			name:      "remove - value to nil",
			a:         tree.NewString("removed"),
			b:         nil,
			wantCount: 1,
			wantType:  ChangeTypeRemove,
		},
		{
			name:      "modify - type changed",
			a:         tree.NewString("42"),
			b:         tree.NewNumber(42),
			wantCount: 1,
			wantType:  ChangeTypeModify,
		},
		{
			name: "no change - numeric string coercion",
			a:    tree.NewString("42"),
			b:    tree.NewNumber(42),
			opts: Options{
				Coercions: Coercions{NumericStrings: true},
			},
			wantCount: 0,
		},
		{
			name: "no change - bool string coercion",
			a:    tree.NewString("true"),
			b:    tree.NewBool(true),
			opts: Options{
				Coercions: Coercions{BoolStrings: true},
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes, err := Diff(tt.a, tt.b, tt.opts)
			if err != nil {
				t.Fatalf("Diff() error = %v", err)
			}
			if len(changes) != tt.wantCount {
				t.Errorf("Diff() got %d changes, want %d", len(changes), tt.wantCount)
			}
			if tt.wantCount > 0 && changes[0].Type != tt.wantType {
				t.Errorf("Change type = %v, want %v", changes[0].Type, tt.wantType)
			}
			if tt.checkFirst != nil && len(changes) > 0 {
				tt.checkFirst(t, changes[0])
			}
		})
	}
}

func TestDiff_Objects(t *testing.T) {
	tests := []struct {
		name      string
		a         *tree.Node
		b         *tree.Node
		opts      Options
		wantCount int
		check     func(*testing.T, []Change)
	}{
		{
			name: "no change - empty objects",
			a:    tree.NewObject(map[string]*tree.Node{}),
			b:    tree.NewObject(map[string]*tree.Node{}),
			wantCount: 0,
		},
		{
			name: "no change - identical objects",
			a: tree.NewObject(map[string]*tree.Node{
				"key": tree.NewString("value"),
			}),
			b: tree.NewObject(map[string]*tree.Node{
				"key": tree.NewString("value"),
			}),
			wantCount: 0,
		},
		{
			name: "add - new key",
			a:    tree.NewObject(map[string]*tree.Node{}),
			b: tree.NewObject(map[string]*tree.Node{
				"newKey": tree.NewString("value"),
			}),
			wantCount: 1,
			check: func(t *testing.T, changes []Change) {
				if changes[0].Type != ChangeTypeAdd {
					t.Errorf("Type = %v, want %v", changes[0].Type, ChangeTypeAdd)
				}
				if changes[0].Path != "/newKey" {
					t.Errorf("Path = %v, want /newKey", changes[0].Path)
				}
			},
		},
		{
			name: "remove - deleted key",
			a: tree.NewObject(map[string]*tree.Node{
				"oldKey": tree.NewString("value"),
			}),
			b:         tree.NewObject(map[string]*tree.Node{}),
			wantCount: 1,
			check: func(t *testing.T, changes []Change) {
				if changes[0].Type != ChangeTypeRemove {
					t.Errorf("Type = %v, want %v", changes[0].Type, ChangeTypeRemove)
				}
				if changes[0].Path != "/oldKey" {
					t.Errorf("Path = %v, want /oldKey", changes[0].Path)
				}
			},
		},
		{
			name: "modify - value changed",
			a: tree.NewObject(map[string]*tree.Node{
				"key": tree.NewString("old"),
			}),
			b: tree.NewObject(map[string]*tree.Node{
				"key": tree.NewString("new"),
			}),
			wantCount: 1,
			check: func(t *testing.T, changes []Change) {
				if changes[0].Type != ChangeTypeModify {
					t.Errorf("Type = %v, want %v", changes[0].Type, ChangeTypeModify)
				}
			},
		},
		{
			name: "nested objects",
			a: tree.NewObject(map[string]*tree.Node{
				"parent": tree.NewObject(map[string]*tree.Node{
					"child": tree.NewString("old"),
				}),
			}),
			b: tree.NewObject(map[string]*tree.Node{
				"parent": tree.NewObject(map[string]*tree.Node{
					"child": tree.NewString("new"),
				}),
			}),
			wantCount: 1,
			check: func(t *testing.T, changes []Change) {
				if changes[0].Path != "/parent/child" {
					t.Errorf("Path = %v, want /parent/child", changes[0].Path)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes, err := Diff(tt.a, tt.b, tt.opts)
			if err != nil {
				t.Fatalf("Diff() error = %v", err)
			}
			if len(changes) != tt.wantCount {
				t.Errorf("Diff() got %d changes, want %d", len(changes), tt.wantCount)
			}
			if tt.check != nil {
				tt.check(t, changes)
			}
		})
	}
}

func TestDiff_Arrays(t *testing.T) {
	tests := []struct {
		name      string
		a         *tree.Node
		b         *tree.Node
		opts      Options
		wantCount int
		check     func(*testing.T, []Change)
	}{
		{
			name:      "no change - empty arrays",
			a:         tree.NewArray([]*tree.Node{}),
			b:         tree.NewArray([]*tree.Node{}),
			wantCount: 0,
		},
		{
			name: "no change - identical arrays",
			a: tree.NewArray([]*tree.Node{
				tree.NewString("a"),
				tree.NewString("b"),
			}),
			b: tree.NewArray([]*tree.Node{
				tree.NewString("a"),
				tree.NewString("b"),
			}),
			wantCount: 0,
		},
		{
			name: "positional - element added",
			a: tree.NewArray([]*tree.Node{
				tree.NewString("a"),
			}),
			b: tree.NewArray([]*tree.Node{
				tree.NewString("a"),
				tree.NewString("b"),
			}),
			wantCount: 1,
			check: func(t *testing.T, changes []Change) {
				if changes[0].Type != ChangeTypeAdd {
					t.Errorf("Type = %v, want %v", changes[0].Type, ChangeTypeAdd)
				}
				if changes[0].Path != "/[1]" {
					t.Errorf("Path = %v, want /[1]", changes[0].Path)
				}
			},
		},
		{
			name: "positional - element removed",
			a: tree.NewArray([]*tree.Node{
				tree.NewString("a"),
				tree.NewString("b"),
			}),
			b: tree.NewArray([]*tree.Node{
				tree.NewString("a"),
			}),
			wantCount: 1,
			check: func(t *testing.T, changes []Change) {
				if changes[0].Type != ChangeTypeRemove {
					t.Errorf("Type = %v, want %v", changes[0].Type, ChangeTypeRemove)
				}
			},
		},
		{
			name: "positional - element changed",
			a: tree.NewArray([]*tree.Node{
				tree.NewString("a"),
				tree.NewString("b"),
			}),
			b: tree.NewArray([]*tree.Node{
				tree.NewString("a"),
				tree.NewString("c"),
			}),
			wantCount: 1,
			check: func(t *testing.T, changes []Change) {
				if changes[0].Type != ChangeTypeModify {
					t.Errorf("Type = %v, want %v", changes[0].Type, ChangeTypeModify)
				}
			},
		},
		{
			name: "set mode - element added by key",
			a: tree.NewArray([]*tree.Node{
				tree.NewObject(map[string]*tree.Node{
					"name": tree.NewString("item1"),
					"value": tree.NewNumber(1),
				}),
			}),
			b: tree.NewArray([]*tree.Node{
				tree.NewObject(map[string]*tree.Node{
					"name": tree.NewString("item1"),
					"value": tree.NewNumber(1),
				}),
				tree.NewObject(map[string]*tree.Node{
					"name": tree.NewString("item2"),
					"value": tree.NewNumber(2),
				}),
			}),
			opts: Options{
				ArraySetKeys: map[string]string{"/": "name"},
			},
			wantCount: 1,
			check: func(t *testing.T, changes []Change) {
				if changes[0].Type != ChangeTypeAdd {
					t.Errorf("Type = %v, want %v", changes[0].Type, ChangeTypeAdd)
				}
				if changes[0].Path != "/[name=item2]" {
					t.Errorf("Path = %v, want /[name=item2]", changes[0].Path)
				}
			},
		},
		{
			name: "set mode - element modified by key",
			a: tree.NewArray([]*tree.Node{
				tree.NewObject(map[string]*tree.Node{
					"name": tree.NewString("item1"),
					"value": tree.NewNumber(1),
				}),
			}),
			b: tree.NewArray([]*tree.Node{
				tree.NewObject(map[string]*tree.Node{
					"name": tree.NewString("item1"),
					"value": tree.NewNumber(2),
				}),
			}),
			opts: Options{
				ArraySetKeys: map[string]string{"/": "name"},
			},
			wantCount: 1,
			check: func(t *testing.T, changes []Change) {
				if changes[0].Type != ChangeTypeModify {
					t.Errorf("Type = %v, want %v", changes[0].Type, ChangeTypeModify)
				}
				if changes[0].Path != "/[name=item1]/value" {
					t.Errorf("Path = %v, want /[name=item1]/value", changes[0].Path)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes, err := Diff(tt.a, tt.b, tt.opts)
			if err != nil {
				t.Fatalf("Diff() error = %v", err)
			}
			if len(changes) != tt.wantCount {
				t.Errorf("Diff() got %d changes, want %d", len(changes), tt.wantCount)
			}
			if tt.check != nil {
				tt.check(t, changes)
			}
		})
	}
}

func TestDiff_IgnorePaths(t *testing.T) {
	tests := []struct {
		name      string
		a         *tree.Node
		b         *tree.Node
		opts      Options
		wantCount int
	}{
		{
			name: "ignore exact path",
			a: tree.NewObject(map[string]*tree.Node{
				"keep":   tree.NewString("old"),
				"ignore": tree.NewString("old"),
			}),
			b: tree.NewObject(map[string]*tree.Node{
				"keep":   tree.NewString("new"),
				"ignore": tree.NewString("new"),
			}),
			opts: Options{
				IgnorePaths: []string{"/ignore"},
			},
			wantCount: 1, // Only /keep changed
		},
		{
			name: "ignore with wildcard",
			a: tree.NewObject(map[string]*tree.Node{
				"metadata": tree.NewObject(map[string]*tree.Node{
					"timestamp": tree.NewString("old"),
					"name":      tree.NewString("old"),
				}),
				"data": tree.NewString("old"),
			}),
			b: tree.NewObject(map[string]*tree.Node{
				"metadata": tree.NewObject(map[string]*tree.Node{
					"timestamp": tree.NewString("new"),
					"name":      tree.NewString("new"),
				}),
				"data": tree.NewString("new"),
			}),
			opts: Options{
				IgnorePaths: []string{"/metadata/*"},
			},
			wantCount: 1, // Only /data changed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes, err := Diff(tt.a, tt.b, tt.opts)
			if err != nil {
				t.Fatalf("Diff() error = %v", err)
			}
			if len(changes) != tt.wantCount {
				t.Errorf("Diff() got %d changes, want %d", len(changes), tt.wantCount)
				for _, c := range changes {
					t.Logf("  Change: %s %s", c.Type, c.Path)
				}
			}
		})
	}
}

func TestDiff_StableOrder(t *testing.T) {
	a := tree.NewObject(map[string]*tree.Node{
		"z": tree.NewString("old"),
		"a": tree.NewString("old"),
		"m": tree.NewString("old"),
	})
	b := tree.NewObject(map[string]*tree.Node{
		"z": tree.NewString("new"),
		"a": tree.NewString("new"),
		"m": tree.NewString("new"),
	})

	changes, err := Diff(a, b, Options{StableOrder: true})
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}

	if len(changes) != 3 {
		t.Fatalf("Diff() got %d changes, want 3", len(changes))
	}

	// Changes should be in alphabetical order
	expectedPaths := []string{"/a", "/m", "/z"}
	for i, want := range expectedPaths {
		if changes[i].Path != want {
			t.Errorf("changes[%d].Path = %v, want %v", i, changes[i].Path, want)
		}
	}
}

func TestDiff_ComplexExample(t *testing.T) {
	// Kubernetes-like deployment diff
	a := tree.NewObject(map[string]*tree.Node{
		"apiVersion": tree.NewString("apps/v1"),
		"kind":       tree.NewString("Deployment"),
		"metadata": tree.NewObject(map[string]*tree.Node{
			"name":      tree.NewString("myapp"),
			"timestamp": tree.NewString("2024-01-01"),
		}),
		"spec": tree.NewObject(map[string]*tree.Node{
			"replicas": tree.NewNumber(3),
			"containers": tree.NewArray([]*tree.Node{
				tree.NewObject(map[string]*tree.Node{
					"name":  tree.NewString("nginx"),
					"image": tree.NewString("nginx:1.19"),
				}),
				tree.NewObject(map[string]*tree.Node{
					"name":  tree.NewString("sidecar"),
					"image": tree.NewString("sidecar:v1"),
				}),
			}),
		}),
	})

	b := tree.NewObject(map[string]*tree.Node{
		"apiVersion": tree.NewString("apps/v1"),
		"kind":       tree.NewString("Deployment"),
		"metadata": tree.NewObject(map[string]*tree.Node{
			"name":      tree.NewString("myapp"),
			"timestamp": tree.NewString("2024-01-02"), // Changed but ignored
		}),
		"spec": tree.NewObject(map[string]*tree.Node{
			"replicas": tree.NewNumber(5), // Changed: 3 -> 5
			"containers": tree.NewArray([]*tree.Node{
				tree.NewObject(map[string]*tree.Node{
					"name":  tree.NewString("nginx"),
					"image": tree.NewString("nginx:1.20"), // Changed: 1.19 -> 1.20
				}),
				tree.NewObject(map[string]*tree.Node{
					"name":  tree.NewString("sidecar"),
					"image": tree.NewString("sidecar:v1"),
				}),
			}),
		}),
	})

	opts := Options{
		IgnorePaths:  []string{"/metadata/timestamp"},
		ArraySetKeys: map[string]string{"/spec/containers": "name"},
		StableOrder:  true,
	}

	changes, err := Diff(a, b, opts)
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}

	// Should have 2 changes: replicas and nginx image
	if len(changes) != 2 {
		t.Errorf("Diff() got %d changes, want 2", len(changes))
		for _, c := range changes {
			t.Logf("  Change: %s %s", c.Type, c.Path)
		}
	}

	// Verify specific changes
	foundReplicas := false
	foundImage := false

	for _, c := range changes {
		if c.Path == "/spec/replicas" && c.Type == ChangeTypeModify {
			foundReplicas = true
			if c.OldValue.Value != 3.0 || c.NewValue.Value != 5.0 {
				t.Errorf("Replicas change values incorrect")
			}
		}
		if c.Path == "/spec/containers[name=nginx]/image" && c.Type == ChangeTypeModify {
			foundImage = true
			if c.OldValue.Value != "nginx:1.19" || c.NewValue.Value != "nginx:1.20" {
				t.Errorf("Image change values incorrect")
			}
		}
	}

	if !foundReplicas {
		t.Error("Did not find replicas change")
	}
	if !foundImage {
		t.Error("Did not find image change")
	}
}

func TestMatchPath(t *testing.T) {
	tests := []struct {
		path    string
		pattern string
		want    bool
	}{
		{"/exact/path", "/exact/path", true},
		{"/exact/path", "/exact/other", false},
		{"/metadata/timestamp", "/metadata/*", true},
		{"/metadata/name", "/metadata/*", true},
		{"/other/timestamp", "/metadata/*", false},
		{"/status/conditions/0/type", "/status/*", true},
	}

	for _, tt := range tests {
		t.Run(tt.path+" vs "+tt.pattern, func(t *testing.T) {
			if got := matchPath(tt.path, tt.pattern); got != tt.want {
				t.Errorf("matchPath(%q, %q) = %v, want %v", tt.path, tt.pattern, got, tt.want)
			}
		})
	}
}
