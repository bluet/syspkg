# Unified Interface Architecture Design

## Overview

This document describes the new unified interface architecture for go-syspkg that enables easy addition of package managers while following the "Less is more" principle.

## Design Goals

1. **Unified Experience**: Same interface whether using APT, npm, Steam, conda, etc.
2. **Easy Plugin Addition**: New package managers require ~50 lines of code
3. **No Constraints**: Plugins only implement operations they support
4. **Graceful Degradation**: Unsupported operations return clear errors
5. **Type Safety**: Compile-time checking with Go interfaces
6. **Extensibility**: Support any future package management system

## Architecture Components

### 1. Core Interface (`manager/interfaces.go`)

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
    Upgrade(ctx context.Context, packages []string, opts *Options) ([]PackageInfo, error)
    
    // Cleanup operations
    Clean(ctx context.Context, opts *Options) error
    AutoRemove(ctx context.Context, opts *Options) ([]PackageInfo, error)
    
    // Health operations
    Verify(ctx context.Context, packages []string, opts *Options) ([]PackageInfo, error)
    Status(ctx context.Context, opts *Options) (ManagerStatus, error)
}
```

### 2. Plugin System (`manager/registry.go`)

- Thread-safe plugin registration and discovery
- Priority-based selection for multiple managers of same type
- Auto-registration pattern via `init()` functions

```go
type Plugin interface {
    CreateManager() PackageManager
    GetPriority() int
}

func Register(name string, plugin Plugin) error
func GetAvailableManagers() map[string]PackageManager
```

### 3. Base Manager (`manager/base.go`)

Provides sensible defaults for 90% of functionality:

```go
type BaseManager struct {
    name        string
    managerType string
    runner      CommandRunner
}

// Default implementations return ErrOperationNotSupported
// Plugins override only what they need
```

## Key Benefits

### 1. "Less is More" Implementation

**Before** (complex, rigid):
```go
// Old approach - forced implementation
type APTManager struct{}
func (a *APTManager) AutoRemove() []PackageInfo {
    // APT supports this - complex implementation
}

type SteamManager struct{}  
func (s *SteamManager) AutoRemove() []PackageInfo {
    // Steam doesn't support this - but forced to implement
    panic("not supported")
}
```

**After** (simple, flexible):
```go
// New approach - implement only what you support
type APTManager struct {
    *manager.BaseManager
}
func (a *APTManager) AutoRemove(ctx context.Context, opts *Options) ([]PackageInfo, error) {
    // APT supports this - implement it
}

type SteamManager struct {
    *manager.BaseManager  
}
// Steam doesn't support AutoRemove - BaseManager provides default error
```

### 2. Easy Plugin Development

Adding support for a new package manager:

```go
// 1. Create manager struct
type PipManager struct {
    *manager.BaseManager
}

func NewPipManager() *PipManager {
    return &PipManager{
        BaseManager: manager.NewBaseManager("pip", manager.TypeLanguage, runner),
    }
}

// 2. Implement only supported operations
func (p *PipManager) Search(ctx context.Context, query []string, opts *manager.Options) ([]manager.PackageInfo, error) {
    // pip search implementation
}

func (p *PipManager) Install(ctx context.Context, packages []string, opts *manager.Options) ([]manager.PackageInfo, error) {
    // pip install implementation  
}

// 3. Create plugin
type PipPlugin struct{}
func (p *PipPlugin) CreateManager() manager.PackageManager { return NewPipManager() }
func (p *PipPlugin) GetPriority() int { return 80 }

// 4. Auto-register
func init() {
    manager.Register("pip", &PipPlugin{})
}
```

### 3. Flexible Package Information

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

// Get language package manager
langPM := manager.GetBestManager(manager.TypeLanguage) 
langPM.Install(ctx, []string{"react"}, opts)
```

### Graceful Degradation

```go
// Operations that aren't supported return clear errors
gameManager, _ := manager.GetManagerByName("steam")
_, err := gameManager.AutoRemove(ctx, opts)
// Returns: "steam: operation not supported by this package manager"
```

## Supported Manager Types

```go
const (
    TypeSystem     = "system"     // APT, YUM, DNF, etc.
    TypeLanguage   = "language"   // npm, pip, cargo, gem, etc.
    TypeVersion    = "version"    // nvm, asdf, pyenv, rbenv, etc.
    TypeContainer  = "container"  // docker, podman, helm, etc.
    TypeGame       = "game"       // steam, lutris, gog, etc.
    TypeScientific = "scientific" // conda, mamba, bioconda, etc.
    TypeBuild      = "build"      // vcpkg, conan, cmake, etc.
    TypeApp        = "app"        // flatpak, snap, appimage, etc.
)
```

## Implementation Status

### ✅ Completed Core Components

1. **Interface Design** - 13 standard operations covering all use cases
2. **Plugin System** - Thread-safe registry with priority-based selection
3. **Base Manager** - Default implementations and helper methods
4. **Working Demo** - Demonstrates APT, npm, and Steam managers
5. **Documentation** - Comprehensive plugin development guide

### 🚧 Next Steps

1. **Migrate Existing Managers** - Update APT, YUM to new interface
2. **Add Real Implementations** - npm, pip, conda, steam managers
3. **CLI Integration** - Update command-line interface
4. **Testing Framework** - Comprehensive test suite
5. **Performance Optimization** - Benchmarking and optimization

## Testing the Architecture

Run the working demo:

```bash
go run examples/working_demo.go
```

Output shows:
- ✅ Plugin registration and discovery
- ✅ Unified interface across package types  
- ✅ Graceful error handling for unsupported operations
- ✅ Flexible metadata for package-specific data
- ✅ Type safety with Go interfaces

## Future Extensibility

The architecture supports any package management system:

- **Version Managers**: `asdf`, `mise`, `g` (Go), `rustup`
- **Container Tools**: `kubectl`, `helm`, `skaffold`  
- **Cloud Tools**: `aws-cli`, `gcloud`, `azure-cli`
- **Development Tools**: `git`, `terraform`, `ansible`
- **Media Tools**: `youtube-dl`, `ffmpeg`
- **And beyond**: Any tool that manages "packages" of any kind

## Design Principles Applied

1. **Single Responsibility**: Each interface has one clear purpose
2. **Open/Closed**: Open for extension (new plugins), closed for modification
3. **Interface Segregation**: Plugins implement only what they support
4. **Dependency Inversion**: Depend on abstractions, not implementations
5. **Composition over Inheritance**: BaseManager provides behavior via composition

This architecture achieves the goal of making package manager addition trivial while maintaining consistency and type safety across the entire ecosystem.