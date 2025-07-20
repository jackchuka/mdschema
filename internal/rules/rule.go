package rules

import (
	"strings"

	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/schema"
)

// ContextualRule is an enhanced rule interface that uses pre-built mappings
type ContextualRule interface {
	// Name returns the rule identifier
	Name() string

	// ValidateWithContext uses pre-established section-schema mappings
	ValidateWithContext(ctx *ValidationContext) []Violation

	// GenerateContent generates markdown content for elements that match this rule
	// Returns true if the rule handled content generation for this element
	GenerateContent(builder *strings.Builder, element schema.StructureElement) bool
}

// Validator manages and runs all rules
type Validator struct {
	rules []ContextualRule
}

// NewValidator creates a new validator with default rules for v0.1 DSL
func NewValidator() *Validator {
	return &Validator{
		rules: []ContextualRule{
			NewStructureRule(),
			NewRequiredTextRule(),
			NewCodeBlockRule(),
			// Add more rules as needed
		},
	}
}

// Validate runs all rules against a document
func (v *Validator) Validate(doc *parser.Document, schema *schema.Schema) []Violation {
	violations := make([]Violation, 0)

	// Create validation context with pre-built mappings
	ctx := NewValidationContext(doc, schema)

	for _, rule := range v.rules {
		// Try to use contextual validation if the rule supports it
		ruleViolations := rule.ValidateWithContext(ctx)
		violations = append(violations, ruleViolations...)
	}

	return violations
}

// NewGenerator creates a generator that uses the same rules as the validator
func NewGenerator() *Generator {
	return &Generator{
		rules: []ContextualRule{
			NewStructureRule(), // Generates structural guidance and ordering info
			NewRequiredTextRule(),
			NewCodeBlockRule(),
		},
	}
}

// Generator creates markdown content using rules
type Generator struct {
	rules []ContextualRule
}

// GenerateContent generates content for an element using all applicable rules
func (g *Generator) GenerateContent(builder *strings.Builder, element schema.StructureElement) {
	contentGenerated := false

	// Let each rule try to generate content for this element
	for _, rule := range g.rules {
		if rule.GenerateContent(builder, element) {
			contentGenerated = true
		}
	}

	// If no rule generated content, add default placeholder
	if !contentGenerated {
		builder.WriteString("TODO: Add content for this section.\n\n")
	}
}
