package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// FileChange represents a changed file in git
type FileChange struct {
	Path      string
	Status    string // A=added, M=modified, D=deleted, R=renamed
	OldPath   string // For renamed files
	Diff      string
	Content   string
	IsBinary  bool
}

// IsGitRepo checks if the current directory is a git repository
func IsGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}

// GetGitDir returns the path to the .git directory
func GetGitDir() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git rev-parse failed: %w: %s", err, output)
	}
	gitDir := strings.TrimSpace(string(output))
	return filepath.Abs(gitDir)
}

// GetRepoRoot returns the root directory of the git repository
func GetRepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git rev-parse failed: %w: %s", err, output)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetStagedChanges returns a list of staged file changes
func GetStagedChanges() ([]FileChange, error) {
	// Get list of staged files with status
	cmd := exec.Command("git", "diff", "--cached", "--name-status")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get staged files: %w", err)
	}

	if len(output) == 0 {
		return nil, nil
	}

	var changes []FileChange
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		change := FileChange{
			Status: parts[0],
			Path:   parts[len(parts)-1],
		}

		// Handle renamed files (R100 old new)
		if strings.HasPrefix(change.Status, "R") {
			if len(parts) >= 3 {
				change.OldPath = parts[1]
				change.Path = parts[2]
			}
			change.Status = "R"
		}

		// Skip deleted files
		if change.Status == "D" {
			continue
		}

		// Check if binary
		change.IsBinary = isBinaryFile(change.Path)

		// Get diff for non-binary files
		if !change.IsBinary {
			diff, err := getStagedDiff(change.Path)
			if err == nil {
				change.Diff = diff
			}

			content, err := getStagedContent(change.Path)
			if err == nil {
				change.Content = content
			}
		}

		changes = append(changes, change)
	}

	return changes, nil
}

// getStagedDiff returns the staged diff for a file
func getStagedDiff(path string) (string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git diff failed for %s: %w: %s", path, err, output)
	}
	return string(output), nil
}

// getStagedContent returns the staged content of a file
func getStagedContent(path string) (string, error) {
	cmd := exec.Command("git", "show", ":"+path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git show failed for %s: %w: %s", path, err, output)
	}
	return string(output), nil
}

// isBinaryFile checks if a file is binary
func isBinaryFile(path string) bool {
	// Check by extension first
	binaryExtensions := []string{
		".png", ".jpg", ".jpeg", ".gif", ".ico", ".webp",
		".pdf", ".zip", ".tar", ".gz", ".exe", ".dll",
		".so", ".dylib", ".woff", ".woff2", ".ttf", ".eot",
	}

	ext := strings.ToLower(filepath.Ext(path))
	for _, binExt := range binaryExtensions {
		if ext == binExt {
			return true
		}
	}

	// Check with git
	cmd := exec.Command("git", "diff", "--cached", "--numstat", "--", path)
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// Binary files show as "-\t-\t" in numstat
	return bytes.HasPrefix(output, []byte("-\t-\t"))
}

// StageFile stages a file
func StageFile(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	cmd := exec.Command("git", "add", "--", path)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stage file %s: %w", path, err)
	}
	return nil
}

// GetCurrentBranch returns the current branch name
func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git rev-parse failed: %w: %s", err, output)
	}
	return strings.TrimSpace(string(output)), nil
}
