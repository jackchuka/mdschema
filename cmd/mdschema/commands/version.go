package commands

import (
	"fmt"

	"github.com/jackchuka/mdschema/internal/version"
	"github.com/spf13/cobra"
)

// NewVersionCmd creates the version command
func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of mdschema",
		Long:  `Print the version number of mdschema`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(version.Info())
			return nil
		},
	}
}
