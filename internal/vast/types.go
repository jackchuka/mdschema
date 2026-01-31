package vast

import (
	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/schema"
)

// Node represents a single schema-bound section in the validation AST.
// Each Node is a 1:1 binding between a schema element and a document section.
type Node struct {
	// Schema binding
	Element schema.StructureElement

	// Document binding (may be nil if required element is missing)
	Section *parser.Section

	// Semantic parent-child relationships (schema-driven, not heading-level-driven)
	Parent   *Node
	Children []*Node

	// Pre-computed validation state
	IsBound bool // True if a matching section was found

	// Ordering metadata
	Order       int // Expected order (position in schema)
	ActualOrder int // Actual order in document (line number for ordering violations)

	// Count tracking for multi-match elements
	MatchCount int // Total matches found for this element (0 if not multi-match)
	MatchIndex int // This node's index (0-based) among siblings with same element
}

// Content returns the section content if bound, empty string otherwise.
func (n *Node) Content() string {
	if n.Section != nil {
		return n.Section.Content
	}
	return ""
}

// CodeBlocks returns the code blocks if bound, empty slice otherwise.
func (n *Node) CodeBlocks() []*parser.CodeBlock {
	if n.Section != nil {
		return n.Section.CodeBlocks
	}
	return nil
}

// Tables returns the tables if bound, empty slice otherwise.
func (n *Node) Tables() []*parser.Table {
	if n.Section != nil {
		return n.Section.Tables
	}
	return nil
}

// Links returns the links if bound, empty slice otherwise.
func (n *Node) Links() []*parser.Link {
	if n.Section != nil {
		return n.Section.Links
	}
	return nil
}

// Images returns the images if bound, empty slice otherwise.
func (n *Node) Images() []*parser.Image {
	if n.Section != nil {
		return n.Section.Images
	}
	return nil
}

// Lists returns the lists if bound, empty slice otherwise.
func (n *Node) Lists() []*parser.List {
	if n.Section != nil {
		return n.Section.Lists
	}
	return nil
}

// Location returns line/column for error reporting.
func (n *Node) Location() (line, column int) {
	if n.Section != nil && n.Section.Heading != nil {
		return n.Section.Heading.Line, n.Section.Heading.Column
	}
	// Default to parent's location or document start
	if n.Parent != nil {
		return n.Parent.Location()
	}
	return 1, 1
}

// HeadingText returns the heading text if bound, otherwise the schema pattern.
func (n *Node) HeadingText() string {
	if n.Section != nil && n.Section.Heading != nil {
		return n.Section.Heading.Text
	}
	return n.Element.Heading.Pattern
}

// Tree represents the complete validation-ready AST.
type Tree struct {
	// Root nodes (top-level schema elements)
	Roots []*Node

	// Document reference for context
	Document *parser.Document

	// All nodes flattened for iteration
	AllNodes []*Node

	// Unmatched sections (sections in document but not in schema)
	UnmatchedSections []*parser.Section
}

// Walk traverses all nodes in depth-first order.
// Returns early if fn returns false.
func (t *Tree) Walk(fn func(*Node) bool) {
	for _, root := range t.Roots {
		if !walkNode(root, fn) {
			return
		}
	}
}

func walkNode(n *Node, fn func(*Node) bool) bool {
	if !fn(n) {
		return false
	}
	for _, child := range n.Children {
		if !walkNode(child, fn) {
			return false
		}
	}
	return true
}

// WalkBound traverses only bound nodes (nodes with matching sections).
func (t *Tree) WalkBound(fn func(*Node) bool) {
	t.Walk(func(n *Node) bool {
		if n.IsBound {
			return fn(n)
		}
		return true
	})
}

// GetByElement finds all nodes matching a specific schema element heading pattern.
func (t *Tree) GetByElement(heading string) []*Node {
	var nodes []*Node
	t.Walk(func(n *Node) bool {
		if n.Element.Heading.Pattern == heading {
			nodes = append(nodes, n)
		}
		return true
	})
	return nodes
}
