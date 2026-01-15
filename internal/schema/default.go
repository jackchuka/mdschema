package schema

import (
	"os"

	yamlcomment "github.com/zijiren233/yaml-comment"
)

const schemaPreamble = `# Markdown Schema Configuration
# This file demonstrates all available schema capabilities.
# See: https://github.com/jackchuka/mdschema for documentation

`

// buildDefaultSchema creates a Schema struct demonstrating all capabilities
func buildDefaultSchema() *Schema {
	return &Schema{
		// Global frontmatter validation
		Frontmatter: &FrontmatterConfig{
			Required: true,
			Fields: []FrontmatterField{
				{Name: "title", Required: true, Type: FieldTypeString},
				{Name: "description", Required: false, Type: FieldTypeString},
				{Name: "author", Required: false, Type: FieldTypeString, Format: FieldFormatEmail},
				{Name: "date", Required: true, Type: FieldTypeDate, Format: FieldFormatDate},
				{Name: "tags", Required: false, Type: FieldTypeArray},
				{Name: "draft", Required: false, Type: FieldTypeBoolean},
				{Name: "version", Required: false, Type: FieldTypeNumber},
				{Name: "repository", Required: false, Type: FieldTypeString, Format: FieldFormatURL},
			},
		},

		// Global heading validation rules
		HeadingRules: &HeadingRules{
			NoSkipLevels:   true,
			Unique:         true,
			UniquePerLevel: false,
			MaxDepth:       4,
		},

		// Global link validation rules
		Links: &LinkRule{
			ValidateInternal: true,
			ValidateFiles:    true,
			ValidateExternal: false,
			ExternalTimeout:  10,
			AllowedDomains:   []string{"github.com", "golang.org", "pkg.go.dev"},
			BlockedDomains:   []string{"example.com"},
		},

		// Document structure with section-scoped rules
		Structure: []StructureElement{
			{
				// Root heading with regex pattern
				Heading: HeadingPattern{
					Pattern: "# [A-Za-z0-9][A-Za-z0-9 _-]*",
					Regex:   true,
				},
				Children: []StructureElement{
					// First section demonstrates ALL rule types so comments appear
					{
						Heading:  HeadingPattern{Pattern: "## Overview"},
						Optional: false,
						SectionRules: &SectionRules{
							RequiredText: []RequiredTextPattern{
								{Pattern: "purpose", Regex: false},
							},
							ForbiddenText: []ForbiddenTextPattern{
								{Pattern: "TODO", Regex: false},
							},
							CodeBlocks: []CodeBlockRule{
								{Lang: "bash", Min: 0, Max: 2},
							},
							Images: []ImageRule{
								{Min: 0, Max: 3, RequireAlt: true, Formats: []string{"png", "jpg", "gif"}},
							},
							Tables: []TableRule{
								{Min: 0, Max: 1, MinColumns: 2, RequiredHeaders: []string{"Column", "Description"}},
							},
							Lists: []ListRule{
								{Min: 0, Max: 5, Type: ListTypeUnordered, MinItems: 2},
							},
							WordCount: &WordCountRule{Min: 50, Max: 500},
						},
					},
					{
						Heading:  HeadingPattern{Pattern: "## Features"},
						Optional: true,
						SectionRules: &SectionRules{
							Lists: []ListRule{
								{Min: 1, Type: ListTypeUnordered, MinItems: 3},
							},
						},
					},
					{
						Heading:  HeadingPattern{Pattern: "## Installation"},
						Optional: false,
						SectionRules: &SectionRules{
							CodeBlocks: []CodeBlockRule{
								{Lang: "bash", Min: 1},
							},
						},
						Children: []StructureElement{
							{
								Heading:  HeadingPattern{Pattern: "### Prerequisites"},
								Optional: true,
							},
							{
								Heading:  HeadingPattern{Pattern: "### Quick Start"},
								Optional: true,
							},
						},
					},
					{
						Heading:  HeadingPattern{Pattern: "## Usage"},
						Optional: false,
						SectionRules: &SectionRules{
							CodeBlocks: []CodeBlockRule{
								{Lang: "go", Min: 1},
							},
						},
					},
					{
						Heading: HeadingPattern{
							Pattern: "## (Contributing|How to Contribute)",
							Regex:   true,
						},
						Optional: true,
					},
				},
			},
			{
				Heading:  HeadingPattern{Pattern: "# License"},
				Optional: true,
			},
		},
	}
}

// CreateDefaultFile creates a .mdschema.yml file with default configuration
// demonstrating all available schema capabilities.
func CreateDefaultFile(path string) error {
	schema := buildDefaultSchema()

	yamlBytes, err := yamlcomment.Marshal(schema)
	if err != nil {
		return err
	}

	content := schemaPreamble + string(yamlBytes)
	return os.WriteFile(path, []byte(content), 0o644)
}
