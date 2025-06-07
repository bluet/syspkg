# Plugin Development Guide

This guide explains how to create new package manager plugins for go-syspkg using the unified interface.

## Overview

The go-syspkg unified interface makes it incredibly easy to add support for any package management system. Whether you want to support:

- **Version Managers**: nvm, asdf, pyenv, rbenv
- **Language Package Managers**: npm, pip, cargo, gem, composer
- **Scientific Computing**: conda, mamba, bioconda
- **Build Tools**: vcpkg, conan, cmake
- **Game Managers**: Steam, Lutris, GOG
- **Container Tools**: Docker, Podman, Helm
- **System Tools**: systemd, homebrew

The process is the same: implement the unified interface and register your plugin.

## Testing Requirements (IMPORTANT)

When developing plugins, follow these testing guidelines:

### **Fixtures are Primary Test Data**
- Use real fixture files from `testing/fixtures/` in unit tests
- Fixtures contain full raw outputs from actual package managers
- Inline mocks are for quick tests and edge cases only

### **Docker for Safety**
- ALWAYS use Docker for fixture generation
- ALWAYS use Docker for integration testing
- NEVER run package manager operations on the development system // WRONG: dev system operations
- Use `make test-docker-*` commands for safe testing

### **Testing Hierarchy**
- Unit tests: Use fixtures (safe, fast, realistic)
- Integration tests: Run in Docker containers
- System tests: Only in CI or dedicated environments

This ensures your plugin works correctly with real-world command outputs and edge cases.

## Quick Start

### 1. Create Your Manager

```go
package mymanager

import (
    "context"
    "github.com/bluet/syspkg/manager"
)

// MyManager implements the unified PackageManager interface
type MyManager struct {
    *manager.BaseManager // Provides 90% of functionality for free
}

// NewMyManager creates a new instance
func NewMyManager() *MyManager {
    // BaseManager handles common operations, logging, validation, etc.
    base := manager.NewBaseManager("my-tool", manager.CategoryLanguage, manager.NewDefaultCommandRunner())
    return &MyManager{
        BaseManager: base,
    }
}
```

### 2. Override Methods You Need

```go
// IsAvailable checks if your tool is installed
func (m *MyManager) IsAvailable() bool {
    _, err := m.GetRunner().Run(context.Background(), "my-tool", []string{"--version"})
    return err == nil
}

// Search implements package search (only if your tool supports it)
func (m *MyManager) Search(ctx context.Context, query []string, opts *manager.Options) ([]manager.PackageInfo, error) {
    if opts == nil {
        opts = manager.DefaultOptions()
    }

    if err := m.ValidatePackageNames(query); err != nil {
        return nil, err
    }

    m.HandleDryRun(opts, "search", query)
    if opts.DryRun {
        return []manager.PackageInfo{}, nil
    }

    // Your search logic here
    args := append([]string{"search"}, query...)
    output, err := m.GetRunner().Run(ctx, "my-tool", args)
    if err != nil {
        return nil, err
    }

    return m.parseSearchOutput(string(output))
}

// Install implements package installation
func (m *MyManager) Install(ctx context.Context, packages []string, opts *manager.Options) ([]manager.PackageInfo, error) {
    if opts == nil {
        opts = manager.DefaultOptions()
    }

    if err := m.ValidatePackageNames(packages); err != nil {
        return nil, err
    }

    m.HandleDryRun(opts, "install", packages)
    if opts.DryRun {
        // Create dry-run results manually
        results := make([]manager.PackageInfo, len(packages))
        for i, pkg := range packages {
            results[i] = manager.NewPackageInfo(pkg, "unknown", "would-install", m.GetType())
        }
        return results, nil
    }

    // Your installation logic here
    args := append([]string{"install"}, packages...)
    output, err := m.GetRunner().Run(ctx, "my-tool", args)
    if err != nil {
        return nil, err
    }

    return m.parseInstallOutput(string(output))
}
```

### 3. Create Plugin Registration

```go
// Plugin represents your package manager plugin
type Plugin struct{}

func (p *Plugin) CreateManager() manager.PackageManager {
    return NewMyManager()
}

func (p *Plugin) GetPriority() int {
    return 70 // Medium priority - adjust based on your use case
}

// Auto-register when package is imported
func init() {
    if err := manager.Register("my-tool", &Plugin{}); err != nil {
        panic("Failed to register my-tool plugin: " + err.Error())
    }
}
```

