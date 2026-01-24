package rules

import (
	"strings"
	"testing"

	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/schema"
	"github.com/jackchuka/mdschema/internal/vast"
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
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				SectionRules: &schema.SectionRules{
					RequiredText: []schema.RequiredTextPattern{{Pattern: "important text"}},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s, "")
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
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				SectionRules: &schema.SectionRules{
					RequiredText: []schema.RequiredTextPattern{{Pattern: "important text"}},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s, "")
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

	// Use a regex with explicit regex: true flag
	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				SectionRules: &schema.SectionRules{
					RequiredText: []schema.RequiredTextPattern{
						{Pattern: "(?i)version: \\d+\\.\\d+\\.\\d+", Regex: true},
					},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s, "")
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
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				SectionRules: &schema.SectionRules{
					RequiredText: []schema.RequiredTextPattern{
						{Pattern: "^Version: \\d+\\.\\d+\\.\\d+", Regex: true},
					},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s, "")
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
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				SectionRules: &schema.SectionRules{
					RequiredText: []schema.RequiredTextPattern{
						{Pattern: "(?i)important note", Regex: true},
					},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s, "")
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
		Heading: schema.HeadingPattern{Pattern: "## Section"},
		SectionRules: &schema.SectionRules{
			RequiredText: []schema.RequiredTextPattern{
				{Pattern: "important text"},
				{Pattern: "another phrase"},
			},
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
		Heading: schema.HeadingPattern{Pattern: "## Section"},
	}

	result := rule.GenerateContent(&builder, element)

	if result {
		t.Error("GenerateContent() should return false when no required text rules")
	}
}

func TestContentContainsPattern(t *testing.T) {
	rule := NewRequiredTextRule()

	tests := []struct {
		content string
		pattern string
		isRegex bool
		want    bool
	}{
		{"Hello world", "world", false, true},
		{"Hello world", "foo", false, false},
		{"Version: 1.0.0", "^Version:", true, true},
		{"No match here", "^Version:", true, false},
		{"HELLO WORLD", "(?i)hello", true, true},
		// Without regex flag, regex metacharacters are treated literally
		{"Hello world", "^Hello", false, false},
		{"^Hello world", "^Hello", false, true},
	}

	for _, tc := range tests {
		got := rule.contentContainsPattern(tc.content, tc.pattern, tc.isRegex)
		if got != tc.want {
			t.Errorf("contentContainsPattern(%q, %q, %v) = %v, want %v",
				tc.content, tc.pattern, tc.isRegex, got, tc.want)
		}
	}
}
