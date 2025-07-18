# SysPkg Project Structure

## Root Directory Organization

```
syspkg/
├── cmd/syspkg/           # CLI application entry point
├── manager/              # Core package manager implementations
├── osinfo/              # OS detection utilities
├── slog-levelmulti/     # Custom logging library
├── testing/             # Testing infrastructure
├── docs/                # Documentation
├── examples/            # Usage examples
├── backup/              # Legacy code (pre-refactor)
└── tmp/                 # Temporary development files
```

## Core Implementation (`manager/`)

### Interface & Registry
- `interfaces.go` - Core `PackageManager` interface definition
- `registry.go` - Thread-safe plugin registration system
- `base.go` - `BaseManager` with common functionality (90% of plugin needs)

### Infrastructure
- `command_runner.go` - Command execution abstraction
- `security.go` - Input validation and injection prevention
- `error_types_test.go` - Error handling patterns

### Package Manager Plugins
```
manager/
├── apt/           # APT (Ubuntu/Debian)
├── yum/           # YUM (RHEL/Rocky/CentOS)
├── apk/           # APK (Alpine Linux)
├── snap/          # Snap packages
├── flatpak/       # Flatpak packages
└── [future]/      # DNF, Pacman, Zypper, etc.
```

## CLI Application (`cmd/syspkg/`)

### Main Components
- `main.go` - Entry point, argument parsing, command dispatch
- `dispatcher.go` - Command routing and execution
- `formatters.go` - Output formatting (JSON, table, human-readable)
- `handlers.go` - Command-specific logic
- `validators.go` - Input validation and safety checks

## Testing Infrastructure (`testing/`)

### Docker Testing
```
testing/docker/
├── docker-compose.test.yml    # Multi-OS test orchestration
├── ubuntu.Dockerfile          # APT testing environment
├── rockylinux.Dockerfile      # YUM testing environment
├── fedora.Dockerfile          # DNF testing environment
└── alpine.Dockerfile          # APK testing environment
```

### Test Fixtures
```
testing/fixtures/
├── apt/           # Real APT command outputs
├── yum/           # Real YUM command outputs
├── dnf/           # Real DNF command outputs
├── snap/          # Real Snap command outputs
├── flatpak/       # Real Flatpak command outputs
└── apk/           # Real APK command outputs
```

### Test Utilities
- `testutil/fixtures.go` - Fixture loading helpers
- `entrypoints/` - Docker container entry scripts for fixture generation

## Documentation (`docs/`)

### Technical Documentation
- `ARCHITECTURE.md` - System design and interfaces
- `TESTING.md` - Comprehensive testing strategy
- `PLUGIN_DEVELOPMENT.md` - Guide for adding new package managers
- `EXIT_CODES.md` - Package manager behavior reference
- `INTEGRATION_GUIDE.md` - Production usage patterns
- `PRODUCTION_GUIDE.md` - Advanced development practices

## Utilities & Libraries

### OS Detection (`osinfo/`)
- `osinfo.go` - Cross-platform OS detection
- `osinfo_test.go` - OS detection validation

### Custom Logging (`slog-levelmulti/`)
- Multi-level structured logging library
- Used for debugging and development

## Development Files

### Configuration
- `Makefile` - Build system and common commands
- `go.mod` - Go module definition (minimal dependencies)
- `.golangci.yml` - Linting configuration
- `.pre-commit-config.yaml` - Quality check automation

### CI/CD
- `.github/workflows/` - GitHub Actions for testing and builds
- `Dockerfile.test` - Testing container configuration

## Naming Conventions

### Package Manager Plugins
- Directory: `manager/[manager-name]/`
- Main file: `plugin.go` (implements interface)
- Tests: `plugin_test.go`
- Utilities: `utils.go` (parsing functions)

### Test Fixtures
- Format: `{operation}.{mode}.{state}.{distro}-{version}.txt`
- Examples:
  - `search-vim.clean-system.ubuntu-2204.txt`
  - `install-vim.dry-run.clean-system.rocky-8.txt`
  - `list-installed.packages-installed.fedora-39.txt`

### File Organization Principles
- **Separation of Concerns**: CLI, core logic, and plugins are separate
- **Plugin Isolation**: Each package manager is self-contained
- **Test Safety**: All destructive tests run in Docker containers
- **Documentation Co-location**: Each component has adjacent documentation
