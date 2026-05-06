package rules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jackchuka/mdschema/internal/schema"
	"github.com/jackchuka/mdschema/internal/vast"
)

// FrontmatterRule validates YAML frontmatter at the start of documents
type FrontmatterRule struct {
}

var _ Rule = (*FrontmatterRule)(nil)
var _ FrontmatterGenerator = (*FrontmatterRule)(nil)

// NewFrontmatterRule creates a new frontmatter rule
func NewFrontmatterRule() *FrontmatterRule {
	return &FrontmatterRule{}
}

// Name returns the rule identifier
func (r *FrontmatterRule) Name() string {
	return "frontmatter"
}

// ValidateWithContext validates using VAST (validation-ready AST)
func (r *FrontmatterRule) ValidateWithContext(ctx *vast.Context) []Violation {
	violations := make([]Violation, 0)

	// Check if frontmatter rules are configured
	if ctx.Schema.Frontmatter == nil {
		return violations
	}

	config := ctx.Schema.Frontmatter
	fm := ctx.Tree.Document.FrontMatter

	// Check if frontmatter is required but missing
	if !config.Optional && fm == nil {
		violations = append(violations,
			NewViolation(r.Name(), "Frontmatter is required but not found", 1, 1))
		return violations
	}

	// If no frontmatter exists and it's not required, nothing to validate
	if fm == nil {
		return violations
	}

	// If frontmatter exists but couldn't be parsed, report error
	if fm.Data == nil {
		violations = append(violations,
			NewViolation(r.Name(), "Frontmatter could not be parsed as valid YAML", 1, 1))
		return violations
	}

	// Validate required fields
	for _, field := range config.Fields {
		value, exists := lookupField(fm.Data, field.Name)

		if !field.Optional && !exists {
			violations = append(violations,
				NewViolation(r.Name(), fmt.Sprintf("Required frontmatter field '%s' is missing", field.Name), 1, 1))
			continue
		}

		if !exists {
			continue
		}

		// Validate field type if specified
		if field.Type != "" {
			if err := r.validateFieldType(field.Name, value, field.Type); err != "" {
				violations = append(violations,
					NewViolation(r.Name(), err, 1, 1))
			}
		}

		// Validate field format if specified
		if field.Format != "" {
			if err := r.validateFieldFormat(field.Name, value, field.Format); err != "" {
				violations = append(violations,
					NewViolation(r.Name(), err, 1, 1))
			}
		}
	}

	return violations
}

// splitFieldPath splits a dot-notation path into segments. A literal dot can
// be escaped with a backslash (e.g. "weird\\.key" → ["weird.key"]).
func splitFieldPath(name string) []string {
	segments := []string{}
	var current strings.Builder
	for i := 0; i < len(name); i++ {
		c := name[i]
		if c == '\\' && i+1 < len(name) && name[i+1] == '.' {
			current.WriteByte('.')
			i++
			continue
		}
		if c == '.' {
			segments = append(segments, current.String())
			current.Reset()
			continue
		}
		current.WriteByte(c)
	}
	segments = append(segments, current.String())
	return segments
}

// lookupField walks a frontmatter data map using a dot-notation path. It
// handles both map[string]any and map[any]any (yaml.v3 may produce the
// latter for nested maps).
func lookupField(data map[string]any, name string) (any, bool) {
	segments := splitFieldPath(name)
	var current any = data
	for _, seg := range segments {
		switch m := current.(type) {
		case map[string]any:
			v, ok := m[seg]
			if !ok {
				return nil, false
			}
			current = v
		case map[any]any:
			v, ok := m[seg]
			if !ok {
				return nil, false
			}
			current = v
		default:
			return nil, false
		}
	}
	return current, true
}

// validateFieldType checks if a field value matches the expected type
func (r *FrontmatterRule) validateFieldType(name string, value any, expectedType schema.FieldType) string {
	switch expectedType {
	case schema.FieldTypeString:
		if _, ok := value.(string); !ok {
			return fmt.Sprintf("Frontmatter field '%s' should be a string", name)
		}
	case schema.FieldTypeNumber:
		switch value.(type) {
		case int, int64, float64:
			// OK
		default:
			return fmt.Sprintf("Frontmatter field '%s' should be a number", name)
		}
	case schema.FieldTypeBoolean:
		if _, ok := value.(bool); !ok {
			return fmt.Sprintf("Frontmatter field '%s' should be a boolean", name)
		}
	case schema.FieldTypeArray:
		if _, ok := value.([]any); !ok {
			return fmt.Sprintf("Frontmatter field '%s' should be an array", name)
		}
	case schema.FieldTypeObject:
		switch value.(type) {
		case map[string]any, map[any]any:
		default:
			return fmt.Sprintf("Frontmatter field '%s' should be an object", name)
		}
	case schema.FieldTypeDate:
		// Date can be string or time.Time depending on YAML parsing
		// YAML may parse dates like 2024-01-15 as time.Time
		if err := r.validateDateValue(value); err != "" {
			return fmt.Sprintf("Frontmatter field '%s' %s", name, err)
		}
	}
	return ""
}