### 4. That's It!

Your package manager is now available:

```go
import _ "your-package/mymanager"  // Auto-registers

managers := manager.GetAvailableManagers()
myMgr := managers["my-tool"]
```

## Interface Reference

### Required Methods

All package managers must implement these methods from the `PackageManager` interface:

```go
type PackageManager interface {
    // Basic info
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

    // Cleanup
    Clean(ctx context.Context, opts *Options) error
    AutoRemove(ctx context.Context, opts *Options) ([]PackageInfo, error)

    // Health
    Verify(ctx context.Context, packages []string, opts *Options) ([]PackageInfo, error)
    Status(ctx context.Context, opts *Options) (ManagerStatus, error)
}
```

### BaseManager Provides

The `BaseManager` provides sensible defaults for all methods:

- **Unsupported operations** return `ErrOperationNotSupported`
- **Input validation** via `ValidatePackageNames()`
- **Logging helpers** via `LogVerbose()`, `LogDebug()`
- **Dry run handling** via `HandleDryRun()`
- **Context-based timeout management** via standard Go patterns
- **Basic status** via default `Status()` implementation

### Only Override What You Need

```go
// Minimal implementation - only search and install
type MinimalManager struct {
    *manager.BaseManager
}

func (m *MinimalManager) Search(ctx context.Context, query []string, opts *manager.Options) ([]manager.PackageInfo, error) {
    // Your search implementation
}

func (m *MinimalManager) Install(ctx context.Context, packages []string, opts *manager.Options) ([]manager.PackageInfo, error) {
    // Your install implementation
}

// All other methods (Remove, Upgrade, etc.) automatically return "not supported"
```

## Manager Types

Use these predefined types for consistency:

```go
const (
    CategorySystem     = "system"     // OS package managers (apt, yum)
    CategoryLanguage   = "language"   // Language-specific (npm, pip, cargo)
    CategoryVersion    = "version"    // Version managers (nvm, asdf, pyenv)
    CategoryContainer  = "container"  // Container management (docker, podman)
    CategoryGame       = "game"       // Game managers (steam, lutris)
    CategoryScientific = "scientific" // Scientific computing (conda, mamba)
    CategoryBuild      = "build"      // Build tools (vcpkg, conan)
    CategoryApp        = "app"        // Application stores (flatpak, snap)
)
```

## Package Info Structure

Use the flexible `PackageInfo` structure:

```go
type PackageInfo struct {
    Name        string                 `json:"name"`
    Version     string                 `json:"version"`     // Current version
    NewVersion  string                 `json:"new_version"` // Available version (for upgrades)
    Status      string                 `json:"status"`      // installed, available, upgradable
    Description string                 `json:"description"`
    Category    string                 `json:"category"`
    ManagerType string                 `json:"manager_type"`
    Metadata    map[string]interface{} `json:"metadata"`    // Tool-specific data
}
```

### Using Metadata for Tool-Specific Data

```go
pkg := manager.NewPackageInfo("my-package", "1.0.0", "installed", m.GetType())

// Add tool-specific metadata
pkg.Metadata["repository"] = "my-repo"
pkg.Metadata["arch"] = "amd64"
pkg.Metadata["download_size"] = 1024000

// For Steam games
pkg.Metadata["appid"] = "123456"
pkg.Metadata["playtime"] = 3600

// For npm packages
pkg.Metadata["global"] = true
pkg.Metadata["dependencies"] = []string{"dep1", "dep2"}
```

## Options Handling

Always handle options properly:

```go
func (m *MyManager) SomeOperation(ctx context.Context, opts *manager.Options) error {
    if opts == nil {
        opts = manager.DefaultOptions()
    }

    // Use options
    if opts.Verbose {
        m.LogVerbose(opts, "Doing something...")
    }

    if opts.DryRun {
        m.LogVerbose(opts, "Would do something")
        return nil
    }

    // Context passed through directly - caller controls timeout
    // Your implementation using ctx
}
```

## Error Handling

### Standardized Return Status (Recommended)

The **best practice** is to use explicit status codes - plugin developers know exactly what happened:

