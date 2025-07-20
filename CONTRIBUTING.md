# Contributing to mdschema

Thank you for your interest in contributing to mdschema! We welcome contributions from the community.

## Development Setup

1. **Clone the repository**:

```bash
git clone https://github.com/jackchuka/mdschema.git
cd mdschema
```

2. **Install Go** (version 1.24 or later):

- Download from [golang.org](https://golang.org/dl/)

3. **Build the project**:

```bash
go build -o mdschema ./cmd/mdschema
```

4. **Run tests and lint**:

```bash
go test ./...
golangci-lint run
```

## Making Changes

### Code Style

- Follow standard Go conventions
- Use `gofmt` to format your code
- Add tests for new functionality
- Update documentation as needed

### Testing

- All tests must pass before submitting a PR
- Add tests for new features and bug fixes
- Integration tests are in `testdata/integration_test.go`

### Commit Messages

- Use clear, descriptive commit messages
- Start with a verb (Add, Fix, Update, etc.)
- Keep the first line under 50 characters
- Add more details in the body if needed

Example:

```
Add support for custom validation rules

This change introduces a plugin system that allows users to
define custom validation rules in their schemas.
```

## Submitting Changes

1. **Fork the repository** on GitHub
2. **Create a feature branch** from main:
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. **Make your changes** and add tests
4. **Run tests** to ensure everything works:
   ```bash
   go test ./...
   ```
5. **Commit your changes** with a clear message
6. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```
7. **Create a Pull Request** on GitHub

## Pull Request Guidelines

- **Describe your changes** clearly in the PR description
- **Link to any relevant issues** using keywords like "Fixes #123"
- **Ensure all tests pass** in CI
- **Be responsive** to code review feedback
- **Keep PRs focused** - one feature or fix per PR

## Types of Contributions

We welcome many types of contributions:

- **Bug fixes** - Fix issues in existing functionality
- **New features** - Add new validation rules or capabilities
- **Documentation** - Improve README, add examples, fix typos
- **Tests** - Add test coverage for existing code
- **Performance** - Optimize existing code
- **Refactoring** - Improve code structure and maintainability

## Code Architecture

Understanding the codebase structure:

- `cmd/mdschema/` - CLI application entry point
- `internal/parser/` - Markdown parsing using goldmark
- `internal/rules/` - Validation rules (structure, code blocks, text)
- `internal/schema/` - Schema loading and definition
- `internal/generator/` - Template generation from schemas
- `internal/reporter/` - Output formatting
- `testdata/` - Integration tests and test schemas

## Getting Help

- **Issues** - Check existing issues or create a new one
- **Discussions** - Use GitHub Discussions for questions
- **Code Review** - Ask questions in your PR if you need guidance

## Code of Conduct

Please be respectful and considerate in all interactions. We want mdschema to be a welcoming project for contributors of all backgrounds and experience levels.

Thank you for contributing! 🎉
