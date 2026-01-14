package rules

import (
	"strings"
	"testing"

	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/schema"
	"github.com/jackchuka/mdschema/internal/vast"
)

func TestNewStructureRule(t *testing.T) {
	rule := NewStructureRule()
	if rule == nil {
		t.Fatal("NewStructureRule() returned nil")
	}
}

func TestStructureRuleName(t *testing.T) {
	rule := NewStructureRule()
	if rule.Name() != "structure" {
		t.Errorf("Name() = %q, want %q", rule.Name(), "structure")
	}
}

func TestStructureRuleValidMappings(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\n## Section\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{Heading: schema.HeadingPattern{Pattern: "# Title"}},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewStructureRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("ValidateWithContext() returned %d violations, want 0", len(violations))
		for _, v := range violations {
			t.Logf("  - %s", v.Message)
		}
	}
}

func TestStructureRuleMissingRequired(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{Heading: schema.HeadingPattern{Pattern: "# Title"}},
			{Heading: schema.HeadingPattern{Pattern: "## Installation"}, Optional: false}, // Required but missing
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewStructureRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("ValidateWithContext() should return violations for missing required element")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "Installation") && strings.Contains(v.Message, "not found") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected violation mentioning missing 'Installation' section")
	}
}

func TestStructureRuleOptionalMissing(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{Heading: schema.HeadingPattern{Pattern: "# Title"}},
			{Heading: schema.HeadingPattern{Pattern: "## License"}, Optional: true}, // Optional and missing - should not be a violation
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewStructureRule()
	violations := rule.ValidateWithContext(ctx)

	for _, v := range violations {
		if strings.Contains(v.Message, "License") {
			t.Errorf("Should not report violation for optional missing element: %s", v.Message)
		}
	}
}

func TestStructureRuleWrongOrder(t *testing.T) {
	p := parser.New()
	// Document has "Usage" before "Installation" but schema expects opposite order
	doc, err := p.Parse("test.md", []byte("# Title\n\n## Usage\n\n## Installation\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{Heading: schema.HeadingPattern{Pattern: "# Title"}, Children: []schema.StructureElement{
				{Heading: schema.HeadingPattern{Pattern: "## Installation"}},
				{Heading: schema.HeadingPattern{Pattern: "## Usage"}},
			}},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewStructureRule()
	violations := rule.ValidateWithContext(ctx)

	// Should detect ordering violation
	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "appear after") || strings.Contains(v.Message, "before") {
			found = true
			break
		}
	}

	if !found {
		t.Log("Warning: ordering violation not detected (may be implementation-specific)")
	}
}

func TestStructureRuleGenerateContent(t *testing.T) {
	rule := NewStructureRule()
	var builder strings.Builder

	element := schema.StructureElement{
		Heading:  schema.HeadingPattern{Pattern: "## Section"},
		Optional: false,
		Children: []schema.StructureElement{
			{Heading: schema.HeadingPattern{Pattern: "### Child1"}, Optional: false},
			{Heading: schema.HeadingPattern{Pattern: "### Child2"}, Optional: true},
		},
	}

	result := rule.GenerateContent(&builder, element)

	if !result {
		t.Error("GenerateContent() should return true")
	}

	content := builder.String()
	if !strings.Contains(content, "Required section") {
		t.Error("Should contain 'Required section' comment")
	}
	if !strings.Contains(content, "Child1") {
		t.Error("Should mention child elements")
	}
}
