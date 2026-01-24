# PreReview

AI-powered code review before you commit. Uses the GitHub Copilot SDK to analyze your staged changes and provide actionable suggestions.

![PreReview Demo](https://vhs.charm.sh/vhs-5kpVg908APYT1zeSPpN0vL.gif)

## Features

- 🔍 **AI-Powered Review** - Uses GitHub Copilot SDK (Claude, GPT-5, Gemini, Grok) to review your code
- 🎯 **Interactive Workflow** - Fix, skip, or review suggestions one by one
- 🪝 **Git Hook Integration** - Automatically runs before each commit
- ⚙️ **Configurable** - Customize models, strictness, and ignore patterns
- 🎨 **Beautiful Terminal UI** - Clean, colorful output with lipgloss

## Prerequisites

- **GitHub Copilot subscription** (Free, Pro, Business, or Enterprise)
- **Copilot CLI** installed:

  ```bash
  brew install copilot-cli
  ```

- Logged in to GitHub:

  ```bash
  copilot auth login
  ```

## Installation

### Using Go

```bash
go install github.com/emilushi/prereview@latest
```

### From Source

```bash
git clone https://github.com/emilushi/prereview.git
cd prereview
go build -o prereview .
sudo mv prereview /usr/local/bin/
```

## Quick Start

1. **Review staged changes:**

   ```bash
   git add .
   prereview
   ```

2. **Install as pre-commit hook:**

   ```bash
   prereview install
   ```

3. **Create config file:**

   ```bash
   prereview config init
   ```

## Usage

### Commands

| Command                              | Description                         |
|--------------------------------------|-------------------------------------|
| `prereview`                          | Review staged changes interactively |
| `prereview review`                   | Same as above                       |
| `prereview install`                  | Install as git pre-commit hook      |
| `prereview uninstall`                | Remove git pre-commit hook          |
| `prereview config init`              | Create default config file          |
| `prereview config list`              | Show current configuration          |
| `prereview config set <key> <value>` | Set a config value                  |
| `prereview config get <key>`         | Get a config value                  |

### Flags

| Flag          | Description                                                           |
|---------------|-----------------------------------------------------------------------|
| `--model`     | AI model to use (claude, gpt-4, gemini, grok)                         |
| `--strict`    | Require all issues to be fixed before committing                      |
| `--tolerance` | Review tolerance: `strict`, `moderate`, `relaxed` (default: moderate) |
| `--force`     | Force commit even with unresolved suggestions                         |
| `--verbose`   | Show detailed output                                                  |
| `--config`    | Path to config file                                                   |

### Tolerance Levels

PreReview supports three tolerance levels to reduce false positives:

| Level      | Description                                                                |
|------------|----------------------------------------------------------------------------|
| `strict`   | Reports all potential issues including style nitpicks                      |
| `moderate` | Reports bugs, security issues, and significant quality concerns (default)  |
| `relaxed`  | Only reports definite bugs and critical security issues                    |

```bash
# Use relaxed mode for less noise
prereview --tolerance relaxed

# Or set in config
prereview config set tolerance relaxed
```

### Confidence Levels

Each suggestion includes a confidence level:

- **high** - Definite issue, should be fixed (>95% confident)
- **medium** - Likely an issue but could be intentional (70-95% confident)
- **low** - Possible issue, may be false positive (<70% confident)

**Only high-confidence errors block commits by default.** Low-confidence suggestions are shown but don't prevent you from committing.

### Interactive Review

When reviewing, you have these options for each suggestion:

- `[f]ix`  - Apply the suggested fix (if available)
- `[s]kip` - Skip this suggestion
- `[v]iew` - View the diff for context
- `[q]uit` - Abort the review

After reviewing all suggestions:

- `[y]es`       - Proceed with commit
- `[n]o`        - Abort
- `[r]e-review` - Review again (after making manual fixes)

## Configuration

Create a `.prereviewrc.yaml` in your project root or `~/.prereviewrc.yaml` for global settings:

```yaml
# AI model to use
model: gpt-4

# Require all issues to be fixed
strict: false

# Review tolerance: strict, moderate, relaxed
# - strict: Report all potential issues
# - moderate: Report bugs and significant issues (default)
# - relaxed: Only report definite bugs and security issues
tolerance: moderate

# What severity level blocks commits: errors, warnings, all, none
# Default: errors (only high-confidence errors block)
block_on: errors

# Show detailed output
verbose: false

# Files to ignore
ignore_patterns:
  - "*.min.js"
  - "vendor/*"
  - "node_modules/*"

# Max file size to review (bytes)
max_file_size: 100000

# Additional coding standard files to detect (beyond auto-detected ones)
# These are file paths relative to repo root
coding_standards:
  - ".custom-lint.json"
  - "config/phpcs-custom.xml"

# Project-specific hints for the AI reviewer
# Use this to provide context that reduces false positives
project_hints:
  - "Factory methods may return different subtypes based on input"
  - "Data is sanitized at input time, not output time"
  - "We use dependency injection throughout the codebase"
```

### Reducing False Positives

If you're experiencing too many false positives:

1. **Use relaxed tolerance:**

   ```bash
   prereview config set tolerance relaxed
   ```

2. **Add project-specific context:**

   ```yaml
   project_hints:
     - "This project uses the Repository pattern"
     - "getEntity() returns different entity types based on the ID prefix"
     - "All user input is sanitized before storage"
   ```

3. **Change what blocks commits:**

   ```yaml
   block_on: errors  # Only block on high-confidence errors
   ```

4. **Force commit when needed:**

   ```bash
   git commit  # Will prompt: "Proceed despite issues? [y/N]"
   # Or bypass completely:
   prereview --force
   git commit --no-verify
   ```

## Authentication

PreReview uses the **GitHub Copilot CLI** for authentication. The CLI handles all authentication automatically using your GitHub Copilot subscription.

### Setup

1. **Install Copilot CLI:**

   ```bash
   brew install copilot-cli
   ```

2. **Login to GitHub:**

   ```bash
   copilot auth login
   ```

3. **Verify authentication:**

   ```bash
   copilot auth status
   ```

That's it! PreReview will use your Copilot CLI credentials automatically.

## Examples

### Basic Review

```bash
$ git add src/utils.js
$ prereview

🔍 Reviewing 1 changed file(s)...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📄 src/utils.js [1/2]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Line 45

  ⚠ Missing null check

  Accessing property on potentially null value.

  Suggested fix:
  if (user?.email) { ... }

  best-practice

  [f]ix | [s]kip | [v]iew diff | [q]uit: f

  ✓ Applied fix

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Summary
  ✓ 1 fixed
  ⏭ 0 skipped
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Proceed with commit? [y]es | [n]o | [r]e-review: y

✓ Review complete: 1 fixed, 0 skipped
```

### Strict Mode

```bash
prereview --strict
```

In strict mode, you cannot proceed with commit if any suggestions are skipped.

## Development

### Requirements

- Go 1.21+
- GitHub Copilot license

### Building

```bash
go build -o prereview .
```

### Testing

```bash
go test ./...
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests
5. Submit a pull request

## Acknowledgments

- [GitHub Copilot](https://github.com/features/copilot) for AI capabilities
- [Cobra](https://github.com/spf13/cobra) for CLI framework
- [Viper](https://github.com/spf13/viper) for configuration
- [Lipgloss](https://github.com/charmbracelet/lipgloss) for terminal styling
