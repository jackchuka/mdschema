package parser

import "github.com/yuin/goldmark/ast"

// Document represents a parsed Markdown document with hierarchical structure
type Document struct {
	Path        string
	Content     []byte
	AST         ast.Node
	Root        *Section     // Root section containing the entire document tree
	FrontMatter *FrontMatter // Optional front matter at document start
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
	Lists      []*List
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

// LineLocatable is implemented by elements that have a line position
type LineLocatable interface {
	GetLine() int
}

// GetLine returns the line number of the code block
func (c *CodeBlock) GetLine() int { return c.Line }

// GetLine returns the line number of the link
func (l *Link) GetLine() int { return l.Line }

// GetLine returns the line number of the list
func (l *List) GetLine() int { return l.Line }

// GetLine returns the line number of the table
func (t *Table) GetLine() int { return t.Line }

// GetLine returns the line number of the image
func (i *Image) GetLine() int { return i.Line }
