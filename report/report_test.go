package report

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/pfrederiksen/configdiff/diff"
	"github.com/pfrederiksen/configdiff/tree"
)

var updateGolden = flag.Bool("update", false, "update golden files")

func TestGenerate(t *testing.T) {
	tests := []struct {
		name    string
		changes []diff.Change
		opts    Options
		golden  string
	}{
		{
			name:    "empty changes",
			changes: []diff.Change{},
			opts:    DefaultOptions(),
			golden:  "empty.txt",
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
			opts:   DefaultOptions(),
			golden: "single_add.txt",
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
			opts:   DefaultOptions(),
			golden: "single_remove.txt",
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
			opts:   DefaultOptions(),
			golden: "single_modify.txt",
		},
		{
			name: "multiple changes",
			changes: []diff.Change{
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
			},
			opts:   DefaultOptions(),
			golden: "multiple_changes.txt",
		},
		{
			name: "complex values",
			changes: []diff.Change{
				{
					Type: diff.ChangeTypeAdd,
					Path: "/config",
					NewValue: tree.NewObject(map[string]*tree.Node{
						"key1": tree.NewString("value1"),
						"key2": tree.NewNumber(42),
					}),
				},
				{
					Type:     diff.ChangeTypeModify,
					Path:     "/items",
					OldValue: tree.NewArray([]*tree.Node{tree.NewString("a")}),
					NewValue: tree.NewArray([]*tree.Node{tree.NewString("a"), tree.NewString("b")}),
				},
			},
			opts:   DefaultOptions(),
			golden: "complex_values.txt",
		},
		{
			name: "compact format",
			changes: []diff.Change{
				{
					Type:     diff.ChangeTypeAdd,
					Path:     "/newKey",
					NewValue: tree.NewString("value"),
				},
				{
					Type:     diff.ChangeTypeRemove,
					Path:     "/oldKey",
					OldValue: tree.NewString("value"),
				},
			},
			opts: Options{
				Compact:    true,
				ShowValues: false,
			},
			golden: "compact_format.txt",
		},
		{
			name: "without values",
			changes: []diff.Change{
				{
					Type:     diff.ChangeTypeAdd,
					Path:     "/a",
					NewValue: tree.NewString("value"),
				},
				{
					Type:     diff.ChangeTypeModify,
					Path:     "/b",
					OldValue: tree.NewNumber(1),
					NewValue: tree.NewNumber(2),
				},
			},
			opts: Options{
				ShowValues: false,
			},
			golden: "without_values.txt",
		},
		{
			name: "value truncation",
			changes: []diff.Change{
				{
					Type:     diff.ChangeTypeAdd,
					Path:     "/longString",
					NewValue: tree.NewString("This is a very long string that should be truncated in the output because it exceeds the maximum length"),
				},
			},
			opts: Options{
				ShowValues:     true,
				MaxValueLength: 30,
			},
			golden: "value_truncation.txt",
		},
		{
			name: "number formatting",
			changes: []diff.Change{
				{
					Type:     diff.ChangeTypeAdd,
					Path:     "/wholeNumber",
					NewValue: tree.NewNumber(42),
				},
				{
					Type:     diff.ChangeTypeAdd,
					Path:     "/decimal",
					NewValue: tree.NewNumber(3.14159),
				},
			},
			opts:   DefaultOptions(),
			golden: "number_formatting.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Generate(tt.changes, tt.opts)

			goldenPath := filepath.Join("..", "testdata", "report", tt.golden)

			if *updateGolden {
				// Update golden file
				if err := os.MkdirAll(filepath.Dir(goldenPath), 0755); err != nil {
					t.Fatalf("Failed to create directory: %v", err)
				}
				if err := os.WriteFile(goldenPath, []byte(got), 0644); err != nil {
					t.Fatalf("Failed to update golden file: %v", err)
				}
			}

			// Read golden file
			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("Failed to read golden file %s: %v (run with -update to create)", goldenPath, err)
			}

			if got != string(want) {
				t.Errorf("Generate() output differs from golden file %s\nGot:\n%s\nWant:\n%s", tt.golden, got, string(want))
				t.Logf("Run with -update flag to update golden files")
			}
		})
	}
}

