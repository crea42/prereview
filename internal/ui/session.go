package ui

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/emilushi/prereview/internal/review"
)

// Action represents the outcome action of a review session
type Action int

const (
	ActionCommit Action = iota
	ActionAbort
	ActionReReview
)

// SessionOutcome represents the result of a review session
type SessionOutcome struct {
	Action  Action
	Fixed   int
	Skipped int
}

// ReviewSession handles an interactive review session
type ReviewSession struct {
	result      *review.ReviewResult
	current     int
	fixed       int
	skipped     int
	suggestions []review.Suggestion
	skippedMap  map[int]bool
}

// NewReviewSession creates a new review session
func NewReviewSession(result *review.ReviewResult) *ReviewSession {
	return &ReviewSession{
		result:      result,
		suggestions: result.Suggestions,
		skippedMap:  make(map[int]bool),
	}
}

// Run starts the interactive review session
func (s *ReviewSession) Run() SessionOutcome {
	total := len(s.suggestions)

	if total == 0 {
		return SessionOutcome{Action: ActionCommit}
	}

	reader := bufio.NewReader(os.Stdin)

	for s.current < total {
		suggestion := s.suggestions[s.current]

		// Print suggestion
		s.printSuggestion(suggestion, s.current+1, total)

		// Get user input
		fmt.Print("\n  " + Option("f") + "ix | " + Option("s") + "kip | " + Option("v") + "iew diff | " + Option("q") + "uit: ")

		input, err := reader.ReadString('\n')
		if err != nil {
			Error("Failed to read input")
			s.current++
			continue
		}

		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "f", "fix":
			if s.applyFix(suggestion) {
				s.fixed++
				Success("  ✓ Applied fix")
			} else {
				Warning("  ⚠ Could not apply fix automatically")
				fmt.Print("  Skip this suggestion? [y/n]: ")
				confirm, _ := reader.ReadString('\n')
				if strings.TrimSpace(strings.ToLower(confirm)) == "y" {
					s.skipped++
					s.skippedMap[s.current] = true
				} else {
					continue // Stay on current suggestion
				}
			}
			s.current++

		case "s", "skip":
			s.skipped++
			s.skippedMap[s.current] = true
			Muted("  ⏭ Skipped")
			s.current++

		case "v", "view":
			s.viewDiff(suggestion)
			// Don't advance, let user decide

		case "q", "quit":
			return SessionOutcome{
				Action:  ActionAbort,
				Fixed:   s.fixed,
				Skipped: s.skipped,
			}

		default:
			Muted("  Invalid option. Use f, s, v, or q.")
		}

		fmt.Println()
	}

	// Show summary
	s.printSummary()

	// Ask what to do
	fmt.Print("\nProceed with commit? " + Option("y") + "es | " + Option("n") + "o | " + Option("r") + "e-review: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	switch input {
	case "y", "yes":
		return SessionOutcome{
			Action:  ActionCommit,
			Fixed:   s.fixed,
			Skipped: s.skipped,
		}
	case "r", "re-review":
		return SessionOutcome{
			Action:  ActionReReview,
			Fixed:   s.fixed,
			Skipped: s.skipped,
		}
	default:
		return SessionOutcome{
			Action:  ActionAbort,
			Fixed:   s.fixed,
			Skipped: s.skipped,
		}
	}
}

