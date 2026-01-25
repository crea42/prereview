# Contributing to PreReview

Thanks for considering contributing! We welcome improvements, bug fixes, and new features.

## Quick Start

1. Fork the repo and clone your fork.
2. Create a branch: `git checkout -b your-feature`.
3. Run tests: `go test ./...`.
4. Commit with a clear message and push.
5. Open a Pull Request.

## Development Setup

- Go 1.21+
- GitHub Copilot CLI installed and authenticated (see README)
- Optional: `gofumpt` / `golangci-lint` if you want to lint locally

## Making Changes

- Keep PRs focused and small when possible.
- Add or update tests for behavior changes.
- Update documentation (README/usage) if behavior changes.
- Follow existing code style; use `gofmt` (Go’s standard formatting).

## Commit Messages

- Use clear, single-line messages (e.g., `fix: handle nil diff` or `feat: add doctor command`).

## Pull Requests

- Fill out the PR template.
- Describe the change and rationale.
- Note testing performed (`go test ./...`, manual steps).

## Reporting Issues

- Use the issue template.
- Include reproduction steps, expected vs actual, environment (OS, Go version).

## Code of Conduct

This project follows the Contributor Covenant. By participating, you agree to uphold it. See CODE_OF_CONDUCT.md.
