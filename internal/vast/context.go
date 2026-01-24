package vast

import (
	"fmt"

	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/schema"
)

// Context provides all validation state with the validation-ready AST.
type Context struct {
	// The validation-ready AST
	Tree *Tree

	// Schema reference (Tree.Document provides document access)
	Schema *schema.Schema

	// RootDir is the root directory for resolving absolute paths (e.g., /path links).
	// Typically the directory containing .mdschema.yml.
	RootDir string

	// Pre-computed indexes for fast lookups
	slugIndex map[string]*parser.Section // For internal link validation
}

// NewContext creates a new validation context with VAST.
func NewContext(doc *parser.Document, s *schema.Schema) *Context {
	return NewContextWithRoot(doc, s, "")
}

// NewContextWithRoot creates a new validation context with VAST and a root directory.
// The rootDir is used for resolving absolute paths (e.g., /path links).
func NewContextWithRoot(doc *parser.Document, s *schema.Schema, rootDir string) *Context {
	builder := NewBuilder()

	ctx := &Context{
		Tree:      builder.Build(doc, s),
		Schema:    s,
		RootDir:   rootDir,
		slugIndex: make(map[string]*parser.Section),
	}

	// Build slug index for link validation
	// Handle duplicates with global uniqueness (avoid collisions with literal "-N" suffixes)
	for _, section := range doc.GetSections() {
		if section.Heading != nil && section.Heading.Slug != "" {
			slug := section.Heading.Slug

			// If slug already exists, find a unique one
			if _, exists := ctx.slugIndex[slug]; exists {
				counter := 1
				for {
					candidate := fmt.Sprintf("%s-%d", section.Heading.Slug, counter)
					if _, exists := ctx.slugIndex[candidate]; !exists {
						slug = candidate
						break
					}
					counter++
				}
			}

			ctx.slugIndex[slug] = section
		}
	}

	return ctx
}

// HasSlug checks if an internal anchor exists.
func (c *Context) HasSlug(slug string) bool {
	_, ok := c.slugIndex[slug]
	return ok
}
