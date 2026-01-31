package schema

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValidSchema(t *testing.T) {
	tmpDir := t.TempDir()
	schemaFile := filepath.Join(tmpDir, "schema.yml")

	content := []byte(`structure:
  - heading: "# Title"
  - heading: "## Section"
    optional: true
`)
	if err := os.WriteFile(schemaFile, content, 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	schema, err := Load(schemaFile)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if schema == nil {
		t.Fatal("Load() returned nil schema")
	}

	if len(schema.Structure) != 2 {
		t.Errorf("expected 2 structure elements, got %d", len(schema.Structure))
	}

	if schema.Structure[0].Heading.Literal != "# Title" {
		t.Errorf("Structure[0].Heading.Literal = %q, want %q", schema.Structure[0].Heading.Literal, "# Title")
	}

	if !schema.Structure[1].Optional {
		t.Error("Structure[1] should be optional")
	}
}

func TestLoadSchemaWithRules(t *testing.T) {
	tmpDir := t.TempDir()
	schemaFile := filepath.Join(tmpDir, "schema.yml")

	content := []byte(`structure:
  - heading: "# Title"
    required_text:
      - "important"
    code_blocks:
      - { lang: go, min: 1 }
`)
	if err := os.WriteFile(schemaFile, content, 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	schema, err := Load(schemaFile)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if schema.Structure[0].SectionRules == nil {
		t.Fatal("SectionRules should not be nil")
	}

	if len(schema.Structure[0].RequiredText) != 1 {
		t.Errorf("expected 1 required text, got %d", len(schema.Structure[0].RequiredText))
	}

	if len(schema.Structure[0].CodeBlocks) != 1 {
		t.Errorf("expected 1 code block rule, got %d", len(schema.Structure[0].CodeBlocks))
	}

	if schema.Structure[0].CodeBlocks[0].Lang != "go" {
		t.Errorf("CodeBlocks[0].Lang = %q, want %q", schema.Structure[0].CodeBlocks[0].Lang, "go")
	}
}

func TestLoadSchemaWithChildren(t *testing.T) {
	tmpDir := t.TempDir()
	schemaFile := filepath.Join(tmpDir, "schema.yml")

	content := []byte(`structure:
  - heading: "# Title"
    children:
      - heading: "## Child"
      - heading: "## Another Child"
        optional: true
`)
	if err := os.WriteFile(schemaFile, content, 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	schema, err := Load(schemaFile)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if len(schema.Structure[0].Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(schema.Structure[0].Children))
	}

	if schema.Structure[0].Children[0].Heading.Literal != "## Child" {
		t.Errorf("Children[0].Heading.Literal = %q, want %q", schema.Structure[0].Children[0].Heading.Literal, "## Child")
	}

	if !schema.Structure[0].Children[1].Optional {
		t.Error("Children[1] should be optional")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	schemaFile := filepath.Join(tmpDir, "invalid.yml")

	content := []byte("this is not valid: yaml: [")
	if err := os.WriteFile(schemaFile, content, 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	_, err := Load(schemaFile)
	if err == nil {
		t.Error("Load() should return error for invalid YAML")
	}
}

func TestLoadNonexistentFile(t *testing.T) {
	_, err := Load("/nonexistent/path/schema.yml")
	if err == nil {
		t.Error("Load() should return error for nonexistent file")
	}
}

func TestFindSchemaInCurrentDir(t *testing.T) {
	tmpDir := t.TempDir()
	schemaFile := filepath.Join(tmpDir, ".mdschema.yml")

	content := []byte(`structure:
  - heading: "# Title"
`)
	if err := os.WriteFile(schemaFile, content, 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	foundPath, err := FindSchema(tmpDir)
	if err != nil {
		t.Fatalf("FindSchema() error: %v", err)
	}

	if foundPath != schemaFile {
		t.Errorf("FindSchema() = %q, want %q", foundPath, schemaFile)
	}
}

func TestFindSchemaInParentDir(t *testing.T) {
	tmpDir := t.TempDir()
	childDir := filepath.Join(tmpDir, "child")
	if err := os.Mkdir(childDir, 0o755); err != nil {
		t.Fatalf("failed to create child dir: %v", err)
	}

	schemaFile := filepath.Join(tmpDir, ".mdschema.yml")
	content := []byte(`structure:
  - heading: "# Title"
`)
	if err := os.WriteFile(schemaFile, content, 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	foundPath, err := FindSchema(childDir)
	if err != nil {
		t.Fatalf("FindSchema() error: %v", err)
	}

	if foundPath != schemaFile {
		t.Errorf("FindSchema() = %q, want %q", foundPath, schemaFile)
	}
}

func TestFindSchemaNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := FindSchema(tmpDir)
	if err == nil {
		t.Error("FindSchema() should return error when schema not found")
	}
}

func TestFindSchemaFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	schemaFile := filepath.Join(tmpDir, ".mdschema.yml")
	markdownFile := filepath.Join(tmpDir, "README.md")

	// Create schema file
	content := []byte(`structure:
  - heading: "# Title"
`)
	if err := os.WriteFile(schemaFile, content, 0o644); err != nil {
		t.Fatalf("failed to create schema file: %v", err)
	}

	// Create markdown file
	if err := os.WriteFile(markdownFile, []byte("# Title"), 0o644); err != nil {
		t.Fatalf("failed to create markdown file: %v", err)
	}

	// FindSchema should work when given a file path
	foundPath, err := FindSchema(markdownFile)
	if err != nil {
		t.Fatalf("FindSchema() error: %v", err)
	}

	if foundPath != schemaFile {
		t.Errorf("FindSchema() = %q, want %q", foundPath, schemaFile)
	}
}

func TestIsDir(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	if !isDir(tmpDir) {
		t.Error("isDir() should return true for directory")
	}

	if isDir(tmpFile) {
		t.Error("isDir() should return false for file")
	}

	if isDir("/nonexistent/path") {
		t.Error("isDir() should return false for nonexistent path")
	}
}
