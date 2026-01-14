package vast

import (
	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/schema"
)

// Builder transforms parser output + schema into a validation-ready AST.
type Builder struct {
	matcher *PatternMatcher
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

	// Track which sections have been bound at each level
	boundSections := make(map[*parser.Section]bool)

	// Build nodes for each top-level schema element
	for i, element := range s.Structure {
		// Find all matching sections at root level (skip already-bound sections)
		matches := b.findMatches(doc.Root.Children, element, boundSections)

		// Create a node for EACH match (not just the first!)
		matchCount := 0
		for _, section := range matches {
			// Double-check that section wasn't bound by a previous match in this loop
			if boundSections[section] {
				continue
			}
			node := b.buildNode(element, section, nil, i, boundSections)
			tree.Roots = append(tree.Roots, node)
			tree.AllNodes = append(tree.AllNodes, node)
			b.collectAllNodes(node, &tree.AllNodes)
			matchCount++
		}

		// If no matches and element is required, create an unbound node
		if matchCount == 0 {
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

	// Build children for THIS specific section (not just the first match!)
	if section != nil && len(element.Children) > 0 {
		for i, childElement := range element.Children {
			// Find matches among THIS section's children (skip already-bound)
			childMatches := b.findMatches(section.Children, childElement, boundSections)

			for _, childSection := range childMatches {
				childNode := b.buildNode(childElement, childSection, node, i, boundSections)
				node.Children = append(node.Children, childNode)
			}

			// If no matches and child element exists in schema, create unbound node
			if len(childMatches) == 0 {
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

// findMatches finds all sections matching a schema element, skipping already-bound sections.
func (b *Builder) findMatches(sections []*parser.Section, element schema.StructureElement, boundSections map[*parser.Section]bool) []*parser.Section {
	matches := make([]*parser.Section, 0)

	for _, section := range sections {
		// Skip already-bound sections to avoid double-matching
		if boundSections[section] {
			continue
		}
		if section.Heading != nil && b.matcher.MatchesHeadingPattern(section.Heading, element.Heading.Pattern, element.Heading.Regex) {
			matches = append(matches, section)
		}
	}

	return matches
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
