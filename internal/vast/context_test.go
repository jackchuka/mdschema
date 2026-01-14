package vast

import (
	"testing"

	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/schema"
)

func TestDuplicateHeadingSlugs(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte(`# Title

First section

# Title

Second section

# Title

Third section
`))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{}
	ctx := NewContext(doc, s)

	// First occurrence should be at "title"
	if !ctx.HasSlug("title") {
		t.Error("Expected slug 'title' to exist")
	}

	// Second occurrence should be at "title-1"
	if !ctx.HasSlug("title-1") {
		t.Error("Expected slug 'title-1' to exist")
	}

	// Third occurrence should be at "title-2"
	if !ctx.HasSlug("title-2") {
		t.Error("Expected slug 'title-2' to exist")
	}

	// Non-existent slug should not exist
	if ctx.HasSlug("title-3") {
		t.Error("Slug 'title-3' should not exist")
	}
}

func TestSlugCollisionWithLiteralSuffix(t *testing.T) {
	p := parser.New()
	// "Title-1" is a literal heading, not a duplicate
	// Second "Title" should get "title-2", not "title-1"
	doc, err := p.Parse("test.md", []byte(`# Title

# Title-1

# Title
`))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{}
	ctx := NewContext(doc, s)

	// First "Title" → "title"
	if !ctx.HasSlug("title") {
		t.Error("Expected slug 'title' to exist")
	}

	// Literal "Title-1" → "title-1"
	if !ctx.HasSlug("title-1") {
		t.Error("Expected slug 'title-1' to exist")
	}

	// Second "Title" should skip "title-1" and use "title-2"
	if !ctx.HasSlug("title-2") {
		t.Error("Expected slug 'title-2' to exist (should skip collision with literal 'title-1')")
	}
}

func TestMixedDuplicateSlugs(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte(`# Introduction

# Features

# Introduction

# Usage

# Features
`))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{}
	ctx := NewContext(doc, s)

	// Check all expected slugs exist
	expectedSlugs := []string{
		"introduction",   // First Introduction
		"features",       // First Features
		"introduction-1", // Second Introduction
		"usage",          // Usage
		"features-1",     // Second Features
	}

	for _, slug := range expectedSlugs {
		if !ctx.HasSlug(slug) {
			t.Errorf("Expected slug '%s' to exist", slug)
		}
	}
}
