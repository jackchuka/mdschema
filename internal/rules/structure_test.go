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
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				Children: []schema.StructureElement{
					{Heading: schema.HeadingPattern{Pattern: "## Section"}},
				},
			},
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
		t.Error("Expected violation for wrong ordering of sections")
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

func TestStructureRuleUnmatchedHeading(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\n## Extra\n"))
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
	if len(violations) == 0 {
		t.Fatal("Expected violations for unmatched heading")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "Unexpected section") && strings.Contains(v.Message, "Extra") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Expected unmatched heading violation for 'Extra', got: %+v", violations)
	}
}

func TestStructureRuleRegexMatchesLicenseHeading(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# TEST\n\n## Overview\n\n# License\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{
					Pattern: "# [A-Za-z0-9][A-Za-z0-9 _-]*",
					Regex:   true,
				},
				Children: []schema.StructureElement{
					{Heading: schema.HeadingPattern{Pattern: "## Overview"}},
				},
			},
			{
				Heading:  schema.HeadingPattern{Pattern: "# License"},
				Optional: true,
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	matches := ctx.Tree.GetByElement("# [A-Za-z0-9][A-Za-z0-9 _-]*")
	if len(matches) != 1 {
		t.Fatalf("Expected regex heading to match 1 root, got %d", len(matches))
	}
	if matches[0].HeadingText() != "TEST" {
		t.Fatalf("Unexpected regex match: %q", matches[0].HeadingText())
	}

	rule := NewStructureRule()
	violations := rule.ValidateWithContext(ctx)
	if len(violations) != 0 {
		t.Fatalf("Expected no violations, got: %+v", violations)
	}
}

func TestStructureRuleRootFirstMismatch(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# 1.test\n\n# TEST\n\n## Overview\n\n# License\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{
					Pattern: "# [A-Za-z0-9][A-Za-z0-9 _-]*",
					Regex:   true,
				},
				Children: []schema.StructureElement{
					{Heading: schema.HeadingPattern{Pattern: "## Overview"}},
				},
			},
			{
				Heading:  schema.HeadingPattern{Pattern: "# License"},
				Optional: true,
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewStructureRule()
	violations := rule.ValidateWithContext(ctx)
	if len(violations) == 0 {
		t.Fatal("Expected root-first violation, got none")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "First heading") && strings.Contains(v.Message, "1.test") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Expected root-first violation mentioning '1.test', got: %+v", violations)
	}
}

func TestStructureRuleOrderMatchingSkipsEarlierHeadings(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# a\n\n# 1.test\n\n# TEST\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{Heading: schema.HeadingPattern{Pattern: "# 1.test"}},
			{
				Heading: schema.HeadingPattern{
					Pattern: "# [A-Za-z0-9][A-Za-z0-9 _-]*",
					Regex:   true,
				},
				Children: []schema.StructureElement{
					{Heading: schema.HeadingPattern{Pattern: "## Overview"}},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewStructureRule()
	violations := rule.ValidateWithContext(ctx)
	if len(violations) == 0 {
		t.Fatal("Expected violations for missing Overview")
	}

	foundOverview := false
	for _, v := range violations {
		if strings.Contains(v.Message, "Overview") {
			foundOverview = true
			if strings.Contains(v.Message, "a") {
				t.Fatalf("Expected missing Overview under TEST, got: %+v", v)
			}
		}
	}
	if !foundOverview {
		t.Fatalf("Expected missing Overview violation, got: %+v", violations)
	}
}
