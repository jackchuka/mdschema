package rules

import (
	"strings"
	"testing"

	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/schema"
	"github.com/jackchuka/mdschema/internal/vast"
)

func TestNewTableRule(t *testing.T) {
	rule := NewTableRule()
	if rule == nil {
		t.Fatal("NewTableRule() returned nil")
	}
}

func TestTableRuleName(t *testing.T) {
	rule := NewTableRule()
	if rule.Name() != "table" {
		t.Errorf("Name() = %q, want %q", rule.Name(), "table")
	}
}

func TestTableRuleMinimum(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\n| A | B |\n|---|---|\n| 1 | 2 |\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Requires 2 tables, but only 1 exists
	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				SectionRules: &schema.SectionRules{
					Tables: []schema.TableRule{
						{Min: 2},
					},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewTableRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Should detect missing tables")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "2") && strings.Contains(v.Message, "1") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Violation should mention required and found count")
	}
}

func TestTableRuleMaximum(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\n| A |\n|---|\n| 1 |\n\n| B |\n|---|\n| 2 |\n\n| C |\n|---|\n| 3 |\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Max 2 tables, but 3 exist
	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				SectionRules: &schema.SectionRules{
					Tables: []schema.TableRule{
						{Max: 2},
					},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewTableRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Should detect too many tables")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "too many") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Violation should mention too many tables")
	}
}

func TestTableRuleMinColumns(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\n| A |\n|---|\n| 1 |\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Requires minimum 3 columns, but only 1 exists
	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				SectionRules: &schema.SectionRules{
					Tables: []schema.TableRule{
						{MinColumns: 3},
					},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewTableRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Should detect too few columns")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "columns") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Violation should mention columns")
	}
}

func TestTableRuleRequiredHeaders(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\n| Name | Age |\n|---|---|\n| John | 30 |\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Requires "Name" and "Description" headers
	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				SectionRules: &schema.SectionRules{
					Tables: []schema.TableRule{
						{RequiredHeaders: []string{"Name", "Description"}},
					},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewTableRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Should detect missing required header")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "Description") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Violation should mention missing header")
	}
}

func TestTableRuleSufficient(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\n| Name | Description |\n|---|---|\n| A | B |\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Requirements met
	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				SectionRules: &schema.SectionRules{
					Tables: []schema.TableRule{
						{Min: 1, MinColumns: 2, RequiredHeaders: []string{"Name", "Description"}},
					},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewTableRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Should have no violations when requirements met, got %d: %v", len(violations), violations)
	}
}

func TestTableRuleGenerateContent(t *testing.T) {
	rule := NewTableRule()
	var builder strings.Builder

	element := schema.StructureElement{
		Heading: schema.HeadingPattern{Pattern: "## Table Section"},
		SectionRules: &schema.SectionRules{
			Tables: []schema.TableRule{
				{Min: 1, RequiredHeaders: []string{"Name", "Value"}},
			},
		},
	}

	result := rule.GenerateContent(&builder, element)

	if !result {
		t.Error("GenerateContent() should return true when table rules exist")
	}

	content := builder.String()
	if !strings.Contains(content, "| Name | Value |") {
		t.Error("Should generate table with required headers")
	}
}

func TestTableRuleGenerateContentNoRules(t *testing.T) {
	rule := NewTableRule()
	var builder strings.Builder

	element := schema.StructureElement{
		Heading: schema.HeadingPattern{Pattern: "## Section"},
	}

	result := rule.GenerateContent(&builder, element)

	if result {
		t.Error("GenerateContent() should return false when no table rules")
	}
}
