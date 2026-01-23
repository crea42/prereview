# PreReview - AI Code Review Tool

## Project Overview
PreReview is a Go-based CLI tool that provides AI-powered code review before git commits using the GitHub Copilot SDK.

## Tech Stack
- **Language**: Go 1.21+
- **CLI Framework**: Cobra
- **Configuration**: Viper
- **Terminal UI**: Lipgloss + Bubbletea
- **AI Backend**: GitHub Copilot SDK

## Project Structure
```
prereview/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command
│   ├── review.go          # Review command (default)
│   ├── install.go         # Install hook command
│   ├── uninstall.go       # Uninstall hook command
│   └── config.go          # Config command
├── internal/
│   ├── git/               # Git operations
│   ├── review/            # Review logic
│   ├── copilot/           # Copilot SDK integration
│   ├── ui/                # Terminal UI components
│   └── config/            # Configuration management
├── .prereviewrc.yaml      # Example config
├── go.mod
├── go.sum
└── main.go
```

## Commands
- `prereview` - Review staged changes interactively
- `prereview install` - Install as git pre-commit hook
- `prereview uninstall` - Remove git pre-commit hook  
- `prereview config` - Configure settings

## Development
```bash
go build -o prereview .
./prereview --help
```
