package standards

import (
	"os"
	"path/filepath"
	"strings"
)

// CodingStandard represents a detected or configured coding standard
type CodingStandard struct {
	Name        string   // e.g., "ESLint", "PHPCS", "WordPress"
	Type        string   // e.g., "linter", "framework", "style-guide"
	ConfigFile  string   // Path to the config file
	Description string   // Brief description of the standard
	Rules       []string // Key rules extracted from config (optional)
}

// StandardsDetector detects coding standards in a project
type StandardsDetector struct {
	repoRoot        string
	customStandards []string // User-configured standard files
}

// KnownStandards maps config files to their coding standards
var KnownStandards = map[string]CodingStandard{
	// JavaScript/TypeScript
	".eslintrc":          {Name: "ESLint", Type: "linter", Description: "JavaScript/TypeScript linting rules"},
	".eslintrc.js":       {Name: "ESLint", Type: "linter", Description: "JavaScript/TypeScript linting rules"},
	".eslintrc.json":     {Name: "ESLint", Type: "linter", Description: "JavaScript/TypeScript linting rules"},
	".eslintrc.yaml":     {Name: "ESLint", Type: "linter", Description: "JavaScript/TypeScript linting rules"},
	".eslintrc.yml":      {Name: "ESLint", Type: "linter", Description: "JavaScript/TypeScript linting rules"},
	"eslint.config.js":   {Name: "ESLint", Type: "linter", Description: "JavaScript/TypeScript linting rules (flat config)"},
	"eslint.config.mjs":  {Name: "ESLint", Type: "linter", Description: "JavaScript/TypeScript linting rules (flat config)"},
	".prettierrc":        {Name: "Prettier", Type: "formatter", Description: "Code formatting rules"},
	".prettierrc.json":   {Name: "Prettier", Type: "formatter", Description: "Code formatting rules"},
	".prettierrc.js":     {Name: "Prettier", Type: "formatter", Description: "Code formatting rules"},
	"prettier.config.js": {Name: "Prettier", Type: "formatter", Description: "Code formatting rules"},
	"biome.json":         {Name: "Biome", Type: "linter", Description: "JavaScript/TypeScript linting and formatting"},
	"biome.jsonc":        {Name: "Biome", Type: "linter", Description: "JavaScript/TypeScript linting and formatting"},

	// PHP
	"phpcs.xml":              {Name: "PHP_CodeSniffer", Type: "linter", Description: "PHP coding standards"},
	"phpcs.xml.dist":         {Name: "PHP_CodeSniffer", Type: "linter", Description: "PHP coding standards"},
	".php-cs-fixer.php":      {Name: "PHP-CS-Fixer", Type: "formatter", Description: "PHP code style fixer"},
	".php-cs-fixer.dist.php": {Name: "PHP-CS-Fixer", Type: "formatter", Description: "PHP code style fixer"},
	"phpstan.neon":           {Name: "PHPStan", Type: "analyzer", Description: "PHP static analysis"},
	"phpstan.neon.dist":      {Name: "PHPStan", Type: "analyzer", Description: "PHP static analysis"},
	"psalm.xml":              {Name: "Psalm", Type: "analyzer", Description: "PHP static analysis"},

	// Python
	"pyproject.toml": {Name: "Python Project", Type: "config", Description: "Python project configuration (may include ruff, black, mypy settings)"},
	".flake8":        {Name: "Flake8", Type: "linter", Description: "Python style guide enforcement"},
	".pylintrc":      {Name: "Pylint", Type: "linter", Description: "Python code analysis"},
	"pylintrc":       {Name: "Pylint", Type: "linter", Description: "Python code analysis"},
	".mypy.ini":      {Name: "Mypy", Type: "type-checker", Description: "Python static type checker"},
	"mypy.ini":       {Name: "Mypy", Type: "type-checker", Description: "Python static type checker"},
	"ruff.toml":      {Name: "Ruff", Type: "linter", Description: "Fast Python linter"},
	".ruff.toml":     {Name: "Ruff", Type: "linter", Description: "Fast Python linter"},

	// Go
	".golangci.yml":  {Name: "golangci-lint", Type: "linter", Description: "Go linters aggregator"},
	".golangci.yaml": {Name: "golangci-lint", Type: "linter", Description: "Go linters aggregator"},
	"golangci.yml":   {Name: "golangci-lint", Type: "linter", Description: "Go linters aggregator"},

	// Ruby
	".rubocop.yml":  {Name: "RuboCop", Type: "linter", Description: "Ruby static code analyzer"},
	".rubocop.yaml": {Name: "RuboCop", Type: "linter", Description: "Ruby static code analyzer"},

	// Rust
	"rustfmt.toml":  {Name: "rustfmt", Type: "formatter", Description: "Rust code formatting"},
	".rustfmt.toml": {Name: "rustfmt", Type: "formatter", Description: "Rust code formatting"},
	"clippy.toml":   {Name: "Clippy", Type: "linter", Description: "Rust linting"},

	// Editor configs
	".editorconfig": {Name: "EditorConfig", Type: "editor", Description: "Cross-editor coding style settings"},

	// StyleLint (CSS)
	".stylelintrc":        {Name: "Stylelint", Type: "linter", Description: "CSS/SCSS linting rules"},
	".stylelintrc.json":   {Name: "Stylelint", Type: "linter", Description: "CSS/SCSS linting rules"},
	"stylelint.config.js": {Name: "Stylelint", Type: "linter", Description: "CSS/SCSS linting rules"},
}

