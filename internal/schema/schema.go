package schema

import (
	"gopkg.in/yaml.v3"
)

// Schema represents the validation rules for Markdown files (v0.1 DSL)
type Schema struct {
	// Document structure with embedded section rules
	Structure []StructureElement `yaml:"structure,omitempty"`

	// Global link validation rules
	Links *LinkRule `yaml:"links,omitempty"`

	// Global heading validation rules
	HeadingRules *HeadingRules `yaml:"heading_rules,omitempty"`

	// Frontmatter validation rules
	Frontmatter *FrontmatterConfig `yaml:"frontmatter,omitempty"`
}

// LinkRule defines validation rules for links in the document
type LinkRule struct {
	// ValidateInternal validates anchor links (#section-name)
	ValidateInternal bool `yaml:"validate_internal,omitempty"`

	// ValidateFiles validates relative file links (./other.md)
	ValidateFiles bool `yaml:"validate_files,omitempty"`

	// ValidateExternal validates external URLs (http/https)
	ValidateExternal bool `yaml:"validate_external,omitempty"`

	// ExternalTimeout is the timeout in seconds for external URL checks (default: 10)
	ExternalTimeout int `yaml:"external_timeout,omitempty"`

	// AllowedDomains restricts external links to these domains only
	AllowedDomains []string `yaml:"allowed_domains,omitempty"`

	// BlockedDomains blocks external links to these domains
	BlockedDomains []string `yaml:"blocked_domains,omitempty"`
}

// StructureElement represents an element in the document structure
// Supports hierarchical structure with children and section-scoped rules
type StructureElement struct {
	// Heading pattern (string or {pattern: "...", regex: true})
	Heading HeadingPattern `yaml:"heading,omitempty"`

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
		se.Heading = HeadingPattern{Pattern: node.Value}
		return nil
	}

	// Object syntax - use a temporary struct to avoid infinite recursion
	type structureElementAlias StructureElement
	alias := (*structureElementAlias)(se)
	return node.Decode(alias)
}

// HeadingPattern defines a heading pattern with optional regex support
type HeadingPattern struct {
	// Pattern is the heading text or regex pattern to match
	Pattern string `yaml:"pattern,omitempty"`

	// Regex indicates the pattern should be treated as a regular expression
	Regex bool `yaml:"regex,omitempty"`
}

// UnmarshalYAML implements custom unmarshaling to support both string and object syntax
func (h *HeadingPattern) UnmarshalYAML(node *yaml.Node) error {
	// Support simple string syntax: "## Features"
	if node.Kind == yaml.ScalarNode {
		h.Pattern = node.Value
		h.Regex = false
		return nil
	}

	// Object syntax: { pattern: "## .*", regex: true }
	type headingPatternAlias HeadingPattern
	alias := (*headingPatternAlias)(h)
	return node.Decode(alias)
}

// SectionRules defines validation rules scoped to a specific heading/section
type SectionRules struct {
	// Required text/substrings within the section
	RequiredText []RequiredTextPattern `yaml:"required_text,omitempty"`

	// Forbidden text patterns that must NOT appear
	ForbiddenText []ForbiddenTextPattern `yaml:"forbidden_text,omitempty"`

	// Code block requirements within this section
	CodeBlocks []CodeBlockRule `yaml:"code_blocks,omitempty"`

	// Image requirements within this section
	Images []ImageRule `yaml:"images,omitempty"`

	// Table requirements within this section
	Tables []TableRule `yaml:"tables,omitempty"`

	// List requirements within this section
	Lists []ListRule `yaml:"lists,omitempty"`

	// Word count requirements for the section
	WordCount *WordCountRule `yaml:"word_count,omitempty"`
}

// RequiredTextPattern defines a required text pattern with optional regex support
type RequiredTextPattern struct {
	// Pattern is the text or regex pattern to match
	Pattern string `yaml:"pattern,omitempty"`

	// Regex indicates the pattern should be treated as a regular expression
	Regex bool `yaml:"regex,omitempty"`
}

