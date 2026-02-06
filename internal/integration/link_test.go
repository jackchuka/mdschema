package integration

import "testing"

// TestLinkValidation tests link validation scenarios
func TestLinkValidation(t *testing.T) {
	testCases := []TestCase{
		// Valid cases
		{
			Name:       "valid internal anchor links",
			FilePath:   testdataDir + "links/valid_internal.md",
			SchemaPath: testdataDir + "links/.mdschema.yml",
			ShouldPass: true,
		},
		{
			Name:       "valid file links",
			FilePath:   testdataDir + "links/valid_file.md",
			SchemaPath: testdataDir + "links/.mdschema.yml",
			ShouldPass: true,
		},
		{
			Name:       "valid external links",
			FilePath:   testdataDir + "links/valid_external.md",
			SchemaPath: testdataDir + "links/.mdschema.yml",
			ShouldPass: true,
		},

		// Invalid cases
		{
			Name:         "broken internal anchor",
			FilePath:     testdataDir + "links/invalid_broken_anchor.md",
			SchemaPath:   testdataDir + "links/.mdschema.yml",
			ShouldPass:   false,
			ExpectedRule: "link",
		},
		{
			Name:         "broken file link",
			FilePath:     testdataDir + "links/invalid_broken_file.md",
			SchemaPath:   testdataDir + "links/.mdschema.yml",
			ShouldPass:   false,
			ExpectedRule: "link",
		},
		{
			Name:         "blocked domain link",
			FilePath:     testdataDir + "links/invalid_blocked_domain.md",
			SchemaPath:   testdataDir + "links/.mdschema.yml",
			ShouldPass:   false,
			ExpectedRule: "link",
		},
		{
			Name:         "blocked domain in frontmatter URL",
			FilePath:     testdataDir + "links/invalid_frontmatter_blocked_domain.md",
			SchemaPath:   testdataDir + "links/.mdschema.yml",
			ShouldPass:   false,
			ExpectedRule: "link",
		},
	}

	runTestCases(t, testCases)
}
