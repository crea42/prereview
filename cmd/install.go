package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/emilushi/prereview/internal/git"
	"github.com/emilushi/prereview/internal/ui"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install prereview as a git pre-commit hook",
	Long:  `Install prereview as a git pre-commit hook in the current repository.`,
	Run:   runInstall,
}

func init() {
	rootCmd.AddCommand(installCmd)
}

const hookScript = `#!/bin/sh
# PreReview - AI-powered code review before commits
# This hook was installed by prereview

# Run prereview in hook mode
prereview --hook

# Capture exit code
exit_code=$?

# If prereview fails, abort the commit
if [ $exit_code -ne 0 ]; then
    echo ""
    echo "Commit aborted by prereview."
    echo "Run 'prereview' manually to review and fix issues."
    exit 1
fi

exit 0
`

func runInstall(cmd *cobra.Command, args []string) {
	// Check if we're in a git repository
	if !git.IsGitRepo() {
		ui.Error("Not a git repository")
		os.Exit(1)
	}

	// Get git hooks directory
	gitDir, err := git.GetGitDir()
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to find .git directory: %v", err))
		os.Exit(1)
	}

	hooksDir := filepath.Join(gitDir, "hooks")
	hookPath := filepath.Join(hooksDir, "pre-commit")

	// Check if hooks directory exists
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		if err := os.MkdirAll(hooksDir, 0755); err != nil {
			ui.Error(fmt.Sprintf("Failed to create hooks directory: %v", err))
			os.Exit(1)
		}
	}

	// Check if pre-commit hook already exists
	if _, err := os.Stat(hookPath); err == nil {
		// Read existing hook to check if it's ours
		content, err := os.ReadFile(hookPath)
		if err != nil {
			ui.Error(fmt.Sprintf("Failed to read existing hook: %v", err))
			os.Exit(1)
		}
		if !strings.Contains(string(content), "# This hook was installed by prereview") {
			ui.Warning("A pre-commit hook already exists.")
			ui.Info("You can manually add prereview to your existing hook:")
			ui.Info("  prereview --hook")
			os.Exit(1)
		}
		ui.Info("Updating existing prereview hook...")
	}

	// Write hook script
	if err := os.WriteFile(hookPath, []byte(hookScript), 0755); err != nil {
		ui.Error(fmt.Sprintf("Failed to write hook: %v", err))
		os.Exit(1)
	}

	ui.Success("✓ Pre-commit hook installed successfully!")
	ui.Info("  PreReview will now run automatically before each commit.")
	ui.Info("  Run 'prereview uninstall' to remove the hook.")
}
