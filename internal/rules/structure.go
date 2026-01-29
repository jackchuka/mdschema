package rules

import (
	"fmt"
	"strings"

	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/schema"
	"github.com/jackchuka/mdschema/internal/vast"
)

// severityFromSchema converts a schema severity string to rules.Severity
func severityFromSchema(s string) Severity {
	switch s {
	case "warning":
		return SeverityWarning
	case "info":
		return SeverityInfo
	default:
		return SeverityError
	}
}

// StructureRule validates document structure using the hierarchical AST
type StructureRule struct {
	matcher *vast.PatternMatcher
}

var _ StructuralRule = (*StructureRule)(nil)

// NewStructureRule creates a new simplified structure rule
func NewStructureRule() *StructureRule {
	return &StructureRule{
		matcher: vast.NewPatternMatcher(),
	}
}

// Name returns the rule identifier
func (r *StructureRule) Name() string {
	return "structure"
}

// ValidateWithContext validates using VAST (validation-ready AST)
func (r *StructureRule) ValidateWithContext(ctx *vast.Context) []Violation {
	violations := make([]Violation, 0)

	// Enforce that the first heading matches the first required element at each level.
	violations = append(violations, r.validateFirstHeadingIssues(ctx)...)

	// Report headings that are not defined in the structure.
	violations = append(violations, r.validateUnmatchedSections(ctx)...)

	// Check for missing required elements
	ctx.Tree.Walk(func(n *vast.Node) bool {
		if !n.IsBound && !n.Element.Optional {
			if hasUnboundAncestor(n) {
				return true
			}
			line, col := n.Location()
			parentName := "document root"
			if n.Parent != nil {
				parentName = n.Parent.HeadingText()
			}

			violations = append(violations,
				NewViolation(r.Name(), fmt.Sprintf("Required element %q not found within %q", n.Element.Heading.Pattern, parentName), line, col).
					WithSeverity(severityFromSchema(n.Element.Severity)))
		}
		return true
	})

	// Check ordering within each level
	violations = append(violations, r.validateOrderingIssues(ctx)...)

	return violations
}

func hasUnboundAncestor(n *vast.Node) bool {
	for p := n.Parent; p != nil; p = p.Parent {
		if !p.IsBound {
			return true
		}
	}
	return false
}

func (r *StructureRule) validateFirstHeadingIssues(ctx *vast.Context) []Violation {
	if ctx == nil || ctx.Tree == nil {
		return nil
	}

	violations := make([]Violation, 0)

	// Check document root level
	if len(ctx.Schema.Structure) > 0 && len(ctx.Tree.Document.Root.Children) > 0 {
		firstExpected := ctx.Schema.Structure[0]
		if !firstExpected.Optional {
			firstActual := ctx.Tree.Document.Root.Children[0]
			if firstActual.Heading != nil {
				if !r.matcher.MatchesHeadingPattern(firstActual.Heading, firstExpected.Heading.Pattern, firstExpected.Heading.Regex) {
					actualHeading := strings.Repeat("#", firstActual.Heading.Level) + " " + firstActual.Heading.Text
					violations = append(violations,
						NewViolation(r.Name(), fmt.Sprintf("First heading under %q is %q but expected %q", "document root", actualHeading, firstExpected.Heading.Pattern), firstActual.Heading.Line, firstActual.Heading.Column).
							WithSeverity(severityFromSchema(firstExpected.Severity)))
				}
			}
		}
	}

	// Check each bound node with children
	ctx.Tree.WalkBound(func(n *vast.Node) bool {
		if n.Section == nil || len(n.Element.Children) == 0 || len(n.Section.Children) == 0 {
			return true
		}

		firstExpected := n.Element.Children[0]
		if firstExpected.Optional {
			return true
		}

		firstActual := n.Section.Children[0]
		if firstActual.Heading == nil {
			return true
		}

		if !r.matcher.MatchesHeadingPattern(firstActual.Heading, firstExpected.Heading.Pattern, firstExpected.Heading.Regex) {
			actualHeading := strings.Repeat("#", firstActual.Heading.Level) + " " + firstActual.Heading.Text
			parentName := n.Section.Heading.Text
			violations = append(violations,
				NewViolation(r.Name(), fmt.Sprintf("First heading under %q is %q but expected %q", parentName, actualHeading, firstExpected.Heading.Pattern), firstActual.Heading.Line, firstActual.Heading.Column).
					WithSeverity(severityFromSchema(firstExpected.Severity)))
		}

		return true
	})

	return violations
}

