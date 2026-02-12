package integration

import "testing"

// TestStructureValidation tests structure rule validation scenarios
func TestStructureValidation(t *testing.T) {
	testCases := []TestCase{
		// Valid cases
		{
			Name:       "complete valid structure",
			FilePath:   testdataDir + "structure/valid_complete.md",
			SchemaPath: testdataDir + "structure/.mdschema.yml",
			ShouldPass: true,
		},
		{
			Name:       "minimal valid structure",
			FilePath:   testdataDir + "structure/valid_minimal.md",
			SchemaPath: testdataDir + "structure/.mdschema.yml",
			ShouldPass: true,
		},

		// Invalid cases
		{
			Name:         "missing root heading",
			FilePath:     testdataDir + "structure/invalid_missing_root.md",
			SchemaPath:   testdataDir + "structure/.mdschema.yml",
			ShouldPass:   false,
			ExpectedRule: "structure",
		},
		{
			Name:         "missing Installation section",
			FilePath:     testdataDir + "structure/invalid_missing_installation.md",
			SchemaPath:   testdataDir + "structure/.mdschema.yml",
			ShouldPass:   false,
			ExpectedRule: "structure",
		},
		{
			Name:         "missing Usage section",
			FilePath:     testdataDir + "structure/invalid_missing_usage.md",
			SchemaPath:   testdataDir + "structure/.mdschema.yml",
			ShouldPass:   false,
			ExpectedRule: "structure",
		},
		{
			Name:         "sections in wrong order",
			FilePath:     testdataDir + "structure/invalid_wrong_order.md",
			SchemaPath:   testdataDir + "structure/.mdschema.yml",
			ShouldPass:   false,
			ExpectedRule: "structure",
		},
		{
			Name:         "children outside parent section",
			FilePath:     testdataDir + "structure/invalid_children_outside_parent.md",
			SchemaPath:   testdataDir + "structure/.mdschema.yml",
			ShouldPass:   false,
			ExpectedRule: "structure",
		},
		{
			Name:         "root heading pattern mismatch",
			FilePath:     testdataDir + "structure/invalid_root_pattern.md",
			SchemaPath:   testdataDir + "structure/.mdschema.yml",
			ShouldPass:   false,
			ExpectedRule: "structure",
		},

		// No structure defined — should not produce structure violations
		{
			Name:       "no structure violations when structure is omitted from schema",
			FilePath:   testdataDir + "schema_only/valid_any_headings.md",
			SchemaPath: testdataDir + "schema_only/.mdschema.yml",
			ShouldPass: true,
		},
	}

	runTestCases(t, testCases)
}
