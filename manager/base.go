// Package manager provides base implementation for package managers
package manager

import (
	"context"
	"fmt"
	"log"
	"time"
)

// BaseManager provides default implementations for PackageManager interface.
// Plugin developers can embed this struct and only override methods they need to customize.
// This follows the "Less is more" principle by providing sensible defaults while allowing
// full customization when needed.
type BaseManager struct {
	name        string
	managerType string
	runner      CommandRunner
}

// NewBaseManager creates a new base manager with the given name and type
func NewBaseManager(name, managerType string, runner CommandRunner) *BaseManager {
	if runner == nil {
		runner = NewDefaultCommandRunner()
	}
	return &BaseManager{
		name:        name,
		managerType: managerType,
		runner:      runner,
	}
}

// === BASIC INFORMATION (Default implementations) ===

func (b *BaseManager) GetName() string {
	return b.name
}

func (b *BaseManager) GetType() string {
	return b.managerType
}

// GetRunner returns the command runner for subclasses
func (b *BaseManager) GetRunner() CommandRunner {
	return b.runner
}

func (b *BaseManager) IsAvailable() bool {
	// Default: try to run the command with --version or --help
	// Subclasses should override this with specific logic
	_, err := b.runner.Run(b.name, "--version")
	if err != nil {
		_, err = b.runner.Run(b.name, "--help")
	}
	return err == nil
}

func (b *BaseManager) GetVersion() (string, error) {
	// Default: try to get version output
	output, err := b.runner.Run(b.name, "--version")
	if err != nil {
		return "", fmt.Errorf("unable to get version for %s: %w", b.name, err)
	}
	return string(output), nil
}

// === CORE PACKAGE OPERATIONS (Default: not supported) ===

func (b *BaseManager) Search(ctx context.Context, query []string, opts *Options) ([]PackageInfo, error) {
	return nil, fmt.Errorf("%s: %w", b.name, ErrOperationNotSupported)
}

func (b *BaseManager) List(ctx context.Context, filter ListFilter, opts *Options) ([]PackageInfo, error) {
	return nil, fmt.Errorf("%s: %w", b.name, ErrOperationNotSupported)
}

func (b *BaseManager) Install(ctx context.Context, packages []string, opts *Options) ([]PackageInfo, error) {
	return nil, fmt.Errorf("%s: %w", b.name, ErrOperationNotSupported)
}

func (b *BaseManager) Remove(ctx context.Context, packages []string, opts *Options) ([]PackageInfo, error) {
	return nil, fmt.Errorf("%s: %w", b.name, ErrOperationNotSupported)
}

func (b *BaseManager) GetInfo(ctx context.Context, packageName string, opts *Options) (PackageInfo, error) {
	return PackageInfo{}, fmt.Errorf("%s: %w", b.name, ErrOperationNotSupported)
}

// === UPDATE/UPGRADE OPERATIONS (Default: not supported) ===

func (b *BaseManager) Refresh(ctx context.Context, opts *Options) error {
	return fmt.Errorf("%s: %w", b.name, ErrOperationNotSupported)
}

func (b *BaseManager) Upgrade(ctx context.Context, packages []string, opts *Options) ([]PackageInfo, error) {
	return nil, fmt.Errorf("%s: %w", b.name, ErrOperationNotSupported)
}

// === CLEANUP OPERATIONS (Default: not supported or no-op) ===

func (b *BaseManager) Clean(ctx context.Context, opts *Options) error {
	// Default: no-op for cleanup (many managers don't need this)
	b.LogVerbosef(opts, "Clean operation not needed for %s", b.name)
	return nil
}

func (b *BaseManager) AutoRemove(ctx context.Context, opts *Options) ([]PackageInfo, error) {
	// Default: no-op (return empty list)
	b.LogVerbosef(opts, "AutoRemove not supported for %s", b.name)
	return []PackageInfo{}, nil
}

// === HEALTH/STATUS OPERATIONS (Default implementations) ===

func (b *BaseManager) Verify(ctx context.Context, packages []string, opts *Options) ([]PackageInfo, error) {
	return nil, fmt.Errorf("%s: %w", b.name, ErrOperationNotSupported)
}

