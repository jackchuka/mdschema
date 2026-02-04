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

	// Check for missing required elements (respecting count constraints)
	ctx.Tree.Walk(func(n *vast.Node) bool {
		if !n.IsBound && !isElementOptional(n.Element) {
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

	// Validate count constraints
	violations = append(violations, r.validateCountConstraints(ctx)...)

	// Check ordering within each level
	violations = append(violations, r.validateOrderingIssues(ctx)...)

	return violations
}

// isElementOptional returns true if the element is optional (0 min matches).
func isElementOptional(element schema.StructureElement) bool {
	if element.Count != nil {
		return element.Count.Min == 0
	}
	return element.Optional
}

func hasUnboundAncestor(n *vast.Node) bool {
	for p := n.Parent; p != nil; p = p.Parent {
		if !p.IsBound {
			return true
		}
	}
	return false
}

// validateCountConstraints validates that multi-match elements have the correct number of matches.
func (r *StructureRule) validateCountConstraints(ctx *vast.Context) []Violation {
	if ctx == nil || ctx.Tree == nil || ctx.Schema == nil {
		return nil
	}

	violations := make([]Violation, 0)

	// Check root-level elements
	violations = append(violations, r.checkCountForElements(ctx.Schema.Structure, ctx.Tree.Roots, "document root")...)

	// Check children of each bound node
	ctx.Tree.WalkBound(func(n *vast.Node) bool {
		if len(n.Element.Children) > 0 {
			parentName := n.HeadingText()
			violations = append(violations, r.checkCountForElements(n.Element.Children, n.Children, parentName)...)
		}
		return true
	})

	return violations
}

// checkCountForElements validates count constraints for a set of schema elements against actual nodes.
func (r *StructureRule) checkCountForElements(elements []schema.StructureElement, nodes []*vast.Node, parentName string) []Violation {
	violations := make([]Violation, 0)

	for _, element := range elements {
		if element.Count == nil {
			continue // No count constraint
		}

		// Count how many nodes match this element
		count := 0
		for _, node := range nodes {
			if node.IsBound && elementMatches(node.Element, element) {
				count++
			}
		}

		minMatches := element.Count.Min
		maxMatches := element.Count.Max

		// Get a readable element name
		elementName := element.Heading.GetReadableName()

		// Check minimum constraint
		if count < minMatches {
			violations = append(violations,
				NewViolation(r.Name(),
					fmt.Sprintf("Element %q within %q requires at least %d occurrence(s), found %d",
						elementName, parentName, minMatches, count),
					1, 1).
					WithSeverity(severityFromSchema(element.Severity)))
		}

		// Check maximum constraint (0 means unlimited)
		if maxMatches > 0 && count > maxMatches {
			violations = append(violations,
				NewViolation(r.Name(),
					fmt.Sprintf("Element %q within %q allows at most %d occurrence(s), found %d",
						elementName, parentName, maxMatches, count),
					1, 1).
					WithSeverity(severityFromSchema(element.Severity)))
		}
	}

	return violations
}

// elementMatches checks if a node's element matches the given schema element.
func elementMatches(nodeElement, schemaElement schema.StructureElement) bool {
	// Compare by heading pattern (all fields)
	return nodeElement.Heading.Pattern == schemaElement.Heading.Pattern &&
		nodeElement.Heading.Literal == schemaElement.Heading.Literal &&
		nodeElement.Heading.Expr == schemaElement.Heading.Expr
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
				if !r.matcher.MatchesHeading(firstActual.Heading, firstExpected.Heading, ctx.Tree.Document.Path) {
					actualHeading := strings.Repeat("#", firstActual.Heading.Level) + " " + firstActual.Heading.Text
					var msg string
					if firstExpected.Heading.Expr != "" {
						// For expressions, show helpful debugging info
						filename := vast.ExtractFilename(ctx.Tree.Document.Path)
						msg = fmt.Sprintf("Heading %q does not match expression %q (filename=%q, heading=%q)",
							actualHeading, firstExpected.Heading.Expr, filename, firstActual.Heading.Text)
					} else {
						msg = fmt.Sprintf("First heading under %q is %q but expected %q",
							"document root", actualHeading, firstExpected.Heading.Pattern)
					}
					violations = append(violations,
						NewViolation(r.Name(), msg, firstActual.Heading.Line, firstActual.Heading.Column).
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

		if !r.matcher.MatchesHeading(firstActual.Heading, firstExpected.Heading, ctx.Tree.Document.Path) {
			actualHeading := strings.Repeat("#", firstActual.Heading.Level) + " " + firstActual.Heading.Text
			parentName := n.Section.Heading.Text
			var msg string
			if firstExpected.Heading.Expr != "" {
				filename := vast.ExtractFilename(ctx.Tree.Document.Path)
				msg = fmt.Sprintf("Heading %q under %q does not match expression %q (filename=%q, heading=%q)",
					actualHeading, parentName, firstExpected.Heading.Expr, filename, firstActual.Heading.Text)
			} else {
				msg = fmt.Sprintf("First heading under %q is %q but expected %q",
					parentName, actualHeading, firstExpected.Heading.Pattern)
			}
			violations = append(violations,
				NewViolation(r.Name(), msg, firstActual.Heading.Line, firstActual.Heading.Column).
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
	violations = append(violations, r.checkSiblingOrder(ctx.Tree.Roots, ctx.Tree.Document.Root.Children, ctx.Tree.Document.Path)...)

	// Check ordering within each bound node
	ctx.Tree.WalkBound(func(n *vast.Node) bool {
		if len(n.Children) > 0 && n.Section != nil {
			violations = append(violations, r.checkSiblingOrder(n.Children, n.Section.Children, ctx.Tree.Document.Path)...)
		}
		return true
	})

	return violations
}

// checkSiblingOrder checks if siblings appear in correct order.
// It detects when an unbound node's pattern matches a section that appears
// before a previously matched sibling (meaning it's out of order).
func (r *StructureRule) checkSiblingOrder(siblings []*vast.Node, sections []*parser.Section, documentPath string) []Violation {
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
				if r.matcher.MatchesHeading(section.Heading, node.Element.Heading, documentPath) {
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
