package schema

import (
	"fmt"
	"reflect"
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
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch node.Kind {
	case yaml.MappingNode:
		if t.Kind() != reflect.Struct || opaqueTypes[t] {
			return
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
			for ft.Kind() == reflect.Ptr {
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
