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
		BaseManager: manager.NewBaseManager("apt", manager.TypeSystem, runner),
	}
}

// NewManagerWithRunner creates APT manager with custom runner (for testing)
func NewManagerWithRunner(runner manager.CommandRunner) *Manager {
	return &Manager{
		BaseManager: manager.NewBaseManager("apt", manager.TypeSystem, runner),
	}
}

// IsAvailable checks if APT is available
func (m *Manager) IsAvailable() bool {
	_, err := exec.LookPath("apt")
	if err != nil {
		return false
	}
	// Verify it's Debian apt (not Java apt)
	output, err := m.GetRunner().Run("apt", "--version")
	return err == nil && strings.Contains(string(output), "apt") && !strings.Contains(string(output), "java")
}

// GetVersion returns APT version
func (m *Manager) GetVersion() (string, error) {
	output, err := m.GetRunner().Run("apt", "--version")
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		parts := strings.Fields(lines[0])
		if len(parts) >= 2 {
			return parts[1], nil
		}
	}
	return strings.TrimSpace(string(output)), nil
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
	output, err := m.GetRunner().RunContext(ctx, "apt", args)
	if err != nil {
		return nil, err
	}

	// apt search already provides basic status information
	packages := parseSearchOutput(string(output))

	// If ShowStatus is requested, enhance with more detailed status info
	// This is useful for getting exact version details and upgradability
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

	output, err := m.GetRunner().RunContext(ctx, "apt", args, "DEBIAN_FRONTEND=noninteractive")
	if err != nil {
		return nil, fmt.Errorf("apt install failed: %w", err)
	}

	return parseInstallOutput(string(output)), nil
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

	output, err := m.GetRunner().RunContext(ctx, "apt", args, "DEBIAN_FRONTEND=noninteractive")
	if err != nil {
		return nil, fmt.Errorf("apt remove failed: %w", err)
	}

	return parseRemoveOutput(string(output)), nil
}

// GetInfo returns detailed package information
func (m *Manager) GetInfo(ctx context.Context, packageName string, opts *manager.Options) (manager.PackageInfo, error) {
	// opts is not used in this function, so no need to handle nil case

	if err := manager.ValidatePackageName(packageName); err != nil {
		return manager.PackageInfo{}, err
	}

	output, err := m.GetRunner().RunContext(ctx, "apt-cache", []string{"show", packageName}, "DEBIAN_FRONTEND=noninteractive")
	if err != nil {
		return manager.PackageInfo{}, fmt.Errorf("apt-cache show failed: %w", err)
	}

	pkg := ParsePackageInfo(string(output))
	if pkg.Name == "" {
		return manager.PackageInfo{}, manager.ErrPackageNotFound
	}
	return pkg, nil
}

// Refresh updates package lists
func (m *Manager) Refresh(ctx context.Context, opts *manager.Options) error {
	_, err := m.GetRunner().RunContext(ctx, "apt", []string{"update"}, "DEBIAN_FRONTEND=noninteractive")
	if err != nil {
		return fmt.Errorf("apt update failed: %w", err)
	}
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

	output, err := m.GetRunner().RunContext(ctx, "apt", args, "DEBIAN_FRONTEND=noninteractive")
	if err != nil {
		return nil, fmt.Errorf("apt upgrade failed: %w", err)
	}

	return parseInstallOutput(string(output)), nil
}

// Clean removes cached packages
func (m *Manager) Clean(ctx context.Context, opts *manager.Options) error {
	_, err := m.GetRunner().RunContext(ctx, "apt", []string{"autoclean"}, "DEBIAN_FRONTEND=noninteractive")
	if err != nil {
		return fmt.Errorf("apt autoclean failed: %w", err)
	}
	return nil
}

// AutoRemove removes orphaned packages
func (m *Manager) AutoRemove(ctx context.Context, opts *manager.Options) ([]manager.PackageInfo, error) {
	args := []string{"autoremove", "-y"}
	if opts != nil && opts.DryRun {
		args = append(args, "--dry-run")
	}

	output, err := m.GetRunner().RunContext(ctx, "apt", args, "DEBIAN_FRONTEND=noninteractive")
	if err != nil {
		return nil, fmt.Errorf("apt autoremove failed: %w", err)
	}

	// Use appropriate parser based on operation mode
	if opts != nil && opts.DryRun {
		return m.parseAutoRemoveOutput(string(output)), nil
	}
	return parseRemoveOutput(string(output)), nil
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
		_, err := m.GetRunner().RunContext(ctx, "dpkg", []string{"-s", pkg})
		status := manager.StatusInstalled
		if err != nil {
			status = "broken"
		}

		results = append(results, manager.NewPackageInfo(pkg, "", status, manager.TypeSystem))
	}

	return results, nil
}

func (m *Manager) listInstalled(ctx context.Context, _ *manager.Options) ([]manager.PackageInfo, error) {
	output, err := m.GetRunner().RunContext(ctx, "dpkg-query",
		[]string{"-W", "-f", "${binary:Package} ${Version} ${Architecture}\n"})
	if err != nil {
		return nil, fmt.Errorf("dpkg-query failed: %w", err)
	}

	var packages []manager.PackageInfo
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			pkg := manager.NewPackageInfo(parts[0], parts[1], manager.StatusInstalled, manager.TypeSystem)

			if len(parts) >= 3 {
				pkg.Metadata["arch"] = parts[2]
			}

			packages = append(packages, pkg)
		}
	}

	return packages, nil
}

func (m *Manager) listUpgradable(ctx context.Context, _ *manager.Options) ([]manager.PackageInfo, error) {
	output, err := m.GetRunner().RunContext(ctx, "apt", []string{"list", "--upgradable"})
	if err != nil {
		return nil, fmt.Errorf("apt list --upgradable failed: %w", err)
	}

	var packages []manager.PackageInfo
	lines := strings.Split(string(output), "\n")

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

				pkg := manager.NewPackageInfo(nameRepo[0], oldVersion, manager.StatusUpgradable, manager.TypeSystem)
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
					pkg := manager.NewPackageInfo(cleanName, "", manager.StatusAvailable, manager.TypeSystem)
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

			pkg := manager.NewPackageInfo(name, currentVersion, manager.StatusUpgradable, manager.TypeSystem)
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
	if _, err := m.GetRunner().RunContext(ctx, "apt", []string{"--version"}); err != nil {
		status.Healthy = false
		status.Issues = append(status.Issues, "APT command not accessible")
	}

	// Get installed package count
	if installed, err := m.listInstalled(ctx, opts); err == nil {
		status.InstalledCount = len(installed)
	}

	// Try to get cache information
	if _, err := m.GetRunner().RunContext(ctx, "apt-cache", []string{"stats"}); err == nil {
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
