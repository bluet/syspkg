// Package apt provides a complete APT package manager implementation
package apt

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/bluet/syspkg/manager"
)

// Manager implements the unified PackageManager interface for APT
type Manager struct {
	*manager.BaseManager
}

// NewManager creates a new APT manager
func NewManager() *Manager {
	runner := manager.NewDefaultCommandRunner()
	return &Manager{
		BaseManager: manager.NewBaseManager("apt", manager.CategorySystem, runner),
	}
}

// NewManagerWithRunner creates APT manager with custom runner (for testing)
func NewManagerWithRunner(runner manager.CommandRunner) *Manager {
	return &Manager{
		BaseManager: manager.NewBaseManager("apt", manager.CategorySystem, runner),
	}
}

// IsAvailable checks if APT is available
func (m *Manager) IsAvailable() bool {
	_, err := exec.LookPath("apt")
	if err != nil {
		return false
	}
	// Verify it's Debian apt (not Java apt)
	result, err := m.GetRunner().Run(context.Background(), "apt", []string{"--version"})
	return err == nil && strings.Contains(string(result.Output), "apt") && !strings.Contains(string(result.Output), "java")
}

// GetVersion returns APT version
func (m *Manager) GetVersion() (string, error) {
	result, err := m.GetRunner().Run(context.Background(), "apt", []string{"--version"})
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(result.Output), "\n")
	if len(lines) > 0 {
		parts := strings.Fields(lines[0])
		if len(parts) >= 2 {
			return parts[1], nil
		}
	}
	return strings.TrimSpace(string(result.Output)), nil
}

// Search finds packages
func (m *Manager) Search(ctx context.Context, query []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	if opts == nil {
		opts = manager.DefaultOptions()
	}

	if err := m.ValidatePackageNames(query); err != nil {
		return nil, err
	}

	args := append([]string{"search"}, query...)
	result, err := m.GetRunner().Run(ctx, "apt", args)
	if err != nil {
		// Command execution failed (e.g., apt not found)
		return nil, manager.WrapReturn(manager.StatusUnavailableError, "apt command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		switch result.ExitCode {
		case 1:
			// APT: no packages found or general error
			stderrStr := string(result.Stderr)
			if strings.Contains(stderrStr, "not found") || strings.Contains(stderrStr, "No packages found") {
				return []manager.PackageInfo{}, nil // Empty results for search is okay
			}
			return nil, manager.WrapReturn(manager.StatusGeneralError, "apt search failed", nil)
		case 2:
			// APT: invalid usage/arguments
			return nil, manager.WrapReturn(manager.StatusUsageError, "invalid search arguments", nil)
		case 100:
			// APT: package cache needs update
			return nil, manager.WrapReturn(manager.StatusGeneralError, "package cache needs update - run 'apt update'", nil)
		default:
			// Unknown exit code
			return nil, manager.WrapReturn(manager.StatusGeneralError, "apt search failed", nil)
		}
	}

	// result.ExitCode == 0: Success - parse and process results
	packages := parseSearchOutput(string(result.Output))

	// If ShowStatus is requested, enhance with more detailed status info
	if opts.ShowStatus {
		return m.addStatusInfo(ctx, packages, opts)
	}

	return packages, nil
}

// addStatusInfo adds real status information to search results
func (m *Manager) addStatusInfo(ctx context.Context, packages []manager.PackageInfo, opts *manager.Options) ([]manager.PackageInfo, error) {
	// Get installed packages for status cross-reference
	installed, err := m.listInstalled(ctx, opts)
	if err != nil {
		// If we can't get installed packages, return packages as-is
		return packages, nil
	}

	// Create lookup map for installed packages
	installedMap := make(map[string]manager.PackageInfo)
	for _, pkg := range installed {
		installedMap[pkg.Name] = pkg
	}

	// Update status information
	for i, pkg := range packages {
		if installedPkg, isInstalled := installedMap[pkg.Name]; isInstalled {
			// Package is installed
			packages[i].Status = manager.StatusInstalled
			packages[i].Version = installedPkg.Version

			// Check if upgradable (repo version differs from installed)
			if pkg.Version != "" && pkg.Version != installedPkg.Version {
				packages[i].Status = manager.StatusUpgradable
				packages[i].NewVersion = pkg.Version
			} else {
				packages[i].NewVersion = installedPkg.Version
			}
		}
		// If not installed, keep original status (usually "available")
	}

	return packages, nil
}

