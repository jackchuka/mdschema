package rules

import (
	"fmt"
	"strings"

	"github.com/jackchuka/mdschema/internal/schema"
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

// ValidateWithContext validates using pre-established section-schema mappings (no string matching)
func (r *StructureRule) ValidateWithContext(ctx *ValidationContext) []Violation {
	violations := make([]Violation, 0)

	// Validate mappings to check for missing required elements and ordering
	violations = append(violations, r.validateMappings(ctx.Mappings, nil)...)

	return violations
}

// validateMappings validates structure and ordering for all mappings
func (r *StructureRule) validateMappings(mappings []*SectionMapping, parent *SectionMapping) []Violation {
	violations := make([]Violation, 0)

	// Track ordering violations
	lastFoundLine := 0
	lastFoundElement := ""

	for _, mapping := range mappings {
		// Check if required element is missing
		if !mapping.Element.Optional && len(mapping.Sections) == 0 {
			parentName := "document root"
			line := 1
			col := 1

			if parent != nil && len(parent.Sections) > 0 && parent.Sections[0].Heading != nil {
				parentName = parent.Sections[0].Heading.Text
				line = parent.Sections[0].Heading.Line
				col = parent.Sections[0].Heading.Column
			}

			violations = append(violations, Violation{
				Rule:    r.Name(),
				Message: fmt.Sprintf("Required element '%s' not found within '%s'", mapping.Element.Heading, parentName),
				Line:    line,
				Column:  col,
			})
		}

		// Check ordering for found elements
		for _, section := range mapping.Sections {
			if section.Heading != nil {
				if lastFoundLine > 0 && section.StartLine < lastFoundLine {
					violations = append(violations, Violation{
						Rule: r.Name(),
						Message: fmt.Sprintf("Element '%s' should appear after '%s' but appears before it",
							section.Heading.Text, lastFoundElement),
						Line:   section.Heading.Line,
						Column: section.Heading.Column,
					})
				}
				lastFoundLine = section.StartLine
				lastFoundElement = section.Heading.Text

				// Recursively validate children
				violations = append(violations, r.validateMappings(mapping.Children, mapping)...)
			}
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
			fmt.Fprintf(builder, "<!-- %d. %s (%s) -->\n", i+1, child.Heading, status)
		}
		builder.WriteString("\n")
	}

	return true
}
