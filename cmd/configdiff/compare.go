package main

import (
	"fmt"
	"os"

	"github.com/pfrederiksen/configdiff"
	"github.com/pfrederiksen/configdiff/internal/cli"
)

// compare performs the diff operation between two files
func compare(oldFile, newFile string) error {
	// Build CLI options from flags
	cliOpts := cli.CLIOptions{
		OldFile:        oldFile,
		NewFile:        newFile,
		Format:         format,
		OldFormat:      oldFormat,
		NewFormat:      newFormat,
		IgnorePaths:    ignorePaths,
		ArrayKeys:      arrayKeys,
		NumericStrings: numericStrings,
		BoolStrings:    boolStrings,
		StableOrder:    stableOrder,
		OutputFormat:   outputFormat,
		NoColor:        noColor,
		MaxValueLength: maxValueLength,
		Quiet:          quiet,
		ExitCode:       exitCode,
	}

	// Apply config file defaults (CLI flags take precedence)
	if cfg != nil {
		cliOpts.ApplyConfigDefaults(cfg)
	}

	// Validate options
	if err := cliOpts.Validate(); err != nil {
		return err
	}

	// Read old file
	oldInput, err := cli.ReadInput(oldFile, cliOpts.GetOldFormat())
	if err != nil {
		return err
	}

	// Read new file
	newInput, err := cli.ReadInput(newFile, cliOpts.GetNewFormat())
	if err != nil {
		return err
	}

	// Convert CLI options to library options
	diffOpts, err := cliOpts.ToLibraryOptions()
	if err != nil {
		return err
	}

	// Perform the diff
	result, err := configdiff.DiffBytes(
		oldInput.Data, oldInput.Format,
		newInput.Data, newInput.Format,
		diffOpts,
	)
	if err != nil {
		return fmt.Errorf("diff failed: %w", err)
	}

	// Format and output results (unless quiet mode)
	if !quiet {
		output, err := cli.FormatOutput(result, cli.OutputOptions{
			Format:         outputFormat,
			NoColor:        noColor,
			MaxValueLength: maxValueLength,
		})
		if err != nil {
			return err
		}

		fmt.Println(output)
	}

	// Handle exit code mode
	if exitCode && cli.HasChanges(result) {
		os.Exit(1)
	}

	return nil
}
