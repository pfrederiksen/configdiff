package cli

import (
	"testing"
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
