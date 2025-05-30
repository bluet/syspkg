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

- The project requires Go 1.23+ (tests with 1.23, 1.24)
- Always run `make lint` before committing to ensure code quality
- When implementing new package managers, focus on parsing command outputs correctly
- The CLI automatically detects available package managers if no flag is specified
- Root privileges are often required for package operations

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

### Docker Testing Capabilities
- **Works Well**: APT, DNF/YUM, APK, Flatpak (limited) - for capturing command outputs and testing parsers
- **Doesn't Work**: Snap (requires systemd), actual package installations
- **Best Practice**: Use Docker to capture real outputs, then use mocks for testing

### Testing Approach
1. **Unit Tests**: Parser functions with captured fixtures
2. **Integration Tests**: Mock exec.Command for package operations
3. **Docker Tests**: Multi-OS parser validation with real command outputs
4. **CI/CD Tests**: Native runners for snap and full integration tests

See `testing/docker/` for implementation details and strategies.
