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

	ctx := vast.NewContext(doc, s, "")
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

	// Frontmatter required but missing (Optional: false is the default)
	s := &schema.Schema{
		Frontmatter: &schema.FrontmatterConfig{
			// Optional: false is default, meaning frontmatter is required
		},
	}

	ctx := vast.NewContext(doc, s, "")
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

	// Frontmatter required and present (Optional: false is the default)
	s := &schema.Schema{
		Frontmatter: &schema.FrontmatterConfig{
			// Optional: false is default, meaning frontmatter is required
		},
	}

	ctx := vast.NewContext(doc, s, "")
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

	// Required field "date" is missing (fields are required by default)
	s := &schema.Schema{
		Frontmatter: &schema.FrontmatterConfig{
			Fields: []schema.FrontmatterField{
				{Name: "title"}, // required by default
				{Name: "date"},  // required by default
			},
		},
	}

	ctx := vast.NewContext(doc, s, "")
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

	// All required fields present (fields are required by default)
	s := &schema.Schema{
		Frontmatter: &schema.FrontmatterConfig{
			Fields: []schema.FrontmatterField{
				{Name: "title"}, // required by default
				{Name: "date"},  // required by default
			},
		},
	}

	ctx := vast.NewContext(doc, s, "")
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
				{Name: "count", Type: schema.FieldTypeNumber},
			},
		},
	}

	ctx := vast.NewContext(doc, s, "")
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
				{Name: "date", Format: schema.FieldFormatDate},
			},
		},
	}

	ctx := vast.NewContext(doc, s, "")
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
				{Name: "date", Format: schema.FieldFormatDate},
			},
		},
	}

	ctx := vast.NewContext(doc, s, "")
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
				{Name: "tags", Type: schema.FieldTypeArray},
			},
		},
	}

	ctx := vast.NewContext(doc, s, "")
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
				{Name: "title"},                  // required by default
				{Name: "author", Optional: true}, // explicitly optional
			},
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewFrontmatterRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Should have no violations for missing optional field, got %d: %v", len(violations), violations)
	}
}

