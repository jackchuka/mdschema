package parser

import (
	"bytes"
	"fmt"
	"os"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"gopkg.in/yaml.v3"
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
	// Extract frontmatter first (before goldmark parsing)
	frontMatter, markdownContent := p.splitFrontMatter(content)

	reader := text.NewReader(markdownContent)
	node := p.md.Parser().Parse(reader)

	// Temporary collections for building the tree
	headings := make([]*Heading, 0)
	codeBlocks := make([]*CodeBlock, 0)
	links := make([]*Link, 0)
	lists := make([]*List, 0)
	tables := make([]*Table, 0)
	images := make([]*Image, 0)

	// Walk the AST and extract elements (use markdownContent since that's what goldmark parsed)
	if err := ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch node := n.(type) {
		case *ast.Heading:
			heading := extractHeading(node, markdownContent)
			headings = append(headings, heading)

		case *ast.FencedCodeBlock:
			block := extractCodeBlock(node, markdownContent)
			codeBlocks = append(codeBlocks, block)

		case *ast.Link:
			link := extractLink(node, markdownContent)
			links = append(links, link)

		case *ast.List:
			list := extractList(node, markdownContent)
			lists = append(lists, list)

		case *east.Table:
			table := extractTable(node, markdownContent)
			tables = append(tables, table)

		case *ast.Image:
			image := extractImage(node, markdownContent)
			images = append(images, image)
		}

		return ast.WalkContinue, nil
	}); err != nil {
		return nil, fmt.Errorf("walking AST: %w", err)
	}

	// Build hierarchical structure (use markdownContent for section content)
	root := p.buildHierarchicalSections(headings, codeBlocks, tables, links, images, lists, markdownContent)

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

// splitFrontMatter extracts YAML front matter and returns it with remaining content
func (p *Parser) splitFrontMatter(content []byte) (*FrontMatter, []byte) {
	// Check if content starts with "---"
	if !bytes.HasPrefix(content, []byte("---")) {
		return nil, content
	}

	// Find the closing "---"
	rest := content[3:]
	endIndex := bytes.Index(rest, []byte("\n---"))
	if endIndex == -1 {
		return nil, content
	}

	// Extract the YAML content (skip leading newline if present)
	yamlContent := rest[:endIndex]
	if len(yamlContent) > 0 && yamlContent[0] == '\n' {
		yamlContent = yamlContent[1:]
	}

	// Calculate where markdown content starts (after closing --- and newline)
	markdownStart := 3 + endIndex + 4 // "---" + content + "\n---"
	if markdownStart < len(content) && content[markdownStart] == '\n' {
		markdownStart++
	}
	markdownContent := content[markdownStart:]

	// Parse the YAML
	var data map[string]any
	if err := yaml.Unmarshal(yamlContent, &data); err != nil {
		// Return frontmatter with raw content even if YAML parsing fails
		return &FrontMatter{
			Format:  "yaml",
			Content: string(yamlContent),
			Data:    nil,
		}, markdownContent
	}

	return &FrontMatter{
		Format:  "yaml",
		Content: string(yamlContent),
		Data:    data,
	}, markdownContent
}