```go
func (m *MyManager) Install(ctx context.Context, packages []string, opts *manager.Options) ([]manager.PackageInfo, error) {
    // Validation errors - you know this is a usage error
    if len(packages) == 0 {
        return nil, manager.WrapReturn(manager.StatusUsageError, "install requires package names", nil)
    }

    if err := m.ValidatePackageNames(packages); err != nil {
        return nil, manager.WrapReturn(manager.StatusUsageError, "invalid package names", err)
    }

    // Run the command
    args := append([]string{"install"}, packages...)
    result, err := m.GetRunner().Run(ctx, "my-tool", args)
    if err != nil {
        // Command execution failed (e.g., tool not found)
        return nil, manager.WrapReturn(manager.StatusUnavailableError, "my-tool command failed", err)
    }

    // Handle error cases first (return-early pattern)
    if result.ExitCode != 0 {
        switch result.ExitCode {
        case 1:
            // Check stderr for specific errors
            stderrStr := string(result.Stderr)
            if strings.Contains(stderrStr, "not found") {
                return nil, manager.WrapReturn(manager.StatusUnavailableError, "package not found", nil)
            }
            return nil, manager.WrapReturn(manager.StatusGeneralError, "installation failed", nil)
        case 2:
            // Invalid usage
            return nil, manager.WrapReturn(manager.StatusUsageError, "invalid arguments", nil)
        case 77:
            // Permission denied (if your tool uses this exit code)
            return nil, manager.WrapReturn(manager.StatusPermissionError, "requires root access", nil)
        default:
            return nil, manager.WrapReturn(manager.StatusGeneralError, "unknown error", nil)
        }
    }

    // result.ExitCode == 0: Success - parse and return results
    return m.parseInstallOutput(string(result.Output)), nil
}
```

### Available Return Status Codes

| Status Code | Exit Code | Usage |
|-------------|-----------|-------|
| `manager.StatusSuccess` | 0 | Operation succeeded |
| `manager.StatusUsageError` | 2 | Invalid arguments, malformed input |
| `manager.StatusPermissionError` | 77 | Permission denied, needs sudo |
| `manager.StatusUnavailableError` | 69 | Service/package not found |
| `manager.StatusGeneralError` | 1 | General failures |

### CommandResult Structure

The `CommandRunner.Run()` method returns a `*CommandResult` with complete command execution details:

```go
type CommandResult struct {
    Output   []byte // stdout
    Stderr   []byte // stderr
    ExitCode int    // exit code (0 = success)
}
```

### The One Function You Need

```go
manager.WrapReturn(status, message, wrappedError)
```

### Complete Example with CommandResult

```go
func (m *MyManager) Install(ctx context.Context, packages []string, opts *manager.Options) ([]manager.PackageInfo, error) {
    // Validation - plugin developer knows this is usage error
    if len(packages) == 0 {
        return nil, manager.WrapReturn(manager.StatusUsageError, "install requires package names", nil)
    }

    // Execute command
    result, err := m.GetRunner().Run(ctx, "my-tool", []string{"install"}, packages...)
    if err != nil {
        // Command execution failed (tool not found)
        return nil, manager.WrapReturn(manager.StatusUnavailableError, "my-tool command failed", err)
    }

    // Handle error cases first (return-early pattern)
    if result.ExitCode != 0 {
        switch result.ExitCode {
        case 1:
            // Generic error - check stderr for specifics
            if strings.Contains(string(result.Stderr), "not found") {
                return nil, manager.WrapReturn(manager.StatusUnavailableError, "package not found", nil)
            }
            return nil, manager.WrapReturn(manager.StatusGeneralError, "installation failed", nil)
        case 2:
            // Invalid usage
            return nil, manager.WrapReturn(manager.StatusUsageError, "invalid arguments", nil)
        case 77:
            // Permission denied (if your tool uses this exit code)
            return nil, manager.WrapReturn(manager.StatusPermissionError, "requires root access", nil)
        default:
            return nil, manager.WrapReturn(manager.StatusGeneralError, "unknown error", nil)
        }
    }

    // Success - parse results and return
    return m.parseInstallOutput(string(result.Output)), nil
}
```

### Why This Design is Better

