package commands

import (
	"errors"
	"fmt"

	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/reporter"
	"github.com/jackchuka/mdschema/internal/rules"
	"github.com/spf13/cobra"
)

// ErrViolationsFound is returned when validation finds violations
var ErrViolationsFound = errors.New("validation violations found")

// NewCheckCmd creates the check command
func NewCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check [globs...]",
		Short: "Validate Markdown files against schema",
		Long:  `Check validates Markdown files matching the given glob patterns against the configured schema.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ConfigFromContext(cmd.Context())
			return runCheck(cfg, args)
		},
	}
}

func runCheck(cfg *Config, globs []string) error {
	// Load schemas
	schemas, err := loadSchemas(cfg)
	if err != nil {
		return fmt.Errorf("loading schemas: %w", err)
	}

	// Find matching files
	files, err := findFiles(globs)
	if err != nil {
		return fmt.Errorf("finding files: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("No matching files found")
		return nil
	}

	// Parse and validate files
	mdParser := parser.New()
	validator := rules.NewValidator()
	allViolations := make([]rules.Violation, 0)

	for _, file := range files {
		doc, err := mdParser.ParseFile(file)
		if err != nil {
			return fmt.Errorf("parsing %s: %w", file, err)
		}

		// Validate against all schemas
		for _, s := range schemas {
			violations := validator.Validate(doc, s)
			// Set file path for each violation
			for i := range violations {
				violations[i] = violations[i].WithPath(file)
			}
			allViolations = append(allViolations, violations...)
		}
	}

	// Report violations
	rep := reporter.New(reporter.Format(cfg.OutputFormat))
	if err := rep.Report(allViolations); err != nil {
		return fmt.Errorf("reporting violations: %w", err)
	}

	// Return error if violations found (caller handles exit code)
	if len(allViolations) > 0 {
		return ErrViolationsFound
	}

	return nil
}
