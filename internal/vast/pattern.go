package vast

import (
	"regexp"
	"strings"

	"github.com/jackchuka/mdschema/internal/parser"
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

// MatchesHeadingPattern checks if a heading matches a pattern with explicit regex flag.
func (pm *PatternMatcher) MatchesHeadingPattern(heading *parser.Heading, pattern string, isRegex bool) bool {
	// Construct full heading text with level markers
	levelMarkers := strings.Repeat("#", heading.Level)
	fullHeading := levelMarkers + " " + heading.Text

	if isRegex {
		return pm.matchRegexPattern(fullHeading, pattern)
	}

	// Exact match - auto-anchor for precise matching
	expectedPattern := "^" + regexp.QuoteMeta(pattern) + "$"
	return pm.matchRegexPattern(fullHeading, expectedPattern)
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