// FrameworkIndicators maps files/patterns to framework-specific standards
var FrameworkIndicators = map[string]CodingStandard{
	"wp-config.php":    {Name: "WordPress", Type: "framework", Description: "WordPress Coding Standards (WPCS)"},
	"wp-content":       {Name: "WordPress", Type: "framework", Description: "WordPress Coding Standards (WPCS)"},
	"artisan":          {Name: "Laravel", Type: "framework", Description: "Laravel/PSR coding standards"},
	"composer.json":    {Name: "Composer/PHP", Type: "package-manager", Description: "PHP project - check for framework-specific standards"},
	"package.json":     {Name: "Node.js", Type: "package-manager", Description: "Node.js project - check for eslint/prettier configs"},
	"Gemfile":          {Name: "Ruby/Rails", Type: "package-manager", Description: "Ruby project - check for rubocop config"},
	"Cargo.toml":       {Name: "Rust", Type: "package-manager", Description: "Rust project - rustfmt and clippy standards"},
	"go.mod":           {Name: "Go", Type: "package-manager", Description: "Go project - gofmt and effective go standards"},
	"requirements.txt": {Name: "Python", Type: "package-manager", Description: "Python project - PEP8 standards"},
	"setup.py":         {Name: "Python", Type: "package-manager", Description: "Python project - PEP8 standards"},
}

// NewStandardsDetector creates a new standards detector
func NewStandardsDetector(repoRoot string, customStandards []string) *StandardsDetector {
	return &StandardsDetector{
		repoRoot:        repoRoot,
		customStandards: customStandards,
	}
}

