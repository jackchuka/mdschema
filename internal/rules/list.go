package rules

import (
	"fmt"
	"strings"

	"github.com/jackchuka/mdschema/internal/schema"
	"github.com/jackchuka/mdschema/internal/vast"
)

// ListRule validates lists using the hierarchical AST
type ListRule struct {
}

var _ ContextualRule = (*ListRule)(nil)

// NewListRule creates a new list rule
func NewListRule() *ListRule {
	return &ListRule{}
}

// Name returns the rule identifier
func (r *ListRule) Name() string {
	return "list"
}

// validateListRequirement validates a specific list requirement for a node
func (r *ListRule) validateListRequirement(n *vast.Node, requirement schema.ListRule) []Violation {
	violations := make([]Violation, 0)
	lists := n.Lists()

	line, col := n.Location()

	// Filter lists by type if specified
	var matchingLists int
	for _, list := range lists {
		if requirement.Type == "" {
			// Any list type matches
			matchingLists++
		} else if requirement.Type == schema.ListTypeOrdered && list.IsOrdered {
			matchingLists++
		} else if requirement.Type == schema.ListTypeUnordered && !list.IsOrdered {
			matchingLists++
		}
	}

	// Check minimum requirement
	if requirement.Min > 0 && matchingLists < requirement.Min {
		message := fmt.Sprintf("Section '%s' requires at least %d lists, found %d",
			n.HeadingText(), requirement.Min, matchingLists)
		if requirement.Type != "" {
			message = fmt.Sprintf("Section '%s' requires at least %d %s lists, found %d",
				n.HeadingText(), requirement.Min, requirement.Type, matchingLists)
		}

		violations = append(violations, Violation{
			Rule:    r.Name(),
			Message: message,
			Line:    line,
			Column:  col,
		})
	}

	// Check maximum requirement
	if requirement.Max > 0 && matchingLists > requirement.Max {
		message := fmt.Sprintf("Section '%s' has too many lists (max %d, found %d)",
			n.HeadingText(), requirement.Max, matchingLists)
		if requirement.Type != "" {
			message = fmt.Sprintf("Section '%s' has too many %s lists (max %d, found %d)",
				n.HeadingText(), requirement.Type, requirement.Max, matchingLists)
		}

		violations = append(violations, Violation{
			Rule:    r.Name(),
			Message: message,
			Line:    line,
			Column:  col,
		})
	}

	return violations
}

// ValidateWithContext validates using VAST (validation-ready AST)
func (r *ListRule) ValidateWithContext(ctx *vast.Context) []Violation {
	violations := make([]Violation, 0)

	// Walk through all bound nodes to find elements with list rules
	ctx.Tree.WalkBound(func(n *vast.Node) bool {
		if n.Element.SectionRules != nil && len(n.Element.Lists) > 0 {
			for _, requirement := range n.Element.Lists {
				violations = append(violations, r.validateListRequirement(n, requirement)...)
			}
		}
		return true
	})

	return violations
}

// GenerateContent generates placeholder lists for list rules
func (r *ListRule) GenerateContent(builder *strings.Builder, element schema.StructureElement) bool {
	if element.SectionRules == nil || len(element.Lists) == 0 {
		return false
	}

	// Add list requirement comments
	builder.WriteString("<!-- List requirements: -->\n")
	for _, rule := range element.Lists {
		if rule.Min > 0 {
			if rule.Type != "" {
				fmt.Fprintf(builder, "<!-- Minimum %d %s lists required -->\n", rule.Min, rule.Type)
			} else {
				fmt.Fprintf(builder, "<!-- Minimum %d lists required -->\n", rule.Min)
			}
		}
		if rule.Max > 0 {
			if rule.Type != "" {
				fmt.Fprintf(builder, "<!-- Maximum %d %s lists allowed -->\n", rule.Max, rule.Type)
			} else {
				fmt.Fprintf(builder, "<!-- Maximum %d lists allowed -->\n", rule.Max)
			}
		}
		if rule.MinItems > 0 {
			fmt.Fprintf(builder, "<!-- Minimum %d items per list -->\n", rule.MinItems)
		}
	}
	builder.WriteString("\n")

	// Add placeholder lists
	for _, rule := range element.Lists {
		if rule.Min > 0 {
			for i := 0; i < rule.Min; i++ {
				isOrdered := rule.Type == schema.ListTypeOrdered
				itemCount := 3
				if rule.MinItems > 0 {
					itemCount = rule.MinItems
				}

				for j := 0; j < itemCount; j++ {
					if isOrdered {
						fmt.Fprintf(builder, "%d. TODO: Add list item\n", j+1)
					} else {
						builder.WriteString("- TODO: Add list item\n")
					}
				}
				builder.WriteString("\n")
			}
		}
	}

	return true
}
