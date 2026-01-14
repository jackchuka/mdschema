package integration

import (
	"testing"

	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/rules"
	"github.com/jackchuka/mdschema/internal/schema"
)

// testdataDir is the path to the testdata directory from this test file
const testdataDir = "../../testdata/"

// TestCase represents a single test case
type TestCase struct {
	Name         string
	FilePath     string
	SchemaPath   string
	ShouldPass   bool
	ExpectedRule string // Optional: specific rule that should fail
}

// TestBaseSchemaValidation tests all base schema validation scenarios
func TestBaseSchemaValidation(t *testing.T) {
	testCases := []TestCase{
		// Valid cases
		{
			Name:       "complete valid document",
			FilePath:   testdataDir + "base/valid/complete.md",
			SchemaPath: testdataDir + ".mdschema.yml",
			ShouldPass: true,
		},
		{
			Name:       "minimal valid document",
			FilePath:   testdataDir + "base/valid/minimal.md",
			SchemaPath: testdataDir + ".mdschema.yml",
			ShouldPass: true,
		},
		{
			Name:       "document with LICENSE",
			FilePath:   testdataDir + "base/valid/with_license.md",
			SchemaPath: testdataDir + ".mdschema.yml",
			ShouldPass: true,
		},
		{
			Name:       "document with nested sections",
			FilePath:   testdataDir + "base/valid/nested.md",
			SchemaPath: testdataDir + ".mdschema.yml",
			ShouldPass: true,
		},

		// Invalid cases
		{
			Name:         "missing root heading",
			FilePath:     testdataDir + "base/invalid/missing_root.md",
			SchemaPath:   testdataDir + ".mdschema.yml",
			ShouldPass:   false,
			ExpectedRule: "structure",
		},
		{
			Name:         "missing Installation section",
			FilePath:     testdataDir + "base/invalid/missing_installation.md",
			SchemaPath:   testdataDir + ".mdschema.yml",
			ShouldPass:   false,
			ExpectedRule: "structure",
		},
		{
			Name:         "missing Usage section",
			FilePath:     testdataDir + "base/invalid/missing_usage.md",
			SchemaPath:   testdataDir + ".mdschema.yml",
			ShouldPass:   false,
			ExpectedRule: "structure",
		},
		{
			Name:         "Installation without bash code",
			FilePath:     testdataDir + "base/invalid/no_bash_code.md",
			SchemaPath:   testdataDir + ".mdschema.yml",
			ShouldPass:   false,
			ExpectedRule: "codeblock",
		},
		{
			Name:         "Usage with insufficient Go code",
			FilePath:     testdataDir + "base/invalid/insufficient_go_code.md",
			SchemaPath:   testdataDir + ".mdschema.yml",
			ShouldPass:   false,
			ExpectedRule: "codeblock",
		},
		{
			Name:         "sections in wrong order",
			FilePath:     testdataDir + "base/invalid/wrong_order.md",
			SchemaPath:   testdataDir + ".mdschema.yml",
			ShouldPass:   false,
			ExpectedRule: "structure",
		},
		{
			Name:         "children outside parent section",
			FilePath:     testdataDir + "base/invalid/children_outside_parent.md",
			SchemaPath:   testdataDir + ".mdschema.yml",
			ShouldPass:   false,
			ExpectedRule: "structure",
		},
		{
			Name:         "root heading pattern mismatch",
			FilePath:     testdataDir + "base/invalid/invalid_root_pattern.md",
			SchemaPath:   testdataDir + ".mdschema.yml",
			ShouldPass:   false,
			ExpectedRule: "structure",
		},
	}

	runTestCases(t, testCases)
}

// Helper function to run a set of test cases
func runTestCases(t *testing.T, testCases []TestCase) {
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Load schema
			s, err := schema.Load(tc.SchemaPath)
			if err != nil {
				t.Fatalf("Failed to load schema %s: %v", tc.SchemaPath, err)
			}

			// Parse document
			p := parser.New()
			doc, err := p.ParseFile(tc.FilePath)
			if err != nil {
				t.Fatalf("Failed to parse document %s: %v", tc.FilePath, err)
			}

			// Validate
			validator := rules.NewValidator()
			violations := validator.Validate(doc, s)

			hasViolations := len(violations) > 0

			if tc.ShouldPass && hasViolations {
				t.Errorf("Expected document to pass validation, but found %d violation(s):", len(violations))
				for _, v := range violations {
					t.Errorf("  - [%s] %s (line %d)", v.Rule, v.Message, v.Line)
				}
			}

			if !tc.ShouldPass && !hasViolations {
				t.Errorf("Expected document to fail validation, but no violations found")
			}

			// Check specific rule if specified
			if tc.ExpectedRule != "" && hasViolations {
				foundExpectedRule := false
				for _, v := range violations {
					if v.Rule == tc.ExpectedRule {
						foundExpectedRule = true
						break
					}
				}
				if !foundExpectedRule {
					t.Errorf("Expected violation from rule '%s', but violations were from: %v",
						tc.ExpectedRule, getUniqueRules(violations))
				}
			}
		})
	}
}

// getUniqueRules returns unique rule names from violations
func getUniqueRules(violations []rules.Violation) []string {
	ruleMap := make(map[string]bool)
	for _, v := range violations {
		ruleMap[v.Rule] = true
	}

	var ruleNames []string
	for rule := range ruleMap {
		ruleNames = append(ruleNames, rule)
	}
	return ruleNames
}
