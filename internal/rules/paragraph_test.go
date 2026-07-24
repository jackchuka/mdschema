package rules

import (
	"strings"
	"testing"

	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/schema"
	"github.com/jackchuka/mdschema/internal/vast"
)

// paragraphSchema builds a schema requiring "# T" > "## Overview" with the
// given paragraph min/max constraints.
func paragraphSchema(min, max int) *schema.Schema {
	return &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# T"},
				Children: []schema.StructureElement{
					{
						Heading: schema.HeadingPattern{Pattern: "## Overview"},
						SectionRules: &schema.SectionRules{
							Paragraphs: &schema.ParagraphRule{Min: min, Max: max},
						},
					},
				},
			},
		},
	}
}

func TestParagraphRuleMinFails(t *testing.T) {
	md := "# T\n\n## Overview\n\n- just a list\n"
	p := parser.New()
	doc, err := p.Parse("test.md", []byte(md))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	sch := paragraphSchema(1, 0) // min:1, max:0(none) on "## Overview"
	ctx := vast.NewContext(doc, sch, "")

	v := NewParagraphRule().ValidateWithContext(ctx)
	if len(v) != 1 {
		t.Fatalf("expected 1 violation, got %d: %+v", len(v), v)
	}
	// A bullet list contains no top-level paragraph, so found == 0.
	if want := "Section 'Overview' has too few paragraphs (minimum 1, found 0)"; v[0].Message != want {
		t.Errorf("message = %q, want %q", v[0].Message, want)
	}
}

func TestParagraphRuleMinPasses(t *testing.T) {
	md := "# T\n\n## Overview\n\nA real prose paragraph.\n"
	p := parser.New()
	doc, err := p.Parse("test.md", []byte(md))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	ctx := vast.NewContext(doc, paragraphSchema(1, 0), "")
	if v := NewParagraphRule().ValidateWithContext(ctx); len(v) != 0 {
		t.Fatalf("expected 0 violations, got %d: %+v", len(v), v)
	}
}

func TestParagraphRuleMaxFails(t *testing.T) {
	md := "# T\n\n## Overview\n\nOne.\n\nTwo.\n\nThree.\n"
	p := parser.New()
	doc, err := p.Parse("test.md", []byte(md))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	ctx := vast.NewContext(doc, paragraphSchema(0, 2), "")
	v := NewParagraphRule().ValidateWithContext(ctx)
	if len(v) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(v))
	}
	if want := "Section 'Overview' has too many paragraphs (maximum 2, found 3)"; v[0].Message != want {
		t.Errorf("message = %q, want %q", v[0].Message, want)
	}
}

func TestParagraphRuleGenerateContent(t *testing.T) {
	rule := NewParagraphRule()
	var builder strings.Builder

	element := schema.StructureElement{
		Heading: schema.HeadingPattern{Pattern: "## Overview"},
		SectionRules: &schema.SectionRules{
			Paragraphs: &schema.ParagraphRule{Min: 1, Max: 3},
		},
	}

	result := rule.GenerateContent(&builder, element)

	if !result {
		t.Error("GenerateContent() should return true when paragraph rules exist")
	}

	content := builder.String()
	if !strings.Contains(content, "<!-- Paragraph requirements: -->") {
		t.Error("Should generate paragraph requirements header")
	}
	if !strings.Contains(content, "<!-- Minimum 1 paragraphs required -->") {
		t.Error("Should generate minimum paragraphs comment")
	}
	if !strings.Contains(content, "<!-- Maximum 3 paragraphs allowed -->") {
		t.Error("Should generate maximum paragraphs comment")
	}
}

func TestParagraphRuleGenerateContentNoRules(t *testing.T) {
	rule := NewParagraphRule()
	var builder strings.Builder

	element := schema.StructureElement{
		Heading: schema.HeadingPattern{Pattern: "## Overview"},
	}

	result := rule.GenerateContent(&builder, element)

	if result {
		t.Error("GenerateContent() should return false when no paragraph rules")
	}
}
