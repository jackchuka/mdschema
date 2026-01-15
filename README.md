# mdschema

[![Test](https://github.com/jackchuka/mdschema/workflows/Test/badge.svg)](https://github.com/jackchuka/mdschema/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/jackchuka/mdschema)](https://goreportcard.com/report/github.com/jackchuka/mdschema)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A declarative schema-based Markdown documentation validator that helps maintain consistent documentation structure across projects.

This README file itself is an example of how to use mdschema to validate and generate documentation.

```bash
mdschema check README.md --schema ./examples/README-schema.yml
✓ No violations found
```

## Features

- **Schema-driven validation** - Define your documentation structure in simple YAML
- **Hierarchical structure** - Support for nested sections and complex document layouts
- **Template generation** - Generate markdown templates from your schemas
- **Comprehensive rules** - Validate headings, code blocks, images, tables, lists, links, and more
- **Frontmatter validation** - Validate YAML frontmatter with type and format checking
- **Link validation** - Check internal anchors, relative files, and external URLs
- **Context-aware** - Uses AST parsing for accurate validation without string matching
- **Fast and lightweight** - Single binary with no dependencies
- **Cross-platform** - Works on Linux, macOS, and Windows

## Installation

### Go Install

```bash
go install github.com/jackchuka/mdschema/cmd/mdschema@latest
```

### From Source

```bash
git clone https://github.com/jackchuka/mdschema.git
cd mdschema
go build -o mdschema ./cmd/mdschema
```

## Quick Start

1. **Initialize a schema** in your project:

```bash
mdschema init
```

2. **Validate your markdown files**:

```bash
mdschema check README.md docs/*.md
```

3. **Generate a template** from your schema:

```bash
mdschema generate -o new-doc.md
```

## Schema Format

Create a `.mdschema.yml` file to define your documentation structure:

```yaml
structure:
  - heading:
      pattern: "# [a-zA-Z0-9_\\- ]+" # Regex pattern for project title
      regex: true
    children:
      - heading: "## Features"
        optional: true
      - heading: "## Installation"
        code_blocks:
          - { lang: bash, min: 1 } # Require at least 1 bash code block
        children:
          - heading: "### Windows"
            optional: true
          - heading: "### macOS"
            optional: true
      - heading: "## Usage"
        code_blocks:
          - { lang: go, min: 2 } # Require at least 2 Go code blocks
        required_text:
          - "example" # Must contain the word "example"
  - heading: "# LICENSE"
    optional: true
```

### Schema Elements

#### Structure Elements

- **`heading`** - Heading pattern (string for literal, or `{pattern: "...", regex: true}` for regex)
- **`optional`** - Whether the section is optional (default: false)
- **`children`** - Nested subsections that must appear within this section

#### Section Rules (apply to each section)

- **`required_text`** - Text that must appear (`"text"` or `{pattern: "...", regex: true}`)
- **`forbidden_text`** - Text that must NOT appear (`"text"` or `{pattern: "...", regex: true}`)
- **`code_blocks`** - Code block requirements: `{lang: "bash", min: 1, max: 3}`
- **`images`** - Image requirements: `{min: 1, require_alt: true, formats: ["png", "svg"]}`
- **`tables`** - Table requirements: `{min: 1, min_columns: 2, required_headers: ["Name"]}`
- **`lists`** - List requirements: `{min: 1, type: "ordered", min_items: 3}`
- **`word_count`** - Word count constraints: `{min: 50, max: 500}`

#### Global Rules (apply to entire document)

- **`links`** - Link validation (internal anchors, relative files, external URLs)
- **`heading_rules`** - Heading constraints (no skipped levels, unique headings, max depth)
- **`frontmatter`** - YAML frontmatter validation (required fields, types, formats)

## Commands

### `check` - Validate Documents

```bash
mdschema check README.md docs/**/*.md
mdschema check --schema custom.yml *.md
```

### `generate` - Create Templates

```bash
# Generate from .mdschema.yml
mdschema generate
# Generate from specific schema file
mdschema generate --schema custom.yml
# Generate and save to file
mdschema generate -o template.md
```

### `init` - Initialize Schema

```bash
# Create .mdschema.yml with defaults
mdschema init
```

### `derive` - Infer Schema from Document

```bash
# Infer schema from existing markdown, output to stdout
mdschema derive README.md

# Save inferred schema to a file
mdschema derive README.md -o inferred-schema.yml
```

## Examples

### Basic README Schema

```yaml
structure:
  - heading:
      pattern: "# .*"
      regex: true
    children:
      - heading: "## Installation"
        code_blocks:
          - { lang: bash, min: 1 }
      - heading: "## Usage"
        code_blocks:
          - { lang: go, min: 1 }
```

### API Documentation Schema

```yaml
structure:
  - heading: "# API Reference"
    children:
      - heading: "## Authentication"
        required_text: ["API key", "Bearer token"]
      - heading: "## Endpoints"
        children:
          - heading: "### GET /users"
            code_blocks:
              - { lang: json, min: 1 }
              - { lang: curl, min: 1 }
```

### Tutorial Schema

```yaml
structure:
  - heading:
      pattern: "# .*"
      regex: true
    children:
      - heading: "## Prerequisites"
      - heading:
          pattern: "## Step 1: .*"
          regex: true
        code_blocks:
          - { min: 1 } # Any language
      - heading:
          pattern: "## Step 2: .*"
          regex: true
        code_blocks:
          - { min: 1 }
      - heading: "## Next Steps"
        optional: true
```

### Blog Post Schema (comprehensive example)

```yaml
# Global rules
frontmatter:
  # optional: false is default, meaning frontmatter is required
  fields:
    - { name: "title" } # required by default
    - { name: "date", format: date } # required by default
    - { name: "author", optional: true, format: email }
    - { name: "tags", optional: true, type: array }

heading_rules:
  no_skip_levels: true
  max_depth: 3

links:
  validate_internal: true
  validate_files: true

# Document structure
structure:
  - heading:
      pattern: "# .*"
      regex: true
    children:
      - heading: "## Introduction"
        word_count: { min: 100, max: 300 }
        forbidden_text: ["TODO", "FIXME"]
      - heading: "## Content"
        images:
          - { min: 1, require_alt: true }
        code_blocks:
          - { min: 1 }
      - heading: "## Conclusion"
        word_count: { min: 50 }
        lists:
          - { min: 1, type: unordered }
```

## Validation Rules

mdschema includes comprehensive validation rules organized into three categories:

### Section Rules (per-section validation)

| Rule               | Description                                        | Options                                         |
| ------------------ | -------------------------------------------------- | ----------------------------------------------- |
| **Structure**      | Ensures sections appear in correct order/hierarchy | `heading`, `optional`, `children`               |
| **Required Text**  | Text/patterns that must appear                     | `pattern`, `regex`                              |
| **Forbidden Text** | Text/patterns that must NOT appear                 | `pattern`, `regex`                              |
| **Code Blocks**    | Code block requirements                            | `lang`, `min`, `max`                            |
| **Images**         | Image presence and format                          | `min`, `max`, `require_alt`, `formats`          |
| **Tables**         | Table structure validation                         | `min`, `max`, `min_columns`, `required_headers` |
| **Lists**          | List presence and type                             | `min`, `max`, `type`, `min_items`               |
| **Word Count**     | Content length constraints                         | `min`, `max`                                    |

### Global Rules (document-wide validation)

#### Link Validation

```yaml
links:
  validate_internal: true # Check anchor links (#section)
  validate_files: true # Check relative file links (./file.md)
  validate_external: false # Check external URLs (slower)
  external_timeout: 10 # Timeout in seconds
  allowed_domains: # Restrict to these domains
    - github.com
    - golang.org
  blocked_domains: # Block these domains
    - example.com
```

#### Heading Rules

```yaml
heading_rules:
  no_skip_levels: true # Disallow h1 -> h3 without h2
  unique: true # All headings must be unique
  unique_per_level: false # Unique within same level only
  max_depth: 4 # Maximum heading depth (h4)
```

#### Frontmatter Validation

```yaml
frontmatter:
  optional: true # Set to make frontmatter optional (default: required)
  fields:
    - { name: "title", type: string } # required by default
    - { name: "date", type: date, format: date } # required by default
    - { name: "author", optional: true, format: email } # explicitly optional
    - { name: "tags", optional: true, type: array }
    - { name: "draft", optional: true, type: boolean }
    - { name: "version", optional: true, type: number }
    - { name: "repo", optional: true, format: url }
```

**Field types:** `string`, `number`, `boolean`, `array`, `date`
**Field formats:** `date` (YYYY-MM-DD), `email`, `url`

## Use Cases

- **Documentation Standards** - Enforce consistent README structure across repositories
- **API Documentation** - Ensure all endpoints have required sections and examples
- **Tutorial Validation** - Verify step-by-step guides follow the expected format
- **CI/CD Integration** - Validate documentation in pull requests
- **Template Generation** - Create starter templates for new projects

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o mdschema ./cmd/mdschema
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change. See [CONTRIBUTING.md](CONTRIBUTING.md) for more details.

## License

MIT License - see [LICENSE](LICENSE) for details.
