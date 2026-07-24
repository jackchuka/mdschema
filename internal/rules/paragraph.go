package rules

import (
	"fmt"
	"strings"

	"github.com/jackchuka/mdschema/internal/schema"
	"github.com/jackchuka/mdschema/internal/vast"
)

// ParagraphRule validates paragraph-count requirements for sections
type ParagraphRule struct {
}

var _ StructuralRule = (*ParagraphRule)(nil)

// NewParagraphRule creates a new paragraph rule
func NewParagraphRule() *ParagraphRule {
	return &ParagraphRule{}
}

// Name returns the rule identifier
func (r *ParagraphRule) Name() string {
	return "paragraph"
}

// ValidateWithContext validates using VAST (validation-ready AST)
func (r *ParagraphRule) ValidateWithContext(ctx *vast.Context) []Violation {
	violations := make([]Violation, 0)

	ctx.Tree.WalkBound(func(n *vast.Node) bool {
		if n.Element.SectionRules != nil && n.Element.Paragraphs != nil {
			rule := n.Element.Paragraphs
			count := len(n.Paragraphs())
			line, col := n.Location()

			if rule.Min > 0 && count < rule.Min {
				violations = append(violations,
					NewViolation(r.Name(), fmt.Sprintf("Section '%s' has too few paragraphs (minimum %d, found %d)", n.HeadingText(), rule.Min, count), line, col))
			}

			if rule.Max > 0 && count > rule.Max {
				violations = append(violations,
					NewViolation(r.Name(), fmt.Sprintf("Section '%s' has too many paragraphs (maximum %d, found %d)", n.HeadingText(), rule.Max, count), line, col))
			}
		}
		return true
	})

	return violations
}

// GenerateContent generates placeholder content for paragraph rules
func (r *ParagraphRule) GenerateContent(builder *strings.Builder, element schema.StructureElement) bool {
	if element.SectionRules == nil || element.Paragraphs == nil {
		return false
	}

	rule := element.Paragraphs

	builder.WriteString("<!-- Paragraph requirements: -->\n")
	if rule.Min > 0 {
		fmt.Fprintf(builder, "<!-- Minimum %d paragraphs required -->\n", rule.Min)
	}
	if rule.Max > 0 {
		fmt.Fprintf(builder, "<!-- Maximum %d paragraphs allowed -->\n", rule.Max)
	}
	builder.WriteString("\n")

	return true
}
