package vast

import (
	"testing"

	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/schema"
)

func TestBuilderBasic(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\n## Section\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{Heading: schema.HeadingPattern{Pattern: "# Title"}},
		},
	}

	ctx := NewContext(doc, s, "")

	if ctx.Tree == nil {
		t.Fatal("Tree should not be nil")
	}

	if len(ctx.Tree.Roots) != 1 {
		t.Errorf("Expected 1 root, got %d", len(ctx.Tree.Roots))
	}

	root := ctx.Tree.Roots[0]
	if !root.IsBound {
		t.Error("Root should be bound")
	}

	if root.HeadingText() != "Title" {
		t.Errorf("Expected heading 'Title', got '%s'", root.HeadingText())
	}
}

func TestBuilderMultipleMatchingSections(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte(`# Project

## Installation

### Windows
Windows install

### macOS
Mac install

## Installation

### Linux
Linux install

`))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Project"},
				Children: []schema.StructureElement{
					{
						Heading: schema.HeadingPattern{Pattern: "## Installation"},
						Children: []schema.StructureElement{
							{Heading: schema.HeadingPattern{Pattern: "### Windows"}, Optional: true},
							{Heading: schema.HeadingPattern{Pattern: "### macOS"}, Optional: true},
							{Heading: schema.HeadingPattern{Pattern: "### Linux"}, Optional: true},
						},
					},
				},
			},
		},
	}

	ctx := NewContext(doc, s, "")

	// Should have 1 root node (# Project)
	if len(ctx.Tree.Roots) != 1 {
		t.Fatalf("Expected 1 root, got %d", len(ctx.Tree.Roots))
	}

	root := ctx.Tree.Roots[0]

	// Root should have 1 Installation child (first match only)
	installCount := 0
	var installNode *Node
	for _, child := range root.Children {
		if child.Element.Heading.Pattern == "## Installation" {
			installCount++
			installNode = child
		}
	}

	if installCount != 1 {
		t.Errorf("Expected 1 Installation node, got %d", installCount)
	}
	if installNode == nil {
		t.Fatal("Expected Installation node to be present")
	}

	// Installation should have Windows/macOS bound and Linux unbound
	windowsBound := false
	macosBound := false
	linuxFound := false
	linuxBound := false
	for _, child := range installNode.Children {
		switch child.Element.Heading.Pattern {
		case "### Windows":
			windowsBound = child.IsBound
		case "### macOS":
			macosBound = child.IsBound
		case "### Linux":
			linuxFound = true
			linuxBound = child.IsBound
		}
	}

	if !windowsBound {
		t.Error("Installation should have Windows bound")
	}
	if !macosBound {
		t.Error("Installation should have macOS bound")
	}
	if !linuxFound {
		t.Error("Installation should include Linux node")
	}
	if linuxBound {
		t.Error("Installation should have Linux unbound")
	}

	unmatchedInstallation := false
	for _, section := range ctx.Tree.UnmatchedSections {
		if section.Heading != nil && section.Heading.Text == "Installation" {
			unmatchedInstallation = true
			break
		}
	}
	if !unmatchedInstallation {
		t.Error("Expected second Installation section to be unmatched")
	}
}

func TestBuilderUnboundNodes(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				Children: []schema.StructureElement{
					{Heading: schema.HeadingPattern{Pattern: "## Required Section"}},
				},
			},
		},
	}

	ctx := NewContext(doc, s, "")

	root := ctx.Tree.Roots[0]

	// Should have 1 child (unbound)
	if len(root.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(root.Children))
	}

	child := root.Children[0]
	if child.IsBound {
		t.Error("Child should be unbound (section doesn't exist)")
	}

	if child.HeadingText() != "## Required Section" {
		t.Errorf("Expected heading '## Required Section', got '%s'", child.HeadingText())
	}
}

