package rules

import (
	"strings"
	"testing"

	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/schema"
	"github.com/jackchuka/mdschema/internal/vast"
)

func TestNewCodeBlockRule(t *testing.T) {
	rule := NewCodeBlockRule()
	if rule == nil {
		t.Fatal("NewCodeBlockRule() returned nil")
	}
}

func TestCodeBlockRuleName(t *testing.T) {
	rule := NewCodeBlockRule()
	if rule.Name() != "codeblock" {
		t.Errorf("Name() = %q, want %q", rule.Name(), "codeblock")
	}
}

func TestCodeBlockRuleMinimum(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\n```go\ncode\n```\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Requires 2 go code blocks, but only 1 exists
	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				SectionRules: &schema.SectionRules{
					CodeBlocks: []schema.CodeBlockRule{
						{Lang: "go", Min: 2},
					},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewCodeBlockRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Should detect missing code blocks")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "go") && strings.Contains(v.Message, "2") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Violation should mention language and required count")
	}
}

func TestCodeBlockRuleMaximum(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\n```go\na\n```\n\n```go\nb\n```\n\n```go\nc\n```\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Max 2 go code blocks, but 3 exist
	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				SectionRules: &schema.SectionRules{
					CodeBlocks: []schema.CodeBlockRule{
						{Lang: "go", Max: 2},
					},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewCodeBlockRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Should detect too many code blocks")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "too many") || strings.Contains(v.Message, "max") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Violation should mention too many code blocks")
	}
}

func TestCodeBlockRuleLanguageSpecific(t *testing.T) {
	p := parser.New()
	// Has 1 bash and 1 go block
	doc, err := p.Parse("test.md", []byte("# Title\n\n```bash\necho\n```\n\n```go\ncode\n```\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Requires 2 bash code blocks
	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				SectionRules: &schema.SectionRules{
					CodeBlocks: []schema.CodeBlockRule{
						{Lang: "bash", Min: 2},
					},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewCodeBlockRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Should detect insufficient bash code blocks (go block shouldn't count)")
	}
}

func TestCodeBlockRuleNoRequirements(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// No code block requirements
	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{Heading: schema.HeadingPattern{Pattern: "# Title"}},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewCodeBlockRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Should have no violations when no code block requirements, got %d", len(violations))
	}
}

func TestCodeBlockRuleSufficient(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\n```go\na\n```\n\n```go\nb\n```\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Requires 2 go code blocks, and 2 exist
	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				SectionRules: &schema.SectionRules{
					CodeBlocks: []schema.CodeBlockRule{
						{Lang: "go", Min: 2},
					},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewCodeBlockRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Should have no violations when requirements met, got %d: %v", len(violations), violations)
	}
}

func TestCodeBlockRuleGenerateContent(t *testing.T) {
	rule := NewCodeBlockRule()
	var builder strings.Builder

	element := schema.StructureElement{
		Heading: schema.HeadingPattern{Pattern: "## Code Section"},
		SectionRules: &schema.SectionRules{
			CodeBlocks: []schema.CodeBlockRule{
				{Lang: "go", Min: 2},
			},
		},
	}

	result := rule.GenerateContent(&builder, element)

	if !result {
		t.Error("GenerateContent() should return true when code block rules exist")
	}

	content := builder.String()
	if !strings.Contains(content, "```go") {
		t.Error("Should generate go code block placeholders")
	}
}

func TestCodeBlockRuleGenerateContentNoRules(t *testing.T) {
	rule := NewCodeBlockRule()
	var builder strings.Builder

	element := schema.StructureElement{
		Heading: schema.HeadingPattern{Pattern: "## Section"},
	}

	result := rule.GenerateContent(&builder, element)

	if result {
		t.Error("GenerateContent() should return false when no code block rules")
	}
}
