package schema

import "testing"

func TestUnknownKeyInStructureElement(t *testing.T) {
	data := []byte(`structure:
  - heading: "## Overview"
    paragraphs:
      min: 1
    paragrafs:
      min: 1
`)
	warnings, err := checkUnknownKeys(data)
	if err != nil {
		t.Fatalf("checkUnknownKeys() error: %v", err)
	}
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d: %+v", len(warnings), warnings)
	}
	if got := warnings[0].Message; got != `unknown key "paragrafs" (ignored)` {
		t.Errorf("message = %q", got)
	}
	if warnings[0].Line == 0 {
		t.Error("expected a non-zero line number")
	}
}

func TestUnknownKeyAtTopLevel(t *testing.T) {
	data := []byte("structure: []\nbogus: true\n")
	warnings, err := checkUnknownKeys(data)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(warnings) != 1 || warnings[0].Message != `unknown key "bogus" (ignored)` {
		t.Fatalf("got %+v", warnings)
	}
}

func TestKnownKeysNoWarnings(t *testing.T) {
	data := []byte(`structure:
  - heading:
      pattern: "## .*"
      regex: true
    optional: true
    word_count:
      min: 1
    required_text:
      - pattern: "foo"
    children:
      - heading: "### Sub"
frontmatter:
  fields:
    - name: title
      type: string
heading_rules:
  no_skip_levels: true
`)
	warnings, err := checkUnknownKeys(data)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected 0 warnings, got %+v", warnings)
	}
}
