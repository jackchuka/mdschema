package integration

import "testing"

// examplesDir is the path to the examples directory from this test file
const examplesDir = "../../examples/"

// TestExampleSchemas validates that example schemas work correctly
func TestExampleSchemas(t *testing.T) {
	testCases := []TestCase{
		// README schema against project README
		{
			Name:       "README schema validates README.md",
			FilePath:   "../../README.md",
			SchemaPath: examplesDir + "README.mdschema.yml",
			ShouldPass: true,
		},
		// Requirements schema against requirements.md
		{
			Name:       "requirements schema validates requirements.md",
			FilePath:   examplesDir + "requirements.md",
			SchemaPath: examplesDir + "requirements.mdschema.yml",
			ShouldPass: true,
		},
		// Tutorial schema against tutorial.md
		{
			Name:       "tutorial schema validates tutorial.md",
			FilePath:   examplesDir + "tutorial.md",
			SchemaPath: examplesDir + "tutorial.mdschema.yml",
			ShouldPass: true,
		},
	}

	runTestCases(t, testCases)
}

// TestHeadingPatternSyntax tests the three heading pattern forms:
// - literal: heading: "## Title"
// - regex:   heading: {pattern: "## .*"}
// - expr:    heading: {expr: "slug(filename) == slug(heading)"}
func TestHeadingPatternSyntax(t *testing.T) {
	testCases := []TestCase{
		// Literal heading match
		{
			Name:       "literal heading - exact match",
			FilePath:   testdataDir + "patterns/literal_match.md",
			SchemaPath: testdataDir + "patterns/literal_schema.yml",
			ShouldPass: true,
		},
		{
			Name:         "literal heading - no match",
			FilePath:     testdataDir + "patterns/literal_nomatch.md",
			SchemaPath:   testdataDir + "patterns/literal_schema.yml",
			ShouldPass:   false,
			ExpectedRule: "structure",
		},

		// Regex heading match
		{
			Name:       "regex heading - matches pattern",
			FilePath:   testdataDir + "patterns/regex_match.md",
			SchemaPath: testdataDir + "patterns/regex_schema.yml",
			ShouldPass: true,
		},
		{
			Name:         "regex heading - no match",
			FilePath:     testdataDir + "patterns/regex_nomatch.md",
			SchemaPath:   testdataDir + "patterns/regex_schema.yml",
			ShouldPass:   false,
			ExpectedRule: "structure",
		},

		// Expression heading match
		{
			Name:       "expr heading - filename matches heading",
			FilePath:   testdataDir + "patterns/my-feature.md",
			SchemaPath: testdataDir + "patterns/expr_schema.yml",
			ShouldPass: true,
		},
	}

	runTestCases(t, testCases)
}

// TestRequiredTextSyntax tests the two required_text forms:
// - literal: required_text: "text"
// - regex:   required_text: {pattern: "\\d+"}
func TestRequiredTextSyntax(t *testing.T) {
	testCases := []TestCase{
		// Literal required text
		{
			Name:       "literal required text - found",
			FilePath:   testdataDir + "patterns/required_literal_match.md",
			SchemaPath: testdataDir + "patterns/required_literal_schema.yml",
			ShouldPass: true,
		},
		{
			Name:         "literal required text - not found",
			FilePath:     testdataDir + "patterns/required_literal_nomatch.md",
			SchemaPath:   testdataDir + "patterns/required_literal_schema.yml",
			ShouldPass:   false,
			ExpectedRule: "required-text",
		},

		// Regex required text
		{
			Name:       "regex required text - matches",
			FilePath:   testdataDir + "patterns/required_regex_match.md",
			SchemaPath: testdataDir + "patterns/required_regex_schema.yml",
			ShouldPass: true,
		},
		{
			Name:         "regex required text - no match",
			FilePath:     testdataDir + "patterns/required_regex_nomatch.md",
			SchemaPath:   testdataDir + "patterns/required_regex_schema.yml",
			ShouldPass:   false,
			ExpectedRule: "required-text",
		},
	}

	runTestCases(t, testCases)
}

// TestCountConstraint tests the count feature for multi-match elements
func TestCountConstraint(t *testing.T) {
	testCases := []TestCase{
		// Valid count - between min and max
		{
			Name:       "count - valid number of matches",
			FilePath:   testdataDir + "patterns/count_valid.md",
			SchemaPath: testdataDir + "patterns/count_schema.yml",
			ShouldPass: true,
		},
		// Too few matches (below min)
		{
			Name:         "count - too few matches",
			FilePath:     testdataDir + "patterns/count_too_few.md",
			SchemaPath:   testdataDir + "patterns/count_schema.yml",
			ShouldPass:   false,
			ExpectedRule: "structure",
		},
		// Too many matches (above max)
		{
			Name:         "count - too many matches",
			FilePath:     testdataDir + "patterns/count_too_many.md",
			SchemaPath:   testdataDir + "patterns/count_schema.yml",
			ShouldPass:   false,
			ExpectedRule: "structure",
		},
		// Content rules applied to each match
		{
			Name:         "count - content rules per match",
			FilePath:     testdataDir + "patterns/count_missing_code.md",
			SchemaPath:   testdataDir + "patterns/count_schema.yml",
			ShouldPass:   false,
			ExpectedRule: "codeblock",
		},
	}

	runTestCases(t, testCases)
}
