package parser

import (
	"fmt"
	"os"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
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
			goldmark.WithExtensions(
				extension.Table,
				meta.Meta,
			),
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
	// Create a parser context to capture metadata
	ctx := parser.NewContext()
	reader := text.NewReader(content)
	node := p.md.Parser().Parse(reader, parser.WithContext(ctx))

	// Extract frontmatter from goldmark-meta
	var frontMatter *FrontMatter
	if metaData := meta.Get(ctx); metaData != nil {
		frontMatter = &FrontMatter{
			Format: "yaml",
			Data:   metaData,
		}
	}

	// Temporary collections for building the tree
	headings := make([]*Heading, 0)
	codeBlocks := make([]*CodeBlock, 0)
	links := make([]*Link, 0)
	lists := make([]*List, 0)
	tables := make([]*Table, 0)
	images := make([]*Image, 0)

	// Walk the AST and extract elements
	if err := ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch node := n.(type) {
		case *ast.Heading:
			heading := extractHeading(node, content)
			headings = append(headings, heading)

		case *ast.FencedCodeBlock:
			block := extractCodeBlock(node, content)
			codeBlocks = append(codeBlocks, block)

		case *ast.Link:
			link := extractLink(node, content)
			links = append(links, link)

		case *ast.List:
			list := extractList(node, content)
			lists = append(lists, list)

		case *east.Table:
			table := extractTable(node, content)
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
	root := p.buildHierarchicalSections(headings, codeBlocks, tables, links, images, lists, content)

	// Create the document
	doc := &Document{
		Path:        path,
		Content:     content,
		AST:         node,
		Root:        root,
		FrontMatter: frontMatter,
	}

	return doc, nil
}