// List packages based on filter
func (m *Manager) List(ctx context.Context, filter manager.ListFilter, opts *manager.Options) ([]manager.PackageInfo, error) {
	switch filter {
	case manager.FilterInstalled:
		return m.listInstalled(ctx, opts)
	case manager.FilterUpgradable:
		return m.listUpgradable(ctx, opts)
	case manager.FilterAvailable:
		return nil, fmt.Errorf("listing all available packages not supported - use Search instead")
	default:
		return nil, fmt.Errorf("unsupported filter: %s", filter)
	}
}

// Install packages
func (m *Manager) Install(ctx context.Context, packages []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	if opts == nil {
		opts = manager.DefaultOptions()
	}

	if err := m.ValidatePackageNames(packages); err != nil {
		return nil, err
	}

	m.HandleDryRun(opts, "install", packages)

	args := []string{"install", "-y"}
	args = append(args, packages...)

	if opts.DryRun {
		args = append(args, "--dry-run")
	}

	result, err := m.GetRunner().Run(ctx, "apt", args, "DEBIAN_FRONTEND=noninteractive")
	if err != nil {
		return nil, manager.WrapReturn(manager.StatusUnavailableError, "apt command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		switch result.ExitCode {
		case 1:
			// APT: general error, check stderr for specifics
			stderrStr := string(result.Stderr)
			if strings.Contains(stderrStr, "not found") || strings.Contains(stderrStr, "No such package") {
				return nil, manager.WrapReturn(manager.StatusUnavailableError, "package not found in repository", nil)
			}
			return nil, manager.WrapReturn(manager.StatusGeneralError, "installation failed", nil)
		case 100:
			// APT: permission or lock error
			return nil, manager.WrapReturn(manager.StatusPermissionError, "installation requires root access or system is locked", nil)
		default:
			return nil, manager.WrapReturn(manager.StatusGeneralError, "apt install failed", nil)
		}
	}

	// result.ExitCode == 0: Success - parse results and return
	return parseInstallOutput(string(result.Output)), nil
}

// Remove packages
func (m *Manager) Remove(ctx context.Context, packages []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	if opts == nil {
		opts = manager.DefaultOptions()
	}

	if err := m.ValidatePackageNames(packages); err != nil {
		return nil, err
	}

	m.HandleDryRun(opts, "remove", packages)

	args := []string{"remove", "-y", "--autoremove"}
	args = append(args, packages...)

	if opts.DryRun {
		args = append(args, "--dry-run")
	}

	result, err := m.GetRunner().Run(ctx, "apt", args, "DEBIAN_FRONTEND=noninteractive")
	if err != nil {
		return nil, manager.WrapReturn(manager.StatusUnavailableError, "apt command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		switch result.ExitCode {
		case 1:
			// APT: general error, check stderr for specifics
			stderrStr := string(result.Stderr)
			if strings.Contains(stderrStr, "not installed") {
				return nil, manager.WrapReturn(manager.StatusUnavailableError, "package not installed", nil)
			}
			return nil, manager.WrapReturn(manager.StatusGeneralError, "removal failed", nil)
		case 100:
			// APT: permission or lock error
			return nil, manager.WrapReturn(manager.StatusPermissionError, "removal requires root access or system is locked", nil)
		default:
			return nil, manager.WrapReturn(manager.StatusGeneralError, "apt remove failed", nil)
		}
	}

	// result.ExitCode == 0: Success - parse results and return
	return parseRemoveOutput(string(result.Output)), nil
}

