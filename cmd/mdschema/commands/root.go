package commands

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// configKey is the context key for Config
type configKey struct{}

// Config holds CLI configuration options
type Config struct {
	SchemaFiles  []string
	OutputFormat string
}

// ConfigFromContext retrieves Config from the command context
func ConfigFromContext(ctx context.Context) *Config {
	if cfg, ok := ctx.Value(configKey{}).(*Config); ok {
		return cfg
	}
	return &Config{OutputFormat: "text"}
}

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
	cfg := &Config{}

	cmd := &cobra.Command{
		Use:   "mdschema",
		Short: "A declarative schema-based Markdown documentation validator",
		Long: `mdschema validates Markdown documentation structure and conventions
using declarative schemas to maintain consistent documentation across projects.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			ctx := context.WithValue(cmd.Context(), configKey{}, cfg)
			cmd.SetContext(ctx)
		},
	}

	// Global flags bound to config
	cmd.PersistentFlags().StringSliceVar(&cfg.SchemaFiles, "schema", []string{}, "Schema file(s) to use")
	cmd.PersistentFlags().StringVar(&cfg.OutputFormat, "format", "text", "Output format: text")

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
		// Don't print ErrViolationsFound - violations already reported
		if !errors.Is(err, ErrViolationsFound) {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}