**Plugin developers know exactly what happened:**
```go
// You KNOW this is a usage error
if len(packages) == 0 {
    return nil, manager.WrapReturn(manager.StatusUsageError, "install requires packages", nil)
}

// You KNOW what your tool's exit codes mean
result, err := m.GetRunner().Run(ctx, "my-tool", args)
if err != nil {
    return nil, manager.WrapReturn(manager.StatusUnavailableError, "my-tool not found", err)
}

// Handle error cases first (return-early pattern)
if result.ExitCode != 0 {
    switch result.ExitCode {
    case 100:
        // APT uses exit code 100 for "package not found"
        return nil, manager.WrapReturn(manager.StatusUnavailableError, "package not found", nil)
    case 77:
        // Your tool uses exit code 77 for permission denied
        return nil, manager.WrapReturn(manager.StatusPermissionError, "requires root", nil)
    default:
        // Check stderr for more context
        if strings.Contains(string(result.Stderr), "invalid") {
            return nil, manager.WrapReturn(manager.StatusUsageError, "invalid input", nil)
        }
        return nil, manager.WrapReturn(manager.StatusGeneralError, "command failed", nil)
    }
}

// Success case - do the main work
return processResults(result.Output), nil
```

**No guessing, no pattern matching, just explicit control.**

### Error Message Best Practices

Create clear, actionable error messages:

```go
// ✅ Good: Clear and actionable
return nil, manager.WrapReturn(manager.StatusPermissionError, "installation requires root access - try with sudo", err)
return nil, manager.WrapReturn(manager.StatusUnavailableError, "package 'vim' not found in enabled repositories", nil)
return nil, manager.WrapReturn(manager.StatusUsageError, "package name cannot contain spaces or special characters", nil)

// ❌ Bad: Vague and unhelpful
return nil, manager.WrapReturn(manager.StatusGeneralError, "error", err)
return nil, manager.WrapReturn(manager.StatusUsageError, "bad input", nil)
```

### Migration Guide

**Update existing error handling:**

```go
// OLD: Generic error handling
if err != nil {
    return nil, fmt.Errorf("install failed: %w", err)
}

// NEW: Explicit status-based handling
if err != nil {
    return nil, manager.WrapReturn(manager.StatusGeneralError, "install failed", err)
}

// BETTER: Specific status based on what you know happened
if err != nil {
    if strings.Contains(err.Error(), "permission denied") {
        return nil, manager.WrapReturn(manager.StatusPermissionError, "installation requires root access", err)
    }
    return nil, manager.WrapReturn(manager.StatusGeneralError, "install failed", err)
}
```

This ensures consistent exit codes across all package managers and provides better user experience.

## Testing Your Plugin

```go
func TestMyManager(t *testing.T) {
    // Use mock command runner for testing
    mockRunner := manager.NewMockCommandRunner()
    manager := NewMyManagerWithRunner(mockRunner)

    // Set up mock responses
    mockRunner.AddCommand("my-tool", []string{"search", "test"},
        []byte("test-package 1.0.0"), nil)

    // Test search
    ctx := context.Background()
    results, err := manager.Search(ctx, []string{"test"}, nil)

    assert.NoError(t, err)
    assert.Len(t, results, 1)
    assert.Equal(t, "test-package", results[0].Name)
}
```

## Advanced Examples

### Version Manager (like nvm)

```go
type NVMManager struct {
    *manager.BaseManager
}

func (m *NVMManager) List(ctx context.Context, filter manager.ListFilter, opts *manager.Options) ([]manager.PackageInfo, error) {
    switch filter {
    case manager.FilterInstalled:
        return m.listInstalled(ctx, opts)
    case manager.FilterAvailable:
        return m.listAvailable(ctx, opts)
    default:
        return nil, fmt.Errorf("filter %s not supported", filter)
    }
}

// Custom operation for version managers
func (m *NVMManager) SetActiveVersion(version string, opts *manager.Options) error {
    _, err := m.GetRunner().Run(context.Background(), "nvm", []string{"use", version})
    return err
}
```

### Game Manager (like Steam)

