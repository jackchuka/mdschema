package rules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jackchuka/mdschema/internal/schema"
	"github.com/jackchuka/mdschema/internal/vast"
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

// ValidateWithContext validates using VAST (validation-ready AST)
func (r *RequiredTextRule) ValidateWithContext(ctx *vast.Context) []Violation {
	violations := make([]Violation, 0)

	// Walk through all bound nodes to find elements with required text rules
	ctx.Tree.WalkBound(func(n *vast.Node) bool {
		if n.Element.SectionRules != nil && len(n.Element.RequiredText) > 0 {
			for _, pattern := range n.Element.RequiredText {
				if !r.contentContainsPattern(n.Content(), pattern.Pattern, pattern.Regex) {
					line, col := n.Location()
					violations = append(violations, Violation{
						Rule:    r.Name(),
						Message: fmt.Sprintf("Required text '%s' not found in section '%s'", pattern.Pattern, n.HeadingText()),
						Line:    line,
						Column:  col,
					})
				}
			}
		}
		return true
	})

	return violations
}

// GenerateContent generates placeholder content for required text rules
func (r *RequiredTextRule) GenerateContent(builder *strings.Builder, element schema.StructureElement) bool {
	if element.SectionRules == nil || len(element.RequiredText) == 0 {
		return false
	}

	// Add required text placeholders
	builder.WriteString("<!-- This section must contain the following text: -->\n")
	for _, pattern := range element.RequiredText {
		if pattern.Regex {
			fmt.Fprintf(builder, "<!-- - %s (regex) -->\n", pattern.Pattern)
		} else {
			fmt.Fprintf(builder, "<!-- - %s -->\n", pattern.Pattern)
		}
	}
	builder.WriteString("\n")

	return true
}

// contentContainsPattern checks if content contains the required pattern
func (r *RequiredTextRule) contentContainsPattern(content, pattern string, isRegex bool) bool {
	if isRegex {
		re, err := regexp.Compile(pattern)
		if err != nil {
			// If regex compilation fails, fall back to substring match
			return strings.Contains(content, pattern)
		}
		return re.MatchString(content)
	}

	// Simple substring match
	return strings.Contains(content, pattern)
}
