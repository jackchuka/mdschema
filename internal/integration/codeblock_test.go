package integration

import "testing"

// TestCodeblockValidation tests codeblock rule validation scenarios
func TestCodeblockValidation(t *testing.T) {
	testCases := []TestCase{
		// Valid cases
		{
			Name:       "valid code blocks",
			FilePath:   testdataDir + "codeblock/valid_complete.md",
			SchemaPath: testdataDir + "codeblock/.mdschema.yml",
			ShouldPass: true,
		},

		// Invalid cases
		{
			Name:         "Installation without bash code",
			FilePath:     testdataDir + "codeblock/invalid_no_bash_code.md",
			SchemaPath:   testdataDir + "codeblock/.mdschema.yml",
			ShouldPass:   false,
			ExpectedRule: "codeblock",
		},
		{
			Name:         "Usage with insufficient Go code",
			FilePath:     testdataDir + "codeblock/invalid_insufficient_go_code.md",
			SchemaPath:   testdataDir + "codeblock/.mdschema.yml",
			ShouldPass:   false,
			ExpectedRule: "codeblock",
		},
	}

	runTestCases(t, testCases)
}
