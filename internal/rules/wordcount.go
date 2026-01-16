package rules

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/jackchuka/mdschema/internal/schema"
	"github.com/jackchuka/mdschema/internal/vast"
)

// WordCountRule validates word count requirements for sections
type WordCountRule struct {
}

var _ StructuralRule = (*WordCountRule)(nil)

// NewWordCountRule creates a new word count rule
func NewWordCountRule() *WordCountRule {
	return &WordCountRule{}
}

// Name returns the rule identifier
func (r *WordCountRule) Name() string {
	return "word-count"
}

// countWords counts words in the content (whitespace-separated tokens)
func (r *WordCountRule) countWords(content string) int {
	// Split on whitespace and count non-empty tokens
	count := 0
	inWord := false

	for _, r := range content {
		if unicode.IsSpace(r) {
			if inWord {
				count++
				inWord = false
			}
		} else {
			inWord = true
		}
	}

	// Count last word if content doesn't end with whitespace
	if inWord {
		count++
	}

	return count
}

// ValidateWithContext validates using VAST (validation-ready AST)
func (r *WordCountRule) ValidateWithContext(ctx *vast.Context) []Violation {
	violations := make([]Violation, 0)

	// Walk through all bound nodes to find elements with word count rules
	ctx.Tree.WalkBound(func(n *vast.Node) bool {
		if n.Element.SectionRules != nil && n.Element.WordCount != nil {
			rule := n.Element.WordCount
			wordCount := r.countWords(n.Content())
			line, col := n.Location()

			// Check minimum requirement
			if rule.Min > 0 && wordCount < rule.Min {
				violations = append(violations,
					NewViolation(r.Name(), fmt.Sprintf("Section '%s' has too few words (minimum %d, found %d)", n.HeadingText(), rule.Min, wordCount), line, col))
			}

			// Check maximum requirement
			if rule.Max > 0 && wordCount > rule.Max {
				violations = append(violations,
					NewViolation(r.Name(), fmt.Sprintf("Section '%s' has too many words (maximum %d, found %d)", n.HeadingText(), rule.Max, wordCount), line, col))
			}
		}
		return true
	})

	return violations
}

// GenerateContent generates placeholder content for word count rules
func (r *WordCountRule) GenerateContent(builder *strings.Builder, element schema.StructureElement) bool {
	if element.SectionRules == nil || element.WordCount == nil {
		return false
	}

	rule := element.WordCount

	// Add word count requirement comments
	builder.WriteString("<!-- Word count requirements: -->\n")
	if rule.Min > 0 {
		fmt.Fprintf(builder, "<!-- Minimum %d words required -->\n", rule.Min)
	}
	if rule.Max > 0 {
		fmt.Fprintf(builder, "<!-- Maximum %d words allowed -->\n", rule.Max)
	}
	builder.WriteString("\n")

	return true
}
