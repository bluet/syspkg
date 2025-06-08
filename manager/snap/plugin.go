// Package snap provides a complete Snap package manager implementation
package snap

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/bluet/syspkg/manager"
)

// ManagerName is the identifier for the Snap package manager
const ManagerName = "snap"

// Manager implements the unified PackageManager interface for Snap
type Manager struct {
	*manager.BaseManager
}

// NewManager creates a new Snap manager
func NewManager() *Manager {
	runner := manager.NewDefaultCommandRunner()
	return &Manager{
		BaseManager: manager.NewBaseManager(ManagerName, manager.CategorySystem, runner),
	}
}

// NewManagerWithRunner creates Snap manager with custom runner (for testing)
func NewManagerWithRunner(runner manager.CommandRunner) *Manager {
	return &Manager{
		BaseManager: manager.NewBaseManager(ManagerName, manager.CategorySystem, runner),
	}
}

// IsAvailable checks if Snap is available
func (m *Manager) IsAvailable() bool {
	_, err := exec.LookPath("snap")
	if err != nil {
		return false
	}
	// Verify snap is working
	result, err := m.GetRunner().Run(context.Background(), "snap", []string{"version"})
	return err == nil && result.ExitCode == 0
}

// GetVersion returns Snap version
func (m *Manager) GetVersion() (string, error) {
	result, err := m.GetRunner().Run(context.Background(), "snap", []string{"version"})
	if err != nil {
		return "", err
	}

	if result.ExitCode != 0 {
		return "", fmt.Errorf("snap version failed with exit code %d", result.ExitCode)
	}

	lines := strings.Split(string(result.Output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "snap ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
	}
	return "", fmt.Errorf("could not parse snap version")
}

// Search searches for packages
func (m *Manager) Search(ctx context.Context, query []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	if len(query) == 0 {
		return nil, fmt.Errorf("search requires a query")
	}

	if err := manager.ValidatePackageNames(query); err != nil {
		return nil, err
	}

	args := []string{"find", "--unicode=never", "--color=never"}
	args = append(args, query...)

	result, err := m.GetRunner().Run(ctx, "snap", args)
	if err != nil {
		return nil, manager.WrapReturn(manager.StatusUnavailableError, "snap command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		switch result.ExitCode {
		case 1:
			// Snap: no packages found - that's okay for search
			return []manager.PackageInfo{}, nil
		default:
			return nil, manager.WrapReturn(manager.StatusGeneralError, "snap find failed", nil)
		}
	}

	// result.ExitCode == 0: Success - parse results and return
	return m.parseSearchOutput(string(result.Output)), nil
}

// List lists packages based on filter
func (m *Manager) List(ctx context.Context, filter manager.ListFilter, opts *manager.Options) ([]manager.PackageInfo, error) {
	switch filter {
	case manager.FilterInstalled:
		return m.listInstalled(ctx, opts)
	case manager.FilterAll:
		// Snap doesn't have a "list all available" command, so use search with broad terms
		return m.Search(ctx, []string{""}, opts)
	case manager.FilterUpgradable:
		return m.listUpgradable(ctx, opts)
	default:
		return nil, fmt.Errorf("unsupported filter: %v", filter)
	}
}

// GetInfo gets information about a specific package
func (m *Manager) GetInfo(ctx context.Context, packageName string, opts *manager.Options) (manager.PackageInfo, error) {
	if err := manager.ValidatePackageNames([]string{packageName}); err != nil {
		return manager.PackageInfo{}, err
	}

	result, err := m.GetRunner().Run(ctx, "snap", []string{"info", packageName})
	if err != nil {
		return manager.PackageInfo{}, manager.WrapReturn(manager.StatusUnavailableError, "snap command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		switch result.ExitCode {
		case 1:
			// Package not found
			return manager.PackageInfo{}, manager.WrapReturn(manager.StatusUnavailableError, "package not found", nil)
		default:
			return manager.PackageInfo{}, manager.WrapReturn(manager.StatusGeneralError, "snap info failed", nil)
		}
	}

	// result.ExitCode == 0: Success - parse results and return
	return m.parseInfoOutput(string(result.Output), packageName), nil
}

// Status returns package manager status
func (m *Manager) Status(ctx context.Context, opts *manager.Options) (manager.ManagerStatus, error) {
	version, _ := m.GetVersion()
	available := m.IsAvailable()

	return manager.ManagerStatus{
		Available: available,
		Healthy:   available,
		Version:   version,
		Issues:    []string{},
	}, nil
}

// Install installs packages
func (m *Manager) Install(ctx context.Context, packages []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	if len(packages) == 0 {
		return nil, fmt.Errorf("install requires package names")
	}

	if err := manager.ValidatePackageNames(packages); err != nil {
		return nil, err
	}

	var results []manager.PackageInfo
	for _, pkg := range packages {
		args := []string{"install"}
		if opts != nil && opts.DryRun {
			// Snap doesn't have a --dry-run, so we'll just return what would be installed
			info, err := m.GetInfo(ctx, pkg, opts)
			if err != nil {
				return nil, err
			}
			info.Status = "would-install"
			results = append(results, info)
			continue
		}

		args = append(args, pkg)
		_, err := m.GetRunner().Run(ctx, "snap", args)
		if err != nil {
			return nil, manager.WrapCommandError(fmt.Sprintf("snap install %s failed", pkg), err)
		}

		results = append(results, manager.NewPackageInfo(pkg, "", manager.StatusInstalled, "snap"))
	}

	return results, nil
}

// Remove removes packages
func (m *Manager) Remove(ctx context.Context, packages []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	if len(packages) == 0 {
		return nil, fmt.Errorf("remove requires package names")
	}

	if err := manager.ValidatePackageNames(packages); err != nil {
		return nil, err
	}

	var results []manager.PackageInfo
	for _, pkg := range packages {
		args := []string{"remove"}
		if opts != nil && opts.DryRun {
			// Snap doesn't have a --dry-run, so we'll just return what would be removed
			results = append(results, manager.NewPackageInfo(pkg, "", "would-remove", "snap"))
			continue
		}

		args = append(args, pkg)
		_, err := m.GetRunner().Run(ctx, "snap", args)
		if err != nil {
			return nil, manager.WrapCommandError(fmt.Sprintf("snap remove %s failed", pkg), err)
		}

		results = append(results, manager.NewPackageInfo(pkg, "", "removed", "snap"))
	}

	return results, nil
}

// Refresh refreshes package lists (equivalent to update)
func (m *Manager) Refresh(ctx context.Context, opts *manager.Options) error {
	_, err := m.GetRunner().Run(ctx, "snap", []string{"refresh", "--list"})
	if err != nil {
		return manager.WrapCommandError("snap refresh --list failed", err)
	}
	return nil
}

// Upgrade upgrades packages
func (m *Manager) Upgrade(ctx context.Context, packages []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	args := []string{"refresh"}

	if len(packages) > 0 {
		if err := manager.ValidatePackageNames(packages); err != nil {
			return nil, err
		}
		args = append(args, packages...)
	}

	if opts != nil && opts.DryRun {
		// Use --list to see what would be upgraded
		listArgs := []string{"refresh", "--list"}
		result, err := m.GetRunner().Run(ctx, "snap", listArgs)
		if err != nil {
			return nil, manager.WrapReturn(manager.StatusUnavailableError, "snap command failed", err)
		}

		// Handle error cases first (return-early pattern)
		if result.ExitCode != 0 {
			return nil, manager.WrapReturn(manager.StatusGeneralError, "snap refresh --list failed", nil)
		}

		// result.ExitCode == 0: Success - parse results and return
		return m.parseUpgradeListOutput(string(result.Output)), nil
	}

	result, err := m.GetRunner().Run(ctx, "snap", args)
	if err != nil {
		return nil, manager.WrapReturn(manager.StatusUnavailableError, "snap command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		switch result.ExitCode {
		case 1:
			// No updates available - not an error
			return []manager.PackageInfo{}, nil
		default:
			return nil, manager.WrapReturn(manager.StatusGeneralError, "snap refresh failed", nil)
		}
	}

	// result.ExitCode == 0: Success - parse results and return
	return m.parseUpgradeOutput(string(result.Output)), nil
}

// Clean is not applicable for Snap (packages are self-contained)
func (m *Manager) Clean(ctx context.Context, opts *manager.Options) error {
	// Snap doesn't have a traditional cache to clean
	return nil
}

// AutoRemove is not applicable for Snap (no dependency orphans)
func (m *Manager) AutoRemove(ctx context.Context, opts *manager.Options) ([]manager.PackageInfo, error) {
	// Snap doesn't have orphaned packages
	return []manager.PackageInfo{}, nil
}

// Verify checks package integrity
func (m *Manager) Verify(ctx context.Context, packages []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	if len(packages) == 0 {
		return nil, fmt.Errorf("verify requires specific package names")
	}

	if err := manager.ValidatePackageNames(packages); err != nil {
		return nil, err
	}

	var results []manager.PackageInfo
	for _, pkg := range packages {
		// Check if package is installed and functioning
		_, err := m.GetRunner().Run(ctx, "snap", []string{"list", pkg})
		status := manager.StatusInstalled
		if err != nil {
			status = "not-installed"
		}

		results = append(results, manager.NewPackageInfo(pkg, "", status, "snap"))
	}

	return results, nil
}

// parseSearchOutput parses snap find output
func (m *Manager) parseSearchOutput(output string) []manager.PackageInfo {
	var packages []manager.PackageInfo
	lines := strings.Split(output, "\n")

	// Skip header line
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		// Format: Name Version Publisher Notes Summary
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			name := parts[0]
			version := parts[1]
			publisher := parts[2]

			pkg := manager.NewPackageInfo(name, version, manager.StatusAvailable, "snap")
			pkg.Metadata["publisher"] = publisher

			// Notes (if present)
			if len(parts) >= 4 && parts[3] != "-" {
				pkg.Metadata["notes"] = parts[3]
			}

			// Summary (rest of the line)
			if len(parts) >= 5 {
				summary := strings.Join(parts[4:], " ")
				pkg.Description = summary
			}

			packages = append(packages, pkg)
		}
	}

	return packages
}

