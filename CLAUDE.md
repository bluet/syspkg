# CLAUDE.md - go-syspkg Project Guidelines

This extends `~/.claude/CLAUDE.md` with project-specific guidance.
Follow all universal principles there, plus these project-specific rules.

## Project Overview

**go-syspkg**: Universal interface for system package managers (apt, yum, snap, flatpak, etc.)

**Philosophy**: Tool-focused, not OS-focused. If apt works in a container or on macOS, we support it.
**Architecture**: Unified interface with 11 package operations, plugin system, CommandRunner pattern.

**Cross-Package Manager Compatibility**: SysPkg normalizes package states for consistent behavior across different package managers. For example, APT's "config-files" state (packages removed but with configuration files remaining) is normalized to "available" status to match the semantics used by other package managers like YUM and Snap.

## 🚨 Critical Safety Rules

1. **ALWAYS use Docker** for package manager operations
   - NEVER run package operations on development system
   - Use `make test-docker-*` for testing
   - Fixtures are generated in Docker for safety

2. **NEVER modify fixture files**
   - Located in `testing/fixtures/`
   - Contain authentic package manager outputs
   - Excluded from formatters to preserve exact output

3. **Input validation is mandatory**
   - Package names must be validated before exec.Command
   - See `manager/security.go` for validation helpers
   - Command injection is a real risk

4. **CommandRunner pattern required**
   - All package managers must use unified interface
   - See APT/YUM implementations as reference
   - Enables testing without system calls

## Testing Rules (CRITICAL)

1. **Fixtures are Primary Test Data**
   - Use real fixture files from `testing/fixtures/` in unit tests
   - Fixtures contain full raw outputs from actual package managers
   - Inline mocks are for quick tests and edge cases only

2. **Testing Hierarchy**
   - **Unit tests**: Use fixtures (safe, fast, realistic)
   - **Integration tests**: Run in Docker containers
   - **System tests**: Only in CI or dedicated environments

3. **Snap Testing Limitations**
   - Snap doesn't work in Docker (requires snapd daemon)
   - Use GitHub Actions native runners for Snap
   - Consider LXC/VM for local Snap testing

## Development Workflow

### Branch Strategy
- Current main branch: `refactor-unified-interface`
- Feature branches from main
- Small, focused commits

### Testing Requirements
```bash
# Daily development - smart OS detection
make test

# Package manager development
make test-docker-rocky    # YUM testing
make test-docker-ubuntu   # APT testing
make test-docker-all      # Comprehensive

# Before commits
make check               # Formatting and linting
```

### When to Use Docker vs Native
- **Native (`make test`)**: Parser tests, core logic, daily development
- **Docker required**: Integration tests, fixture generation, cross-PM testing
- **Never native**: Package installation/removal operations

### Commit Guidelines
- Prefix: `fix:`, `feat:`, `docs:`, `test:`, `refactor:`
- Scope: `(apt)`, `(yum)`, `(core)`, etc.
- Example: `fix(apt): correct exit code handling for package not found`
- Run `make check` before committing

## Essential Commands

```bash
# Build
make build              # Current platform
make build-all-arch     # All platforms

# Test
make test               # Smart OS-aware testing
make test-docker-ubuntu # APT in Ubuntu container
make test-docker-rocky  # YUM in Rocky Linux container

# Code Quality
make check              # Run before committing
make format             # Fix formatting issues
make lint               # Detailed linting
```

## Testing Strategy

### Fixture Naming Convention
`operation-description-context-os.txt`

Examples:
- `search-vim-ubuntu2204.txt`
- `install-vim.dry-run.clean-system.ubuntu-2204.txt`
- `list-installed.packages-installed.rocky-8.txt`

### Test Organization
- **Unit tests**: Use fixtures, test parsers (`*_test.go`)
- **Integration tests**: Real commands in Docker (`*_integration_test.go`)
- **Mock tests**: Full logic without system calls (CommandRunner)

### Exit Code Complexity
Each package manager has unique exit codes. Never assume:
- APT: 0=success, 1=general error, 100=dpkg error
- YUM: 0=success, 1=error, specific codes vary
- See `docs/EXIT_CODES.md` for details

## Important Project Requirements

- **License**: Apache License 2.0 (provides patent protection and enterprise-grade legal clarity)
- **Go Version**: Requires Go 1.23+ (CI tests with 1.23, 1.24)
- **Code Quality**: Always run `make check` before committing to ensure quality
- **Pre-commit**: Hooks automatically enforce formatting, linting, and security checks
- **Package Managers**: When implementing new ones, focus on parsing command outputs correctly
- **CLI Detection**: Automatically detects available package managers if no flag is specified
- **Privileges**: Root privileges are often required for package operations

## Current Development Focus

**Active Branch**: `refactor-unified-interface`
- APT ✅ Complete with all 11 package operations
- YUM ✅ Complete with fixtures
- Snap/Flatpak 🚧 Need migration to unified interface

**Priority Issues**:
- #28, #29: Snap/Flatpak CommandRunner migration
- #3: CLI upgrade display fix
- See [GitHub Issues](https://github.com/bluet/syspkg/issues)

## Documentation Verification

### Project-Specific Red Flags
- `GetType()` - doesn't exist, use `GetName()` or `GetCategory()`
- `NewPackageInfo(..., CategorySystem)` - 4th param is manager name like "apt"
- Interface definitions must match `manager/interfaces.go`
- Examples must compile with current interfaces

### Verification Commands
```bash
# Check for non-existent methods
rg "GetType|GetBestManager" docs/ examples/

# Verify NewPackageInfo usage
rg "NewPackageInfo.*Category" docs/ examples/

# Test examples work
go run examples/complete_demo.go
```

## Quick References

- **Architecture**: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
- **Testing Guide**: [docs/TESTING.md](docs/TESTING.md)
- **Exit Codes**: [docs/EXIT_CODES.md](docs/EXIT_CODES.md)
- **Contributing**: [CONTRIBUTING.md](CONTRIBUTING.md)
- **Plugin Development**: [docs/PLUGIN_DEVELOPMENT.md](docs/PLUGIN_DEVELOPMENT.md)

## Project Memories

- Package manager exit codes are NOT standardized - always verify
- Fixtures are sacred - they represent real PM behavior
- Think clearly about module responsibilities vs user/integrator responsibilities
