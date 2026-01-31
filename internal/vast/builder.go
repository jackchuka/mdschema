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
		if hasMultiMatch(element) {
			// Multi-match: find all matching sections
			maxMatches := getMaxMatches(element)
			matches := b.findAllMatchesAfter(doc.Root.Children, element, boundSections, lastMatchedLine, maxMatches)

			if len(matches) > 0 {
				for j, match := range matches {
					node := b.buildNode(element, match, nil, i, boundSections)
					node.MatchCount = len(matches)
					node.MatchIndex = j
					tree.Roots = append(tree.Roots, node)
					tree.AllNodes = append(tree.AllNodes, node)
					b.collectAllNodes(node, &tree.AllNodes)
					lastMatchedLine = match.StartLine
				}
			} else if getMinMatches(element) > 0 {
				// No matches found but element is required - create unbound node
				node := &Node{
					Element:    element,
					Section:    nil,
					Parent:     nil,
					IsBound:    false,
					Order:      i,
					MatchCount: 0,
				}
				tree.Roots = append(tree.Roots, node)
				tree.AllNodes = append(tree.AllNodes, node)
				b.buildUnboundChildren(node, element.Children, boundSections)
			}
		} else {
			// Single match: existing behavior
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

	// Build children for THIS specific section
	if section != nil && len(element.Children) > 0 {
		lastMatchedLine := 0
		for i, childElement := range element.Children {
			if hasMultiMatch(childElement) {
				// Multi-match child: find all matching sections
				maxMatches := getMaxMatches(childElement)
				childMatches := b.findAllMatchesAfter(section.Children, childElement, boundSections, lastMatchedLine, maxMatches)

				if len(childMatches) > 0 {
					for j, childMatch := range childMatches {
						childNode := b.buildNode(childElement, childMatch, node, i, boundSections)
						childNode.MatchCount = len(childMatches)
						childNode.MatchIndex = j
						node.Children = append(node.Children, childNode)
						lastMatchedLine = childMatch.StartLine
					}
				} else if getMinMatches(childElement) > 0 {
					// No matches but required - create unbound node
					childNode := &Node{
						Element:    childElement,
						Section:    nil,
						Parent:     node,
						IsBound:    false,
						Order:      i,
						Children:   make([]*Node, 0),
						MatchCount: 0,
					}
					node.Children = append(node.Children, childNode)
					b.buildUnboundChildren(childNode, childElement.Children, boundSections)
				}
			} else {
				// Single match child: existing behavior
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

// findAllMatchesAfter finds ALL matching sections after minLine, up to maxMatches (0 = unlimited).
func (b *Builder) findAllMatchesAfter(sections []*parser.Section, element schema.StructureElement, boundSections map[*parser.Section]bool, minLine int, maxMatches int) []*parser.Section {
	var matches []*parser.Section
	for _, section := range sections {
		// Skip already-bound sections to avoid double-matching
		if boundSections[section] {
			continue
		}
		if section.StartLine <= minLine {
			continue
		}
		if section.Heading != nil && b.matcher.MatchesHeading(section.Heading, element.Heading, b.filename) {
			matches = append(matches, section)
			if maxMatches > 0 && len(matches) >= maxMatches {
				break
			}
		}
	}
	return matches
}

// getMaxMatches returns the max matches allowed for an element (0 = unlimited).
func getMaxMatches(element schema.StructureElement) int {
	if element.Count != nil {
		return element.Count.Max
	}
	return 1 // Default: exactly one match
}

// getMinMatches returns the minimum matches required for an element.
func getMinMatches(element schema.StructureElement) int {
	if element.Count != nil {
		return element.Count.Min
	}
	if element.Optional {
		return 0
	}
	return 1 // Default: exactly one match required
}

// hasMultiMatch returns true if element can match more than once.
func hasMultiMatch(element schema.StructureElement) bool {
	if element.Count == nil {
		return false
	}
	// Multi-match if max > 1 or max is unlimited (0)
	return element.Count.Max != 1
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