func TestGenerateCompact(t *testing.T) {
	changes := []diff.Change{
		{
			Type:     diff.ChangeTypeAdd,
			Path:     "/a",
			NewValue: tree.NewString("value"),
		},
	}

	output := GenerateCompact(changes)
	if output == "" {
		t.Error("GenerateCompact() returned empty string")
	}
	// Compact output should not show values
	if len(output) > 100 {
		t.Error("GenerateCompact() output is too long")
	}
}

func TestGenerateDetailed(t *testing.T) {
	changes := []diff.Change{
		{
			Type:     diff.ChangeTypeModify,
			Path:     "/key",
			OldValue: tree.NewString("old"),
			NewValue: tree.NewString("new"),
		},
	}

	output := GenerateDetailed(changes)
	if output == "" {
		t.Error("GenerateDetailed() returned empty string")
	}
	// Detailed output should show values
	if !contains(output, "old") || !contains(output, "new") {
		t.Error("GenerateDetailed() should show old and new values")
	}
}

func TestSummarizeChanges(t *testing.T) {
	changes := []diff.Change{
		{Type: diff.ChangeTypeAdd},
		{Type: diff.ChangeTypeAdd},
		{Type: diff.ChangeTypeRemove},
		{Type: diff.ChangeTypeModify},
		{Type: diff.ChangeTypeModify},
		{Type: diff.ChangeTypeModify},
	}

	summary := summarizeChanges(changes)

	if summary.Total != 6 {
		t.Errorf("Total = %d, want 6", summary.Total)
	}
	if summary.Added != 2 {
		t.Errorf("Added = %d, want 2", summary.Added)
	}
	if summary.Removed != 1 {
		t.Errorf("Removed = %d, want 1", summary.Removed)
	}
	if summary.Modified != 3 {
		t.Errorf("Modified = %d, want 3", summary.Modified)
	}
}

