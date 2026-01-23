# PreReview - AI Code Review Tool

## Project Overview
PreReview is a Go-based CLI tool that provides AI-powered code review before git commits using the GitHub Copilot SDK.

## Tech Stack
- **Language**: Go 1.21+
- **CLI Framework**: Cobra
- **Configuration**: Viper
- **Terminal UI**: Lipgloss
- **AI Backend**: GitHub Copilot SDK

## Project Structure
```
prereview/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command and flags
│   ├── review.go          # Review command (default)
│   ├── install.go         # Install git pre-commit hook
│   ├── uninstall.go       # Uninstall git pre-commit hook
│   └── config.go          # Config init/list/get/set commands
├── internal/
│   ├── copilot/           # Copilot SDK client wrapper
│   │   └── client.go      # Chat API, model mapping
│   ├── git/               # Git operations
│   │   └── git.go         # Staged files, diffs, file staging
│   ├── output/            # Output formatting
│   │   └── markdown.go    # Markdown suggestions file generation
│   ├── review/            # Core review logic
│   │   └── review.go      # AI prompts, suggestion parsing
│   ├── standards/         # Coding standards detection
│   │   └── detector.go    # Detect project linters/formatters
│   └── ui/                # Terminal UI components
│       ├── session.go     # Interactive review session
│       └── styles.go      # Lipgloss styles and colors
├── go.mod
├── go.sum
├── main.go
├── LICENSE                # MIT License
└── README.md
```

## Commands
- `prereview` - Review staged changes interactively
- `prereview review` - Same as above (explicit)
- `prereview review --markdown` - Output suggestions to markdown file
- `prereview install` - Install as git pre-commit hook
- `prereview uninstall` - Remove git pre-commit hook  
- `prereview config init` - Create default .prereviewrc.yaml
- `prereview config list` - Show current configuration
- `prereview config get <key>` - Get a config value
- `prereview config set <key> <value>` - Set a config value

## Flags
- `--model` - AI model (claude, gpt-4, gpt-5, gemini, grok)
- `--strict` - Require all issues fixed before commit
- `--verbose` - Show detailed output
- `--config` - Path to config file
- `--markdown` - Output suggestions to markdown file (review command)

## Development
```bash
go build -o prereview .
./prereview --help
```

## Testing
```bash
# Review staged changes
git add .
./prereview --model claude

# Generate markdown output
./prereview review --markdown
```