// UnmarshalYAML implements custom unmarshaling to support both string and object syntax
func (r *RequiredTextPattern) UnmarshalYAML(node *yaml.Node) error {
	// Support simple string syntax: "some text"
	if node.Kind == yaml.ScalarNode {
		r.Pattern = node.Value
		r.Regex = false
		return nil
	}

	// Object syntax: { pattern: "...", regex: true }
	type requiredTextPatternAlias RequiredTextPattern
	alias := (*requiredTextPatternAlias)(r)
	return node.Decode(alias)
}

// CodeBlockRule defines validation for code blocks within a section
type CodeBlockRule struct {
	Lang string `yaml:"lang"`
	Min  int    `yaml:"min,omitempty"`
	Max  int    `yaml:"max,omitempty"`
}

// ForbiddenTextPattern defines a text pattern that must NOT appear
type ForbiddenTextPattern struct {
	// Pattern is the text or regex pattern to match
	Pattern string `yaml:"pattern,omitempty"`

	// Regex indicates the pattern should be treated as a regular expression
	Regex bool `yaml:"regex,omitempty"`
}

// UnmarshalYAML implements custom unmarshaling to support both string and object syntax
func (f *ForbiddenTextPattern) UnmarshalYAML(node *yaml.Node) error {
	// Support simple string syntax: "TODO"
	if node.Kind == yaml.ScalarNode {
		f.Pattern = node.Value
		f.Regex = false
		return nil
	}

	// Object syntax: { pattern: "...", regex: true }
	type forbiddenTextPatternAlias ForbiddenTextPattern
	alias := (*forbiddenTextPatternAlias)(f)
	return node.Decode(alias)
}

// ImageRule defines validation for images within a section
type ImageRule struct {
	Min        int      `yaml:"min,omitempty"`
	Max        int      `yaml:"max,omitempty"`
	RequireAlt bool     `yaml:"require_alt,omitempty"`
	Formats    []string `yaml:"formats,omitempty"`
}

// TableRule defines validation for tables within a section
type TableRule struct {
	Min             int      `yaml:"min,omitempty"`
	Max             int      `yaml:"max,omitempty"`
	MinColumns      int      `yaml:"min_columns,omitempty"`
	RequiredHeaders []string `yaml:"required_headers,omitempty"`
}

// ListRule defines validation for lists within a section
type ListRule struct {
	Min      int    `yaml:"min,omitempty"`
	Max      int    `yaml:"max,omitempty"`
	Type     string `yaml:"type,omitempty"` // "ordered", "unordered", or empty for any
	MinItems int    `yaml:"min_items,omitempty"`
}

// WordCountRule defines word count constraints for a section
type WordCountRule struct {
	Min int `yaml:"min,omitempty"`
	Max int `yaml:"max,omitempty"`
}

// HeadingRules defines global validation rules for document headings
type HeadingRules struct {
	// NoSkipLevels ensures heading levels are not skipped (e.g., h1 -> h3 without h2)
	NoSkipLevels bool `yaml:"no_skip_levels,omitempty"`

	// Unique ensures all headings in the document are unique
	Unique bool `yaml:"unique,omitempty"`

	// UniquePerLevel ensures headings are unique within the same level
	UniquePerLevel bool `yaml:"unique_per_level,omitempty"`

	// MaxDepth limits the maximum heading depth (1-6, where 1 is h1)
	MaxDepth int `yaml:"max_depth,omitempty"`
}

// FrontmatterConfig defines validation rules for YAML frontmatter
type FrontmatterConfig struct {
	// Required indicates frontmatter must be present
	Required bool `yaml:"required,omitempty"`

	// Fields defines the required/optional fields and their constraints
	Fields []FrontmatterField `yaml:"fields,omitempty"`
}

// FrontmatterField defines a single frontmatter field requirement
type FrontmatterField struct {
	// Name is the field name (required)
	Name string `yaml:"name"`

	// Required indicates whether this field must be present
	Required bool `yaml:"required,omitempty"`

	// Type is the expected type: "string", "number", "boolean", "array", "date"
	Type string `yaml:"type,omitempty"`

	// Format specifies format validation (e.g., "date" for YYYY-MM-DD)
	Format string `yaml:"format,omitempty"`
}
