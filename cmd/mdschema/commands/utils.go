package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jackchuka/mdschema/internal/schema"
)

// loadSchemas loads schemas based on config or discovery
func loadSchemas(cfg *Config) ([]*schema.Schema, error) {
	if len(cfg.SchemaFiles) > 0 {
		// Load explicitly specified schemas
		loaded, err := schema.LoadMultiple(cfg.SchemaFiles)
		if err != nil {
			return nil, err
		}

		return loaded, nil
	}

	// Try to discover schema in current directory
	schemaPath, err := schema.FindSchema(".")
	if err != nil {
		return nil, fmt.Errorf("finding schema: %w", err)
	}

	loaded, err := schema.Load(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("loading discovered schema: %w", err)
	}
	return []*schema.Schema{loaded}, nil
}

const (
	maxFileSize  = 50 * 1024 * 1024 // 50MB per file
	maxFileCount = 1000             // Maximum files to process
)

// findFiles finds all files matching the given glob patterns with validation and limits
// Supports ** for recursive directory matching (e.g., docs/**/*.md)
func findFiles(patterns []string) ([]string, error) {
	seen := make(map[string]bool)
	var files []string

	for _, pattern := range patterns {
		var matches []string
		var err error

		// Check if pattern contains ** for recursive matching
		if containsDoublestar(pattern) {
			matches, err = globWithDoublestar(pattern)
		} else {
			matches, err = filepath.Glob(pattern)
		}

		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern %s: %w", pattern, err)
		}

		for _, match := range matches {
			// Check file count limit early
			if len(files) >= maxFileCount {
				return nil, fmt.Errorf("too many files matched (limit: %d). Use more specific patterns", maxFileCount)
			}

			// Only process .md and .mdx files
			ext := filepath.Ext(match)
			if ext != ".md" && ext != ".mdx" {
				continue
			}

			// Get absolute path for deduplication
			absPath, err := filepath.Abs(match)
			if err != nil {
				return nil, fmt.Errorf("getting absolute path for %s: %w", match, err)
			}

			// Skip if already seen
			if seen[absPath] {
				continue
			}

			// Validate file exists and is accessible
			fileInfo, err := os.Stat(absPath)
			if err != nil {
				if os.IsNotExist(err) {
					continue // Skip non-existent files silently
				}
				if os.IsPermission(err) {
					return nil, fmt.Errorf("permission denied accessing file %s", absPath)
				}
				return nil, fmt.Errorf("error accessing file %s: %w", absPath, err)
			}

			// Skip directories
			if fileInfo.IsDir() {
				continue
			}

			// Check file size limit
			if fileInfo.Size() > maxFileSize {
				return nil, fmt.Errorf("file %s is too large (%d bytes, limit: %d bytes)",
					absPath, fileInfo.Size(), maxFileSize)
			}

			// Validate file is readable
			file, err := os.Open(absPath)
			if err != nil {
				return nil, fmt.Errorf("cannot read file %s: %w", absPath, err)
			}
			_ = file.Close() // Just checking readability

			seen[absPath] = true
			files = append(files, absPath)
		}
	}

	return files, nil
}

// containsDoublestar checks if a pattern contains **
func containsDoublestar(pattern string) bool {
	return len(pattern) >= 2 && (pattern == "**" ||
		(len(pattern) >= 3 && (pattern[:3] == "**/" || pattern[len(pattern)-3:] == "/**")) ||
		(len(pattern) >= 4 && contains(pattern, "/**/")))
}

// contains checks if s contains substr
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && indexOf(s, substr) >= 0)
}

// indexOf returns the index of substr in s, or -1 if not found
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// globWithDoublestar handles glob patterns containing **
// It walks the directory tree and matches files against the pattern
func globWithDoublestar(pattern string) ([]string, error) {
	var matches []string

	// Split pattern at **
	parts := splitDoublestar(pattern)
	if len(parts) == 0 {
		return nil, nil
	}

	// Get the base directory (before **)
	baseDir := parts[0]
	if baseDir == "" {
		baseDir = "."
	}

	// Get the pattern after ** (if any)
	var suffix string
	if len(parts) > 1 {
		suffix = parts[1]
	}

	// Walk the directory tree
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip permission errors and continue walking
			if os.IsPermission(err) {
				return nil
			}
			return err
		}

		// Skip directories for matching (but continue walking into them)
		if info.IsDir() {
			return nil
		}

		// If there's a suffix pattern, match against it
		if suffix != "" {
			// Get the relative path from base
			relPath, err := filepath.Rel(baseDir, path)
			if err != nil {
				return nil
			}

			// Match the filename or relative path against the suffix pattern
			matched, err := filepath.Match(suffix, filepath.Base(path))
			if err != nil {
				return err
			}
			if !matched {
				// Also try matching the full relative path for patterns like **/*.md
				matched, _ = filepath.Match(suffix, relPath)
			}
			if !matched {
				return nil
			}
		}

		matches = append(matches, path)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return matches, nil
}

// splitDoublestar splits a pattern at ** into base directory and suffix
func splitDoublestar(pattern string) []string {
	// Handle different ** positions
	idx := indexOf(pattern, "/**/")
	if idx >= 0 {
		return []string{pattern[:idx], pattern[idx+4:]}
	}

	// Pattern ends with /**
	if len(pattern) >= 3 && pattern[len(pattern)-3:] == "/**" {
		return []string{pattern[:len(pattern)-3], ""}
	}

	// Pattern starts with **/
	if len(pattern) >= 3 && pattern[:3] == "**/" {
		return []string{".", pattern[3:]}
	}

	// Pattern is just **
	if pattern == "**" {
		return []string{".", ""}
	}

	// No ** found (shouldn't happen if containsDoublestar returned true)
	return []string{pattern}
}
