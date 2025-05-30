# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Build
```bash
make build          # Build for current platform
make build-all-arch # Build for all supported platforms
make all           # Run tests and then build
```

Binary output: `bin/syspkg`

### Test
```bash
make test          # Run all tests with verbose output
go test ./manager/apt -v  # Run tests for specific package
```

### Lint and Format
```bash
make lint          # Run go mod tidy, golangci-lint, and gofmt
make format        # Format code with gofmt and goimports
make check         # Run all code quality checks (formatting, linting, vet)
make install-tools # Install required tools (golangci-lint)
```

### Pre-commit Hooks
```bash
pre-commit install        # Install pre-commit hooks
pre-commit run --all-files # Run all hooks on all files
```

**Pre-commit hooks include:**
- File hygiene (trailing whitespace, EOF, merge conflicts)
- Go tools (gofmt, goimports, go vet, go mod tidy, golangci-lint)
- Build verification (go build, go mod verify)
- Security-focused using local system tools only

### CI/CD Workflows
Located in `.github/workflows/`:
- **test-and-coverage.yml**: Go 1.23/1.24 testing with coverage
- **lint-and-format.yml**: Code quality and formatting checks
- **build.yml**: Multi-version build verification
- **release-binaries.yml**: Cross-platform binary releases

## Architecture Overview

### Core Interfaces
- `PackageManager` (interface.go): Defines methods all package managers must implement
- `SysPkg` (interface.go): High-level interface for managing multiple package managers

### Package Structure
- `/cmd/syspkg/`: CLI application using urfave/cli/v2 framework
- `/manager/`: Package manager implementations
  - Each manager (apt, snap, flatpak) has its own directory
  - Common types in `options.go` and `packageinfo.go`
- `/osinfo/`: OS detection utilities for determining available package managers

### Key Design Patterns
- Interface-based abstraction allows easy addition of new package managers
- Factory pattern in `syspkg.go` for creating manager instances
- Options pattern for configurable behavior (manager.Options)

### Adding New Package Managers
1. Create new directory under `/manager/`
2. Implement the `PackageManager` interface
3. Add to `initializePackageManagers()` in `syspkg.go`
4. Add tests following existing patterns (parse functions, availability checks)

### Testing Patterns
- Unit tests use table-driven tests with mock data
- Integration tests check actual package manager availability
- Parse functions are heavily tested with real command output examples
- Tests use `*_test` package naming for better encapsulation

### CLI Command Structure
Main commands: `install`, `delete`, `refresh`, `upgrade`, `find`, `show`
Package manager flags: `--apt`, `--flatpak`, `--snap`
Options: `--debug`, `--assume-yes`, `--dry-run`, `--interactive`, `--verbose`

## Important Notes

- **License**: Apache License 2.0 (provides patent protection and enterprise-grade legal clarity)
- **Go Version**: Requires Go 1.23+ (CI tests with 1.23, 1.24)
- **Code Quality**: Always run `make check` before committing to ensure quality
- **Pre-commit**: Hooks automatically enforce formatting, linting, and security checks
- **Package Managers**: When implementing new ones, focus on parsing command outputs correctly
- **CLI Detection**: Automatically detects available package managers if no flag is specified
- **Privileges**: Root privileges are often required for package operations

## Philosophy

**Tool-Focused Approach**: SysPkg focuses on supporting package manager tools based on their functionality rather than the operating system they're running on. If apt+dpkg work correctly in a container, on macOS via Homebrew, or in any other environment, SysPkg will support them. This makes the project more flexible and useful across different development environments.

## Project Improvement Roadmap

*Note: To-do list consolidated 2025-05-30 - removed duplicates, feature creep items, and over-engineering. Focused on core security, testing, and platform support.*

### 🔴 High Priority (Security & Critical Bugs) - 4 items
1. **Fix command injection vulnerability** - validate/sanitize package names before exec.Command
2. **Implement input validation helper function** for package names and arguments
3. **Fix resource leaks** in error handling paths
4. **Add security scanning with Snyk** to CI/CD pipeline
5. **Review and merge PR #12** - fix GetPackageManager("") panic bug ✅

### 🟡 Medium Priority (Code Quality & Testing) - 7 items
**Testing:**
- Create integration tests with mocked command execution
- Add unit tests for snap package manager
- Add unit tests for flatpak package manager

