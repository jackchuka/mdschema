package rules

import (
	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/schema"
)

// SectionMapping represents a mapping between a schema element and document sections
type SectionMapping struct {
	Element  schema.StructureElement
	Sections []*parser.Section // All document sections that match this element
	Parent   *SectionMapping   // Parent mapping in the hierarchy
	Children []*SectionMapping // Child mappings
}

// ValidationContext provides clean access to section-schema mappings for rules
type ValidationContext struct {
	Document *parser.Document
	Schema   *schema.Schema
	Mappings []*SectionMapping // Root-level mappings
}

// NewValidationContext creates a context with pre-established section-to-schema mappings
func NewValidationContext(doc *parser.Document, s *schema.Schema) *ValidationContext {
	ctx := &ValidationContext{
		Document: doc,
		Schema:   s,
	}

	// Build mappings starting from root
	ctx.Mappings = buildMappings(doc.Root, s.Structure, nil)

	return ctx
}

// buildMappings recursively builds section-to-schema mappings
func buildMappings(section *parser.Section, expectedElements []schema.StructureElement, parent *SectionMapping) []*SectionMapping {
	mappings := make([]*SectionMapping, 0)
	matcher := NewPatternMatcher()

	for _, element := range expectedElements {
		mapping := &SectionMapping{
			Element:  element,
			Sections: make([]*parser.Section, 0),
			Parent:   parent,
			Children: make([]*SectionMapping, 0),
		}

		isRegex := containsRegexMetachars(element.Heading)

		// Find all matching sections for this element
		for _, child := range section.Children {
			if child.Heading != nil && matcher.MatchesHeadingPattern(child.Heading, element.Heading, isRegex) {
				mapping.Sections = append(mapping.Sections, child)
			}
		}

		// Build child mappings only for the first matching section to avoid duplicates
		if len(mapping.Sections) > 0 && len(element.Children) > 0 {
			childMappings := buildMappings(mapping.Sections[0], element.Children, mapping)
			mapping.Children = append(mapping.Children, childMappings...)
		}

		mappings = append(mappings, mapping)
	}

	return mappings
}

// containsRegexMetachars checks if a pattern contains regex metacharacters
func containsRegexMetachars(pattern string) bool {
	for _, char := range pattern {
		switch char {
		case '[', ']', '*', '+', '?', '^', '$', '\\':
			return true
		}
	}
	return false
}
