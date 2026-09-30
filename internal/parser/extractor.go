package parser

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
)

func extractHeading(node *ast.Heading, content []byte) *Heading {
	// Use ast.Walk to recursively extract all text (handles emphasis, code, links, etc.)
	var textBuf bytes.Buffer
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if t, ok := n.(*ast.Text); ok {
			textBuf.Write(t.Segment.Value(content))
		}
		return ast.WalkContinue, nil
	})

	text := strings.TrimSpace(textBuf.String())
	line, col := getPosition(node, content)

	return &Heading{
		Level:  node.Level,
		Text:   text,
		Line:   line,
		Column: col,
		Slug:   GenerateSlug(text),
	}
}

func extractCodeBlock(node *ast.FencedCodeBlock, content []byte) *CodeBlock {
	var lang string
	if node.Info != nil {
		lang = string(node.Info.Segment.Value(content))
	}

	line, col := getPosition(node, content)
	return &CodeBlock{
		Lang:   lang,
		Line:   line,
		Column: col,
	}
}

// extractFrontmatterLinks extracts link-like values from frontmatter data.
func extractFrontmatterLinks(data map[string]any) []*Link {
	if data == nil {
		return nil
	}

	var links []*Link
	for _, value := range data {
		str, ok := value.(string)
		if !ok {
			continue
		}
		if strings.HasPrefix(str, "#") || strings.HasPrefix(str, "http://") || strings.HasPrefix(str, "https://") {
			links = append(links, &Link{
				URL:        str,
				IsInternal: isInternalLink(str),
				Line:       1,
				Column:     1,
			})
		}
	}
	return links
}

func extractLink(node *ast.Link, content []byte) *Link {
	// Use ast.Walk to recursively extract all text (handles emphasis, code, etc.)
	var textBuf bytes.Buffer
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if t, ok := n.(*ast.Text); ok {
			textBuf.Write(t.Segment.Value(content))
		}
		return ast.WalkContinue, nil
	})

	url := string(node.Destination)
	line, col := getPosition(node, content)

	return &Link{
		URL:        url,
		Text:       textBuf.String(),
		IsInternal: isInternalLink(url),
		Line:       line,
		Column:     col,
	}
}

func extractList(node *ast.List, content []byte) *List {
	line, col := getPosition(node, content)
	list := &List{
		IsOrdered: node.IsOrdered(),
		Nested:    node.Parent() != nil && node.Parent().Kind() == ast.KindListItem,
		Items:     make([]*ListItem, 0),
		Line:      line,
		Column:    col,
	}

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if item, ok := child.(*ast.ListItem); ok {
			list.Items = append(list.Items, extractListItem(item, content, line, col))
		}
	}

	return list
}

// extractListItem reads the raw source of the item's first text block.
// Soft line breaks are joined with a single space.
func extractListItem(node *ast.ListItem, content []byte, fallbackLine, fallbackCol int) *ListItem {
	item := &ListItem{Line: fallbackLine, Column: fallbackCol}

	block := node.FirstChild()
	if block == nil {
		return item
	}
	switch block.Kind() {
	case ast.KindTextBlock, ast.KindParagraph:
	default:
		return item
	}

	lines := block.Lines()
	if lines.Len() == 0 {
		return item
	}
	parts := make([]string, 0, lines.Len())
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		parts = append(parts, strings.TrimSpace(string(seg.Value(content))))
	}
	item.Text = strings.Join(parts, " ")
	item.Line, item.Column = calculateLineColumn(content, lines.At(0).Start)
	return item
}

func extractParagraph(node *ast.Paragraph, content []byte) *Paragraph {
	line, col := getPosition(node, content)
	return &Paragraph{
		Line:   line,
		Column: col,
	}
}

func extractTable(node *east.Table, content []byte) *Table {
	headers := make([]string, 0)

	// Extract headers from first row
	if node.FirstChild() != nil && node.FirstChild().Kind() == east.KindTableHeader {
		headerRow := node.FirstChild()
		for cell := headerRow.FirstChild(); cell != nil; cell = cell.NextSibling() {
			var textBuf bytes.Buffer
			if err := ast.Walk(cell, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
				if !entering {
					return ast.WalkContinue, nil
				}
				if t, ok := n.(*ast.Text); ok {
					textBuf.Write(t.Segment.Value(content))
				}
				return ast.WalkContinue, nil
			}); err != nil {
				fmt.Printf("Error extracting table header: %v\n", err)
				continue
			}
			headers = append(headers, strings.TrimSpace(textBuf.String()))
		}
	}

	line, col := getPosition(node, content)
	return &Table{
		Headers: headers,
		Line:    line,
		Column:  col,
	}
}

func extractImage(node *ast.Image, content []byte) *Image {
	// Use ast.Walk to recursively extract all alt text (handles emphasis, code, etc.)
	var altBuf bytes.Buffer
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if t, ok := n.(*ast.Text); ok {
			altBuf.Write(t.Segment.Value(content))
		}
		return ast.WalkContinue, nil
	})

	line, col := getPosition(node, content)
	return &Image{
		URL:    string(node.Destination),
		Alt:    altBuf.String(),
		Line:   line,
		Column: col,
	}
}
