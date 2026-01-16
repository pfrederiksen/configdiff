package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCLI(t *testing.T) {
	// Create temporary test files
	tmpDir := t.TempDir()

	oldFile := filepath.Join(tmpDir, "old.yaml")
	newFile := filepath.Join(tmpDir, "new.yaml")

	oldContent := []byte("name: test\nvalue: 1")
	newContent := []byte("name: test\nvalue: 2")

	if err := os.WriteFile(oldFile, oldContent, 0644); err != nil {
		t.Fatalf("Failed to write old file: %v", err)
	}
	if err := os.WriteFile(newFile, newContent, 0644); err != nil {
		t.Fatalf("Failed to write new file: %v", err)
	}

	tests := []struct {
		name    string
		oldFile string
		newFile string
		wantErr bool
	}{
		{
			name:    "basic comparison",
			oldFile: oldFile,
			newFile: newFile,
			wantErr: false,
		},
		{
			name:    "non-existent old file",
			oldFile: "/nonexistent/old.yaml",
			newFile: newFile,
			wantErr: true,
		},
		{
			name:    "non-existent new file",
			oldFile: oldFile,
			newFile: "/nonexistent/new.yaml",
			wantErr: true,
		},
		{
			name:    "both stdin",
			oldFile: "-",
			newFile: "-",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set quiet mode to avoid output during tests
			quiet = true
			exitCode = false

			err := compare(tt.oldFile, tt.newFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("compare() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVersionInfo(t *testing.T) {
	// Test that version variables exist
	if version == "" {
		t.Error("version should not be empty")
	}
	if commit == "" {
		t.Error("commit should not be empty")
	}
	if date == "" {
		t.Error("date should not be empty")
	}
	if builtBy == "" {
		t.Error("builtBy should not be empty")
	}
}
