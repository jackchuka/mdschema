package rules

import (
	"strings"
	"testing"

	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/schema"
	"github.com/jackchuka/mdschema/internal/vast"
)

func TestNewListRule(t *testing.T) {
	rule := NewListRule()
	if rule == nil {
		t.Fatal("NewListRule() returned nil")
	}
}

func TestListRuleName(t *testing.T) {
	rule := NewListRule()
	if rule.Name() != "list" {
		t.Errorf("Name() = %q, want %q", rule.Name(), "list")
	}
}

func TestListRuleMinimum(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\n- item 1\n- item 2\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Requires 2 lists, but only 1 exists
	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				SectionRules: &schema.SectionRules{
					Lists: []schema.ListRule{
						{Min: 2},
					},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewListRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Should detect missing lists")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "2") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Violation should mention required count")
	}
}

func TestListRuleMaximum(t *testing.T) {
	p := parser.New()
	// Use paragraphs between lists to ensure they're parsed as separate lists
	doc, err := p.Parse("test.md", []byte("# Title\n\n- a\n- b\n\nSome text.\n\n- c\n- d\n\nMore text.\n\n- e\n- f\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Max 2 lists, but 3 exist
	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				SectionRules: &schema.SectionRules{
					Lists: []schema.ListRule{
						{Max: 2},
					},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewListRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Should detect too many lists")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "too many") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Violation should mention too many lists")
	}
}

func TestListRuleTypeOrdered(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\n- unordered item\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Requires ordered list, but only unordered exists
	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				SectionRules: &schema.SectionRules{
					Lists: []schema.ListRule{
						{Min: 1, Type: "ordered"},
					},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewListRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Should detect missing ordered list")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "ordered") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Violation should mention ordered lists")
	}
}

func TestListRuleTypeUnordered(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\n1. ordered item\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Requires unordered list, but only ordered exists
	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				SectionRules: &schema.SectionRules{
					Lists: []schema.ListRule{
						{Min: 1, Type: "unordered"},
					},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewListRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Should detect missing unordered list")
	}
}

func TestListRuleSufficient(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\n1. item 1\n2. item 2\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Requires 1 ordered list, and 1 exists
	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				SectionRules: &schema.SectionRules{
					Lists: []schema.ListRule{
						{Min: 1, Type: "ordered"},
					},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewListRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Should have no violations when requirements met, got %d: %v", len(violations), violations)
	}
}

func TestListRuleGenerateContent(t *testing.T) {
	rule := NewListRule()
	var builder strings.Builder

	element := schema.StructureElement{
		Heading: schema.HeadingPattern{Pattern: "## List Section"},
		SectionRules: &schema.SectionRules{
			Lists: []schema.ListRule{
				{Min: 1, Type: "ordered"},
			},
		},
	}

	result := rule.GenerateContent(&builder, element)

	if !result {
		t.Error("GenerateContent() should return true when list rules exist")
	}

	content := builder.String()
	if !strings.Contains(content, "1.") {
		t.Error("Should generate ordered list placeholders")
	}
}

func TestListRuleGenerateContentUnordered(t *testing.T) {
	rule := NewListRule()
	var builder strings.Builder

	element := schema.StructureElement{
		Heading: schema.HeadingPattern{Pattern: "## List Section"},
		SectionRules: &schema.SectionRules{
			Lists: []schema.ListRule{
				{Min: 1, Type: "unordered"},
			},
		},
	}

	result := rule.GenerateContent(&builder, element)

	if !result {
		t.Error("GenerateContent() should return true when list rules exist")
	}

	content := builder.String()
	if !strings.Contains(content, "-") {
		t.Error("Should generate unordered list placeholders")
	}
}

func TestListRuleGenerateContentNoRules(t *testing.T) {
	rule := NewListRule()
	var builder strings.Builder

	element := schema.StructureElement{
		Heading: schema.HeadingPattern{Pattern: "## Section"},
	}

	result := rule.GenerateContent(&builder, element)

	if result {
		t.Error("GenerateContent() should return false when no list rules")
	}
}
