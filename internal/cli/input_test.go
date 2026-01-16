package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadInput(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")
	testContent := []byte("name: test\nvalue: 123")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name       string
		path       string
		formatHint string
		wantFormat string
		wantErr    bool
	}{
		{
			name:       "read yaml file",
			path:       testFile,
			formatHint: "auto",
			wantFormat: "yaml",
			wantErr:    false,
		},
		{
			name:       "explicit format",
			path:       testFile,
			formatHint: "json",
			wantFormat: "json",
			wantErr:    false,
		},
		{
			name:       "non-existent file",
			path:       "/nonexistent/file.yaml",
			formatHint: "auto",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := ReadInput(tt.path, tt.formatHint)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if input.Format != tt.wantFormat {
					t.Errorf("ReadInput() format = %v, want %v", input.Format, tt.wantFormat)
				}
				if input.Path != tt.path {
					t.Errorf("ReadInput() path = %v, want %v", input.Path, tt.path)
				}
			}
		})
	}
}

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		data    []byte
		want    string
	}{
		{
			name: "yaml extension",
			path: "test.yaml",
			data: []byte("name: test"),
			want: "yaml",
		},
		{
			name: "json extension",
			path: "test.json",
			data: []byte(`{"name": "test"}`),
			want: "json",
		},
		{
			name: "json content",
			path: "test.txt",
			data: []byte(`{"name": "test"}`),
			want: "json",
		},
		{
			name: "yaml content",
			path: "test.txt",
			data: []byte("name: test\nvalue: 123"),
			want: "yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectFormat(tt.path, tt.data)
			if got != tt.want {
				t.Errorf("detectFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectFromContent(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "json object",
			data: []byte(`{"key": "value"}`),
			want: "json",
		},
		{
			name: "json array",
			data: []byte(`["item1", "item2"]`),
			want: "json",
		},
		{
			name: "yaml",
			data: []byte("key: value\nother: test"),
			want: "yaml",
		},
		{
			name: "empty",
			data: []byte(""),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectFromContent(tt.data)
			if got != tt.want {
				t.Errorf("detectFromContent() = %v, want %v", got, tt.want)
			}
		})
	}
}
