package schema

import (
	"github.com/invopop/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"
)

// Schema represents the validation rules for Markdown files (v0.1 DSL)
type Schema struct {
	// Document structure with embedded section rules
	Structure []StructureElement `yaml:"structure,omitempty" json:"structure,omitempty" hc:"Document structure defines required/optional sections and their validation rules"`

	// Global link validation rules
	Links *LinkRule `yaml:"links,omitempty" json:"links,omitempty" hc:"Global link validation settings"`

	// Global heading validation rules
	HeadingRules *HeadingRules `yaml:"heading_rules,omitempty" json:"heading_rules,omitempty" hc:"Global heading validation rules"`

	// Frontmatter validation rules
	Frontmatter *FrontmatterConfig `yaml:"frontmatter,omitempty" json:"frontmatter,omitempty" hc:"YAML frontmatter validation"`
}

// LinkRule defines validation rules for links in the document
type LinkRule struct {
	// ValidateInternal validates anchor links (#section-name)
	ValidateInternal bool `yaml:"validate_internal,omitempty" json:"validate_internal,omitempty" lc:"check anchor links (#section-name)"`

	// ValidateFiles validates relative file links (./other.md)
	ValidateFiles bool `yaml:"validate_files,omitempty" json:"validate_files,omitempty" lc:"check relative file links (./other.md)"`

	// ValidateExternal validates external URLs (http/https)
	ValidateExternal bool `yaml:"validate_external,omitempty" json:"validate_external,omitempty" lc:"check external URLs (http/https)"`

	// ExternalTimeout is the timeout in seconds for external URL checks (default: 10)
	ExternalTimeout int `yaml:"external_timeout,omitempty" json:"external_timeout,omitempty" lc:"timeout in seconds for external URL checks"`

	// AllowedDomains restricts external links to these domains only
	AllowedDomains []string `yaml:"allowed_domains,omitempty" json:"allowed_domains,omitempty" lc:"restrict external links to these domains"`

	// BlockedDomains blocks external links to these domains
	BlockedDomains []string `yaml:"blocked_domains,omitempty" json:"blocked_domains,omitempty" lc:"block links to these domains"`
}

