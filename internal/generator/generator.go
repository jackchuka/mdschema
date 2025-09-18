package generator

import (
	"strings"

	"github.com/jackchuka/mdschema/internal/rules"
	"github.com/jackchuka/mdschema/internal/schema"
)

// Generator creates markdown templates from schemas using rules
type Generator struct {
	ruleGenerator *rules.Generator
}

// New creates a new Generator
func New() *Generator {
	return &Generator{
		ruleGenerator: rules.NewGenerator(),
	}
}

// Generate creates a markdown template from the schema structure
func (g *Generator) Generate(s *schema.Schema) string {
	var builder strings.Builder

	builder.WriteString("<!-- Generated from schema -->\n\n")

	for _, element := range s.Structure {
		g.generateElement(&builder, element, 1)
	}

	return builder.String()
}

// generateElement recursively generates markdown for a structure element
func (g *Generator) generateElement(builder *strings.Builder, element schema.StructureElement, level int) {
	// Generate heading - extract text from schema pattern
	headingText := g.extractHeadingText(element.Heading)
	heading := strings.Repeat("#", level) + " " + headingText
	builder.WriteString(heading + "\n\n")

	// Add optional marker if applicable
	if element.Optional {
		builder.WriteString("<!-- Optional section -->\n\n")
	}

	// Use rule-based content generation
	g.ruleGenerator.GenerateContent(builder, element)

	// Generate children elements
	for _, child := range element.Children {
		g.generateElement(builder, child, level+1)
	}
}

// extractHeadingText extracts human-readable text from a heading pattern
func (g *Generator) extractHeadingText(pattern string) string {
	// Remove heading prefix (# ## ###)
	text := strings.TrimSpace(pattern)

	// Remove markdown heading prefix
	for strings.HasPrefix(text, "#") {
		text = strings.TrimSpace(text[1:])
	}

	return text
}
