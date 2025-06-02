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

	return m.parseSearchOutput(string(output)), nil
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

	return m.parseInstallOutput(string(output)), nil
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

	return m.parseRemoveOutput(string(output)), nil
}

// GetInfo returns detailed package information
func (m *Manager) GetInfo(ctx context.Context, packageName string, opts *manager.Options) (manager.PackageInfo, error) {
	if opts == nil {
		opts = manager.DefaultOptions()
	}

	if err := manager.ValidatePackageName(packageName); err != nil {
		return manager.PackageInfo{}, err
	}

	output, err := m.GetRunner().RunContext(ctx, "apt-cache", []string{"show", packageName}, "DEBIAN_FRONTEND=noninteractive")
	if err != nil {
		return manager.PackageInfo{}, fmt.Errorf("apt-cache show failed: %w", err)
	}

	return m.parsePackageInfo(string(output)), nil
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

	return m.parseInstallOutput(string(output)), nil
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

	return m.parseRemoveOutput(string(output)), nil
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

// Simple parsing - real implementation would be more robust
func (m *Manager) parseSearchOutput(output string) []manager.PackageInfo {
	var packages []manager.PackageInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.Contains(line, "/") && !strings.HasPrefix(line, "Sorting") && !strings.HasPrefix(line, "Full Text") && !strings.HasPrefix(line, "WARNING") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				nameRepo := strings.Split(parts[0], "/")
				if len(nameRepo) >= 2 {
					pkg := manager.NewPackageInfo(nameRepo[0], "", manager.StatusAvailable, manager.TypeSystem)
					pkg.NewVersion = parts[1]
					pkg.Category = nameRepo[1]
					pkg.Metadata["arch"] = parts[2]
					packages = append(packages, pkg)
				}
			}
		}
	}

	return packages
}

func (m *Manager) parseInstallOutput(output string) []manager.PackageInfo {
	var packages []manager.PackageInfo
	lines := strings.Split(output, "\n")

	// Look for "Setting up package:arch (version) ..." lines
	settingUpRegex := regexp.MustCompile(`Setting up ([^:]+)(?::([^:]+))? \(([^)]+)\)`)

	for _, line := range lines {
		if match := settingUpRegex.FindStringSubmatch(line); match != nil {
			name := match[1]
			arch := match[2]
			version := match[3]

			pkg := manager.NewPackageInfo(name, version, manager.StatusInstalled, manager.TypeSystem)

			if arch != "" {
				pkg.Metadata["arch"] = arch
			}

			packages = append(packages, pkg)
		}
	}

	return packages
}

func (m *Manager) parseRemoveOutput(output string) []manager.PackageInfo {
	var packages []manager.PackageInfo
	lines := strings.Split(output, "\n")

	// Look for "Removing package:arch (version) ..." lines
	removingRegex := regexp.MustCompile(`Removing ([^:]+)(?::([^:]+))? \(([^)]+)\)`)

	for _, line := range lines {
		if match := removingRegex.FindStringSubmatch(line); match != nil {
			name := match[1]
			arch := match[2]
			version := match[3]

			pkg := manager.NewPackageInfo(name, version, manager.StatusAvailable, manager.TypeSystem)

			if arch != "" {
				pkg.Metadata["arch"] = arch
			}

			packages = append(packages, pkg)
		}
	}

	return packages
}

func (m *Manager) parsePackageInfo(output string) manager.PackageInfo {
	pkg := manager.NewPackageInfo("", "", manager.StatusUnknown, manager.TypeSystem)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				switch key {
				case "Package":
					pkg.Name = value
				case "Version":
					pkg.Version = value
				case "Architecture":
					pkg.Metadata["arch"] = value
				case "Section":
					pkg.Category = value
				case "Description":
					pkg.Description = value
				case "Installed-Size":
					pkg.Metadata["installed_size"] = value
				case "Maintainer":
					pkg.Metadata["maintainer"] = value
				}
			}
		}
	}

	// Determine status based on whether we have version info
	if pkg.Version != "" {
		pkg.Status = manager.StatusInstalled
	} else {
		pkg.Status = manager.StatusAvailable
	}

	return pkg
}

func (m *Manager) listInstalled(ctx context.Context, opts *manager.Options) ([]manager.PackageInfo, error) {
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

func (m *Manager) listUpgradable(ctx context.Context, opts *manager.Options) ([]manager.PackageInfo, error) {
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

// Plugin for registration
type Plugin struct{}

func (p *Plugin) CreateManager() manager.PackageManager { return NewManager() }
func (p *Plugin) GetPriority() int                      { return 90 }

// Auto-register
func init() {
	manager.Register("apt", &Plugin{})
}
