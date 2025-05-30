# CI/CD Workflows

This directory contains GitHub Actions workflows for the go-syspkg project.

## Workflows

### 1. Lint and Format (`lint.yml`)
Runs on every push and pull request to ensure code quality:
- **gofmt**: Checks code formatting
- **go vet**: Reports suspicious constructs
- **golangci-lint**: Comprehensive linting with multiple linters
- **go mod tidy**: Ensures go.mod and go.sum are up to date

### 2. Test (`test.yml`)
Runs comprehensive tests across multiple platforms:
- **OS Matrix**: Ubuntu and macOS
- **Go Versions**: 1.21, 1.22, 1.23
- **Coverage**: Uploads test coverage to Codecov
- **Race Detection**: Runs tests with race detector enabled

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
pip install pre-commit
pre-commit install
```

## Configuration Files

- `.golangci.yml`: Configures golangci-lint with enabled/disabled linters
- `.pre-commit-config.yaml`: Defines pre-commit hooks
- `Makefile`: Contains format and check targets

## Adding New Checks

To add new linting rules:
1. Update `.golangci.yml` to enable/disable linters
2. Update the `lint.yml` workflow if needed
3. Test locally with `make check`