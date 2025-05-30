# CI/CD Workflows

This directory contains GitHub Actions workflows for the go-syspkg project.

## Workflows

### 1. Test and Coverage (`test-and-coverage.yml`)
Runs comprehensive tests with coverage reporting:
- **Go Versions**: 1.23, 1.24 (project requires Go 1.23+)
- **Platform**: Ubuntu Latest
- **Coverage**: Uploads test coverage to Codecov
- **Race Detection**: Runs tests with race detector enabled

### 2. Lint and Format (`lint-and-format.yml`)
Ensures code quality and formatting standards:
- **gofmt**: Checks code formatting
- **go vet**: Reports suspicious constructs
- **golangci-lint**: Comprehensive linting with multiple linters
- **go mod tidy**: Ensures go.mod and go.sum are up to date

### 3. Build (`build.yml`)
Verifies code builds across Go versions:
- **Go Versions**: 1.23, 1.24
- **Build Targets**: All packages and CLI binary
- **Platform**: Ubuntu Latest

### 4. Release Binaries (`release-binaries.yml`)
Creates cross-platform release binaries:
- **Platforms**: Linux, Windows, Darwin
- **Architectures**: amd64, arm64, 386 (where supported)
- **Go Version**: 1.24
- **Trigger**: On GitHub releases

## Local Development

### Running Checks Locally
```bash
# Format code
make format

# Check formatting and run linters
make check

# Run specific checks
gofmt -l .                    # List files that need formatting
go vet ./...                  # Run go vet
golangci-lint run            # Run all linters
```

### Pre-commit Hooks
Install pre-commit to run checks automatically before commits:
```bash
pre-commit install
```

Pre-commit hooks include:
- File hygiene (trailing whitespace, EOF, merge conflicts)
- Go tools (gofmt, goimports, go vet, go mod tidy, golangci-lint)
- Build verification (go build, go mod verify)
- Security-focused using local system tools only

## Configuration Files

- `.golangci.yml`: Configures golangci-lint with enabled/disabled linters
- `.pre-commit-config.yaml`: Defines pre-commit hooks
- `Makefile`: Contains format and check targets

## Adding New Checks

To add new linting rules:
1. Update `.golangci.yml` to enable/disable linters
2. Update the `lint-and-format.yml` workflow if needed
3. Test locally with `make check`

## Project Standards

- **License**: Apache License 2.0 (patent protection and enterprise clarity)
- **Go Version**: 1.23+ required, CI tests with 1.23 and 1.24
- **Code Quality**: Enforced via pre-commit hooks and CI workflows
- **Security**: Local system tools used in pre-commit for maximum security
