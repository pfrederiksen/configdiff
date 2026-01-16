package cli

import (
	"strings"
	"testing"

	"github.com/pfrederiksen/configdiff"
	"github.com/pfrederiksen/configdiff/diff"
	"github.com/pfrederiksen/configdiff/patch"
	"github.com/pfrederiksen/configdiff/tree"
)

func TestFormatOutput(t *testing.T) {
	// Create a simple test result
	oldNode := &tree.Node{Kind: tree.KindString, Value: "old", Path: "/test"}
	newNode := &tree.Node{Kind: tree.KindString, Value: "new", Path: "/test"}

	changes := []diff.Change{
		{
			Type:     diff.ChangeTypeModify,
			Path:     "/test",
			OldValue: oldNode,
			NewValue: newNode,
		},
	}

	testPatch, _ := patch.FromChanges(changes)

	result := &configdiff.Result{
		Changes: changes,
		Patch:   testPatch,
		Report:  "test report",
	}

	tests := []struct {
		name    string
		opts    OutputOptions
		wantErr bool
		check   func(string) bool
	}{
		{
			name: "report format",
			opts: OutputOptions{
				Format:         "report",
				MaxValueLength: 80,
			},
			wantErr: false,
			check: func(s string) bool {
				return strings.Contains(s, "Summary:")
			},
		},
		{
			name: "compact format",
			opts: OutputOptions{
				Format: "compact",
			},
			wantErr: false,
			check: func(s string) bool {
				return strings.Contains(s, "/test")
			},
		},
		{
			name: "json format",
			opts: OutputOptions{
				Format: "json",
			},
			wantErr: false,
			check: func(s string) bool {
				return strings.Contains(s, "\"Type\"") || strings.Contains(s, "\"type\"")
			},
		},
		{
			name: "patch format",
			opts: OutputOptions{
				Format: "patch",
			},
			wantErr: false,
			check: func(s string) bool {
				return strings.Contains(s, "operations")
			},
		},
		{
			name: "invalid format",
			opts: OutputOptions{
				Format: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := FormatOutput(result, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("FormatOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && tt.check != nil {
				if !tt.check(output) {
					t.Errorf("FormatOutput() output check failed, got: %s", output)
				}
			}
		})
	}
}

func TestHasChanges(t *testing.T) {
	tests := []struct {
		name   string
		result *configdiff.Result
		want   bool
	}{
		{
			name: "has changes",
			result: &configdiff.Result{
				Changes: []diff.Change{
					{Type: diff.ChangeTypeAdd, Path: "/test"},
				},
			},
			want: true,
		},
		{
			name: "no changes",
			result: &configdiff.Result{
				Changes: []diff.Change{},
			},
			want: false,
		},
		{
			name: "nil changes",
			result: &configdiff.Result{
				Changes: nil,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasChanges(tt.result)
			if got != tt.want {
				t.Errorf("HasChanges() = %v, want %v", got, tt.want)
			}
		})
	}
}