// parseInfoOutput parses snap info output
func (m *Manager) parseInfoOutput(output, packageName string) manager.PackageInfo {
	pkg := manager.NewPackageInfo(packageName, "", manager.StatusAvailable, "snap")

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				switch key {
				case "summary":
					pkg.Description = value
				case "publisher":
					pkg.Metadata["publisher"] = value
				case "license":
					pkg.Metadata["license"] = value
				case "description":
					pkg.Metadata["description"] = value
				}
			}
		}
	}

	return pkg
}

// listInstalled lists installed snap packages
func (m *Manager) listInstalled(ctx context.Context, _ *manager.Options) ([]manager.PackageInfo, error) {
	result, err := m.GetRunner().Run(ctx, "snap", []string{"list", "--unicode=never", "--color=never"})
	if err != nil {
		return nil, manager.WrapReturn(manager.StatusUnavailableError, "snap command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		return nil, manager.WrapReturn(manager.StatusGeneralError, "snap list failed", nil)
	}

	// result.ExitCode == 0: Success - parse results and return
	var packages []manager.PackageInfo
	lines := strings.Split(string(result.Output), "\n")

	// Skip header line
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		// Format: Name Version Rev Tracking Publisher Notes
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			name := parts[0]
			version := parts[1]

			pkg := manager.NewPackageInfo(name, version, manager.StatusInstalled, "snap")

			if len(parts) >= 3 {
				pkg.Metadata["rev"] = parts[2]
			}
			if len(parts) >= 4 {
				pkg.Metadata["tracking"] = parts[3]
			}
			if len(parts) >= 5 {
				pkg.Metadata["publisher"] = parts[4]
			}
			if len(parts) >= 6 {
				pkg.Metadata["notes"] = parts[5]
			}

			packages = append(packages, pkg)
		}
	}

	return packages, nil
}

