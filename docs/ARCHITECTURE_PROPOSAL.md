# Architecture Proposal: Supporting Diverse Package Management Systems

## Executive Summary

After extensive research into various package management ecosystems, it's clear that go-syspkg needs a more flexible architecture to support not just system package managers (APT, YUM, Snap, Flatpak) but also:

- **Version Managers**: nvm, asdf, sdkman, rbenv, pyenv
- **Language-Specific Package Managers**: npm, pip, cargo, gem, composer
- **Scientific Computing**: conda, mamba, micromamba, bioconda
- **C++ Package Managers**: vcpkg, conan
- **Game Managers**: Steam, Lutris, GOG
- **Container/App Managers**: Docker, Podman, Helm
- **Windows Package Managers**: winget, chocolatey, scoop
- **Unified Interfaces**: UniGetUI (Windows), asdf (multi-language)

## Key Insights from Research

### 1. Fundamental Differences

Each category of package manager has fundamentally different:
- **Installation Patterns**: System-wide vs user-local vs project-specific
- **Version Management**: Single active version vs multiple concurrent versions
- **Dependency Models**: OS packages vs language modules vs containerized apps
- **State Management**: Installed/removed vs active/inactive versions
- **Scope**: System libraries vs development tools vs applications vs games

### 2. Common Patterns

Despite differences, most package managers share:
- List/Search capabilities
- Install/Remove operations
- Version querying
- Some form of update mechanism

### 3. Unique Features

Many managers have unique concepts that don't map to others:
- **asdf/nvm**: "use" command to switch active version
- **conda**: Environments with isolated dependency graphs
- **vcpkg/conan**: Build configurations and toolchains
- **Steam/Lutris**: Game-specific operations like verify/repair
- **Docker/Podman**: Image vs container lifecycle

## Proposed Architecture

### 1. Flexible Interface Hierarchy

Instead of one monolithic `PackageManager` interface, create a hierarchy:

```go
// Core interface - minimal common operations
type PackageManager interface {
    IsAvailable() bool
    GetManagerType() string
    GetManagerName() string
}

// Searchable - can search/list packages
type Searchable interface {
    Search(query string, opts *Options) ([]PackageResult, error)
    List(filter ListFilter, opts *Options) ([]PackageResult, error)
}

// Installable - can install/remove
type Installable interface {
    Install(packages []string, opts *Options) ([]PackageResult, error)
    Remove(packages []string, opts *Options) ([]PackageResult, error)
}

// Versionable - supports multiple versions
type Versionable interface {
    ListVersions(package string, opts *Options) ([]Version, error)
    SetActiveVersion(package string, version string, opts *Options) error
    GetActiveVersion(package string, opts *Options) (Version, error)
}

// Upgradable - supports updates
type Upgradable interface {
    Upgrade(packages []string, opts *Options) ([]PackageResult, error)
    ListUpgradable(opts *Options) ([]PackageResult, error)
}

// Environments - supports isolated environments
type EnvironmentManager interface {
    CreateEnvironment(name string, opts *Options) error
    ListEnvironments(opts *Options) ([]Environment, error)
    ActivateEnvironment(name string, opts *Options) error
    RemoveEnvironment(name string, opts *Options) error
}
```

### 2. Flexible Result Types

Replace rigid `PackageInfo` with flexible result types:

```go
// Generic result that can represent different concepts
type PackageResult struct {
    ID          string                 // Unique identifier
    Name        string                 // Human-readable name
    Type        string                 // package, version, environment, game, etc.
    Status      string                 // Manager-specific status
    Version     string                 // Current version if applicable
    Metadata    map[string]interface{} // Flexible metadata
}

// Examples of metadata usage:
// APT: {"architecture": "amd64", "section": "utils"}
// npm: {"global": true, "dependencies": [...]}
// Steam: {"appid": "220", "playtime": 1234}
// asdf: {"active": true, "path": "/home/user/.asdf/installs/node/16.0.0"}
```

