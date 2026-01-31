# Getting Started with mdschema

## Prerequisites

Before you begin, you need to install Go 1.21 or later.

## Step 1: Install mdschema

Install mdschema using Go:

```bash
go install github.com/jackchuka/mdschema/cmd/mdschema@latest
```

## Step 2: Create a Schema

Create a `.mdschema.yml` file:

```yaml
structure:
  - heading: "# My Project"
    children:
      - heading: "## Features"
      - heading: "## Installation"
```

## Step 3: Validate a Markdown File

Run mdschema to validate your Markdown file:

```bash
mdschema validate path/to/yourfile.md --schema path/to/.mdschema.yml
```

## Troubleshooting

If you encounter issues, check that Go is in your PATH.

## Next Steps

Explore more advanced schema features in the documentation.
