# Contributing to configdiff

Thank you for considering contributing to configdiff!

## Development Workflow

This project follows strict GitHub workflow practices:

1. **Never commit directly to `main`** - All changes must go through Pull Requests
2. **Branch protection is enabled** - PRs must pass all status checks before merging
3. **Linear history preferred** - Use rebase workflows when possible

## Getting Started

1. Fork the repository
2. Clone your fork locally
3. Create a feature branch from `main`
4. Make your changes with tests
5. Ensure all tests and checks pass
6. Push your branch and open a Pull Request

## Development Setup

```bash
# Clone the repository
git clone https://github.com/pfrederiksen/configdiff.git
cd configdiff

# Install dependencies
go mod download

# Run tests
go test ./...

# Run tests with race detector
go test -race ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run linter
golangci-lint run
```

## Code Quality Standards

- **Test Coverage**: Maintain high test coverage. All new features must include tests.
- **Documentation**: All exported types and functions must have GoDoc comments.
- **Linting**: Code must pass `golangci-lint` checks.
- **Performance**: Consider performance implications; correctness comes first.
- **Simplicity**: Avoid over-engineering. Keep solutions focused and minimal.

## Testing Guidelines

- Write table-driven tests for unit tests
- Use golden files in `testdata/` for output validation
- Include fuzz tests where appropriate (parsing, normalization)
- Test edge cases and error conditions
- Ensure deterministic test behavior

## Pull Request Guidelines

- **Title**: Use conventional commit style (e.g., `feat: add array set support`, `fix: handle null values`)
- **Description**: Clearly explain what changed and why
- **Tests**: Include test evidence (coverage reports, test output)
- **Scope**: Keep PRs focused on a single concern
- **Documentation**: Update README and GoDoc as needed

## Commit Message Format

Follow conventional commits:

```
<type>: <short description>

<optional body>

<optional footer>
```

Types: `feat`, `fix`, `docs`, `test`, `refactor`, `perf`, `chore`, `ci`

## Questions?

Open an issue for discussion before starting work on major features.
