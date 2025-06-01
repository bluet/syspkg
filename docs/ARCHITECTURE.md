# SysPkg Architecture

This document describes the technical architecture, design patterns, and core interfaces of the SysPkg project.

## Core Interfaces

### PackageManager Interface
The `PackageManager` interface (defined in `interface.go`) defines methods that all package managers must implement:

```go
type PackageManager interface {
    IsAvailable() bool
    GetPackageManager() string
    Install(pkgs []string, opts *Options) ([]PackageInfo, error)
    Delete(pkgs []string, opts *Options) ([]PackageInfo, error)
    Find(keywords []string, opts *Options) ([]PackageInfo, error)
    ListInstalled(opts *Options) ([]PackageInfo, error)
    ListUpgradable(opts *Options) ([]PackageInfo, error)
    Upgrade(pkgs []string, opts *Options) ([]PackageInfo, error)
    UpgradeAll(opts *Options) ([]PackageInfo, error)
    Refresh(opts *Options) error
    Clean(opts *Options) error
    GetPackageInfo(pkg string, opts *Options) (PackageInfo, error)
    AutoRemove(opts *Options) ([]PackageInfo, error)
}
```

### SysPkg Interface
The `SysPkg` interface provides high-level package management across multiple package managers:

```go
type SysPkg interface {
    FindPackageManagers(include IncludeOptions) (map[string]PackageManager, error)
    RefreshPackageManagers(include IncludeOptions) (map[string]PackageManager, error)
    GetPackageManager(name string) (PackageManager, error)
}
```

## Command Execution Architecture

### Current State (Issue #20)
Mixed patterns across package managers:
- **YUM**: Uses CommandRunner for some operations (Find), direct exec.Command for others
- **APT/Snap/Flatpak**: All use direct exec.Command calls

### Target Architecture: CommandBuilder Pattern (Option C)

```go
type CommandBuilder interface {
    CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd
}
```

**Why CommandBuilder (not Extended CommandRunner)**:
- **Exit code complexity**: Each PM has unique exit code behaviors (APT 100=error, YUM 100=success)
- **Maximum flexibility**: Full access to exec.Cmd features (env, stdin/stdout, working dir)
- **Simple interface**: Only 1 method vs multiple in extended interface
- **Go idiomatic**: Returns standard `*exec.Cmd` type
- **No generic helpers needed**: Each PM handles its own exit codes

**Critical Discovery**: Package managers have wildly inconsistent exit codes:
- APT: Exit code 100 = any error
- YUM: Exit code 100 = updates available (success!)
- Snap: Exit code 64 = usage error (not "no packages found")

See [EXIT_CODES.md](EXIT_CODES.md) for comprehensive documentation.

## Project Structure

### Directory Layout
```
syspkg/
├── cmd/syspkg/           # CLI application using urfave/cli/v2
├── manager/              # Package manager implementations
│   ├── apt/             # APT (Ubuntu/Debian)
│   ├── yum/             # YUM (Rocky Linux/AlmaLinux/RHEL)
│   ├── snap/            # Snap packages
│   ├── flatpak/         # Flatpak packages
│   ├── options.go       # Common options structure
│   └── packageinfo.go   # Common package info structure
├── osinfo/              # OS detection utilities
├── testing/             # Testing infrastructure
└── docs/                # Documentation
```

### Package Manager Implementation Structure
Each package manager follows a consistent structure:
```
manager/{pm}/
├── {pm}.go              # Main implementation (PackageManager interface)
├── utils.go             # Parser functions (Parse*Output)
├── {pm}_test.go         # Unit tests
├── behavior_test.go     # Fixture-based behavior tests
├── {pm}_integration_test.go  # Integration tests (when available)
└── EXIT_CODES.md        # Package manager specific exit code docs
```

## Key Design Patterns

### Interface-Based Abstraction
- Allows easy addition of new package managers
- Consistent API across all supported package managers
- Clear separation between interface and implementation

### Factory Pattern
Factory pattern in `syspkg.go` for creating manager instances:
```go
func New(includeOptions IncludeOptions) (*Impl, error) {
    packageManagers := initializePackageManagers(includeOptions)
    return &Impl{packageManagers: packageManagers}, nil
}
```

### Options Pattern
Configurable behavior using `manager.Options`:
```go
type Options struct {
    DryRun      bool
    Interactive bool
    Verbose     bool
    AssumeYes   bool
    Debug       bool
}
```