func (b *BaseManager) Status(ctx context.Context, opts *Options) (ManagerStatus, error) {
	// Default implementation provides basic status
	status := ManagerStatus{
		Available:      b.IsAvailable(),
		Healthy:        true, // Assume healthy if available
		Version:        "",
		LastRefresh:    "unknown",
		CacheSize:      0,
		PackageCount:   0,
		InstalledCount: 0,
		Issues:         []string{},
		Metadata:       make(map[string]interface{}),
	}

	// Try to get version
	if version, err := b.GetVersion(); err == nil {
		status.Version = version
	}

	// If not available, mark as unhealthy
	if !status.Available {
		status.Healthy = false
		status.Issues = append(status.Issues, fmt.Sprintf("%s is not available on this system", b.name))
	}

	return status, nil
}

// === HELPER METHODS ===

// LogVerbosef logs a message only if verbose mode is enabled
func (b *BaseManager) LogVerbosef(opts *Options, format string, args ...interface{}) {
	if opts != nil && opts.Verbose {
		log.Printf("[%s] "+format, append([]interface{}{b.name}, args...)...)
	}
}

// LogDebugf logs a message only if debug mode is enabled
func (b *BaseManager) LogDebugf(opts *Options, format string, args ...interface{}) {
	if opts != nil && opts.Debug {
		log.Printf("[%s DEBUG] "+format, append([]interface{}{b.name}, args...)...)
	}
}

// GetTimeoutContext creates a context with timeout from options
func (b *BaseManager) GetTimeoutContext(ctx context.Context, opts *Options) (context.Context, context.CancelFunc) {
	timeout := 30 * time.Second // Default timeout

	if opts != nil && opts.TimeoutSecs > 0 {
		timeout = time.Duration(opts.TimeoutSecs) * time.Second
	}

	return context.WithTimeout(ctx, timeout)
}

// ValidatePackageNames validates package names for security
func (b *BaseManager) ValidatePackageNames(packages []string) error {
	for _, pkg := range packages {
		if err := ValidatePackageName(pkg); err != nil {
			return fmt.Errorf("invalid package name '%s': %w", pkg, err)
		}
	}
	return nil
}

// HandleDryRun logs what would be done in dry-run mode
func (b *BaseManager) HandleDryRun(opts *Options, operation string, packages []string) {
	if opts != nil && opts.DryRun {
		if len(packages) > 0 {
			log.Printf("[%s DRY-RUN] Would %s packages: %v", b.name, operation, packages)
		} else {
			log.Printf("[%s DRY-RUN] Would %s", b.name, operation)
		}
	}
}

// SimplePlugin provides a basic plugin implementation that can be used by most package managers
type SimplePlugin struct {
	name        string
	managerType string
	priority    int
	createFunc  func() PackageManager
}

// NewSimplePlugin creates a new simple plugin
func NewSimplePlugin(name, managerType string, priority int, createFunc func() PackageManager) *SimplePlugin {
	return &SimplePlugin{
		name:        name,
		managerType: managerType,
		priority:    priority,
		createFunc:  createFunc,
	}
}

func (p *SimplePlugin) CreateManager() PackageManager {
	return p.createFunc()
}

func (p *SimplePlugin) GetPriority() int {
	return p.priority
}

// === COMMON UTILITIES ===

// ParsePackageNameVersion splits "package@version" or "package==version" into name and version
func ParsePackageNameVersion(pkg string) (name, version string) {
	// Common separators: @, ==, =
	for _, sep := range []string{"@", "==", "="} {
		if idx := indexString(pkg, sep); idx != -1 {
			return pkg[:idx], pkg[idx+len(sep):]
		}
	}
	return pkg, ""
}

// indexString returns the index of substr in s, or -1 if not found
func indexString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// NewPackageInfo creates a PackageInfo with common fields set
func NewPackageInfo(name, version, status, managerType string) PackageInfo {
	return PackageInfo{
		Name:        name,
		Version:     version,
		Status:      status,
		ManagerType: managerType,
		Metadata:    make(map[string]interface{}),
	}
}