// StructureElement represents an element in the document structure
// Supports hierarchical structure with children and section-scoped rules
type StructureElement struct {
	// Heading pattern (string or {pattern: "...", regex: true})
	Heading HeadingPattern `yaml:"heading,omitempty" json:"heading,omitempty"`

	// Description is guidance text shown in generated output as an HTML comment
	Description string `yaml:"description,omitempty" json:"description,omitempty" lc:"section description shown in generated output"`

	// Optional element flag
	Optional bool `yaml:"optional,omitempty" json:"optional,omitempty" lc:"section is not required"`

	// Count defines how many times this element can match (default: exactly once)
	// When specified, takes precedence over Optional
	Count *CountConstraint `yaml:"count,omitempty" json:"count,omitempty" lc:"occurrence constraints {min, max}"`

	// Severity level for violations (error, warning, info). Default: error
	Severity string `yaml:"severity,omitempty" json:"severity,omitempty" lc:"violation severity: error, warning, or info" jsonschema:"enum=error,enum=warning,enum=info"`

	// AllowAdditional permits extra subsections not defined in children
	AllowAdditional bool `yaml:"allow_additional,omitempty" json:"allow_additional,omitempty" lc:"allow extra subsections not in schema"`

	// Hierarchical children elements
	Children []StructureElement `yaml:"children,omitempty" json:"children,omitempty" lc:"nested subsections"`

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

// JSONSchema implements jsonschema.JSONSchemer for union type support (string | object)
func (StructureElement) JSONSchema() *jsonschema.Schema {
	props := jsonschema.NewProperties()
	props.Set("heading", &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			{Type: "string", Description: "Simple heading text (e.g., '## Features')"},
			{
				Type:        "object",
				Description: "Heading pattern with optional regex support",
				Properties: func() *orderedmap.OrderedMap[string, *jsonschema.Schema] {
					p := jsonschema.NewProperties()
					p.Set("pattern", &jsonschema.Schema{Type: "string", Description: "Heading text or regex pattern"})
					p.Set("regex", &jsonschema.Schema{Type: "boolean", Description: "Treat pattern as regular expression"})
					return p
				}(),
			},
		},
		Description: "Heading pattern to match",
	})
	props.Set("description", &jsonschema.Schema{Type: "string", Description: "Section description shown in generated output"})
	props.Set("optional", &jsonschema.Schema{Type: "boolean", Description: "Section is not required"})
	props.Set("count", &jsonschema.Schema{Ref: "#/$defs/CountConstraint", Description: "Occurrence constraints {min, max}"})
	props.Set("severity", &jsonschema.Schema{Type: "string", Enum: []any{"error", "warning", "info"}, Description: "Violation severity: error, warning, or info"})
	props.Set("allow_additional", &jsonschema.Schema{Type: "boolean", Description: "Allow extra subsections not in schema"})
	props.Set("children", &jsonschema.Schema{
		Type:        "array",
		Description: "Nested subsections",
		Items:       &jsonschema.Schema{Ref: "#/$defs/StructureElement"},
	})
	// SectionRules inline fields
	props.Set("required_text", &jsonschema.Schema{
		Type:        "array",
		Description: "Text that must appear in this section",
		Items: &jsonschema.Schema{
			OneOf: []*jsonschema.Schema{
				{Type: "string", Description: "Simple text to match"},
				{
					Type:        "object",
					Description: "Pattern with optional regex support",
					Properties: func() *orderedmap.OrderedMap[string, *jsonschema.Schema] {
						p := jsonschema.NewProperties()
						p.Set("pattern", &jsonschema.Schema{Type: "string", Description: "Text or regex to match"})
						p.Set("regex", &jsonschema.Schema{Type: "boolean", Description: "Treat as regex"})
						return p
					}(),
				},
			},
		},
	})
	props.Set("forbidden_text", &jsonschema.Schema{
		Type:        "array",
		Description: "Text that must NOT appear",
		Items: &jsonschema.Schema{
			OneOf: []*jsonschema.Schema{
				{Type: "string", Description: "Simple text that must not appear"},
				{
					Type:        "object",
					Description: "Pattern with optional regex support",
					Properties: func() *orderedmap.OrderedMap[string, *jsonschema.Schema] {
						p := jsonschema.NewProperties()
						p.Set("pattern", &jsonschema.Schema{Type: "string", Description: "Text or regex that must NOT appear"})
						p.Set("regex", &jsonschema.Schema{Type: "boolean", Description: "Treat as regex"})
						return p
					}(),
				},
			},
		},
	})
	props.Set("code_blocks", &jsonschema.Schema{
		Type:        "array",
		Description: "Code block requirements",
		Items:       &jsonschema.Schema{Ref: "#/$defs/CodeBlockRule"},
	})
	props.Set("images", &jsonschema.Schema{
		Type:        "array",
		Description: "Image requirements",
		Items:       &jsonschema.Schema{Ref: "#/$defs/ImageRule"},
	})
	props.Set("tables", &jsonschema.Schema{
		Type:        "array",
		Description: "Table requirements",
		Items:       &jsonschema.Schema{Ref: "#/$defs/TableRule"},
	})
	props.Set("lists", &jsonschema.Schema{
		Type:        "array",
		Description: "List requirements",
		Items:       &jsonschema.Schema{Ref: "#/$defs/ListRule"},
	})
	props.Set("word_count", &jsonschema.Schema{Ref: "#/$defs/WordCountRule", Description: "Word count constraints"})

	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			{Type: "string", Description: "Simple heading text (shorthand for {heading: '...'})"},
			{
				Type:                 "object",
				Properties:           props,
				AdditionalProperties: jsonschema.FalseSchema,
				Description:          "Structure element with heading pattern and validation rules",
			},
		},
		Description: "Document structure element",
	}
}

