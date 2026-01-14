package parser

import "strings"

// buildHierarchicalSections creates a hierarchical tree of sections
func (p *Parser) buildHierarchicalSections(headings []*Heading, codeBlocks []*CodeBlock, tables []*Table, links []*Link, images []*Image, lists []*List, content []byte) *Section {
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
		Lists:      make([]*List, 0),
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
			Lists:      make([]*List, 0),
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
	p.setEndLinesAndContent(root, codeBlocks, tables, links, images, lists, content)

	return root
}

// setEndLinesAndContent sets end lines first, then associates content with sections
func (p *Parser) setEndLinesAndContent(section *Section, codeBlocks []*CodeBlock, tables []*Table, links []*Link, images []*Image, lists []*List, content []byte) {
	contentLines := strings.Split(string(content), "\n")

	// First pass: set all end lines correctly (bottom-up via recursion)
	p.setEndLines(section, len(contentLines))

	// Second pass: associate elements and extract content
	p.associateContent(section, codeBlocks, tables, links, images, lists, contentLines)
}

// setEndLines recursively sets end lines for all sections (top-down)
func (p *Parser) setEndLines(section *Section, boundary int) {
	// First, determine this section's end boundary
	if section.Parent != nil {
		// Check for next sibling
		if nextStart := p.findNextSiblingStart(section); nextStart > 0 {
			section.EndLine = nextStart - 1
		} else {
			// No next sibling - use parent's boundary
			section.EndLine = boundary
		}
	} else {
		// Root section uses document end
		section.EndLine = boundary
	}

	// Now process children with this section's boundary
	for _, child := range section.Children {
		p.setEndLines(child, section.EndLine)
	}
}

// findNextSiblingStart finds the start line of the next sibling, or 0 if none
func (p *Parser) findNextSiblingStart(section *Section) int {
	if section.Parent == nil {
		return 0
	}
	parentChildren := section.Parent.Children
	for i, child := range parentChildren {
		if child == section && i+1 < len(parentChildren) {
			return parentChildren[i+1].StartLine
		}
	}
	return 0
}

// associateContent recursively extracts content and associates elements with sections
func (p *Parser) associateContent(section *Section, codeBlocks []*CodeBlock, tables []*Table, links []*Link, images []*Image, lists []*List, contentLines []string) {
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

	// Associate elements with this section (including root section for intro content)
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

	for _, link := range links {
		if link.Line >= section.StartLine && link.Line <= section.EndLine {
			// Check if it belongs to a child section
			belongsToChild := false
			for _, child := range section.Children {
				if link.Line >= child.StartLine && link.Line <= child.EndLine {
					belongsToChild = true
					break
				}
			}
			if !belongsToChild {
				section.Links = append(section.Links, link)
			}
		}
	}

	for _, image := range images {
		if image.Line >= section.StartLine && image.Line <= section.EndLine {
			// Check if it belongs to a child section
			belongsToChild := false
			for _, child := range section.Children {
				if image.Line >= child.StartLine && image.Line <= child.EndLine {
					belongsToChild = true
					break
				}
			}
			if !belongsToChild {
				section.Images = append(section.Images, image)
			}
		}
	}

	for _, list := range lists {
		if list.Line >= section.StartLine && list.Line <= section.EndLine {
			// Check if it belongs to a child section
			belongsToChild := false
			for _, child := range section.Children {
				if list.Line >= child.StartLine && list.Line <= child.EndLine {
					belongsToChild = true
					break
				}
			}
			if !belongsToChild {
				section.Lists = append(section.Lists, list)
			}
		}
	}

	// Recursively process children
	for _, child := range section.Children {
		p.associateContent(child, codeBlocks, tables, links, images, lists, contentLines)
	}
}
