package rules

import (
	"regexp"
	"strings"
)

// containsTextPattern reports whether content contains literal as a substring,
// or, when literal is empty, matches pattern as a regex. An invalid regex falls
// back to a substring match.
func containsTextPattern(content, literal, pattern string) bool {
	if literal != "" {
		return strings.Contains(content, literal)
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return strings.Contains(content, pattern)
	}
	return re.MatchString(content)
}

func patternString(literal, pattern string) string {
	if literal != "" {
		return literal
	}
	return pattern
}

func patternHint(literal, pattern string) string {
	if literal != "" {
		return literal
	}
	return pattern + " (regex)"
}