func (r *StructureRule) validateUnmatchedSections(ctx *vast.Context) []Violation {
	if ctx == nil || ctx.Tree == nil {
		return nil
	}

	violations := make([]Violation, 0)
	for _, section := range ctx.Tree.UnmatchedSections {
		if section == nil || section.Heading == nil {
			continue
		}

		// Check if parent allows additional sections
		if r.ancestorAllowsAdditional(ctx, section) {
			continue
		}

		heading := strings.Repeat("#", section.Heading.Level) + " " + section.Heading.Text
		parent := "document root"
		if section.Parent != nil && section.Parent.Heading != nil {
			parent = section.Parent.Heading.Text
		}
		violations = append(violations,
			NewViolation(r.Name(), fmt.Sprintf("Unexpected section %q found under %q", heading, parent), section.Heading.Line, section.Heading.Column))
	}

	return violations
}

// ancestorAllowsAdditional recursively checks if any ancestor allows additional sections.
func (r *StructureRule) ancestorAllowsAdditional(ctx *vast.Context, section *parser.Section) bool {
	if section == nil {
		return false
	}

	// Root level - no heading means document root
	if section.Heading == nil {
		return false
	}

	// Check if this section is bound to a schema element
	var node *vast.Node
	ctx.Tree.Walk(func(n *vast.Node) bool {
		if n.Section == section {
			node = n
			return false // stop walking
		}
		return true
	})

	if node != nil {
		// This section is bound to a schema element
		return node.Element.AllowAdditional
	}

	// This section is unmatched - check if its parent allows additional
	// If so, this section and all its descendants are allowed
	return r.ancestorAllowsAdditional(ctx, section.Parent)
}

func (r *StructureRule) validateOrderingIssues(ctx *vast.Context) []Violation {
	if ctx == nil || ctx.Tree == nil {
		return nil
	}

	violations := make([]Violation, 0)

	// Check ordering at root level
	violations = append(violations, r.checkSiblingOrder(ctx.Tree.Roots, ctx.Tree.Document.Root.Children)...)

	// Check ordering within each bound node
	ctx.Tree.WalkBound(func(n *vast.Node) bool {
		if len(n.Children) > 0 && n.Section != nil {
			violations = append(violations, r.checkSiblingOrder(n.Children, n.Section.Children)...)
		}
		return true
	})

	return violations
}

// checkSiblingOrder checks if siblings appear in correct order.
// It detects when an unbound node's pattern matches a section that appears
// before a previously matched sibling (meaning it's out of order).
func (r *StructureRule) checkSiblingOrder(siblings []*vast.Node, sections []*parser.Section) []Violation {
	violations := make([]Violation, 0)

	// Track the maximum line number seen so far among bound siblings
	maxBoundLine := 0
	maxBoundText := ""

	for _, node := range siblings {
		if node.IsBound && node.Section != nil && node.Section.Heading != nil {
			// For bound nodes, just track the line progression
			if node.Section.StartLine > maxBoundLine {
				maxBoundLine = node.Section.StartLine
				maxBoundText = node.Section.Heading.Text
			}
		} else if !node.IsBound && maxBoundLine > 0 {
			// For unbound nodes, check if there's a matching section BEFORE maxBoundLine
			// This would indicate the section exists but is out of order
			for _, section := range sections {
				if section.Heading == nil || section.StartLine >= maxBoundLine {
					continue
				}
				if r.matcher.MatchesHeadingPattern(section.Heading, node.Element.Heading.Pattern, node.Element.Heading.Regex) {
					violations = append(violations,
						NewViolation(r.Name(), fmt.Sprintf("Element %q should appear after %q but appears before it", section.Heading.Text, maxBoundText), section.Heading.Line, section.Heading.Column).
							WithSeverity(severityFromSchema(node.Element.Severity)))
					break
				}
			}
		}
	}

	return violations
}

// GenerateContent generates structural organization and ordering information
func (r *StructureRule) GenerateContent(builder *strings.Builder, element schema.StructureElement) bool {
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