### Parser Pattern
Each package manager implements parser functions for different operations:
- `ParseInstallOutput()` - Parse installation results
- `ParseFindOutput()` - Parse search results
- `ParseListInstalledOutput()` - Parse installed package lists
- `ParsePackageInfoOutput()` - Parse detailed package information

## Cross-Package Manager Compatibility

### Status Normalization
SysPkg normalizes package states for consistent behavior:
- APT's "config-files" state maps to "available" status
- Consistent status reporting across all package managers

### Field Usage Patterns
Consistent field usage across operations:

| Operation | Version | NewVersion | Status |
|-----------|---------|------------|--------|
| **Install** | `installed_version` | `installed_version` | `installed` |
| **Delete** | `removed_version` | `""` | `available` |
| **Find** | `""` | `available_version` | `available/installed` |
| **ListInstalled** | `installed_version` | `""` | `installed` |
| **ListUpgradable** | `current_version` | `upgrade_version` | `upgradable` |

## Testing Architecture

### Three-Layer Testing Strategy

#### 1. Unit Tests (Pure Logic)
- Parser functions with fixtures
- OS detection logic
- Command construction
- No actual package manager execution

#### 2. Integration Tests (Real Commands)
- Real package manager availability checks
- Command output capture for test fixtures
- Limited package operations (list, search, show)

#### 3. Mock Tests (Full Logic)
- Test complete method logic with dependency injection
- Use MockCommandRunner for controlled responses
- Test error conditions and edge cases

### Environment-Aware Testing
Tests automatically detect the current OS and determine which package managers to test:

```go
env, err := testenv.GetTestEnvironment()
if skip, reason := env.ShouldSkipTest("yum"); skip {
    t.Skip(reason)
}
```

## CLI Command Structure

### Main Commands
- `install` - Install packages
- `delete`/`remove` - Remove packages
- `refresh` - Update package lists
- `upgrade` - Upgrade packages
- `find`/`search` - Search for packages
- `show` - Show package information

### Package Manager Flags
- `--apt` - Use APT package manager
- `--yum` - Use YUM package manager
- `--flatpak` - Use Flatpak package manager
- `--snap` - Use Snap package manager

### Options
- `--debug` - Enable debug output
- `--assume-yes` - Automatically answer yes to prompts
- `--dry-run` - Show what would be done without executing
- `--interactive` - Enable interactive mode
- `--verbose` - Enable verbose output

## Adding New Package Managers

### Implementation Steps

1. **Create package directory**: `manager/newpm/`

2. **Implement PackageManager interface**: `manager/newpm/newpm.go`
   ```go
   type PackageManager struct{}

   func (pm *PackageManager) IsAvailable() bool { ... }
   func (pm *PackageManager) Install(...) { ... }
   // ... implement all interface methods
   ```

3. **Add parser functions**: `manager/newpm/utils.go`
   ```go
   func ParseInstallOutput(output string, opts *manager.Options) []manager.PackageInfo { ... }
   func ParseSearchOutput(output string, opts *manager.Options) []manager.PackageInfo { ... }
   ```

4. **Create tests**: `manager/newpm/newpm_test.go`
   ```go
   func TestParseInstallOutput(t *testing.T) { ... }
   func TestNewPMAvailability(t *testing.T) { ... }
   ```

5. **Add to factory**: Update `initializePackageManagers()` in `syspkg.go`

6. **Document exit codes**: Create `manager/newpm/EXIT_CODES.md`

7. **Add Docker support**: `testing/docker/newos.Dockerfile`

8. **Update testing matrix**: `testing/os-matrix.yaml`

### Exit Code Documentation Requirements
**Critical**: Never assume exit codes work like other package managers!

- Document actual exit codes (not assumptions)
- Verify behavior through testing
- Document special cases and edge behaviors
- Provide testing commands for verification

## Philosophy

### Tool-Focused Approach
SysPkg focuses on supporting package manager tools based on their functionality rather than the operating system they're running on. If apt+dpkg work correctly in a container, on macOS via Homebrew, or in any other environment, SysPkg will support them.

### Cross-Package Manager Compatibility
SysPkg normalizes package states for consistent behavior across different package managers while preserving the unique characteristics of each tool.

### Interface-Driven Design
Clear interfaces allow for easy testing, mocking, and extension while maintaining backward compatibility.

## Related Documentation

- **[CONTRIBUTING.md](../CONTRIBUTING.md)** - Development workflow and testing guide
- **[EXIT_CODES.md](EXIT_CODES.md)** - Package manager exit code behavior
- **[../README.md](../README.md)** - Project overview and usage examples
- **[../CLAUDE.md](../CLAUDE.md)** - AI assistant development guidelines
