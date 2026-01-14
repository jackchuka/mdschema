package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
	if p.md == nil {
		t.Fatal("New() returned parser with nil goldmark instance")
	}
}

func TestParseBasicDocument(t *testing.T) {
	p := New()
	content := []byte(`# Title

Some content here.

## Section One

More content.
`)
	doc, err := p.Parse("test.md", content)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if doc.Path != "test.md" {
		t.Errorf("doc.Path = %q, want %q", doc.Path, "test.md")
	}

	sections := doc.GetSections()
	if len(sections) != 2 {
		t.Fatalf("GetSections() returned %d sections, want 2", len(sections))
	}

	if sections[0].Heading.Text != "Title" {
		t.Errorf("sections[0].Heading.Text = %q, want %q", sections[0].Heading.Text, "Title")
	}
	if sections[0].Heading.Level != 1 {
		t.Errorf("sections[0].Heading.Level = %d, want 1", sections[0].Heading.Level)
	}

	if sections[1].Heading.Text != "Section One" {
		t.Errorf("sections[1].Heading.Text = %q, want %q", sections[1].Heading.Text, "Section One")
	}
	if sections[1].Heading.Level != 2 {
		t.Errorf("sections[1].Heading.Level = %d, want 2", sections[1].Heading.Level)
	}
}

func TestParseHeadingLevels(t *testing.T) {
	p := New()
	content := []byte(`# H1
## H2
### H3
#### H4
##### H5
###### H6
`)
	doc, err := p.Parse("test.md", content)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	sections := doc.GetSections()
	if len(sections) != 6 {
		t.Fatalf("GetSections() returned %d sections, want 6", len(sections))
	}

	expectedLevels := []int{1, 2, 3, 4, 5, 6}
	expectedTexts := []string{"H1", "H2", "H3", "H4", "H5", "H6"}

	for i, section := range sections {
		if section.Heading.Level != expectedLevels[i] {
			t.Errorf("section[%d].Heading.Level = %d, want %d", i, section.Heading.Level, expectedLevels[i])
		}
		if section.Heading.Text != expectedTexts[i] {
			t.Errorf("section[%d].Heading.Text = %q, want %q", i, section.Heading.Text, expectedTexts[i])
		}
	}
}

func TestParseCodeBlocks(t *testing.T) {
	p := New()
	content := []byte("# Title\n\n```go\nfunc main() {}\n```\n\n```bash\necho hello\n```\n")

	doc, err := p.Parse("test.md", content)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	sections := doc.GetSections()
	if len(sections) != 1 {
		t.Fatalf("GetSections() returned %d sections, want 1", len(sections))
	}

	codeBlocks := sections[0].CodeBlocks
	if len(codeBlocks) != 2 {
		t.Fatalf("section has %d code blocks, want 2", len(codeBlocks))
	}

	if codeBlocks[0].Lang != "go" {
		t.Errorf("codeBlocks[0].Lang = %q, want %q", codeBlocks[0].Lang, "go")
	}
	if codeBlocks[1].Lang != "bash" {
		t.Errorf("codeBlocks[1].Lang = %q, want %q", codeBlocks[1].Lang, "bash")
	}
}

func TestParseNestedSections(t *testing.T) {
	p := New()
	content := []byte(`# Root

## Child 1

### Grandchild 1.1

## Child 2

### Grandchild 2.1

### Grandchild 2.2
`)
	doc, err := p.Parse("test.md", content)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// The root section should have children
	rootSection := doc.Root.Children[0]
	if rootSection.Heading.Text != "Root" {
		t.Errorf("root section text = %q, want %q", rootSection.Heading.Text, "Root")
	}

	// Root should have 2 direct children (Child 1 and Child 2)
	if len(rootSection.Children) != 2 {
		t.Fatalf("root has %d children, want 2", len(rootSection.Children))
	}

	child1 := rootSection.Children[0]
	if child1.Heading.Text != "Child 1" {
		t.Errorf("child1 text = %q, want %q", child1.Heading.Text, "Child 1")
	}
	if len(child1.Children) != 1 {
		t.Fatalf("child1 has %d children, want 1", len(child1.Children))
	}

	child2 := rootSection.Children[1]
	if child2.Heading.Text != "Child 2" {
		t.Errorf("child2 text = %q, want %q", child2.Heading.Text, "Child 2")
	}
	if len(child2.Children) != 2 {
		t.Fatalf("child2 has %d children, want 2", len(child2.Children))
	}
}

func TestParseEmptyDocument(t *testing.T) {
	p := New()
	doc, err := p.Parse("test.md", []byte(""))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	sections := doc.GetSections()
	if len(sections) != 0 {
		t.Errorf("GetSections() returned %d sections, want 0", len(sections))
	}
}

func TestParseDocumentWithOnlyText(t *testing.T) {
	p := New()
	content := []byte("Just some text without any headings.\n\nMore text here.")
	doc, err := p.Parse("test.md", content)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	sections := doc.GetSections()
	if len(sections) != 0 {
		t.Errorf("GetSections() returned %d sections, want 0", len(sections))
	}
}

