package schema

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Load reads and parses a schema file
func Load(path string) (*Schema, error) {
	return simpleLoadYAML(path)
}

// simpleLoadYAML loads a schema from a YAML file
func simpleLoadYAML(path string) (*Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading schema file: %w", err)
	}

	var schema Schema
	if err := yaml.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("parsing schema YAML: %w", err)
	}

	return &schema, nil
}

// LoadMultiple loads multiple schemas
func LoadMultiple(paths []string) ([]*Schema, error) {
	schemas := make([]*Schema, 0, len(paths))
	for _, path := range paths {
		schema, err := Load(path)
		if err != nil {
			return nil, fmt.Errorf("loading schema %s: %w", path, err)
		}
		schemas = append(schemas, schema)
	}
	return schemas, nil
}

// FindSchema discovers schema files in the directory hierarchy
func FindSchema(startPath string) (string, error) {
	dir := startPath
	if !isDir(dir) {
		dir = filepath.Dir(dir)
	}

	for {
		schemaPath := filepath.Join(dir, ".mdschema.yml")
		if _, err := os.Stat(schemaPath); err == nil {
			return schemaPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("no .mdschema.yml found in directory hierarchy")
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
