package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pfrederiksen/configdiff"
	"github.com/pfrederiksen/configdiff/internal/cli"
)

// compare performs the diff operation between two files or directories
func compare(oldFile, newFile string) error {
	// Check if inputs are directories
	oldInfo, oldErr := os.Stat(oldFile)
	newInfo, newErr := os.Stat(newFile)

	// Handle directory comparison
	if oldErr == nil && newErr == nil && oldInfo.IsDir() && newInfo.IsDir() {
		if !recursive {
			return fmt.Errorf("comparing directories requires --recursive flag")
		}
		hasChanges, err := compareDirectories(oldFile, newFile)
		if err != nil {
			return err
		}

		// Handle exit code mode for directory comparison
		if exitCode && hasChanges {
			os.Exit(1)
		}

		return nil
	}

	// One is a directory and one isn't
	if oldErr == nil && oldInfo.IsDir() {
		return fmt.Errorf("cannot compare directory %q with file %q", oldFile, newFile)
	}
	if newErr == nil && newInfo.IsDir() {
		return fmt.Errorf("cannot compare file %q with directory %q", oldFile, newFile)
	}

	// Both are files (or stdin), proceed with normal comparison
	hasChanges, err := compareFiles(oldFile, newFile)
	if err != nil {
		return err
	}

	// Handle exit code mode for single file comparison
	if exitCode && hasChanges {
		os.Exit(1)
	}

	return nil
}

// compareFiles performs the diff operation between two files.
// Returns true if changes were found, false otherwise.
func compareFiles(oldFile, newFile string) (bool, error) {
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
		return false, err
	}

	// Read old file
	oldInput, err := cli.ReadInput(oldFile, cliOpts.GetOldFormat())
	if err != nil {
		return false, err
	}

	// Read new file
	newInput, err := cli.ReadInput(newFile, cliOpts.GetNewFormat())
	if err != nil {
		return false, err
	}

	// Convert CLI options to library options
	diffOpts, err := cliOpts.ToLibraryOptions()
	if err != nil {
		return false, err
	}

	// Perform the diff
	result, err := configdiff.DiffBytes(
		oldInput.Data, oldInput.Format,
		newInput.Data, newInput.Format,
		diffOpts,
	)
	if err != nil {
		return false, fmt.Errorf("diff failed: %w", err)
	}

	// Format and output results (unless quiet mode)
	var output string
	if !quiet {
		output, err = cli.FormatOutput(result, cli.OutputOptions{
			Format:         outputFormat,
			NoColor:        noColor,
			MaxValueLength: maxValueLength,
			OldFile:        oldFile,
			NewFile:        newFile,
		})
		if err != nil {
			return false, err
		}

		fmt.Println(output)
	}

	// Write GitHub Actions outputs if in GHA environment
	hasChanges := cli.HasChanges(result)
	if githubOutput := os.Getenv("GITHUB_OUTPUT"); githubOutput != "" {
		if err := writeGitHubOutputs(githubOutput, hasChanges, output); err != nil {
			// Log error but don't fail the command
			fmt.Fprintf(os.Stderr, "Warning: Failed to write GitHub Actions outputs: %v\n", err)
		}
	}

	// Return whether changes were found
	return hasChanges, nil
}

// compareDirectories recursively compares two directories.
// Returns true if any changes were found, false otherwise.
func compareDirectories(oldDir, newDir string) (bool, error) {
	// Collect all config files from both directories
	oldFiles, err := collectConfigFiles(oldDir)
	if err != nil {
		return false, fmt.Errorf("failed to scan old directory: %w", err)
	}

	newFiles, err := collectConfigFiles(newDir)
	if err != nil {
		return false, fmt.Errorf("failed to scan new directory: %w", err)
	}

	// Build set of all relative paths
	allPaths := make(map[string]bool)
	for _, path := range oldFiles {
		rel, _ := filepath.Rel(oldDir, path)
		allPaths[rel] = true
	}
	for _, path := range newFiles {
		rel, _ := filepath.Rel(newDir, path)
		allPaths[rel] = true
	}

	// Track if any differences found
	hasAnyChanges := false
	filesCompared := 0
	filesAdded := 0
	filesRemoved := 0

	// Compare each file
	for relPath := range allPaths {
		oldPath := filepath.Join(oldDir, relPath)
		newPath := filepath.Join(newDir, relPath)

		oldExists := fileExists(oldPath)
		newExists := fileExists(newPath)

		if oldExists && newExists {
			// File exists in both directories - compare them
			if !quiet {
				fmt.Printf("\n=== %s ===\n", relPath)
			}

			fileHasChanges, err := compareFiles(oldPath, newPath)
			if err != nil {
				if !quiet {
					fmt.Printf("Error: %v\n", err)
				}
				continue
			}
			filesCompared++
			if fileHasChanges {
				hasAnyChanges = true
			}
		} else if newExists && !oldExists {
			// File added
			filesAdded++
			if !quiet {
				fmt.Printf("\n+++ %s (added)\n", relPath)
			}
			hasAnyChanges = true
		} else if oldExists && !newExists {
			// File removed
			filesRemoved++
			if !quiet {
				fmt.Printf("\n--- %s (removed)\n", relPath)
			}
			hasAnyChanges = true
		}
	}

	// Print summary
	if !quiet {
		fmt.Printf("\n")
		fmt.Printf("Summary: %d files compared, %d added, %d removed\n",
			filesCompared, filesAdded, filesRemoved)
	}

	// Return whether any changes were found
	return hasAnyChanges, nil
}

// collectConfigFiles recursively finds all config files in a directory
func collectConfigFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if it's a config file by extension
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".yaml", ".yml", ".json", ".hcl", ".tf", ".toml":
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// writeGitHubOutputs writes GitHub Actions outputs to the GITHUB_OUTPUT file
func writeGitHubOutputs(outputFile string, hasChanges bool, diffOutput string) error {
	f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write has-changes output
	if _, err := fmt.Fprintf(f, "has-changes=%v\n", hasChanges); err != nil {
		return err
	}

	// Generate random delimiter to prevent injection attacks
	// See: https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/workflow-commands-for-github-actions#multiline-strings
	delimiterBytes := make([]byte, 16)
	if _, err := rand.Read(delimiterBytes); err != nil {
		return fmt.Errorf("failed to generate random delimiter: %w", err)
	}
	delimiter := "ghadelimiter_" + hex.EncodeToString(delimiterBytes)

	// Write diff-output using heredoc format with random delimiter
	if _, err := fmt.Fprintf(f, "diff-output<<%s\n%s\n%s\n", delimiter, diffOutput, delimiter); err != nil {
		return err
	}

	return nil
}
