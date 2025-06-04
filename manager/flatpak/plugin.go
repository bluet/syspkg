// Package flatpak provides a complete Flatpak package manager implementation
package flatpak

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/bluet/syspkg/manager"
)

// Manager implements the unified PackageManager interface for Flatpak
type Manager struct {
	*manager.BaseManager
}

// NewManager creates a new Flatpak manager
func NewManager() *Manager {
	runner := manager.NewDefaultCommandRunner()
	return &Manager{
		BaseManager: manager.NewBaseManager("flatpak", manager.TypeSystem, runner),
	}
}

// NewManagerWithRunner creates Flatpak manager with custom runner (for testing)
func NewManagerWithRunner(runner manager.CommandRunner) *Manager {
	return &Manager{
		BaseManager: manager.NewBaseManager("flatpak", manager.TypeSystem, runner),
	}
}

// IsAvailable checks if Flatpak is available
func (m *Manager) IsAvailable() bool {
	_, err := exec.LookPath("flatpak")
	if err != nil {
		return false
	}
	// Verify flatpak is working
	_, err = m.GetRunner().Run("flatpak", "--version")
	return err == nil
}

// GetVersion returns Flatpak version
func (m *Manager) GetVersion() (string, error) {
	output, err := m.GetRunner().RunContext(context.Background(), "flatpak", []string{"--version"}, "LANG=C")
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Flatpak ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
	}
	return "", fmt.Errorf("could not parse flatpak version")
}

// Search searches for packages
func (m *Manager) Search(ctx context.Context, query []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	if len(query) == 0 {
		return nil, fmt.Errorf("search requires a query")
	}

	if err := manager.ValidatePackageNames(query); err != nil {
		return nil, err
	}

	args := []string{"search"}
	args = append(args, query...)

	output, err := m.GetRunner().RunContext(ctx, "flatpak", args, "LANG=C")
	if err != nil {
		return nil, fmt.Errorf("flatpak search failed: %w", err)
	}

	return m.parseSearchOutput(string(output)), nil
}

// List lists packages based on filter
func (m *Manager) List(ctx context.Context, filter manager.ListFilter, opts *manager.Options) ([]manager.PackageInfo, error) {
	switch filter {
	case manager.FilterInstalled:
		return m.listInstalled(ctx, opts)
	case manager.FilterAll:
		// Flatpak doesn't have a "list all available" command, so use search with broad terms
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

	output, err := m.GetRunner().RunContext(ctx, "flatpak", []string{"info", packageName}, "LANG=C")
	if err != nil {
		return manager.PackageInfo{}, fmt.Errorf("flatpak info failed: %w", err)
	}

	return m.parseInfoOutput(string(output), packageName), nil
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
		args := []string{"install", "-y"}
		if opts != nil && opts.DryRun {
			// Flatpak doesn't have a --dry-run, so we'll just return what would be installed
			info, err := m.GetInfo(ctx, pkg, opts)
			if err != nil {
				return nil, err
			}
			info.Status = "would-install"
			results = append(results, info)
			continue
		}

		args = append(args, pkg)
		_, err := m.GetRunner().RunContext(ctx, "flatpak", args, "LANG=C")
		if err != nil {
			return nil, fmt.Errorf("flatpak install %s failed: %w", pkg, err)
		}

		results = append(results, manager.NewPackageInfo(pkg, "", manager.StatusInstalled, manager.TypeSystem))
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
		args := []string{"uninstall", "-y"}
		if opts != nil && opts.DryRun {
			// Flatpak doesn't have a --dry-run, so we'll just return what would be removed
			results = append(results, manager.NewPackageInfo(pkg, "", "would-remove", manager.TypeSystem))
			continue
		}

		args = append(args, pkg)
		_, err := m.GetRunner().RunContext(ctx, "flatpak", args, "LANG=C")
		if err != nil {
			return nil, fmt.Errorf("flatpak uninstall %s failed: %w", pkg, err)
		}

		results = append(results, manager.NewPackageInfo(pkg, "", "removed", manager.TypeSystem))
	}

	return results, nil
}

