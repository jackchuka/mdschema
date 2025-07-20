package rules

import (
	"fmt"
	"strings"

	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/schema"
)

// CodeBlockRule validates code blocks using the hierarchical AST
type CodeBlockRule struct {
}

var _ ContextualRule = (*CodeBlockRule)(nil)

// NewCodeBlockRule creates a new simplified code block rule
func NewCodeBlockRule() *CodeBlockRule {
	return &CodeBlockRule{}
}

// Name returns the rule identifier
func (r *CodeBlockRule) Name() string {
	return "codeblock"
}

// validateCodeBlockRequirement validates a specific code block requirement for a section
func (r *CodeBlockRule) validateCodeBlockRequirement(section *parser.Section, requirement schema.CodeBlockRule) []Violation {
	violations := make([]Violation, 0)

	// Count matching code blocks directly from the section
	count := 0
	for _, block := range section.CodeBlocks {
		if requirement.Lang == "" || block.Lang == requirement.Lang {
			count++
		}
	}

	// Check minimum requirement
	if requirement.Min > 0 && count < requirement.Min {
		message := fmt.Sprintf("Section '%s' requires at least %d code blocks", section.Heading.Text, requirement.Min)
		if requirement.Lang != "" {
			message = fmt.Sprintf("Section '%s' requires at least %d '%s' code blocks, found %d",
				section.Heading.Text, requirement.Min, requirement.Lang, count)
		}

		violations = append(violations, Violation{
			Rule:    r.Name(),
			Message: message,
			Line:    section.Heading.Line,
			Column:  section.Heading.Column,
		})
	}

	// Check maximum requirement
	if requirement.Max > 0 && count > requirement.Max {
		message := fmt.Sprintf("Section '%s' has too many code blocks (max %d)", section.Heading.Text, requirement.Max)
		if requirement.Lang != "" {
			message = fmt.Sprintf("Section '%s' has too many '%s' code blocks (max %d, found %d)",
				section.Heading.Text, requirement.Lang, requirement.Max, count)
		}

		violations = append(violations, Violation{
			Rule:    r.Name(),
			Message: message,
			Line:    section.Heading.Line,
			Column:  section.Heading.Column,
		})
	}

	return violations
}

// ValidateWithContext validates using pre-established section-schema mappings (no string matching)
func (r *CodeBlockRule) ValidateWithContext(ctx *ValidationContext) []Violation {
	violations := make([]Violation, 0)

	// Walk through all mappings to find elements with code block rules
	violations = append(violations, r.validateCodeBlockMappings(ctx.Mappings)...)

	return violations
}

// GenerateContent generates placeholder code blocks for code block rules
func (r *CodeBlockRule) GenerateContent(builder *strings.Builder, element schema.StructureElement) bool {
	if element.SectionRules == nil || len(element.CodeBlocks) == 0 {
		return false
	}

	// Add code block placeholders
	builder.WriteString("<!-- Code block requirements: -->\n")
	for _, rule := range element.CodeBlocks {
		if rule.Lang != "" {
			if rule.Min > 0 {
				fmt.Fprintf(builder, "<!-- Minimum %d %s code blocks required -->\n", rule.Min, rule.Lang)
			}
			if rule.Max > 0 {
				fmt.Fprintf(builder, "<!-- Maximum %d %s code blocks allowed -->\n", rule.Max, rule.Lang)
			}
		} else {
			if rule.Min > 0 {
				fmt.Fprintf(builder, "<!-- Minimum %d code blocks required -->\n", rule.Min)
			}
			if rule.Max > 0 {
				fmt.Fprintf(builder, "<!-- Maximum %d code blocks allowed -->\n", rule.Max)
			}
		}
	}
	builder.WriteString("\n")

	// Add placeholder code blocks
	for _, rule := range element.CodeBlocks {
		if rule.Min > 0 {
			lang := rule.Lang
			if lang == "" {
				lang = "text"
			}
			// Generate the minimum required number of code blocks
			for i := 0; i < rule.Min; i++ {
				fmt.Fprintf(builder, "```%s\n", lang)
				builder.WriteString("// TODO: Add your code here\n")
				builder.WriteString("```\n\n")
			}
		}
	}

	return true
}

// validateCodeBlockMappings recursively validates code block requirements for all mappings
func (r *CodeBlockRule) validateCodeBlockMappings(mappings []*SectionMapping) []Violation {
	violations := make([]Violation, 0)

	for _, mapping := range mappings {
		// Check if this element has code block rules
		if mapping.Element.SectionRules != nil && len(mapping.Element.CodeBlocks) > 0 {
			// Validate code block requirements in all matching sections
			for _, section := range mapping.Sections {
				for _, requirement := range mapping.Element.CodeBlocks {
					violations = append(violations, r.validateCodeBlockRequirement(section, requirement)...)
				}
			}
		}

		// Recursively validate children
		violations = append(violations, r.validateCodeBlockMappings(mapping.Children)...)
	}

	return violations
}
