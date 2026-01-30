package vast

import (
	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/schema"
)

// Builder transforms parser output + schema into a validation-ready AST.
type Builder struct {
	matcher  *PatternMatcher
	filename string // Document filename for expression matching
}

// NewBuilder creates a new VAST builder.
func NewBuilder() *Builder {
	return &Builder{
		matcher: NewPatternMatcher(),
	}
}

// Build creates a validation-ready AST from document and schema.
func (b *Builder) Build(doc *parser.Document, s *schema.Schema) *Tree {
	tree := &Tree{
		Document:          doc,
		Roots:             make([]*Node, 0),
		AllNodes:          make([]*Node, 0),
		UnmatchedSections: make([]*parser.Section, 0),
	}

	// Store document path for expression matching
	b.filename = doc.Path

	// Track which sections have been bound at each level
	boundSections := make(map[*parser.Section]bool)

	// Build nodes for each top-level schema element
	lastMatchedLine := 0
	for i, element := range s.Structure {
		// Find the first matching section at root level after the last match
		match := b.findFirstMatchAfter(doc.Root.Children, element, boundSections, lastMatchedLine)

		if match != nil {
			node := b.buildNode(element, match, nil, i, boundSections)
			tree.Roots = append(tree.Roots, node)
			tree.AllNodes = append(tree.AllNodes, node)
			b.collectAllNodes(node, &tree.AllNodes)
			lastMatchedLine = match.StartLine
		} else {
			// If no match, create an unbound node
			node := &Node{
				Element: element,
				Section: nil,
				Parent:  nil,
				IsBound: false,
				Order:   i,
			}
			tree.Roots = append(tree.Roots, node)
			tree.AllNodes = append(tree.AllNodes, node)
			// Also build unbound children for this unbound node
			b.buildUnboundChildren(node, element.Children, boundSections)
		}
	}

	// Collect unmatched sections
	b.collectUnmatched(doc.Root, boundSections, &tree.UnmatchedSections)

	return tree
}

// buildNode recursively builds a node and its children.
func (b *Builder) buildNode(element schema.StructureElement, section *parser.Section, parent *Node, order int, boundSections map[*parser.Section]bool) *Node {
	node := &Node{
		Element:  element,
		Section:  section,
		Parent:   parent,
		IsBound:  section != nil,
		Order:    order,
		Children: make([]*Node, 0),
	}

	if section != nil {
		boundSections[section] = true
		node.ActualOrder = section.StartLine
	}

	// Build children for THIS specific section (first match only)
	if section != nil && len(element.Children) > 0 {
		lastMatchedLine := 0
		for i, childElement := range element.Children {
			// Find the first match among THIS section's children after the last match
			childMatch := b.findFirstMatchAfter(section.Children, childElement, boundSections, lastMatchedLine)

			if childMatch != nil {
				childNode := b.buildNode(childElement, childMatch, node, i, boundSections)
				node.Children = append(node.Children, childNode)
				lastMatchedLine = childMatch.StartLine
			} else {
				// If no match, create an unbound node
				childNode := &Node{
					Element:  childElement,
					Section:  nil,
					Parent:   node,
					IsBound:  false,
					Order:    i,
					Children: make([]*Node, 0),
				}
				node.Children = append(node.Children, childNode)
				// Recursively build unbound children
				b.buildUnboundChildren(childNode, childElement.Children, boundSections)
			}
		}
	}

	return node
}

// buildUnboundChildren creates unbound child nodes for a schema element with no matching section.
func (b *Builder) buildUnboundChildren(parent *Node, children []schema.StructureElement, boundSections map[*parser.Section]bool) {
	for i, childElement := range children {
		childNode := &Node{
			Element:  childElement,
			Section:  nil,
			Parent:   parent,
			IsBound:  false,
			Order:    i,
			Children: make([]*Node, 0),
		}
		parent.Children = append(parent.Children, childNode)
		// Recursively build unbound grandchildren
		b.buildUnboundChildren(childNode, childElement.Children, boundSections)
	}
}

// findFirstMatchAfter finds the first matching section after minLine, skipping already-bound sections.
func (b *Builder) findFirstMatchAfter(sections []*parser.Section, element schema.StructureElement, boundSections map[*parser.Section]bool, minLine int) *parser.Section {
	for _, section := range sections {
		// Skip already-bound sections to avoid double-matching
		if boundSections[section] {
			continue
		}
		if section.StartLine <= minLine {
			continue
		}
		if section.Heading != nil && b.matcher.MatchesHeading(section.Heading, element.Heading, b.filename) {
			return section
		}
	}

	return nil
}

// collectAllNodes recursively collects all nodes from a node's children.
func (b *Builder) collectAllNodes(n *Node, nodes *[]*Node) {
	for _, child := range n.Children {
		*nodes = append(*nodes, child)
		b.collectAllNodes(child, nodes)
	}
}

// collectUnmatched finds sections not bound to any schema element.
func (b *Builder) collectUnmatched(section *parser.Section, bound map[*parser.Section]bool, unmatched *[]*parser.Section) {
	if section.Heading != nil && !bound[section] {
		*unmatched = append(*unmatched, section)
	}
	for _, child := range section.Children {
		b.collectUnmatched(child, bound, unmatched)
	}
}
