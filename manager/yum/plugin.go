// Package yum provides a complete YUM package manager implementation
package yum

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/bluet/syspkg/manager"
)

// Timeouts for different YUM operations are handled by BaseManager

// Manager implements the unified PackageManager interface for YUM
type Manager struct {
	*manager.BaseManager
}

// NewManager creates a new YUM manager
func NewManager() *Manager {
	runner := manager.NewDefaultCommandRunner()
	return &Manager{
		BaseManager: manager.NewBaseManager("yum", manager.TypeSystem, runner),
	}
}

// NewManagerWithRunner creates YUM manager with custom runner (for testing)
func NewManagerWithRunner(runner manager.CommandRunner) *Manager {
	return &Manager{
		BaseManager: manager.NewBaseManager("yum", manager.TypeSystem, runner),
	}
}

// IsAvailable checks if YUM is available
func (m *Manager) IsAvailable() bool {
	// First try using the command runner (works for testing with mocks)
	output, err := m.GetRunner().Run("yum", "--version")
	if err == nil && (strings.Contains(strings.ToLower(string(output)), "yum") || strings.Contains(strings.ToLower(string(output)), "rpm")) {
		return true
	}

	// Fallback to checking if yum binary exists in PATH (for real systems)
	_, pathErr := exec.LookPath("yum")
	if pathErr != nil {
		return false
	}

	// Try again with exec (in case runner failed but binary exists)
	return err == nil
}

// GetVersion returns YUM version
func (m *Manager) GetVersion() (string, error) {
	output, err := m.GetRunner().Run("yum", "--version")
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}
	return "unknown", nil
}

// Search searches for packages
func (m *Manager) Search(ctx context.Context, query []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	if opts == nil {
		opts = manager.DefaultOptions()
	}

	if err := m.ValidatePackageNames(query); err != nil {
		return nil, err
	}

	args := append([]string{"search"}, query...)
	output, err := m.GetRunner().RunContext(ctx, "yum", args)
	if err != nil {
		// YUM returns exit code 1 when no packages found, but this is not an error for search
		if strings.Contains(string(output), "No matches found") {
			return []manager.PackageInfo{}, nil
		}
		return nil, fmt.Errorf("yum search failed: %w", err)
	}

	packages := parseSearchOutput(string(output))

	// If status enhancement is requested, get detailed status
	if opts.ShowStatus {
		packages = m.enhanceWithDetailedStatus(packages)
	}

	return packages, nil
}

// List packages based on filter
func (m *Manager) List(ctx context.Context, filter manager.ListFilter, opts *manager.Options) ([]manager.PackageInfo, error) {
	switch filter {
	case manager.FilterInstalled:
		return m.ListInstalled(ctx, opts)
	case manager.FilterUpgradable:
		return m.ListUpgradable(ctx, opts)
	case manager.FilterAvailable:
		return nil, fmt.Errorf("listing all available packages not supported - use Search instead")
	default:
		return nil, fmt.Errorf("unsupported filter: %s", filter)
	}
}

// ListInstalled lists installed packages
func (m *Manager) ListInstalled(ctx context.Context, opts *manager.Options) ([]manager.PackageInfo, error) {
	// opts is not used in this function, so no need to handle nil case

	output, err := m.GetRunner().RunContext(ctx, "yum", []string{"list", "installed"})
	if err != nil {
		return nil, fmt.Errorf("yum list installed failed: %w", err)
	}

	return parseListOutput(string(output)), nil
}

// ListUpgradable lists packages that can be upgraded
func (m *Manager) ListUpgradable(ctx context.Context, opts *manager.Options) ([]manager.PackageInfo, error) {
	// opts is not used in this function, so no need to handle nil case

	output, err := m.GetRunner().RunContext(ctx, "yum", []string{"list", "updates"})
	if err != nil {
		// YUM returns exit code 1 when no updates available, this is not an error
		if strings.Contains(string(output), "No packages marked for update") {
			return []manager.PackageInfo{}, nil
		}
		return nil, fmt.Errorf("yum list updates failed: %w", err)
	}

	return parseListOutput(string(output)), nil
}

