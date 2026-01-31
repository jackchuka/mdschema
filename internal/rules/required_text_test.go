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
						{Pattern: "(?i)version: \\d+\\.\\d+\\.\\d+"},
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
						{Pattern: "^Version: \\d+\\.\\d+\\.\\d+"},
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
						{Pattern: "(?i)important note"},
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
				{Literal: "important text"},
				{Literal: "another phrase"},
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
		pattern schema.RequiredTextPattern
		want    bool
	}{
		// Literal patterns (substring match)
		{"Hello world", schema.RequiredTextPattern{Literal: "world"}, true},
		{"Hello world", schema.RequiredTextPattern{Literal: "foo"}, false},
		// Regex patterns
		{"Version: 1.0.0", schema.RequiredTextPattern{Pattern: "^Version:"}, true},
		{"No match here", schema.RequiredTextPattern{Pattern: "^Version:"}, false},
		{"HELLO WORLD", schema.RequiredTextPattern{Pattern: "(?i)hello"}, true},
		// Literal treats regex metacharacters as literal
		{"Hello world", schema.RequiredTextPattern{Literal: "^Hello"}, false},
		{"^Hello world", schema.RequiredTextPattern{Literal: "^Hello"}, true},
	}

	for _, tc := range tests {
		got := rule.contentContainsPattern(tc.content, tc.pattern)
		patternStr := tc.pattern.Literal
		if patternStr == "" {
			patternStr = tc.pattern.Pattern
		}
		if got != tc.want {
			t.Errorf("contentContainsPattern(%q, %q) = %v, want %v",
				tc.content, patternStr, got, tc.want)
		}
	}
}