// Refresh refreshes package lists (equivalent to update)
func (m *Manager) Refresh(ctx context.Context, opts *manager.Options) error {
	_, err := m.GetRunner().RunContext(ctx, "flatpak", []string{"update", "--appstream"}, "LANG=C")
	if err != nil {
		return fmt.Errorf("flatpak update --appstream failed: %w", err)
	}
	return nil
}

// Upgrade upgrades packages
func (m *Manager) Upgrade(ctx context.Context, packages []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	args := []string{"update", "-y"}

	if len(packages) > 0 {
		if err := manager.ValidatePackageNames(packages); err != nil {
			return nil, err
		}
		args = append(args, packages...)
	}

	if opts != nil && opts.DryRun {
		// Use list to see what's installed, then check for updates
		installed, err := m.listInstalled(ctx, opts)
		if err != nil {
			return nil, err
		}

		// Mark all as would-upgrade
		var results []manager.PackageInfo
		for _, pkg := range installed {
			pkg.Status = "would-upgrade"
			results = append(results, pkg)
		}
		return results, nil
	}

	output, err := m.GetRunner().RunContext(ctx, "flatpak", args, "LANG=C")
	if err != nil {
		return nil, fmt.Errorf("flatpak update failed: %w", err)
	}

	return m.parseUpdateOutput(string(output)), nil
}

// Clean is not applicable for Flatpak (automatic cleanup)
func (m *Manager) Clean(ctx context.Context, opts *manager.Options) error {
	// Clean unused runtimes
	_, err := m.GetRunner().RunContext(ctx, "flatpak", []string{"uninstall", "--unused", "-y"}, "LANG=C")
	if err != nil {
		return fmt.Errorf("flatpak uninstall --unused failed: %w", err)
	}
	return nil
}

// AutoRemove removes unused runtimes
func (m *Manager) AutoRemove(ctx context.Context, opts *manager.Options) ([]manager.PackageInfo, error) {
	args := []string{"uninstall", "--unused"}
	if opts != nil && !opts.DryRun {
		args = append(args, "-y")
	}

	output, err := m.GetRunner().RunContext(ctx, "flatpak", args, "LANG=C")
	if err != nil {
		return nil, fmt.Errorf("flatpak uninstall --unused failed: %w", err)
	}

	return m.parseUninstallOutput(string(output)), nil
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
		// Check if package is installed
		_, err := m.GetRunner().RunContext(ctx, "flatpak", []string{"info", pkg}, "LANG=C")
		status := manager.StatusInstalled
		if err != nil {
			status = "not-installed"
		}

		results = append(results, manager.NewPackageInfo(pkg, "", status, manager.TypeSystem))
	}

	return results, nil
}

// parseSearchOutput parses flatpak search output (tab-separated)
func (m *Manager) parseSearchOutput(output string) []manager.PackageInfo {
	var packages []manager.PackageInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Format: Name\tDescription\tApplication ID\tVersion\tBranch\tRemotes
		parts := strings.Split(line, "\t")
		if len(parts) >= 4 {
			name := strings.TrimSpace(parts[0])
			description := strings.TrimSpace(parts[1])
			appId := strings.TrimSpace(parts[2])
			version := strings.TrimSpace(parts[3])

			pkg := manager.NewPackageInfo(name, version, manager.StatusAvailable, manager.TypeSystem)
			pkg.Description = description
			pkg.Metadata["app_id"] = appId

			if len(parts) >= 5 {
				pkg.Metadata["branch"] = strings.TrimSpace(parts[4])
			}
			if len(parts) >= 6 {
				pkg.Metadata["remotes"] = strings.TrimSpace(parts[5])
			}

			packages = append(packages, pkg)
		}
	}

	return packages
}

// parseInfoOutput parses flatpak info output
func (m *Manager) parseInfoOutput(output, packageName string) manager.PackageInfo {
	pkg := manager.NewPackageInfo(packageName, "", manager.StatusAvailable, manager.TypeSystem)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				switch strings.ToLower(key) {
				case "id":
					pkg.Metadata["app_id"] = value
				case "version":
					pkg.Version = value
				case "description":
					pkg.Description = value
				case "license":
					pkg.Metadata["license"] = value
				case "origin":
					pkg.Metadata["origin"] = value
				}
			}
		}
	}

	return pkg
}

