package infer

import (
	"reflect"
	"testing"

	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/schema"
)

func TestFromDocumentBuildsStructure(t *testing.T) {
	content := `# Project Title

Intro text

## Overview
Context here

### Background
Details

## Usage

### Install
Steps

### Run
Steps
`

	p := parser.New()
	doc, err := p.Parse("test.md", []byte(content))
	if err != nil {
		t.Fatalf("parse markdown: %v", err)
	}

	got, err := FromDocument(doc)
	if err != nil {
		t.Fatalf("FromDocument returned error: %v", err)
	}

	want := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: "# Project Title",
				Children: []schema.StructureElement{
					{
						Heading: "## Overview",
						Children: []schema.StructureElement{
							{Heading: "### Background"},
						},
					},
					{
						Heading: "## Usage",
						Children: []schema.StructureElement{
							{Heading: "### Install"},
							{Heading: "### Run"},
						},
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FromDocument mismatch\nwant: %#v\n got: %#v", want, got)
	}
}

func TestFromDocumentRequiresHeadings(t *testing.T) {
	content := "plain text with no headings"

	p := parser.New()
	doc, err := p.Parse("test.md", []byte(content))
	if err != nil {
		t.Fatalf("parse markdown: %v", err)
	}

	if _, err := FromDocument(doc); err == nil {
		t.Fatal("expected error when document has no headings, got nil")
	}
}

func TestFromDocumentSkipsFrontMatter(t *testing.T) {
	content := `---
title: "Example"
author: "Doc Writer"
---

# First

Body

## Child (Optional)

More
`

	p := parser.New()
	doc, err := p.Parse("test.md", []byte(content))
	if err != nil {
		t.Fatalf("parse markdown: %v", err)
	}

	got, err := FromDocument(doc)
	if err != nil {
		t.Fatalf("FromDocument returned error: %v", err)
	}

	if len(got.Structure) == 0 {
		t.Fatalf("expected structure elements, got none")
	}

	if got.Structure[0].Heading != "# First" {
		t.Fatalf("expected first heading '# First', got %q", got.Structure[0].Heading)
	}

	if len(got.Structure[0].Children) != 1 || got.Structure[0].Children[0].Heading != "## Child (Optional)" {
		t.Fatalf("expected child heading '## Child (Optional)', got %#v", got.Structure[0].Children)
	}
}
