package rules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/schema"
	"github.com/jackchuka/mdschema/internal/vast"
)

func TestNewLinkValidationRule(t *testing.T) {
	rule := NewLinkValidationRule()
	if rule == nil {
		t.Fatal("NewLinkValidationRule() returned nil")
	}
}

func TestLinkValidationRuleName(t *testing.T) {
	rule := NewLinkValidationRule()
	if rule.Name() != "link" {
		t.Errorf("Name() = %q, want %q", rule.Name(), "link")
	}
}

func TestLinkValidationNoRulesConfigured(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\n[link](#section)\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Schema with no link rules
	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{Heading: schema.HeadingPattern{Pattern: "# Title"}},
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewLinkValidationRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Expected no violations when link rules not configured, got %d", len(violations))
	}
}

func TestLinkValidationValidAnchor(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\n## Section\n\n[link](#section)\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{Heading: schema.HeadingPattern{Pattern: "# Title"}},
		},
		Links: &schema.LinkRule{
			ValidateInternal: true,
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewLinkValidationRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Expected no violations for valid anchor, got %d:", len(violations))
		for _, v := range violations {
			t.Logf("  - %s", v.Message)
		}
	}
}

func TestLinkValidationBrokenAnchor(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\n[link](#nonexistent)\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{Heading: schema.HeadingPattern{Pattern: "# Title"}},
		},
		Links: &schema.LinkRule{
			ValidateInternal: true,
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewLinkValidationRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Expected violation for broken anchor, got none")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "nonexistent") && strings.Contains(v.Message, "does not exist") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected violation mentioning broken anchor")
	}
}

func TestLinkValidationValidFileLink(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a target file
	targetFile := filepath.Join(tmpDir, "other.md")
	if err := os.WriteFile(targetFile, []byte("# Other"), 0o644); err != nil {
		t.Fatalf("Failed to create target file: %v", err)
	}

	// Create the main document
	mainFile := filepath.Join(tmpDir, "main.md")
	content := []byte("# Title\n\n[link](./other.md)\n")
	if err := os.WriteFile(mainFile, content, 0o644); err != nil {
		t.Fatalf("Failed to create main file: %v", err)
	}

	p := parser.New()
	doc, err := p.ParseFile(mainFile)
	if err != nil {
		t.Fatalf("ParseFile() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{Heading: schema.HeadingPattern{Pattern: "# Title"}},
		},
		Links: &schema.LinkRule{
			ValidateFiles: true,
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewLinkValidationRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Expected no violations for valid file link, got %d:", len(violations))
		for _, v := range violations {
			t.Logf("  - %s", v.Message)
		}
	}
}

func TestLinkValidationBrokenFileLink(t *testing.T) {
	tmpDir := t.TempDir()

	// Create only the main document (no target file)
	mainFile := filepath.Join(tmpDir, "main.md")
	content := []byte("# Title\n\n[link](./missing.md)\n")
	if err := os.WriteFile(mainFile, content, 0o644); err != nil {
		t.Fatalf("Failed to create main file: %v", err)
	}

	p := parser.New()
	doc, err := p.ParseFile(mainFile)
	if err != nil {
		t.Fatalf("ParseFile() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{Heading: schema.HeadingPattern{Pattern: "# Title"}},
		},
		Links: &schema.LinkRule{
			ValidateFiles: true,
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewLinkValidationRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Expected violation for broken file link, got none")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "missing.md") && strings.Contains(v.Message, "does not exist") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected violation mentioning broken file link")
	}
}

func TestLinkValidationBlockedDomain(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\n[link](https://blocked.com/page)\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{Heading: schema.HeadingPattern{Pattern: "# Title"}},
		},
		Links: &schema.LinkRule{
			BlockedDomains: []string{"blocked.com"},
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewLinkValidationRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Expected violation for blocked domain, got none")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "blocked domain") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected violation mentioning blocked domain")
	}
}

