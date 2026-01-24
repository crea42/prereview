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
	Confidence   Confidence // How confident the AI is about this suggestion
	Title        string
	Description  string
	OriginalCode string // Original code to be replaced
	SuggestFix   string // Suggested replacement code
	Category     string // security, performance, style, etc.
}

// Confidence levels for suggestions
type Confidence string

const (
	ConfidenceHigh   Confidence = "high"   // Definite issue, should be fixed
	ConfidenceMedium Confidence = "medium" // Likely an issue, review recommended
	ConfidenceLow    Confidence = "low"    // Possible issue, may be false positive
)

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
	projectHints     []string // User-provided hints about the project
	tolerance        string   // strict, moderate, relaxed
}

// NewReviewer creates a new Reviewer instance
func NewReviewer(model string, repoRoot string, customStandards []string, projectHints []string, tolerance string) (*Reviewer, error) {
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

	// Default tolerance
	if tolerance == "" {
		tolerance = "moderate"
	}

	return &Reviewer{
		client:           client,
		model:            model,
		standardsContext: standardsContext,
		projectHints:     projectHints,
		tolerance:        tolerance,
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
	prompt := buildReviewPrompt(change, r.standardsContext, r.projectHints, r.tolerance)

	response, err := r.client.Chat(r.model, prompt)
	if err != nil {
		return nil, err
	}

	return parseReviewResponse(response, change.Path)
}

// buildReviewPrompt creates the prompt for code review
func buildReviewPrompt(change git.FileChange, standardsContext string, projectHints []string, tolerance string) string {
	// Build tolerance-specific guidance
	var toleranceGuidance string
	switch tolerance {
	case "strict":
		toleranceGuidance = `
TOLERANCE: STRICT
- Report all potential issues including style nitpicks
- Mark uncertain issues with CONFIDENCE: low
- Only mark as CONFIDENCE: high when you are 100% certain of an issue`
	case "relaxed":
		toleranceGuidance = `
TOLERANCE: RELAXED
- Only report definite bugs, security vulnerabilities, or critical performance issues
- Skip style suggestions, minor improvements, and best-practice recommendations
- If you're not at least 90% confident about an issue, don't report it
- Skip issues that might be intentional design decisions or framework-specific patterns
- When in doubt, assume the developer knows what they're doing`
	default: // moderate
		toleranceGuidance = `
TOLERANCE: MODERATE
- Report bugs, security issues, and significant code quality concerns
- Skip minor style nitpicks unless they affect readability significantly
- Mark uncertain issues with CONFIDENCE: low or medium
- Consider framework-specific patterns - what looks wrong might be idiomatic`
	}

	basePrompt := `You are a pragmatic code reviewer. Your goal is to be HELPFUL, not pedantic.

IMPORTANT GUIDELINES:
1. AVOID FALSE POSITIVES - When uncertain, don't report. Users hate being blocked by incorrect suggestions.
2. UNDERSTAND CONTEXT - A function that returns multiple types based on input is common (e.g., factory patterns, polymorphism).
3. TRUST THE DEVELOPER - If code works and is reasonable, don't suggest rewrites for minor improvements.
4. FRAMEWORK AWARENESS - Many frameworks have patterns that look wrong but are correct:
   - Factory methods returning different subtypes based on input parameters
   - Type hints may be broader than actual runtime types - this is often intentional
   - Data may be sanitized/escaped at storage time, not output time
   - Different color formats (hex, rgba, hsl) are all valid depending on context
5. DON'T CREATE LOOPS - If suggesting a change would create a new issue, reconsider the suggestion.
` + toleranceGuidance + `

For each GENUINE issue found, respond in this exact format:
---
LINE: <line number where issue starts>
END_LINE: <end line number if multi-line, otherwise same as LINE>
SEVERITY: <error|warning|info|hint>
CONFIDENCE: <high|medium|low>
CATEGORY: <security|performance|style|bug|best-practice>
TITLE: <short title>
DESCRIPTION: <detailed description explaining WHY this is an issue and the RISK if not fixed>
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

CONFIDENCE LEVELS:
- high: You are certain this is a bug, security issue, or definite problem (>95% confident)
- medium: This is likely an issue but could be intentional (~70-95% confident)  
- low: This might be an issue, or might be a valid pattern (<70% confident)

CRITICAL RULES:
1. ORIGINAL must be copied EXACTLY from the file content - character for character
2. Include enough context (2-3 lines before/after) to make the match unique
3. Preserve ALL whitespace, tabs, and indentation exactly as they appear
4. For multi-line code, include all lines between <<< and >>>
5. If no code fix is applicable, use: N/A (without <<< >>>)
6. NEVER suggest a fix that would cause a different issue
7. If the code is already sanitized/escaped upstream, don't flag it again
8. Consider the full context - a "wrong type" might be polymorphic

Focus on:
- Security vulnerabilities (CONFIDENCE: high only for definite issues)
- Actual bugs that will cause runtime errors
- Performance issues with measurable impact
- Error handling gaps that could cause crashes
`

	// Add coding standards context if available
	if standardsContext != "" {
		basePrompt += standardsContext
	}

	// Add project-specific hints if provided
	if len(projectHints) > 0 {
		basePrompt += "\n\nPROJECT-SPECIFIC CONTEXT (trust these hints from the developer):\n"
		for _, hint := range projectHints {
			basePrompt += "- " + hint + "\n"
		}
	}

	basePrompt += `
If no issues are found (or only uncertain low-confidence issues in relaxed mode), respond with: NO_ISSUES

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
		} else if hasPrefix(line, "CONFIDENCE:") {
			current.Confidence = Confidence(parseStringValue(line, "CONFIDENCE:"))
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
