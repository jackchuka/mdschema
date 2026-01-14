package parser

import (
	"fmt"
	"strings"

	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
)

func getPosition(node ast.Node, content []byte) (line, col int) {
	// Try to get position from different node types
	switch n := node.(type) {
	case *ast.Heading:
		if n.Lines().Len() > 0 {
			return calculateLineColumn(content, n.Lines().At(0).Start)
		}
	case *ast.FencedCodeBlock:
		if n.Lines().Len() > 0 {
			return calculateLineColumn(content, n.Lines().At(0).Start)
		}
	case *ast.Text:
		return calculateLineColumn(content, n.Segment.Start)
	case *ast.List:
		if n.Lines().Len() > 0 {
			return calculateLineColumn(content, n.Lines().At(0).Start)
		}
	case *east.Table:
		if n.Lines().Len() > 0 {
			return calculateLineColumn(content, n.Lines().At(0).Start)
		}
	}

	// Fallback to finding the first text node
	var firstOffset int
	if err := ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if t, ok := n.(*ast.Text); ok && firstOffset == 0 {
			firstOffset = t.Segment.Start
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	}); err != nil {
		fmt.Printf("Error getting position: %v\n", err)
		return 1, 1 // Default to line 1, column 1 on error
	}

	if firstOffset > 0 {
		return calculateLineColumn(content, firstOffset)
	}

	// Default fallback
	return 1, 1
}

func calculateLineColumn(content []byte, offset int) (line, col int) {
	line = 1
	col = 1

	for i := 0; i < offset && i < len(content); i++ {
		if content[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}

	return line, col
}

func generateSlug(text string) string {
	// Simple slug generation
	slug := strings.ToLower(text)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.Trim(slug, "-")
	return slug
}

func isInternalLink(url string) bool {
	return !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://")
}
