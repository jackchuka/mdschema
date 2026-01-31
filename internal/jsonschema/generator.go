// Package jsonschema provides JSON Schema generation for mdschema configuration files.
package jsonschema

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/jackchuka/mdschema/internal/schema"
)

// lookupComment reads descriptions from lc: (line comment) or hc: (head comment) struct tags.
// This makes lc:/hc: the single source of truth for both yaml-comment and JSON Schema descriptions.
func lookupComment(t reflect.Type, fieldName string) string {
	if fieldName == "" {
		return ""
	}
	f, found := t.FieldByName(fieldName)
	if !found {
		return ""
	}
	// Try lc: first (line comment - most common)
	if desc := f.Tag.Get("lc"); desc != "" {
		return capitalizeFirst(desc)
	}
	// Fall back to hc: (head comment - used on top-level Schema fields)
	return capitalizeFirst(f.Tag.Get("hc"))
}

// capitalizeFirst capitalizes the first letter of a string.
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// Generate creates a JSON Schema from the Schema type for editor autocomplete and validation.
func Generate() ([]byte, error) {
	r := &jsonschema.Reflector{
		DoNotReference: false,
		LookupComment:  lookupComment,
	}

	// Reflect the main Schema type
	s := r.Reflect(&schema.Schema{})

	// Also reflect the types that are referenced in custom JSONSchema() methods
	// but not automatically discovered by the reflector.
	// We use a fresh reflector with ExpandedStruct to get inline definitions.
	inlineReflector := &jsonschema.Reflector{
		DoNotReference: true,
		ExpandedStruct: true,
		LookupComment:  lookupComment,
	}

	// Additional types that need explicit definitions
	additionalTypes := []struct {
		name string
		typ  reflect.Type
	}{
		{"CodeBlockRule", reflect.TypeOf(schema.CodeBlockRule{})},
		{"ImageRule", reflect.TypeOf(schema.ImageRule{})},
		{"TableRule", reflect.TypeOf(schema.TableRule{})},
		{"ListRule", reflect.TypeOf(schema.ListRule{})},
		{"WordCountRule", reflect.TypeOf(schema.WordCountRule{})},
	}

	for _, t := range additionalTypes {
		def := inlineReflector.ReflectFromType(t.typ)
		// Clear metadata that shouldn't be in definitions
		def.Version = ""
		def.ID = ""
		if s.Definitions == nil {
			s.Definitions = make(jsonschema.Definitions)
		}
		s.Definitions[t.name] = def
	}

	// Set schema metadata
	s.ID = "https://raw.githubusercontent.com/jackchuka/mdschema/main/schema.json"
	s.Title = "mdschema"
	s.Description = "Schema for mdschema YAML configuration files (.mdschema.yml)"

	return json.MarshalIndent(s, "", "  ")
}