### 3. Manager Categories

Define categories to help users understand capabilities:

```go
type ManagerCategory string

const (
    CategorySystem      ManagerCategory = "system"      // APT, YUM, etc.
    CategoryLanguage    ManagerCategory = "language"    // npm, pip, cargo
    CategoryVersion     ManagerCategory = "version"     // nvm, asdf, pyenv
    CategoryContainer   ManagerCategory = "container"   // docker, podman
    CategoryGame        ManagerCategory = "game"        // steam, lutris
    CategoryScientific  ManagerCategory = "scientific"  // conda, mamba
    CategoryBuild       ManagerCategory = "build"       // vcpkg, conan
)
```

### 4. Plugin Architecture

Move from static registration to dynamic plugin discovery:

```go
// Plugin interface for registering new package managers
type Plugin interface {
    Name() string
    Category() ManagerCategory
    CreateManager(config map[string]interface{}) (PackageManager, error)
}

// Registry for plugins
type Registry struct {
    plugins map[string]Plugin
}

func (r *Registry) Register(plugin Plugin) {
    r.plugins[plugin.Name()] = plugin
}

// Auto-discovery of plugins
func DiscoverPlugins(dirs []string) *Registry {
    // Look for .so files or Go plugins
    // Or use build tags for compile-time plugins
}
```

### 5. Unified Operations with Adapters

For common operations, provide adapters:

```go
// Adapter to make version managers work like package managers
type VersionManagerAdapter struct {
    vm Versionable
}

func (a *VersionManagerAdapter) Install(packages []string, opts *Options) ([]PackageResult, error) {
    // "Install" means download a new version
    results := []PackageResult{}
    for _, pkg := range packages {
        // Parse package@version format
        name, version := parsePackageVersion(pkg)
        // Implementation to download/install version
    }
    return results, nil
}
```

## Benefits of New Architecture

1. **Extensibility**: Easy to add new categories of package managers
2. **Flexibility**: Managers only implement interfaces they support
3. **Type Safety**: Compile-time checking of capabilities
4. **User Clarity**: Categories help users understand what each manager does
5. **Future Proof**: Can support unforeseen package manager types

## Migration Strategy

1. **Phase 1**: Create new interfaces alongside existing ones
2. **Phase 2**: Implement adapters for current package managers
3. **Phase 3**: Migrate existing managers to new interfaces
4. **Phase 4**: Deprecate old interfaces
5. **Phase 5**: Add new package manager types

## Example Implementations

### System Package Manager (APT)
```go
type APTManager struct {
    runner CommandRunner
}

// Implements: PackageManager, Searchable, Installable, Upgradable
func (a *APTManager) GetManagerType() string { return "system" }
func (a *APTManager) Install(packages []string, opts *Options) ([]PackageResult, error) {
    // Current implementation
}
```

### Version Manager (asdf)
```go
type ASDFManager struct {
    runner CommandRunner
}

// Implements: PackageManager, Searchable, Installable, Versionable
func (a *ASDFManager) GetManagerType() string { return "version" }
func (a *ASDFManager) SetActiveVersion(pkg, version string, opts *Options) error {
    // Run: asdf global <pkg> <version>
}
```

### Game Manager (Steam)
```go
type SteamManager struct {
    steamcmd CommandRunner
}

// Implements: PackageManager, Searchable, Installable, custom GameManager interface
func (s *SteamManager) VerifyIntegrity(appID string) error {
    // Steam-specific operation
}
```

## Conclusion

This architecture provides the flexibility to support any type of package manager while maintaining type safety and clear interfaces. It follows the "Less is more" principle by:

1. **Minimal Core Interface**: Only requires 3 methods
2. **Composition over Inheritance**: Use interface composition
3. **No Forced Abstractions**: Managers only implement what makes sense
4. **Clear Separation**: Each interface has a single responsibility

The architecture is designed to grow with the ecosystem while maintaining backward compatibility through adapters.