// HeadingPattern defines a heading pattern with optional regex or expression support
type HeadingPattern struct {
	// Literal is set when using scalar form (e.g., heading: "## Features") - exact match
	Literal string `yaml:"-" json:"-"`

	// Pattern is the heading regex pattern to match (always treated as regex)
	Pattern string `yaml:"pattern,omitempty" json:"pattern,omitempty" lc:"heading regex pattern"`

	// Expr is a boolean expression for dynamic matching (e.g., "slug(filename) == slug(heading)")
	// Available variables: filename (without extension), heading (heading text)
	// Available functions: slug, lower, upper, trim, hasPrefix, hasSuffix, strContains, match, replace, trimPrefix, trimSuffix
	Expr string `yaml:"expr,omitempty" json:"expr,omitempty" lc:"boolean expression for dynamic matching"`
}

// UnmarshalYAML implements custom unmarshaling to support both string and object syntax
func (h *HeadingPattern) UnmarshalYAML(node *yaml.Node) error {
	// Support simple string syntax: "## Features" (literal match)
	if node.Kind == yaml.ScalarNode {
		h.Literal = node.Value
		return nil
	}

	// Object syntax: { pattern: "## .*" } or { expr: "..." }
	type headingPatternAlias HeadingPattern
	alias := (*headingPatternAlias)(h)
	return node.Decode(alias)
}

// JSONSchema implements jsonschema.JSONSchemer for union type support (string | object)
func (HeadingPattern) JSONSchema() *jsonschema.Schema {
	props := jsonschema.NewProperties()
	props.Set("pattern", &jsonschema.Schema{
		Type:        "string",
		Description: "Heading regex pattern to match",
	})
	props.Set("expr", &jsonschema.Schema{
		Type:        "string",
		Description: "Boolean expression for dynamic matching (e.g., 'slug(filename) == slug(heading)')",
	})

	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			{Type: "string", Description: "Simple heading text for literal match (e.g., '## Features')"},
			{
				Type:                 "object",
				Properties:           props,
				AdditionalProperties: jsonschema.FalseSchema,
				Description:          "Heading pattern with regex or expression support",
			},
		},
		Description: "Heading pattern",
	}
}

// SectionRules defines validation rules scoped to a specific heading/section
type SectionRules struct {
	// Required text/substrings within the section
	RequiredText []RequiredTextPattern `yaml:"required_text,omitempty" json:"required_text,omitempty" lc:"text that must appear in this section"`

	// Forbidden text patterns that must NOT appear
	ForbiddenText []ForbiddenTextPattern `yaml:"forbidden_text,omitempty" json:"forbidden_text,omitempty" lc:"text that must NOT appear"`

	// Code block requirements within this section
	CodeBlocks []CodeBlockRule `yaml:"code_blocks,omitempty" json:"code_blocks,omitempty" lc:"code block requirements"`

	// Image requirements within this section
	Images []ImageRule `yaml:"images,omitempty" json:"images,omitempty" lc:"image requirements"`

	// Table requirements within this section
	Tables []TableRule `yaml:"tables,omitempty" json:"tables,omitempty" lc:"table requirements"`

	// List requirements within this section
	Lists []ListRule `yaml:"lists,omitempty" json:"lists,omitempty" lc:"list requirements"`

	// Word count requirements for the section
	WordCount *WordCountRule `yaml:"word_count,omitempty" json:"word_count,omitempty" lc:"word count constraints"`
}

// RequiredTextPattern defines a required text pattern with optional regex support
type RequiredTextPattern struct {
	// Literal is set when using scalar form (e.g., required_text: "text") - substring match
	Literal string `yaml:"-" json:"-"`

	// Pattern is the regex pattern to match (always treated as regex)
	Pattern string `yaml:"pattern,omitempty" json:"pattern,omitempty" lc:"regex pattern to match"`
}

// UnmarshalYAML implements custom unmarshaling to support both string and object syntax
func (r *RequiredTextPattern) UnmarshalYAML(node *yaml.Node) error {
	// Support simple string syntax: "some text" (literal substring match)
	if node.Kind == yaml.ScalarNode {
		r.Literal = node.Value
		return nil
	}

	// Object syntax: { pattern: "..." }
	type requiredTextPatternAlias RequiredTextPattern
	alias := (*requiredTextPatternAlias)(r)
	return node.Decode(alias)
}

