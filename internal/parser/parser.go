package parser

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// Document represents a parsed Markdown document with hierarchical structure
type Document struct {
	Path    string
	Content []byte
	AST     ast.Node
	Root    *Section // Root section containing the entire document tree
}

// GetSections returns all sections in document order
func (d *Document) GetSections() []*Section {
	var sections []*Section
	d.Root.collectSections(&sections)
	return sections
}

// Heading represents a heading in the document
type Heading struct {
	Level  int
	Text   string
	Line   int
	Column int
	Slug   string
}

// Section represents a section of content under a heading
type Section struct {
	Children  []*Section // Nested subsections
	Parent    *Section   // Parent section (nil for root)
	Content   string
	StartLine int
	EndLine   int

	Heading    *Heading
	CodeBlocks []*CodeBlock
	Tables     []*Table
	Links      []*Link
	Images     []*Image
}

// collectSections recursively collects all sections in document order
func (s *Section) collectSections(sections *[]*Section) {
	if s.Heading != nil { // Don't include root section
		*sections = append(*sections, s)
	}
	for _, child := range s.Children {
		child.collectSections(sections)
	}
}

// CodeBlock represents a code block
type CodeBlock struct {
	Lang   string
	Line   int
	Column int
	Parent *Heading // The heading this block appears under
}

// Link represents a link in the document
type Link struct {
	URL        string
	Text       string
	IsInternal bool
	Line       int
	Column     int
}

// List represents a list in the document
type List struct {
	IsOrdered bool
	Line      int
	Column    int
}

// Table represents a table in the document
type Table struct {
	Headers []string
	Line    int
	Column  int
	Parent  *Heading
}

// Image represents an image in the document
type Image struct {
	URL    string
	Alt    string
	Line   int
	Column int
}

// FrontMatter represents document front matter
type FrontMatter struct {
	Format  string // "yaml" or "toml"
	Content string
	Data    map[string]any
}

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

// buildHierarchicalSections creates a hierarchical tree of sections
func (p *Parser) buildHierarchicalSections(headings []*Heading, codeBlocks []*CodeBlock, tables []*Table, links []*Link, images []*Image, content []byte) *Section {
	// Create root section for top-level content
	root := &Section{
		Heading:    nil, // No heading for root
		StartLine:  1,
		EndLine:    len(strings.Split(string(content), "\n")),
		Children:   make([]*Section, 0),
		Parent:     nil,
		CodeBlocks: make([]*CodeBlock, 0),
		Tables:     make([]*Table, 0),
		Links:      make([]*Link, 0),
		Images:     make([]*Image, 0),
	}

	// Stack to track current nesting level
	sectionStack := []*Section{root}

	// Process each heading in order
	for _, heading := range headings {
		// Create section for this heading
		section := &Section{
			Heading:    heading,
			StartLine:  heading.Line,
			Children:   make([]*Section, 0),
			CodeBlocks: make([]*CodeBlock, 0),
			Tables:     make([]*Table, 0),
			Links:      make([]*Link, 0),
			Images:     make([]*Image, 0),
		}

		// Find appropriate parent based on heading level
		// Pop stack until we find a section with lower level (higher in hierarchy)
		for len(sectionStack) > 1 {
			parent := sectionStack[len(sectionStack)-1]
			if parent.Heading == nil || parent.Heading.Level < heading.Level {
				break
			}
			sectionStack = sectionStack[:len(sectionStack)-1]
		}

		// Set parent-child relationship
		parent := sectionStack[len(sectionStack)-1]
		section.Parent = parent
		parent.Children = append(parent.Children, section)

		// Add this section to the stack
		sectionStack = append(sectionStack, section)
	}

	// Second pass: set end lines and associate content
	p.setEndLinesAndContent(root, codeBlocks, tables, links, images, content)

	return root
}

// setEndLinesAndContent recursively sets end lines and associates content with sections
func (p *Parser) setEndLinesAndContent(section *Section, codeBlocks []*CodeBlock, tables []*Table, links []*Link, images []*Image, content []byte) {
	contentLines := strings.Split(string(content), "\n")

	// Determine end line
	if len(section.Children) > 0 {
		// Will be updated after processing children
		section.EndLine = len(contentLines)
	} else if section.Parent != nil {
		// If no children, find next sibling or parent's end
		nextSiblingStart := 0
		parentChildren := section.Parent.Children
		for i, child := range parentChildren {
			if child == section && i+1 < len(parentChildren) {
				nextSiblingStart = parentChildren[i+1].StartLine
				break
			}
		}
		if nextSiblingStart > 0 {
			section.EndLine = nextSiblingStart - 1
		} else {
			section.EndLine = section.Parent.EndLine
		}
	}

	// Extract content
	if section.StartLine > 0 && section.EndLine <= len(contentLines) {
		sectionContent := make([]string, 0)
		startIdx := section.StartLine
		if section.Heading != nil {
			startIdx = section.StartLine + 1 // Skip heading line
		}
		for lineIdx := startIdx; lineIdx <= section.EndLine && lineIdx <= len(contentLines); lineIdx++ {
			if lineIdx-1 < len(contentLines) {
				sectionContent = append(sectionContent, contentLines[lineIdx-1])
			}
		}
		section.Content = strings.Join(sectionContent, "\n")
	}

	// Associate elements with this section
	if section.Heading != nil {
		for _, codeBlock := range codeBlocks {
			if codeBlock.Line >= section.StartLine && codeBlock.Line <= section.EndLine {
				// Check if it belongs to a child section
				belongsToChild := false
				for _, child := range section.Children {
					if codeBlock.Line >= child.StartLine && codeBlock.Line <= child.EndLine {
						belongsToChild = true
						break
					}
				}
				if !belongsToChild {
					section.CodeBlocks = append(section.CodeBlocks, codeBlock)
				}
			}
		}

		for _, table := range tables {
			if table.Line >= section.StartLine && table.Line <= section.EndLine {
				// Check if it belongs to a child section
				belongsToChild := false
				for _, child := range section.Children {
					if table.Line >= child.StartLine && table.Line <= child.EndLine {
						belongsToChild = true
						break
					}
				}
				if !belongsToChild {
					section.Tables = append(section.Tables, table)
				}
			}
		}
	}

	// Recursively process children
	for _, child := range section.Children {
		p.setEndLinesAndContent(child, codeBlocks, tables, links, images, content)
	}

	// After processing children, update end line to include all children content
	// but not exceed the next sibling's start
	if len(section.Children) > 0 {
		lastChild := section.Children[len(section.Children)-1]

		// Find next sibling to limit end line
		nextSiblingStart := 0
		if section.Parent != nil {
			parentChildren := section.Parent.Children
			for i, child := range parentChildren {
				if child == section && i+1 < len(parentChildren) {
					nextSiblingStart = parentChildren[i+1].StartLine
					break
				}
			}
		}

		if nextSiblingStart > 0 {
			section.EndLine = nextSiblingStart - 1
		} else {
			section.EndLine = lastChild.EndLine
		}
	}
}
