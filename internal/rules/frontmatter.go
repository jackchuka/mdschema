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
		value, exists := fm.Data[field.Name]

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

	builder.WriteString("---\n")

	for _, field := range s.Frontmatter.Fields {
		placeholder := r.getPlaceholder(field)
		if !field.Optional {
			builder.WriteString(field.Name + ": " + placeholder + " # required\n")
		} else {
			builder.WriteString(field.Name + ": " + placeholder + "\n")
		}
	}

	builder.WriteString("---\n\n")
	return true
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
	default:
		return "\"TODO\""
	}
}
