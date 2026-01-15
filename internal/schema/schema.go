package schema

import (
	"gopkg.in/yaml.v3"
)

// Schema represents the validation rules for Markdown files (v0.1 DSL)
type Schema struct {
	// Document structure with embedded section rules
	Structure []StructureElement `yaml:"structure,omitempty" hc:"Document structure defines required/optional sections and their validation rules"`

	// Global link validation rules
	Links *LinkRule `yaml:"links,omitempty" hc:"Global link validation settings"`

	// Global heading validation rules
	HeadingRules *HeadingRules `yaml:"heading_rules,omitempty" hc:"Global heading validation rules"`

	// Frontmatter validation rules
	Frontmatter *FrontmatterConfig `yaml:"frontmatter,omitempty" hc:"YAML frontmatter validation"`
}

// LinkRule defines validation rules for links in the document
type LinkRule struct {
	// ValidateInternal validates anchor links (#section-name)
	ValidateInternal bool `yaml:"validate_internal,omitempty" lc:"check anchor links (#section-name)"`

	// ValidateFiles validates relative file links (./other.md)
	ValidateFiles bool `yaml:"validate_files,omitempty" lc:"check relative file links (./other.md)"`

	// ValidateExternal validates external URLs (http/https)
	ValidateExternal bool `yaml:"validate_external,omitempty" lc:"check external URLs (http/https)"`

	// ExternalTimeout is the timeout in seconds for external URL checks (default: 10)
	ExternalTimeout int `yaml:"external_timeout,omitempty" lc:"timeout in seconds for external URL checks"`

	// AllowedDomains restricts external links to these domains only
	AllowedDomains []string `yaml:"allowed_domains,omitempty" lc:"restrict external links to these domains"`

	// BlockedDomains blocks external links to these domains
	BlockedDomains []string `yaml:"blocked_domains,omitempty" lc:"block links to these domains"`
}

// StructureElement represents an element in the document structure
// Supports hierarchical structure with children and section-scoped rules
type StructureElement struct {
	// Heading pattern (string or {pattern: "...", regex: true})
	Heading HeadingPattern `yaml:"heading,omitempty"`

	// Optional element flag
	Optional bool `yaml:"optional,omitempty" lc:"section is not required"`

	// Hierarchical children elements
	Children []StructureElement `yaml:"children,omitempty" lc:"nested subsections"`

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
	Pattern string `yaml:"pattern,omitempty" lc:"heading text or regex pattern"`

	// Regex indicates the pattern should be treated as a regular expression
	Regex bool `yaml:"regex,omitempty" lc:"treat pattern as regular expression"`
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
	RequiredText []RequiredTextPattern `yaml:"required_text,omitempty" lc:"text that must appear in this section"`

	// Forbidden text patterns that must NOT appear
	ForbiddenText []ForbiddenTextPattern `yaml:"forbidden_text,omitempty" lc:"text that must NOT appear"`

	// Code block requirements within this section
	CodeBlocks []CodeBlockRule `yaml:"code_blocks,omitempty" lc:"code block requirements"`

	// Image requirements within this section
	Images []ImageRule `yaml:"images,omitempty" lc:"image requirements"`

	// Table requirements within this section
	Tables []TableRule `yaml:"tables,omitempty" lc:"table requirements"`

	// List requirements within this section
	Lists []ListRule `yaml:"lists,omitempty" lc:"list requirements"`

	// Word count requirements for the section
	WordCount *WordCountRule `yaml:"word_count,omitempty" lc:"word count constraints"`
}

// RequiredTextPattern defines a required text pattern with optional regex support
type RequiredTextPattern struct {
	// Pattern is the text or regex pattern to match
	Pattern string `yaml:"pattern,omitempty" lc:"text or regex to match"`

	// Regex indicates the pattern should be treated as a regular expression
	Regex bool `yaml:"regex,omitempty" lc:"treat as regex"`
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
	Lang string `yaml:"lang" lc:"language identifier (bash, go, python, etc.)"`
	Min  int    `yaml:"min,omitempty" lc:"minimum required blocks"`
	Max  int    `yaml:"max,omitempty" lc:"maximum allowed blocks"`
}

