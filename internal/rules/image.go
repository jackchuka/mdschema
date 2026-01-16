package rules

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jackchuka/mdschema/internal/schema"
	"github.com/jackchuka/mdschema/internal/vast"
)

// ImageRule validates images using the hierarchical AST
type ImageRule struct {
}

var _ StructuralRule = (*ImageRule)(nil)

// NewImageRule creates a new image rule
func NewImageRule() *ImageRule {
	return &ImageRule{}
}

// Name returns the rule identifier
func (r *ImageRule) Name() string {
	return "image"
}

// validateImageRequirement validates a specific image requirement for a node
func (r *ImageRule) validateImageRequirement(n *vast.Node, requirement schema.ImageRule) []Violation {
	violations := make([]Violation, 0)
	images := n.Images()

	line, col := n.Location()

	// Count images (optionally filtering by format)
	count := 0
	for _, img := range images {
		if len(requirement.Formats) > 0 {
			ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(img.URL), "."))
			matched := false
			for _, format := range requirement.Formats {
				if strings.EqualFold(ext, format) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		count++
	}

	// Check minimum requirement
	if requirement.Min > 0 && count < requirement.Min {
		message := fmt.Sprintf("Section '%s' requires at least %d images, found %d",
			n.HeadingText(), requirement.Min, count)
		if len(requirement.Formats) > 0 {
			message = fmt.Sprintf("Section '%s' requires at least %d images (%s), found %d",
				n.HeadingText(), requirement.Min, strings.Join(requirement.Formats, ", "), count)
		}

		violations = append(violations, NewViolation(r.Name(), message, line, col))
	}

	// Check maximum requirement
	if requirement.Max > 0 && count > requirement.Max {
		message := fmt.Sprintf("Section '%s' has too many images (max %d, found %d)",
			n.HeadingText(), requirement.Max, count)

		violations = append(violations, NewViolation(r.Name(), message, line, col))
	}

	// Check alt text requirement
	if requirement.RequireAlt {
		for _, img := range images {
			if strings.TrimSpace(img.Alt) == "" {
				violations = append(violations,
					NewViolation(r.Name(), fmt.Sprintf("Image in section '%s' is missing alt text", n.HeadingText()), img.Line, img.Column))
			}
		}
	}

	return violations
}

// ValidateWithContext validates using VAST (validation-ready AST)
func (r *ImageRule) ValidateWithContext(ctx *vast.Context) []Violation {
	violations := make([]Violation, 0)

	// Walk through all bound nodes to find elements with image rules
	ctx.Tree.WalkBound(func(n *vast.Node) bool {
		if n.Element.SectionRules != nil && len(n.Element.Images) > 0 {
			for _, requirement := range n.Element.Images {
				violations = append(violations, r.validateImageRequirement(n, requirement)...)
			}
		}
		return true
	})

	return violations
}

// GenerateContent generates placeholder images for image rules
func (r *ImageRule) GenerateContent(builder *strings.Builder, element schema.StructureElement) bool {
	if element.SectionRules == nil || len(element.Images) == 0 {
		return false
	}

	// Add image requirement comments
	builder.WriteString("<!-- Image requirements: -->\n")
	for _, rule := range element.Images {
		if rule.Min > 0 {
			fmt.Fprintf(builder, "<!-- Minimum %d images required -->\n", rule.Min)
		}
		if rule.Max > 0 {
			fmt.Fprintf(builder, "<!-- Maximum %d images allowed -->\n", rule.Max)
		}
		if rule.RequireAlt {
			builder.WriteString("<!-- All images must have alt text -->\n")
		}
		if len(rule.Formats) > 0 {
			fmt.Fprintf(builder, "<!-- Allowed formats: %s -->\n", strings.Join(rule.Formats, ", "))
		}
	}
	builder.WriteString("\n")

	// Add placeholder images
	for _, rule := range element.Images {
		if rule.Min > 0 {
			for i := 0; i < rule.Min; i++ {
				builder.WriteString("![TODO: Add alt text](path/to/image.png)\n\n")
			}
		}
	}

	return true
}
