package rules

import (
	"strings"

	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/schema"
	"github.com/jackchuka/mdschema/internal/vast"
)

// Rule is the base interface for all validation rules
type Rule interface {
	// Name returns the rule identifier
	Name() string

	// ValidateWithContext uses pre-established section-schema mappings via VAST
	ValidateWithContext(ctx *vast.Context) []Violation
}

// StructuralRule validates and generates content for structure elements (sections)
type StructuralRule interface {
	Rule

	// GenerateContent generates markdown content for elements that match this rule
	// Returns true if the rule handled content generation for this element
	GenerateContent(builder *strings.Builder, element schema.StructureElement) bool
}

// FrontmatterGenerator generates document-level frontmatter content
type FrontmatterGenerator interface {
	Generate(builder *strings.Builder, s *schema.Schema) bool
}

// Validator manages and runs all rules
type Validator struct {
	rules []Rule
}

// defaultStructuralRules returns the standard set of structural validation rules
func defaultStructuralRules() []StructuralRule {
	return []StructuralRule{
		NewStructureRule(),
		NewRequiredTextRule(),
		NewForbiddenTextRule(),
		NewCodeBlockRule(),
		NewImageRule(),
		NewTableRule(),
		NewListRule(),
		NewWordCountRule(),
	}
}

// defaultDocumentRules returns validation rules that operate at document level
func defaultDocumentRules() []Rule {
	return []Rule{
		NewHeadingRule(),
		NewLinkValidationRule(),
	}
}

// defaultRules returns all rules as base Rule interface
func defaultRules() []Rule {
	rules := make([]Rule, 0)
	for _, r := range defaultStructuralRules() {
		rules = append(rules, r)
	}
	rules = append(rules, defaultDocumentRules()...)
	rules = append(rules, NewFrontmatterRule())
	return rules
}

// NewValidator creates a new validator with default rules for v0.1 DSL
func NewValidator() *Validator {
	return &Validator{
		rules: defaultRules(),
	}
}

// Validate runs all rules against a document with a specified root directory.
// The rootDir is used for resolving absolute paths (e.g., /path links).
func (v *Validator) Validate(doc *parser.Document, s *schema.Schema, rootDir string) []Violation {
	violations := make([]Violation, 0)

	// Create validation context with VAST
	ctx := vast.NewContext(doc, s, rootDir)

	for _, rule := range v.rules {
		ruleViolations := rule.ValidateWithContext(ctx)
		violations = append(violations, ruleViolations...)
	}

	return violations
}

// Generator creates markdown content using rules
type Generator struct {
	structuralRules      []StructuralRule
	frontmatterGenerator FrontmatterGenerator
}

// NewGenerator creates a generator that uses the same rules as the validator
func NewGenerator() *Generator {
	return &Generator{
		structuralRules:      defaultStructuralRules(),
		frontmatterGenerator: NewFrontmatterRule(),
	}
}

// GenerateContent generates content for an element using all applicable rules
func (g *Generator) GenerateContent(builder *strings.Builder, element schema.StructureElement) {
	contentGenerated := false

	for _, rule := range g.structuralRules {
		if rule.GenerateContent(builder, element) {
			contentGenerated = true
		}
	}

	// If no rule generated content, add default placeholder
	if !contentGenerated {
		builder.WriteString("TODO: Add content for this section.\n\n")
	}
}

// GenerateFrontmatter generates document frontmatter using the frontmatter generator
func (g *Generator) GenerateFrontmatter(builder *strings.Builder, s *schema.Schema) {
	g.frontmatterGenerator.Generate(builder, s)
}
