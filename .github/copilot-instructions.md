# GitHub Copilot Instructions for go-sqlcmd

This document provides guidance for GitHub Copilot when working with the go-sqlcmd repository.

## Project Overview

go-sqlcmd is a Go-based command line tool (`sqlcmd`) for working with Microsoft SQL Server, Azure SQL Database, and Azure Synapse. The project aims to be a complete port of the original ODBC sqlcmd to Go, utilizing the [go-mssqldb](https://github.com/microsoft/go-mssqldb) driver.

## Repository Structure

```
/
├── cmd/                    # Entry points for the application
│   ├── modern/             # Modern CLI entry point (Cobra-based)
│   │   ├── root/           # Root command and subcommands
│   │   └── sqlconfig/      # SQL configuration management
│   └── sqlcmd/             # Legacy CLI entry point (Kong-based)
├── pkg/                    # Public packages consumable by other hosts
│   └── sqlcmd/             # Core sqlcmd functionality
├── internal/               # Internal packages (not for external use)
│   ├── buffer/             # Buffer management
│   ├── cmdparser/          # Command parsing utilities
│   ├── color/              # Console coloring
│   ├── config/             # Configuration management
│   ├── container/          # Docker/Podman container management
│   ├── credman/            # Credential management
│   ├── http/               # HTTP utilities
│   ├── io/                 # I/O utilities
│   ├── localizer/          # Localization support
│   ├── net/                # Network utilities
│   ├── output/             # Output formatting
│   ├── pal/                # Platform abstraction layer
│   ├── secret/             # Secret/credential management
│   ├── sql/                # SQL-related utilities
│   ├── test/               # Test utilities
│   ├── tools/              # Development tools
│   └── translations/       # Localized string translations
├── build/                  # Build scripts and templates
├── testdata/               # Test data files
├── release/                # Release-related files
└── .github/                # GitHub workflows and configurations
```

## Building the Project

### Build Commands

```bash
# Build the sqlcmd executable
./build/build.sh       # Linux/macOS
.\build\build.cmd      # Windows

# Or build directly with Go
go build -o sqlcmd ./cmd/modern
```

### Dependencies

The project uses Go modules. Run `go mod download` to fetch dependencies.

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test -v ./pkg/sqlcmd/...
go test -v ./internal/config/...
```

### Test Environment Variables

Tests may require the following environment variables for database connectivity:
- `SQLCMDSERVER` - SQL Server hostname
- `SQLCMDPORT` - SQL Server port (default: 1433)
- `SQLCMDUSER` - Username (e.g., `sa`)
- `SQLCMDPASSWORD` - Password
- `SQLCMDDATABASE` - Database name

## Code Style and Conventions

### Go Standards

- Follow standard Go conventions and idioms
- Use `gofmt` for formatting
- Use tabs for indentation (as specified in `.editorconfig`)
- Follow effective Go guidelines: https://go.dev/doc/effective_go

### File Headers

All Go files should include the following copyright header:
```go
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
```

### Package Documentation

Each package should have a `doc.go` file with package-level documentation.

### Error Handling

- Use the standard Go error handling patterns
- For user-facing errors, prefer localized error messages via the `internal/localizer` package
- Wrap errors with context when propagating

### Naming Conventions

- Use camelCase for unexported identifiers
- Use PascalCase for exported identifiers
- Acronyms should be all uppercase (e.g., `SQL`, `URL`, `HTTP`)
- Package names should be lowercase, single-word if possible

## CLI Architecture

### Modern CLI (Cobra-based)

The modern CLI is located in `cmd/modern/` and uses the [Cobra](https://github.com/spf13/cobra) library. Key points:
- Root command is in `cmd/modern/root.go`
- Subcommands are in `cmd/modern/root/` directory
- Uses dependency injection for testability

### Legacy CLI (Kong-based)

The legacy CLI is in `cmd/sqlcmd/` and maintains backward compatibility with the original ODBC sqlcmd.

### Command Structure

When adding new commands:
1. Create the command in `cmd/modern/root/`
2. Follow the existing pattern for subcommands (see `query.go`, `start.go`, `stop.go`)
3. Add corresponding tests with `_test.go` suffix

## Configuration

- Configuration files are stored in `~/.sqlcmd/sqlconfig`
- Use the `internal/config` package for configuration management
- Viper is used for configuration parsing (`internal/config/viper.go`)

## Container Support

The project supports creating SQL Server instances using Docker or Podman:
- Container management is in `internal/container/`
- Supports SQL Server images

## Localization

- Localized strings are in `internal/translations/`
- Use the `internal/localizer` package for localized messages
- Supported languages: Chinese (Simplified/Traditional), English, French, German, Italian, Japanese, Korean, Portuguese (Brazil), Russian, Spanish

### Adding Localizable Strings

When adding user-facing strings to the code, use the `localizer` package:

```go
import "github.com/microsoft/go-sqlcmd/internal/localizer"

// Use localizer.Sprintf for formatted strings
message := localizer.Sprintf("This is a localizable message with %s", value)

// Use localizer.Errorf for localized errors
err := localizer.Errorf("Error: %s failed", operation)
```

Constants that are not user-facing (like environment variable names, command names) should be placed in `internal/localizer/constants.go` and do not need localization.

### Generating Localization Files

After adding new localizable strings, you **must** regenerate the translation catalog files before committing. The build scripts handle this automatically.

#### On Windows

```cmd
build\build.cmd
```

This script:
- Installs `gotext` if not already installed
- Runs `go generate` which executes the gotext command defined in `internal/translations/translations.go`
- Generates/updates the translation catalog in `internal/translations/catalog.go`
- Reports any conflicting localizable strings that need to be fixed

#### On Linux/macOS

Run the following commands manually:

```bash
# Install gotext if not already installed
go install golang.org/x/text/cmd/gotext@latest

# Generate translation files
go generate ./...
```

### Important Notes

- Always run the build script after adding new user-facing strings
- Check the build output for "conflicting localizable strings" warnings and resolve them
- The `SQLCMD_LANG` environment variable controls the runtime language (e.g., `de-de`, `fr-fr`)
- Test your changes with different language settings to ensure proper localization

## Azure Authentication

- Azure AD authentication is supported via the `azidentity` package
- Authentication code is in `pkg/sqlcmd/azure_auth.go`
- Supports multiple authentication methods: DefaultAzureCredential, Password, Interactive, ManagedIdentity, ServicePrincipal

## Security Considerations

- Never commit secrets or credentials
- Use environment variables or secure credential stores for sensitive data
- Follow the Microsoft Security Development Lifecycle (SDL)
- Report security vulnerabilities via SECURITY.md

## Pull Request Guidelines

1. Ensure all tests pass locally
2. Follow the existing code style
3. Update documentation if adding new features
4. Add tests for new functionality
5. Keep commits focused and well-described

## Common Tasks

### Adding a New Command (Modern CLI)

For new commands related to context management, container operations, or configuration:

1. Create a new file in `cmd/modern/root/` (e.g., `mycommand.go`)
2. Define a command struct that embeds `cmdparser.Cmd`
3. Implement the `DefineCommand` method to set up command options, flags, and examples
4. Implement the `run` method with the command logic
5. Add corresponding tests in `mycommand_test.go`

Example structure:
```go
type MyCommand struct {
    cmdparser.Cmd
    // flags
}

func (c *MyCommand) DefineCommand(...cmdparser.CommandOptions) {
    options := cmdparser.CommandOptions{
        Use:   "mycommand",
        Short: localizer.Sprintf("Description"),
        Run:   c.run,
    }
    c.Cmd.DefineCommand(options)
    // Add flags
}

func (c *MyCommand) run() {
    // Command logic
}
```

### Adding a New Configuration Option (Modern CLI)

1. Update the config struct in `internal/config/config.go`
2. Add validation if needed
3. Update the Viper bindings in `internal/config/viper.go`
4. Add tests

### Adding Features (Legacy CLI)

For new features related to querying SQL Server and displaying query results, add them to the legacy CLI:

1. Add new fields to the `SQLCmdArguments` struct in `cmd/sqlcmd/sqlcmd.go`
2. Register new flags in the `setFlags` function
3. Add validation logic in the `Validate` method if needed
4. Determine from existing patterns whether to add a SQLCMD variable to support it
5. Update `setVars` or `setConnect` functions to use the new arguments
6. Implement the feature logic in the `run` function or related functions
7. Add corresponding tests in `cmd/sqlcmd/sqlcmd_test.go`
8. Update README.md to show example usage

### Working with SQL Connections

- Use `pkg/sqlcmd` for SQL connection management
- Connection options are defined in `pkg/sqlcmd/connect.go`
- Support for various transports: TCP, named pipes, shared memory

## CI/CD

- GitHub Actions workflows are in `.github/workflows/`
- `golangci-lint.yml` - Linting with golangci-lint
- `pr-validation.yml` - Build and test validation for PRs
- Azure Pipelines configurations are in `.pipelines/`

## Linting

The project uses golangci-lint. To run locally:

```bash
golangci-lint run
```

## Useful Resources

- [go-mssqldb driver](https://github.com/microsoft/go-mssqldb)
- [sqlcmd documentation](https://docs.microsoft.com/sql/tools/go-sqlcmd-utility)
- [Cobra CLI library](https://github.com/spf13/cobra)
- [Viper configuration library](https://github.com/spf13/viper)