// GetInfo gets detailed package information
func (m *Manager) GetInfo(ctx context.Context, packageName string, opts *manager.Options) (manager.PackageInfo, error) {
	// opts is not used in this function, so no need to handle nil case

	if err := m.ValidatePackageNames([]string{packageName}); err != nil {
		return manager.PackageInfo{}, err
	}

	output, err := m.GetRunner().RunContext(ctx, "yum", []string{"info", packageName})
	if err != nil {
		if strings.Contains(string(output), "No matching Packages") ||
			strings.Contains(string(output), "Error: No matching Packages") {
			return manager.PackageInfo{}, manager.ErrPackageNotFound
		}
		return manager.PackageInfo{}, fmt.Errorf("yum info failed: %w", err)
	}

	info, err := parseInfoOutput(string(output), packageName)
	if err != nil {
		return manager.PackageInfo{}, err
	}
	return info, nil
}

// Install installs packages
func (m *Manager) Install(ctx context.Context, packageNames []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	localOpts := opts
	if localOpts == nil {
		localOpts = manager.DefaultOptions()
	}

	if err := m.ValidatePackageNames(packageNames); err != nil {
		return nil, err
	}

	args := []string{"install"}
	if localOpts.DryRun {
		args = append(args, "--setopt=tsflags=test")
	}
	if localOpts.AssumeYes {
		args = append(args, "-y")
	}
	args = append(args, packageNames...)

	output, err := m.GetRunner().RunContext(ctx, "yum", args)
	if err != nil {
		if strings.Contains(err.Error(), "Nothing to do") ||
			strings.Contains(err.Error(), "already installed") {
			return []manager.PackageInfo{}, nil // Not an error if already installed
		}
		return nil, fmt.Errorf("yum install failed: %w", err)
	}

	// Parse the output to extract installed packages
	return parseInstallOutput(string(output)), nil
}

// Remove removes packages
func (m *Manager) Remove(ctx context.Context, packageNames []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	if opts == nil {
		opts = manager.DefaultOptions()
	}

	if err := m.ValidatePackageNames(packageNames); err != nil {
		return nil, err
	}

	args := []string{"remove"}
	if opts.DryRun {
		args = append(args, "--setopt=tsflags=test")
	}
	if opts.AssumeYes {
		args = append(args, "-y")
	}
	args = append(args, packageNames...)

	output, err := m.GetRunner().RunContext(ctx, "yum", args)
	if err != nil {
		if strings.Contains(err.Error(), "No Packages marked for removal") {
			return nil, manager.ErrPackageNotFound
		}
		return nil, fmt.Errorf("yum remove failed: %w", err)
	}

	// Parse the output to extract removed packages
	return parseRemoveOutput(string(output)), nil
}

// Refresh updates package lists (refresh metadata)
func (m *Manager) Refresh(ctx context.Context, opts *manager.Options) error {
	return m.Update(ctx, opts)
}

// Update updates package lists (refresh metadata)
func (m *Manager) Update(ctx context.Context, opts *manager.Options) error {
	// opts is not used in this function, so no need to handle nil case

	_, err := m.GetRunner().RunContext(ctx, "yum", []string{"makecache", "fast"})
	if err != nil {
		return fmt.Errorf("yum makecache failed: %w", err)
	}

	return nil
}

// Upgrade upgrades packages
func (m *Manager) Upgrade(ctx context.Context, packageNames []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	if opts == nil {
		opts = manager.DefaultOptions()
	}

	if len(packageNames) > 0 {
		if err := m.ValidatePackageNames(packageNames); err != nil {
			return nil, err
		}
	}

	args := []string{"update"}
	if opts.DryRun {
		args = append(args, "--setopt=tsflags=test")
	}
	if opts.AssumeYes {
		args = append(args, "-y")
	}
	args = append(args, packageNames...)

	output, err := m.GetRunner().RunContext(ctx, "yum", args)
	if err != nil {
		if strings.Contains(err.Error(), "Nothing to do") {
			return []manager.PackageInfo{}, nil // Not an error if nothing to update
		}
		return nil, fmt.Errorf("yum update failed: %w", err)
	}

	// Parse the output to extract upgraded packages
	// YUM update output format is similar to install output
	return parseInstallOutput(string(output)), nil
}

