package rules

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackchuka/mdschema/internal/parser"
	"github.com/jackchuka/mdschema/internal/schema"
)

// LinkValidationRule validates internal and external links in the document
type LinkValidationRule struct {
}

var _ ContextualRule = (*LinkValidationRule)(nil)

// NewLinkValidationRule creates a new link validation rule
func NewLinkValidationRule() *LinkValidationRule {
	return &LinkValidationRule{}
}

// Name returns the rule identifier
func (r *LinkValidationRule) Name() string {
	return "link"
}

// ValidateWithContext validates links based on schema configuration
func (r *LinkValidationRule) ValidateWithContext(ctx *ValidationContext) []Violation {
	violations := make([]Violation, 0)

	// Skip if no link rules configured
	if ctx.Schema.Links == nil {
		return violations
	}

	linkRule := ctx.Schema.Links

	// Collect all links from the document
	links := r.collectAllLinks(ctx.Document.Root)

	// Build a map of valid heading slugs
	slugs := r.buildSlugMap(ctx.Document)

	// Get document directory for relative path resolution
	docDir := filepath.Dir(ctx.Document.Path)

	for _, link := range links {
		violations = append(violations, r.validateLink(link, linkRule, slugs, docDir)...)
	}

	return violations
}

// collectAllLinks recursively collects all links from sections
func (r *LinkValidationRule) collectAllLinks(section *parser.Section) []*parser.Link {
	links := make([]*parser.Link, 0)
	links = append(links, section.Links...)

	for _, child := range section.Children {
		links = append(links, r.collectAllLinks(child)...)
	}

	return links
}

// buildSlugMap creates a map of valid heading slugs from the document
func (r *LinkValidationRule) buildSlugMap(doc *parser.Document) map[string]bool {
	slugs := make(map[string]bool)

	sections := doc.GetSections()
	for _, section := range sections {
		if section.Heading != nil && section.Heading.Slug != "" {
			slugs[section.Heading.Slug] = true
		}
	}

	return slugs
}

// validateLink validates a single link according to the rules
func (r *LinkValidationRule) validateLink(link *parser.Link, rule *schema.LinkRule, slugs map[string]bool, docDir string) []Violation {
	violations := make([]Violation, 0)

	url := link.URL

	// Anchor links (#section-name)
	if anchor, found := strings.CutPrefix(url, "#"); found {
		if rule.ValidateInternal {
			if !slugs[anchor] {
				violations = append(violations, Violation{
					Rule:    r.Name(),
					Message: fmt.Sprintf("Broken internal link: anchor '%s' does not exist in the document", url),
					Line:    link.Line,
					Column:  link.Column,
				})
			}
		}
		return violations
	}

	// External links (http/https)
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		violations = append(violations, r.validateExternalLink(link, rule)...)
		return violations
	}

	// File links (relative paths)
	if rule.ValidateFiles {
		violations = append(violations, r.validateFileLink(link, docDir)...)
	}

	return violations
}

// validateExternalLink validates an external URL
func (r *LinkValidationRule) validateExternalLink(link *parser.Link, rule *schema.LinkRule) []Violation {
	violations := make([]Violation, 0)

	parsedURL, err := url.Parse(link.URL)
	if err != nil {
		violations = append(violations, Violation{
			Rule:    r.Name(),
			Message: fmt.Sprintf("Invalid URL format: %s", link.URL),
			Line:    link.Line,
			Column:  link.Column,
		})
		return violations
	}

	host := parsedURL.Hostname()

	// Check blocked domains
	if len(rule.BlockedDomains) > 0 {
		for _, blocked := range rule.BlockedDomains {
			if strings.EqualFold(host, blocked) || strings.HasSuffix(strings.ToLower(host), "."+strings.ToLower(blocked)) {
				violations = append(violations, Violation{
					Rule:    r.Name(),
					Message: fmt.Sprintf("Link to blocked domain: %s", host),
					Line:    link.Line,
					Column:  link.Column,
				})
				return violations
			}
		}
	}

	// Check allowed domains (if configured, link must be to one of these)
	if len(rule.AllowedDomains) > 0 {
		allowed := false
		for _, domain := range rule.AllowedDomains {
			if strings.EqualFold(host, domain) || strings.HasSuffix(strings.ToLower(host), "."+strings.ToLower(domain)) {
				allowed = true
				break
			}
		}
		if !allowed {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Message: fmt.Sprintf("Link to domain '%s' is not in the allowed domains list", host),
				Line:    link.Line,
				Column:  link.Column,
			})
			return violations
		}
	}

	// Validate external URL accessibility
	if rule.ValidateExternal {
		timeout := rule.ExternalTimeout
		if timeout <= 0 {
			timeout = 10 // Default timeout
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodHead, link.URL, nil)
		if err != nil {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Message: fmt.Sprintf("Failed to create request for URL: %s", link.URL),
				Line:    link.Line,
				Column:  link.Column,
			})
			return violations
		}

		req.Header.Set("User-Agent", "mdschema-link-validator/1.0")

		client := &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		}

		resp, err := client.Do(req)
		if err != nil {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Message: fmt.Sprintf("Failed to reach URL '%s': %v", link.URL, err),
				Line:    link.Line,
				Column:  link.Column,
			})
			return violations
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		if resp.StatusCode >= 400 {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Message: fmt.Sprintf("URL '%s' returned status %d", link.URL, resp.StatusCode),
				Line:    link.Line,
				Column:  link.Column,
			})
		}
	}

	return violations
}

// validateFileLink validates a relative file path link
func (r *LinkValidationRule) validateFileLink(link *parser.Link, docDir string) []Violation {
	violations := make([]Violation, 0)

	// Parse URL to handle anchors in file links (e.g., ./other.md#section)
	linkURL := link.URL
	if idx := strings.Index(linkURL, "#"); idx != -1 {
		linkURL = linkURL[:idx]
	}

	// Skip empty paths (just anchors)
	if linkURL == "" {
		return violations
	}

	// Resolve relative path
	targetPath := filepath.Join(docDir, linkURL)
	targetPath = filepath.Clean(targetPath)

	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		violations = append(violations, Violation{
			Rule:    r.Name(),
			Message: fmt.Sprintf("Broken file link: '%s' does not exist", link.URL),
			Line:    link.Line,
			Column:  link.Column,
		})
	}

	return violations
}

// GenerateContent generates placeholder content (links don't generate content)
func (r *LinkValidationRule) GenerateContent(builder *strings.Builder, element schema.StructureElement) bool {
	// Link validation rule doesn't generate content
	return false
}
