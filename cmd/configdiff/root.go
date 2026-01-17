package main

import (
	"fmt"

	"github.com/pfrederiksen/configdiff/internal/config"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	format         string
	oldFormat      string
	newFormat      string
	ignorePaths    []string
	arrayKeys      []string
	numericStrings bool
	boolStrings    bool
	stableOrder    bool
	outputFormat   string
	noColor        bool
	maxValueLength int
	quiet          bool
	exitCode       bool

	// Config file loaded at startup
	cfg *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "configdiff [flags] <old-file> <new-file>",
	Short: "Semantic diff for YAML/JSON/HCL configuration files",
	Long: `configdiff provides semantic, human-grade diffs for configuration files.

It understands the structure of your configuration and can apply customizable
rules for semantic comparison, ignore specific paths, treat arrays as sets,
handle type coercions, and generate both machine-readable patches and
human-friendly reports.

Use "-" for stdin input (only one file can be stdin).`,
	Example: `  # Basic comparison
  configdiff old.yaml new.yaml

  # Compare with stdin
  kubectl get deploy myapp -o yaml | configdiff old.yaml -

  # Ignore paths
  configdiff old.yaml new.yaml -i /metadata/generation -i /status/*

  # Array-as-set comparison
  configdiff old.yaml new.yaml --array-key /spec/containers=name

  # Different output formats
  configdiff old.yaml new.yaml -o compact
  configdiff old.yaml new.yaml -o json
  configdiff old.yaml new.yaml -o patch

  # Exit code mode for CI
  if configdiff old.yaml new.yaml --exit-code; then
    echo "No changes detected"
  fi`,
	Args:              cobra.ExactArgs(2),
	RunE:              runCompare,
	SilenceUsage:      true,
	SilenceErrors:     true,
	DisableAutoGenTag: true,
}

func init() {
	// Load config file (errors are ignored - config is optional)
	cfg, _ = config.Load()

	// Format flags
	rootCmd.Flags().StringVarP(&format, "format", "f", "auto", "Input format (yaml, json, auto)")
	rootCmd.Flags().StringVar(&oldFormat, "old-format", "", "Old file format override")
	rootCmd.Flags().StringVar(&newFormat, "new-format", "", "New file format override")

	// Diff option flags
	rootCmd.Flags().StringSliceVarP(&ignorePaths, "ignore", "i", nil, "Paths to ignore (can be repeated)")
	rootCmd.Flags().StringSliceVar(&arrayKeys, "array-key", nil, "Array paths to key fields (format: path=key)")
	rootCmd.Flags().BoolVar(&numericStrings, "numeric-strings", false, "Coerce numeric strings to numbers")
	rootCmd.Flags().BoolVar(&boolStrings, "bool-strings", false, "Coerce bool strings to booleans")
	rootCmd.Flags().BoolVar(&stableOrder, "stable-order", true, "Sort output deterministically")

	// Output flags
	rootCmd.Flags().StringVarP(&outputFormat, "output", "o", "report", "Output format (report, compact, json, patch)")
	rootCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	rootCmd.Flags().IntVar(&maxValueLength, "max-value-length", 80, "Truncate values longer than N chars (0 = no limit)")
	rootCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Quiet mode (no output)")
	rootCmd.Flags().BoolVar(&exitCode, "exit-code", false, "Exit with code 1 if differences found")

	// Add version command
	rootCmd.AddCommand(versionCmd)
}

// runCompare is the main entry point for the compare command
func runCompare(cmd *cobra.Command, args []string) error {
	oldFile := args[0]
	newFile := args[1]

	// Validate that both files aren't stdin
	if oldFile == "-" && newFile == "-" {
		return fmt.Errorf("both old-file and new-file cannot be stdin (\"-\")\nHint: Save one file to disk or use process substitution:\n  configdiff <(command1) <(command2)")
	}

	// This will be implemented in compare.go
	return compare(oldFile, newFile)
}
