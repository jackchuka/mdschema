package commands

import (
	"fmt"
	"os"

	"github.com/jackchuka/mdschema/internal/schema"
	"github.com/spf13/cobra"
)

// NewInitCmd creates the init command
func NewInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a schema in your project",
		Long:  `Creates a .mdschema.yml file with a basic schema configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit()
		},
	}
}

func runInit() error {
	schemaPath := ".mdschema.yml"

	// Check if schema already exists
	if _, err := os.Stat(schemaPath); err == nil {
		fmt.Printf("Schema file already exists at %s\n", schemaPath)
		return nil
	}

	// Create default schema
	if err := schema.CreateDefaultFile(schemaPath); err != nil {
		return fmt.Errorf("creating schema file: %w", err)
	}

	fmt.Printf("✓ Created %s with default configuration\n", schemaPath)
	return nil
}
