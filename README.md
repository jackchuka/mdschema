<div align="center">

# mdschema

**Lint the _structure_ of your Markdown, not just its style.**

Declare what your docs must contain in a small YAML schema,<br>
then validate, scaffold, and enforce it in CI.

[![Test](https://github.com/jackchuka/mdschema/actions/workflows/test.yml/badge.svg)](https://github.com/jackchuka/mdschema/actions/workflows/test.yml)
[![Release](https://img.shields.io/github/v/release/jackchuka/mdschema?sort=semver)](https://github.com/jackchuka/mdschema/releases)
[![npm](https://img.shields.io/npm/v/@jackchuka/mdschema)](https://www.npmjs.com/package/@jackchuka/mdschema)
[![Go Report Card](https://goreportcard.com/badge/github.com/jackchuka/mdschema)](https://goreportcard.com/report/github.com/jackchuka/mdschema)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

<img src="assets/demo.gif" alt="mdschema checking a README, then generating a template that passes" width="800">

[Installation](#installation) · [Quick Start](#quick-start) · [Schema Reference](#schema-reference) · [GitHub Action](#github-action) · [Examples](#examples)

</div>

---

Style linters tell you a line is too long. **mdschema** tells you the README is missing its `## Installation` section, that the install section has no `bash` snippet, or that a blog post's `date` isn't a valid date.

The same schema also **generates** a ready-to-fill template for new documents, and can be **derived** from a document you already like.

> [!TIP]
> This README is validated against [`examples/README.mdschema.yml`](examples/README.mdschema.yml) in CI.

## Features

- **Schema-driven** — describe the expected document in plain YAML, with nested sections and repeatable headings
- **Content rules** — require or forbid text, code blocks by language, images, tables, lists, and word counts per section
- **Frontmatter checks** — required fields, types, formats (`date`, `email`, `url`), enums, and nested keys
- **Link checks** — internal anchors, relative files, and (optionally) external URLs
- **Dynamic headings** — match headings against the filename with expressions like `slug(filename) == slug(heading)`
- **Templates both ways** — `generate` a skeleton from a schema, or `derive` a schema from an existing document
- **AST-based** — parses Markdown properly, so headings inside code blocks never confuse it
- **Runs anywhere** — single static binary for Linux, macOS, and Windows; also on npm and as a GitHub Action
- **Editor support** — JSON Schema for completion and inline validation of `.mdschema.yml`

## Installation

### Homebrew

```bash
brew install jackchuka/tap/mdschema
```

### npm

```bash
npm install -g @jackchuka/mdschema
# or run without installing
npx @jackchuka/mdschema check README.md
```

### Go

```bash
go install github.com/jackchuka/mdschema/cmd/mdschema@latest
```

### Binary

Download a prebuilt binary from the [releases page](https://github.com/jackchuka/mdschema/releases).

### From Source

```bash
git clone https://github.com/jackchuka/mdschema.git
cd mdschema
go build -o mdschema ./cmd/mdschema
```

## Quick Start

1. **Create a schema** — start from defaults, or infer one from a document you already like:

   ```bash
   mdschema init                              # writes .mdschema.yml
   mdschema derive README.md -o .mdschema.yml # or derive it from README.md
   ```

2. **Validate** your Markdown files:

   ```bash
   mdschema check README.md "docs/**/*.md"
   ```

3. **Scaffold** a new document that already satisfies the schema:

   ```bash
   mdschema generate -o new-doc.md
   ```

## Commands

| Command                  | Description                                              |
| ------------------------ | -------------------------------------------------------- |
| `mdschema check [globs]` | Validate files against the schema; exits 1 on violation  |
| `mdschema generate`      | Print (or `-o` save) a Markdown template from the schema |
| `mdschema init`          | Create a starter `.mdschema.yml`                         |
| `mdschema derive <file>` | Infer a schema from an existing document                 |
| `mdschema schema`        | Print the JSON Schema for `.mdschema.yml` files          |
| `mdschema version`       | Print the version                                        |

Every command accepts `--schema <path>` (default: `.mdschema.yml`).

## Schema Reference

A schema has one `structure` tree plus optional document-wide rules:

```yaml
structure: [...] # expected sections, in order
heading_rules: { ... } # heading levels and uniqueness
links: { ... } # link validation
frontmatter: { ... } # YAML frontmatter validation
```

### Sections

Each entry in `structure` (and in any `children`) describes one section:

| Key                | Description                                                            |
| ------------------ | ---------------------------------------------------------------------- |
| `heading`          | `"## Title"` (literal), `{pattern: "## .*"}` (regex), or `{expr: ...}` |
| `description`      | Guidance emitted as an HTML comment by `generate`                      |
| `optional`         | Section may be omitted (default `false`)                               |
| `count`            | Repeatable section: `{min: 1, max: 0}` (`0` = unlimited)               |
| `allow_additional` | Allow subsections not listed in `children` (default `false`)           |
| `children`         | Nested subsections, in order                                           |

…and these content rules, checked inside that section:

| Rule             | Example                                                |
| ---------------- | ------------------------------------------------------ |
| `required_text`  | `["API key", {pattern: "v\\d+\\.\\d+"}]`               |
| `forbidden_text` | `["TODO", "FIXME"]`                                    |
| `code_blocks`    | `[{lang: bash, min: 1, max: 3}]`                       |
| `images`         | `[{min: 1, require_alt: true, formats: [png, svg]}]`   |
| `tables`         | `[{min: 1, min_columns: 2, required_headers: [Name]}]` |
| `lists`          | `[{min: 1, type: ordered, min_items: 3}]`              |
| `word_count`     | `{min: 50, max: 500}`                                  |

### Heading Expressions

Use `expr` to match headings dynamically, for example to require that `my-file.md` starts with `# My File`:

```yaml
structure:
  - heading:
      expr: "slug(filename) == slug(heading)"
    children:
      - heading: "## Features"
```

Variables: `filename` (without extension), `heading` (heading text), `level` (1–6).

<details>
<summary><b>Available functions</b></summary>

| Function                 | Description         | Example                                         |
| ------------------------ | ------------------- | ----------------------------------------------- |
| `slug(s)`                | URL-friendly slug   | `slug("My File")` → `"my-file"`                 |
| `kebab(s)`               | PascalCase to kebab | `kebab("MyFile")` → `"my-file"`                 |
| `lower(s)` / `upper(s)`  | Case conversion     | `lower("README")` → `"readme"`                  |
| `trimPrefix(s, pattern)` | Remove regex prefix | `trimPrefix("01-file", "^\\d+-")` → `"file"`    |
| `trimSuffix(s, pattern)` | Remove regex suffix | `trimSuffix("file_draft", "_draft")` → `"file"` |
| `hasPrefix(s, prefix)`   | Check prefix        | `hasPrefix("api-ref", "api")` → `true`          |
| `hasSuffix(s, suffix)`   | Check suffix        | `hasSuffix("file_v2", "_v2")` → `true`          |
| `strContains(s, substr)` | Check contains      | `strContains("api-ref", "api")` → `true`        |
| `match(s, pattern)`      | Regex match         | `match("test-123", "test-\\d+")` → `true`       |
| `replace(s, old, new)`   | Replace all         | `replace("a-b-c", "-", "_")` → `"a_b_c"`        |

</details>

### Heading Rules

```yaml
heading_rules:
  no_skip_levels: true # disallow h1 -> h3 without h2
  unique: true # all headings must be unique
  unique_per_level: false # unique within the same level only
  max_depth: 4 # deepest allowed heading (h4)
```

### Links

```yaml
links:
  validate_internal: true # anchors like (#section)
  validate_files: true # relative files like (./file.md)
  validate_external: false # external URLs (slower)
  external_timeout: 10 # seconds
  allowed_domains: [github.com, golang.org]
  blocked_domains: [example.com]
```

### Frontmatter

Frontmatter is required once a `frontmatter` block exists; fields are required unless marked `optional`.

```yaml
frontmatter:
  optional: false
  fields:
    - { name: "title", type: string }
    - { name: "date", format: date }
    - { name: "author", optional: true, format: email }
    - { name: "tags", optional: true, type: array, enum: [go, cli, markdown] }
    - { name: "status", enum: [draft, published, archived] }
    - { name: "metadata.version" } # nested key via dot-notation
```

- **Types:** `string`, `number`, `boolean`, `array`, `date`, `object`
- **Formats:** `date` (YYYY-MM-DD), `email`, `url`
- **Enums:** with `type: array`, every element must be listed. Entries are checked against the declared `type`/`format` when the schema loads, so an enum no value could satisfy (e.g. `{ type: number, enum: [low, high] }`) is reported as a schema warning.
- **Nested keys:** `metadata.author` validates `author` under `metadata`. Escape a literal dot with a backslash: `name: "weird\\.key"`.

## GitHub Action

```yaml
name: Docs
on: [push, pull_request]

jobs:
  mdschema:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: jackchuka/mdschema@v0.15.3
        with:
          files: "README.md docs/**/*.md"
          schema: ".mdschema.yml"
```

| Input               | Description                                    | Default         |
| ------------------- | ---------------------------------------------- | --------------- |
| `version`           | mdschema CLI version (use `latest` for newest) | Action ref      |
| `files`             | Files or glob patterns                         | `**/*.md`       |
| `schema`            | Path to schema file                            | `.mdschema.yml` |
| `args`              | Additional CLI arguments                       | (empty)         |
| `working-directory` | Working directory for validation               | `.`             |
| `token`             | Token for GitHub API version lookups           | `github.token`  |

For a monorepo, point `working-directory` at the package (e.g. `./packages/docs`).

## Editor Support

`.mdschema.yml` has a [JSON Schema](https://json-schema.org/) for completion, hover docs, and validation in any editor using the YAML Language Server (VS Code, Neovim, JetBrains, …). Add this line to the top of the file:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/jackchuka/mdschema/main/schema.json
```

Or map it once in VS Code's `settings.json`:

```json
{
  "yaml.schemas": {
    "https://raw.githubusercontent.com/jackchuka/mdschema/main/schema.json": ".mdschema.yml"
  }
}
```

## Examples

Complete schema + document pairs live in [`examples/`](examples):

| Schema                                                            | Document                                      | Shows                                  |
| ----------------------------------------------------------------- | --------------------------------------------- | -------------------------------------- |
| [`README.mdschema.yml`](examples/README.mdschema.yml)             | [this README](README.md)                      | Regex headings, link and heading rules |
| [`blog-post.mdschema.yml`](examples/blog-post.mdschema.yml)       | [`blog-post.md`](examples/blog-post.md)       | Frontmatter, word counts, images       |
| [`tutorial.mdschema.yml`](examples/tutorial.mdschema.yml)         | [`tutorial.md`](examples/tutorial.md)         | Repeatable `## Step N` sections        |
| [`requirements.mdschema.yml`](examples/requirements.mdschema.yml) | [`requirements.md`](examples/requirements.md) | Structured specification documents     |

### Repeatable Steps

```yaml
structure:
  - heading: { pattern: "# .*" }
    children:
      - heading: "## Prerequisites"
      - heading: { pattern: "## Step [0-9]+: .*" }
        count: { min: 1, max: 0 } # one or more steps
        code_blocks: [{ min: 1 }] # each step needs a code block
      - heading: "## Next Steps"
        optional: true
```

### Flexible Sections

```yaml
structure:
  - heading: "# Project Name"
    allow_additional: true # authors may add FAQ, Troubleshooting, ...
    children:
      - heading: "## Overview"
      - heading: "## Installation"
        code_blocks: [{ lang: bash, min: 1 }]
```

## Development

```bash
go test ./...
go build -o mdschema ./cmd/mdschema
```

## Contributing

Contributions are welcome! For major changes, please open an issue first to discuss what you would like to change. See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

[MIT](LICENSE)
