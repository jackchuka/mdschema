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

func TestEnumTypeMismatchWarnings(t *testing.T) {
	tests := []struct {
		name         string
		data         string
		wantMessages []string
	}{
		{
			name: "every bad entry is reported",
			data: "frontmatter:\n  fields:\n    - name: priority\n      type: number\n      enum: [low, high]\n",
			wantMessages: []string{
				`enum value "low" for field "priority" is not a number`,
				`enum value "high" for field "priority" is not a number`,
			},
		},
		{
			name: "one bad entry among good ones",
			data: "frontmatter:\n  fields:\n    - name: priority\n      type: number\n      enum: [1, 2, three]\n",
			wantMessages: []string{
				`enum value "three" for field "priority" is not a number`,
			},
		},
		{
			name: "number under a string field",
			data: "frontmatter:\n  fields:\n    - name: status\n      type: string\n      enum: [draft, 2]\n",
			wantMessages: []string{
				`enum value "2" for field "status" is not a string`,
			},
		},
		{
			name: "non-date under a date field",
			data: "frontmatter:\n  fields:\n    - name: released\n      type: date\n      enum: [2024-01-01, soon]\n",
			wantMessages: []string{
				`enum value "soon" for field "released" is not a date`,
			},
		},
		{
			// Sequence and mapping nodes have an empty Value, so they are named
			// by shape rather than quoted as an empty string.
			name: "list entry is described by shape",
			data: "frontmatter:\n  fields:\n    - name: status\n      type: string\n      enum:\n        - [a, b]\n",
			wantMessages: []string{
				`enum value for field "status" is a list, not a string`,
			},
		},
		{
			name: "mapping entry is described by shape",
			data: "frontmatter:\n  fields:\n    - name: status\n      type: string\n      enum:\n        - {a: b}\n",
			wantMessages: []string{
				`enum value for field "status" is a mapping, not a string`,
			},
		},
		{
			// A format constrains the value type even with no explicit type.
			name: "non-string under an email format",
			data: "frontmatter:\n  fields:\n    - name: contact\n      format: email\n      enum: [1, 2]\n",
			wantMessages: []string{
				`enum value "1" for field "contact" is not a string`,
				`enum value "2" for field "contact" is not a string`,
			},
		},
		{
			name: "non-date under a date format",
			data: "frontmatter:\n  fields:\n    - name: released\n      format: date\n      enum: [2024-01-01, soon]\n",
			wantMessages: []string{
				`enum value "soon" for field "released" is not a date`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings, err := checkUnknownKeys([]byte(tt.data))
			if err != nil {
				t.Fatalf("checkUnknownKeys() error: %v", err)
			}
			if len(warnings) != len(tt.wantMessages) {
				t.Fatalf("expected %d warnings, got %d: %+v", len(tt.wantMessages), len(warnings), warnings)
			}
			for i, want := range tt.wantMessages {
				if got := warnings[i].Message; got != want {
					t.Errorf("message[%d] = %q, want %q", i, got, want)
				}
				if warnings[i].Line == 0 {
					t.Errorf("warning[%d]: expected a non-zero line number", i)
				}
			}
		})
	}
}

func TestEnumTypeConsistentNoWarnings(t *testing.T) {
	// Array fields constrain their elements, not the field itself, and
	// untyped fields constrain nothing — neither should warn.
	data := []byte(`frontmatter:
  fields:
    - name: status
      type: string
      enum: [draft, published]
    - name: priority
      type: number
      enum: [1, 2.5]
    - name: draft
      type: boolean
      enum: [true, false]
    - name: released
      type: date
      enum: [2024-01-01]
    - name: tags
      type: array
      enum: [go, cli]
    - name: anything
      enum: [1, foo, true]
    - name: contact
      format: email
      enum: [user@example.com]
    - name: homepage
      format: url
      enum: [https://example.com]
    - name: shipped
      format: date
      enum: [2024-01-01]
`)
	warnings, err := checkUnknownKeys(data)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected 0 warnings, got %+v", warnings)
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
