package reporter

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/jackchuka/mdschema/internal/rules"
)

// TextReporter outputs violations in human-readable text format
type TextReporter struct {
	writer io.Writer
	colors bool
}

// NewTextReporter creates a new text reporter
func NewTextReporter() *TextReporter {
	return &TextReporter{
		writer: os.Stdout,
		colors: isTerminal(),
	}
}

// Report outputs violations in text format
func (r *TextReporter) Report(violations []rules.Violation) error {
	if len(violations) == 0 {
		_, _ = fmt.Fprintln(r.writer, r.formatSuccess("✓ No violations found"))
		return nil
	}

	// Sort violations by file and line
	sort.Slice(violations, func(i, j int) bool {
		if violations[i].Path != violations[j].Path {
			return violations[i].Path < violations[j].Path
		}
		return violations[i].Line < violations[j].Line
	})

	// Group violations by file
	fileViolations := make(map[string][]rules.Violation)
	for _, v := range violations {
		fileViolations[v.Path] = append(fileViolations[v.Path], v)
	}

	// Output violations
	for path, vList := range fileViolations {
		_, _ = fmt.Fprintln(r.writer, r.formatFile(path))
		for _, v := range vList {
			_, _ = fmt.Fprintln(r.writer, r.formatViolation(v))
		}
		_, _ = fmt.Fprintln(r.writer)
	}

	// Summary
	_, _ = fmt.Fprintf(r.writer, r.formatError("✗ Found %d violation(s) in %d file(s)\n"),
		len(violations), len(fileViolations))

	return nil
}

func (r *TextReporter) formatViolation(v rules.Violation) string {
	var icon string
	var colorFunc func(string) string

	switch v.Severity {
	case rules.SeverityError:
		icon = "✗"
		colorFunc = r.formatError
	case rules.SeverityWarning:
		icon = "⚠"
		colorFunc = r.formatWarning
	case rules.SeverityInfo:
		icon = "ℹ"
		colorFunc = r.formatInfo
	}

	position := fmt.Sprintf("%d:%d", v.Line, v.Column)
	return fmt.Sprintf("  %s %s %s %s",
		colorFunc(icon),
		r.formatDim(position),
		r.formatRule(v.Rule),
		v.Message)
}

func (r *TextReporter) formatFile(path string) string {
	return r.formatBold(path)
}

func (r *TextReporter) formatRule(rule string) string {
	return r.formatDim(fmt.Sprintf("[%s]", rule))
}

// Color formatting functions
func (r *TextReporter) formatError(s string) string {
	if r.colors {
		return fmt.Sprintf("\033[31m%s\033[0m", s) // Red
	}
	return s
}

func (r *TextReporter) formatWarning(s string) string {
	if r.colors {
		return fmt.Sprintf("\033[33m%s\033[0m", s) // Yellow
	}
	return s
}

func (r *TextReporter) formatInfo(s string) string {
	if r.colors {
		return fmt.Sprintf("\033[36m%s\033[0m", s) // Cyan
	}
	return s
}

func (r *TextReporter) formatSuccess(s string) string {
	if r.colors {
		return fmt.Sprintf("\033[32m%s\033[0m", s) // Green
	}
	return s
}

func (r *TextReporter) formatBold(s string) string {
	if r.colors {
		return fmt.Sprintf("\033[1m%s\033[0m", s) // Bold
	}
	return s
}

func (r *TextReporter) formatDim(s string) string {
	if r.colors {
		return fmt.Sprintf("\033[2m%s\033[0m", s) // Dim
	}
	return s
}

// isTerminal checks if stdout is a terminal
func isTerminal() bool {
	// Simple check - could be enhanced with platform-specific code
	return os.Getenv("TERM") != "" && !strings.Contains(os.Getenv("TERM"), "dumb")
}