// ForbiddenTextPattern defines a text pattern that must NOT appear
type ForbiddenTextPattern struct {
	// Pattern is the text or regex pattern to match
	Pattern string `yaml:"pattern,omitempty" lc:"text or regex that must NOT appear"`

	// Regex indicates the pattern should be treated as a regular expression
	Regex bool `yaml:"regex,omitempty" lc:"treat as regex"`
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
	Min        int      `yaml:"min,omitempty" lc:"minimum required images"`
	Max        int      `yaml:"max,omitempty" lc:"maximum allowed images"`
	RequireAlt bool     `yaml:"require_alt,omitempty" lc:"require alt text"`
	Formats    []string `yaml:"formats,omitempty" lc:"allowed formats (png, jpg, gif, etc.)"`
}

// TableRule defines validation for tables within a section
type TableRule struct {
	Min             int      `yaml:"min,omitempty" lc:"minimum required tables"`
	Max             int      `yaml:"max,omitempty" lc:"maximum allowed tables"`
	MinColumns      int      `yaml:"min_columns,omitempty" lc:"minimum columns per table"`
	RequiredHeaders []string `yaml:"required_headers,omitempty" lc:"headers that must exist"`
}

// ListType represents the type of a list
type ListType string

// List type constants
const (
	ListTypeOrdered   ListType = "ordered"
	ListTypeUnordered ListType = "unordered"
)

// ListRule defines validation for lists within a section
type ListRule struct {
	Min      int      `yaml:"min,omitempty" lc:"minimum required lists"`
	Max      int      `yaml:"max,omitempty" lc:"maximum allowed lists"`
	Type     ListType `yaml:"type,omitempty" lc:"ordered, unordered, or empty for any"`
	MinItems int      `yaml:"min_items,omitempty" lc:"minimum items per list"`
}

// WordCountRule defines word count constraints for a section
type WordCountRule struct {
	Min int `yaml:"min,omitempty" lc:"minimum words"`
	Max int `yaml:"max,omitempty" lc:"maximum words"`
}

// HeadingRules defines global validation rules for document headings
type HeadingRules struct {
	// NoSkipLevels ensures heading levels are not skipped (e.g., h1 -> h3 without h2)
	NoSkipLevels bool `yaml:"no_skip_levels,omitempty" lc:"disallow skipping levels (e.g., h1 -> h3)"`

	// Unique ensures all headings in the document are unique
	Unique bool `yaml:"unique,omitempty" lc:"all headings must be unique"`

	// UniquePerLevel ensures headings are unique within the same level
	UniquePerLevel bool `yaml:"unique_per_level,omitempty" lc:"headings unique within same level"`

	// MaxDepth limits the maximum heading depth (1-6, where 1 is h1)
	MaxDepth int `yaml:"max_depth,omitempty" lc:"maximum heading depth (1-6)"`
}

// FrontmatterConfig defines validation rules for YAML frontmatter
type FrontmatterConfig struct {
	// Optional indicates frontmatter block is not required (default: false = required)
	Optional bool `yaml:"optional,omitempty" lc:"frontmatter block is not required"`

	// Fields defines the required/optional fields and their constraints
	Fields []FrontmatterField `yaml:"fields,omitempty" lc:"field definitions"`
}

// FieldType represents the type of a frontmatter field
type FieldType string

// Field type constants for frontmatter validation
const (
	FieldTypeString  FieldType = "string"
	FieldTypeNumber  FieldType = "number"
	FieldTypeBoolean FieldType = "boolean"
	FieldTypeArray   FieldType = "array"
	FieldTypeDate    FieldType = "date"
)

// FieldFormat represents the format of a frontmatter field
type FieldFormat string

// Field format constants for frontmatter validation
const (
	FieldFormatDate  FieldFormat = "date"  // YYYY-MM-DD
	FieldFormatEmail FieldFormat = "email" // valid email address
	FieldFormatURL   FieldFormat = "url"   // http:// or https://
)

// FrontmatterField defines a single frontmatter field requirement
type FrontmatterField struct {
	// Name is the field name (required)
	Name string `yaml:"name" lc:"field name"`

	// Optional indicates whether this field is not required (default: false = required)
	Optional bool `yaml:"optional,omitempty" lc:"field is not required"`

	// Type is the expected type (use FieldType* constants)
	Type FieldType `yaml:"type,omitempty" lc:"string, number, boolean, array, date"`

	// Format specifies format validation (use FieldFormat* constants)
	Format FieldFormat `yaml:"format,omitempty" lc:"date, email, or url"`
}
