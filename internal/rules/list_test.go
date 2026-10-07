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

	ctx := vast.NewContext(doc, s, "")
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

	ctx := vast.NewContext(doc, s, "")
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
						{Min: 1, Type: schema.ListTypeOrdered},
					},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s, "")
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
						{Min: 1, Type: schema.ListTypeUnordered},
					},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s, "")
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
						{Min: 1, Type: schema.ListTypeOrdered},
					},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s, "")
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
				{Min: 1, Type: schema.ListTypeOrdered},
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
				{Min: 1, Type: schema.ListTypeUnordered},
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

func itemsMustMatchSchema(rule schema.ListRule) *schema.Schema {
	return &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading:      schema.HeadingPattern{Pattern: "# Title"},
				SectionRules: &schema.SectionRules{Lists: []schema.ListRule{rule}},
			},
		},
	}
}

func regexItems(patterns ...string) []schema.RequiredTextPattern {
	result := make([]schema.RequiredTextPattern, 0, len(patterns))
	for _, p := range patterns {
		result = append(result, schema.RequiredTextPattern{Pattern: p})
	}
	return result
}

func TestListRuleItemsMustMatch(t *testing.T) {
	p := parser.New()
	md := "# Title\n\n- Button\n- **Card**\n  - field\n- `X`\n- Dialog: modal window\n"
	doc, err := p.Parse("test.md", []byte(md))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := itemsMustMatchSchema(schema.ListRule{ItemsMustMatch: regexItems(`^[A-Za-z][A-Za-z0-9]*(\s.*|:.*)?$`)})
	violations := NewListRule().ValidateWithContext(vast.NewContext(doc, s, ""))

	if len(violations) != 2 {
		t.Fatalf("expected 2 violations, got %d: %v", len(violations), violations)
	}

	wantItems := []string{"'**Card**'", "'`X`'"}
	wantLines := []int{4, 6}
	for i, v := range violations {
		if !strings.Contains(v.Message, wantItems[i]) {
			t.Errorf("violation %d message %q should contain raw item %s", i, v.Message, wantItems[i])
		}
		if !strings.Contains(v.Message, "does not match") {
			t.Errorf("violation %d message %q should mention the pattern", i, v.Message)
		}
		if v.Line != wantLines[i] {
			t.Errorf("violation %d line = %d, want %d", i, v.Line, wantLines[i])
		}
	}
}

func TestListRuleItemsMustMatchAllPatterns(t *testing.T) {
	p := parser.New()
	md := "# Title\n\n- Alpha one\n- Beta\n- gamma one\n"
	doc, err := p.Parse("test.md", []byte(md))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := itemsMustMatchSchema(schema.ListRule{ItemsMustMatch: regexItems(`^[A-Z]`, `\sone$`)})
	violations := NewListRule().ValidateWithContext(vast.NewContext(doc, s, ""))

	// "Alpha one" passes both; "Beta" fails the 2nd; "gamma one" fails the 1st.
	want := []string{"'Beta' in section 'Title' does not match '\\sone$'", "'gamma one' in section 'Title' does not match '^[A-Z]'"}
	if len(violations) != len(want) {
		t.Fatalf("expected %d violations, got %d: %v", len(want), len(violations), violations)
	}
	for i, v := range violations {
		if !strings.Contains(v.Message, want[i]) {
			t.Errorf("violation %d message %q should contain %q", i, v.Message, want[i])
		}
	}
}

func TestListRuleItemsMustMatchLiteral(t *testing.T) {
	p := parser.New()
	md := "# Title\n\n- see [docs](x.md)\n- no link here\n"
	doc, err := p.Parse("test.md", []byte(md))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := itemsMustMatchSchema(schema.ListRule{ItemsMustMatch: []schema.RequiredTextPattern{{Literal: "]("}}})
	violations := NewListRule().ValidateWithContext(vast.NewContext(doc, s, ""))

	if len(violations) != 1 || !strings.Contains(violations[0].Message, "'no link here'") {
		t.Errorf("literal form should be a raw substring match, got %v", violations)
	}
}

func TestListRuleItemsMustMatchIgnoresNestedItems(t *testing.T) {
	p := parser.New()
	md := "# Title\n\n- Parent\n  - nested-not-matching!\n    - deeper!\n"
	doc, err := p.Parse("test.md", []byte(md))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := itemsMustMatchSchema(schema.ListRule{ItemsMustMatch: regexItems(`^[A-Z][a-z]+$`)})
	violations := NewListRule().ValidateWithContext(vast.NewContext(doc, s, ""))

	if len(violations) != 0 {
		t.Errorf("nested items should not be checked, got %v", violations)
	}
}

func TestListRuleItemsMustMatchRespectsType(t *testing.T) {
	p := parser.New()
	md := "# Title\n\n- bad item\n\nText.\n\n1. Good\n"
	doc, err := p.Parse("test.md", []byte(md))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := itemsMustMatchSchema(schema.ListRule{Type: schema.ListTypeOrdered, ItemsMustMatch: regexItems(`^[A-Z]`)})
	violations := NewListRule().ValidateWithContext(vast.NewContext(doc, s, ""))

	if len(violations) != 0 {
		t.Errorf("unordered list should be ignored by an ordered rule, got %v", violations)
	}
}

func TestListRuleItemsMustMatchMultilineItem(t *testing.T) {
	p := parser.New()
	md := "# Title\n\n- Alpha first line\n  continued\n"
	doc, err := p.Parse("test.md", []byte(md))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := itemsMustMatchSchema(schema.ListRule{ItemsMustMatch: regexItems(`^Alpha first line continued$`)})
	violations := NewListRule().ValidateWithContext(vast.NewContext(doc, s, ""))

	if len(violations) != 0 {
		t.Errorf("soft-wrapped lines should be joined with a space, got %v", violations)
	}
}

func TestListRuleGenerateContentItemsMustMatch(t *testing.T) {
	var builder strings.Builder
	element := schema.StructureElement{
		Heading: schema.HeadingPattern{Pattern: "## Components"},
		SectionRules: &schema.SectionRules{
			Lists: []schema.ListRule{{Min: 1, ItemsMustMatch: []schema.RequiredTextPattern{{Pattern: `^[A-Z]`}, {Literal: "TODO"}}}},
		},
	}

	NewListRule().GenerateContent(&builder, element)

	out := builder.String()
	for _, want := range []string{"Each list item must match: ^[A-Z] (regex)", "Each list item must contain: TODO"} {
		if !strings.Contains(out, want) {
			t.Errorf("generated content should contain %q, got:\n%s", want, out)
		}
	}
}
