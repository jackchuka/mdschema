package infer

import (
	"fmt"
	"strings"

	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/schema"
)

// FromDocument converts a parsed Markdown document into a schema definition.
func FromDocument(doc *parser.Document) (*schema.Schema, error) {
	if doc == nil {
		return nil, fmt.Errorf("document is nil")
	}

	if doc.Root == nil {
		return nil, fmt.Errorf("document has no structural information")
	}

	if len(doc.Root.Children) == 0 {
		return nil, fmt.Errorf("document has no headings to infer structure")
	}

	// Note: Frontmatter is now stripped by the parser before parsing,
	// so we don't need to skip sections based on frontmatter line numbers.
	// The parser extracts frontmatter into doc.FrontMatter.

	structure := make([]schema.StructureElement, 0, len(doc.Root.Children))
	for _, section := range doc.Root.Children {
		structure = append(structure, buildElement(section))
	}

	if len(structure) == 0 {
		return nil, fmt.Errorf("document has no headings to infer structure")
	}

	return &schema.Schema{Structure: structure}, nil
}

func buildElement(section *parser.Section) schema.StructureElement {
	var heading schema.HeadingPattern
	if section.Heading != nil {
		heading = schema.HeadingPattern{Pattern: headingPattern(section.Heading)}
	}

	var children []schema.StructureElement
	if len(section.Children) > 0 {
		children = make([]schema.StructureElement, 0, len(section.Children))
		for _, child := range section.Children {
			children = append(children, buildElement(child))
		}
	}

	return schema.StructureElement{
		Heading:  heading,
		Children: children,
	}
}

func headingPattern(h *parser.Heading) string {
	if h == nil {
		return ""
	}

	prefix := strings.Repeat("#", h.Level)
	text := strings.TrimSpace(h.Text)
	if text == "" {
		return prefix
	}
	return fmt.Sprintf("%s %s", prefix, text)
}
