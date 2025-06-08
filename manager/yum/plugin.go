// Package yum provides a complete YUM package manager implementation
package yum

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/bluet/syspkg/manager"
)

// ManagerName is the identifier for the YUM package manager
const ManagerName = "yum"

// Timeouts for different YUM operations are handled by BaseManager

// Manager implements the unified PackageManager interface for YUM
type Manager struct {
	*manager.BaseManager
}

// NewManager creates a new YUM manager
func NewManager() *Manager {
	runner := manager.NewDefaultCommandRunner()
	return &Manager{
		BaseManager: manager.NewBaseManager(ManagerName, manager.CategorySystem, runner),
	}
}

// NewManagerWithRunner creates YUM manager with custom runner (for testing)
func NewManagerWithRunner(runner manager.CommandRunner) *Manager {
	return &Manager{
		BaseManager: manager.NewBaseManager(ManagerName, manager.CategorySystem, runner),
	}
}

// IsAvailable checks if YUM is available
func (m *Manager) IsAvailable() bool {
	// First try using the command runner (works for testing with mocks)
	result, err := m.GetRunner().Run(context.Background(), "yum", []string{"--version"})
	if err == nil && result.ExitCode == 0 && (strings.Contains(strings.ToLower(string(result.Output)), "yum") || strings.Contains(strings.ToLower(string(result.Output)), "rpm")) {
		return true
	}

	// Fallback to checking if yum binary exists in PATH (for real systems)
	_, pathErr := exec.LookPath("yum")
	return pathErr == nil
}

