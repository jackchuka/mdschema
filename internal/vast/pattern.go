package vast

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/schema"
)

// PatternMatcher provides utilities for matching heading patterns.
type PatternMatcher struct {
	// Cache compiled regexes to avoid recompilation
	regexCache map[string]*regexp.Regexp
}

// NewPatternMatcher creates a new pattern matcher with caching.
func NewPatternMatcher() *PatternMatcher {
	return &PatternMatcher{
		regexCache: make(map[string]*regexp.Regexp),
	}
}

// MatchesHeading checks if a heading matches a HeadingPattern (literal, pattern, or expr).
func (pm *PatternMatcher) MatchesHeading(heading *parser.Heading, hp schema.HeadingPattern, filename string) bool {
	// If expression is provided, use expression matching
	if hp.Expr != "" {
		return pm.matchesHeadingExpr(heading, hp.Expr, filename)
	}

	// Construct full heading text with level markers
	levelMarkers := strings.Repeat("#", heading.Level)
	fullHeading := levelMarkers + " " + heading.Text

	// If literal is provided (scalar form), use exact match
	if hp.Literal != "" {
		expectedPattern := "^" + regexp.QuoteMeta(hp.Literal) + "$"
		return pm.matchRegexPattern(fullHeading, expectedPattern)
	}

	// Otherwise use pattern matching (always regex)
	return pm.matchRegexPattern(fullHeading, hp.Pattern)
}

// matchesHeadingExpr evaluates an expression to check if heading matches.
func (pm *PatternMatcher) matchesHeadingExpr(heading *parser.Heading, expression, documentPath string) bool {
	// Extract filename without extension
	filename := ExtractFilename(documentPath)
	if filename == "" {
		return false
	}

	// Build expression environment
	env := map[string]any{
		"filename":    filename,
		"heading":     heading.Text,
		"level":       heading.Level,
		"slug":        parser.GenerateSlug,
		"kebab":       toKebabCase,
		"lower":       strings.ToLower,
		"upper":       strings.ToUpper,
		"trim":        strings.TrimSpace,
		"strContains": strings.Contains,
		"hasPrefix":   strings.HasPrefix,
		"hasSuffix":   strings.HasSuffix,
		"replace":     strings.ReplaceAll,
		"trimPrefix":  trimPrefixRegex,
		"trimSuffix":  trimSuffixRegex,
		"match":       matchRegex,
	}

	program, err := expr.Compile(expression, expr.Env(env), expr.AsBool())
	if err != nil {
		return false
	}

	result, err := expr.Run(program, env)
	if err != nil {
		return false
	}

	matched, ok := result.(bool)
	return ok && matched
}

// matchRegexPattern compiles and matches a regex pattern with caching.
func (pm *PatternMatcher) matchRegexPattern(text, pattern string) bool {
	// Auto-anchor if not already anchored for explicit regex
	if !strings.HasPrefix(pattern, "^") {
		pattern = "^" + pattern
	}
	if !strings.HasSuffix(pattern, "$") {
		pattern = pattern + "$"
	}

	// Check cache first
	re, exists := pm.regexCache[pattern]
	if !exists {
		var err error
		re, err = regexp.Compile(pattern)
		if err != nil {
			// If regex compilation fails, treat as literal string
			return text == strings.TrimPrefix(strings.TrimSuffix(pattern, "$"), "^")
		}
		pm.regexCache[pattern] = re
	}

	return re.MatchString(text)
}

// ExtractFilename extracts the filename without extension from a path
func ExtractFilename(path string) string {
	if path == "" {
		return ""
	}
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

// trimPrefixRegex removes a regex pattern from the start of a string
func trimPrefixRegex(s, pattern string) string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return s
	}
	loc := re.FindStringIndex(s)
	if loc != nil && loc[0] == 0 {
		return s[loc[1]:]
	}
	return s
}

// trimSuffixRegex removes a regex pattern from the end of a string
func trimSuffixRegex(s, pattern string) string {
	re, err := regexp.Compile(pattern + "$")
	if err != nil {
		return s
	}
	return re.ReplaceAllString(s, "")
}

// matchRegex checks if a string matches a regex pattern
func matchRegex(s, pattern string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(s)
}

// toKebabCase converts PascalCase/camelCase to kebab-case
// Examples: "CreateUnit" -> "create-unit", "XMLParser" -> "xml-parser"
func toKebabCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			// Check if previous char was lowercase or next char is lowercase (for acronyms)
			prevLower := i > 0 && s[i-1] >= 'a' && s[i-1] <= 'z'
			nextLower := i+1 < len(s) && s[i+1] >= 'a' && s[i+1] <= 'z'
			if prevLower || nextLower {
				result.WriteRune('-')
			}
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