func TestFormatSummary(t *testing.T) {
	tests := []struct {
		name    string
		summary Summary
		want    string
	}{
		{
			name:    "only adds",
			summary: Summary{Total: 2, Added: 2},
			want:    "Summary: +2 added (2 total)\n",
		},
		{
			name:    "only removes",
			summary: Summary{Total: 1, Removed: 1},
			want:    "Summary: -1 removed (1 total)\n",
		},
		{
			name:    "mixed",
			summary: Summary{Total: 4, Added: 1, Removed: 1, Modified: 2},
			want:    "Summary: +1 added, -1 removed, ~2 modified (4 total)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := Options{NoColor: true} // Disable color in tests
			got := formatSummary(tt.summary, opts)
			if got != tt.want {
				t.Errorf("formatSummary() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetChangeSymbol(t *testing.T) {
	tests := []struct {
		changeType diff.ChangeType
		want       string
	}{
		{diff.ChangeTypeAdd, "+"},
		{diff.ChangeTypeRemove, "-"},
		{diff.ChangeTypeModify, "~"},
		{diff.ChangeTypeMove, "↔"},
	}

	for _, tt := range tests {
		t.Run(string(tt.changeType), func(t *testing.T) {
			got := getChangeSymbol(tt.changeType)
			if got != tt.want {
				t.Errorf("getChangeSymbol(%v) = %v, want %v", tt.changeType, got, tt.want)
			}
		})
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name   string
		node   *tree.Node
		maxLen int
		want   string
	}{
		{
			name:   "nil node",
			node:   nil,
			maxLen: 0,
			want:   "<nil>",
		},
		{
			name:   "null",
			node:   tree.NewNull(),
			maxLen: 0,
			want:   "null",
		},
		{
			name:   "bool true",
			node:   tree.NewBool(true),
			maxLen: 0,
			want:   "true",
		},
		{
			name:   "whole number",
			node:   tree.NewNumber(42),
			maxLen: 0,
			want:   "42",
		},
		{
			name:   "decimal number",
			node:   tree.NewNumber(3.14),
			maxLen: 0,
			want:   "3.14",
		},
		{
			name:   "string",
			node:   tree.NewString("hello"),
			maxLen: 0,
			want:   `"hello"`,
		},
		{
			name:   "object",
			node:   tree.NewObject(map[string]*tree.Node{"a": tree.NewNull()}),
			maxLen: 0,
			want:   "{...} (1 keys)",
		},
		{
			name:   "array",
			node:   tree.NewArray([]*tree.Node{tree.NewNull(), tree.NewNull()}),
			maxLen: 0,
			want:   "[...] (2 items)",
		},
		{
			name:   "truncation",
			node:   tree.NewString("this is a long string"),
			maxLen: 10,
			want:   `"this i...`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatValue(tt.node, tt.maxLen)
			if got != tt.want {
				t.Errorf("formatValue() = %q, want %q", got, tt.want)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestGenerateStat(t *testing.T) {
	tests := []struct {
		name    string
		changes []diff.Change
		golden  string
	}{
		{
			name:    "empty changes",
			changes: []diff.Change{},
			golden:  "stat_empty.txt",
		},
		{
			name: "single modification",
			changes: []diff.Change{
				{
					Type:     diff.ChangeTypeModify,
					Path:     "/version",
					OldValue: tree.NewString("1.0"),
					NewValue: tree.NewString("2.0"),
				},
			},
			golden: "stat_single_modify.txt",
		},
		{
			name: "multiple changes",
			changes: []diff.Change{
				{
					Type:     diff.ChangeTypeAdd,
					Path:     "/newKey",
					NewValue: tree.NewString("value"),
				},
				{
					Type:     diff.ChangeTypeRemove,
					Path:     "/oldKey",
					OldValue: tree.NewString("value"),
				},
				{
					Type:     diff.ChangeTypeModify,
					Path:     "/changedKey",
					OldValue: tree.NewString("old"),
					NewValue: tree.NewString("new"),
				},
			},
			golden: "stat_multiple.txt",
		},
		{
			name: "mixed changes",
			changes: []diff.Change{
				{
					Type:     diff.ChangeTypeAdd,
					Path:     "/config/new",
					NewValue: tree.NewString("value"),
				},
				{
					Type:     diff.ChangeTypeAdd,
					Path:     "/config/another",
					NewValue: tree.NewNumber(42),
				},
				{
					Type:     diff.ChangeTypeRemove,
					Path:     "/old/setting",
					OldValue: tree.NewBool(true),
				},
				{
					Type:     diff.ChangeTypeModify,
					Path:     "/replicas",
					OldValue: tree.NewNumber(2),
					NewValue: tree.NewNumber(5),
				},
				{
					Type:     diff.ChangeTypeMove,
					Path:     "/position",
					OldValue: tree.NewNumber(0),
					NewValue: tree.NewNumber(1),
				},
			},
			golden: "stat_mixed.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateStat(tt.changes)

			goldenPath := filepath.Join("..", "testdata", "report", tt.golden)

			if *updateGolden {
				// Update golden file
				if err := os.MkdirAll(filepath.Dir(goldenPath), 0755); err != nil {
					t.Fatalf("Failed to create directory: %v", err)
				}
				if err := os.WriteFile(goldenPath, []byte(got), 0644); err != nil {
					t.Fatalf("Failed to update golden file: %v", err)
				}
			}

			// Read golden file
			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("Failed to read golden file %s: %v (run with -update to create)", goldenPath, err)
			}

			if got != string(want) {
				t.Errorf("GenerateStat() output differs from golden file %s\nGot:\n%s\nWant:\n%s", tt.golden, got, string(want))
				t.Logf("Run with -update flag to update golden files")
			}
		})
	}
}

func TestGenerateSideBySide(t *testing.T) {
	tests := []struct {
		name    string
		changes []diff.Change
		opts    Options
		golden  string
	}{
		{
			name:    "empty changes",
			changes: []diff.Change{},
			opts:    Options{NoColor: true},
			golden:  "side_by_side_empty.txt",
		},
		{
			name: "single addition",
			changes: []diff.Change{
				{
					Type:     diff.ChangeTypeAdd,
					Path:     "/newKey",
					NewValue: tree.NewString("value"),
				},
			},
			opts:   Options{NoColor: true},
			golden: "side_by_side_add.txt",
		},
		{
			name: "single removal",
			changes: []diff.Change{
				{
					Type:     diff.ChangeTypeRemove,
					Path:     "/oldKey",
					OldValue: tree.NewString("value"),
				},
			},
			opts:   Options{NoColor: true},
			golden: "side_by_side_remove.txt",
		},
		{
			name: "modification",
			changes: []diff.Change{
				{
					Type:     diff.ChangeTypeModify,
					Path:     "/key",
					OldValue: tree.NewString("old"),
					NewValue: tree.NewString("new"),
				},
			},
			opts:   Options{NoColor: true},
			golden: "side_by_side_modify.txt",
		},
		{
			name: "multiple changes",
			changes: []diff.Change{
				{
					Type:     diff.ChangeTypeAdd,
					Path:     "/env",
					NewValue: tree.NewString("production"),
				},
				{
					Type:     diff.ChangeTypeRemove,
					Path:     "/debug",
					OldValue: tree.NewBool(true),
				},
				{
					Type:     diff.ChangeTypeModify,
					Path:     "/replicas",
					OldValue: tree.NewNumber(2),
					NewValue: tree.NewNumber(5),
				},
			},
			opts:   Options{NoColor: true},
			golden: "side_by_side_multiple.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateSideBySide(tt.changes, tt.opts)

			goldenPath := filepath.Join("..", "testdata", "report", tt.golden)

			if *updateGolden {
				// Update golden file
				if err := os.MkdirAll(filepath.Dir(goldenPath), 0755); err != nil {
					t.Fatalf("Failed to create directory: %v", err)
				}
				if err := os.WriteFile(goldenPath, []byte(got), 0644); err != nil {
					t.Fatalf("Failed to update golden file: %v", err)
				}
			}

			// Read golden file
			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("Failed to read golden file %s: %v (run with -update to create)", goldenPath, err)
			}

			if got != string(want) {
				t.Errorf("GenerateSideBySide() output differs from golden file %s\nGot:\n%s\nWant:\n%s", tt.golden, got, string(want))
				t.Logf("Run with -update flag to update golden files")
			}
		})
	}
}

func TestGenerateGitDiff(t *testing.T) {
	tests := []struct {
		name    string
		changes []diff.Change
		oldFile string
		newFile string
		golden  string
	}{
		{
			name:    "empty changes",
			changes: []diff.Change{},
			oldFile: "old.yaml",
			newFile: "new.yaml",
			golden:  "git_diff_empty.txt",
		},
		{
			name: "single addition",
			changes: []diff.Change{
				{
					Type:     diff.ChangeTypeAdd,
					Path:     "/newKey",
					NewValue: tree.NewString("value"),
				},
			},
			oldFile: "old.yaml",
			newFile: "new.yaml",
			golden:  "git_diff_add.txt",
		},
		{
			name: "single removal",
			changes: []diff.Change{
				{
					Type:     diff.ChangeTypeRemove,
					Path:     "/oldKey",
					OldValue: tree.NewString("value"),
				},
			},
			oldFile: "old.yaml",
			newFile: "new.yaml",
			golden:  "git_diff_remove.txt",
		},
		{
			name: "modification",
			changes: []diff.Change{
				{
					Type:     diff.ChangeTypeModify,
					Path:     "/key",
					OldValue: tree.NewString("old"),
					NewValue: tree.NewString("new"),
				},
			},
			oldFile: "old.yaml",
			newFile: "new.yaml",
			golden:  "git_diff_modify.txt",
		},
		{
			name: "multiple changes",
			changes: []diff.Change{
				{
					Type:     diff.ChangeTypeAdd,
					Path:     "/env",
					NewValue: tree.NewString("production"),
				},
				{
					Type:     diff.ChangeTypeRemove,
					Path:     "/debug",
					OldValue: tree.NewBool(true),
				},
				{
					Type:     diff.ChangeTypeModify,
					Path:     "/replicas",
					OldValue: tree.NewNumber(2),
					NewValue: tree.NewNumber(5),
				},
			},
			oldFile: "config.yaml",
			newFile: "config.yaml",
			golden:  "git_diff_multiple.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateGitDiff(tt.changes, tt.oldFile, tt.newFile)

			goldenPath := filepath.Join("..", "testdata", "report", tt.golden)

			if *updateGolden {
				// Update golden file
				if err := os.MkdirAll(filepath.Dir(goldenPath), 0755); err != nil {
					t.Fatalf("Failed to create directory: %v", err)
				}
				if err := os.WriteFile(goldenPath, []byte(got), 0644); err != nil {
					t.Fatalf("Failed to update golden file: %v", err)
				}
			}

			// Read golden file
			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("Failed to read golden file %s: %v (run with -update to create)", goldenPath, err)
			}

			if got != string(want) {
				t.Errorf("GenerateGitDiff() output differs from golden file %s\nGot:\n%s\nWant:\n%s", tt.golden, got, string(want))
				t.Logf("Run with -update flag to update golden files")
			}
		})
	}
}
