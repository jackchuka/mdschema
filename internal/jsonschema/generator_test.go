package jsonschema

import (
	"encoding/json"
	"testing"
)

func TestGenerate(t *testing.T) {
	schemaBytes, err := Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]any
	if err := json.Unmarshal(schemaBytes, &parsed); err != nil {
		t.Fatalf("Generated schema is not valid JSON: %v", err)
	}

	// Verify key schema properties exist
	if _, ok := parsed["$schema"]; !ok {
		t.Error("Generated schema missing $schema")
	}

	if _, ok := parsed["$id"]; !ok {
		t.Error("Generated schema missing $id")
	}

	if title, ok := parsed["title"]; !ok || title != "mdschema" {
		t.Errorf("Generated schema title = %v, want 'mdschema'", title)
	}

	// Verify definitions exist
	defs, ok := parsed["$defs"].(map[string]any)
	if !ok {
		t.Fatal("Generated schema missing $defs")
	}

	// Verify Schema type is defined
	schemaDef, ok := defs["Schema"].(map[string]any)
	if !ok {
		t.Fatal("Generated schema missing Schema definition")
	}

	// Verify Schema has properties
	props, ok := schemaDef["properties"].(map[string]any)
	if !ok {
		t.Fatal("Schema definition missing properties")
	}

	requiredProps := []string{"structure", "links", "heading_rules", "frontmatter"}
	for _, prop := range requiredProps {
		if _, ok := props[prop]; !ok {
			t.Errorf("Schema definition missing property %q", prop)
		}
	}

	// Verify definitions exist for referenced types
	requiredDefs := []string{"CodeBlockRule", "ImageRule", "TableRule", "ListRule", "WordCountRule", "StructureElement"}
	for _, def := range requiredDefs {
		if _, ok := defs[def]; !ok {
			t.Errorf("Generated schema missing definition %q", def)
		}
	}
}

func TestGenerateSchemaHasDescriptions(t *testing.T) {
	schemaBytes, err := Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(schemaBytes, &parsed); err != nil {
		t.Fatalf("Generated schema is not valid JSON: %v", err)
	}

	// Check that description exists at top level
	if desc, ok := parsed["description"]; !ok || desc == "" {
		t.Error("Generated schema missing top-level description")
	}
}

func TestGenerateSchemaStructureElementHasOneOf(t *testing.T) {
	schemaBytes, err := Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(schemaBytes, &parsed); err != nil {
		t.Fatalf("Generated schema is not valid JSON: %v", err)
	}

	defs := parsed["$defs"].(map[string]any)
	structureElement := defs["StructureElement"].(map[string]any)

	// StructureElement should have oneOf for string | object union type
	if _, ok := structureElement["oneOf"]; !ok {
		t.Error("StructureElement missing oneOf for union type support")
	}
}
