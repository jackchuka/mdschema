package parser

import "strings"

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