// printSuggestion displays a suggestion
func (s *ReviewSession) printSuggestion(sug review.Suggestion, num, total int) {
	Divider()

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color("#374151")).
		Foreground(lipgloss.Color("#F9FAFB")).
		Padding(0, 1)

	header := fmt.Sprintf("📄 %s [%d/%d]", sug.File, num, total)
	fmt.Println(headerStyle.Render(header))

	Divider()

	// Location
	if sug.Line > 0 {
		locationStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF"))
		location := fmt.Sprintf("Line %d", sug.Line)
		if sug.EndLine > sug.Line {
			location = fmt.Sprintf("Lines %d-%d", sug.Line, sug.EndLine)
		}
		fmt.Println(locationStyle.Render("  " + location))
	}

	// Severity and title
	sevStyle := SeverityStyle(string(sug.Severity))
	icon := SeverityIcon(string(sug.Severity))
	fmt.Println()
	fmt.Println(sevStyle.Render("  " + icon + " " + sug.Title))

	// Description
	if sug.Description != "" {
		descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#D1D5DB"))
		fmt.Println()
		fmt.Println(descStyle.Render("  " + sug.Description))
	}

	// Suggested fix
	if sug.SuggestFix != "" {
		fmt.Println()
		fixLabelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Bold(true)
		fmt.Println(fixLabelStyle.Render("  Suggested fix:"))
		codeBlockStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("#1F2937")).
			Foreground(lipgloss.Color("#A7F3D0")).
			Padding(0, 1).
			MarginLeft(2)
		fmt.Println(codeBlockStyle.Render(sug.SuggestFix))
	}

	// Category badge
	if sug.Category != "" {
		badgeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			Background(lipgloss.Color("#374151")).
			Padding(0, 1)
		fmt.Println()
		fmt.Println("  " + badgeStyle.Render(sug.Category))
	}
}

// applyFix applies a suggested fix
func (s *ReviewSession) applyFix(sug review.Suggestion) bool {
	// Check if we have both original and fix code
	if sug.SuggestFix == "" || sug.SuggestFix == "N/A" {
		return false
	}
	if sug.OriginalCode == "" || sug.OriginalCode == "N/A" {
		return false
	}

	// Read the file
	content, err := os.ReadFile(sug.File)
	if err != nil {
		return false
	}

	fileContent := string(content)

	// Try to find and replace the original code
	if !strings.Contains(fileContent, sug.OriginalCode) {
		return false
	}

	// Replace the original code with the fix
	newContent := strings.Replace(fileContent, sug.OriginalCode, sug.SuggestFix, 1)

	// Check if replacement actually happened
	if newContent == fileContent {
		return false
	}

	// Write the file back (preserving original permissions)
	fileInfo, _ := os.Stat(sug.File)
	perm := os.FileMode(0644)
	if fileInfo != nil {
		perm = fileInfo.Mode().Perm()
	}
	if err := os.WriteFile(sug.File, []byte(newContent), perm); err != nil {
		return false
	}

	// Stage the change
	cmd := exec.Command("git", "add", sug.File)
	if err := cmd.Run(); err != nil {
		// File was modified but not staged - still consider it a success
		Warning("  File modified but could not stage: " + err.Error())
	}

	return true
}

// viewDiff shows the diff for a file
func (s *ReviewSession) viewDiff(sug review.Suggestion) {
	fmt.Println()

	// Get the staged diff for the file
	cmd := exec.Command("git", "diff", "--cached", "--color=always", "--", sug.File)
	output, err := cmd.Output()
	if err != nil {
		Muted("  Could not retrieve diff: " + err.Error())
		fmt.Println()
		return
	}

	if len(output) == 0 {
		Muted("  No staged changes for this file")
		fmt.Println()
		return
	}

	// Print header
	diffHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#60A5FA"))
	fmt.Println("  " + diffHeaderStyle.Render("Diff for "+sug.File))
	fmt.Println()

	// Print the diff with some indentation
	lines := strings.Split(string(output), "\n")
	const maxLines = 50 // Limit output to avoid overwhelming the terminal
	
	for i, line := range lines {
		if i >= maxLines {
			Muted(fmt.Sprintf("  ... (%d more lines, use 'git diff --cached %s' to see full diff)", len(lines)-maxLines, sug.File))
			break
		}
		fmt.Println("  " + line)
	}
	fmt.Println()
}

// printSummary shows the review summary
func (s *ReviewSession) printSummary() {
	Divider()

	summaryStyle := lipgloss.NewStyle().Bold(true)
	fmt.Println(summaryStyle.Render("Summary"))

	total := len(s.suggestions)
	remaining := total - s.fixed - s.skipped

	fmt.Printf("  %s %d fixed\n", successStyle.Render("✓"), s.fixed)
	fmt.Printf("  %s %d skipped\n", warningStyle.Render("⏭"), s.skipped)
	if remaining > 0 {
		fmt.Printf("  %s %d remaining\n", errorStyle.Render("•"), remaining)
	}

	Divider()
}
