# Contributing to abp-gen

Thank you for your interest in contributing to abp-gen! This document provides guidelines and instructions for contributing.

## Code of Conduct

This project adheres to a Code of Conduct that all contributors are expected to follow. Please read [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) before contributing.

## How to Contribute

### Reporting Bugs

If you find a bug, please open an issue using the [bug report template](.github/ISSUE_TEMPLATE/bug_report.md). Include:

- A clear description of the bug
- Steps to reproduce
- Expected vs actual behavior
- Environment details (OS, Go version, etc.)
- Any relevant logs or error messages

### Suggesting Features

Feature requests are welcome! Please use the [feature request template](.github/ISSUE_TEMPLATE/feature_request.md) and include:

- A clear description of the feature
- Use cases and examples
- Potential implementation approach (if you have ideas)

### Pull Requests

1. **Fork the repository** and clone it locally
2. **Create a branch** from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. **Make your changes** following the coding standards
4. **Write or update tests** for your changes
5. **Run tests** to ensure everything passes:
   ```bash
   go test ./...
   go vet ./...
   ```
6. **Commit your changes** with clear, descriptive messages
7. **Push to your fork** and open a pull request

## Development Setup

### Prerequisites

- Go 1.21 or later
- Git

### Getting Started

1. Clone the repository:
   ```bash
   git clone https://github.com/mohamedhabibwork/abp-gen.git
   cd abp-gen
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build the project:
   ```bash
   go build ./cmd/abp-gen
   ```

4. Run tests:
   ```bash
   go test ./...
   ```

### Project Structure

```
abp-gen/
├── cmd/abp-gen/          # Main CLI application
├── internal/
│   ├── detector/         # Solution/project detection
│   ├── generator/        # Code generators
│   ├── merger/           # File merging system
│   ├── prompts/          # Interactive prompts
│   ├── schema/           # Schema parsing and validation
│   ├── templates/        # Code templates
│   └── writer/           # File writing utilities
├── examples/             # Example schemas
├── test/                 # Test files
└── .github/              # GitHub workflows and templates
```

## Coding Standards

### Go Code Style

- Follow standard Go formatting (`go fmt`)
- Use `gofmt` and `golint` before committing
- Write clear, self-documenting code
- Add comments for exported functions and types
- Keep functions focused and small

### Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `style:` Code style changes (formatting, etc.)
- `refactor:` Code refactoring
- `test:` Adding or updating tests
- `chore:` Maintenance tasks

Example:
```
feat: add support for MongoDB repositories

- Add MongoDB repository generator
- Update schema to support MongoDB provider
- Add MongoDB templates
```

### Testing

- Write tests for new features
- Maintain or improve test coverage
- Test edge cases and error conditions
- Run all tests before submitting PR:
  ```bash
  go test -v -race ./...
  ```

### Documentation

- Update README.md for user-facing changes
- Add code comments for complex logic
- Update CHANGELOG.md for significant changes
- Keep examples up to date

## Development Workflow

1. **Create an issue** to discuss major changes
2. **Fork and branch** from `main`
3. **Make changes** following coding standards
4. **Test thoroughly** locally
5. **Update documentation** as needed
6. **Submit PR** with clear description
7. **Address review feedback**

## Review Process

- All PRs require at least one approval
- CI must pass before merging
- Maintainers will review code quality, tests, and documentation
- Be responsive to feedback and questions

## Questions?

- Open an issue for questions or discussions
- Check existing issues and PRs first
- Be respectful and constructive in all interactions

Thank you for contributing to abp-gen! 🎉