// DetectStandards scans the project for coding standards configurations
func (d *StandardsDetector) DetectStandards() []CodingStandard {
	var standards []CodingStandard
	seen := make(map[string]bool)

	// First, check custom standards specified by user
	for _, customFile := range d.customStandards {
		path := filepath.Join(d.repoRoot, customFile)
		if _, err := os.Stat(path); err == nil {
			std := CodingStandard{
				Name:        filepath.Base(customFile),
				Type:        "custom",
				ConfigFile:  customFile,
				Description: "User-specified coding standard configuration",
			}
			standards = append(standards, std)
			seen[std.Name] = true
		}
	}

	// Detect known standards config files
	for filename, standard := range KnownStandards {
		if seen[standard.Name] {
			continue
		}

		path := filepath.Join(d.repoRoot, filename)
		if _, err := os.Stat(path); err == nil {
			std := standard
			std.ConfigFile = filename
			standards = append(standards, std)
			seen[std.Name] = true
		}
	}

	// Detect framework indicators
	for indicator, framework := range FrameworkIndicators {
		if seen[framework.Name] {
			continue
		}

		path := filepath.Join(d.repoRoot, indicator)
		if _, err := os.Stat(path); err == nil {
			// For WordPress, also check if it's actually a WP project
			if framework.Name == "WordPress" {
				if d.isWordPressProject() {
					standards = append(standards, framework)
					seen[framework.Name] = true
				}
			} else {
				standards = append(standards, framework)
				seen[framework.Name] = true
			}
		}
	}

	return standards
}

// isWordPressProject checks if this is actually a WordPress project
func (d *StandardsDetector) isWordPressProject() bool {
	wpIndicators := []string{
		"wp-config.php",
		"wp-content",
		"wp-includes",
		"wp-admin",
		"style.css", // Theme
	}

	count := 0
	for _, indicator := range wpIndicators {
		path := filepath.Join(d.repoRoot, indicator)
		if _, err := os.Stat(path); err == nil {
			count++
		}
	}

	// Also check for theme/plugin structure
	if count == 0 {
		// Check if it's a theme (has style.css with Theme Name header)
		stylePath := filepath.Join(d.repoRoot, "style.css")
		if file, err := os.Open(stylePath); err == nil {
			header := make([]byte, 8192)
			n, _ := file.Read(header)
			file.Close()
			if strings.Contains(string(header[:n]), "Theme Name:") {
				return true
			}
		}

		// Check if it's a plugin (has main PHP file with Plugin Name header)
		files, err := filepath.Glob(filepath.Join(d.repoRoot, "*.php"))
		if err == nil {
			maxFilesToCheck := 50
			for i, filename := range files {
				if i >= maxFilesToCheck {
					break
				}
				if file, err := os.Open(filename); err == nil {
					header := make([]byte, 8192)
					n, _ := file.Read(header)
					file.Close()
					if strings.Contains(string(header[:n]), "Plugin Name:") {
						return true
					}
				}
			}
		}
	}

	return count >= 2
}

// GetStandardsContext returns a formatted string of detected standards for AI context
func (d *StandardsDetector) GetStandardsContext() string {
	standards := d.DetectStandards()
	if len(standards) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n\nCoding Standards Detected in Project:\n")

	for _, std := range standards {
		sb.WriteString("- ")
		sb.WriteString(std.Name)
		if std.ConfigFile != "" {
			sb.WriteString(" (")
			sb.WriteString(std.ConfigFile)
			sb.WriteString(")")
		}
		sb.WriteString(": ")
		sb.WriteString(std.Description)
		sb.WriteString("\n")
	}

	sb.WriteString("\nPlease ensure your code review suggestions align with these coding standards.\n")

	return sb.String()
}

// ReadConfigContent reads the content of a standards config file (for detailed context)
func (d *StandardsDetector) ReadConfigContent(configFile string, maxBytes int) (string, error) {
	if maxBytes == 0 {
		maxBytes = 2048 // Default max bytes to read
	}

	path := filepath.Join(d.repoRoot, filepath.Clean(configFile))
	// Ensure path stays within repoRoot to prevent path traversal
	if !strings.HasPrefix(path, d.repoRoot) {
		return "", os.ErrPermission
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	content := string(data)
	if len(content) > maxBytes {
		// Ensure we don't break UTF-8 encoding
		content = content[:maxBytes]
		for len(content) > 0 && content[len(content)-1] >= 0x80 && content[len(content)-1] < 0xC0 {
			content = content[:len(content)-1]
		}
		content += "\n... (truncated)"
	}

	return content, nil
}
