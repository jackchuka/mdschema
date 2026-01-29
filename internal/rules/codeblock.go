package rules

import (
	"fmt"
	"strings"

	"github.com/jackchuka/mdschema/internal/schema"
	"github.com/jackchuka/mdschema/internal/vast"
)

// CodeBlockRule validates code blocks using the hierarchical AST
type CodeBlockRule struct {
}

var _ StructuralRule = (*CodeBlockRule)(nil)

// NewCodeBlockRule creates a new simplified code block rule
func NewCodeBlockRule() *CodeBlockRule {
	return &CodeBlockRule{}
}

// Name returns the rule identifier
func (r *CodeBlockRule) Name() string {
	return "codeblock"
}

// validateCodeBlockRequirement validates a specific code block requirement for a node
func (r *CodeBlockRule) validateCodeBlockRequirement(n *vast.Node, requirement schema.CodeBlockRule) []Violation {
	violations := make([]Violation, 0)

	// Count matching code blocks directly from the node
	count := 0
	for _, block := range n.CodeBlocks() {
		if requirement.Lang == "" || block.Lang == requirement.Lang {
			count++
		}
	}

	line, col := n.Location()

	// Check minimum requirement
	if requirement.Min > 0 && count < requirement.Min {
		message := fmt.Sprintf("Section '%s' requires at least %d code blocks", n.HeadingText(), requirement.Min)
		if requirement.Lang != "" {
			message = fmt.Sprintf("Section '%s' requires at least %d '%s' code blocks, found %d",
				n.HeadingText(), requirement.Min, requirement.Lang, count)
		}

		violations = append(violations, NewViolation(r.Name(), message, line, col))
	}

	// Check maximum requirement
	if requirement.Max > 0 && count > requirement.Max {
		message := fmt.Sprintf("Section '%s' has too many code blocks (max %d)", n.HeadingText(), requirement.Max)
		if requirement.Lang != "" {
			message = fmt.Sprintf("Section '%s' has too many '%s' code blocks (max %d, found %d)",
				n.HeadingText(), requirement.Lang, requirement.Max, count)
		}

		violations = append(violations, NewViolation(r.Name(), message, line, col))
	}

	return violations
}

// ValidateWithContext validates using VAST (validation-ready AST)
func (r *CodeBlockRule) ValidateWithContext(ctx *vast.Context) []Violation {
	violations := make([]Violation, 0)

	// Walk through all bound nodes to find elements with code block rules
	ctx.Tree.WalkBound(func(n *vast.Node) bool {
		if n.Element.SectionRules != nil && len(n.Element.CodeBlocks) > 0 {
			for _, requirement := range n.Element.CodeBlocks {
				violations = append(violations, r.validateCodeBlockRequirement(n, requirement)...)
			}
		}
		return true
	})

	return violations
}

// commentForLang returns an appropriate TODO comment for the given language
func commentForLang(lang string) string {
	switch lang {
	case "bash", "sh", "shell", "zsh", "yaml", "yml", "python", "py", "ruby", "rb", "perl", "r", "toml", "dockerfile":
		return "# TODO: Add your code here"
	case "html", "xml", "svg":
		return "<!-- TODO: Add your code here -->"
	case "css", "scss", "sass", "less":
		return "/* TODO: Add your code here */"
	case "sql", "lua", "haskell", "hs":
		return "-- TODO: Add your code here"
	default:
		// go, js, ts, java, c, cpp, rust, swift, kotlin, etc.
		return "// TODO: Add your code here"
	}
}

// GenerateContent generates placeholder code blocks for code block rules
func (r *CodeBlockRule) GenerateContent(builder *strings.Builder, element schema.StructureElement) bool {
	if element.SectionRules == nil || len(element.CodeBlocks) == 0 {
		return false
	}

	// Add code block placeholders
	builder.WriteString("<!-- Code block requirements: -->\n")
	for _, rule := range element.CodeBlocks {
		if rule.Lang != "" {
			if rule.Min > 0 {
				fmt.Fprintf(builder, "<!-- Minimum %d %s code blocks required -->\n", rule.Min, rule.Lang)
			}
			if rule.Max > 0 {
				fmt.Fprintf(builder, "<!-- Maximum %d %s code blocks allowed -->\n", rule.Max, rule.Lang)
			}
		} else {
			if rule.Min > 0 {
				fmt.Fprintf(builder, "<!-- Minimum %d code blocks required -->\n", rule.Min)
			}
			if rule.Max > 0 {
				fmt.Fprintf(builder, "<!-- Maximum %d code blocks allowed -->\n", rule.Max)
			}
		}
	}
	builder.WriteString("\n")

	// Add placeholder code blocks
	for _, rule := range element.CodeBlocks {
		if rule.Min > 0 {
			lang := rule.Lang
			if lang == "" {
				lang = "text"
			}
			// Generate the minimum required number of code blocks
			for i := 0; i < rule.Min; i++ {
				fmt.Fprintf(builder, "```%s\n", lang)
				builder.WriteString(commentForLang(lang) + "\n")
				builder.WriteString("```\n\n")
			}
		}
	}

	return true
}