func TestParseTables(t *testing.T) {
	p := New()
	content := []byte(`# Table Test

| Header 1 | Header 2 |
| -------- | -------- |
| Cell 1   | Cell 2   |
`)
	doc, err := p.Parse("test.md", content)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	sections := doc.GetSections()
	if len(sections) != 1 {
		t.Fatalf("GetSections() returned %d sections, want 1", len(sections))
	}

	tables := sections[0].Tables
	if len(tables) != 1 {
		t.Fatalf("section has %d tables, want 1", len(tables))
	}

	if len(tables[0].Headers) != 2 {
		t.Errorf("table has %d headers, want 2", len(tables[0].Headers))
	}
}

func TestParseFile(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.md")

	content := []byte("# Test File\n\nContent here.")
	if err := os.WriteFile(tmpFile, content, 0o644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	p := New()
	doc, err := p.ParseFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseFile() error: %v", err)
	}

	if doc.Path != tmpFile {
		t.Errorf("doc.Path = %q, want %q", doc.Path, tmpFile)
	}

	sections := doc.GetSections()
	if len(sections) != 1 {
		t.Fatalf("GetSections() returned %d sections, want 1", len(sections))
	}

	if sections[0].Heading.Text != "Test File" {
		t.Errorf("heading text = %q, want %q", sections[0].Heading.Text, "Test File")
	}
}

