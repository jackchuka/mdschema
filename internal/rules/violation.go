package rules

import "fmt"

// Violation represents a rule violation
type Violation struct {
	Rule     string
	Message  string
	Path     string
	Line     int
	Column   int
	Severity Severity
}

// Severity levels for violations
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// Error returns the violation as an error string
func (v Violation) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s: %s", v.Path, v.Line, v.Column, v.Rule, v.Message)
}

// Position returns the formatted position string
func (v Violation) Position() string {
	return fmt.Sprintf("%s:%d:%d", v.Path, v.Line, v.Column)
}
