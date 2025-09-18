package commands

import (
	"fmt"
	"os"

	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/schema/infer"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// NewDeriveCmd creates the derive command
func NewDeriveCmd() *cobra.Command {
	var outputFile string

	cmd := &cobra.Command{
		Use:   "derive <markdown-file>",
		Short: "Infer a schema from an existing Markdown document",
		Long:  "Analyze a Markdown document and generate a schema that reflects its heading hierarchy.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			markdownPath := args[0]
			return runDerive(cmd, markdownPath, outputFile)
		},
	}

	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")

	return cmd
}

func runDerive(cmd *cobra.Command, markdownPath, outputFile string) error {
	p := parser.New()
	doc, err := p.ParseFile(markdownPath)
	if err != nil {
		return fmt.Errorf("parsing markdown: %w", err)
	}

	inferred, err := infer.FromDocument(doc)
	if err != nil {
		return fmt.Errorf("inferring schema: %w", err)
	}

	data, err := yaml.Marshal(inferred)
	if err != nil {
		return fmt.Errorf("encoding schema: %w", err)
	}

	if outputFile != "" {
		if err := os.WriteFile(outputFile, data, 0o644); err != nil {
			return fmt.Errorf("writing schema to %s: %w", outputFile, err)
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "✓ Derived schema written to %s\n", outputFile)
		return nil
	}

	if _, err := cmd.OutOrStdout().Write(data); err != nil {
		return fmt.Errorf("writing schema to stdout: %w", err)
	}

	return nil
}