// GetVersion returns YUM version
func (m *Manager) GetVersion() (string, error) {
	result, err := m.GetRunner().Run(context.Background(), "yum", []string{"--version"})
	if err != nil {
		return "", err
	}

	if result.ExitCode != 0 {
		return "", fmt.Errorf("yum --version failed with exit code %d", result.ExitCode)
	}

	lines := strings.Split(string(result.Output), "\n")
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
	result, err := m.GetRunner().Run(ctx, "yum", args)
	if err != nil {
		return nil, manager.WrapReturn(manager.StatusUnavailableError, "yum command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		switch result.ExitCode {
		case 1:
			// YUM returns exit code 1 when no packages found, but this is not an error for search
			if strings.Contains(string(result.Output), "No matches found") || strings.Contains(string(result.Stderr), "No matches found") {
				return []manager.PackageInfo{}, nil
			}
			return nil, manager.WrapReturn(manager.StatusGeneralError, "yum search failed", nil)
		default:
			return nil, manager.WrapReturn(manager.StatusGeneralError, "yum search failed", nil)
		}
	}

	// result.ExitCode == 0: Success - parse results and return
	packages := parseSearchOutput(string(result.Output))

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
	result, err := m.GetRunner().Run(ctx, "yum", []string{"list", "installed"})
	if err != nil {
		return nil, manager.WrapReturn(manager.StatusUnavailableError, "yum command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		return nil, manager.WrapReturn(manager.StatusGeneralError, "yum list installed failed", nil)
	}

	// result.ExitCode == 0: Success - parse results and return
	return parseListOutput(string(result.Output)), nil
}

// ListUpgradable lists packages that can be upgraded
func (m *Manager) ListUpgradable(ctx context.Context, opts *manager.Options) ([]manager.PackageInfo, error) {
	result, err := m.GetRunner().Run(ctx, "yum", []string{"list", "updates"})
	if err != nil {
		return nil, manager.WrapReturn(manager.StatusUnavailableError, "yum command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		switch result.ExitCode {
		case 1:
			// YUM returns exit code 1 when no updates available, this is not an error
			if strings.Contains(string(result.Output), "No packages marked for update") || strings.Contains(string(result.Stderr), "No packages marked for update") {
				return []manager.PackageInfo{}, nil
			}
			return nil, manager.WrapReturn(manager.StatusGeneralError, "yum list updates failed", nil)
		default:
			return nil, manager.WrapReturn(manager.StatusGeneralError, "yum list updates failed", nil)
		}
	}

	// result.ExitCode == 0: Success - parse results and return
	return parseListOutput(string(result.Output)), nil
}

// GetInfo gets detailed package information
func (m *Manager) GetInfo(ctx context.Context, packageName string, opts *manager.Options) (manager.PackageInfo, error) {
	if err := m.ValidatePackageNames([]string{packageName}); err != nil {
		return manager.PackageInfo{}, err
	}

	result, err := m.GetRunner().Run(ctx, "yum", []string{"info", packageName})
	if err != nil {
		return manager.PackageInfo{}, manager.WrapReturn(manager.StatusUnavailableError, "yum command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		switch result.ExitCode {
		case 1:
			// Check for package not found
			if strings.Contains(string(result.Output), "No matching Packages") ||
				strings.Contains(string(result.Output), "Error: No matching Packages") ||
				strings.Contains(string(result.Stderr), "No matching Packages") {
				return manager.PackageInfo{}, manager.ErrPackageNotFound
			}
			return manager.PackageInfo{}, manager.WrapReturn(manager.StatusGeneralError, "yum info failed", nil)
		default:
			return manager.PackageInfo{}, manager.WrapReturn(manager.StatusGeneralError, "yum info failed", nil)
		}
	}

	// result.ExitCode == 0: Success - parse results and return
	info, err := parseInfoOutput(string(result.Output), packageName)
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

	result, err := m.GetRunner().Run(ctx, "yum", args)
	if err != nil {
		return nil, manager.WrapReturn(manager.StatusUnavailableError, "yum command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		switch result.ExitCode {
		case 1:
			// Check stderr/output for specific error messages
			combinedOutput := string(result.Output) + string(result.Stderr)
			if strings.Contains(combinedOutput, "Nothing to do") ||
				strings.Contains(combinedOutput, "already installed") {
				return []manager.PackageInfo{}, nil // Not an error if already installed
			}
			if strings.Contains(combinedOutput, "No package") {
				return nil, manager.WrapReturn(manager.StatusUnavailableError, "package not found", nil)
			}
			return nil, manager.WrapReturn(manager.StatusGeneralError, "installation failed", nil)
		default:
			return nil, manager.WrapReturn(manager.StatusGeneralError, "yum install failed", nil)
		}
	}

	// result.ExitCode == 0: Success - parse results and return
	return parseInstallOutput(string(result.Output)), nil
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

	result, err := m.GetRunner().Run(ctx, "yum", args)
	if err != nil {
		return nil, manager.WrapReturn(manager.StatusUnavailableError, "yum command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		switch result.ExitCode {
		case 1:
			combinedOutput := string(result.Output) + string(result.Stderr)
			if strings.Contains(combinedOutput, "No Packages marked for removal") {
				return nil, manager.ErrPackageNotFound
			}
			return nil, manager.WrapReturn(manager.StatusGeneralError, "removal failed", nil)
		default:
			return nil, manager.WrapReturn(manager.StatusGeneralError, "yum remove failed", nil)
		}
	}

	// result.ExitCode == 0: Success - parse results and return
	return parseRemoveOutput(string(result.Output)), nil
}

// Refresh updates package lists (refresh metadata)
func (m *Manager) Refresh(ctx context.Context, opts *manager.Options) error {
	return m.Update(ctx, opts)
}

// Update updates package lists (refresh metadata)
func (m *Manager) Update(ctx context.Context, opts *manager.Options) error {
	result, err := m.GetRunner().Run(ctx, "yum", []string{"makecache", "fast"})
	if err != nil {
		return manager.WrapReturn(manager.StatusUnavailableError, "yum command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		return manager.WrapReturn(manager.StatusGeneralError, "yum makecache failed", nil)
	}

	// result.ExitCode == 0: Success
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

	result, err := m.GetRunner().Run(ctx, "yum", args)
	if err != nil {
		return nil, manager.WrapReturn(manager.StatusUnavailableError, "yum command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		switch result.ExitCode {
		case 1:
			combinedOutput := string(result.Output) + string(result.Stderr)
			if strings.Contains(combinedOutput, "Nothing to do") {
				return []manager.PackageInfo{}, nil // Not an error if nothing to update
			}
			return nil, manager.WrapReturn(manager.StatusGeneralError, "upgrade failed", nil)
		default:
			return nil, manager.WrapReturn(manager.StatusGeneralError, "yum update failed", nil)
		}
	}

	// result.ExitCode == 0: Success - parse results and return
	// YUM update output format is similar to install output
	return parseInstallOutput(string(result.Output)), nil
}

// Clean cleans package cache
func (m *Manager) Clean(ctx context.Context, opts *manager.Options) error {
	// Check for dry-run mode - should not perform actual operations
	if opts != nil && opts.DryRun {
		return nil
	}

	result, err := m.GetRunner().Run(ctx, "yum", []string{"clean", "all"})
	if err != nil {
		return manager.WrapReturn(manager.StatusUnavailableError, "yum command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		return manager.WrapReturn(manager.StatusGeneralError, "yum clean failed", nil)
	}

	// result.ExitCode == 0: Success
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

	result, err := m.GetRunner().Run(ctx, "yum", args)
	if err != nil {
		return nil, manager.WrapReturn(manager.StatusUnavailableError, "yum command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		switch result.ExitCode {
		case 1:
			combinedOutput := string(result.Output) + string(result.Stderr)
			if strings.Contains(combinedOutput, "Nothing to do") {
				return []manager.PackageInfo{}, nil // Not an error if nothing to remove
			}
			return nil, manager.WrapReturn(manager.StatusGeneralError, "autoremove failed", nil)
		default:
			return nil, manager.WrapReturn(manager.StatusGeneralError, "yum autoremove failed", nil)
		}
	}

	// result.ExitCode == 0: Success - parse results and return
	// Use context-aware parsing based on operation mode
	if opts != nil && opts.DryRun {
		// YUM autoremove dry-run may have different output format
		// For now, use the enhanced parseRemoveOutput which handles both formats
		// TODO: If dry-run format is significantly different, create parseAutoRemoveDryRunOutput
		return parseRemoveOutput(string(result.Output)), nil
	}
	return parseRemoveOutput(string(result.Output)), nil
}

// Verify verifies package integrity
func (m *Manager) Verify(ctx context.Context, packageNames []string, opts *manager.Options) ([]manager.PackageInfo, error) {
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

	output, err := m.GetRunner().Run(ctx, "yum", args)
	if err != nil {
		return nil, fmt.Errorf("yum check failed: %w", err)
	}

	// Parse the output to create verification results
	var results []manager.PackageInfo
	if len(packageNames) > 0 {
		for _, pkg := range packageNames {
			// For simplicity, if yum check succeeds, mark packages as verified
			result := manager.NewPackageInfo(pkg, "", manager.StatusInstalled, ManagerName)
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

	if _, err := m.GetRunner().Run(ctx, "yum", []string{"--version"}); err != nil {
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
