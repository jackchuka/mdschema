package rules

import (
	"fmt"
	"strings"

	"github.com/jackchuka/mdschema/internal/schema"
	"github.com/jackchuka/mdschema/internal/vast"
)

// TableRule validates tables using the hierarchical AST
type TableRule struct {
}

var _ ContextualRule = (*TableRule)(nil)

// NewTableRule creates a new table rule
func NewTableRule() *TableRule {
	return &TableRule{}
}

// Name returns the rule identifier
func (r *TableRule) Name() string {
	return "table"
}

// validateTableRequirement validates a specific table requirement for a node
func (r *TableRule) validateTableRequirement(n *vast.Node, requirement schema.TableRule) []Violation {
	violations := make([]Violation, 0)
	tables := n.Tables()

	line, col := n.Location()

	// Count tables
	count := len(tables)

	// Check minimum requirement
	if requirement.Min > 0 && count < requirement.Min {
		violations = append(violations, Violation{
			Rule:    r.Name(),
			Message: fmt.Sprintf("Section '%s' requires at least %d tables, found %d", n.HeadingText(), requirement.Min, count),
			Line:    line,
			Column:  col,
		})
	}

	// Check maximum requirement
	if requirement.Max > 0 && count > requirement.Max {
		violations = append(violations, Violation{
			Rule:    r.Name(),
			Message: fmt.Sprintf("Section '%s' has too many tables (max %d, found %d)", n.HeadingText(), requirement.Max, count),
			Line:    line,
			Column:  col,
		})
	}

	// Check minimum columns requirement
	if requirement.MinColumns > 0 {
		for _, table := range tables {
			if len(table.Headers) < requirement.MinColumns {
				violations = append(violations, Violation{
					Rule:    r.Name(),
					Message: fmt.Sprintf("Table in section '%s' has too few columns (minimum %d, found %d)", n.HeadingText(), requirement.MinColumns, len(table.Headers)),
					Line:    table.Line,
					Column:  table.Column,
				})
			}
		}
	}

	// Check required headers
	if len(requirement.RequiredHeaders) > 0 {
		for _, table := range tables {
			headerSet := make(map[string]bool)
			for _, h := range table.Headers {
				headerSet[strings.ToLower(strings.TrimSpace(h))] = true
			}

			for _, required := range requirement.RequiredHeaders {
				if !headerSet[strings.ToLower(required)] {
					violations = append(violations, Violation{
						Rule:    r.Name(),
						Message: fmt.Sprintf("Table in section '%s' is missing required header '%s'", n.HeadingText(), required),
						Line:    table.Line,
						Column:  table.Column,
					})
				}
			}
		}
	}

	return violations
}

// ValidateWithContext validates using VAST (validation-ready AST)
func (r *TableRule) ValidateWithContext(ctx *vast.Context) []Violation {
	violations := make([]Violation, 0)

	// Walk through all bound nodes to find elements with table rules
	ctx.Tree.WalkBound(func(n *vast.Node) bool {
		if n.Element.SectionRules != nil && len(n.Element.Tables) > 0 {
			for _, requirement := range n.Element.Tables {
				violations = append(violations, r.validateTableRequirement(n, requirement)...)
			}
		}
		return true
	})

	return violations
}

// GenerateContent generates placeholder tables for table rules
func (r *TableRule) GenerateContent(builder *strings.Builder, element schema.StructureElement) bool {
	if element.SectionRules == nil || len(element.Tables) == 0 {
		return false
	}

	// Add table requirement comments
	builder.WriteString("<!-- Table requirements: -->\n")
	for _, rule := range element.Tables {
		if rule.Min > 0 {
			fmt.Fprintf(builder, "<!-- Minimum %d tables required -->\n", rule.Min)
		}
		if rule.Max > 0 {
			fmt.Fprintf(builder, "<!-- Maximum %d tables allowed -->\n", rule.Max)
		}
		if rule.MinColumns > 0 {
			fmt.Fprintf(builder, "<!-- Minimum %d columns required -->\n", rule.MinColumns)
		}
		if len(rule.RequiredHeaders) > 0 {
			fmt.Fprintf(builder, "<!-- Required headers: %s -->\n", strings.Join(rule.RequiredHeaders, ", "))
		}
	}
	builder.WriteString("\n")

	// Add placeholder tables
	for _, rule := range element.Tables {
		if rule.Min > 0 {
			for i := 0; i < rule.Min; i++ {
				// Generate table with required headers or default headers
				headers := rule.RequiredHeaders
				if len(headers) == 0 {
					cols := rule.MinColumns
					if cols == 0 {
						cols = 2
					}
					headers = make([]string, cols)
					for j := range headers {
						headers[j] = fmt.Sprintf("Column %d", j+1)
					}
				}

				// Write table header
				builder.WriteString("| ")
				builder.WriteString(strings.Join(headers, " | "))
				builder.WriteString(" |\n")

				// Write separator
				builder.WriteString("|")
				for range headers {
					builder.WriteString(" --- |")
				}
				builder.WriteString("\n")

				// Write placeholder row
				builder.WriteString("| ")
				placeholders := make([]string, len(headers))
				for j := range placeholders {
					placeholders[j] = "TODO"
				}
				builder.WriteString(strings.Join(placeholders, " | "))
				builder.WriteString(" |\n\n")
			}
		}
	}

	return true
}
