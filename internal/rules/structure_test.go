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

	ctx := vast.NewContext(doc, s, "")
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

	ctx := vast.NewContext(doc, s, "")
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

	ctx := vast.NewContext(doc, s, "")
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

	ctx := vast.NewContext(doc, s, "")
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
	if !strings.Contains(content, "Child1") {
		t.Error("Should mention child elements")
	}
	if !strings.Contains(content, "subsections in order") {
		t.Error("Should contain ordering guidance for children")
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

	ctx := vast.NewContext(doc, s, "")
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

	ctx := vast.NewContext(doc, s, "")
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

	ctx := vast.NewContext(doc, s, "")
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

	ctx := vast.NewContext(doc, s, "")
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

func TestStructureRuleAllowAdditional(t *testing.T) {
	p := parser.New()
	// Document has an extra "## Notes" section under "# Title" that is not in schema
	doc, err := p.Parse("test.md", []byte("# Title\n\n## Overview\n\n## Notes\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading:         schema.HeadingPattern{Pattern: "# Title"},
				AllowAdditional: true, // Allow extra subsections
				Children: []schema.StructureElement{
					{Heading: schema.HeadingPattern{Pattern: "## Overview"}},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewStructureRule()
	violations := rule.ValidateWithContext(ctx)

	// Should NOT report violation for "Notes" since parent allows additional
	for _, v := range violations {
		if strings.Contains(v.Message, "Notes") {
			t.Errorf("Should not report violation for extra section when AllowAdditional is true: %s", v.Message)
		}
	}
}

func TestStructureRuleAllowAdditionalFalse(t *testing.T) {
	p := parser.New()
	// Document has an extra "## Notes" section under "# Title" that is not in schema
	doc, err := p.Parse("test.md", []byte("# Title\n\n## Overview\n\n## Notes\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading:         schema.HeadingPattern{Pattern: "# Title"},
				AllowAdditional: false, // Do not allow extra subsections (default)
				Children: []schema.StructureElement{
					{Heading: schema.HeadingPattern{Pattern: "## Overview"}},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewStructureRule()
	violations := rule.ValidateWithContext(ctx)

	// Should report violation for "Notes" since parent does not allow additional
	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "Notes") && strings.Contains(v.Message, "Unexpected") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected violation for extra section when AllowAdditional is false")
	}
}

func TestStructureRuleAllowAdditionalNested(t *testing.T) {
	p := parser.New()
	// Nested structure with extra sections at different levels
	doc, err := p.Parse("test.md", []byte("# Title\n\n## Overview\n\n### Details\n\n### Extra\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				Children: []schema.StructureElement{
					{
						Heading:         schema.HeadingPattern{Pattern: "## Overview"},
						AllowAdditional: true, // Allow extra subsections under Overview
						Children: []schema.StructureElement{
							{Heading: schema.HeadingPattern{Pattern: "### Details"}},
						},
					},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewStructureRule()
	violations := rule.ValidateWithContext(ctx)

	// Should NOT report violation for "Extra" since parent (Overview) allows additional
	for _, v := range violations {
		if strings.Contains(v.Message, "Extra") {
			t.Errorf("Should not report violation for extra nested section when AllowAdditional is true: %s", v.Message)
		}
	}
}

func TestStructureRuleAllowAdditionalDeeplyNested(t *testing.T) {
	p := parser.New()
	// Extra section with its own children - all should be allowed
	doc, err := p.Parse("test.md", []byte("# Title\n\n## Overview\n\n## Extra\n\n### Extra Child\n\n#### Extra Grandchild\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading:         schema.HeadingPattern{Pattern: "# Title"},
				AllowAdditional: true, // Allow extra subsections
				Children: []schema.StructureElement{
					{Heading: schema.HeadingPattern{Pattern: "## Overview"}},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewStructureRule()
	violations := rule.ValidateWithContext(ctx)

	// Should NOT report any violations - Extra, Extra Child, and Extra Grandchild are all allowed
	if len(violations) > 0 {
		t.Errorf("Should not report violations for descendants of allowed extra sections: %+v", violations)
	}
}

func TestStructureRuleSeverityLevels(t *testing.T) {
	tests := []struct {
		name             string
		markdown         string
		severity         string
		expectedSeverity Severity
	}{
		{
			name:             "default severity is error",
			markdown:         "# Title\n",
			severity:         "",
			expectedSeverity: SeverityError,
		},
		{
			name:             "warning severity",
			markdown:         "# Title\n",
			severity:         "warning",
			expectedSeverity: SeverityWarning,
		},
		{
			name:             "info severity",
			markdown:         "# Title\n",
			severity:         "info",
			expectedSeverity: SeverityInfo,
		},
		{
			name:             "error severity (explicit)",
			markdown:         "# Title\n",
			severity:         "error",
			expectedSeverity: SeverityError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New()
			doc, err := p.Parse("test.md", []byte(tt.markdown))
			if err != nil {
				t.Fatalf("Parse() error: %v", err)
			}

			s := &schema.Schema{
				Structure: []schema.StructureElement{
					{Heading: schema.HeadingPattern{Pattern: "# Title"}},
					{
						Heading:  schema.HeadingPattern{Pattern: "## Changelog"},
						Severity: tt.severity,
					},
				},
			}

			ctx := vast.NewContext(doc, s, "")
			rule := NewStructureRule()
			violations := rule.ValidateWithContext(ctx)

			if len(violations) == 0 {
				t.Fatal("Expected violations for missing Changelog section")
			}

			found := false
			for _, v := range violations {
				if strings.Contains(v.Message, "Changelog") {
					found = true
					if v.Severity != tt.expectedSeverity {
						t.Errorf("Severity = %q, want %q", v.Severity, tt.expectedSeverity)
					}
					break
				}
			}

			if !found {
				t.Error("Expected violation mentioning missing 'Changelog' section")
			}
		})
	}
}

// Tests for expression-based heading matching

func TestStructureRuleExprSlugMatch(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("getting-started.md", []byte("# Getting Started\n\nContent."))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{
					Expr: "slug(filename) == slug(heading)",
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewStructureRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Should match slug-style, got %d violations: %v", len(violations), violations)
	}
}

func TestStructureRuleExprSlugMismatch(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("installation.md", []byte("# Setup Guide\n\nContent."))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{
					Expr: "slug(filename) == slug(heading)",
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewStructureRule()
	violations := rule.ValidateWithContext(ctx)

	// The document should still bind (structure rule doesn't fail on mismatch during binding)
	// But validateFirstHeadingIssues should report the mismatch
	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "expected") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Should detect mismatch between heading and filename")
	}
}

func TestStructureRuleExprTrimPrefix(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("01-getting-started.md", []byte("# Getting Started\n\nContent."))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{
					Expr: "slug(trimPrefix(filename, `^\\d+-`)) == slug(heading)",
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewStructureRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Should match after trimming prefix, got %d violations: %v", len(violations), violations)
	}
}

func TestStructureRuleExprCombinedWithPattern(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("my-project.md", []byte("# My Project\n\n## Features\n\nContent."))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{
					Expr: "slug(filename) == slug(heading)", // Dynamic match for title
				},
				Children: []schema.StructureElement{
					{
						Heading: schema.HeadingPattern{
							Pattern: "## Features", // Static pattern for child
						},
					},
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewStructureRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Should match expr for root and pattern for child, got %d violations: %v", len(violations), violations)
	}
}

func TestStructureRuleExprREADME(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("README.md", []byte("# README\n\nContent."))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{
					Expr: "slug(filename) == slug(heading)",
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewStructureRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Should match README, got %d violations: %v", len(violations), violations)
	}
}

func TestStructureRuleExprExactMatch(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("README.md", []byte("# README\n\nContent."))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{
					Expr: "filename == heading",
				},
			},
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewStructureRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Exact match should pass, got %d violations: %v", len(violations), violations)
	}
}