func TestBuilderSectionsNotDoubleMatched(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte(`# My Project

## Section

# LICENSE

MIT
`))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Schema where first element is a regex that could match both headings
	// but LICENSE should only be matched by the second explicit element
	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{Heading: schema.HeadingPattern{Pattern: "# [A-Z][a-z].*", Regex: true}}, // Matches "My Project" but not "LICENSE"
			{Heading: schema.HeadingPattern{Pattern: "# LICENSE"}},                   // Exact match
		},
	}

	ctx := NewContext(doc, s, "")

	// Should have 2 roots
	if len(ctx.Tree.Roots) != 2 {
		t.Fatalf("Expected 2 roots, got %d", len(ctx.Tree.Roots))
	}

	// First root should be My Project
	first := ctx.Tree.Roots[0]
	if !first.IsBound || first.HeadingText() != "My Project" {
		t.Errorf("First root should be 'My Project', got '%s' (bound: %v)", first.HeadingText(), first.IsBound)
	}

	// Second root should be LICENSE
	second := ctx.Tree.Roots[1]
	if !second.IsBound || second.HeadingText() != "LICENSE" {
		t.Errorf("Second root should be 'LICENSE', got '%s' (bound: %v)", second.HeadingText(), second.IsBound)
	}
}

func TestTreeWalk(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\n## Section\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				Children: []schema.StructureElement{
					{Heading: schema.HeadingPattern{Pattern: "## Section"}},
				},
			},
		},
	}

	ctx := NewContext(doc, s, "")

	count := 0
	ctx.Tree.Walk(func(n *Node) bool {
		count++
		return true
	})

	if count != 2 {
		t.Errorf("Walk should visit 2 nodes, visited %d", count)
	}
}

func TestTreeWalkBound(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				Children: []schema.StructureElement{
					{Heading: schema.HeadingPattern{Pattern: "## Missing"}},
				},
			},
		},
	}

	ctx := NewContext(doc, s, "")

	boundCount := 0
	ctx.Tree.WalkBound(func(n *Node) bool {
		boundCount++
		return true
	})

	if boundCount != 1 {
		t.Errorf("WalkBound should visit 1 bound node, visited %d", boundCount)
	}
}

func TestNodeAccessors(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\nSome content.\n\n```go\ncode\n```\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{Heading: schema.HeadingPattern{Pattern: "# Title"}},
		},
	}

	ctx := NewContext(doc, s, "")
	root := ctx.Tree.Roots[0]

	// Test Content()
	content := root.Content()
	if content == "" {
		t.Error("Content() should return section content")
	}

	// Test CodeBlocks()
	codeBlocks := root.CodeBlocks()
	if len(codeBlocks) != 1 {
		t.Errorf("Expected 1 code block, got %d", len(codeBlocks))
	}

	// Test Location()
	line, _ := root.Location()
	if line != 1 {
		t.Errorf("Expected line 1, got %d", line)
	}
}

func TestRegexFlagRequired(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte(`# Hello World

# [A-Z].*
`))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Without regex flag, pattern should be treated as literal
	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{Heading: schema.HeadingPattern{Pattern: "# [A-Z].*"}}, // No Regex flag - should match literally
		},
	}

	ctx := NewContext(doc, s, "")

	if len(ctx.Tree.Roots) != 1 {
		t.Fatalf("Expected 1 root, got %d", len(ctx.Tree.Roots))
	}

	root := ctx.Tree.Roots[0]
	// Should match the literal "# [A-Z].*" heading, not "Hello World"
	if root.HeadingText() != "[A-Z].*" {
		t.Errorf("Expected literal match '[A-Z].*', got '%s'", root.HeadingText())
	}
}

func TestUnboundNodeAccessors(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				Children: []schema.StructureElement{
					{Heading: schema.HeadingPattern{Pattern: "## Missing"}},
				},
			},
		},
	}

	ctx := NewContext(doc, s, "")
	missingNode := ctx.Tree.Roots[0].Children[0]

	// Unbound node should return empty/nil for accessors
	if missingNode.Content() != "" {
		t.Error("Unbound node Content() should be empty")
	}

	if missingNode.CodeBlocks() != nil {
		t.Error("Unbound node CodeBlocks() should be nil")
	}

	// HeadingText should return schema pattern for unbound nodes
	if missingNode.HeadingText() != "## Missing" {
		t.Errorf("Unbound HeadingText() should be schema pattern, got '%s'", missingNode.HeadingText())
	}

	// Location should fall back to parent
	line, _ := missingNode.Location()
	if line != 1 {
		t.Errorf("Unbound Location() should fall back to parent, got line %d", line)
	}
}
