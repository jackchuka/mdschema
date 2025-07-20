package rules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jackchuka/mdschema/internal/schema"
)

// RequiredTextRule validates required text within sections
type RequiredTextRule struct {
}

var _ ContextualRule = (*RequiredTextRule)(nil)

// NewRequiredTextRule creates a new section content rule
func NewRequiredTextRule() *RequiredTextRule {
	return &RequiredTextRule{}
}

// Name returns the rule identifier
func (r *RequiredTextRule) Name() string {
	return "required-text"
}

// ValidateWithContext validates using pre-established section-schema mappings (no string matching)
func (r *RequiredTextRule) ValidateWithContext(ctx *ValidationContext) []Violation {
	violations := make([]Violation, 0)

	// Walk through all mappings to find elements with required text rules
	violations = append(violations, r.validateMappings(ctx.Mappings)...)

	return violations
}

// validateMappings recursively validates required text for all mappings
func (r *RequiredTextRule) validateMappings(mappings []*SectionMapping) []Violation {
	violations := make([]Violation, 0)

	for _, mapping := range mappings {
		// Check if this element has required text rules
		if mapping.Element.SectionRules != nil && len(mapping.Element.RequiredText) > 0 {
			// Validate required text in all matching sections
			for _, section := range mapping.Sections {
				for _, requiredText := range mapping.Element.RequiredText {
					if !r.contentContainsText(section.Content, requiredText) {
						violations = append(violations, Violation{
							Rule:    r.Name(),
							Message: fmt.Sprintf("Required text '%s' not found in section '%s'", requiredText, section.Heading.Text),
							Line:    section.Heading.Line,
							Column:  section.Heading.Column,
						})
					}
				}
			}
		}

		// Recursively validate children
		violations = append(violations, r.validateMappings(mapping.Children)...)
	}

	return violations
}

// GenerateContent generates placeholder content for required text rules
func (r *RequiredTextRule) GenerateContent(builder *strings.Builder, element schema.StructureElement) bool {
	if element.SectionRules == nil || len(element.RequiredText) == 0 {
		return false
	}

	// Add required text placeholders
	builder.WriteString("<!-- This section must contain the following text: -->\n")
	for _, text := range element.RequiredText {
		fmt.Fprintf(builder, "<!-- - %s -->\n", text)
	}
	builder.WriteString("\n")

	return true
}

// contentContainsText checks if content contains the required text (supports regex)
func (r *RequiredTextRule) contentContainsText(content, requiredText string) bool {
	// Check if it's a regex pattern (starts with ^ or contains regex metacharacters)
	if strings.HasPrefix(requiredText, "^") || strings.Contains(requiredText, "(?i)") {
		re, err := regexp.Compile(requiredText)
		if err != nil {
			// If regex compilation fails, fall back to substring match
			return strings.Contains(content, requiredText)
		}
		return re.MatchString(content)
	}

	// Simple substring match
	return strings.Contains(content, requiredText)
}
