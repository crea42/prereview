package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/emilushi/prereview/internal/git"
	"github.com/emilushi/prereview/internal/output"
	"github.com/emilushi/prereview/internal/review"
	"github.com/emilushi/prereview/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// DefaultMaxFileSize is the default maximum file size to review (100KB)
const DefaultMaxFileSize int64 = 100000

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Review staged changes",
	Long:  `Review staged git changes and provide AI-powered suggestions before committing.`,
	Run:   runReview,
}

func init() {
	rootCmd.AddCommand(reviewCmd)
	reviewCmd.Flags().Bool("markdown", false, "Output suggestions to a markdown file instead of interactive mode")
	viper.BindPFlag("output_markdown", reviewCmd.Flags().Lookup("markdown"))
}

func runReview(cmd *cobra.Command, args []string) {
	// Check if we're in a git repository
	if !git.IsGitRepo() {
		ui.Error("Not a git repository")
		os.Exit(1)
	}

	// Get repo root for standards detection
	repoRoot, err := git.GetRepoRoot()
	if err != nil {
		ui.Warning("Could not determine repository root")
		repoRoot = "."
	}

	// Get staged changes
	changes, err := git.GetStagedChanges()
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to get staged changes: %v", err))
		os.Exit(1)
	}

	// Filter out ignored files
	ignorePatterns := viper.GetStringSlice("ignore_patterns")
	maxFileSize := viper.GetInt64("max_file_size")
	if maxFileSize == 0 {
		maxFileSize = DefaultMaxFileSize
	}

	filteredChanges := make([]git.FileChange, 0, len(changes))
	for _, change := range changes {
		if shouldIgnoreFile(change.Path, ignorePatterns) {
			if viper.GetBool("verbose") {
				ui.Info(fmt.Sprintf("  Skipping ignored file: %s", change.Path))
			}
			continue
		}
		// Check file size
		if int64(len(change.Content)) > maxFileSize {
			if viper.GetBool("verbose") {
				ui.Info(fmt.Sprintf("  Skipping large file: %s (%d bytes)", change.Path, len(change.Content)))
			}
			continue
		}
		filteredChanges = append(filteredChanges, change)
	}
	changes = filteredChanges

	if len(changes) == 0 {
		ui.Info("No staged changes to review")
		return
	}

	ui.Info(fmt.Sprintf("🔍 Reviewing %d changed file(s)...\n", len(changes)))

	// Get custom coding standards from config
	customStandards := viper.GetStringSlice("coding_standards")

	// Create reviewer with coding standards context
	reviewer, err := review.NewReviewer(viper.GetString("model"), repoRoot, customStandards)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to initialize reviewer: %v", err))
		os.Exit(1)
	}
	defer reviewer.Close()

	// Run review
	result, err := reviewer.Review(changes)
	if err != nil {
		ui.Error(fmt.Sprintf("Review failed: %v", err))
		os.Exit(1)
	}

	if len(result.Suggestions) == 0 {
		ui.Success("✓ No issues found! Your code looks good.")
		return
	}

	// Check if markdown output is enabled
	if viper.GetBool("output_markdown") {
		generator := output.NewMarkdownGenerator(repoRoot)
		filePath, err := generator.GenerateSuggestionsFile(result)
		if err != nil {
			ui.Error(fmt.Sprintf("Failed to generate markdown file: %v", err))
			os.Exit(1)
		}
		ui.Success(fmt.Sprintf("✓ Generated suggestions file: %s", filePath))
		ui.Info(fmt.Sprintf("  Found %d suggestion(s) across %d file(s)", len(result.Suggestions), len(result.Files)))
		return
	}

	// Interactive review session
	session := ui.NewReviewSession(result)
	outcome := session.Run()

	// Handle outcome
	switch outcome.Action {
	case ui.ActionCommit:
		ui.Success(fmt.Sprintf("\n✓ Review complete: %d fixed, %d skipped", outcome.Fixed, outcome.Skipped))
		if viper.GetBool("strict") && outcome.Skipped > 0 {
			ui.Warning("Strict mode: Cannot commit with skipped issues")
			os.Exit(1)
		}
	case ui.ActionAbort:
		ui.Info("\n✗ Review aborted")
		os.Exit(1)
	case ui.ActionReReview:
		ui.Info("\n🔄 Re-reviewing changes...")
		runReview(cmd, args)
	}
}

// shouldIgnoreFile checks if a file path matches any ignore patterns
func shouldIgnoreFile(filePath string, patterns []string) bool {
	// Normalize path separators
	normalizedPath := filepath.ToSlash(filePath)
	baseName := filepath.Base(filePath)
	verbose := viper.GetBool("verbose")

	for _, pattern := range patterns {
		normalizedPattern := filepath.ToSlash(pattern)

		// Check if pattern matches the full path
		matched, err := filepath.Match(normalizedPattern, normalizedPath)
		if err != nil && verbose {
			ui.Warning(fmt.Sprintf("Invalid ignore pattern '%s': %v", pattern, err))
		}
		if matched {
			return true
		}

		// Check if pattern matches the base name (e.g., "*.min.js")
		matched, err = filepath.Match(normalizedPattern, baseName)
		if err != nil && verbose {
			ui.Warning(fmt.Sprintf("Invalid ignore pattern '%s': %v", pattern, err))
		}
		if matched {
			return true
		}

		// Handle directory patterns like "build/*", "vendor/*", "node_modules/*"
		// These should match the directory anywhere in the path
		if strings.HasSuffix(normalizedPattern, "/*") {
			dir := normalizedPattern[:len(normalizedPattern)-2]
			// Match: "build/*" should match "foo/build/bar.js" and "build/bar.js"
			if strings.HasPrefix(normalizedPath, dir+"/") ||
				strings.Contains(normalizedPath, "/"+dir+"/") {
				return true
			}
		}

		// Handle "**" patterns (recursive match)
		if strings.Contains(normalizedPattern, "**") {
			// Simple glob matching for common cases
			// e.g., "**/build/*" matches any path containing "/build/"
			if strings.HasPrefix(normalizedPattern, "**/") {
				suffix := normalizedPattern[3:] // Remove "**/"
				if strings.HasSuffix(suffix, "/*") {
					dir := suffix[:len(suffix)-2]
					if strings.Contains(normalizedPath, "/"+dir+"/") ||
						strings.HasPrefix(normalizedPath, dir+"/") {
						return true
					}
				} else {
					// Match file pattern anywhere in path
					matched, err := filepath.Match(suffix, baseName)
					if err != nil && verbose {
						ui.Warning(fmt.Sprintf("Invalid ignore pattern '%s': %v", pattern, err))
					}
					if matched {
						return true
					}
				}
			}
		}
	}
	return false
}
