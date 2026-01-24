package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	primaryColor   = lipgloss.Color("#7C3AED")
	successColor   = lipgloss.Color("#10B981")
	warningColor   = lipgloss.Color("#F59E0B")
	errorColor     = lipgloss.Color("#EF4444")
	infoColor      = lipgloss.Color("#3B82F6")
	mutedColor     = lipgloss.Color("#6B7280")

	// Styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	successStyle = lipgloss.NewStyle().
			Foreground(successColor)

	warningStyle = lipgloss.NewStyle().
			Foreground(warningColor)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor)

	infoStyle = lipgloss.NewStyle().
			Foreground(infoColor)

	mutedStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	fileStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#60A5FA"))

	lineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A78BFA"))

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4B5563")).
			Padding(0, 1)

	suggestionHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Background(lipgloss.Color("#1F2937")).
				Foreground(lipgloss.Color("#F9FAFB")).
				Padding(0, 1).
				Width(60)

	codeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1F2937")).
			Foreground(lipgloss.Color("#D1D5DB")).
			Padding(0, 1)

	optionStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)
)

// Title prints a styled title
func Title(text string) {
	fmt.Println(titleStyle.Render(text))
}

// Success prints a success message
func Success(text string) {
	fmt.Println(successStyle.Render(text))
}

// Warning prints a warning message
func Warning(text string) {
	fmt.Println(warningStyle.Render(text))
}

// Error prints an error message
func Error(text string) {
	fmt.Fprintln(os.Stderr, errorStyle.Render(text))
}

// Info prints an info message
func Info(text string) {
	fmt.Println(infoStyle.Render(text))
}

// Muted prints muted text
func Muted(text string) {
	fmt.Println(mutedStyle.Render(text))
}

// File formats a file path
func File(path string) string {
	return fileStyle.Render(path)
}

// Line formats a line number
func Line(num int) string {
	return lineStyle.Render(fmt.Sprintf("L%d", num))
}

// Box wraps content in a box
func Box(content string) string {
	return boxStyle.Render(content)
}

// Code formats code
func Code(content string) string {
	return codeStyle.Render(content)
}

// Option formats an option key
func Option(key string) string {
	return optionStyle.Render("[" + key + "]")
}

// SeverityStyle returns the style for a severity level
func SeverityStyle(severity string) lipgloss.Style {
	switch severity {
	case "error":
		return errorStyle
	case "warning":
		return warningStyle
	case "info":
		return infoStyle
	default:
		return mutedStyle
	}
}

// SeverityIcon returns an icon for a severity level
func SeverityIcon(severity string) string {
	switch severity {
	case "error":
		return "✗"
	case "warning":
		return "⚠"
	case "info":
		return "ℹ"
	default:
		return "•"
	}
}

// Divider prints a horizontal divider
func Divider() {
	fmt.Println(mutedStyle.Render(strings.Repeat("━", 60)))
}

// SuccessIcon returns a green checkmark
func SuccessIcon() string {
	return successStyle.Render("✓")
}

// ErrorIcon returns a red X
func ErrorIcon() string {
	return errorStyle.Render("✗")
}

// WarningIcon returns a yellow warning symbol
func WarningIcon() string {
	return warningStyle.Render("⚠")
}

// InfoIcon returns a blue info symbol
func InfoIcon() string {
	return infoStyle.Render("ℹ")
}
