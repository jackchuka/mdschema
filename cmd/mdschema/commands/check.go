package commands

import (
	"fmt"
	"os"

	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/reporter"
	"github.com/jackchuka/mdschema/internal/rules"
	"github.com/spf13/cobra"
)

// NewCheckCmd creates the check command
func NewCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check [globs...]",
		Short: "Validate Markdown files against schema",
		Long:  `Check validates Markdown files matching the given glob patterns against the configured schema.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheck(args)
		},
	}
}

func runCheck(globs []string) error {
	// Load schemas
	schemas, err := loadSchemas()
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
			// Set file path and default severity for each violation
			for i := range violations {
				violations[i].Path = file
				if violations[i].Severity == "" {
					violations[i].Severity = rules.SeverityError
				}
			}
			allViolations = append(allViolations, violations...)
		}
	}

	// Report violations
	rep := reporter.New(reporter.Format(outputFormat))
	if err := rep.Report(allViolations); err != nil {
		return fmt.Errorf("reporting violations: %w", err)
	}

	// Exit with non-zero status if violations found
	if len(allViolations) > 0 {
		os.Exit(1)
	}

	return nil
}
