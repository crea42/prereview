package review

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/emilushi/prereview/internal/copilot"
	"github.com/emilushi/prereview/internal/git"
	"github.com/emilushi/prereview/internal/standards"
)

// Suggestion represents a code review suggestion
type Suggestion struct {
	File         string
	Line         int
	EndLine      int
	Severity     Severity
	Title        string
	Description  string
	OriginalCode string // Original code to be replaced
	SuggestFix   string // Suggested replacement code
	Category     string // security, performance, style, etc.
}

// Severity levels for suggestions
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
	SeverityHint    Severity = "hint"
)

// ReviewResult contains the results of a code review
type ReviewResult struct {
	Files       []string
	Suggestions []Suggestion
	Summary     string
}

// Reviewer handles code review using AI
type Reviewer struct {
	client           *copilot.Client
	model            string
	standardsContext string
}

// NewReviewer creates a new Reviewer instance
func NewReviewer(model string, repoRoot string, customStandards []string) (*Reviewer, error) {
	client, err := copilot.NewClient()
	if err != nil {
		return nil, err
	}

	// If no model specified, leave empty to let Copilot SDK pick the default
	// This may provide better model selection similar to VS Code's "auto" mode

	// Detect coding standards
	var standardsContext string
	if repoRoot != "" {
		detector := standards.NewStandardsDetector(repoRoot, customStandards)
		standardsContext = detector.GetStandardsContext()
	}

	return &Reviewer{
		client:           client,
		model:            model,
		standardsContext: standardsContext,
	}, nil
}

// Close stops the Copilot client
func (r *Reviewer) Close() {
	if r.client != nil {
		r.client.Close()
	}
}

// Review performs a code review on the given file changes
func (r *Reviewer) Review(changes []git.FileChange) (*ReviewResult, error) {
	result := &ReviewResult{
		Files: make([]string, 0, len(changes)),
	}

	for i, change := range changes {
		result.Files = append(result.Files, change.Path)

		if change.IsBinary {
			continue
		}

		// Show progress
		fmt.Printf("  [%d/%d] Reviewing %s...\n", i+1, len(changes), change.Path)

		// Request review from Copilot
		suggestions, err := r.reviewFile(change)
		if err != nil {
			// Show error to user but continue with other files
			fmt.Printf("    ✗ Error: %v\n", err)
			continue
		}

		if len(suggestions) > 0 {
			fmt.Printf("    ✓ Found %d suggestion(s)\n", len(suggestions))
		}

		result.Suggestions = append(result.Suggestions, suggestions...)
	}

	return result, nil
}

// reviewFile reviews a single file and returns suggestions
func (r *Reviewer) reviewFile(change git.FileChange) ([]Suggestion, error) {
	prompt := buildReviewPrompt(change, r.standardsContext)

	response, err := r.client.Chat(r.model, prompt)
	if err != nil {
		return nil, err
	}

	return parseReviewResponse(response, change.Path)
}

// buildReviewPrompt creates the prompt for code review
func buildReviewPrompt(change git.FileChange, standardsContext string) string {
	basePrompt := `You are a code reviewer. Review the following code changes and provide suggestions for improvements.

For each issue found, respond in this exact format:
---
LINE: <line number where issue starts>
END_LINE: <end line number if multi-line, otherwise same as LINE>
SEVERITY: <error|warning|info|hint>
CATEGORY: <security|performance|style|bug|best-practice>
TITLE: <short title>
DESCRIPTION: <detailed description>
ORIGINAL:
<<<
the exact original code lines copied verbatim from the file
include multiple lines if needed, preserving all whitespace and indentation
>>>
FIX:
<<<
the exact replacement code
include multiple lines if needed, preserving all whitespace and indentation
>>>
---

CRITICAL RULES for ORIGINAL and FIX:
1. ORIGINAL must be copied EXACTLY from the file content provided below - character for character
2. Include enough context (2-3 lines before/after) to make the match unique
3. Preserve ALL whitespace, tabs, and indentation exactly as they appear
4. For multi-line code, include all lines between <<< and >>>
5. If no code fix is applicable, use: N/A (without <<< >>>)

Focus on:
- Security vulnerabilities
- Performance issues
- Bug risks
- Code style and best practices
- Error handling
`

	// Add coding standards context if available
	if standardsContext != "" {
		basePrompt += standardsContext
	}

	basePrompt += `
If no issues are found, respond with: NO_ISSUES

File: ` + change.Path + `

Diff:
` + change.Diff + `

Full staged content:
` + change.Content

	return basePrompt
}

