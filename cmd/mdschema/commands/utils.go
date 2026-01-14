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
func findFiles(patterns []string) ([]string, error) {
	seen := make(map[string]bool)
	var files []string

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
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
