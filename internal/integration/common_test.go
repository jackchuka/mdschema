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

// runTestCases is a helper function to run a set of test cases
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
