package schema

import "os"

// CreateDefaultFile creates a .mdschema.yml file with default configuration
func CreateDefaultFile(path string) error {
	content := `# Markdown Schema Configuration
# See: https://github.com/jackchuka/mdschema for documentation

structure:
  - heading: "# [a-zA-Z0-9_\\- ]+"
    children: # if children the content must be in this root section
      - heading: "## Features"
        optional: true
      - heading: "## Installation"
        children:
          - heading: "### Windows"
            optional: true
          - heading: "### macOS"
            optional: true
        code_blocks:
          - { lang: bash, min: 1 }
      - heading: "## Usage"
        code_blocks:
          - { lang: go, min: 2 }
  - heading: "# LICENSE"
    optional: true
`
	return os.WriteFile(path, []byte(content), 0o644)
}