```go
type SteamManager struct {
    *manager.BaseManager
}

func (m *SteamManager) Install(ctx context.Context, packages []string, opts *manager.Options) ([]manager.PackageInfo, error) {
    // Steam uses app IDs instead of package names
    for _, appID := range packages {
        _, err := m.GetRunner().Run(ctx, "steamcmd", []string{"+app_update", appID, "+quit"})
        if err != nil {
            return nil, err
        }
    }

    // Return results with Steam-specific metadata
    var results []manager.PackageInfo
    for _, appID := range packages {
        pkg := manager.NewPackageInfo(appID, "unknown", "installed", m.GetType())
        pkg.Metadata["appid"] = appID
        pkg.Metadata["platform"] = "steam"
        results = append(results, pkg)
    }

    return results, nil
}

// Custom Steam operations
func (m *SteamManager) VerifyGameIntegrity(appID string) error {
    _, err := m.GetRunner().Run(context.Background(), "steamcmd", []string{"+app_update", appID, "validate", "+quit"})
    return err
}
```

## Best Practices

### 1. Follow the "Less is More" Principle

- Only implement operations your tool actually supports
- Use `BaseManager` for everything else
- Don't try to fake unsupported operations

```go
// ❌ WRONG: Don't fake unsupported operations
func (m *Manager) AutoRemove(ctx context.Context, opts *Options) ([]PackageInfo, error) {
    return []PackageInfo{}, nil // WRONG: fake implementation
}

// ✅ CORRECT: Use BaseManager for unsupported operations
// (BaseManager.AutoRemove will return ErrOperationNotSupported)
```

### 2. Consistent Naming

```go
// Good
manager.NewBaseManager("npm", manager.CategoryLanguage, runner)
manager.NewBaseManager("steam", manager.CategoryGame, runner)
manager.NewBaseManager("apt", manager.CategorySystem, runner)

// Bad
manager.NewBaseManager("Node Package Manager", "nodejs", runner) // BAD: verbose name
```

### 3. Proper Error Messages

```go
// Good
return nil, fmt.Errorf("failed to install %s: package not found in registry", pkg)

// Bad
return nil, errors.New("error") // BAD: vague error
```

### 4. Use Metadata Wisely

```go
// Store tool-specific data in metadata
pkg.Metadata["repository_url"] = "https://registry.npmjs.org"
pkg.Metadata["license"] = "MIT"
pkg.Metadata["download_count"] = 1000000

// Don't abuse it for core data
pkg.Name = "package-name"        // Good
pkg.Metadata["name"] = "package-name"  // BAD: wrong field
```

### 5. Handle Edge Cases

```go
func (m *MyManager) Search(ctx context.Context, query []string, opts *manager.Options) ([]manager.PackageInfo, error) {
    // Handle empty query
    if len(query) == 0 {
        return []manager.PackageInfo{}, nil
    }

    // Handle network issues
    args := []string{"search", strings.Join(query, " ")}
    output, err := m.GetRunner().Run(ctx, "my-tool", args)
    if err != nil {
        if strings.Contains(err.Error(), "network") {
            return nil, fmt.Errorf("network error: %w", err)
        }
        return nil, fmt.Errorf("search failed: %w", err)
    }

    // Handle no results
    if strings.TrimSpace(string(output)) == "" {
        return []manager.PackageInfo{}, nil
    }

    return m.parseOutput(string(output))
}
```

## Complete Example: pip Manager

