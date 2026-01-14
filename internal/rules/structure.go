package rules

import (
	"fmt"
	"strings"

	"github.com/jackchuka/mdschema/internal/schema"
	"github.com/jackchuka/mdschema/internal/vast"
)

// StructureRule validates document structure using the hierarchical AST
type StructureRule struct {
}

var _ ContextualRule = (*StructureRule)(nil)

// NewStructureRule creates a new simplified structure rule
func NewStructureRule() *StructureRule {
	return &StructureRule{}
}

// Name returns the rule identifier
func (r *StructureRule) Name() string {
	return "structure"
}

// ValidateWithContext validates using VAST (validation-ready AST)
func (r *StructureRule) ValidateWithContext(ctx *vast.Context) []Violation {
	violations := make([]Violation, 0)

	// Check for missing required elements
	ctx.Tree.Walk(func(n *vast.Node) bool {
		if !n.IsBound && !n.Element.Optional {
			line, col := n.Location()
			parentName := "document root"
			if n.Parent != nil {
				parentName = n.Parent.HeadingText()
			}

			violations = append(violations, Violation{
				Rule:    r.Name(),
				Message: fmt.Sprintf("Required element '%s' not found within '%s'", n.Element.Heading.Pattern, parentName),
				Line:    line,
				Column:  col,
			})
		}
		return true
	})

	// Check ordering within each level
	violations = append(violations, r.validateOrdering(ctx.Tree.Roots)...)

	return violations
}

// validateOrdering validates structure ordering for all nodes at a level
func (r *StructureRule) validateOrdering(nodes []*vast.Node) []Violation {
	violations := make([]Violation, 0)

	lastLine := 0
	lastText := ""

	for _, n := range nodes {
		if n.IsBound {
			if lastLine > 0 && n.ActualOrder < lastLine {
				line, col := n.Location()
				violations = append(violations, Violation{
					Rule: r.Name(),
					Message: fmt.Sprintf("Element '%s' should appear after '%s' but appears before it",
						n.HeadingText(), lastText),
					Line:   line,
					Column: col,
				})
			}
			lastLine = n.ActualOrder
			lastText = n.HeadingText()

			// Recursively check children
			violations = append(violations, r.validateOrdering(n.Children)...)
		}
	}

	return violations
}

// GenerateContent generates structural organization and ordering information
func (r *StructureRule) GenerateContent(builder *strings.Builder, element schema.StructureElement) bool {
	// Add structural comments for required vs optional elements
	if !element.Optional {
		builder.WriteString("<!-- Required section -->\n")
	}

	// If this element has children, add ordering guidance
	if len(element.Children) > 0 {
		builder.WriteString("<!-- This section should contain the following subsections in order: -->\n")
		for i, child := range element.Children {
			status := "required"
			if child.Optional {
				status = "optional"
			}
			fmt.Fprintf(builder, "<!-- %d. %s (%s) -->\n", i+1, child.Heading.Pattern, status)
		}
		builder.WriteString("\n")
	}

	return true
}
