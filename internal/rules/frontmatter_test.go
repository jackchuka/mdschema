package rules

import (
	"strings"
	"testing"

	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/schema"
	"github.com/jackchuka/mdschema/internal/vast"
)

func TestNewFrontmatterRule(t *testing.T) {
	rule := NewFrontmatterRule()
	if rule == nil {
		t.Fatal("NewFrontmatterRule() returned nil")
	}
}

func TestFrontmatterRuleName(t *testing.T) {
	rule := NewFrontmatterRule()
	if rule.Name() != "frontmatter" {
		t.Errorf("Name() = %q, want %q", rule.Name(), "frontmatter")
	}
}

func TestFrontmatterRuleNoConfig(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// No frontmatter config
	s := &schema.Schema{}

	ctx := vast.NewContext(doc, s)
	rule := NewFrontmatterRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Should have no violations when no frontmatter config, got %d", len(violations))
	}
}

func TestFrontmatterRuleRequiredMissing(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Frontmatter required but missing
	s := &schema.Schema{
		Frontmatter: &schema.FrontmatterConfig{
			Required: true,
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewFrontmatterRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Should detect missing required frontmatter")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "required") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Violation should mention required frontmatter")
	}
}

func TestFrontmatterRuleRequiredPresent(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("---\ntitle: Test\n---\n\n# Title\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Frontmatter required and present
	s := &schema.Schema{
		Frontmatter: &schema.FrontmatterConfig{
			Required: true,
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewFrontmatterRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Should have no violations when frontmatter present, got %d: %v", len(violations), violations)
	}
}

func TestFrontmatterRuleRequiredFieldMissing(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("---\ntitle: Test\n---\n\n# Title\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Required field "date" is missing
	s := &schema.Schema{
		Frontmatter: &schema.FrontmatterConfig{
			Fields: []schema.FrontmatterField{
				{Name: "title", Required: true},
				{Name: "date", Required: true},
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewFrontmatterRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Should detect missing required field")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "date") && strings.Contains(v.Message, "missing") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Violation should mention missing 'date' field")
	}
}

func TestFrontmatterRuleAllFieldsPresent(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("---\ntitle: Test\ndate: 2024-01-15\n---\n\n# Title\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// All required fields present
	s := &schema.Schema{
		Frontmatter: &schema.FrontmatterConfig{
			Fields: []schema.FrontmatterField{
				{Name: "title", Required: true},
				{Name: "date", Required: true},
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewFrontmatterRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Should have no violations when all fields present, got %d: %v", len(violations), violations)
	}
}

func TestFrontmatterRuleTypeValidation(t *testing.T) {
	p := parser.New()
	// "count" should be a number but is a string
	doc, err := p.Parse("test.md", []byte("---\ncount: not-a-number\n---\n\n# Title\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Frontmatter: &schema.FrontmatterConfig{
			Fields: []schema.FrontmatterField{
				{Name: "count", Type: "number"},
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewFrontmatterRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Should detect type mismatch")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "number") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Violation should mention expected type 'number'")
	}
}

func TestFrontmatterRuleDateFormat(t *testing.T) {
	p := parser.New()
	// Invalid date format
	doc, err := p.Parse("test.md", []byte("---\ndate: 01/15/2024\n---\n\n# Title\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Frontmatter: &schema.FrontmatterConfig{
			Fields: []schema.FrontmatterField{
				{Name: "date", Format: "date"},
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewFrontmatterRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Should detect invalid date format")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "YYYY-MM-DD") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Violation should mention YYYY-MM-DD format")
	}
}

func TestFrontmatterRuleValidDateFormat(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("---\ndate: 2024-01-15\n---\n\n# Title\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Frontmatter: &schema.FrontmatterConfig{
			Fields: []schema.FrontmatterField{
				{Name: "date", Format: "date"},
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewFrontmatterRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Should have no violations for valid date, got %d: %v", len(violations), violations)
	}
}

func TestFrontmatterRuleArrayType(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("---\ntags:\n  - go\n  - markdown\n---\n\n# Title\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Frontmatter: &schema.FrontmatterConfig{
			Fields: []schema.FrontmatterField{
				{Name: "tags", Type: "array"},
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewFrontmatterRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Should have no violations for valid array, got %d: %v", len(violations), violations)
	}
}

func TestFrontmatterRuleOptionalField(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("---\ntitle: Test\n---\n\n# Title\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Optional field "author" is missing - should be OK
	s := &schema.Schema{
		Frontmatter: &schema.FrontmatterConfig{
			Fields: []schema.FrontmatterField{
				{Name: "title", Required: true},
				{Name: "author", Required: false},
			},
		},
	}

	ctx := vast.NewContext(doc, s)
	rule := NewFrontmatterRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Should have no violations for missing optional field, got %d: %v", len(violations), violations)
	}
}

func TestFrontmatterRuleGenerateContent(t *testing.T) {
	rule := NewFrontmatterRule()
	var builder strings.Builder

	element := schema.StructureElement{
		Heading: schema.HeadingPattern{Pattern: "## Section"},
	}

	result := rule.GenerateContent(&builder, element)

	if result {
		t.Error("GenerateContent() should return false for frontmatter rules (document-level)")
	}
}

func TestFrontmatterParsing(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("---\ntitle: My Document\nauthor: John Doe\n---\n\n# Title\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if doc.FrontMatter == nil {
		t.Fatal("FrontMatter should be parsed")
	}

	if doc.FrontMatter.Format != "yaml" {
		t.Errorf("Format = %q, want 'yaml'", doc.FrontMatter.Format)
	}

	if doc.FrontMatter.Data == nil {
		t.Fatal("FrontMatter.Data should be parsed")
	}

	if doc.FrontMatter.Data["title"] != "My Document" {
		t.Errorf("title = %v, want 'My Document'", doc.FrontMatter.Data["title"])
	}

	if doc.FrontMatter.Data["author"] != "John Doe" {
		t.Errorf("author = %v, want 'John Doe'", doc.FrontMatter.Data["author"])
	}
}

func TestFrontmatterNoFrontmatter(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\nContent.\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if doc.FrontMatter != nil {
		t.Error("FrontMatter should be nil when not present")
	}
}