// JSONSchema implements jsonschema.JSONSchemer for union type support (string | object)
func (RequiredTextPattern) JSONSchema() *jsonschema.Schema {
	props := jsonschema.NewProperties()
	props.Set("pattern", &jsonschema.Schema{
		Type:        "string",
		Description: "Regex pattern to match",
	})

	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			{Type: "string", Description: "Simple text to match (substring)"},
			{
				Type:                 "object",
				Properties:           props,
				AdditionalProperties: jsonschema.FalseSchema,
				Description:          "Regex pattern to match",
			},
		},
		Description: "Required text pattern",
	}
}

// CodeBlockRule defines validation for code blocks within a section
type CodeBlockRule struct {
	Lang string `yaml:"lang,omitempty" json:"lang,omitempty" lc:"language identifier (bash, go, python, etc.) - omit for any language"`
	Min  int    `yaml:"min,omitempty" json:"min,omitempty" lc:"minimum required blocks"`
	Max  int    `yaml:"max,omitempty" json:"max,omitempty" lc:"maximum allowed blocks"`
}

// ForbiddenTextPattern defines a text pattern that must NOT appear
type ForbiddenTextPattern struct {
	// Literal is set when using scalar form (e.g., forbidden_text: "TODO") - substring match
	Literal string `yaml:"-" json:"-"`

	// Pattern is the regex pattern to match (always treated as regex)
	Pattern string `yaml:"pattern,omitempty" json:"pattern,omitempty" lc:"regex pattern that must NOT appear"`
}

// UnmarshalYAML implements custom unmarshaling to support both string and object syntax
func (f *ForbiddenTextPattern) UnmarshalYAML(node *yaml.Node) error {
	// Support simple string syntax: "TODO" (literal substring match)
	if node.Kind == yaml.ScalarNode {
		f.Literal = node.Value
		return nil
	}

	// Object syntax: { pattern: "..." }
	type forbiddenTextPatternAlias ForbiddenTextPattern
	alias := (*forbiddenTextPatternAlias)(f)
	return node.Decode(alias)
}

// JSONSchema implements jsonschema.JSONSchemer for union type support (string | object)
func (ForbiddenTextPattern) JSONSchema() *jsonschema.Schema {
	props := jsonschema.NewProperties()
	props.Set("pattern", &jsonschema.Schema{
		Type:        "string",
		Description: "Regex pattern that must NOT appear",
	})

	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			{Type: "string", Description: "Simple text that must not appear (substring)"},
			{
				Type:                 "object",
				Properties:           props,
				AdditionalProperties: jsonschema.FalseSchema,
				Description:          "Regex pattern that must not appear",
			},
		},
		Description: "Forbidden text pattern",
	}
}

// ImageRule defines validation for images within a section
type ImageRule struct {
	Min        int      `yaml:"min,omitempty" json:"min,omitempty" lc:"minimum required images"`
	Max        int      `yaml:"max,omitempty" json:"max,omitempty" lc:"maximum allowed images"`
	RequireAlt bool     `yaml:"require_alt,omitempty" json:"require_alt,omitempty" lc:"require alt text"`
	Formats    []string `yaml:"formats,omitempty" json:"formats,omitempty" lc:"allowed formats (png, jpg, gif, etc.)"`
}

// TableRule defines validation for tables within a section
type TableRule struct {
	Min             int      `yaml:"min,omitempty" json:"min,omitempty" lc:"minimum required tables"`
	Max             int      `yaml:"max,omitempty" json:"max,omitempty" lc:"maximum allowed tables"`
	MinColumns      int      `yaml:"min_columns,omitempty" json:"min_columns,omitempty" lc:"minimum columns per table"`
	RequiredHeaders []string `yaml:"required_headers,omitempty" json:"required_headers,omitempty" lc:"headers that must exist"`
}

// ListType represents the type of a list
type ListType string

// List type constants
const (
	ListTypeOrdered   ListType = "ordered"
	ListTypeUnordered ListType = "unordered"
)

// JSONSchema implements jsonschema.JSONSchemer to add enum constraint
func (ListType) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "string",
		Enum:        []any{"ordered", "unordered"},
		Description: "List type: ordered or unordered",
	}
}

