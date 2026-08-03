package schema

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Warning describes a non-fatal issue found while loading a schema.
type Warning struct {
	Message string
	Line    int
}

// opaqueTypes are leaf types with custom UnmarshalYAML that intentionally
// accept keys (e.g. "regex") not present as struct fields. We do not descend
// into them, to avoid false-positive "unknown key" warnings.
var opaqueTypes = map[reflect.Type]bool{
	reflect.TypeOf(HeadingPattern{}):       true,
	reflect.TypeOf(RequiredTextPattern{}):  true,
	reflect.TypeOf(ForbiddenTextPattern{}): true,
}

// checkUnknownKeys walks the raw YAML node tree in parallel with the Schema
// type graph and reports mapping keys that have no corresponding yaml tag.
func checkUnknownKeys(data []byte) ([]Warning, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	var warnings []Warning
	if doc.Kind == yaml.DocumentNode && len(doc.Content) > 0 {
		walkNode(doc.Content[0], reflect.TypeOf(Schema{}), &warnings)
	}
	return warnings, nil
}

func walkNode(node *yaml.Node, t reflect.Type, warnings *[]Warning) {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	switch node.Kind {
	case yaml.MappingNode:
		if t.Kind() != reflect.Struct || opaqueTypes[t] {
			return
		}
		if t == reflect.TypeOf(FrontmatterField{}) {
			checkEnumTypes(node, warnings)
		}
		allowed := allowedKeys(t)
		// Mapping content alternates key, value, key, value, ...
		for i := 0; i+1 < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valNode := node.Content[i+1]
			ft, ok := allowed[keyNode.Value]
			if !ok {
				*warnings = append(*warnings, Warning{
					Message: fmt.Sprintf("unknown key %q (ignored)", keyNode.Value),
					Line:    keyNode.Line,
				})
				continue
			}
			walkNode(valNode, ft, warnings)
		}
	case yaml.SequenceNode:
		if t.Kind() != reflect.Slice && t.Kind() != reflect.Array {
			return
		}
		for _, child := range node.Content {
			walkNode(child, t.Elem(), warnings)
		}
	}
}

var dateValueRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// checkEnumTypes reports enum entries that can never match the field's declared
// type — e.g. `{type: number, enum: [low, high]}`, which rejects every possible
// value. Array and object fields are skipped: for arrays the entries describe
// elements rather than the field itself, and object entries have no scalar tag
// to compare. A field with neither type nor format constrains nothing, so it is
// skipped too.
func checkEnumTypes(field *yaml.Node, warnings *[]Warning) {
	var name string
	var declared FieldType
	var format FieldFormat
	var enum *yaml.Node
	for i := 0; i+1 < len(field.Content); i += 2 {
		switch field.Content[i].Value {
		case "name":
			name = field.Content[i+1].Value
		case "type":
			declared = FieldType(field.Content[i+1].Value)
		case "format":
			format = FieldFormat(field.Content[i+1].Value)
		case "enum":
			enum = field.Content[i+1]
		}
	}
	if enum == nil || enum.Kind != yaml.SequenceNode {
		return
	}
	// A format implies a value type, so an enum entry can be impossible on
	// format alone — `{format: email, enum: [1, 2]}` matches nothing.
	if declared == "" {
		declared = typeForFormat(format)
	}

	for _, entry := range enum.Content {
		if enumEntryMatchesType(entry, declared) {
			continue
		}
		var msg string
		if shape := entryShape(entry); shape != "" {
			msg = fmt.Sprintf("enum value for field %q is %s, not a %s", name, shape, declared)
		} else {
			msg = fmt.Sprintf("enum value %q for field %q is not a %s", entry.Value, name, declared)
		}
		*warnings = append(*warnings, Warning{Message: msg, Line: entry.Line})
	}
}

// typeForFormat maps a format to the type its values must have, so enum entries
// can be checked against `format` when `type` is omitted.
func typeForFormat(format FieldFormat) FieldType {
	switch format {
	case FieldFormatDate:
		return FieldTypeDate
	case FieldFormatEmail, FieldFormatURL:
		return FieldTypeString
	}
	return ""
}

// entryShape names a non-scalar enum entry, returning "" for scalars. Sequence
// and mapping nodes have an empty Value, so quoting it reads as `enum value ""`.
func entryShape(entry *yaml.Node) string {
	switch entry.Kind {
	case yaml.SequenceNode:
		return "a list"
	case yaml.MappingNode:
		return "a mapping"
	}
	return ""
}

// enumEntryMatchesType compares a raw YAML node against a declared field type
// using the tag yaml.v3 resolved for it (!!int, !!str, !!timestamp, ...).
func enumEntryMatchesType(entry *yaml.Node, declared FieldType) bool {
	switch declared {
	case FieldTypeString:
		return entry.Tag == "!!str"
	case FieldTypeNumber:
		return entry.Tag == "!!int" || entry.Tag == "!!float"
	case FieldTypeBoolean:
		return entry.Tag == "!!bool"
	case FieldTypeDate:
		// Quoted dates stay !!str, so accept those that look like YYYY-MM-DD.
		return entry.Tag == "!!timestamp" ||
			(entry.Tag == "!!str" && dateValueRegex.MatchString(entry.Value))
	}
	return true
}

// allowedKeys returns the yaml keys a struct type accepts, mapped to the field
// type. It flattens ,inline embedded structs (following pointers).
func allowedKeys(t reflect.Type) map[string]reflect.Type {
	keys := make(map[string]reflect.Type)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		name, opts := parseYAMLTag(f.Tag.Get("yaml"))
		if name == "-" {
			continue
		}
		if hasOpt(opts, "inline") {
			ft := f.Type
			for ft.Kind() == reflect.Pointer {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct {
				for k, v := range allowedKeys(ft) {
					keys[k] = v
				}
			}
			continue
		}
		if name == "" {
			name = strings.ToLower(f.Name)
		}
		keys[name] = f.Type
	}
	return keys
}

func parseYAMLTag(tag string) (name string, opts []string) {
	parts := strings.Split(tag, ",")
	return parts[0], parts[1:]
}

func hasOpt(opts []string, want string) bool {
	for _, o := range opts {
		if o == want {
			return true
		}
	}
	return false
}
