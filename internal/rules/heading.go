package rules

import (
	"fmt"
	"strings"

	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/schema"
	"github.com/jackchuka/mdschema/internal/vast"
)

// HeadingRule validates heading structure across the document
type HeadingRule struct {
}

var _ ContextualRule = (*HeadingRule)(nil)

// NewHeadingRule creates a new heading rule
func NewHeadingRule() *HeadingRule {
	return &HeadingRule{}
}

// Name returns the rule identifier
func (r *HeadingRule) Name() string {
	return "heading"
}

// collectHeadings recursively collects all headings from the document
func (r *HeadingRule) collectHeadings(section *parser.Section, headings *[]*parser.Heading) {
	if section.Heading != nil {
		*headings = append(*headings, section.Heading)
	}
	for _, child := range section.Children {
		r.collectHeadings(child, headings)
	}
}

// ValidateWithContext validates using VAST (validation-ready AST)
func (r *HeadingRule) ValidateWithContext(ctx *vast.Context) []Violation {
	violations := make([]Violation, 0)

	// Check if heading rules are configured
	if ctx.Schema.HeadingRules == nil {
		return violations
	}

	rules := ctx.Schema.HeadingRules

	// Collect all headings from the document
	var headings []*parser.Heading
	r.collectHeadings(ctx.Tree.Document.Root, &headings)

	// Validate no skip levels
	if rules.NoSkipLevels {
		violations = append(violations, r.validateNoSkipLevels(headings)...)
	}

	// Validate unique headings
	if rules.Unique {
		violations = append(violations, r.validateUniqueHeadings(headings)...)
	}

	// Validate unique per level
	if rules.UniquePerLevel {
		violations = append(violations, r.validateUniquePerLevel(headings)...)
	}

	// Validate max depth
	if rules.MaxDepth > 0 {
		violations = append(violations, r.validateMaxDepth(headings, rules.MaxDepth)...)
	}

	return violations
}

// validateNoSkipLevels checks that heading levels are not skipped
func (r *HeadingRule) validateNoSkipLevels(headings []*parser.Heading) []Violation {
	violations := make([]Violation, 0)

	if len(headings) == 0 {
		return violations
	}

	// Track the minimum level seen (first heading establishes the baseline)
	minLevel := headings[0].Level
	prevLevel := minLevel

	for i, h := range headings {
		if i == 0 {
			continue
		}

		// If heading level increases by more than 1, it's a skip
		if h.Level > prevLevel+1 {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Message: fmt.Sprintf("Heading level skipped: '%s' (h%d) after h%d", h.Text, h.Level, prevLevel),
				Line:    h.Line,
				Column:  h.Column,
			})
		}

		prevLevel = h.Level
	}

	return violations
}

// validateUniqueHeadings checks that all headings are unique
func (r *HeadingRule) validateUniqueHeadings(headings []*parser.Heading) []Violation {
	violations := make([]Violation, 0)

	seen := make(map[string]*parser.Heading)

	for _, h := range headings {
		normalizedText := strings.ToLower(strings.TrimSpace(h.Text))
		if existing, ok := seen[normalizedText]; ok {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Message: fmt.Sprintf("Duplicate heading '%s' (first occurrence at line %d)", h.Text, existing.Line),
				Line:    h.Line,
				Column:  h.Column,
			})
		} else {
			seen[normalizedText] = h
		}
	}

	return violations
}

// validateUniquePerLevel checks that headings are unique within each level
func (r *HeadingRule) validateUniquePerLevel(headings []*parser.Heading) []Violation {
	violations := make([]Violation, 0)

	// Map of level -> (text -> first heading)
	seenByLevel := make(map[int]map[string]*parser.Heading)

	for _, h := range headings {
		if seenByLevel[h.Level] == nil {
			seenByLevel[h.Level] = make(map[string]*parser.Heading)
		}

		normalizedText := strings.ToLower(strings.TrimSpace(h.Text))
		if existing, ok := seenByLevel[h.Level][normalizedText]; ok {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Message: fmt.Sprintf("Duplicate h%d heading '%s' (first occurrence at line %d)", h.Level, h.Text, existing.Line),
				Line:    h.Line,
				Column:  h.Column,
			})
		} else {
			seenByLevel[h.Level][normalizedText] = h
		}
	}

	return violations
}

// validateMaxDepth checks that no heading exceeds the maximum depth
func (r *HeadingRule) validateMaxDepth(headings []*parser.Heading, maxDepth int) []Violation {
	violations := make([]Violation, 0)

	for _, h := range headings {
		if h.Level > maxDepth {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Message: fmt.Sprintf("Heading '%s' (h%d) exceeds maximum depth of %d", h.Text, h.Level, maxDepth),
				Line:    h.Line,
				Column:  h.Column,
			})
		}
	}

	return violations
}

// GenerateContent does not generate content for heading rules
func (r *HeadingRule) GenerateContent(builder *strings.Builder, element schema.StructureElement) bool {
	return false
}