// GetInfo returns detailed package information
func (m *Manager) GetInfo(ctx context.Context, packageName string, opts *manager.Options) (manager.PackageInfo, error) {
	// opts is not used in this function, so no need to handle nil case

	if err := manager.ValidatePackageName(packageName); err != nil {
		return manager.PackageInfo{}, err
	}

	result, err := m.GetRunner().Run(ctx, "apt-cache", []string{"show", packageName}, "DEBIAN_FRONTEND=noninteractive")
	if err != nil {
		return manager.PackageInfo{}, manager.WrapReturn(manager.StatusUnavailableError, "apt-cache command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		switch result.ExitCode {
		case 1:
			// Package not found
			return manager.PackageInfo{}, manager.WrapReturn(manager.StatusUnavailableError, "package not found", nil)
		default:
			return manager.PackageInfo{}, manager.WrapReturn(manager.StatusGeneralError, "apt-cache show failed", nil)
		}
	}

	// result.ExitCode == 0: Success - parse results and return
	pkg := ParsePackageInfo(string(result.Output))
	if pkg.Name == "" {
		return manager.PackageInfo{}, manager.ErrPackageNotFound
	}
	return pkg, nil
}

// Refresh updates package lists
func (m *Manager) Refresh(ctx context.Context, opts *manager.Options) error {
	result, err := m.GetRunner().Run(ctx, "apt", []string{"update"}, "DEBIAN_FRONTEND=noninteractive")
	if err != nil {
		return manager.WrapReturn(manager.StatusUnavailableError, "apt command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		switch result.ExitCode {
		case 100:
			// APT: permission or lock error
			return manager.WrapReturn(manager.StatusPermissionError, "update requires root access or system is locked", nil)
		default:
			return manager.WrapReturn(manager.StatusGeneralError, "apt update failed", nil)
		}
	}

	// result.ExitCode == 0: Success
	return nil
}

// Upgrade packages
func (m *Manager) Upgrade(ctx context.Context, packages []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	if len(packages) > 0 {
		if err := manager.ValidatePackageNames(packages); err != nil {
			return nil, err
		}
	}

	var args []string
	if len(packages) > 0 {
		// Upgrade specific packages
		args = append([]string{"install", "-y"}, packages...)
	} else {
		// Upgrade all packages
		args = []string{"upgrade", "-y"}
	}

	if opts != nil && opts.DryRun {
		args = append(args, "--dry-run")
	}

	result, err := m.GetRunner().Run(ctx, "apt", args, "DEBIAN_FRONTEND=noninteractive")
	if err != nil {
		return nil, manager.WrapReturn(manager.StatusUnavailableError, "apt command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		switch result.ExitCode {
		case 1:
			// APT: general error
			return nil, manager.WrapReturn(manager.StatusGeneralError, "upgrade failed", nil)
		case 100:
			// APT: permission or lock error
			return nil, manager.WrapReturn(manager.StatusPermissionError, "upgrade requires root access or system is locked", nil)
		default:
			return nil, manager.WrapReturn(manager.StatusGeneralError, "apt upgrade failed", nil)
		}
	}

	// result.ExitCode == 0: Success - parse results and return
	return parseInstallOutput(string(result.Output)), nil
}

// Clean removes cached packages
func (m *Manager) Clean(ctx context.Context, opts *manager.Options) error {
	result, err := m.GetRunner().Run(ctx, "apt", []string{"autoclean"}, "DEBIAN_FRONTEND=noninteractive")
	if err != nil {
		return manager.WrapReturn(manager.StatusUnavailableError, "apt command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		switch result.ExitCode {
		case 100:
			// APT: permission or lock error
			return manager.WrapReturn(manager.StatusPermissionError, "clean requires root access or system is locked", nil)
		default:
			return manager.WrapReturn(manager.StatusGeneralError, "apt autoclean failed", nil)
		}
	}

	// result.ExitCode == 0: Success
	return nil
}