```go
package pip

import (
    "context"
    "encoding/json"
    "fmt"
    "strings"

    "github.com/bluet/syspkg/manager"
)

type PipManager struct {
    *manager.BaseManager
}

func NewPipManager() *PipManager {
    base := manager.NewBaseManager("pip", manager.CategoryLanguage, manager.NewDefaultCommandRunner())
    return &PipManager{BaseManager: base}
}

func (m *PipManager) IsAvailable() bool {
    _, err := m.GetRunner().Run(context.Background(), "pip", []string{"--version"})
    return err == nil
}

func (m *PipManager) Search(ctx context.Context, query []string, opts *manager.Options) ([]manager.PackageInfo, error) {
    if opts == nil {
        opts = manager.DefaultOptions()
    }

    if len(query) == 0 {
        return []manager.PackageInfo{}, nil
    }

    m.HandleDryRun(opts, "search", query)
    if opts.DryRun {
        return []manager.PackageInfo{}, nil
    }

    // Use context directly - caller controls timeouts

    // Use pip search (note: deprecated in newer pip versions)
    searchTerm := strings.Join(query, " ")
    output, err := m.GetRunner().Run(ctx, "pip", []string{"search", searchTerm})
    if err != nil {
        return nil, fmt.Errorf("pip search failed: %w", err)
    }

    return m.parseSearchOutput(string(output))
}

func (m *PipManager) Install(ctx context.Context, packages []string, opts *manager.Options) ([]manager.PackageInfo, error) {
    if opts == nil {
        opts = manager.DefaultOptions()
    }

    if err := m.ValidatePackageNames(packages); err != nil {
        return nil, err
    }

    m.HandleDryRun(opts, "install", packages)
    if opts.DryRun {
        results := make([]manager.PackageInfo, len(packages))
        for i, pkg := range packages {
            results[i] = manager.NewPackageInfo(pkg, "unknown", "would-install", m.GetType())
        }
        return results, nil
    }

    // Use context directly - caller controls timeouts

    args := []string{"install"}
    if opts.GlobalScope {
        args = append(args, "--user")
    }
    args = append(args, packages...)

    output, err := m.GetRunner().Run(ctx, "pip", args)
    if err != nil {
        return nil, fmt.Errorf("pip install failed: %w", err)
    }

    m.LogVerbose(opts, "Install output: %s", string(output))

    // Assume successful installation
    results := make([]manager.PackageInfo, len(packages))
    for i, pkg := range packages {
        results[i] = manager.NewPackageInfo(pkg, "unknown", "installed", m.GetType())
    }

    return results, nil
}

func (m *PipManager) List(ctx context.Context, filter manager.ListFilter, opts *manager.Options) ([]manager.PackageInfo, error) {
    if filter != manager.FilterInstalled {
        return nil, fmt.Errorf("pip only supports listing installed packages")
    }

    // Use context directly - caller controls timeouts

    output, err := m.GetRunner().Run(ctx, "pip", []string{"list", "--format=json"})
    if err != nil {
        return nil, fmt.Errorf("pip list failed: %w", err)
    }

    return m.parseListOutput(string(output))
}

// Parsing helpers
func (m *PipManager) parseSearchOutput(output string) ([]manager.PackageInfo, error) {
    var packages []manager.PackageInfo
    lines := strings.Split(output, "\n")

    for _, line := range lines {
        if strings.TrimSpace(line) == "" {
            continue
        }

        // Simple parsing - real implementation would be more robust
        parts := strings.Fields(line)
        if len(parts) >= 2 {
            pkg := manager.NewPackageInfo(parts[0], parts[1], "available", m.GetType())
            if len(parts) > 2 {
                pkg.Description = strings.Join(parts[2:], " ")
            }
            packages = append(packages, pkg)
        }
    }

    return packages, nil
}

func (m *PipManager) parseListOutput(output string) ([]manager.PackageInfo, error) {
    var packages []manager.PackageInfo

    // pip list --format=json returns array of objects
    var pipPackages []map[string]string
    if err := json.Unmarshal([]byte(output), &pipPackages); err != nil {
        return nil, fmt.Errorf("failed to parse pip list output: %w", err)
    }

    for _, pipPkg := range pipPackages {
        if name, ok := pipPkg["name"]; ok {
            pkg := manager.NewPackageInfo(name, pipPkg["version"], "installed", m.GetType())
            packages = append(packages, pkg)
        }
    }

    return packages, nil
}

// Plugin registration
type Plugin struct{}

func (p *Plugin) CreateManager() manager.PackageManager {
    return NewPipManager()
}

func (p *Plugin) GetPriority() int {
    return 80 // High priority for Python environments
}

func init() {
    if err := manager.Register("pip", &Plugin{}); err != nil {
        panic("Failed to register pip plugin: " + err.Error())
    }
}
```

## Conclusion

The unified interface makes adding new package managers incredibly straightforward:

1. **Embed `BaseManager`** for 90% of functionality
2. **Override only what you need** - search, install, etc.
3. **Register your plugin** with auto-initialization
4. **Done!** Your manager works with the entire syspkg ecosystem

The architecture follows the "Less is more" principle by providing a minimal but powerful interface that doesn't force unnecessary complexity on plugin developers while ensuring a consistent user experience across all package managers.
