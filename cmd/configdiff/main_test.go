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

func TestCollectConfigFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	testFiles := []string{
		"config.yaml",
		"config.yml",
		"data.json",
		"terraform.tf",
		"vars.hcl",
		"Cargo.toml",
		"subdir/nested.yaml",
		"README.md",     // Should not be collected
		"script.sh",     // Should not be collected
	}

	for _, f := range testFiles {
		path := filepath.Join(tmpDir, f)
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", f, err)
		}
	}

	files, err := collectConfigFiles(tmpDir)
	if err != nil {
		t.Fatalf("collectConfigFiles() error = %v", err)
	}

	// Should collect exactly 6 config files (not README.md or script.sh)
	if len(files) != 7 {
		t.Errorf("collectConfigFiles() found %d files, want 7", len(files))
	}

	// Check that config files are present
	wantExtensions := map[string]bool{
		".yaml": false,
		".yml":  false,
		".json": false,
		".tf":   false,
		".hcl":  false,
		".toml": false,
	}

	for _, f := range files {
		ext := filepath.Ext(f)
		if _, ok := wantExtensions[ext]; ok {
			wantExtensions[ext] = true
		}
	}

	for ext, found := range wantExtensions {
		if !found {
			t.Errorf("No %s file found in collected files", ext)
		}
	}
}

func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()

	existingFile := filepath.Join(tmpDir, "exists.txt")
	if err := os.WriteFile(existingFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "existing file",
			path: existingFile,
			want: true,
		},
		{
			name: "non-existent file",
			path: filepath.Join(tmpDir, "nonexistent.txt"),
			want: false,
		},
		{
			name: "directory",
			path: tmpDir,
			want: false, // directories should return false
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fileExists(tt.path)
			if got != tt.want {
				t.Errorf("fileExists(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestCompareDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	oldDir := filepath.Join(tmpDir, "old")
	newDir := filepath.Join(tmpDir, "new")

	if err := os.MkdirAll(oldDir, 0755); err != nil {
		t.Fatalf("Failed to create old dir: %v", err)
	}
	if err := os.MkdirAll(newDir, 0755); err != nil {
		t.Fatalf("Failed to create new dir: %v", err)
	}

	// Create files in both directories
	commonFile := "config.yaml"
	oldOnlyFile := "old-only.json"
	newOnlyFile := "new-only.yaml"

	// File in both (with different content)
	if err := os.WriteFile(filepath.Join(oldDir, commonFile), []byte("version: 1.0"), 0644); err != nil {
		t.Fatalf("Failed to write common file to old dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(newDir, commonFile), []byte("version: 2.0"), 0644); err != nil {
		t.Fatalf("Failed to write common file to new dir: %v", err)
	}

	// File only in old
	if err := os.WriteFile(filepath.Join(oldDir, oldOnlyFile), []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to write old-only file: %v", err)
	}

	// File only in new
	if err := os.WriteFile(filepath.Join(newDir, newOnlyFile), []byte("new: value"), 0644); err != nil {
		t.Fatalf("Failed to write new-only file: %v", err)
	}

	// Test the comparison
	quiet = true // Suppress output during test
	exitCode = false

	err := compareDirectories(oldDir, newDir)
	if err != nil {
		t.Errorf("compareDirectories() error = %v", err)
	}
}

func TestCompareWithDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	dir := filepath.Join(tmpDir, "dir")
	file := filepath.Join(tmpDir, "file.yaml")

	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}
	if err := os.WriteFile(file, []byte("test: value"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	tests := []struct {
		name    string
		oldPath string
		newPath string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "directory without recursive flag",
			oldPath: dir,
			newPath: dir,
			wantErr: true,
			errMsg:  "requires --recursive",
		},
		{
			name:    "directory vs file",
			oldPath: dir,
			newPath: file,
			wantErr: true,
			errMsg:  "cannot compare directory",
		},
		{
			name:    "file vs directory",
			oldPath: file,
			newPath: dir,
			wantErr: true,
			errMsg:  "cannot compare file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recursive = false
			quiet = true
			exitCode = false

			err := compare(tt.oldPath, tt.newPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("compare() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("compare() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
