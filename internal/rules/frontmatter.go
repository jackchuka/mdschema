package rules

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/jackchuka/mdschema/internal/schema"
	"github.com/jackchuka/mdschema/internal/vast"
)

// dateLayout is the only date form mdschema validates (see FieldFormatDate).
const dateLayout = "2006-01-02"

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
		wellFormed := true
		if field.Type != "" {
			if err := r.validateFieldType(field.Name, value, field.Type); err != "" {
				violations = append(violations,
					NewViolation(r.Name(), err, 1, 1))
				wellFormed = false
			}
		}

		// Validate field format if specified
		if field.Format != "" {
			if err := r.validateFieldFormat(field.Name, value, field.Format); err != "" {
				violations = append(violations,
					NewViolation(r.Name(), err, 1, 1))
				wellFormed = false
			}
		}

		// Validate allowed values if specified. A value that already failed its
		// type or format check cannot match any enum entry either, and reporting
		// both leaves the reader with two violations for one mistake — the
		// earlier, more specific message stands on its own.
		if len(field.Enum) > 0 && wellFormed {
			for _, err := range r.validateFieldEnum(field, value) {
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

// validateFieldEnum checks if a field value is one of the allowed values. For
// array fields every element is checked, so a document listing several
// disallowed values reports all of them in a single run.
func (r *FrontmatterRule) validateFieldEnum(field schema.FrontmatterField, value any) []string {
	if field.Type == schema.FieldTypeArray {
		arr, ok := value.([]any)
		if !ok {
			// validateFieldType already reports the wrong shape.
			return nil
		}
		var errs []string
		for _, elem := range arr {
			if !enumContains(field.Enum, elem) {
				errs = append(errs, fmt.Sprintf("Frontmatter field '%s' contains %s, allowed values: %s",
					field.Name, formatEnumValue(elem), formatEnumList(field.Enum)))
			}
		}
		return errs
	}
	if !enumContains(field.Enum, value) {
		return []string{fmt.Sprintf("Frontmatter field '%s' has value %s, allowed values: %s",
			field.Name, formatEnumValue(value), formatEnumList(field.Enum))}
	}
	return nil
}

func enumContains(allowed []any, value any) bool {
	for _, a := range allowed {
		if enumEqual(a, value) {
			return true
		}
	}
	return false
}

// enumEqual compares an allowed value against a frontmatter value. Numeric
// types (int, int64, float64) are interchangeable since YAML parsing may
// produce any of them. Dates are compared as YYYY-MM-DD because the two sides
// come from different decoders: schemas via yaml.v3, which resolves timestamps
// to time.Time, and document frontmatter via goldmark-meta's yaml.v2, which
// leaves them as strings. Maps and slices recurse so that nested dates and
// numbers get the same treatment as top-level ones.
func enumEqual(a, b any) bool {
	if af, aok := toFloat(a); aok {
		bf, bok := toFloat(b)
		return bok && af == bf
	}
	if ad, aok := toDateString(a); aok {
		bd, bok := toDateString(b)
		return bok && ad == bd
	}
	if am, aok := toStringMap(a); aok {
		bm, bok := toStringMap(b)
		if !bok || len(am) != len(bm) {
			return false
		}
		for k, av := range am {
			bv, ok := bm[k]
			if !ok || !enumEqual(av, bv) {
				return false
			}
		}
		return true
	}
	if as, aok := a.([]any); aok {
		bs, bok := b.([]any)
		if !bok || len(as) != len(bs) {
			return false
		}
		for i := range as {
			if !enumEqual(as[i], bs[i]) {
				return false
			}
		}
		return true
	}
	// reflect.DeepEqual rather than ==: == panics on non-comparable dynamic
	// types, and only scalars are guaranteed to have reached this point.
	return reflect.DeepEqual(a, b)
}

// toStringMap normalizes either map shape a YAML decoder may produce. Nested
// maps arrive as map[string]any from schemas (yaml.v3) but as map[any]any from
// document frontmatter (goldmark-meta's yaml.v2), so comparing the two sides
// directly rejects structurally identical maps on their type alone. Non-string
// keys are rendered with %v, which is enough to compare them.
func toStringMap(v any) (map[string]any, bool) {
	switch m := v.(type) {
	case map[string]any:
		return m, true
	case map[any]any:
		out := make(map[string]any, len(m))
		for k, elem := range m {
			out[fmt.Sprintf("%v", k)] = elem
		}
		return out, true
	}
	return nil, false
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case float64:
		return n, true
	}
	return 0, false
}

// toDateString normalizes a date to YYYY-MM-DD. Timestamps carrying a clock
// component are left alone, since mdschema's date support covers whole days
// only and truncating them would make distinct instants compare equal.
func toDateString(v any) (string, bool) {
	switch d := v.(type) {
	case time.Time:
		if d.Hour() == 0 && d.Minute() == 0 && d.Second() == 0 && d.Nanosecond() == 0 {
			return d.Format(dateLayout), true
		}
	case string:
		if isValidDateFormat(d) {
			return d, true
		}
	}
	return "", false
}

func formatEnumValue(v any) string {
	switch val := v.(type) {
	case nil:
		return "null"
	case string:
		return fmt.Sprintf("%q", val)
	case time.Time:
		if s, ok := toDateString(val); ok {
			return s
		}
	}
	return fmt.Sprintf("%v", v)
}

func formatEnumList(allowed []any) string {
	parts := make([]string, len(allowed))
	for i, a := range allowed {
		parts[i] = formatEnumValue(a)
	}
	return "[" + strings.Join(parts, ", ") + "]"
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
	// An enum pins the value down to a known set, so use its first entry
	if len(field.Enum) > 0 {
		if field.Type == schema.FieldTypeArray {
			return "[" + formatEnumValue(field.Enum[0]) + "]"
		}
		return formatEnumValue(field.Enum[0])
	}

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
