package schema

import (
	"gopkg.in/yaml.v3"
)

// Schema represents the validation rules for Markdown files (v0.1 DSL)
type Schema struct {
	// Document structure with embedded section rules
	Structure []StructureElement `yaml:"structure,omitempty"`
}

// StructureElement represents an element in the document structure
// Supports hierarchical structure with children and section-scoped rules
type StructureElement struct {
	// Heading pattern
	Heading string `yaml:"heading,omitempty"`

	// Optional element flag
	Optional bool `yaml:"optional,omitempty"`

	// Hierarchical children elements
	Children []StructureElement `yaml:"children,omitempty"`

	// Embedded section rules for validation within this element's scope
	*SectionRules `yaml:",inline"`
}

// UnmarshalYAML implements custom unmarshaling to support the new hierarchical syntax
func (se *StructureElement) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		se.Heading = node.Value
		return nil
	}

	// Object syntax - use a temporary struct to avoid infinite recursion
	type structureElementAlias StructureElement
	alias := (*structureElementAlias)(se)
	if err := node.Decode(alias); err != nil {
		return err
	}

	return nil
}

// SectionRules defines validation rules scoped to a specific heading/section
type SectionRules struct {
	// Required text/substrings within the section
	RequiredText []string `yaml:"required_text,omitempty"`

	// Code block requirements within this section
	CodeBlocks []CodeBlockRule `yaml:"code_blocks,omitempty"`
}

// CodeBlockRule defines validation for code blocks within a section
type CodeBlockRule struct {
	Lang string `yaml:"lang"`
	Min  int    `yaml:"min,omitempty"`
	Max  int    `yaml:"max,omitempty"`
}
