package parser

import (
	"fmt"
	"os"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// Parser handles Markdown parsing
type Parser struct {
	md goldmark.Markdown
}

// New creates a new parser instance
func New() *Parser {
	return &Parser{
		md: goldmark.New(
			goldmark.WithExtensions(extension.Table),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
			),
		),
	}
}

// ParseFile parses a Markdown file
func (p *Parser) ParseFile(path string) (*Document, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	return p.Parse(path, content)
}

// Parse parses Markdown content
func (p *Parser) Parse(path string, content []byte) (*Document, error) {
	reader := text.NewReader(content)
	node := p.md.Parser().Parse(reader)

	// Temporary collections for building the tree
	headings := make([]*Heading, 0)
	codeBlocks := make([]*CodeBlock, 0)
	links := make([]*Link, 0)
	lists := make([]*List, 0)
	tables := make([]*Table, 0)
	images := make([]*Image, 0)

	// Walk the AST and extract elements
	var currentHeading *Heading
	if err := ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch node := n.(type) {
		case *ast.Heading:
			heading := extractHeading(node, content)
			headings = append(headings, heading)
			currentHeading = heading

		case *ast.FencedCodeBlock:
			block := extractCodeBlock(node, content, currentHeading)
			codeBlocks = append(codeBlocks, block)

		case *ast.Link:
			link := extractLink(node, content)
			links = append(links, link)

		case *ast.List:
			list := extractList(node, content)
			lists = append(lists, list)

		case *east.Table:
			table := extractTable(node, content, currentHeading)
			tables = append(tables, table)

		case *ast.Image:
			image := extractImage(node, content)
			images = append(images, image)
		}

		return ast.WalkContinue, nil
	}); err != nil {
		return nil, fmt.Errorf("walking AST: %w", err)
	}

	// Build hierarchical structure
	root := p.buildHierarchicalSections(headings, codeBlocks, tables, links, images, content)

	// Create the document
	doc := &Document{
		Path:    path,
		Content: content,
		AST:     node,
		Root:    root,
	}

	return doc, nil
}