// validateDateValue checks if a value is a valid date
func (r *FrontmatterRule) validateDateValue(value any) string {
	switch v := value.(type) {
	case string:
		if !isValidDateFormat(v) {
			return "should be in YYYY-MM-DD format"
		}
	default:
		// YAML v3 parses dates as time.Time, which is valid
		// Check if it's a time.Time by seeing if it has the right methods
		if _, ok := value.(interface{ Year() int }); ok {
			return "" // Valid time.Time
		}
		return "should be a date (YYYY-MM-DD)"
	}
	return ""
}

// validateFieldFormat checks if a field value matches the expected format
func (r *FrontmatterRule) validateFieldFormat(name string, value any, format schema.FieldFormat) string {
	switch format {
	case schema.FieldFormatDate:
		// Date format can be validated on string or time.Time
		if err := r.validateDateValue(value); err != "" {
			return fmt.Sprintf("Frontmatter field '%s' %s", name, err)
		}
		return ""
	}

	// Other formats require string values
	str, ok := value.(string)
	if !ok {
		return fmt.Sprintf("Frontmatter field '%s' format validation requires a string value", name)
	}

	switch format {
	case schema.FieldFormatEmail:
		if !isValidEmail(str) {
			return fmt.Sprintf("Frontmatter field '%s' should be a valid email address", name)
		}
	case schema.FieldFormatURL:
		if !isValidURL(str) {
			return fmt.Sprintf("Frontmatter field '%s' should be a valid URL", name)
		}
	}
	return ""
}

// isValidDateFormat checks if a string is in YYYY-MM-DD format
func isValidDateFormat(s string) bool {
	dateRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	return dateRegex.MatchString(s)
}

// isValidEmail checks if a string looks like an email address
func isValidEmail(s string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(s)
}

// isValidURL checks if a string looks like a URL
func isValidURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// Generate generates YAML frontmatter based on schema configuration
func (r *FrontmatterRule) Generate(builder *strings.Builder, s *schema.Schema) bool {
	if s.Frontmatter == nil || len(s.Frontmatter.Fields) == 0 {
		return false
	}

	tree := buildFrontmatterTree(s.Frontmatter.Fields)

	builder.WriteString("---\n")
	r.writeFrontmatterTree(builder, tree, 0)
	builder.WriteString("---\n\n")
	return true
}

// fmTreeNode is a node in the tree built from dot-notation field paths.
// Branches (nodes with children) emit as YAML maps; leaves emit a placeholder.
type fmTreeNode struct {
	key      string
	field    *schema.FrontmatterField
	children []*fmTreeNode
}

func buildFrontmatterTree(fields []schema.FrontmatterField) []*fmTreeNode {
	var roots []*fmTreeNode
	for i := range fields {
		f := fields[i]
		segments := splitFieldPath(f.Name)
		insertFrontmatterPath(&roots, segments, &f, 0)
	}
	return roots
}

func insertFrontmatterPath(siblings *[]*fmTreeNode, segments []string, f *schema.FrontmatterField, depth int) {
	seg := segments[depth]
	var node *fmTreeNode
	for _, s := range *siblings {
		if s.key == seg {
			node = s
			break
		}
	}
	if node == nil {
		node = &fmTreeNode{key: seg}
		*siblings = append(*siblings, node)
	}
	if depth == len(segments)-1 {
		node.field = f
		return
	}
	insertFrontmatterPath(&node.children, segments, f, depth+1)
}

func (r *FrontmatterRule) writeFrontmatterTree(builder *strings.Builder, nodes []*fmTreeNode, depth int) {
	indent := strings.Repeat("  ", depth)
	for _, n := range nodes {
		if len(n.children) > 0 {
			if hasRequiredDescendant(n) {
				builder.WriteString(indent + n.key + ": # required\n")
			} else {
				builder.WriteString(indent + n.key + ":\n")
			}
			r.writeFrontmatterTree(builder, n.children, depth+1)
			continue
		}
		if n.field == nil {
			continue
		}
		placeholder := r.getPlaceholder(*n.field)
		if !n.field.Optional {
			builder.WriteString(indent + n.key + ": " + placeholder + " # required\n")
		} else {
			builder.WriteString(indent + n.key + ": " + placeholder + "\n")
		}
	}
}

func hasRequiredDescendant(n *fmTreeNode) bool {
	if n.field != nil && !n.field.Optional {
		return true
	}
	for _, c := range n.children {
		if hasRequiredDescendant(c) {
			return true
		}
	}
	return false
}

// getPlaceholder returns an appropriate placeholder value based on field type/format
func (r *FrontmatterRule) getPlaceholder(field schema.FrontmatterField) string {
	// Check format first as it's more specific
	switch field.Format {
	case schema.FieldFormatDate:
		return "2024-01-01"
	case schema.FieldFormatEmail:
		return "user@example.com"
	case schema.FieldFormatURL:
		return "https://example.com"
	}

	// Fall back to type
	switch field.Type {
	case schema.FieldTypeString:
		return "\"TODO\""
	case schema.FieldTypeNumber:
		return "0"
	case schema.FieldTypeBoolean:
		return "false"
	case schema.FieldTypeArray:
		return "[\"item1\", \"item2\"]"
	case schema.FieldTypeDate:
		return "2024-01-01"
	case schema.FieldTypeObject:
		return "{}"
	default:
		return "\"TODO\""
	}
}
