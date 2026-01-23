package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/emilushi/prereview/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View or modify configuration",
	Long:  `View or modify prereview configuration settings.`,
	Run:   runConfig,
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	Run:   runConfigSet,
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	Run:   runConfigGet,
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	Run:   runConfigList,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a default configuration file",
	Run:   runConfigInit,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configInitCmd)
}

func runConfig(cmd *cobra.Command, args []string) {
	// Default: show help. Error can be safely ignored as Help() only fails
	// if writing to output fails, which cobra handles internally.
	_ = cmd.Help()
}

func runConfigSet(cmd *cobra.Command, args []string) {
	key := args[0]
	value := args[1]

	viper.Set(key, value)

	// Write config
	configPath := viper.ConfigFileUsed()
	if configPath == "" {
		configPath = ".prereviewrc.yaml"
	}

	if err := viper.SafeWriteConfigAs(configPath); err != nil {
		// If file exists, try regular WriteConfigAs
		if os.IsExist(err) {
			if err := viper.WriteConfigAs(configPath); err != nil {
				ui.Error(fmt.Sprintf("Failed to write config: %v", err))
				os.Exit(1)
			}
		} else {
			ui.Error(fmt.Sprintf("Failed to write config: %v", err))
			os.Exit(1)
		}
	}

	ui.Success(fmt.Sprintf("✓ Set %s = %s", key, value))
}

func runConfigGet(cmd *cobra.Command, args []string) {
	key := args[0]
	value := viper.Get(key)

	if value == nil {
		ui.Info(fmt.Sprintf("%s: (not set)", key))
	} else {
		ui.Info(fmt.Sprintf("%s: %v", key, value))
	}
}

func runConfigList(cmd *cobra.Command, args []string) {
	settings := viper.AllSettings()

	if len(settings) == 0 {
		ui.Info("No configuration set. Using defaults.")
		return
	}

	ui.Info("Current configuration:")
	keys := make([]string, 0, len(settings))
	for key := range settings {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("  %s: %v\n", key, settings[key])
	}

	if configFile := viper.ConfigFileUsed(); configFile != "" {
		fmt.Printf("\nConfig file: %s\n", configFile)
	}
}

func runConfigInit(cmd *cobra.Command, args []string) {
	configPath := ".prereviewrc.yaml"

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		ui.Warning("Configuration file already exists: " + configPath)
		return
	}

	defaultConfig := `# PreReview Configuration
# https://github.com/emilushi/prereview

# AI model to use: claude, gpt-5, gpt-4, gemini, grok (default: gpt-4o-mini)
model: gpt-4

# Require all issues to be fixed before committing
strict: false

# Show detailed output
verbose: false

# Output suggestions to markdown file instead of interactive terminal
# When enabled, generates suggestions_<commit_hash>.md in project root
output_markdown: false

# File patterns to ignore (glob patterns)
ignore_patterns:
  - "*.min.js"
  - "*.min.css"
  - "vendor/*"
  - "node_modules/*"
  - "*.lock"
  - "go.sum"

# Maximum file size to review (in bytes)
max_file_size: 100000

# Coding standards configuration files to use for review context
# PreReview auto-detects common files like .eslintrc, phpcs.xml, etc.
# Add custom paths here for additional standards files
# coding_standards:
#   - ".custom-lint-rules.json"
#   - "config/coding-standards.yaml"

# Review focus areas (optional)
# focus:
#   - security
#   - performance
#   - best-practices
`

	if err := os.WriteFile(configPath, []byte(defaultConfig), 0600); err != nil {
		ui.Error(fmt.Sprintf("Failed to create config file: %v", err))
		os.Exit(1)
	}

	ui.Success("✓ Created configuration file: " + configPath)
}
