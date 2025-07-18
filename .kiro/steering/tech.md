# SysPkg Technology Stack

## Language & Runtime
- **Go 1.23+** - Primary language with modern Go features
- **Standard Library** - Minimal external dependencies, leveraging Go's robust stdlib
- **Context Support** - Proper timeout and cancellation handling throughout

## Architecture
- **Plugin-based Architecture** - Auto-registration via `init()` functions
- **Unified Interface** - Single `PackageManager` interface for all package managers
- **Concurrent Operations** - Thread-safe registry with `sync.RWMutex` protection
- **Priority-based Selection** - Automatic best-match selection for package managers

## Build System & Commands

### Core Build Commands
```bash
# Development workflow
make test              # Smart OS-aware testing
make check             # Code quality checks (lint, format, vet)
make build             # Build for current platform
make build-all-arch    # Cross-platform builds

# Testing commands
make test-docker-ubuntu    # APT testing in Ubuntu container
make test-docker-rocky     # YUM testing in Rocky Linux container
make test-docker-all       # Multi-OS Docker testing
make test-fixtures         # Generate authentic test fixtures
```

### Quality & Formatting
```bash
make lint              # golangci-lint + gofmt + go mod tidy
make format            # gofmt + goimports with local imports
make install-tools     # Install required dev tools
```

## Testing Infrastructure
- **Three-layer Testing**: Unit (fixtures) → Integration (Docker) → System (CI)
- **Docker-based Safety**: All integration tests run in containers
- **Fixture-based Validation**: Real package manager outputs as test data
- **Multi-OS Support**: Ubuntu, Rocky Linux, Fedora, Alpine containers

## Security Features
- **Input Validation**: Regex-based package name validation
- **Command Injection Prevention**: No shell interpretation of user input
- **Parameterized Commands**: Safe command construction
- **Timeout Management**: Context-based operation timeouts

## Dependencies
- **Minimal External Deps**: Primarily uses Go standard library
- **golangci-lint**: Code quality and linting
- **Docker**: Multi-OS testing and fixture generation
- **Pre-commit**: Automated quality checks

## Core Architecture Files
- `manager/interfaces.go` - Core PackageManager interface
- `manager/base.go` - BaseManager with common functionality
- `manager/registry.go` - Thread-safe plugin registration
- `manager/command_runner.go` - Command execution abstraction
- `manager/security.go` - Input validation and security

## Performance Characteristics
- **Startup**: <50ms (plugin registration)
- **Memory**: <10MB baseline
- **Concurrent Ops**: 3x performance improvement via `*AllConcurrent` methods
- **Thread Safety**: Full protection for concurrent access
