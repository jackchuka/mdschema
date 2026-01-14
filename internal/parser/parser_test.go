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
		{"With   Multiple   Spaces", "with---multiple---spaces"}, // preserves multiple dashes
		{"Special!@#Characters", "special!@#characters"},         // preserves special chars
		{"Numbers 123", "numbers-123"},
		{"", ""},
		{"Already-Slugged", "already-slugged"},
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
