package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "prereview",
	Short: "AI-powered code review before commits",
	Long: `PreReview is a CLI tool that provides AI-powered code review
before you commit your changes. It uses the GitHub Copilot SDK
to analyze staged changes and provide suggestions.

Run without arguments to review staged changes:
  prereview

Install as a git pre-commit hook:
  prereview install

Configure settings:
  prereview config`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default behavior: run review
		reviewCmd.Run(cmd, args)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .prereviewrc.yaml)")
	rootCmd.PersistentFlags().String("model", "", "AI model to use (claude, gpt-5, gpt-4, gemini, grok)")
	rootCmd.PersistentFlags().Bool("strict", false, "Require all issues to be fixed before committing")
	rootCmd.PersistentFlags().Bool("verbose", false, "Show detailed output")
	rootCmd.PersistentFlags().Bool("hook", false, "Run in pre-commit hook mode (non-interactive, exits with error if issues found)")

	_ = viper.BindPFlag("model", rootCmd.PersistentFlags().Lookup("model"))
	_ = viper.BindPFlag("strict", rootCmd.PersistentFlags().Lookup("strict"))
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("hook", rootCmd.PersistentFlags().Lookup("hook"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Look for config in current directory first
		viper.AddConfigPath(".")
		// Then home directory
		home, err := os.UserHomeDir()
		if err == nil {
			viper.AddConfigPath(home)
		}
		viper.SetConfigType("yaml")
		viper.SetConfigName(".prereviewrc")
	}

	// Set defaults
	viper.SetDefault("model", "gpt-4o-mini")
	viper.SetDefault("strict", false)
	viper.SetDefault("verbose", false)
	viper.SetDefault("ignore_patterns", []string{})
	viper.SetDefault("max_file_size", 100000) // 100KB

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	} else if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
		// Config file found but another error occurred (e.g., parse error)
		if viper.GetBool("verbose") {
			fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
		}
	}
}
