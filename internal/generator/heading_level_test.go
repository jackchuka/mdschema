package generator

import (
	"strings"
	"testing"

	"github.com/jackchuka/mdschema/internal/schema"
)

func TestGenerateChildHeadingLevel(t *testing.T) {
	g := New()

	s := &schema.Schema{
		Structure: []schema.StructureElement{
			{
				Heading: schema.HeadingPattern{Pattern: "# Title"},
				Children: []schema.StructureElement{
					{
						Heading:  schema.HeadingPattern{Pattern: "## Child"},
						Children: []schema.StructureElement{{Heading: schema.HeadingPattern{Pattern: "### Grandchild"}}},
					},
				},
			},
		},
	}

	output := g.Generate(s, "")

	// The parent's comment also lists "## Child", so check the heading lines themselves.
	for _, heading := range []string{"\n## Child\n", "\n### Grandchild\n"} {
		if !strings.Contains(output, heading) {
			t.Errorf("Generated output should contain %q one level below its parent, got:\n%s", strings.TrimSpace(heading), output)
		}
	}
}
