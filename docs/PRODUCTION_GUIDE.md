# Production-Quality Plugin Development Guide

## Overview

This guide covers the advanced techniques, patterns, and domain knowledge needed to create production-ready package manager plugins that are robust, secure, and maintainable.

## Table of Contents

1. [Real-World Parsing Strategies](#real-world-parsing-strategies)
2. [Error Handling Patterns](#error-handling-patterns)
3. [Testing Strategies](#testing-strategies)
4. [Security Considerations](#security-considerations)
5. [Performance Optimization](#performance-optimization)
6. [Package Manager Domain Knowledge](#package-manager-domain-knowledge)
7. [Production Readiness Checklist](#production-readiness-checklist)

## Real-World Parsing Strategies

### Understanding Command Output Complexity

Package managers produce complex, inconsistent output that requires sophisticated parsing:

```go
// ❌ NAIVE APPROACH - Will break in real use
func (m *Manager) parseInstallOutput(output string) []PackageInfo {
    lines := strings.Split(output, "\n")
    for _, line := range lines {
        if strings.Contains(line, "installed") { // WRONG: naive string matching
            // This is too simplistic and will match irrelevant lines
        }
    }
}

// ✅ PRODUCTION APPROACH - Robust regex with validation
func (m *Manager) parseInstallOutput(output string) []PackageInfo {
    var packages []PackageInfo

    // Normalize input - handle different line endings
    output = strings.ReplaceAll(output, "\r\n", "\n")
    output = strings.TrimSpace(output)

    // Use specific regex patterns for exact matching
    // APT example: "Setting up package:arch (version) ..."
    settingUpRegex := regexp.MustCompile(`Setting up ([^:\s]+)(?::([^:\s]+))?\s+\(([^)]+)\)`)

    lines := strings.Split(output, "\n")
    for _, rawLine := range lines {
        line := strings.TrimSpace(rawLine)

        // Skip empty lines and progress indicators
        if line == "" || strings.HasPrefix(line, "Reading") ||
           strings.HasPrefix(line, "Building") {
            continue
        }

        if match := settingUpRegex.FindStringSubmatch(line); match != nil {
            name := match[1]
            arch := match[2] // May be empty
            version := match[3]

            // Validate extracted data
            if name == "" || version == "" {
                continue
            }

            pkg := manager.NewPackageInfo(name, version, manager.StatusInstalled, "apt")

            if arch != "" {
                pkg.Metadata["arch"] = arch
            }

            packages = append(packages, pkg)
        }
    }

    return packages
}
```

### Command Selection Strategies

Understanding when to use which commands is crucial:

```go
// APT Command Selection Guide
func (m *APTManager) chooseListCommand(filter ListFilter) (string, []string, error) {
    switch filter {
    case FilterInstalled:
        // Use dpkg-query for speed and reliability
        // APT's "apt list --installed" is slower and less reliable
        return "dpkg-query", []string{
            "-W", "-f", "${binary:Package} ${Version} ${Architecture}\n",
        }, nil

    case FilterUpgradable:
        // Only apt can determine upgradable packages
        return "apt", []string{"list", "--upgradable"}, nil

    case FilterAvailable:
        // This would be too expensive - guide users to Search instead
        return "", nil, fmt.Errorf("listing all available packages not supported - use Search instead")

    default:
        return "", nil, fmt.Errorf("unsupported filter: %s", filter)
    }
}

// Environment Variables - Critical for Reliability
func (m *APTManager) getEnvironment() []string {
    return []string{
        "DEBIAN_FRONTEND=noninteractive", // Prevent interactive prompts
        "DEBCONF_NONINTERACTIVE_SEEN=true", // Skip debconf questions
        "LC_ALL=C", // Ensure consistent output format
        "TERM=dumb", // Prevent terminal escape sequences
    }
}
```

## Error Handling Patterns

### Sophisticated Error Classification

```go
import (
    "errors"
    "fmt"
    "os/exec"
    "syscall"
)

// Define specific error types for better handling
type PackageManagerError struct {
    Operation string
    Package   string
    ExitCode  int
    Stderr    string
    Stdout    string
    Cause     error
}

func (e *PackageManagerError) Error() string {
    return fmt.Sprintf("%s failed for package '%s': %s", e.Operation, e.Package, e.Cause)
}

func (e *PackageManagerError) Unwrap() error {
    return e.Cause
}

// Enhanced error handling with context
func (m *Manager) executeWithErrorHandling(ctx context.Context, operation string, cmd string, args []string, packages []string) ([]byte, error) {
    output, err := m.GetRunner().Run(ctx, cmd, args, m.getEnvironment()...)

    if err != nil {
        // Extract exit code and classify error
        var exitCode int
        if exitError, ok := err.(*exec.ExitError); ok {
            if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
                exitCode = status.ExitStatus()
            }
        }

        // Classify errors by exit code (package manager specific)
        switch exitCode {
        case 0:
            // Success - should not reach here
            return output, nil

        case 1:
            // APT: Package not found or general error
            if strings.Contains(string(output), "Unable to locate package") {
                return nil, &PackageManagerError{
                    Operation: operation,
                    Package:   strings.Join(packages, ", "),
                    ExitCode:  exitCode,
                    Stdout:    string(output),
                    Cause:     ErrPackageNotFound,
                }
            }

        case 100:
            // APT: Package operation failed (dependencies, conflicts, etc.)
            return nil, &PackageManagerError{
                Operation: operation,
                Package:   strings.Join(packages, ", "),
                ExitCode:  exitCode,
                Stdout:    string(output),
                Cause:     fmt.Errorf("package operation failed - check dependencies"),
            }

        default:
            // Unknown error
            return nil, &PackageManagerError{
                Operation: operation,
                Package:   strings.Join(packages, ", "),
                ExitCode:  exitCode,
                Stdout:    string(output),
                Cause:     err,
            }
        }
    }

    return output, nil
}

### Context and Timeout Best Practices

**Plugin implementations should respect caller-provided context directly:**

```go
// ✅ CORRECT: Use context directly - let caller control timeouts
func (m *Manager) Install(ctx context.Context, packages []string, opts *Options) ([]PackageInfo, error) {
    // Validate inputs
    if err := m.ValidatePackageNames(packages); err != nil {
        return nil, err
    }

    // Use context directly - caller controls timeouts
    args := append([]string{"install"}, packages...)
    output, err := m.GetRunner().Run(ctx, "apt", args, "DEBIAN_FRONTEND=noninteractive")
    if err != nil {
        return nil, fmt.Errorf("apt install failed: %w", err)
    }

    return parseInstallOutput(string(output)), nil
}

// ❌ WRONG: Don't add arbitrary timeout defaults
func (m *Manager) installWithDefaults(ctx context.Context, packages []string) error {
    // Don't do this - plugin authors can't know deployment environment
    timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Minute) // WRONG: arbitrary timeout
    defer cancel()
    // ...
}
```

**Why plugins shouldn't set timeout defaults:**
- **Environment variability**: Bare metal vs VM performance differs dramatically
- **Package managers handle timeouts**: APT, YUM already have configurable timeouts
- **System administrators tune these**: Based on their specific environment
- **Plugin authors can't know context**: Local mirrors vs slow external repos

## Testing Strategies

### Comprehensive Mock Testing Framework

```go
// Advanced MockCommandRunner with realistic behaviors
type AdvancedMockCommandRunner struct {
    commands     map[string]string // command -> output
    errors       map[string]error  // command -> error
    delays       map[string]time.Duration // command -> delay
    callHistory  []MockCall
    environment  []string
    interactive  bool
}

type MockCall struct {
    Command     string
    Args        []string
    Environment []string
    Timestamp   time.Time
    Interactive bool
}

func NewAdvancedMockRunner() *AdvancedMockCommandRunner {
    return &AdvancedMockCommandRunner{
        commands:    make(map[string]string),
        errors:      make(map[string]error),
        delays:      make(map[string]time.Duration),
        callHistory: make([]MockCall, 0),
    }
}

// Set realistic command outputs based on real package manager behavior
func (m *AdvancedMockCommandRunner) SetAPTSearchOutput(query string, packages []APTPackage) {
    var output strings.Builder
    output.WriteString("Sorting...\nFull Text Search...\n")

    for _, pkg := range packages {
        output.WriteString(fmt.Sprintf("%s/%s %s %s\n",
            pkg.Name, pkg.Suite, pkg.Version, pkg.Arch))
        if pkg.Description != "" {
            output.WriteString(fmt.Sprintf("  %s\n", pkg.Description))
        }
        output.WriteString("\n")
    }

    m.commands[fmt.Sprintf("apt search %s", query)] = output.String()
}

// Test data structures for realistic scenarios
type APTPackage struct {
    Name        string
    Version     string
    Arch        string
    Suite       string
    Description string
    Status      string
}

// Fixture data for comprehensive testing
func GetTestFixtures() map[string][]APTPackage {
    return map[string][]APTPackage{
        "vim": {
            {
                Name:        "vim",
                Version:     "2:8.2.3458-2ubuntu2.5",
                Arch:        "amd64",
                Suite:       "jammy",
                Description: "Vi IMproved - enhanced vi editor",
                Status:      "available",
            },
            {
                Name:        "vim-common",
                Version:     "2:8.2.3458-2ubuntu2.5",
                Arch:        "all",
                Suite:       "jammy",
                Description: "Vi IMproved - Common files",
                Status:      "available",
            },
        },
        "curl": {
            {
                Name:        "curl",
                Version:     "7.81.0-1ubuntu1.19",
                Arch:        "amd64",
                Suite:       "jammy-updates",
                Description: "command line tool for transferring data with URL syntax",
                Status:      "installed",
            },
        },
    }
}

// Comprehensive test scenarios
func TestAPTManager_Search_ComprehensiveScenarios(t *testing.T) {
    tests := []struct {
        name          string
        query         []string
        mockOutput    string
        mockError     error
        expectedPkgs  int
        expectedError bool
        validate      func([]PackageInfo) error
    }{
        {
            name:  "successful search with multiple results",
            query: []string{"vim"},
            mockOutput: `Sorting...
Full Text Search...
vim/jammy 2:8.2.3458-2ubuntu2.5 amd64
  Vi IMproved - enhanced vi editor

vim-common/jammy,jammy 2:8.2.3458-2ubuntu2.5 all
  Vi IMproved - Common files`,
            expectedPkgs: 2,
            validate: func(pkgs []PackageInfo) error {
                if pkgs[0].Name != "vim" {
                    return fmt.Errorf("expected first package name 'vim', got '%s'", pkgs[0].Name)
                }
                if pkgs[0].NewVersion != "2:8.2.3458-2ubuntu2.5" {
                    return fmt.Errorf("expected version '2:8.2.3458-2ubuntu2.5', got '%s'", pkgs[0].NewVersion)
                }
                return nil
            },
        },
        {
            name:          "empty search results",
            query:         []string{"nonexistent-package-xyz"},
            mockOutput:    "Sorting...\nFull Text Search...\n",
            expectedPkgs:  0,
            expectedError: false,
        },
        {
            name:          "malformed output handling",
            query:         []string{"test"},
            mockOutput:    "ERROR: Invalid package name\nSorting...\ngarbage data\n",
            expectedPkgs:  0,
            expectedError: false, // Should handle gracefully
        },
        {
            name:          "command failure",
            query:         []string{"test"},
            mockError:     &exec.ExitError{},
            expectedError: true,
        },
        {
            name:          "invalid package name input",
            query:         []string{"invalid;package"},
            expectedError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            runner := NewAdvancedMockRunner()
            if tt.mockOutput != "" {
                runner.SetOutput(fmt.Sprintf("apt search %s", strings.Join(tt.query, " ")), tt.mockOutput)
            }
            if tt.mockError != nil {
                runner.SetError(fmt.Sprintf("apt search %s", strings.Join(tt.query, " ")), tt.mockError)
            }

            mgr := NewManagerWithRunner(runner)
            ctx := context.Background()

            packages, err := mgr.Search(ctx, tt.query, nil)

            if tt.expectedError && err == nil {
                t.Errorf("expected error but got none")
            }
            if !tt.expectedError && err != nil {
                t.Errorf("unexpected error: %v", err)
            }

            if len(packages) != tt.expectedPkgs {
                t.Errorf("expected %d packages, got %d", tt.expectedPkgs, len(packages))
            }

            if tt.validate != nil && len(packages) > 0 {
                if err := tt.validate(packages); err != nil {
                    t.Errorf("validation failed: %v", err)
                }
            }
        })
    }
}

// Integration tests with real commands (when available)
func TestAPTManager_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration tests in short mode")
    }

    // Only run if APT is actually available
    if _, err := exec.LookPath("apt"); err != nil {
        t.Skip("APT not available, skipping integration tests")
    }

    mgr := NewManager()
    if !mgr.IsAvailable() {
        t.Skip("APT not available on this system")
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    t.Run("real search operation", func(t *testing.T) {
        packages, err := mgr.Search(ctx, []string{"curl"}, nil)
        if err != nil {
            t.Fatalf("search failed: %v", err)
        }

        if len(packages) == 0 {
            t.Error("expected at least one package for 'curl'")
        }

        // Validate structure of real results
        for _, pkg := range packages[:min(3, len(packages))] {
            if pkg.Name == "" {
                t.Error("package name should not be empty")
            }
            if pkg.Manager != "apt" {
                t.Errorf("expected manager '%s', got '%s'", "apt", pkg.Manager)
            }
        }
    })
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
```

## Security Considerations

### Advanced Input Validation

```go
import (
    "regexp"
    "unicode"
    "unicode/utf8"
)

// Enhanced package name validation with detailed security checks
var (
    // Allow package names, versions, and architecture specifications
    packageNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9+._-]*(?::[a-zA-Z0-9][a-zA-Z0-9+._-]*)?(?:=[a-zA-Z0-9][a-zA-Z0-9+._:~-]*)?$`)

    // Dangerous patterns that could indicate injection attempts
    dangerousPatterns = []*regexp.Regexp{
        regexp.MustCompile(`[;&|$\x60\\<>(){}[\]*?~]`), // Shell metacharacters
        regexp.MustCompile(`\.\./`),                      // Path traversal
        regexp.MustCompile(`^-`),                         // Option injection
        regexp.MustCompile(`\x00`),                       // Null bytes
        regexp.MustCompile(`\r|\n`),                      // Line breaks
    }
)

func ValidatePackageNameAdvanced(name string) error {
    // Basic checks
    if name == "" {
        return fmt.Errorf("package name cannot be empty")
    }

    if len(name) > 255 {
        return fmt.Errorf("package name too long (max 255 characters)")
    }

    // UTF-8 validation
    if !utf8.ValidString(name) {
        return fmt.Errorf("package name contains invalid UTF-8")
    }

    // Check for control characters
    for _, r := range name {
        if unicode.IsControl(r) && r != '\t' {
            return fmt.Errorf("package name contains control character: %U", r)
        }
    }

    // Check dangerous patterns
    for _, pattern := range dangerousPatterns {
        if pattern.MatchString(name) {
            return fmt.Errorf("package name contains dangerous characters")
        }
    }

    // Package manager specific validation
    if !packageNameRegex.MatchString(name) {
        return fmt.Errorf("invalid package name format")
    }

    return nil
}

// Command injection prevention
func (m *Manager) sanitizeEnvironment(env []string) []string {
    safe := make([]string, 0, len(env))

    for _, envVar := range env {
        // Validate environment variables
        if strings.Contains(envVar, ";") || strings.Contains(envVar, "|") {
            continue // Skip potentially dangerous env vars
        }

        // Ensure format is KEY=VALUE
        if !strings.Contains(envVar, "=") {
            continue
        }

        parts := strings.SplitN(envVar, "=", 2)
        if len(parts) != 2 {
            continue
        }

        key, value := parts[0], parts[1]

        // Validate key
        if !regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`).MatchString(key) {
            continue
        }

        // Basic value sanitization
        if strings.ContainsAny(value, ";&|$`\\") {
            continue
        }

        safe = append(safe, envVar)
    }

    return safe
}

// Secure command execution wrapper
func (m *Manager) executeSecurely(ctx context.Context, command string, args []string, env []string) ([]byte, error) {
    // Validate command
    if filepath.Base(command) != command {
        return nil, fmt.Errorf("command must not contain path separators")
    }

    // Validate arguments
    for _, arg := range args {
        if strings.ContainsAny(arg, ";&|$`\\<>") {
            return nil, fmt.Errorf("argument contains dangerous characters: %s", arg)
        }
    }

    // Sanitize environment
    safeEnv := m.sanitizeEnvironment(env)

    return m.GetRunner().Run(ctx, command, args, safeEnv...)
}
```

## Performance Optimization

### Efficient Batch Operations

```go
// Batch processing for multiple package operations
func (m *Manager) InstallBatch(ctx context.Context, packages []string, opts *Options) ([]PackageInfo, error) {
    if len(packages) == 0 {
        return []PackageInfo{}, nil
    }

    // Validate all packages first
    if err := ValidatePackageNames(packages); err != nil {
        return nil, err
    }

    // Determine optimal batch size based on package manager
    batchSize := m.getOptimalBatchSize()

    var allResults []PackageInfo
    var errors []error

    // Process in batches
    for i := 0; i < len(packages); i += batchSize {
        end := i + batchSize
        if end > len(packages) {
            end = len(packages)
        }

        batch := packages[i:end]
        results, err := m.installSingleBatch(ctx, batch, opts)

        if err != nil {
            errors = append(errors, err)
            continue
        }

        allResults = append(allResults, results...)
    }

    // Handle partial failures
    if len(errors) > 0 {
        if len(allResults) == 0 {
            return nil, fmt.Errorf("all installations failed: %v", errors[0])
        }

        // Return partial results with warning
        return allResults, fmt.Errorf("some installations failed (%d errors), %d succeeded",
            len(errors), len(allResults))
    }

    return allResults, nil
}

func (m *Manager) getOptimalBatchSize() int {
    // APT can handle many packages in one command efficiently
    // Other managers might need smaller batches
    return 50
}

// Concurrent operations for independent tasks
func (m *Manager) GetInfoBatch(ctx context.Context, packages []string, opts *Options) (map[string]PackageInfo, error) {
    if len(packages) == 0 {
        return make(map[string]PackageInfo), nil
    }

    results := make(map[string]PackageInfo)
    var mu sync.Mutex
    var wg sync.WaitGroup
    errChan := make(chan error, len(packages))

    // Limit concurrency to avoid overwhelming the system
    semaphore := make(chan struct{}, 5)

    for _, pkg := range packages {
        wg.Add(1)
        go func(packageName string) {
            defer wg.Done()

            // Acquire semaphore
            semaphore <- struct{}{}
            defer func() { <-semaphore }()

            info, err := m.GetInfo(ctx, packageName, opts)
            if err != nil {
                errChan <- fmt.Errorf("failed to get info for %s: %w", packageName, err)
                return
            }

            mu.Lock()
            results[packageName] = info
            mu.Unlock()
        }(pkg)
    }

    wg.Wait()
    close(errChan)

    // Collect errors
    var errors []error
    for err := range errChan {
        errors = append(errors, err)
    }

    if len(errors) > 0 {
        return results, fmt.Errorf("encountered %d errors: %v", len(errors), errors[0])
    }

    return results, nil
}

// Caching for expensive operations
type CacheEntry struct {
    Data      interface{}
    Timestamp time.Time
    TTL       time.Duration
}

type ManagerCache struct {
    data map[string]CacheEntry
    mu   sync.RWMutex
}

func NewManagerCache() *ManagerCache {
    return &ManagerCache{
        data: make(map[string]CacheEntry),
    }
}

func (c *ManagerCache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    entry, exists := c.data[key]
    if !exists {
        return nil, false
    }

    // Check if expired
    if time.Since(entry.Timestamp) > entry.TTL {
        return nil, false
    }

    return entry.Data, true
}

func (c *ManagerCache) Set(key string, data interface{}, ttl time.Duration) {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.data[key] = CacheEntry{
        Data:      data,
        Timestamp: time.Now(),
        TTL:       ttl,
    }
}

// Use caching for expensive operations like search
func (m *Manager) SearchWithCache(ctx context.Context, query []string, opts *Options) ([]PackageInfo, error) {
    cacheKey := fmt.Sprintf("search:%s", strings.Join(query, "+"))

    if cached, found := m.cache.Get(cacheKey); found {
        return cached.([]PackageInfo), nil
    }

    results, err := m.Search(ctx, query, opts)
    if err != nil {
        return nil, err
    }

    // Cache for 5 minutes
    m.cache.Set(cacheKey, results, 5*time.Minute)

    return results, nil
}
```

## Package Manager Domain Knowledge

### APT-Specific Implementation Details

```go
// APT Command Reference and Best Practices
const (
    // Why we use these specific commands:

    // dpkg-query vs apt list --installed:
    // - dpkg-query is faster (direct database access)
    // - apt list includes more formatting overhead
    // - dpkg-query has more stable output format
    listInstalledCmd = "dpkg-query"

    // apt-cache vs apt show:
    // - apt-cache show is faster and more reliable
    // - apt show adds extra formatting that's harder to parse
    // - apt-cache has been stable for decades
    packageInfoCmd = "apt-cache"

    // Environment variables critical for reliability:
    // DEBIAN_FRONTEND=noninteractive - prevents interactive prompts
    // DEBCONF_NONINTERACTIVE_SEEN=true - skips debconf questions
    // LC_ALL=C - ensures consistent output format
    // TERM=dumb - prevents terminal escape sequences
)

// APT Exit Code Reference
func (m *APTManager) interpretExitCode(exitCode int, operation string) error {
    switch exitCode {
    case 0:
        return nil // Success

    case 1:
        return fmt.Errorf("general error or package not found")

    case 2:
        return fmt.Errorf("command line syntax error")

    case 100:
        return fmt.Errorf("package operation failed - likely dependency issues")

    case 101:
        return fmt.Errorf("package manager locked - another operation in progress")

    default:
        return fmt.Errorf("unknown error (exit code %d)", exitCode)
    }
}

// APT Package States and Status Mapping
func (m *APTManager) normalizeAPTStatus(dpkgStatus string) string {
    switch dpkgStatus {
    case "install ok installed":
        return StatusInstalled

    case "install ok config-files":
        // Package removed but config files remain
        // Normalize to "available" for cross-manager compatibility
        return StatusAvailable

    case "deinstall ok config-files":
        return StatusAvailable

    case "install ok unpacked":
        return "unpacked" // Installation in progress

    case "install ok half-configured":
        return "broken" // Installation failed

    default:
        return StatusUnknown
    }
}

// APT-specific parsing patterns
var aptPatterns = struct {
    // "Setting up package:arch (version) ..."
    InstallSuccess *regexp.Regexp

    // "Removing package:arch (version) ..."
    RemoveSuccess *regexp.Regexp

    // "package/suite version arch [upgradable from: old_version]"
    UpgradableFormat *regexp.Regexp

    // "E: Unable to locate package xyz"
    PackageNotFound *regexp.Regexp
}{
    InstallSuccess:   regexp.MustCompile(`Setting up ([^:\s]+)(?::([^:\s]+))?\s+\(([^)]+)\)`),
    RemoveSuccess:    regexp.MustCompile(`Removing ([^:\s]+)(?::([^:\s]+))?\s+\(([^)]+)\)`),
    UpgradableFormat: regexp.MustCompile(`^([^/]+)/([^\s]+)\s+([^\s]+)\s+([^\s]+)\s+\[upgradable from:\s+([^\]]+)\]`),
    PackageNotFound:  regexp.MustCompile(`E: Unable to locate package (.+)`),
}
```

### Package Manager Comparison Guide

```go
// Cross-Package Manager Implementation Patterns
type ManagerProfile struct {
    Name                string
    SearchCommand       []string
    InstallCommand      []string
    RemoveCommand       []string
    ListInstalledCmd    []string
    RequiresElevation   bool
    SupportsParallel    bool
    OutputFormat        string
    CommonEnvironment   []string
}

var managerProfiles = map[string]ManagerProfile{
    "apt": {
        Name:              "apt",
        SearchCommand:     []string{"search"},
        InstallCommand:    []string{"install", "-y"},
        RemoveCommand:     []string{"remove", "-y", "--autoremove"},
        ListInstalledCmd:  []string{"dpkg-query", "-W", "-f", "${binary:Package} ${Version}\n"},
        RequiresElevation: true,
        SupportsParallel:  false, // APT locks prevent parallel operations
        OutputFormat:      "structured",
        CommonEnvironment: []string{"DEBIAN_FRONTEND=noninteractive", "LC_ALL=C"},
    },

    "npm": {
        Name:              "npm",
        SearchCommand:     []string{"search", "--json"},
        InstallCommand:    []string{"install"},
        RemoveCommand:     []string{"uninstall"},
        ListInstalledCmd:  []string{"list", "--json"},
        RequiresElevation: false, // User-space installation
        SupportsParallel:  true,  // npm can handle concurrent operations
        OutputFormat:      "json",
        CommonEnvironment: []string{"NPM_CONFIG_PROGRESS=false"},
    },

    "pip": {
        Name:              "pip",
        SearchCommand:     []string{"search"}, // Deprecated in newer versions
        InstallCommand:    []string{"install"},
        RemoveCommand:     []string{"uninstall", "-y"},
        ListInstalledCmd:  []string{"list", "--format=json"},
        RequiresElevation: false, // Typically user-space
        SupportsParallel:  true,
        OutputFormat:      "mixed", // Some JSON, some text
        CommonEnvironment: []string{"PIP_DISABLE_PIP_VERSION_CHECK=1"},
    },
}

// Template for implementing new package managers
func (m *GenericManager) implementFromProfile(profile ManagerProfile) {
    // Use profile to generate standard implementations
    // This reduces boilerplate and ensures consistency
}
```

## Production Readiness Checklist

### Essential Production Features

```go
// Production-ready manager checklist
type ProductionFeatures struct {
    // Core functionality
    AllOperationsImplemented  bool
    ErrorHandlingRobust      bool
    InputValidationComplete  bool
    SecurityAudited          bool

    // Performance
    TimeoutSupport           bool
    BatchOperationSupport    bool
    CachingImplemented       bool
    ResourceLeakPrevention   bool

    // Reliability
    ComprehensiveTests       bool
    IntegrationTests         bool
    ErrorRecovery           bool
    PartialFailureHandling   bool

    // Observability
    StructuredLogging       bool
    OperationMetrics        bool
    HealthChecks            bool
    DebugModeSupport        bool

    // User Experience
    ProgressIndicators      bool
    ClearErrorMessages      bool
    HelpText               bool
    ConfigurationSupport    bool
}

// Logging patterns for production
func (m *Manager) logOperation(operation string, packages []string, duration time.Duration, err error) {
    logLevel := "info"
    if err != nil {
        logLevel = "error"
    }

    log.WithFields(log.Fields{
        "manager":   m.GetName(),
        "operation": operation,
        "packages":  packages,
        "duration":  duration,
        "error":     err,
        "timestamp": time.Now(),
    }).Log(logLevel, fmt.Sprintf("%s operation completed", operation))
}

// Health check implementation
func (m *Manager) HealthCheck(ctx context.Context) error {
    // Check if package manager binary is available
    if !m.IsAvailable() {
        return fmt.Errorf("package manager %s is not available", m.GetName())
    }

    // Check if we can execute basic commands
    _, err := m.GetVersion()
    if err != nil {
        return fmt.Errorf("failed to get version: %w", err)
    }

    // Check if package database is accessible (quick operation)
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    _, err = m.List(ctx, FilterInstalled, &Options{Quiet: true})
    if err != nil {
        return fmt.Errorf("failed to list packages: %w", err)
    }

    return nil
}

// Configuration support
type ManagerConfig struct {
    DefaultTimeout  time.Duration `yaml:"default_timeout"`
    BatchSize      int           `yaml:"batch_size"`
    CacheEnabled   bool          `yaml:"cache_enabled"`
    CacheTTL       time.Duration `yaml:"cache_ttl"`
    LogLevel       string        `yaml:"log_level"`
    Environment    []string      `yaml:"environment"`
}

func (m *Manager) LoadConfig(configPath string) error {
    // Load and validate configuration
    // Apply to manager behavior
    return nil
}
```

### Testing Templates

```go
// Standard test template for any package manager
func TestManagerCompliance(t *testing.T, createManager func() PackageManager) {
    manager := createManager()
    ctx := context.Background()

    // Test basic interface compliance
    t.Run("interface_compliance", func(t *testing.T) {
        // Test that all methods are callable
        name := manager.GetName()
        if name == "" {
            t.Error("GetName() should return non-empty string")
        }

        managerCategory := manager.GetCategory()
        if managerCategory == "" {
            t.Error("GetCategory() should return non-empty string")
        }

        available := manager.IsAvailable()
        t.Logf("Manager available: %v", available)

        if available {
            version, err := manager.GetVersion()
            if err != nil {
                t.Errorf("GetVersion() failed: %v", err)
            }
            t.Logf("Manager version: %s", version)
        }
    })

    // Test error handling
    t.Run("error_handling", func(t *testing.T) {
        // Test invalid input handling
        _, err := manager.Search(ctx, []string{"invalid;package"}, nil)
        if err == nil {
            t.Error("Search should reject invalid package names")
        }

        _, err = manager.Install(ctx, []string{""}, nil)
        if err == nil {
            t.Error("Install should reject empty package names")
        }
    })

    // Test graceful degradation
    t.Run("graceful_degradation", func(t *testing.T) {
        // Test operations that might not be supported
        _, err := manager.Verify(ctx, []string{"test"}, nil)
        if err != nil {
            // Should return clear error message, not panic
            if !errors.Is(err, ErrOperationNotSupported) {
                t.Logf("Verify not supported: %v", err)
            }
        }
    })
}

// Performance test template
func BenchmarkManagerOperations(b *testing.B, createManager func() PackageManager) {
    manager := createManager()
    ctx := context.Background()

    if !manager.IsAvailable() {
        b.Skip("Manager not available")
    }

    b.Run("search", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _, err := manager.Search(ctx, []string{"curl"}, nil)
            if err != nil {
                b.Fatalf("Search failed: %v", err)
            }
        }
    })

    b.Run("list_installed", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _, err := manager.List(ctx, FilterInstalled, nil)
            if err != nil {
                b.Fatalf("List failed: %v", err)
            }
        }
    })
}
```

## Summary

This production guide provides the deep domain knowledge, patterns, and practices needed to create truly robust package manager plugins. Key takeaways:

1. **Parsing**: Use specific regex patterns, handle edge cases, normalize input
2. **Error Handling**: Classify errors by exit codes, provide context, handle timeouts
3. **Testing**: Comprehensive scenarios, realistic fixtures, integration tests
4. **Security**: Advanced input validation, environment sanitization, injection prevention
5. **Performance**: Batch operations, caching, concurrency with limits
6. **Domain Knowledge**: Package manager quirks, command selection, environment variables

With this guide, developers can create plugins that match production quality standards and handle real-world complexity gracefully.
