package parser

import "testing"

func TestParagraphsTopLevelOnly(t *testing.T) {
	src := []byte(`# Title

## Overview

First prose paragraph.

Second prose paragraph.

- a bullet

- another bullet

> a blockquote paragraph
`)
	p := New()
	doc, err := p.Parse("test.md", src)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	var overview *Section
	for _, s := range doc.GetSections() {
		if s.Heading != nil && s.Heading.Text == "Overview" {
			overview = s
		}
	}
	if overview == nil {
		t.Fatal("Overview section not found")
	}

	// Two prose paragraphs; the list and blockquote paragraphs must NOT count.
	if got := len(overview.Paragraphs); got != 2 {
		t.Errorf("len(Paragraphs) = %d, want 2", got)
	}
}

func TestParagraphsEmptySection(t *testing.T) {
	src := []byte("# Title\n\n## Empty\n\n## Next\n\ntext\n")
	p := New()
	doc, err := p.Parse("test.md", src)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	for _, s := range doc.GetSections() {
		if s.Heading != nil && s.Heading.Text == "Empty" {
			if got := len(s.Paragraphs); got != 0 {
				t.Errorf("Empty section len(Paragraphs) = %d, want 0", got)
			}
		}
	}
}