// Clean cleans package cache
func (m *Manager) Clean(ctx context.Context, opts *manager.Options) error {
	// opts is not used in this function, so no need to handle nil case

	_, err := m.GetRunner().RunContext(ctx, "yum", []string{"clean", "all"})
	if err != nil {
		return fmt.Errorf("yum clean failed: %w", err)
	}

	return nil
}

// AutoRemove removes automatically installed packages that are no longer needed
func (m *Manager) AutoRemove(ctx context.Context, opts *manager.Options) ([]manager.PackageInfo, error) {
	if opts == nil {
		opts = manager.DefaultOptions()
	}

	args := []string{"autoremove"}
	if opts.DryRun {
		args = append(args, "--setopt=tsflags=test")
	}
	if opts.AssumeYes {
		args = append(args, "-y")
	}

	output, err := m.GetRunner().RunContext(ctx, "yum", args)
	if err != nil {
		if strings.Contains(err.Error(), "Nothing to do") {
			return []manager.PackageInfo{}, nil // Not an error if nothing to remove
		}
		return nil, fmt.Errorf("yum autoremove failed: %w", err)
	}

	// Parse the output to extract removed packages
	// Use context-aware parsing based on operation mode
	if opts != nil && opts.DryRun {
		// YUM autoremove dry-run may have different output format
		// For now, use the enhanced parseRemoveOutput which handles both formats
		// TODO: If dry-run format is significantly different, create parseAutoRemoveDryRunOutput
		return parseRemoveOutput(string(output)), nil
	}
	return parseRemoveOutput(string(output)), nil
}

// Verify verifies package integrity
func (m *Manager) Verify(ctx context.Context, packageNames []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	// opts is not used in this function, so no need to handle nil case

	if len(packageNames) > 0 {
		if err := m.ValidatePackageNames(packageNames); err != nil {
			return nil, err
		}
	}

	// Use yum check for package verification
	args := []string{"check"}
	if len(packageNames) > 0 {
		args = append(args, packageNames...)
	}

	output, err := m.GetRunner().RunContext(ctx, "yum", args)
	if err != nil {
		return nil, fmt.Errorf("yum check failed: %w", err)
	}

	// Parse the output to create verification results
	var results []manager.PackageInfo
	if len(packageNames) > 0 {
		for _, pkg := range packageNames {
			// For simplicity, if yum check succeeds, mark packages as verified
			result := manager.NewPackageInfo(pkg, "", manager.StatusInstalled, manager.TypeSystem)
			result.Metadata = make(map[string]interface{})
			result.Metadata["verified"] = true
			results = append(results, result)
		}
	}

	_ = output
	return results, nil
}

// Status returns overall status/health of the YUM package manager
func (m *Manager) Status(ctx context.Context, opts *manager.Options) (manager.ManagerStatus, error) {
	if opts == nil {
		opts = manager.DefaultOptions()
	}

	status := manager.ManagerStatus{
		Available: m.IsAvailable(),
		Healthy:   true,
		Issues:    []string{},
		Metadata:  make(map[string]interface{}),
	}

	// Get version
	if version, err := m.GetVersion(); err == nil {
		status.Version = version
	}

	// Check if we can access YUM
	if _, err := m.GetRunner().RunContext(ctx, "yum", []string{"--version"}); err != nil {
		status.Healthy = false
		status.Issues = append(status.Issues, "YUM command not accessible")
	}

	// Get installed package count
	if installed, err := m.ListInstalled(ctx, opts); err == nil {
		status.InstalledCount = len(installed)
	}

	return status, nil
}

// Plugin for registration
type Plugin struct{}

func (p *Plugin) CreateManager() manager.PackageManager { return NewManager() }
func (p *Plugin) GetPriority() int                      { return 80 }

// Auto-register
func init() {
	_ = manager.Register("yum", &Plugin{})
}
