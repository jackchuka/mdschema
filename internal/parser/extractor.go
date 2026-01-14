package parser

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
)

func extractHeading(node *ast.Heading, content []byte) *Heading {
	// Pre-allocate slice with estimated capacity to avoid reallocations
	textParts := make([][]byte, 0, 4)
	totalLen := 0

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if t, ok := child.(*ast.Text); ok {
			segment := t.Segment.Value(content)
			textParts = append(textParts, segment)
			totalLen += len(segment)
		}
	}

	// Build text efficiently using single allocation
	var text string
	if len(textParts) == 1 {
		// Common case: single text node
		text = string(textParts[0])
	} else if len(textParts) > 1 {
		// Multiple text nodes: join efficiently
		result := make([]byte, 0, totalLen)
		for _, part := range textParts {
			result = append(result, part...)
		}
		text = string(result)
	}

	text = strings.TrimSpace(text)
	line, col := getPosition(node, content)

	return &Heading{
		Level:  node.Level,
		Text:   text,
		Line:   line,
		Column: col,
		Slug:   generateSlug(text),
	}
}

func extractCodeBlock(node *ast.FencedCodeBlock, content []byte, parent *Heading) *CodeBlock {
	var lang string
	if node.Info != nil {
		lang = string(node.Info.Segment.Value(content))
	}

	line, col := getPosition(node, content)
	return &CodeBlock{
		Lang:   lang,
		Line:   line,
		Column: col,
		Parent: parent,
	}
}

func extractLink(node *ast.Link, content []byte) *Link {
	// Pre-allocate for text extraction
	textParts := make([][]byte, 0, 2)
	totalLen := 0

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if t, ok := child.(*ast.Text); ok {
			segment := t.Segment.Value(content)
			textParts = append(textParts, segment)
			totalLen += len(segment)
		}
	}

	// Build text efficiently
	var text string
	if len(textParts) == 1 {
		text = string(textParts[0])
	} else if len(textParts) > 1 {
		result := make([]byte, 0, totalLen)
		for _, part := range textParts {
			result = append(result, part...)
		}
		text = string(result)
	}

	url := string(node.Destination)
	line, col := getPosition(node, content)

	return &Link{
		URL:        url,
		Text:       text,
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

func extractTable(node *east.Table, content []byte, parent *Heading) *Table {
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
		Parent:  parent,
	}
}

func extractImage(node *ast.Image, content []byte) *Image {
	var altBuf bytes.Buffer
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if t, ok := child.(*ast.Text); ok {
			altBuf.Write(t.Segment.Value(content))
		}
	}

	line, col := getPosition(node, content)
	return &Image{
		URL:    string(node.Destination),
		Alt:    altBuf.String(),
		Line:   line,
		Column: col,
	}
}
