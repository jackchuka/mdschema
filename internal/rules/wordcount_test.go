package rules

import (
	"strings"
	"testing"

	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/schema"
	"github.com/jackchuka/mdschema/internal/vast"
)

func TestNewWordCountRule(t *testing.T) {
	rule := NewWordCountRule()
	if rule == nil {
		t.Fatal("NewWordCountRule() returned nil")
	}
}

func TestWordCountRuleName(t *testing.T) {
	rule := NewWordCountRule()
	if rule.Name() != "word-count" {
		t.Errorf("Name() = %q, want %q", rule.Name(), "word-count")
	}
}

func TestWordCountRuleTooFew(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\nShort content.\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Requires minimum 10 words, but only 2 exist
	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				SectionRules: &schema.SectionRules{
					WordCount: &schema.WordCountRule{Min: 10},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewWordCountRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Should detect too few words")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "too few") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Violation should mention too few words")
	}
}

func TestWordCountRuleTooMany(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\nThis is a very long sentence with many words that should exceed the maximum limit.\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Max 5 words, but content has more
	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				SectionRules: &schema.SectionRules{
					WordCount: &schema.WordCountRule{Max: 5},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewWordCountRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Should detect too many words")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "too many") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Violation should mention too many words")
	}
}

func TestWordCountRuleSufficient(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\nThis is enough content for the section.\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Requires 5-20 words
	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				SectionRules: &schema.SectionRules{
					WordCount: &schema.WordCountRule{Min: 5, Max: 20},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewWordCountRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Should have no violations when word count is within range, got %d: %v", len(violations), violations)
	}
}

func TestWordCountRuleCountWords(t *testing.T) {
	rule := NewWordCountRule()

	tests := []struct {
		content string
		want    int
	}{
		{"", 0},
		{"one", 1},
		{"one two", 2},
		{"one  two", 2},
		{"  one  two  ", 2},
		{"one two three four five", 5},
		{"word\nword", 2},
		{"word\t\tword", 2},
	}

	for _, tt := range tests {
		got := rule.countWords(tt.content)
		if got != tt.want {
			t.Errorf("countWords(%q) = %d, want %d", tt.content, got, tt.want)
		}
	}
}

func TestWordCountRuleGenerateContent(t *testing.T) {
	rule := NewWordCountRule()
	var builder strings.Builder

	element := schema.StructureElement{
		Heading: schema.HeadingPattern{Pattern: "## Section"},
		SectionRules: &schema.SectionRules{
			WordCount: &schema.WordCountRule{Min: 50, Max: 200},
		},
	}

	result := rule.GenerateContent(&builder, element)

	if !result {
		t.Error("GenerateContent() should return true when word count rules exist")
	}

	content := builder.String()
	if !strings.Contains(content, "50") || !strings.Contains(content, "200") {
		t.Error("Should generate word count requirements")
	}
}

func TestWordCountRuleGenerateContentNoRules(t *testing.T) {
	rule := NewWordCountRule()
	var builder strings.Builder

	element := schema.StructureElement{
		Heading: schema.HeadingPattern{Pattern: "## Section"},
	}

	result := rule.GenerateContent(&builder, element)

	if result {
		t.Error("GenerateContent() should return false when no word count rules")
	}
}
