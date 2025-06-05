# syspkg Architecture Overview

## Status: ✅ Production Ready

This document provides a comprehensive overview of the implemented unified interface architecture in syspkg v2.0.

## Quick Start

```bash
# Build and use
go build -o bin/syspkg ./cmd/syspkg/
./bin/syspkg managers              # Show available package managers
./bin/syspkg search vim            # Search across all managers
./bin/syspkg install curl --dry-run # Dry run install

# Test
go test ./manager/apt/ -v          # Test APT plugin
go test ./manager/yum/ -v          # Test YUM plugin
go test ./manager/ -v              # Test unified interface
```

## Architecture Highlights

### 🎯 **Unified Interface**
- **Single API** for all package managers (APT, YUM, npm, Steam, conda, etc.)
- **13 essential operations** covering all package management needs
- **Type safety** with Go interfaces and compile-time checking

### 🔌 **Plugin System**
- **Auto-registration** via `init()` functions
- **Priority-based selection** for multiple managers of same type
- **~50 lines per plugin** - embed BaseManager, implement 2-3 methods

### 🛡️ **Security & Reliability**
- **Input validation** prevents command injection attacks
- **Context support** for timeouts and cancellation
- **Graceful degradation** with clear error messages

## Core Components

### 1. PackageManager Interface (`manager/interfaces.go`)

```go
type PackageManager interface {
    // Basic information
    GetName() string
    GetType() string
    IsAvailable() bool
    GetVersion() (string, error)

    // Core operations
    Search(ctx context.Context, query []string, opts *Options) ([]PackageInfo, error)
    List(ctx context.Context, filter ListFilter, opts *Options) ([]PackageInfo, error)
    Install(ctx context.Context, packages []string, opts *Options) ([]PackageInfo, error)
    Remove(ctx context.Context, packages []string, opts *Options) ([]PackageInfo, error)
    GetInfo(ctx context.Context, packageName string, opts *Options) (PackageInfo, error)

    // Update operations
    Refresh(ctx context.Context, opts *Options) error
    Update(ctx context.Context, opts *Options) error
    Upgrade(ctx context.Context, packages []string, opts *Options) ([]PackageInfo, error)

    // Cleanup operations
    Clean(ctx context.Context, opts *Options) error
    AutoRemove(ctx context.Context, opts *Options) ([]PackageInfo, error)

    // Advanced operations
    Verify(ctx context.Context, packages []string, opts *Options) ([]PackageInfo, error)
    Status(ctx context.Context, opts *Options) (ManagerStatus, error)
}
```

### 2. Package Information Structure

```go
type PackageInfo struct {
    Name        string                 `json:"name"`
    Version     string                 `json:"version"`
    NewVersion  string                 `json:"new_version"`
    Status      string                 `json:"status"`
    Description string                 `json:"description"`
    Category    string                 `json:"category"`
    ManagerType string                 `json:"manager_type"`
    Metadata    map[string]interface{} `json:"metadata"`     // Flexible!
}

// Examples:
// APT: Metadata["arch"] = "amd64"
// npm: Metadata["global"] = true, Metadata["registry"] = "..."
// Steam: Metadata["appid"] = "730", Metadata["playtime"] = 1234
```

### 3. BaseManager (`manager/base.go`)

Provides 90% of functionality for new plugins:

```go
type BaseManager struct {
    name        string
    managerType string
    runner      CommandRunner
}

// New plugins only need to implement 2-3 methods:
// - IsAvailable() bool
// - Search() or specific operations they support
// - Custom parsing for their package manager output
```

### 4. Plugin Registration (`manager/registry.go`)

```go
// Auto-registration via init() functions
func init() {
    _ = manager.Register("apt", &Plugin{})
}

// Thread-safe registry with priority-based selection
func GetBestManager(managerType string) PackageManager
func GetAvailableManagers() map[string]PackageManager
```

## Implementation Architecture

### Package Manager Types

```go
const (
    TypeSystem    = "system"    // APT, YUM, DNF, APK, Pacman
    TypeLanguage  = "language"  // npm, pip, cargo, gem, composer
    TypeContainer = "container" // Docker, Podman, Helm
    TypeApp       = "app"       // Snap, Flatpak, AppImage
    TypeGame      = "game"      // Steam, Lutris, GOG
    TypeVersion   = "version"   // nvm, rbenv, pyenv, asdf
)
```

### Security Features

