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
	switch format {
	case FormatText:
		return NewTextReporter()
	case FormatSARIF, FormatJUnit:
		// TODO: SARIF and JUnit reporters not yet implemented
		return NewTextReporter()
	default:
		return NewTextReporter()
	}
}
