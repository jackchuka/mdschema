package rules

import (
	"strings"
	"testing"

	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/schema"
	"github.com/jackchuka/mdschema/internal/vast"
)

func TestNewImageRule(t *testing.T) {
	rule := NewImageRule()
	if rule == nil {
		t.Fatal("NewImageRule() returned nil")
	}
}

func TestImageRuleName(t *testing.T) {
	rule := NewImageRule()
	if rule.Name() != "image" {
		t.Errorf("Name() = %q, want %q", rule.Name(), "image")
	}
}

func TestImageRuleMinimum(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\n![alt](image.png)\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Requires 2 images, but only 1 exists
	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				SectionRules: &schema.SectionRules{
					Images: []schema.ImageRule{
						{Min: 2},
					},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewImageRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Should detect missing images")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "2") && strings.Contains(v.Message, "1") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Violation should mention required and found count")
	}
}

func TestImageRuleMaximum(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\n![a](1.png)\n![b](2.png)\n![c](3.png)\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Max 2 images, but 3 exist
	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				SectionRules: &schema.SectionRules{
					Images: []schema.ImageRule{
						{Max: 2},
					},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewImageRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Should detect too many images")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "too many") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Violation should mention too many images")
	}
}

func TestImageRuleRequireAlt(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\n![](image.png)\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Requires alt text
	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				SectionRules: &schema.SectionRules{
					Images: []schema.ImageRule{
						{RequireAlt: true},
					},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewImageRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Should detect missing alt text")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "alt") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Violation should mention alt text")
	}
}

func TestImageRuleSufficient(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\n![alt1](1.png)\n![alt2](2.png)\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Requires 2 images with alt, and 2 exist
	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				SectionRules: &schema.SectionRules{
					Images: []schema.ImageRule{
						{Min: 2, RequireAlt: true},
					},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewImageRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Should have no violations when requirements met, got %d: %v", len(violations), violations)
	}
}

func TestImageRuleGenerateContent(t *testing.T) {
	rule := NewImageRule()
	var builder strings.Builder

	element := schema.StructureElement{
		Heading: schema.HeadingPattern{Pattern: "## Images Section"},
		SectionRules: &schema.SectionRules{
			Images: []schema.ImageRule{
				{Min: 2, RequireAlt: true},
			},
		},
	}

	result := rule.GenerateContent(&builder, element)

	if !result {
		t.Error("GenerateContent() should return true when image rules exist")
	}

	content := builder.String()
	if !strings.Contains(content, "![") {
		t.Error("Should generate image placeholders")
	}
}

func TestImageRuleGenerateContentNoRules(t *testing.T) {
	rule := NewImageRule()
	var builder strings.Builder

	element := schema.StructureElement{
		Heading: schema.HeadingPattern{Pattern: "## Section"},
	}

	result := rule.GenerateContent(&builder, element)

	if result {
		t.Error("GenerateContent() should return false when no image rules")
	}
}