func TestLinkValidationAllowedDomains(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\n[link](https://notallowed.com/page)\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{Heading: schema.HeadingPattern{Pattern: "# Title"}},
		},
		Links: &schema.LinkRule{
			AllowedDomains: []string{"github.com", "example.com"},
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewLinkValidationRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Expected violation for domain not in allowed list, got none")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "not in the allowed domains list") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected violation mentioning domain not in allowed list")
	}
}

func TestLinkValidationAllowedDomainsPass(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\n[link](https://github.com/user/repo)\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{Heading: schema.HeadingPattern{Pattern: "# Title"}},
		},
		Links: &schema.LinkRule{
			AllowedDomains: []string{"github.com", "example.com"},
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewLinkValidationRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Expected no violations for allowed domain, got %d:", len(violations))
		for _, v := range violations {
			t.Logf("  - %s", v.Message)
		}
	}
}

func TestLinkValidationSubdomainBlocked(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\n[link](https://sub.blocked.com/page)\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{Heading: schema.HeadingPattern{Pattern: "# Title"}},
		},
		Links: &schema.LinkRule{
			BlockedDomains: []string{"blocked.com"},
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewLinkValidationRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Expected violation for subdomain of blocked domain, got none")
	}
}

func TestLinkValidationFileLinkWithAnchor(t *testing.T) {
	tmpDir := t.TempDir()

	// Create target file
	targetFile := filepath.Join(tmpDir, "other.md")
	if err := os.WriteFile(targetFile, []byte("# Other"), 0o644); err != nil {
		t.Fatalf("Failed to create target file: %v", err)
	}

	// Create main document with file link containing anchor
	mainFile := filepath.Join(tmpDir, "main.md")
	content := []byte("# Title\n\n[link](./other.md#section)\n")
	if err := os.WriteFile(mainFile, content, 0o644); err != nil {
		t.Fatalf("Failed to create main file: %v", err)
	}

	p := parser.New()
	doc, err := p.ParseFile(mainFile)
	if err != nil {
		t.Fatalf("ParseFile() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{Heading: schema.HeadingPattern{Pattern: "# Title"}},
		},
		Links: &schema.LinkRule{
			ValidateFiles: true,
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewLinkValidationRule()
	violations := rule.ValidateWithContext(ctx)

	// Should pass because the file exists (we don't validate anchors in other files)
	if len(violations) != 0 {
		t.Errorf("Expected no violations for valid file link with anchor, got %d:", len(violations))
		for _, v := range violations {
			t.Logf("  - %s", v.Message)
		}
	}
}

func TestLinkValidationInternalDisabled(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("# Title\n\n[link](#nonexistent)\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{Heading: schema.HeadingPattern{Pattern: "# Title"}},
		},
		Links: &schema.LinkRule{
			ValidateInternal: false, // Explicitly disabled
			ValidateFiles:    true,
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewLinkValidationRule()
	violations := rule.ValidateWithContext(ctx)

	// Should not report anchor errors when ValidateInternal is false
	for _, v := range violations {
		if strings.Contains(v.Message, "anchor") {
			t.Errorf("Should not validate anchors when ValidateInternal is false: %s", v.Message)
		}
	}
}

func TestLinkValidationRootRelativePath(t *testing.T) {
	// Create directory structure:
	// tmpDir/
	//   depth1/folder1/  <- target directory at "root"
	//   docs/
	//     README.md        <- document with root-relative link
	tmpDir := t.TempDir()

	// Create target directory at root level
	targetDir := filepath.Join(tmpDir, "depth1", "folder1")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}

	// Create docs directory
	docsDir := filepath.Join(tmpDir, "docs")
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		t.Fatalf("Failed to create docs directory: %v", err)
	}

	// Create document with root-relative link (starts with /)
	mainFile := filepath.Join(docsDir, "README.md")
	content := []byte("# Title\n\n[folder1](/depth1/folder1/)\n")
	if err := os.WriteFile(mainFile, content, 0o644); err != nil {
		t.Fatalf("Failed to create main file: %v", err)
	}

	p := parser.New()
	doc, err := p.ParseFile(mainFile)
	if err != nil {
		t.Fatalf("ParseFile() error: %v", err)
	}

	s := &schema.Schema{
		Links: &schema.LinkRule{
			ValidateFiles: true,
		},
	}

	// Create context with rootDir set to tmpDir (simulating schema location)
	ctx := vast.NewContext(doc, s, tmpDir)
	rule := NewLinkValidationRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Expected no violations for valid root-relative link, got %d:", len(violations))
		for _, v := range violations {
			t.Logf("  - %s at line %d", v.Message, v.Line)
		}
	}
}