// listUpgradable lists packages that can be upgraded
func (m *Manager) listUpgradable(ctx context.Context, _ *manager.Options) ([]manager.PackageInfo, error) {
	result, err := m.GetRunner().Run(ctx, "snap", []string{"refresh", "--list"})
	if err != nil {
		return nil, manager.WrapReturn(manager.StatusUnavailableError, "snap command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		switch result.ExitCode {
		case 1:
			// No updates available - not an error
			return []manager.PackageInfo{}, nil
		default:
			return nil, manager.WrapReturn(manager.StatusGeneralError, "snap refresh --list failed", nil)
		}
	}

	// result.ExitCode == 0: Success - parse results and return
	return m.parseUpgradeListOutput(string(result.Output)), nil
}

// parseUpgradeListOutput parses the output of snap refresh --list
func (m *Manager) parseUpgradeListOutput(output string) []manager.PackageInfo {
	var packages []manager.PackageInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "Name") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			name := parts[0]
			newVersion := parts[1]

			pkg := manager.NewPackageInfo(name, "", manager.StatusUpgradable, "snap")
			pkg.NewVersion = newVersion

			packages = append(packages, pkg)
		}
	}

	return packages
}

// parseUpgradeOutput parses the output of snap refresh
func (m *Manager) parseUpgradeOutput(output string) []manager.PackageInfo {
	var packages []manager.PackageInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.Contains(line, "refreshed") {
			// Look for lines like "packagename refreshed"
			parts := strings.Fields(line)
			if len(parts) >= 2 && parts[1] == "refreshed" {
				name := parts[0]
				pkg := manager.NewPackageInfo(name, "", "upgraded", "snap")
				packages = append(packages, pkg)
			}
		}
	}

	return packages
}

// Plugin for registration
type Plugin struct{}

func (p *Plugin) CreateManager() manager.PackageManager { return NewManager() }
func (p *Plugin) GetPriority() int                      { return 80 }

// Auto-register
func init() {
	_ = manager.Register("snap", &Plugin{})
}
