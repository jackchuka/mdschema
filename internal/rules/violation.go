package rules

import "fmt"

// Severity levels for violations
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// Violation represents a rule violation
type Violation struct {
	Rule     string
	Message  string
	Path     string
	Line     int
	Column   int
	Severity Severity
}

// NewViolation creates a violation with default severity (error)
func NewViolation(rule, message string, line, column int) Violation {
	return Violation{
		Rule:     rule,
		Message:  message,
		Line:     line,
		Column:   column,
		Severity: SeverityError,
	}
}

// WithSeverity returns a copy of the violation with the specified severity
func (v Violation) WithSeverity(s Severity) Violation {
	v.Severity = s
	return v
}

func (v Violation) WithPath(path string) Violation {
	v.Path = path
	return v
}

// Error returns the violation as an error string
func (v Violation) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s: %s", v.Path, v.Line, v.Column, v.Rule, v.Message)
}

// Position returns the formatted position string
func (v Violation) Position() string {
	return fmt.Sprintf("%s:%d:%d", v.Path, v.Line, v.Column)
}
