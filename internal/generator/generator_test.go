package generator

import (
	"strings"
	"testing"

	"github.com/jackchuka/mdschema/internal/schema"
)

func TestNew(t *testing.T) {
	g := New()
	if g == nil {
		t.Fatal("New() returned nil")
	}
	if g.ruleGenerator == nil {
		t.Fatal("New() returned generator with nil ruleGenerator")
	}
}

func TestGenerateBasicStructure(t *testing.T) {
	g := New()

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{Heading: schema.HeadingPattern{Pattern: "# Title"}},
		},
	}

	output := g.Generate(s)

	if !strings.Contains(output, "# Title") {
		t.Error("Generated output should contain the heading")
	}

	if !strings.Contains(output, "<!-- Generated from schema -->") {
		t.Error("Generated output should contain header comment")
	}
}

func TestGenerateOptionalSection(t *testing.T) {
	g := New()

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{Heading: schema.HeadingPattern{Pattern: "# Title"}, Optional: true},
		},
	}

	output := g.Generate(s)

	if !strings.Contains(output, "Optional section") {
		t.Error("Generated output should mark optional sections")
	}
}

func TestGenerateNestedChildren(t *testing.T) {
	g := New()

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				Children: []schema.StructureElement{
					{Heading: schema.HeadingPattern{Pattern: "## Child"}},
					{Heading: schema.HeadingPattern{Pattern: "## Another Child"}},
				},
			},
		},
	}

	output := g.Generate(s)

	if !strings.Contains(output, "# Title") {
		t.Error("Generated output should contain parent heading")
	}

	if !strings.Contains(output, "## Child") {
		t.Error("Generated output should contain child heading")
	}

	if !strings.Contains(output, "## Another Child") {
		t.Error("Generated output should contain second child heading")
	}
}

func TestGenerateWithCodeBlocks(t *testing.T) {
	g := New()

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Installation"},
				SectionRules: &schema.SectionRules{
					CodeBlocks: []schema.CodeBlockRule{
						{Lang: "bash", Min: 1},
					},
				},
			},
		},
	}

	output := g.Generate(s)

	if !strings.Contains(output, "```bash") {
		t.Error("Generated output should contain bash code block")
	}
}

func TestGenerateWithRequiredText(t *testing.T) {
	g := New()

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Section"},
				SectionRules: &schema.SectionRules{
					RequiredText: []schema.RequiredTextPattern{{Pattern: "important text"}},
				},
			},
		},
	}

	output := g.Generate(s)

	if !strings.Contains(output, "important text") {
		t.Error("Generated output should mention required text")
	}
}

func TestExtractHeadingText(t *testing.T) {
	g := New()

	tests := []struct {
		input string
		want  string
	}{
		{"# Title", "Title"},
		{"## Section", "Section"},
		{"### Subsection", "Subsection"},
		{"# [a-zA-Z]+", "[a-zA-Z]+"},
		{"Title", "Title"},
		{"  # Spaced Title  ", "Spaced Title"},
	}

	for _, tc := range tests {
		got := g.extractHeadingText(tc.input)
		if got != tc.want {
			t.Errorf("extractHeadingText(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestGenerateMultipleTopLevel(t *testing.T) {
	g := New()

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{Heading: schema.HeadingPattern{Pattern: "# First"}},
			{Heading: schema.HeadingPattern{Pattern: "# Second"}},
			{Heading: schema.HeadingPattern{Pattern: "# Third"}},
		},
	}

	output := g.Generate(s)

	if strings.Count(output, "# First") != 1 {
		t.Error("Should generate First heading exactly once")
	}
	if strings.Count(output, "# Second") != 1 {
		t.Error("Should generate Second heading exactly once")
	}
	if strings.Count(output, "# Third") != 1 {
		t.Error("Should generate Third heading exactly once")
	}
}

func TestGenerateDeeplyNested(t *testing.T) {
	g := New()

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Level 1"},
				Children: []schema.StructureElement{
					{
						Heading: schema.HeadingPattern{Pattern: "## Level 2"},
						Children: []schema.StructureElement{
							{
								Heading: schema.HeadingPattern{Pattern: "### Level 3"},
							},
						},
					},
				},
			},
		},
	}

	output := g.Generate(s)

	if !strings.Contains(output, "# Level 1") {
		t.Error("Should contain level 1 heading")
	}
	if !strings.Contains(output, "## Level 2") {
		t.Error("Should contain level 2 heading")
	}
	if !strings.Contains(output, "### Level 3") {
		t.Error("Should contain level 3 heading")
	}
}

func TestGenerateEmptySchema(t *testing.T) {
	g := New()

	s := &schema.Schema{
		Structure: []schema.StructureElement{},
	}

	output := g.Generate(s)

	// Should just have the header comment
	if !strings.Contains(output, "<!-- Generated from schema -->") {
		t.Error("Should contain header comment even for empty schema")
	}
}
