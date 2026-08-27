package vast

import "testing"

func TestEvaluateHeadingExpr(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		filename   string
		heading    string
		level      int
		want       bool
	}{
		{
			name:       "plain equality holds when heading equals filename",
			expression: "filename == heading",
			filename:   "moveInventoryLots",
			heading:    "moveInventoryLots",
			want:       true,
		},
		{
			name:       "plain equality fails when heading differs from filename",
			expression: "filename == heading",
			filename:   "moveInventoryLots",
			heading:    "SomethingElse",
			want:       false,
		},
		{
			name:       "same transform on both sides holds when heading equals filename",
			expression: "slug(filename) == slug(heading)",
			filename:   "CreateOrder",
			heading:    "CreateOrder",
			want:       true,
		},
		{
			name:       "same transform on both sides also holds for differently-formatted equivalents",
			expression: "slug(filename) == slug(heading)",
			filename:   "Create Order",
			heading:    "create-order",
			want:       true,
		},
		{
			name:       "an expression unrelated to filename/heading equality is evaluated as-is",
			expression: "heading == \"FixedTitle\"",
			filename:   "anything",
			heading:    "FixedTitle",
			want:       true,
		},
		{
			name:       "an expression unrelated to filename/heading equality correctly fails",
			expression: "heading == \"FixedTitle\"",
			filename:   "anything",
			heading:    "SomethingElse",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EvaluateHeadingExpr(tt.expression, tt.filename, tt.heading, tt.level)
			if err != nil {
				t.Fatalf("EvaluateHeadingExpr() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("EvaluateHeadingExpr(%q, filename=%q, heading=%q) = %v, want %v",
					tt.expression, tt.filename, tt.heading, got, tt.want)
			}
		})
	}
}

func TestEvaluateHeadingExprInvalidExpression(t *testing.T) {
	_, err := EvaluateHeadingExpr("not a valid expr (", "f", "h", 1)
	if err == nil {
		t.Error("EvaluateHeadingExpr() expected an error for an invalid expression, got nil")
	}
}
