package integration

import "testing"

// TestParagraphValidation tests paragraph rule validation scenarios
func TestParagraphValidation(t *testing.T) {
	testCases := []TestCase{
		// Valid cases
		{
			Name:       "valid section with paragraph",
			FilePath:   testdataDir + "paragraph/valid_has_paragraph.md",
			SchemaPath: testdataDir + "paragraph/.mdschema.yml",
			ShouldPass: true,
		},

		// Invalid cases
		{
			Name:         "Overview with only a bullet list, no paragraph",
			FilePath:     testdataDir + "paragraph/invalid_no_paragraph.md",
			SchemaPath:   testdataDir + "paragraph/.mdschema.yml",
			ShouldPass:   false,
			ExpectedRule: "paragraph",
		},
	}

	runTestCases(t, testCases)
}
