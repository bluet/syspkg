# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Rules
Read ~/.claude/CLAUDE.md

## Development Rules
- Use The twelve-factor methodology.
- Follow the language/framework specific coding conventions and best practices. If there is no specific coding style requirement in the particular programming language/framework, follow the KNF (Kernel Normal Form) coding convention.
- Use KISS (Keep It Simple and Stupid) and DRY (Don't Repeat Yourself) rules and modern software development principles and best practices.
- The project/code must be clean, documented, easy to maintain, secure, and efficient (high performance).
- Smaller commit instead of a large fat one, easier to review.
- For security reasons, only add the files intended to be added, avoid using `git add .` or `git add -A`, which might include sensitive data by accident.

  1. Methodical Investigation First

  "Before making any changes, use multiple Read/Grep tools to understand:
    1. What the code ACTUALLY does (not what you assume)
    2. What the test ACTUALLY tests (read the exact lines)
    3. What the error/issue ACTUALLY says (don't interpret)
    Then explain your findings before proposing solutions."

  2. Verify Assumptions

  "For every assumption you make, explicitly state it and verify it with tools.
  Examples:
    - 'I assume this test calls method X' → Read the test and confirm
    - 'I assume this runs on environment Y' → Check CI config and confirm
    - 'I assume this should behave like Z' → Check actual behavior and confirm"

  3. Understand Context Before Acting

  "When someone reports a problem:
    1. Read the actual error/comment they're referring to
    2. Understand what component is involved (parser vs full method vs CI vs local)
    3. Trace the execution path to understand what's really happening
    4. Only then propose a solution with evidence"

  4. Test Your Understanding

  "Before implementing any fix:
    1. Explain what you think the problem is
    2. Explain why your proposed solution addresses that specific problem
    3. Predict what will happen when you apply the fix
    4. Run tests to verify your understanding"

  5. Follow Your Own Guidelines

  "Remember and follow the development guidelines in CLAUDE.md:
    - Use KISS and DRY principles
    - Be careful and methodical
    - Follow testing philosophy (what fixtures are for)
    - Don't assume - verify with tools"

  6. Admit Uncertainty

  "When you don't know something, say 'I don't know, let me investigate' rather than making educated guesses. Use tools to gather facts before reasoning."

  Red Flags That Should Trigger Caution:

    - When someone questions your reasoning → Stop and re-investigate
    - When tests pass but "shouldn't" → Investigate why instead of assuming
    - When making changes to tests → Double-check what the test actually does
    - When CI behavior seems inconsistent → Check actual CI config

  Better Session Management:

  "At the start of complex debugging:
    1. First, use TodoWrite to break down the investigation steps
    2. Work through each step methodically with tools
    3. Update todos as you learn facts
    4. Only propose solutions after gathering complete information"


### Testing Philosophy
- Focus on behavior and contracts, not implementation
- Avoid mocking internal methods or testing private attributes
- Tests should document expected usage patterns
- Don't test third-party libraries
- Follow modern testing principles and best practices

## Helper Tools
Suggest and use suitable tools if applicable. If not installed yet, plan the installation in Todo and ask if user want to install the tool(s).
If the language/framework has built-in tool as officially recommend, and the best practice is to use it (ex, gofmt), please always remember to use them, follow official suggestions.

Snyk (security scan):
  - Scan code with `snyk test` and `snyk code test`.
  - `snyk test` command scans your project, tests dependencies for vulnerabilities, and reports how many vulnerabilities are found.
  - `snyk code test` analyzes source code for security issues, often referred to as Static Application Security Testing (SAST).


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
- File hygiene (trailing whitespace, EOF, merge conflicts) - **excludes fixtures/**
- Go tools (gofmt, goimports, go vet, go mod tidy, golangci-lint)
- Build verification (go build, go mod verify)
- Security-focused using local system tools only
- **Fixture protection**: Test fixtures excluded from formatting to preserve authentic output

### CI/CD Workflows
Located in `.github/workflows/`:
- **test-and-coverage.yml**: Go 1.23/1.24 testing with coverage
- **lint-and-format.yml**: Code quality and formatting checks
- **build.yml**: Multi-version build verification
- **release-binaries.yml**: Cross-platform binary releases

## Architecture Overview

For detailed technical architecture, design patterns, and implementation guidelines, see **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)**.

**Quick Reference:**
- **Core Interfaces**: `PackageManager` and `SysPkg` (interface.go)
- **CommandBuilder Pattern**: Target architecture for Issue #20
- **Package Structure**: `/cmd`, `/manager`, `/osinfo`, `/testing`
- **Testing Strategy**: Three-layer approach (unit, integration, mock)
- **Exit Code Complexity**: Each PM has unique behaviors (see docs/EXIT_CODES.md)

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

**Cross-Package Manager Compatibility**: SysPkg normalizes package states for consistent behavior across different package managers. For example, APT's "config-files" state (packages removed but with configuration files remaining) is normalized to "available" status to match the semantics used by other package managers like YUM and Snap.

## Project Improvement Roadmap

*Note: To-do list consolidated 2025-05-30 - removed duplicates, feature creep items, and over-engineering. Focused on core security, testing, and platform support.*

### 🔴 High Priority (Security & Critical Bugs) - 15 items
1. **Fix command injection vulnerability** - validate/sanitize package names before exec.Command
2. **Implement input validation helper function** for package names and arguments
3. **Fix resource leaks** in error handling paths
4. **Add security scanning with Snyk** to CI/CD pipeline
5. **Review and merge PR #12** - fix GetPackageManager("") panic bug ✅
6. **Cross-package manager status normalization** ✅ - APT config-files → available
7. **GitHub workflow compatibility fixes** ✅ - Go 1.23.4, Docker multi-OS testing
8. **Fix APT exit code bug** - Remove incorrect handling of exit code 100 as "no packages found" (it means error!)
9. **Fix Snap exit code bug** - Remove incorrect handling of exit code 64 as "no packages found" (it means usage error!)
10. **Implement CommandBuilder interface (Issue #20)** - Replace direct exec.Command calls with testable CommandBuilder pattern
11. **Add exit code documentation** ✅ - Created comprehensive exit code docs for all package managers

### ✅ CRITICAL INVESTIGATION COMPLETED
**Investigation Results: No design flaw found - tests are correct**
8. **RESOLVED: Tests are correctly testing parser functions** ✅ - `behavior_test.go` tests `ParseFindOutput()` directly, not `Find()` method
9. **RESOLVED: ParseFindOutput() vs Find() distinction clarified** ✅:
   - `ParseFindOutput()`: Pure parser, returns all packages as "available" (YUM limitation)
   - `Find()`: Enhanced method that adds rpm -q status detection for accurate results
10. **RESOLVED: CommandRunner interface verified** ✅ - Correctly implemented in `enhancePackagesWithStatus()`
11. **RESOLVED: Test execution paths confirmed** ✅ - Tests correctly test parsers, not enhanced methods
12. **RESOLVED: Fixtures validated as authentic** ✅ - Real YUM output that correctly lacks status info
13. **RESOLVED: Interface implementation verified** ✅ - All methods properly implemented and registered
14. **RESOLVED: Created integration tests** ✅ - Added `yum_integration_test.go` demonstrating three-layer testing approach

### 🟡 Medium Priority (Code Quality & Testing) - 8 items
**Testing:**
- **Create integration tests** ✅ - Added `yum_integration_test.go` with three-layer testing approach
- **Document testing strategy** ✅ - Added comprehensive testing documentation to CONTRIBUTING.md
- **Implement CommandRunner dependency injection (Issue #20)** 🚧 - YUM partially implemented, architecture decision updated to Option C (CommandBuilder)
- Add unit tests for snap package manager
- Add unit tests for flatpak package manager
- **APT fixture cleanup and behavior testing** ✅ - Reduced 16→7 fixtures, full test coverage
- **Cross-platform parsing robustness** ✅ - CRLF/whitespace handling, regex optimization
- **YUM fixture analysis and cleanup** ✅ **COMPLETED** (Issue #16) - Following APT pattern:
  - ✅ Analyzed YUM fixtures to determine what's needed for comprehensive testing
  - ✅ Removed redundant/duplicate files (search-vim-rockylinux.txt)
  - ✅ Standardized fixture naming convention (rocky8 vs rockylinux inconsistency)
  - ✅ Renamed info-vim-rockylinux.txt → info-vim-installed-rocky8.txt for clarity
  - ✅ Added missing edge case fixtures (empty results, not found, clean, refresh)
  - ✅ Created comprehensive behavior_test.go following APT fixture pattern
  - ✅ Converted YUM tests from inline data to fixture-based tests
  - ✅ Verified fixture compatibility and completeness with all tests passing
- **YUM operations implementation** ✅ **COMPLETED** - Comprehensive YUM package manager:
  - ✅ Implemented all missing operations: Install, Delete, ListUpgradable, Upgrade, UpgradeAll, AutoRemove
  - ✅ Added complete parser functions for all operation outputs
  - ✅ Created comprehensive behavior tests covering all operations and edge cases
  - ✅ Generated real fixtures using Rocky Linux Docker for authentic test data
  - ✅ Documented YUM-specific behaviors and cross-package manager compatibility
  - ✅ **YUM Find() status detection** - Implemented rpm -q integration for accurate installation status
  - ✅ **Cross-package manager API consistency** - YUM Find() now matches APT behavior exactly
  - ✅ All tests passing with 100% security scan clearance

**Documentation:**
- **API and behavior documentation** ✅ - Enhanced interface docs, status normalization, cross-PM compatibility
- **Error handling best practices** ✅ - Fixed ignored errors in documentation examples
- **Accuracy improvements** ✅ - Fixed misleading comments about status handling
- **YUM documentation updates** ✅ - Updated all outdated behavior comments to reflect Find() status detection capabilities

**Code Improvements:**
- Implement context support for cancellation and timeouts
- Create custom error types for better error handling
- Extract common parsing logic to shared utilities (DRY principle)
- Replace magic strings/numbers with named constants
- **Fix APT multi-arch package parsing** (Issue #15) - cosmetic fix for empty package names

**Removed from roadmap (2025-05-30):**
- ~~Structured logging~~ (over-engineering for project scope)
- ~~Progress indicators~~ (feature creep for CLI/library)
- ~~Architecture diagrams~~ (low ROI for library project)
- ~~TODO comment fixes~~ (covered by security improvements)

### 🟢 Low Priority (Platform Support) - 3 items
**New Package Managers:**
- Add proper macOS support with brew package manager implementation
- Add Windows support with chocolatey/scoop/winget package managers
- Implement dnf package manager support (Red Hat/Fedora) - uses yum backend
- ~~Implement yum package manager support (Red Hat/Fedora)~~ ✅ **COMPLETED**

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
env, err := testenv.GetTestEnvironment()
if err != nil {
    t.Fatalf("failed to get test environment: %v", err)
}
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

**Docker Matrix**: Tests run across multiple OSes in parallel:
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

Fixtures are automatically generated from real package manager outputs across different OSes:
- `testing/fixtures/apt/search-vim-ubuntu22.txt`
- `testing/fixtures/yum/search-vim-rocky8.txt`
- `testing/fixtures/dnf/search-vim-fedora39.txt`

This ensures parsers work correctly with real-world output variations across distributions.

See `testing/docker/`, `testing/os-matrix.yaml`, and [CONTRIBUTING.md](CONTRIBUTING.md) for complete details.
