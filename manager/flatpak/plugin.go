// Package flatpak provides a complete Flatpak package manager implementation
package flatpak

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/bluet/syspkg/manager"
)

// ManagerName is the identifier for the Flatpak package manager
const ManagerName = "flatpak"

// Manager implements the unified PackageManager interface for Flatpak
type Manager struct {
	*manager.BaseManager
}

// NewManager creates a new Flatpak manager
func NewManager() *Manager {
	runner := manager.NewDefaultCommandRunner()
	return &Manager{
		BaseManager: manager.NewBaseManager(ManagerName, manager.CategorySystem, runner),
	}
}

// NewManagerWithRunner creates Flatpak manager with custom runner (for testing)
func NewManagerWithRunner(runner manager.CommandRunner) *Manager {
	return &Manager{
		BaseManager: manager.NewBaseManager(ManagerName, manager.CategorySystem, runner),
	}
}

// IsAvailable checks if Flatpak is available
func (m *Manager) IsAvailable() bool {
	_, err := exec.LookPath("flatpak")
	if err != nil {
		return false
	}
	// Verify flatpak is working
	result, err := m.GetRunner().Run(context.Background(), "flatpak", []string{"--version"})
	return err == nil && result.ExitCode == 0
}

// GetVersion returns Flatpak version
func (m *Manager) GetVersion() (string, error) {
	result, err := m.GetRunner().Run(context.Background(), "flatpak", []string{"--version"}, "LANG=C")
	if err != nil {
		return "", err
	}

	if result.ExitCode != 0 {
		return "", fmt.Errorf("flatpak --version failed with exit code %d", result.ExitCode)
	}

	lines := strings.Split(string(result.Output), "\n")
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

	result, err := m.GetRunner().Run(ctx, "flatpak", args, "LANG=C")
	if err != nil {
		return nil, manager.WrapReturn(manager.StatusUnavailableError, "flatpak command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		switch result.ExitCode {
		case 1:
			// Flatpak: no packages found - that's okay for search
			return []manager.PackageInfo{}, nil
		default:
			return nil, manager.WrapReturn(manager.StatusGeneralError, "flatpak search failed", nil)
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

	result, err := m.GetRunner().Run(ctx, "flatpak", []string{"info", packageName}, "LANG=C")
	if err != nil {
		return manager.PackageInfo{}, manager.WrapReturn(manager.StatusUnavailableError, "flatpak command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		switch result.ExitCode {
		case 1:
			// Package not found
			return manager.PackageInfo{}, manager.WrapReturn(manager.StatusUnavailableError, "package not found", nil)
		default:
			return manager.PackageInfo{}, manager.WrapReturn(manager.StatusGeneralError, "flatpak info failed", nil)
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
		result, err := m.GetRunner().Run(ctx, "flatpak", args, "LANG=C")
		if err != nil {
			return nil, manager.WrapReturn(manager.StatusUnavailableError, "flatpak command failed", err)
		}

		// Handle error cases first (return-early pattern)
		if result.ExitCode != 0 {
			switch result.ExitCode {
			case 1:
				// Package not found or already installed
				combinedOutput := string(result.Output) + string(result.Stderr)
				if strings.Contains(combinedOutput, "not found") {
					return nil, manager.WrapReturn(manager.StatusUnavailableError, fmt.Sprintf("package %s not found", pkg), nil)
				}
				return nil, manager.WrapReturn(manager.StatusGeneralError, fmt.Sprintf("installation of %s failed", pkg), nil)
			default:
				return nil, manager.WrapReturn(manager.StatusGeneralError, fmt.Sprintf("flatpak install %s failed", pkg), nil)
			}
		}

		// result.ExitCode == 0: Success
		results = append(results, manager.NewPackageInfo(pkg, "", manager.StatusInstalled, ManagerName))
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
			results = append(results, manager.NewPackageInfo(pkg, "", "would-remove", ManagerName))
			continue
		}

		args = append(args, pkg)
		_, err := m.GetRunner().Run(ctx, "flatpak", args, "LANG=C")
		if err != nil {
			return nil, manager.WrapCommandError(fmt.Sprintf("flatpak uninstall %s failed", pkg), err)
		}

		results = append(results, manager.NewPackageInfo(pkg, "", "removed", ManagerName))
	}

	return results, nil
}

// Refresh refreshes package lists (equivalent to update)
func (m *Manager) Refresh(ctx context.Context, opts *manager.Options) error {
	_, err := m.GetRunner().Run(ctx, "flatpak", []string{"update", "--appstream"}, "LANG=C")
	if err != nil {
		return manager.WrapCommandError("flatpak update --appstream failed", err)
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

	result, err := m.GetRunner().Run(ctx, "flatpak", args, "LANG=C")
	if err != nil {
		return nil, manager.WrapReturn(manager.StatusUnavailableError, "flatpak command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		switch result.ExitCode {
		case 1:
			// No updates available - not an error
			return []manager.PackageInfo{}, nil
		default:
			return nil, manager.WrapReturn(manager.StatusGeneralError, "flatpak update failed", nil)
		}
	}

	// result.ExitCode == 0: Success - parse results and return
	return m.parseUpdateOutput(string(result.Output)), nil
}

// Clean is not applicable for Flatpak (automatic cleanup)
func (m *Manager) Clean(ctx context.Context, opts *manager.Options) error {
	// Check for dry-run mode - should not perform actual operations
	if opts != nil && opts.DryRun {
		return nil
	}

	// Clean unused runtimes
	_, err := m.GetRunner().Run(ctx, "flatpak", []string{"uninstall", "--unused", "-y"}, "LANG=C")
	if err != nil {
		return manager.WrapCommandError("flatpak uninstall --unused failed", err)
	}
	return nil
}

// AutoRemove removes unused runtimes
func (m *Manager) AutoRemove(ctx context.Context, opts *manager.Options) ([]manager.PackageInfo, error) {
	args := []string{"uninstall", "--unused"}
	if opts != nil && !opts.DryRun {
		args = append(args, "-y")
	}

	result, err := m.GetRunner().Run(ctx, "flatpak", args, "LANG=C")
	if err != nil {
		return nil, manager.WrapReturn(manager.StatusUnavailableError, "flatpak command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		switch result.ExitCode {
		case 1:
			// No unused packages - not an error
			return []manager.PackageInfo{}, nil
		default:
			return nil, manager.WrapReturn(manager.StatusGeneralError, "flatpak uninstall --unused failed", nil)
		}
	}

	// result.ExitCode == 0: Success - parse results and return
	return m.parseUninstallOutput(string(result.Output)), nil
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
		_, err := m.GetRunner().Run(ctx, "flatpak", []string{"info", pkg}, "LANG=C")
		status := manager.StatusInstalled
		if err != nil {
			status = "not-installed"
		}

		results = append(results, manager.NewPackageInfo(pkg, "", status, ManagerName))
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

			// Skip entries with empty package names
			if name == "" {
				continue
			}

			pkg := manager.NewPackageInfo(name, version, manager.StatusAvailable, ManagerName)
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
	pkg := manager.NewPackageInfo(packageName, "", manager.StatusAvailable, ManagerName)

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
	result, err := m.GetRunner().Run(ctx, "flatpak", []string{"list", "--user", "--columns=name,version,origin"}, "LANG=C")
	if err != nil {
		return nil, manager.WrapReturn(manager.StatusUnavailableError, "flatpak command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		return nil, manager.WrapReturn(manager.StatusGeneralError, "flatpak list failed", nil)
	}

	// result.ExitCode == 0: Success - parse results and return
	var packages []manager.PackageInfo
	lines := strings.Split(string(result.Output), "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Format: Name\tVersion\tOrigin (3 columns from --columns=name,version,origin)
		parts := strings.Split(line, "\t")
		if len(parts) >= 3 {
			name := strings.TrimSpace(parts[0])
			version := strings.TrimSpace(parts[1])
			origin := strings.TrimSpace(parts[2])

			pkg := manager.NewPackageInfo(name, version, manager.StatusInstalled, ManagerName)
			pkg.Metadata["origin"] = origin

			packages = append(packages, pkg)
		}
	}

	return packages, nil
}

// listUpgradable lists packages that can be upgraded
func (m *Manager) listUpgradable(ctx context.Context, _ *manager.Options) ([]manager.PackageInfo, error) {
	// Check for updates without applying them
	result, err := m.GetRunner().Run(ctx, "flatpak", []string{"list", "--user", "--updates", "--columns=name,version,origin"}, "LANG=C")
	if err != nil {
		return nil, manager.WrapReturn(manager.StatusUnavailableError, "flatpak command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		switch result.ExitCode {
		case 1:
			// No updates available - not an error
			return []manager.PackageInfo{}, nil
		default:
			return nil, manager.WrapReturn(manager.StatusGeneralError, "flatpak remote-ls --updates failed", nil)
		}
	}

	// result.ExitCode == 0: Success - parse results and return
	return m.parseUpdatesOutput(string(result.Output)), nil
}

// parseUpdatesOutput parses the output of flatpak remote-ls --updates
func (m *Manager) parseUpdatesOutput(output string) []manager.PackageInfo {
	var packages []manager.PackageInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Format: Name\tVersion\tOrigin (3 columns from --columns=name,version,origin)
		parts := strings.Split(line, "\t")
		if len(parts) >= 3 {
			name := strings.TrimSpace(parts[0])
			version := strings.TrimSpace(parts[1])
			origin := strings.TrimSpace(parts[2])

			pkg := manager.NewPackageInfo(name, version, manager.StatusUpgradable, ManagerName)
			pkg.Metadata["origin"] = origin

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
				fullName := parts[1] // Usually format: "app.id/arch/branch"
				// Extract just the app ID (before first slash)
				name := fullName
				if slashIndex := strings.Index(fullName, "/"); slashIndex != -1 {
					name = fullName[:slashIndex]
				}
				pkg := manager.NewPackageInfo(name, "", "upgraded", ManagerName)
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
				pkg := manager.NewPackageInfo(name, "", "removed", ManagerName)
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