// parseReviewResponse parses the AI response into suggestions
func parseReviewResponse(response string, file string) ([]Suggestion, error) {
	var suggestions []Suggestion

	if response == "NO_ISSUES" || response == "" {
		return suggestions, nil
	}

	// Parse the structured response
	// This is a simplified parser - in production you'd want more robust parsing
	suggestions = parseStructuredResponse(response, file)

	return suggestions, nil
}

// parseStructuredResponse parses the structured AI response
func parseStructuredResponse(response string, file string) []Suggestion {
	var suggestions []Suggestion

	lines := splitLines(response)
	var current *Suggestion
	var multiLineField string // "ORIGINAL" or "FIX"
	var multiLineContent []string
	inMultiLine := false
	lastField := "" // Track last field we saw (ORIGINAL: or FIX:)

	for _, line := range lines {
		// Check for end of multi-line block
		if inMultiLine && line == ">>>" {
			content := joinLines(multiLineContent)
			if multiLineField == "ORIGINAL" {
				current.OriginalCode = content
			} else if multiLineField == "FIX" {
				current.SuggestFix = content
			}
			inMultiLine = false
			multiLineField = ""
			multiLineContent = nil
			continue
		}

		// Accumulate multi-line content
		if inMultiLine {
			multiLineContent = append(multiLineContent, line)
			continue
		}

		// Start of new suggestion
		if line == "---" {
			if current != nil && current.Title != "" {
				suggestions = append(suggestions, *current)
			}
			current = &Suggestion{File: file}
			lastField = ""
			continue
		}

		if current == nil {
			continue
		}

		// Handle "<<<" - start of multi-line block
		if line == "<<<" {
			multiLineField = lastField
			inMultiLine = true
			multiLineContent = nil
			continue
		}

		// Parse fields
		if hasPrefix(line, "LINE:") {
			current.Line = parseIntValue(line, "LINE:")
		} else if hasPrefix(line, "END_LINE:") {
			current.EndLine = parseIntValue(line, "END_LINE:")
		} else if hasPrefix(line, "SEVERITY:") {
			current.Severity = Severity(parseStringValue(line, "SEVERITY:"))
		} else if hasPrefix(line, "CATEGORY:") {
			current.Category = parseStringValue(line, "CATEGORY:")
		} else if hasPrefix(line, "TITLE:") {
			current.Title = parseStringValue(line, "TITLE:")
		} else if hasPrefix(line, "DESCRIPTION:") {
			current.Description = parseStringValue(line, "DESCRIPTION:")
		} else if hasPrefix(line, "ORIGINAL:") {
			lastField = "ORIGINAL"
			rest := parseStringValue(line, "ORIGINAL:")
			if rest != "" && rest != "N/A" {
				// Single-line format
				current.OriginalCode = rest
			} else if rest == "N/A" {
				current.OriginalCode = ""
			}
			// If empty, expect <<< on next line
		} else if hasPrefix(line, "FIX:") {
			lastField = "FIX"
			rest := parseStringValue(line, "FIX:")
			if rest != "" && rest != "N/A" {
				// Single-line format
				current.SuggestFix = rest
			} else if rest == "N/A" {
				current.SuggestFix = ""
			}
			// If empty, expect <<< on next line
		}
	}

	// Don't forget the last one
	if current != nil && current.Title != "" {
		suggestions = append(suggestions, *current)
	}

	return suggestions
}

func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}

func splitLines(s string) []string {
	return strings.Split(s, "\n")
}

func hasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

func parseStringValue(line, prefix string) string {
	if len(line) <= len(prefix) {
		return ""
	}
	return strings.TrimSpace(line[len(prefix):])
}

func parseIntValue(line, prefix string) int {
	s := parseStringValue(line, prefix)
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0 // Invalid or missing line number
	}
	return n
}
