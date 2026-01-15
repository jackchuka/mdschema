package rules

import (
	"strings"
	"testing"

	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/schema"
	"github.com/jackchuka/mdschema/internal/vast"
)

func TestNewHeadingRule(t *testing.T) {
	rule := NewHeadingRule()
	if rule == nil {
		t.Fatal("NewHeadingRule() returned nil")
	}
}

func TestHeadingRuleName(t *testing.T) {
	rule := NewHeadingRule()
	if rule.Name() != "heading" {
		t.Errorf("Name() = %q, want %q", rule.Name(), "heading")
	}
}

func TestHeadingRuleNoRulesConfigured(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\n### Skipped\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// No heading rules configured
	s := &schema.Schema{}

	ctx := vast.NewContext(doc, s)
	rule := NewHeadingRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Should have no violations when no heading rules configured, got %d", len(violations))
	}
}

func TestHeadingRuleNoSkipLevels(t *testing.T) {
	p := parser.New()
	// h1 -> h3 skips h2
	doc, err := p.Parse("test.md", []byte("# Title\n\n### Skipped h2\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		HeadingRules: &schema.HeadingRules{
			NoSkipLevels: true,
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewHeadingRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Should detect skipped heading level")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "skipped") || strings.Contains(v.Message, "Skipped") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Violation should mention skipped level")
	}
}

func TestHeadingRuleNoSkipLevelsValid(t *testing.T) {
	p := parser.New()
	// Proper hierarchy: h1 -> h2 -> h3
	doc, err := p.Parse("test.md", []byte("# Title\n\n## Section\n\n### Subsection\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		HeadingRules: &schema.HeadingRules{
			NoSkipLevels: true,
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewHeadingRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Should have no violations for proper heading hierarchy, got %d: %v", len(violations), violations)
	}
}

func TestHeadingRuleNoSkipLevelsDecreasing(t *testing.T) {
	p := parser.New()
	// h1 -> h2 -> h3 -> h2 (going back up is OK)
	doc, err := p.Parse("test.md", []byte("# Title\n\n## Section 1\n\n### Subsection\n\n## Section 2\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		HeadingRules: &schema.HeadingRules{
			NoSkipLevels: true,
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewHeadingRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Should allow decreasing heading levels, got %d: %v", len(violations), violations)
	}
}

func TestHeadingRuleUniqueHeadings(t *testing.T) {
	p := parser.New()
	// Duplicate headings
	doc, err := p.Parse("test.md", []byte("# Title\n\n## Features\n\n## Features\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		HeadingRules: &schema.HeadingRules{
			Unique: true,
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewHeadingRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Should detect duplicate headings")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "Duplicate") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Violation should mention duplicate")
	}
}

func TestHeadingRuleUniqueHeadingsValid(t *testing.T) {
	p := parser.New()
	// All unique headings
	doc, err := p.Parse("test.md", []byte("# Title\n\n## Features\n\n## Installation\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		HeadingRules: &schema.HeadingRules{
			Unique: true,
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewHeadingRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Should have no violations for unique headings, got %d: %v", len(violations), violations)
	}
}

func TestHeadingRuleUniquePerLevel(t *testing.T) {
	p := parser.New()
	// Duplicate h2 headings (same level)
	doc, err := p.Parse("test.md", []byte("# Title\n\n## Features\n\n## Features\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		HeadingRules: &schema.HeadingRules{
			UniquePerLevel: true,
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewHeadingRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Should detect duplicate headings at same level")
	}
}

func TestHeadingRuleUniquePerLevelDifferentLevels(t *testing.T) {
	p := parser.New()
	// Same text at different levels is OK
	doc, err := p.Parse("test.md", []byte("# Overview\n\n## Overview\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		HeadingRules: &schema.HeadingRules{
			UniquePerLevel: true,
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewHeadingRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Should allow same heading text at different levels, got %d: %v", len(violations), violations)
	}
}

func TestHeadingRuleMaxDepth(t *testing.T) {
	p := parser.New()
	// h4 exceeds max depth of 3
	doc, err := p.Parse("test.md", []byte("# Title\n\n## Section\n\n### Subsection\n\n#### Too Deep\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		HeadingRules: &schema.HeadingRules{
			MaxDepth: 3,
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewHeadingRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Should detect heading exceeding max depth")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "exceeds") || strings.Contains(v.Message, "depth") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Violation should mention exceeding depth")
	}
}

func TestHeadingRuleMaxDepthValid(t *testing.T) {
	p := parser.New()
	// All headings within max depth of 3
	doc, err := p.Parse("test.md", []byte("# Title\n\n## Section\n\n### Subsection\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		HeadingRules: &schema.HeadingRules{
			MaxDepth: 3,
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewHeadingRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Should have no violations within max depth, got %d: %v", len(violations), violations)
	}
}

func TestHeadingRuleCombinedRules(t *testing.T) {
	p := parser.New()
	// Multiple violations: skip level, duplicate, exceeds depth
	doc, err := p.Parse("test.md", []byte("# Title\n\n### Skipped\n\n### Skipped\n\n##### Too Deep\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		HeadingRules: &schema.HeadingRules{
			NoSkipLevels: true,
			Unique:       true,
			MaxDepth:     4,
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewHeadingRule()
	violations := rule.ValidateWithContext(ctx)

	// Should have at least 3 violations: skip, duplicate, depth
	if len(violations) < 3 {
		t.Errorf("Should detect multiple violations, got %d: %v", len(violations), violations)
	}
}

func TestHeadingRuleGenerateContent(t *testing.T) {
	rule := NewHeadingRule()
	var builder strings.Builder

	element := schema.StructureElement{
		Heading: schema.HeadingPattern{Pattern: "## Section"},
	}

	result := rule.GenerateContent(&builder, element)

	if result {
		t.Error("GenerateContent() should return false for heading rules")
	}
}
