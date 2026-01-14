package rules

import (
	"testing"

	"github.com/jackchuka/mdschema/internal/parser"
)

func TestNewPatternMatcher(t *testing.T) {
	pm := NewPatternMatcher()
	if pm == nil {
		t.Fatal("NewPatternMatcher() returned nil")
	}
	if pm.regexCache == nil {
		t.Fatal("NewPatternMatcher() returned nil regexCache")
	}
}

func TestMatchesHeadingPatternExact(t *testing.T) {
	pm := NewPatternMatcher()

	tests := []struct {
		heading *parser.Heading
		pattern string
		isRegex bool
		want    bool
	}{
		{
			heading: &parser.Heading{Level: 1, Text: "Title"},
			pattern: "# Title",
			isRegex: false,
			want:    true,
		},
		{
			heading: &parser.Heading{Level: 2, Text: "Section"},
			pattern: "## Section",
			isRegex: false,
			want:    true,
		},
		{
			heading: &parser.Heading{Level: 1, Text: "Title"},
			pattern: "# Wrong",
			isRegex: false,
			want:    false,
		},
		{
			heading: &parser.Heading{Level: 1, Text: "Title"},
			pattern: "## Title", // Wrong level
			isRegex: false,
			want:    false,
		},
	}

	for _, tc := range tests {
		got := pm.MatchesHeadingPattern(tc.heading, tc.pattern, tc.isRegex)
		if got != tc.want {
			t.Errorf("MatchesHeadingPattern(%v, %q, %v) = %v, want %v",
				tc.heading, tc.pattern, tc.isRegex, got, tc.want)
		}
	}
}

func TestMatchesHeadingPatternRegex(t *testing.T) {
	pm := NewPatternMatcher()

	tests := []struct {
		heading *parser.Heading
		pattern string
		isRegex bool
		want    bool
	}{
		{
			heading: &parser.Heading{Level: 1, Text: "My Project"},
			pattern: "# [a-zA-Z ]+",
			isRegex: true,
			want:    true,
		},
		{
			heading: &parser.Heading{Level: 2, Text: "Installation"},
			pattern: "## (Installation|Setup)",
			isRegex: true,
			want:    true,
		},
		{
			heading: &parser.Heading{Level: 2, Text: "Setup"},
			pattern: "## (Installation|Setup)",
			isRegex: true,
			want:    true,
		},
		{
			heading: &parser.Heading{Level: 2, Text: "Other"},
			pattern: "## (Installation|Setup)",
			isRegex: true,
			want:    false,
		},
	}

	for _, tc := range tests {
		got := pm.MatchesHeadingPattern(tc.heading, tc.pattern, tc.isRegex)
		if got != tc.want {
			t.Errorf("MatchesHeadingPattern(%v, %q, %v) = %v, want %v",
				tc.heading, tc.pattern, tc.isRegex, got, tc.want)
		}
	}
}

func TestPatternMatcherCaching(t *testing.T) {
	pm := NewPatternMatcher()

	heading := &parser.Heading{Level: 1, Text: "Test"}
	pattern := "# Test"

	// First call should cache the pattern
	pm.MatchesHeadingPattern(heading, pattern, false)

	// The expected pattern should be in the cache
	expectedCachedPattern := "^\\# Test$"
	if _, exists := pm.regexCache[expectedCachedPattern]; !exists {
		// Check cache size at least grew
		if len(pm.regexCache) == 0 {
			t.Error("pattern should have been cached")
		}
	}

	// Second call should use the cache (we verify cache size doesn't grow)
	cacheSize := len(pm.regexCache)
	pm.MatchesHeadingPattern(heading, pattern, false)

	if len(pm.regexCache) != cacheSize {
		t.Errorf("cache size changed from %d to %d, expected no change", cacheSize, len(pm.regexCache))
	}
}

func TestContainsRegexMetachars(t *testing.T) {
	tests := []struct {
		pattern string
		want    bool
	}{
		{"# Simple", false},
		{"## Also Simple", false},
		{"# [a-z]+", true},
		{"# Title*", true},
		{"# Title?", true},
		{"# Title+", true},
		{"^# Title", true},
		{"# Title$", true},
		{"# Title\\.", true},
	}

	for _, tc := range tests {
		got := containsRegexMetachars(tc.pattern)
		if got != tc.want {
			t.Errorf("containsRegexMetachars(%q) = %v, want %v", tc.pattern, got, tc.want)
		}
	}
}