func TestFrontmatterRuleGenerateDocumentPreamble(t *testing.T) {
	rule := NewFrontmatterRule()
	var builder strings.Builder

	// Test with no frontmatter config
	s := &schema.Schema{}
	result := rule.Generate(&builder, s)
	if result {
		t.Error("GenerateDocumentPreamble() should return false when no frontmatter config")
	}

	// Test with frontmatter config
	builder.Reset()
	s = &schema.Schema{
		Frontmatter: &schema.FrontmatterConfig{
			Fields: []schema.FrontmatterField{
				{Name: "title", Type: schema.FieldTypeString}, // required by default
				{Name: "date", Type: schema.FieldTypeDate},    // required by default
			},
		},
	}
	result = rule.Generate(&builder, s)
	if !result {
		t.Error("GenerateDocumentPreamble() should return true when frontmatter config exists")
	}

	output := builder.String()
	if !strings.Contains(output, "---") {
		t.Error("Output should contain frontmatter delimiters")
	}
	if !strings.Contains(output, "title:") {
		t.Error("Output should contain title field")
	}
	if !strings.Contains(output, "# required") {
		t.Error("Output should mark required fields")
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

func TestFrontmatterRuleNestedFields(t *testing.T) {
	const agentSkills = `---
name: my-skill
metadata:
  author: example-org
  version: "1.0"
---

# Title
`

	tests := []struct {
		name              string
		content           string
		fields            []schema.FrontmatterField
		wantViolation     bool
		wantMessageSubstr string
	}{
		{
			name:    "nested keys present",
			content: agentSkills,
			fields: []schema.FrontmatterField{
				{Name: "name"},
				{Name: "metadata.author"},
				{Name: "metadata.version"},
			},
		},
		{
			name:    "nested key missing",
			content: "---\nname: x\nmetadata:\n  version: \"1.0\"\n---\n\n# T\n",
			fields: []schema.FrontmatterField{
				{Name: "metadata.author"},
			},
			wantViolation:     true,
			wantMessageSubstr: "metadata.author",
		},
		{
			name:    "parent missing for nested key",
			content: "---\nname: x\n---\n\n# T\n",
			fields: []schema.FrontmatterField{
				{Name: "metadata.author"},
			},
			wantViolation:     true,
			wantMessageSubstr: "metadata.author",
		},
		{
			name:    "nested key optional missing",
			content: "---\nname: x\n---\n\n# T\n",
			fields: []schema.FrontmatterField{
				{Name: "metadata.author", Optional: true},
			},
		},
		{
			name:    "nested key wrong type",
			content: "---\nmetadata:\n  version: 1.0\n---\n\n# T\n",
			fields: []schema.FrontmatterField{
				{Name: "metadata.version", Type: schema.FieldTypeString},
			},
			wantViolation:     true,
			wantMessageSubstr: "string",
		},
		{
			name:    "object type on parent",
			content: "---\nmetadata:\n  author: x\n---\n\n# T\n",
			fields: []schema.FrontmatterField{
				{Name: "metadata", Type: schema.FieldTypeObject},
			},
		},
		{
			name:    "object type rejects scalar",
			content: "---\nmetadata: just-a-string\n---\n\n# T\n",
			fields: []schema.FrontmatterField{
				{Name: "metadata", Type: schema.FieldTypeObject},
			},
			wantViolation:     true,
			wantMessageSubstr: "object",
		},
		{
			name:    "scalar parent rejects nested lookup",
			content: "---\nmetadata: just-a-string\n---\n\n# T\n",
			fields: []schema.FrontmatterField{
				{Name: "metadata.author"},
			},
			wantViolation:     true,
			wantMessageSubstr: "metadata.author",
		},
		{
			name:    "format on nested url",
			content: "---\nmetadata:\n  homepage: not-a-url\n---\n\n# T\n",
			fields: []schema.FrontmatterField{
				{Name: "metadata.homepage", Format: schema.FieldFormatURL},
			},
			wantViolation:     true,
			wantMessageSubstr: "URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New()
			doc, err := p.Parse("test.md", []byte(tt.content))
			if err != nil {
				t.Fatalf("Parse() error: %v", err)
			}

			s := &schema.Schema{
				Frontmatter: &schema.FrontmatterConfig{Fields: tt.fields},
			}

			ctx := vast.NewContext(doc, s, "")
			rule := NewFrontmatterRule()
			violations := rule.ValidateWithContext(ctx)

			if tt.wantViolation {
				if len(violations) == 0 {
					t.Fatalf("expected violation, got none")
				}
				found := false
				for _, v := range violations {
					if strings.Contains(v.Message, tt.wantMessageSubstr) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected violation containing %q, got %+v", tt.wantMessageSubstr, violations)
				}
				return
			}
			if len(violations) != 0 {
				t.Errorf("expected no violations, got %+v", violations)
			}
		})
	}
}

func TestSplitFieldPath(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"author", []string{"author"}},
		{"metadata.author", []string{"metadata", "author"}},
		{"a.b.c", []string{"a", "b", "c"}},
		{`weird\.key`, []string{"weird.key"}},
		{`a.b\.c.d`, []string{"a", "b.c", "d"}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := splitFieldPath(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("got %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestFrontmatterRuleGenerateNested(t *testing.T) {
	rule := NewFrontmatterRule()
	var builder strings.Builder

	s := &schema.Schema{
		Frontmatter: &schema.FrontmatterConfig{
			Fields: []schema.FrontmatterField{
				{Name: "name", Type: schema.FieldTypeString},
				{Name: "metadata.author", Type: schema.FieldTypeString},
				{Name: "metadata.version", Type: schema.FieldTypeString},
				{Name: "metadata.homepage", Optional: true, Format: schema.FieldFormatURL},
			},
		},
	}

	if !rule.Generate(&builder, s) {
		t.Fatal("Generate should return true")
	}

	output := builder.String()
	wantSubstrings := []string{
		"---\n",
		"name: \"TODO\" # required",
		"metadata: # required",
		"  author: \"TODO\" # required",
		"  version: \"TODO\" # required",
		"  homepage: https://example.com",
	}
	for _, s := range wantSubstrings {
		if !strings.Contains(output, s) {
			t.Errorf("output missing %q\nGot:\n%s", s, output)
		}
	}
}
