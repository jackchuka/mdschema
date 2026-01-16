package reporter

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/fatih/color"
	"github.com/jackchuka/mdschema/internal/rules"
)

// TextReporter outputs violations in human-readable text format
type TextReporter struct {
	writer io.Writer
}

// NewTextReporter creates a new text reporter
func NewTextReporter() *TextReporter {
	return &TextReporter{
		writer: os.Stdout,
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
	return color.RedString(s)
}

func (r *TextReporter) formatWarning(s string) string {
	return color.YellowString(s)
}

func (r *TextReporter) formatInfo(s string) string {
	return color.CyanString(s)
}

func (r *TextReporter) formatSuccess(s string) string {
	return color.GreenString(s)
}

func (r *TextReporter) formatBold(s string) string {
	return color.New(color.Bold).Sprint(s)
}

func (r *TextReporter) formatDim(s string) string {
	return color.New(color.Faint).Sprint(s)
}
