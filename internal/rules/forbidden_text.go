package rules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jackchuka/mdschema/internal/schema"
	"github.com/jackchuka/mdschema/internal/vast"
)

// ForbiddenTextRule validates that forbidden text patterns do not appear in sections
type ForbiddenTextRule struct {
}

var _ StructuralRule = (*ForbiddenTextRule)(nil)

// NewForbiddenTextRule creates a new forbidden text rule
func NewForbiddenTextRule() *ForbiddenTextRule {
	return &ForbiddenTextRule{}
}

// Name returns the rule identifier
func (r *ForbiddenTextRule) Name() string {
	return "forbidden-text"
}

// ValidateWithContext validates using VAST (validation-ready AST)
func (r *ForbiddenTextRule) ValidateWithContext(ctx *vast.Context) []Violation {
	violations := make([]Violation, 0)

	// Walk through all bound nodes to find elements with forbidden text rules
	ctx.Tree.WalkBound(func(n *vast.Node) bool {
		if n.Element.SectionRules != nil && len(n.Element.ForbiddenText) > 0 {
			for _, pattern := range n.Element.ForbiddenText {
				if r.contentContainsPattern(n.Content(), pattern) {
					patternStr := pattern.Literal
					if patternStr == "" {
						patternStr = pattern.Pattern
					}
					line, col := n.Location()
					violations = append(violations,
						NewViolation(r.Name(), fmt.Sprintf("Forbidden text '%s' found in section '%s'", patternStr, n.HeadingText()), line, col))
				}
			}
		}
		return true
	})

	return violations
}

// GenerateContent generates placeholder content for forbidden text rules
func (r *ForbiddenTextRule) GenerateContent(builder *strings.Builder, element schema.StructureElement) bool {
	if element.SectionRules == nil || len(element.ForbiddenText) == 0 {
		return false
	}

	// Add forbidden text warnings
	builder.WriteString("<!-- WARNING: This section must NOT contain the following: -->\n")
	for _, pattern := range element.ForbiddenText {
		if pattern.Pattern != "" {
			fmt.Fprintf(builder, "<!-- - %s (regex) -->\n", pattern.Pattern)
		} else {
			fmt.Fprintf(builder, "<!-- - %s -->\n", pattern.Literal)
		}
	}
	builder.WriteString("\n")

	return true
}

// contentContainsPattern checks if content contains the forbidden pattern
func (r *ForbiddenTextRule) contentContainsPattern(content string, pattern schema.ForbiddenTextPattern) bool {
	// If literal is set (scalar form), use substring match
	if pattern.Literal != "" {
		return strings.Contains(content, pattern.Literal)
	}

	// Otherwise use regex match
	re, err := regexp.Compile(pattern.Pattern)
	if err != nil {
		// If regex compilation fails, fall back to substring match
		return strings.Contains(content, pattern.Pattern)
	}
	return re.MatchString(content)
}
