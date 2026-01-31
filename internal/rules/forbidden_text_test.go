package rules

import (
	"strings"
	"testing"

	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/schema"
	"github.com/jackchuka/mdschema/internal/vast"
)

func TestNewForbiddenTextRule(t *testing.T) {
	rule := NewForbiddenTextRule()
	if rule == nil {
		t.Fatal("NewForbiddenTextRule() returned nil")
	}
}

func TestForbiddenTextRuleName(t *testing.T) {
	rule := NewForbiddenTextRule()
	if rule.Name() != "forbidden-text" {
		t.Errorf("Name() = %q, want %q", rule.Name(), "forbidden-text")
	}
}

func TestForbiddenTextRuleViolation(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\nThis is a TODO item.\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Forbids "TODO"
	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				SectionRules: &schema.SectionRules{
					ForbiddenText: []schema.ForbiddenTextPattern{
						{Pattern: "TODO"},
					},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewForbiddenTextRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Should detect forbidden text")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "TODO") && strings.Contains(v.Message, "Forbidden") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Violation should mention forbidden text pattern")
	}
}

func TestForbiddenTextRuleRegex(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\nFIXME-123: Fix this bug.\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Forbids FIXME pattern with regex
	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				SectionRules: &schema.SectionRules{
					ForbiddenText: []schema.ForbiddenTextPattern{
						{Pattern: "FIXME-\\d+"},
					},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewForbiddenTextRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Should detect forbidden regex pattern")
	}
}

func TestForbiddenTextRuleNoViolation(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\nThis is clean content.\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Forbids "TODO" but content doesn't have it
	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				SectionRules: &schema.SectionRules{
					ForbiddenText: []schema.ForbiddenTextPattern{
						{Pattern: "TODO"},
						{Pattern: "FIXME"},
					},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewForbiddenTextRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Should have no violations when forbidden text absent, got %d: %v", len(violations), violations)
	}
}

func TestForbiddenTextRuleGenerateContent(t *testing.T) {
	rule := NewForbiddenTextRule()
	var builder strings.Builder

	element := schema.StructureElement{
		Heading: schema.HeadingPattern{Pattern: "## Section"},
		SectionRules: &schema.SectionRules{
			ForbiddenText: []schema.ForbiddenTextPattern{
				{Pattern: "TODO"},
			},
		},
	}

	result := rule.GenerateContent(&builder, element)

	if !result {
		t.Error("GenerateContent() should return true when forbidden text rules exist")
	}

	content := builder.String()
	if !strings.Contains(content, "NOT contain") {
		t.Error("Should generate warning about forbidden text")
	}
}

func TestForbiddenTextRuleGenerateContentNoRules(t *testing.T) {
	rule := NewForbiddenTextRule()
	var builder strings.Builder

	element := schema.StructureElement{
		Heading: schema.HeadingPattern{Pattern: "## Section"},
	}

	result := rule.GenerateContent(&builder, element)

	if result {
		t.Error("GenerateContent() should return false when no forbidden text rules")
	}
}
