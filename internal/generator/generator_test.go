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

func TestGenerateFrontmatter(t *testing.T) {
	g := New()

	s := &schema.Schema{
		Frontmatter: &schema.FrontmatterConfig{
			// Optional: false is default, meaning frontmatter is required
			Fields: []schema.FrontmatterField{
				{Name: "title", Type: schema.FieldTypeString},               // required by default
				{Name: "date", Type: schema.FieldTypeDate},                  // required by default
				{Name: "tags", Optional: true, Type: schema.FieldTypeArray}, // explicitly optional
				{Name: "draft", Type: schema.FieldTypeBoolean},              // required by default
				{Name: "version", Type: schema.FieldTypeNumber},             // required by default
			},
		},
		Structure: []schema.StructureElement{
			{Heading: schema.HeadingPattern{Pattern: "# Title"}},
		},
	}

	output := g.Generate(s)

	// Check frontmatter delimiters
	if !strings.Contains(output, "---\n") {
		t.Error("Generated output should contain frontmatter delimiters")
	}

	// Check required fields have comments
	if !strings.Contains(output, "title: \"TODO\" # required") {
		t.Error("Generated output should contain title field with required comment")
	}

	if !strings.Contains(output, "date: 2024-01-01 # required") {
		t.Error("Generated output should contain date field")
	}

	// Check optional fields don't have required comment
	if !strings.Contains(output, "tags: [\"item1\", \"item2\"]\n") {
		t.Error("Generated output should contain tags array field without required comment")
	}

	// Check boolean placeholder (now required by default)
	if !strings.Contains(output, "draft: false # required") {
		t.Error("Generated output should contain draft boolean field with required comment")
	}

	// Check number placeholder (now required by default)
	if !strings.Contains(output, "version: 0 # required") {
		t.Error("Generated output should contain version number field with required comment")
	}
}

func TestGenerateFrontmatterFormats(t *testing.T) {
	g := New()

	s := &schema.Schema{
		Frontmatter: &schema.FrontmatterConfig{
			Fields: []schema.FrontmatterField{
				{Name: "author_email", Format: schema.FieldFormatEmail},
				{Name: "website", Format: schema.FieldFormatURL},
				{Name: "published", Format: schema.FieldFormatDate},
			},
		},
		Structure: []schema.StructureElement{},
	}

	output := g.Generate(s)

	if !strings.Contains(output, "author_email: user@example.com") {
		t.Error("Generated output should contain email format placeholder")
	}

	if !strings.Contains(output, "website: https://example.com") {
		t.Error("Generated output should contain URL format placeholder")
	}

	if !strings.Contains(output, "published: 2024-01-01") {
		t.Error("Generated output should contain date format placeholder")
	}
}

func TestGenerateNoFrontmatter(t *testing.T) {
	g := New()

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{Heading: schema.HeadingPattern{Pattern: "# Title"}},
		},
	}

	output := g.Generate(s)

	// Should not contain frontmatter delimiters when no frontmatter config
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line == "---" {
			t.Error("Generated output should not contain frontmatter when not configured")
			break
		}
	}
}

func TestGenerateLiteral(t *testing.T) {
	g := New()

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Literal: "# Title"},
				Children: []schema.StructureElement{
					{Heading: schema.HeadingPattern{Literal: "## Required Child"}},
					{Heading: schema.HeadingPattern{Literal: "## Optional Child"}, Optional: true},
				},
			},
		},
	}

	output := g.Generate(s)

	if !strings.Contains(output, "# Title") {
		t.Error("Generated output should contain parent heading")
	}

	if !strings.Contains(output, "## Required Child") {
		t.Error("Generated output should contain required child heading")
	}

	if !strings.Contains(output, "## Optional Child") {
		t.Error("Generated output should contain optional child heading")
	}

	if !strings.Contains(output, "<!-- 1. ## Required Child (required) -->") {
		t.Error("Generated output should contain required child heading comment")
	}

	if !strings.Contains(output, "<!-- 2. ## Optional Child (optional) -->") {
		t.Error("Generated output should contain optional child heading comment")
	}
}
