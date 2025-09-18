package infer

import (
	"bufio"
	"bytes"
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

	frontMatterEnd, hasFrontMatter := detectFrontMatter(doc.Content)

	structure := make([]schema.StructureElement, 0, len(doc.Root.Children))
	for _, section := range doc.Root.Children {
		if hasFrontMatter && section.StartLine <= frontMatterEnd {
			continue
		}
		structure = append(structure, buildElement(section))
	}

	if len(structure) == 0 {
		return nil, fmt.Errorf("document has no headings to infer structure")
	}

	return &schema.Schema{Structure: structure}, nil
}

func buildElement(section *parser.Section) schema.StructureElement {
	heading := ""
	if section.Heading != nil {
		heading = headingPattern(section.Heading)
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

func detectFrontMatter(content []byte) (endLine int, ok bool) {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	lineNum := 0

	if !scanner.Scan() {
		return 0, false
	}
	lineNum++
	line := scanner.Text()
	line = strings.TrimPrefix(line, "\ufeff")
	if strings.TrimSpace(line) != "---" {
		return 0, false
	}

	for scanner.Scan() {
		lineNum++
		trimmed := strings.TrimSpace(scanner.Text())
		if trimmed == "---" || trimmed == "..." {
			return lineNum, true
		}
	}

	return 0, false
}
