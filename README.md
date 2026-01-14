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

- 🔍 **Schema-driven validation** - Define your documentation structure in simple YAML
- 🌳 **Hierarchical structure** - Support for nested sections and complex document layouts
- 📝 **Template generation** - Generate markdown templates from your schemas
- 🔧 **Rule-based validation** - Validate headings, code blocks, required text, and structure order
- 🎯 **Context-aware** - Uses AST parsing for accurate validation without string matching guesswork
- 🚀 **Fast and lightweight** - Single binary with no dependencies
- 💻 **Cross-platform** - Works on Linux, macOS, and Windows

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
  - heading: "# [a-zA-Z0-9_\\- ]+" # Regex pattern for project title
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

- **`heading`** - Heading pattern (supports regex)
- **`optional`** - Whether the section is optional (default: false)
- **`children`** - Nested subsections that must appear within this section
- **`code_blocks`** - Code block requirements with language and count constraints
- **`required_text`** - Text that must appear within the section

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
  - heading: "# .*"
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
  - heading: "# *"
    children:
      - heading: "## Prerequisites"
      - heading: "## Step 1: *"
        code_blocks:
          - { min: 1 } # Any language
      - heading: "## Step 2: *"
        code_blocks:
          - { min: 1 }
      - heading: "## Next Steps"
        optional: true
```

## Validation Rules

mdschema includes several built-in validation rules:

- **Structure** - Ensures sections appear in the correct order and hierarchy
- **Required Text** - Validates that required text appears in sections
- **Code Blocks** - Enforces code block requirements (language, minimum/maximum count)

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
