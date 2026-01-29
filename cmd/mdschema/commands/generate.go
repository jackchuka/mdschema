package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jackchuka/mdschema/internal/generator"
	"github.com/jackchuka/mdschema/internal/schema"
	"github.com/spf13/cobra"
)

// NewGenerateCmd creates the generate command
func NewGenerateCmd() *cobra.Command {
	var outputFile string

	cmd := &cobra.Command{
		Use:   "generate [schema-file]",
		Short: "Generate markdown template from schema",
		Long:  `Generate a markdown file template that matches the given schema structure.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ConfigFromContext(cmd.Context())
			var schemaFile string
			if len(args) > 0 {
				schemaFile = args[0]
			}
			return runGenerate(cfg, schemaFile, outputFile)
		},
	}

	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")

	return cmd
}

func runGenerate(cfg *Config, schemaFile, outputFile string) error {
	// Load schema
	var s *schema.Schema
	var err error

	if schemaFile != "" {
		s, err = schema.Load(schemaFile)
		if err != nil {
			return fmt.Errorf("loading schema from %s: %w", schemaFile, err)
		}
	} else {
		// Load schema using existing utility
		s, _, err = loadSchema(cfg)
		if err != nil {
			return fmt.Errorf("loading schema: %w", err)
		}
	}

	// Generate markdown content using the generator package
	gen := generator.New()
	content := gen.Generate(s)

	// Output to file or stdout
	if outputFile != "" {
		if dir := filepath.Dir(outputFile); dir != "." {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("creating directory %s: %w", dir, err)
			}
		}
		err := os.WriteFile(outputFile, []byte(content), 0o644)
		if err != nil {
			return fmt.Errorf("writing to %s: %w", outputFile, err)
		}
		fmt.Printf("✓ Generated markdown template at %s\n", outputFile)
	} else {
		fmt.Print(content)
	}

	return nil
}
