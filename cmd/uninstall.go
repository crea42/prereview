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

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove prereview git pre-commit hook",
	Long:  `Remove the prereview pre-commit hook from the current repository.`,
	Run:   runUninstall,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}

func runUninstall(cmd *cobra.Command, args []string) {
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

	hookPath := filepath.Join(gitDir, "hooks", "pre-commit")

	// Check if hook exists
	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		ui.Info("No pre-commit hook found")
		return
	}

	// Read hook to verify it's ours
	content, err := os.ReadFile(hookPath)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to read hook: %v", err))
		os.Exit(1)
	}

	if !strings.Contains(string(content), "# This hook was installed by prereview") {
		ui.Warning("The pre-commit hook was not installed by prereview")
		ui.Info("Not removing to avoid breaking your existing hook")
		os.Exit(1)
	}

	// Remove the hook
	if err := os.Remove(hookPath); err != nil {
		ui.Error(fmt.Sprintf("Failed to remove hook: %v", err))
		os.Exit(1)
	}

	ui.Success("✓ Pre-commit hook removed successfully!")
}
