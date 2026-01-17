package cli

import (
	"testing"

	"github.com/pfrederiksen/configdiff/internal/config"
)

func TestCLIOptions_ToLibraryOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    CLIOptions
		wantErr bool
	}{
		{
			name: "basic options",
			opts: CLIOptions{
				IgnorePaths:    []string{"/metadata/generation"},
				NumericStrings: true,
				BoolStrings:    true,
				StableOrder:    true,
			},
			wantErr: false,
		},
		{
			name: "array keys",
			opts: CLIOptions{
				ArrayKeys: []string{"/spec/containers=name", "volumes=name"},
			},
			wantErr: false,
		},
		{
			name: "invalid array key format",
			opts: CLIOptions{
				ArrayKeys: []string{"invalid"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			libOpts, err := tt.opts.ToLibraryOptions()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToLibraryOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if libOpts.StableOrder != tt.opts.StableOrder {
					t.Errorf("StableOrder = %v, want %v", libOpts.StableOrder, tt.opts.StableOrder)
				}
				if libOpts.Coercions.NumericStrings != tt.opts.NumericStrings {
					t.Errorf("NumericStrings = %v, want %v", libOpts.Coercions.NumericStrings, tt.opts.NumericStrings)
				}
			}
		})
	}
}

func TestCLIOptions_GetOldFormat(t *testing.T) {
	tests := []struct {
		name string
		opts CLIOptions
		want string
	}{
		{
			name: "old format specified",
			opts: CLIOptions{Format: "yaml", OldFormat: "json"},
			want: "json",
		},
		{
			name: "fall back to format",
			opts: CLIOptions{Format: "yaml"},
			want: "yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.opts.GetOldFormat()
			if got != tt.want {
				t.Errorf("GetOldFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCLIOptions_GetNewFormat(t *testing.T) {
	tests := []struct {
		name string
		opts CLIOptions
		want string
	}{
		{
			name: "new format specified",
			opts: CLIOptions{Format: "yaml", NewFormat: "json"},
			want: "json",
		},
		{
			name: "fall back to format",
			opts: CLIOptions{Format: "json"},
			want: "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.opts.GetNewFormat()
			if got != tt.want {
				t.Errorf("GetNewFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCLIOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    CLIOptions
		wantErr bool
	}{
		{
			name: "valid options",
			opts: CLIOptions{
				Format:       "yaml",
				OutputFormat: "report",
			},
			wantErr: false,
		},
		{
			name: "invalid output format",
			opts: CLIOptions{
				Format:       "yaml",
				OutputFormat: "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid input format",
			opts: CLIOptions{
				Format:       "invalid",
				OutputFormat: "report",
			},
			wantErr: true,
		},
		{
			name: "invalid old format",
			opts: CLIOptions{
				Format:       "yaml",
				OldFormat:    "invalid",
				OutputFormat: "report",
			},
			wantErr: true,
		},
		{
			name: "invalid new format",
			opts: CLIOptions{
				Format:       "yaml",
				NewFormat:    "invalid",
				OutputFormat: "report",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
func TestCLIOptions_ApplyConfigDefaults(t *testing.T) {
	tests := []struct {
		name   string
		opts   CLIOptions
		config *config.Config
		want   CLIOptions
	}{
		{
			name: "empty config no-op",
			opts: CLIOptions{
				OutputFormat: "report",
			},
			config: &config.Config{},
			want: CLIOptions{
				OutputFormat: "report",
			},
		},
		{
			name: "apply config defaults when CLI empty",
			opts: CLIOptions{},
			config: &config.Config{
				IgnorePaths:    []string{"/status"},
				NumericStrings: true,
				BoolStrings:    true,
				StableOrder:    true,
				OutputFormat:   "compact",
				MaxValueLength: 50,
				NoColor:        true,
			},
			want: CLIOptions{
				IgnorePaths:    []string{"/status"},
				NumericStrings: true,
				BoolStrings:    true,
				StableOrder:    true,
				OutputFormat:   "compact",
				MaxValueLength: 50,
				NoColor:        true,
			},
		},
		{
			name: "CLI values take precedence",
			opts: CLIOptions{
				IgnorePaths:    []string{"/metadata"},
				NumericStrings: true,
				OutputFormat:   "json",
				MaxValueLength: 100,
			},
			config: &config.Config{
				IgnorePaths:    []string{"/status"},
				NumericStrings: false,
				OutputFormat:   "compact",
				MaxValueLength: 50,
			},
			want: CLIOptions{
				IgnorePaths:    []string{"/metadata", "/status"}, // Merged
				NumericStrings: true,                             // CLI wins (already set)
				OutputFormat:   "json",                           // CLI wins
				MaxValueLength: 100,                              // CLI wins
			},
		},
		{
			name: "merge ignore paths deduplication",
			opts: CLIOptions{
				IgnorePaths: []string{"/metadata", "/status"},
			},
			config: &config.Config{
				IgnorePaths: []string{"/status", "/timestamp"},
			},
			want: CLIOptions{
				IgnorePaths: []string{"/metadata", "/status", "/timestamp"},
			},
		},
		{
			name: "merge array keys",
			opts: CLIOptions{
				ArrayKeys: []string{"/containers=name"},
			},
			config: &config.Config{
				ArrayKeys: map[string]string{
					"/volumes": "name",
					"/ports":   "port",
				},
			},
			want: CLIOptions{
				ArrayKeys: []string{"/containers=name", "/volumes=name", "/ports=port"},
			},
		},
		{
			name: "boolean flags - config applies when CLI is false",
			opts: CLIOptions{
				NumericStrings: false,
				BoolStrings:    false,
				StableOrder:    false,
				NoColor:        false,
			},
			config: &config.Config{
				NumericStrings: true,
				BoolStrings:    true,
				StableOrder:    true,
				NoColor:        true,
			},
			want: CLIOptions{
				NumericStrings: true,
				BoolStrings:    true,
				StableOrder:    true,
				NoColor:        true,
			},
		},
		{
			name: "string defaults - config applies when CLI is default",
			opts: CLIOptions{
				OutputFormat: "report", // Default value
			},
			config: &config.Config{
				OutputFormat: "compact",
			},
			want: CLIOptions{
				OutputFormat: "compact",
			},
		},
		{
			name: "numeric defaults - config applies when CLI is zero",
			opts: CLIOptions{
				MaxValueLength: 0, // Default value
			},
			config: &config.Config{
				MaxValueLength: 100,
			},
			want: CLIOptions{
				MaxValueLength: 100,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := tt.opts
			opts.ApplyConfigDefaults(tt.config)

			// Check arrays with element comparison (order-independent)
			if !containsAll(opts.IgnorePaths, tt.want.IgnorePaths) {
				t.Errorf("IgnorePaths = %v, want to contain all of %v", opts.IgnorePaths, tt.want.IgnorePaths)
			}

			// Check array keys
			if len(opts.ArrayKeys) != len(tt.want.ArrayKeys) {
				t.Errorf("ArrayKeys length = %d, want %d", len(opts.ArrayKeys), len(tt.want.ArrayKeys))
			}

			// Check boolean flags
			if opts.NumericStrings != tt.want.NumericStrings {
				t.Errorf("NumericStrings = %v, want %v", opts.NumericStrings, tt.want.NumericStrings)
			}
			if opts.BoolStrings != tt.want.BoolStrings {
				t.Errorf("BoolStrings = %v, want %v", opts.BoolStrings, tt.want.BoolStrings)
			}
			if opts.StableOrder != tt.want.StableOrder {
				t.Errorf("StableOrder = %v, want %v", opts.StableOrder, tt.want.StableOrder)
			}
			if opts.NoColor != tt.want.NoColor {
				t.Errorf("NoColor = %v, want %v", opts.NoColor, tt.want.NoColor)
			}

			// Check string options
			if opts.OutputFormat != tt.want.OutputFormat {
				t.Errorf("OutputFormat = %v, want %v", opts.OutputFormat, tt.want.OutputFormat)
			}

			// Check numeric options
			if opts.MaxValueLength != tt.want.MaxValueLength {
				t.Errorf("MaxValueLength = %v, want %v", opts.MaxValueLength, tt.want.MaxValueLength)
			}
		})
	}
}

// containsAll checks if actual contains all elements from expected
func containsAll(actual, expected []string) bool {
	if len(actual) < len(expected) {
		return false
	}
	expectedMap := make(map[string]bool)
	for _, e := range expected {
		expectedMap[e] = true
	}
	for _, a := range actual {
		if expectedMap[a] {
			delete(expectedMap, a)
		}
	}
	return len(expectedMap) == 0
}
