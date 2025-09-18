package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	schemaFiles  []string
	outputFormat string
)

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mdschema",
		Short: "A declarative schema-based Markdown documentation validator",
		Long: `mdschema validates Markdown documentation structure and conventions
using declarative schemas to maintain consistent documentation across projects.`,
	}

	// Global flags
	cmd.PersistentFlags().StringSliceVar(&schemaFiles, "schema", []string{}, "Schema file(s) to use")
	cmd.PersistentFlags().StringVar(&outputFormat, "format", "text", "Output format: text")

	// Add subcommands
	cmd.AddCommand(NewInitCmd())
	cmd.AddCommand(NewCheckCmd())
	cmd.AddCommand(NewGenerateCmd())
	cmd.AddCommand(NewDeriveCmd())
	cmd.AddCommand(NewVersionCmd())

	return cmd
}

// Execute runs the root command
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