1. **Input Validation**:
   ```go
   func ValidatePackageNames(packages []string) error {
       for _, pkg := range packages {
           if containsInjection(pkg) {
               return fmt.Errorf("invalid package name: %s", pkg)
           }
       }
       return nil
   }
   ```

2. **Command Injection Prevention**:
   - Regex validation of package names
   - Parameterized command construction
   - No shell interpretation of user input

3. **Timeout Management**:
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
   defer cancel()
   ```

## Usage Examples

### Multi-Manager Search

```go
// Search across all available managers
managers := manager.GetAvailableManagers()
for name, pm := range managers {
    results, err := pm.Search(ctx, []string{"vim"}, opts)
    if err != nil {
        fmt.Printf("%s: %v\n", name, err)
    } else {
        fmt.Printf("%s found %d packages\n", name, len(results))
    }
}
```

### Type-Specific Operations

```go
// Get best system package manager (APT on Ubuntu, YUM on RHEL)
systemPM := manager.GetBestManager(manager.TypeSystem)
systemPM.Install(ctx, []string{"vim"}, opts)

// Get application manager
appPM := manager.GetBestManager(manager.TypeApp)
appPM.Install(ctx, []string{"discord"}, opts)
```

### CLI Integration

```bash
# Automatic manager detection
syspkg search vim

# Specific manager
syspkg search vim --manager apt

# Cross-manager operations
syspkg managers --type system
```

## Testing Architecture

### Three-Layer Testing Strategy

1. **Unit Tests**: Parser functions with authentic fixtures
2. **Integration Tests**: Real command execution in Docker
3. **System Tests**: Full operations with actual package managers

### Fixture-Based Testing

```go
// Use real command outputs for authentic testing
fixture := testutil.LoadAPTFixture(t, "search-vim.clean-system.ubuntu-2204.txt")
packages := parseSearchOutput(fixture)
```

### Docker Safety

```bash
# All integration tests run in containers
make test-docker-ubuntu       # APT testing
make test-docker-rocky        # YUM testing
make test-docker-all          # Cross-platform testing
```

## Current Implementation Status

### ✅ Production Ready Package Managers

| Manager | Status | Test Coverage | Fixtures | Security |
|---------|--------|---------------|----------|----------|
| **APT** | ✅ Complete | 100% | 44 authentic | ✅ Full |
| **YUM** | ✅ Complete | 100% | 29 authentic | ✅ Full |

### 🚧 In Development

| Manager | Status | Priority |
|---------|--------|----------|
| **Snap** | Partial | High |
| **Flatpak** | Partial | High |
| **DNF** | Planned | Medium |
| **APK** | Planned | Medium |

## Performance Characteristics

- **Startup Time**: <50ms (plugin registration)
- **Memory Usage**: <10MB baseline
- **Command Execution**: Inherits from underlying package manager
- **Concurrent Operations**: Thread-safe registry and operations

## Error Handling Strategy

1. **Graceful Degradation**: Operations return clear errors for unsupported features
2. **Type Safety**: Compile-time checking prevents runtime errors
3. **Validation**: Input validation prevents injection attacks
4. **Context Awareness**: Proper timeout and cancellation support

## Extension Points

### Adding New Package Managers

1. **Create plugin file** (e.g., `manager/npm/plugin.go`)
2. **Embed BaseManager** for common functionality
3. **Implement 2-3 required methods** (IsAvailable, Search, etc.)
4. **Add auto-registration** via `init()` function
5. **Add tests** with authentic fixtures

### Custom Operations

```go
// Extend interface for manager-specific features
type ExtendedManager interface {
    PackageManager
    CustomOperation(args ...string) error
}
```

## File Organization

```
manager/
├── interfaces.go       # Core PackageManager interface
├── base.go            # BaseManager implementation
├── registry.go        # Plugin registration system
├── command_runner.go  # Command execution abstraction
├── security.go        # Input validation and security
├── apt/               # APT plugin implementation
├── yum/               # YUM plugin implementation
└── [new-manager]/     # Template for new managers
```

## Related Documentation

- **[Plugin Development Guide](PLUGIN_DEVELOPMENT.md)** - How to create new package manager plugins
- **[Production Guide](PRODUCTION_GUIDE.md)** - Advanced development and deployment
- **[Testing Guide](TESTING.md)** - Comprehensive testing strategies
- **[Exit Codes Reference](EXIT_CODES.md)** - Package manager behavior reference

---

**Architecture Status**: ✅ **Production Ready**
**Last Updated**: 2025-06-06
**Version**: 2.0 (Unified Interface)