// ListRule defines validation for lists within a section
type ListRule struct {
	Min      int      `yaml:"min,omitempty" json:"min,omitempty" lc:"minimum required lists"`
	Max      int      `yaml:"max,omitempty" json:"max,omitempty" lc:"maximum allowed lists"`
	Type     ListType `yaml:"type,omitempty" json:"type,omitempty" lc:"ordered, unordered, or empty for any"`
	MinItems int      `yaml:"min_items,omitempty" json:"min_items,omitempty" lc:"minimum items per list"`
}

// WordCountRule defines word count constraints for a section
type WordCountRule struct {
	Min int `yaml:"min,omitempty" json:"min,omitempty" lc:"minimum words"`
	Max int `yaml:"max,omitempty" json:"max,omitempty" lc:"maximum words"`
}

// CountConstraint defines how many times a structure element can match
type CountConstraint struct {
	Min int `yaml:"min,omitempty" json:"min,omitempty" lc:"minimum occurrences required"`
	Max int `yaml:"max,omitempty" json:"max,omitempty" lc:"maximum occurrences allowed (0 = unlimited)"`
}

// HeadingRules defines global validation rules for document headings
type HeadingRules struct {
	// NoSkipLevels ensures heading levels are not skipped (e.g., h1 -> h3 without h2)
	NoSkipLevels bool `yaml:"no_skip_levels,omitempty" json:"no_skip_levels,omitempty" lc:"disallow skipping levels (e.g., h1 -> h3)"`

	// Unique ensures all headings in the document are unique
	Unique bool `yaml:"unique,omitempty" json:"unique,omitempty" lc:"all headings must be unique"`

	// UniquePerLevel ensures headings are unique within the same level
	UniquePerLevel bool `yaml:"unique_per_level,omitempty" json:"unique_per_level,omitempty" lc:"headings unique within same level"`

	// MaxDepth limits the maximum heading depth (1-6, where 1 is h1)
	MaxDepth int `yaml:"max_depth,omitempty" json:"max_depth,omitempty" lc:"maximum heading depth (1-6)"`
}

// FrontmatterConfig defines validation rules for YAML frontmatter
type FrontmatterConfig struct {
	// Optional indicates frontmatter block is not required (default: false = required)
	Optional bool `yaml:"optional,omitempty" json:"optional,omitempty" lc:"frontmatter block is not required"`

	// Fields defines the required/optional fields and their constraints
	Fields []FrontmatterField `yaml:"fields,omitempty" json:"fields,omitempty" lc:"field definitions"`
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

// JSONSchema implements jsonschema.JSONSchemer to add enum constraint
func (FieldType) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "string",
		Enum:        []any{"string", "number", "boolean", "array", "date"},
		Description: "Field type: string, number, boolean, array, or date",
	}
}

// FieldFormat represents the format of a frontmatter field
type FieldFormat string

// Field format constants for frontmatter validation
const (
	FieldFormatDate  FieldFormat = "date"  // YYYY-MM-DD
	FieldFormatEmail FieldFormat = "email" // valid email address
	FieldFormatURL   FieldFormat = "url"   // http:// or https://
)

// JSONSchema implements jsonschema.JSONSchemer to add enum constraint
func (FieldFormat) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "string",
		Enum:        []any{"date", "email", "url"},
		Description: "Field format: date (YYYY-MM-DD), email, or url",
	}
}

// FrontmatterField defines a single frontmatter field requirement
type FrontmatterField struct {
	// Name is the field name (required)
	Name string `yaml:"name" json:"name" lc:"field name"`

	// Optional indicates whether this field is not required (default: false = required)
	Optional bool `yaml:"optional,omitempty" json:"optional,omitempty" lc:"field is not required"`

	// Type is the expected type (use FieldType* constants)
	Type FieldType `yaml:"type,omitempty" json:"type,omitempty" lc:"string, number, boolean, array, date"`

	// Format specifies format validation (use FieldFormat* constants)
	Format FieldFormat `yaml:"format,omitempty" json:"format,omitempty" lc:"date, email, or url"`
}