**Code Improvements:**
- Implement context support for cancellation and timeouts
- Create custom error types for better error handling
- Extract common parsing logic to shared utilities (DRY principle)
- Replace magic strings/numbers with named constants

**Removed from roadmap (2025-05-30):**
- ~~Structured logging~~ (over-engineering for project scope)
- ~~Progress indicators~~ (feature creep for CLI/library)
- ~~Architecture diagrams~~ (low ROI for library project)
- ~~TODO comment fixes~~ (covered by security improvements)

### 🟢 Low Priority (Platform Support) - 3 items
**New Package Managers:**
- Add proper macOS support with brew package manager implementation
- Add Windows support with chocolatey/scoop/winget package managers
- Implement dnf/yum package manager support (Red Hat/Fedora)

**Removed from roadmap (2025-05-30):**
- ~~zypper, apk support~~ (lower priority than core platforms)
- ~~Parallel operations~~ (premature optimization)

## Testing Strategy Notes

SysPkg uses a comprehensive multi-layered testing approach to ensure package managers work correctly across different operating systems.

### OS/Package Manager Matrix Testing

**Configuration-Driven Testing**: `testing/os-matrix.yaml` defines which package managers should be tested on which OS distributions.

**Supported Testing Environments**:
- **Ubuntu/Debian**: APT, Flatpak, Snap
- **RHEL/Rocky/Alma**: YUM (v8), DNF (v9+)
- **Fedora**: DNF, Flatpak
- **Alpine**: APK
- **Arch** (planned): Pacman

### Multi-Layer Test Architecture

#### 1. **Unit Tests** (Run Everywhere)
```bash
make test-unit
```
- Parser functions with OS-specific fixtures
- OS detection logic
- Command construction
- No actual package manager execution

#### 2. **Integration Tests** (Docker + Native)
```bash
make test-integration
```
- Real package manager availability checks
- Command output capture for test fixtures
- Limited package operations (list, search, show)

#### 3. **Docker-Based Multi-OS Testing**
```bash
make test-docker-all          # All OS
make test-docker-ubuntu       # APT testing
make test-docker-rocky        # YUM testing
make test-docker-alma         # YUM testing
make test-docker-fedora       # DNF testing
make test-docker-alpine       # APK testing
```

**Docker Benefits**:
- Test YUM on Rocky Linux/AlmaLinux
- Test APT on various Ubuntu/Debian versions
- Generate real command outputs for fixtures
- Isolated, reproducible test environments

#### 4. **System Tests** (Native CI Only)
- Actual package installation/removal
- Privileged operations
- Snap/systemd dependent features

### Environment-Aware Testing

**Automatic Detection**: Tests automatically detect the current OS and determine which package managers to test:

```go
env, _ := testenv.GetTestEnvironment()
if skip, reason := env.ShouldSkipTest("yum"); skip {
    t.Skip(reason)
}
```

**Test Tags**: Tests use build tags for selective execution:
- `unit`: Parser and core logic tests
- `integration`: Real command execution (limited)
- `system`: Full package operations (privileged)
- `apt`, `yum`, `dnf`, `apk`: Package manager specific

### CI/CD Multi-OS Pipeline

**Docker Matrix**: Tests run across multiple OS in parallel:
```yaml
strategy:
  matrix:
    include:
      - os: ubuntu, pm: apt
      - os: rockylinux, pm: yum
      - os: fedora, pm: dnf
      - os: alpine, pm: apk
```

**Native Tests**: For systemd-dependent features like Snap:
```yaml
- os: ubuntu, runner: ubuntu-latest, pm: snap
```

### Local Development Workflow

**For detailed development workflows, see [CONTRIBUTING.md](CONTRIBUTING.md)**

**Quick reference:**
1. **Daily development**: `make test` (smart OS-aware testing)
2. **Package manager work**: `make test-docker-rocky` (YUM), `make test-docker-fedora` (DNF)
3. **Comprehensive validation**: `make test-docker-all`
4. **Fixture updates**: `make test-fixtures`

### Test Fixture Generation

Fixtures are automatically generated from real package manager outputs across different OS:
- `testing/fixtures/apt/search-vim-ubuntu22.txt`
- `testing/fixtures/yum/search-vim-rocky8.txt`
- `testing/fixtures/dnf/search-vim-fedora39.txt`

This ensures parsers work correctly with real-world output variations across distributions.

See `testing/docker/`, `testing/os-matrix.yaml`, and [CONTRIBUTING.md](CONTRIBUTING.md) for complete details.
