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
		Line:      line,
		Column:    col,
	}

	return list
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