func TestLinkValidationFrontmatterBlockedDomain(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("---\nrepo: https://blocked.com/repo\n---\n\n# Title\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Links: &schema.LinkRule{
			BlockedDomains: []string{"blocked.com"},
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewLinkValidationRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Expected violation for frontmatter URL to blocked domain")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "blocked domain") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected violation mentioning blocked domain, got: %v", violations)
	}
}

func TestLinkValidationFrontmatterAllowedDomains(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("---\nrepo: https://notallowed.com/repo\n---\n\n# Title\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Links: &schema.LinkRule{
			AllowedDomains: []string{"github.com"},
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewLinkValidationRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Expected violation for frontmatter URL not in allowed domains")
	}
}

func TestLinkValidationFrontmatterAllowedDomainPass(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("---\nrepo: https://github.com/user/repo\n---\n\n# Title\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Links: &schema.LinkRule{
			AllowedDomains: []string{"github.com"},
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewLinkValidationRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Expected no violations for allowed domain, got %d: %v", len(violations), violations)
	}
}

func TestLinkValidationFrontmatterNonStringSkipped(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("---\ncount: 123\n---\n\n# Title\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Links: &schema.LinkRule{
			BlockedDomains: []string{"blocked.com"},
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewLinkValidationRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Expected no violations for non-URL frontmatter value, got %d", len(violations))
	}
}

func TestLinkValidationFrontmatterInternalAnchor(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("---\nref: \"#nonexistent\"\n---\n\n# Title\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Links: &schema.LinkRule{
			ValidateInternal: true,
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewLinkValidationRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Expected violation for broken anchor in frontmatter")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "nonexistent") && strings.Contains(v.Message, "does not exist") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected violation mentioning broken anchor, got: %v", violations)
	}
}

func TestLinkValidationFrontmatterValidAnchor(t *testing.T) {
	p := parser.New()
	doc, err := p.Parse("test.md", []byte("---\nref: \"#details\"\n---\n\n# Title\n\n## Details\n\nSome content.\n"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s := &schema.Schema{
		Links: &schema.LinkRule{
			ValidateInternal: true,
		},
	}

	ctx := vast.NewContext(doc, s, "")
	rule := NewLinkValidationRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) != 0 {
		t.Errorf("Expected no violations for valid anchor in frontmatter, got %d: %v", len(violations), violations)
	}
}

func TestLinkValidationRootRelativePathBroken(t *testing.T) {
	tmpDir := t.TempDir()

	// Create docs directory but NOT the target directory
	docsDir := filepath.Join(tmpDir, "docs")
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		t.Fatalf("Failed to create docs directory: %v", err)
	}

	// Create document with root-relative link to non-existent directory
	mainFile := filepath.Join(docsDir, "README.md")
	content := []byte("# Title\n\n[folder1](/depth1/nonexistent/)\n")
	if err := os.WriteFile(mainFile, content, 0o644); err != nil {
		t.Fatalf("Failed to create main file: %v", err)
	}

	p := parser.New()
	doc, err := p.ParseFile(mainFile)
	if err != nil {
		t.Fatalf("ParseFile() error: %v", err)
	}

	s := &schema.Schema{
		Links: &schema.LinkRule{
			ValidateFiles: true,
		},
	}

	// Create context with rootDir set to tmpDir
	ctx := vast.NewContext(doc, s, tmpDir)
	rule := NewLinkValidationRule()
	violations := rule.ValidateWithContext(ctx)

	if len(violations) == 0 {
		t.Fatal("Expected violation for broken root-relative link, got none")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "/depth1/nonexistent/") && strings.Contains(v.Message, "does not exist") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected violation mentioning broken root-relative link")
	}
}
