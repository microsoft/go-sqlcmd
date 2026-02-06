# Contributing to go-sqlcmd

Thank you for your interest in contributing to go-sqlcmd! This document provides guidelines for contributing to the project.

## Development Setup

### Prerequisites

- Go 1.24 or later
- Git

### Building the Project

```bash
# Clone the repository
git clone https://github.com/microsoft/go-sqlcmd.git
cd go-sqlcmd

# Build sqlcmd
./build/build.sh       # Linux/macOS
.\build\build.cmd      # Windows
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test -v ./pkg/sqlcmd/...
```

## Release Process

This project uses [Release Please](https://github.com/googleapis/release-please) for automated version management and releases. Release Please analyzes commits since the last release and automatically creates a Release PR that:

- Bumps the version based on [Conventional Commits](https://www.conventionalcommits.org/)
- Updates the CHANGELOG.md with changes
- Creates a GitHub release when the PR is merged

### Conventional Commits

All commits should follow the [Conventional Commits](https://www.conventionalcommits.org/) specification. This enables automated version bumping and changelog generation.

#### Commit Message Format

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

#### Types and Version Bumping

| Type | Version Bump | Description | Example |
|------|--------------|-------------|---------|
| `feat:` | Minor (0.X.0) | New feature | `feat: add query timeout option` |
| `fix:` | Patch (0.0.X) | Bug fix | `fix: resolve connection timeout issue` |
| `feat!:` or `BREAKING CHANGE:` | Major (X.0.0) | Breaking change | `feat!: change default port to 1433` |
| `docs:` | No bump | Documentation only | `docs: update README with examples` |
| `chore:` | No bump | Maintenance tasks | `chore: update dependencies` |
| `ci:` | No bump | CI/CD changes | `ci: add workflow for linting` |
| `test:` | No bump | Test changes | `test: add unit tests for parser` |
| `refactor:` | No bump | Code refactoring | `refactor: simplify connection logic` |
| `perf:` | No bump | Performance improvements | `perf: optimize query execution` |
| `build:` | No bump | Build system changes | `build: update build script` |

#### Examples

**Feature (Minor bump):**
```
feat: add support for Azure AD authentication

Adds new authentication method for connecting to Azure SQL
databases using Azure Active Directory credentials.
```

**Bug fix (Patch bump):**
```
fix: correct handling of null values in output

Previously, null values were displayed as empty strings.
Now they are properly displayed as "NULL".
```

**Breaking change (Major bump):**
```
feat!: change default encryption to mandatory

BREAKING CHANGE: The default encryption setting has changed
from optional to mandatory. Users must explicitly set
encryption to optional if needed.
```

**Non-version-bumping changes:**
```
docs: add examples for container commands
chore: update go-mssqldb dependency
ci: add codeql security scanning
test: add integration tests for query command
```

### How Releases Work

1. **Make changes** following conventional commit guidelines
2. **Create PR** with your changes
3. **Merge PR to main** - Release Please will analyze commits
4. **Release Please creates/updates a Release PR** that:
   - Bumps version in relevant files
   - Updates CHANGELOG.md
5. **Review and merge the Release PR** - This triggers:
   - Creation of a git tag for the new version
   - Creation of a GitHub Release
   - Publishing of release artifacts

### Manual Release Process (if needed)

If you need to create a release manually:

1. Update the version in `.release-please-manifest.json`
2. Update CHANGELOG.md manually
3. Create and push a git tag:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```
4. Create a GitHub release from the tag

## Pull Request Guidelines

- Follow the conventional commit format in PR titles
- Keep PRs focused on a single feature or fix
- Add tests for new functionality
- Ensure all tests pass locally before submitting
- Update documentation if needed

## Code Style

- Follow standard Go conventions and idioms
- Use `gofmt` for formatting
- Run `golangci-lint` before submitting

## Questions?

If you have questions, feel free to:
- Open an issue for discussion
- Check existing issues and discussions
- Review the [README](README.md) for more information

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
