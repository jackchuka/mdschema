package generator

import (
	"strings"

	"github.com/jackchuka/mdschema/internal/rules"
	"github.com/jackchuka/mdschema/internal/schema"
	"github.com/jackchuka/mdschema/internal/vast"
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

// Generate creates a markdown template from the schema structure. outputPath (pass "" if
// unknown) resolves an expr-based heading to its own filename instead of the expression text.
func (g *Generator) Generate(s *schema.Schema, outputPath string) string {
	var builder strings.Builder

	// Generate frontmatter if applicable
	g.ruleGenerator.GenerateFrontmatter(&builder, s)

	for _, element := range s.Structure {
		g.generateElement(&builder, element, 1, outputPath)
	}

	return builder.String()
}

// generateElement recursively generates markdown for a structure element
func (g *Generator) generateElement(builder *strings.Builder, element schema.StructureElement, level int, outputPath string) {
	// Generate heading - extract text from schema pattern
	headingText := g.resolveHeadingText(element.Heading, outputPath, level)
	heading := strings.Repeat("#", level) + " " + headingText
	builder.WriteString(heading + "\n\n")

	// Add description as HTML comment if present
	if element.Description != "" {
		builder.WriteString("<!-- " + element.Description + " -->\n\n")
	}

	// Add optional marker if applicable
	if element.Optional {
		builder.WriteString("<!-- Optional section -->\n\n")
	}

	// Use rule-based content generation
	g.ruleGenerator.GenerateContent(builder, element)

	// Generate children elements
	for _, child := range element.Children {
		g.generateElement(builder, child, level+1, outputPath)
	}
}

// resolveHeadingText tries the filename as the heading for an expr-based pattern, keeping it
// only if EvaluateHeadingExpr confirms the expression holds for heading == filename.
func (g *Generator) resolveHeadingText(hp schema.HeadingPattern, outputPath string, level int) string {
	if hp.Pattern == "" && hp.Literal == "" && hp.Expr != "" && outputPath != "" {
		filename := vast.ExtractFilename(outputPath)
		if matched, err := vast.EvaluateHeadingExpr(hp.Expr, filename, filename, level); err == nil && matched {
			return filename
		}
	}
	return g.extractHeadingText(hp.GetReadableName())
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
