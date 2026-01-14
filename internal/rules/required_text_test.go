package rules

import (
	"strings"
	"testing"

	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/schema"
)

func TestNewRequiredTextRule(t *testing.T) {
	rule := NewRequiredTextRule()
	if rule == nil {
		t.Fatal("NewRequiredTextRule() returned nil")
	}
}

func TestRequiredTextRuleName(t *testing.T) {
	rule := NewRequiredTextRule()
	if rule.Name() != "required-text" {
		t.Errorf("Name() = %q, want %q", rule.Name(), "required-text")
	}
}

func TestRequiredTextExactMatch(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\nThis section contains important text.\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: "# Title",
				SectionRules: &schema.SectionRules{
					RequiredText: []string{"important text"},
				},
			},
		},
	}

	ctx := NewValidationContext(doc, s)
	rule := NewRequiredTextRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Should have no violations when required text is present, got %d", len(violations))
		for _, v := range violations {
			t.Logf("  - %s", v.Message)
		}
	}
}

func TestRequiredTextMissing(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\nSome other content.\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: "# Title",
				SectionRules: &schema.SectionRules{
					RequiredText: []string{"important text"},
				},
			},
		},
	}

	ctx := NewValidationContext(doc, s)
	rule := NewRequiredTextRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Should detect missing required text")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "important text") && strings.Contains(v.Message, "not found") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Violation should mention the missing text")
	}
}

func TestRequiredTextRegexMatch(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\nVersion: 1.2.3\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Use a regex that doesn't require start anchor since content includes full section
	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: "# Title",
				SectionRules: &schema.SectionRules{
					RequiredText: []string{"(?i)version: \\d+\\.\\d+\\.\\d+"},
				},
			},
		},
	}

	ctx := NewValidationContext(doc, s)
	rule := NewRequiredTextRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Should match regex pattern, got %d violations", len(violations))
		for _, v := range violations {
			t.Logf("  - %s", v.Message)
		}
	}
}

func TestRequiredTextRegexNoMatch(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\nVersion: abc\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: "# Title",
				SectionRules: &schema.SectionRules{
					RequiredText: []string{"^Version: \\d+\\.\\d+\\.\\d+"},
				},
			},
		},
	}

	ctx := NewValidationContext(doc, s)
	rule := NewRequiredTextRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Error("Should not match version pattern 'abc'")
	}
}

func TestRequiredTextCaseInsensitive(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\nIMPORTANT NOTE\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: "# Title",
				SectionRules: &schema.SectionRules{
					RequiredText: []string{"(?i)important note"},
				},
			},
		},
	}

	ctx := NewValidationContext(doc, s)
	rule := NewRequiredTextRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Should match case-insensitively, got %d violations", len(violations))
	}
}

func TestRequiredTextGenerateContent(t *testing.T) {
	rule := NewRequiredTextRule()
	var builder strings.Builder

	element := schema.StructureElement{
		Heading: "## Section",
		SectionRules: &schema.SectionRules{
			RequiredText: []string{"important text", "another phrase"},
		},
	}

	result := rule.GenerateContent(&builder, element)

	if !result {
		t.Error("GenerateContent() should return true when required text rules exist")
	}

	content := builder.String()
	if !strings.Contains(content, "important text") {
		t.Error("Should mention required text in generated content")
	}
	if !strings.Contains(content, "another phrase") {
		t.Error("Should mention all required text in generated content")
	}
}

func TestRequiredTextGenerateContentNoRules(t *testing.T) {
	rule := NewRequiredTextRule()
	var builder strings.Builder

	element := schema.StructureElement{
		Heading: "## Section",
	}

	result := rule.GenerateContent(&builder, element)

	if result {
		t.Error("GenerateContent() should return false when no required text rules")
	}
}

func TestContentContainsText(t *testing.T) {
	rule := NewRequiredTextRule()

	tests := []struct {
		content      string
		requiredText string
		want         bool
	}{
		{"Hello world", "world", true},
		{"Hello world", "foo", false},
		{"Version: 1.0.0", "^Version:", true},
		{"No match here", "^Version:", false},
		{"HELLO WORLD", "(?i)hello", true},
	}

	for _, tc := range tests {
		got := rule.contentContainsText(tc.content, tc.requiredText)
		if got != tc.want {
			t.Errorf("contentContainsText(%q, %q) = %v, want %v",
				tc.content, tc.requiredText, got, tc.want)
		}
	}
}