// listInstalled lists installed flatpak packages
func (m *Manager) listInstalled(ctx context.Context, _ *manager.Options) ([]manager.PackageInfo, error) {
	output, err := m.GetRunner().RunContext(ctx, "flatpak", []string{"list", "--app"}, "LANG=C")
	if err != nil {
		return nil, fmt.Errorf("flatpak list failed: %w", err)
	}

	var packages []manager.PackageInfo
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Format: Name\tApplication ID\tVersion\tBranch\tInstallation
		parts := strings.Split(line, "\t")
		if len(parts) >= 3 {
			name := strings.TrimSpace(parts[0])
			appId := strings.TrimSpace(parts[1])
			version := strings.TrimSpace(parts[2])

			pkg := manager.NewPackageInfo(name, version, manager.StatusInstalled, manager.TypeSystem)
			pkg.Metadata["app_id"] = appId

			if len(parts) >= 4 {
				pkg.Metadata["branch"] = strings.TrimSpace(parts[3])
			}
			if len(parts) >= 5 {
				pkg.Metadata["installation"] = strings.TrimSpace(parts[4])
			}

			packages = append(packages, pkg)
		}
	}

	return packages, nil
}

// listUpgradable lists packages that can be upgraded
func (m *Manager) listUpgradable(ctx context.Context, _ *manager.Options) ([]manager.PackageInfo, error) {
	// Check for updates without applying them
	output, err := m.GetRunner().RunContext(ctx, "flatpak", []string{"remote-ls", "--updates"}, "LANG=C")
	if err != nil {
		return nil, fmt.Errorf("flatpak remote-ls --updates failed: %w", err)
	}

	return m.parseUpdatesOutput(string(output)), nil
}

// parseUpdatesOutput parses the output of flatpak remote-ls --updates
func (m *Manager) parseUpdatesOutput(output string) []manager.PackageInfo {
	var packages []manager.PackageInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Format: Application ID\tVersion\tBranch\tRemote
		parts := strings.Split(line, "\t")
		if len(parts) >= 2 {
			appId := strings.TrimSpace(parts[0])
			version := strings.TrimSpace(parts[1])

			pkg := manager.NewPackageInfo(appId, "", manager.StatusUpgradable, manager.TypeSystem)
			pkg.NewVersion = version
			pkg.Metadata["app_id"] = appId

			if len(parts) >= 3 {
				pkg.Metadata["branch"] = strings.TrimSpace(parts[2])
			}
			if len(parts) >= 4 {
				pkg.Metadata["remote"] = strings.TrimSpace(parts[3])
			}

			packages = append(packages, pkg)
		}
	}

	return packages
}

// parseUpdateOutput parses the output of flatpak update
func (m *Manager) parseUpdateOutput(output string) []manager.PackageInfo {
	var packages []manager.PackageInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.Contains(line, "Updated") {
			// Look for lines indicating updates
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				name := parts[1] // Usually the second field is the app name
				pkg := manager.NewPackageInfo(name, "", "upgraded", manager.TypeSystem)
				packages = append(packages, pkg)
			}
		}
	}

	return packages
}

// parseUninstallOutput parses the output of flatpak uninstall --unused
func (m *Manager) parseUninstallOutput(output string) []manager.PackageInfo {
	var packages []manager.PackageInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.Contains(line, "Uninstalling") {
			// Look for lines indicating removal
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				name := parts[1]
				pkg := manager.NewPackageInfo(name, "", "removed", manager.TypeSystem)
				packages = append(packages, pkg)
			}
		}
	}

	return packages
}

// Plugin for registration
type Plugin struct{}

func (p *Plugin) CreateManager() manager.PackageManager { return NewManager() }
func (p *Plugin) GetPriority() int                      { return 70 }

// Auto-register
func init() {
	_ = manager.Register("flatpak", &Plugin{})
}