// AutoRemove removes orphaned packages
func (m *Manager) AutoRemove(ctx context.Context, opts *manager.Options) ([]manager.PackageInfo, error) {
	args := []string{"autoremove", "-y"}
	if opts != nil && opts.DryRun {
		args = append(args, "--dry-run")
	}

	result, err := m.GetRunner().Run(ctx, "apt", args, "DEBIAN_FRONTEND=noninteractive")
	if err != nil {
		return nil, manager.WrapReturn(manager.StatusUnavailableError, "apt command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		switch result.ExitCode {
		case 100:
			// APT: permission or lock error
			return nil, manager.WrapReturn(manager.StatusPermissionError, "autoremove requires root access or system is locked", nil)
		default:
			return nil, manager.WrapReturn(manager.StatusGeneralError, "apt autoremove failed", nil)
		}
	}

	// result.ExitCode == 0: Success - parse results and return
	// Use appropriate parser based on operation mode
	if opts != nil && opts.DryRun {
		return m.parseAutoRemoveOutput(string(result.Output)), nil
	}
	return parseRemoveOutput(string(result.Output)), nil
}

// Verify checks package integrity (delegated to dpkg)
func (m *Manager) Verify(ctx context.Context, packages []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	if len(packages) == 0 {
		return nil, fmt.Errorf("verify requires specific package names")
	}

	if err := manager.ValidatePackageNames(packages); err != nil {
		return nil, err
	}

	var results []manager.PackageInfo
	for _, pkg := range packages {
		_, err := m.GetRunner().Run(ctx, "dpkg", []string{"-s", pkg})
		status := manager.StatusInstalled
		if err != nil {
			status = "broken"
		}

		results = append(results, manager.NewPackageInfo(pkg, "", status, "apt"))
	}

	return results, nil
}

func (m *Manager) listInstalled(ctx context.Context, _ *manager.Options) ([]manager.PackageInfo, error) {
	result, err := m.GetRunner().Run(ctx, "dpkg-query",
		[]string{"-W", "-f", "${binary:Package} ${Version} ${Architecture}\n"})
	if err != nil {
		return nil, manager.WrapReturn(manager.StatusUnavailableError, "dpkg-query command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		return nil, manager.WrapReturn(manager.StatusGeneralError, "dpkg-query failed", nil)
	}

	// result.ExitCode == 0: Success - parse results and return
	var packages []manager.PackageInfo
	lines := strings.Split(string(result.Output), "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			pkg := manager.NewPackageInfo(parts[0], parts[1], manager.StatusInstalled, "apt")

			if len(parts) >= 3 {
				pkg.Metadata["arch"] = parts[2]
			}

			packages = append(packages, pkg)
		}
	}

	return packages, nil
}

func (m *Manager) listUpgradable(ctx context.Context, _ *manager.Options) ([]manager.PackageInfo, error) {
	result, err := m.GetRunner().Run(ctx, "apt", []string{"list", "--upgradable"})
	if err != nil {
		return nil, manager.WrapReturn(manager.StatusUnavailableError, "apt command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		switch result.ExitCode {
		case 100:
			// APT: permission or cache issue
			return nil, manager.WrapReturn(manager.StatusGeneralError, "apt cache needs update - run 'apt update'", nil)
		default:
			return nil, manager.WrapReturn(manager.StatusGeneralError, "apt list --upgradable failed", nil)
		}
	}

	// result.ExitCode == 0: Success - parse results and return
	var packages []manager.PackageInfo
	lines := strings.Split(string(result.Output), "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "Listing") || strings.TrimSpace(line) == "" {
			continue
		}

		// Parse "package/repo version arch [upgradable from: oldversion]"
		parts := strings.Fields(line)
		if len(parts) >= 4 && strings.Contains(line, "upgradable from") {
			nameRepo := strings.Split(parts[0], "/")
			if len(nameRepo) >= 2 {
				// Extract old version from "[upgradable from: version]"
				oldVersion := ""
				if len(parts) >= 6 {
					oldVersion = strings.TrimSuffix(parts[5], "]")
				}

				pkg := manager.NewPackageInfo(nameRepo[0], oldVersion, manager.StatusUpgradable, "apt")
				pkg.NewVersion = parts[1]
				pkg.Category = nameRepo[1]
				pkg.Metadata["arch"] = parts[2]

				packages = append(packages, pkg)
			}
		}
	}

	return packages, nil
}

