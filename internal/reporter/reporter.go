package reporter

import (
	"github.com/jackchuka/mdschema/internal/rules"
)

// Reporter is the interface for outputting violations
type Reporter interface {
	Report(violations []rules.Violation) error
}

// Format represents the output format
type Format string

const (
	FormatText  Format = "text"
	FormatSARIF Format = "sarif"
	FormatJUnit Format = "junit"
)

// New creates a reporter for the specified format
func New(format Format) Reporter {
	// Only text format supported for now
	return NewTextReporter()
}