func TestParseFileNotFound(t *testing.T) {
	p := New()
	_, err := p.ParseFile("/nonexistent/path/file.md")
	if err == nil {
		t.Error("ParseFile() with nonexistent file should return error")
	}
}

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Simple Title", "simple-title"},
		{"UPPERCASE", "uppercase"},
		{"With   Multiple   Spaces", "with-multiple-spaces"}, // collapses multiple hyphens (GitHub-like)
		{"Special!@#Characters", "specialcharacters"},        // removes special chars (GitHub-like)
		{"Numbers 123", "numbers-123"},
		{"", ""},
		{"Already-Slugged", "already-slugged"},
		{"Hello, World!", "hello-world"},           // punctuation removed
		{"foo_bar_baz", "foo_bar_baz"},             // underscores preserved
		{"API Reference (v2)", "api-reference-v2"}, // parens removed
	}

	for _, tc := range tests {
		got := generateSlug(tc.input)
		if got != tc.want {
			t.Errorf("generateSlug(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestIsInternalLink(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"#section", true},
		{"./relative/path.md", true},
		{"../parent/path.md", true},
		{"path/to/file.md", true},
		{"https://example.com", false},
		{"http://example.com", false},
		{"//example.com", true},           // only http/https are checked as external
		{"mailto:test@example.com", true}, // only http/https are checked as external
	}

	for _, tc := range tests {
		got := isInternalLink(tc.url)
		if got != tc.want {
			t.Errorf("isInternalLink(%q) = %v, want %v", tc.url, got, tc.want)
		}
	}
}

func TestCalculateLineColumn(t *testing.T) {
	content := []byte("line1\nline2\nline3")

	tests := []struct {
		offset   int
		wantLine int
		wantCol  int
	}{
		{0, 1, 1},  // start of file
		{5, 1, 6},  // end of line1
		{6, 2, 1},  // start of line2
		{11, 2, 6}, // end of line2
		{12, 3, 1}, // start of line3
		{16, 3, 5}, // end of line3
	}

	for _, tc := range tests {
		line, col := calculateLineColumn(content, tc.offset)
		if line != tc.wantLine || col != tc.wantCol {
			t.Errorf("calculateLineColumn(content, %d) = (%d, %d), want (%d, %d)",
				tc.offset, line, col, tc.wantLine, tc.wantCol)
		}
	}
}

func TestSectionParentChild(t *testing.T) {
	p := New()
	content := []byte(`# Parent

## Child
`)
	doc, err := p.Parse("test.md", content)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	sections := doc.GetSections()
	if len(sections) != 2 {
		t.Fatalf("GetSections() returned %d sections, want 2", len(sections))
	}

	parent := sections[0]
	child := sections[1]

	if child.Parent != parent {
		t.Error("child.Parent should point to parent section")
	}

	if len(parent.Children) != 1 {
		t.Fatalf("parent has %d children, want 1", len(parent.Children))
	}

	if parent.Children[0] != child {
		t.Error("parent.Children[0] should be the child section")
	}
}

func TestHeadingSlug(t *testing.T) {
	p := New()
	content := []byte(`# My Title

## Another Section
`)
	doc, err := p.Parse("test.md", content)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	sections := doc.GetSections()
	if sections[0].Heading.Slug != "my-title" {
		t.Errorf("slug = %q, want %q", sections[0].Heading.Slug, "my-title")
	}
	if sections[1].Heading.Slug != "another-section" {
		t.Errorf("slug = %q, want %q", sections[1].Heading.Slug, "another-section")
	}
}

func TestHeadingWithInlineFormatting(t *testing.T) {
	p := New()
	content := []byte(`# **Bold** Heading

## Heading with ` + "`code`" + `

### Heading with *emphasis*

#### Link to [something](url) here
`)
	doc, err := p.Parse("test.md", content)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	sections := doc.GetSections()
	if len(sections) != 4 {
		t.Fatalf("GetSections() returned %d sections, want 4", len(sections))
	}

	tests := []struct {
		index    int
		wantText string
		wantSlug string
	}{
		{0, "Bold Heading", "bold-heading"},
		{1, "Heading with code", "heading-with-code"},
		{2, "Heading with emphasis", "heading-with-emphasis"},
		{3, "Link to something here", "link-to-something-here"},
	}

	for _, tc := range tests {
		if sections[tc.index].Heading.Text != tc.wantText {
			t.Errorf("sections[%d].Heading.Text = %q, want %q", tc.index, sections[tc.index].Heading.Text, tc.wantText)
		}
		if sections[tc.index].Heading.Slug != tc.wantSlug {
			t.Errorf("sections[%d].Heading.Slug = %q, want %q", tc.index, sections[tc.index].Heading.Slug, tc.wantSlug)
		}
	}
}

// TestSiblingContentNotCaptured verifies that elements from a sibling section
// are not captured by a parent or its leaf child (tests top-down boundary fix)
func TestSiblingContentNotCaptured(t *testing.T) {
	p := New()
	// Structure: Section A has child A1 (leaf), Section B is sibling to A with a code block
	// Neither Section A nor its child A1 should capture B's code block
	content := []byte(`# Section A

## Section A1

Content in A1

# Section B

` + "```go" + `
func main() {}
` + "```" + `
`)
	doc, err := p.Parse("test.md", content)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Find sections
	var sectionA, sectionA1, sectionB *Section
	for _, child := range doc.Root.Children {
		if child.Heading.Text == "Section A" {
			sectionA = child
			for _, grandchild := range child.Children {
				if grandchild.Heading.Text == "Section A1" {
					sectionA1 = grandchild
				}
			}
		}
		if child.Heading.Text == "Section B" {
			sectionB = child
		}
	}

	if sectionA == nil || sectionA1 == nil || sectionB == nil {
		t.Fatal("Could not find Section A, A1, or B")
	}

	// Verify Section A has no code blocks
	if len(sectionA.CodeBlocks) != 0 {
		t.Errorf("Section A should have 0 code blocks, got %d", len(sectionA.CodeBlocks))
	}

	// Verify Section A1 (leaf child) has no code blocks - this is the key test
	if len(sectionA1.CodeBlocks) != 0 {
		t.Errorf("Section A1 should have 0 code blocks, got %d (leaf child captured parent's sibling content)", len(sectionA1.CodeBlocks))
	}

	// Verify Section B has the code block
	if len(sectionB.CodeBlocks) != 1 {
		t.Errorf("Section B should have 1 code block, got %d", len(sectionB.CodeBlocks))
	}

	// Verify boundaries: A1 ends before B starts
	if sectionA1.EndLine >= sectionB.StartLine {
		t.Errorf("Section A1 EndLine (%d) should be less than Section B StartLine (%d)",
			sectionA1.EndLine, sectionB.StartLine)
	}

	// Verify boundaries: A ends before B starts
	if sectionA.EndLine >= sectionB.StartLine {
		t.Errorf("Section A EndLine (%d) should be less than Section B StartLine (%d)",
			sectionA.EndLine, sectionB.StartLine)
	}
}

// TestIntroContentAttachedToRoot verifies that content before the first heading
// is attached to the root section (Medium #1 fix)
func TestIntroContentAttachedToRoot(t *testing.T) {
	p := New()
	content := []byte(`This is intro content with a [link](https://example.com).

` + "```bash" + `
echo "intro code block"
` + "```" + `

# First Heading

Content after heading.
`)
	doc, err := p.Parse("test.md", content)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Root section should have the intro link and code block
	if len(doc.Root.Links) != 1 {
		t.Errorf("Root section should have 1 link (intro), got %d", len(doc.Root.Links))
	}
	if len(doc.Root.CodeBlocks) != 1 {
		t.Errorf("Root section should have 1 code block (intro), got %d", len(doc.Root.CodeBlocks))
	}

	// The first heading section should not have the intro elements
	if len(doc.Root.Children) < 1 {
		t.Fatal("Expected at least 1 child section")
	}
	firstSection := doc.Root.Children[0]
	if len(firstSection.Links) != 0 {
		t.Errorf("First heading section should have 0 links, got %d", len(firstSection.Links))
	}
	if len(firstSection.CodeBlocks) != 0 {
		t.Errorf("First heading section should have 0 code blocks, got %d", len(firstSection.CodeBlocks))
	}
}