// Additional parser methods for comprehensive fixture testing

func (m *Manager) parseAutoRemoveOutput(output string) []manager.PackageInfo {
	var packages []manager.PackageInfo
	lines := strings.Split(output, "\n")

	// Look for "The following packages will be REMOVED:"
	inRemoveSection := false
	for _, line := range lines {
		if strings.Contains(line, "The following packages will be REMOVED:") {
			inRemoveSection = true
			continue
		}

		if inRemoveSection && strings.TrimSpace(line) != "" && !strings.Contains(line, "upgraded") {
			// Parse package names from the removal list
			packageNames := strings.Fields(line)
			for _, name := range packageNames {
				// Clean up package name (remove any special characters)
				cleanName := strings.Trim(name, " \t")
				if cleanName != "" && !strings.Contains(cleanName, "operation") {
					pkg := manager.NewPackageInfo(cleanName, "", manager.StatusAvailable, "apt")
					packages = append(packages, pkg)
				}
			}
		}

		// Stop at summary lines
		if strings.Contains(line, "upgraded") || strings.Contains(line, "After this operation") {
			break
		}
	}

	return packages
}

func (m *Manager) parseCleanOutput(output string) []manager.PackageInfo {
	// Clean operation typically doesn't return package info, just verify no parsing errors
	return []manager.PackageInfo{}
}

func (m *Manager) parseUpdateOutput(output string) []manager.PackageInfo {
	// Update operation typically doesn't return package info, just verify no parsing errors
	return []manager.PackageInfo{}
}

func (m *Manager) parseUpgradeOutput(output string) []manager.PackageInfo {
	var packages []manager.PackageInfo
	lines := strings.Split(output, "\n")

	// Look for "Inst package [currentversion] (newversion ...)" lines in dry run output
	instRegex := regexp.MustCompile(`Inst ([^\s]+) \[([^\]]+)\] \(([^\s]+)`)

	for _, line := range lines {
		if match := instRegex.FindStringSubmatch(line); match != nil {
			name := match[1]
			currentVersion := match[2]
			newVersion := match[3]

			pkg := manager.NewPackageInfo(name, currentVersion, manager.StatusUpgradable, "apt")
			pkg.NewVersion = newVersion

			packages = append(packages, pkg)
		}
	}

	return packages
}

// Status returns overall status/health of the APT package manager
func (m *Manager) Status(ctx context.Context, opts *manager.Options) (manager.ManagerStatus, error) {
	if opts == nil {
		opts = manager.DefaultOptions()
	}

	status := manager.ManagerStatus{
		Available:      m.IsAvailable(),
		Healthy:        true,
		Issues:         []string{},
		Metadata:       make(map[string]interface{}),
		LastRefresh:    "unknown",
		CacheSize:      0,
		PackageCount:   0,
		InstalledCount: 0,
	}

	// Get version
	if version, err := m.GetVersion(); err == nil {
		status.Version = version
	}

	// Check if we can access APT
	if _, err := m.GetRunner().Run(ctx, "apt", []string{"--version"}); err != nil {
		status.Healthy = false
		status.Issues = append(status.Issues, "APT command not accessible")
	}

	// Get installed package count
	if installed, err := m.listInstalled(ctx, opts); err == nil {
		status.InstalledCount = len(installed)
	}

	// Try to get cache information
	if _, err := m.GetRunner().Run(ctx, "apt-cache", []string{"stats"}); err == nil {
		// APT cache is accessible - could parse more detailed info if needed
		status.Metadata["cache_accessible"] = true
	}

	// If not available, mark as unhealthy
	if !status.Available {
		status.Healthy = false
		status.Issues = append(status.Issues, "APT is not available on this system")
	}

	return status, nil
}

// Plugin for registration
type Plugin struct{}

func (p *Plugin) CreateManager() manager.PackageManager { return NewManager() }
func (p *Plugin) GetPriority() int                      { return 90 }

// Auto-register
func init() {
	_ = manager.Register("apt", &Plugin{})
}
