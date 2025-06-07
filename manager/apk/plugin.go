// Package apk provides a complete APK package manager implementation
package apk

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/bluet/syspkg/manager"
)

// Manager implements the unified PackageManager interface for APK
type Manager struct {
	*manager.BaseManager
}

// NewManager creates a new APK manager
func NewManager() *Manager {
	runner := manager.NewDefaultCommandRunner()
	return &Manager{
		BaseManager: manager.NewBaseManager("apk", manager.CategorySystem, runner),
	}
}

// NewManagerWithRunner creates APK manager with custom runner (for testing)
func NewManagerWithRunner(runner manager.CommandRunner) *Manager {
	return &Manager{
		BaseManager: manager.NewBaseManager("apk", manager.CategorySystem, runner),
	}
}

// IsAvailable checks if APK is available
func (m *Manager) IsAvailable() bool {
	_, err := exec.LookPath("apk")
	if err != nil {
		return false
	}
	// Verify apk is working
	_, err = m.GetRunner().Run(context.Background(), "apk", []string{"--version"})
	return err == nil
}

// GetVersion returns APK version
func (m *Manager) GetVersion() (string, error) {
	result, err := m.GetRunner().Run(context.Background(), "apk", []string{"--version"})
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(result.Output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "apk-tools ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
	}
	return strings.TrimSpace(string(result.Output)), nil
}

// Search searches for packages
func (m *Manager) Search(ctx context.Context, query []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	if len(query) == 0 {
		return nil, fmt.Errorf("search requires a query")
	}

	if err := m.ValidatePackageNames(query); err != nil {
		return nil, err
	}

	args := []string{"search"}
	args = append(args, query...)

	result, err := m.GetRunner().Run(ctx, "apk", args)
	if err != nil {
		// Command execution failed (e.g., apk not found)
		return nil, manager.WrapReturn(manager.StatusUnavailableError, "apk command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		switch result.ExitCode {
		case 1:
			// No packages found - that's okay for search
			return []manager.PackageInfo{}, nil
		case 2:
			// Invalid usage
			return nil, manager.WrapReturn(manager.StatusUsageError, "invalid search query", nil)
		default:
			// Unknown exit code
			return nil, manager.WrapReturn(manager.StatusGeneralError, "apk search failed", nil)
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
	case manager.FilterUpgradable:
		return m.listUpgradable(ctx, opts)
	case manager.FilterAvailable:
		return nil, fmt.Errorf("listing all available packages not supported - use Search instead")
	default:
		return nil, fmt.Errorf("unsupported filter: %s", filter)
	}
}

// GetInfo gets information about a specific package
func (m *Manager) GetInfo(ctx context.Context, packageName string, opts *manager.Options) (manager.PackageInfo, error) {
	if err := m.ValidatePackageNames([]string{packageName}); err != nil {
		return manager.PackageInfo{}, err
	}

	output, err := m.GetRunner().Run(ctx, "apk", []string{"info", packageName})
	if err != nil {
		return manager.PackageInfo{}, manager.WrapCommandError("apk info failed", err)
	}

	return m.parseInfoOutput(string(output.Output), packageName), nil
}

// Status returns package manager status
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

	// Check if we can access APK
	if _, err := m.GetRunner().Run(ctx, "apk", []string{"--version"}); err != nil {
		status.Healthy = false
		status.Issues = append(status.Issues, "APK command not accessible")
	}

	// Get installed package count
	if installed, err := m.listInstalled(ctx, opts); err == nil {
		status.InstalledCount = len(installed)
	}

	// If not available, mark as unhealthy
	if !status.Available {
		status.Healthy = false
		status.Issues = append(status.Issues, "APK is not available on this system")
	}

	return status, nil
}

// Install installs packages
func (m *Manager) Install(ctx context.Context, packages []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	// Plugin developer knows this is a usage error
	if len(packages) == 0 {
		return nil, manager.WrapReturn(manager.StatusUsageError, "install requires package names", nil)
	}

	if err := m.ValidatePackageNames(packages); err != nil {
		return nil, manager.WrapReturn(manager.StatusUsageError, "invalid package names", err)
	}

	args := []string{"add"}
	if opts != nil && opts.DryRun {
		args = append(args, "--simulate")
	}
	args = append(args, packages...)

	result, err := m.GetRunner().Run(ctx, "apk", args)
	if err != nil {
		// Command execution failed (e.g., apk not found)
		return nil, manager.WrapReturn(manager.StatusUnavailableError, "apk command failed", err)
	}

	// Handle error cases first (return-early pattern)
	if result.ExitCode != 0 {
		switch result.ExitCode {
		case 1:
			// Check stderr to determine specific error
			stderrStr := string(result.Stderr)
			if strings.Contains(stderrStr, "not found") || strings.Contains(stderrStr, "No such package") {
				return nil, manager.WrapReturn(manager.StatusUnavailableError, "package not found in repository", nil)
			}
			return nil, manager.WrapReturn(manager.StatusGeneralError, "installation failed", nil)
		case 77:
			// Permission denied (APK uses this exit code)
			return nil, manager.WrapReturn(manager.StatusPermissionError, "installation requires root access", nil)
		default:
			// Unknown exit code
			return nil, manager.WrapReturn(manager.StatusGeneralError, "apk install failed", nil)
		}
	}

	// result.ExitCode == 0: Success - parse results and return
	return m.parseInstallOutput(string(result.Output)), nil
}

// Remove removes packages
func (m *Manager) Remove(ctx context.Context, packages []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	if len(packages) == 0 {
		return nil, fmt.Errorf("remove requires package names")
	}

	if err := m.ValidatePackageNames(packages); err != nil {
		return nil, err
	}

	args := []string{"del"}
	if opts != nil && opts.DryRun {
		args = append(args, "--simulate")
	}
	args = append(args, packages...)

	output, err := m.GetRunner().Run(ctx, "apk", args)
	if err != nil {
		return nil, manager.WrapCommandError("apk del failed", err)
	}

	return m.parseRemoveOutput(string(output.Output)), nil
}

// Refresh refreshes package lists (equivalent to update)
func (m *Manager) Refresh(ctx context.Context, opts *manager.Options) error {
	return m.Update(ctx, opts)
}

// Update updates package lists
func (m *Manager) Update(ctx context.Context, opts *manager.Options) error {
	_, err := m.GetRunner().Run(ctx, "apk", []string{"update"})
	if err != nil {
		return manager.WrapCommandError("apk update failed", err)
	}
	return nil
}

// Upgrade upgrades packages
func (m *Manager) Upgrade(ctx context.Context, packages []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	args := []string{"upgrade"}

	if len(packages) > 0 {
		if err := m.ValidatePackageNames(packages); err != nil {
			return nil, err
		}
		args = append(args, packages...)
	}

	if opts != nil && opts.DryRun {
		args = append(args, "--simulate")
	}

	output, err := m.GetRunner().Run(ctx, "apk", args)
	if err != nil {
		return nil, manager.WrapCommandError("apk upgrade failed", err)
	}

	return m.parseUpgradeOutput(string(output.Output)), nil
}

// Clean cleans package cache
func (m *Manager) Clean(ctx context.Context, opts *manager.Options) error {
	_, err := m.GetRunner().Run(ctx, "apk", []string{"cache", "clean"})
	if err != nil {
		return manager.WrapCommandError("apk cache clean failed", err)
	}
	return nil
}

// AutoRemove is not applicable for APK (no orphaned packages concept)
func (m *Manager) AutoRemove(ctx context.Context, opts *manager.Options) ([]manager.PackageInfo, error) {
	// APK doesn't have orphaned packages like APT
	return []manager.PackageInfo{}, nil
}

// Verify checks package integrity
func (m *Manager) Verify(ctx context.Context, packages []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	if len(packages) == 0 {
		return nil, fmt.Errorf("verify requires specific package names")
	}

	if err := m.ValidatePackageNames(packages); err != nil {
		return nil, err
	}

	var results []manager.PackageInfo
	for _, pkg := range packages {
		// Check if package is installed using apk info
		_, err := m.GetRunner().Run(ctx, "apk", []string{"info", "-e", pkg})
		status := manager.StatusInstalled
		if err != nil {
			status = "not-installed"
		}

		results = append(results, manager.NewPackageInfo(pkg, "", status, "apk"))
	}

	return results, nil
}

// parseSearchOutput parses apk search output
func (m *Manager) parseSearchOutput(output string) []manager.PackageInfo {
	var packages []manager.PackageInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Format: package-name-version
		pkg := manager.NewPackageInfo(line, "", manager.StatusAvailable, "apk")
		packages = append(packages, pkg)
	}

	return packages
}

// parseInfoOutput parses apk info output
func (m *Manager) parseInfoOutput(output, packageName string) manager.PackageInfo {
	pkg := manager.NewPackageInfo(packageName, "", manager.StatusAvailable, "apk")

	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if strings.Contains(line, " description:") {
			// Extract package name and version from the description line
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				nameVersion := parts[0]
				if strings.Contains(nameVersion, "-") {
					// Extract version from package-name-version format
					lastDash := strings.LastIndex(nameVersion, "-")
					if lastDash > 0 {
						pkg.Name = nameVersion[:lastDash]
						pkg.Version = nameVersion[lastDash+1:]
					}
				}
			}

			// Next line contains the description
			if i+1 < len(lines) {
				pkg.Description = strings.TrimSpace(lines[i+1])
			}
		}

		if strings.Contains(line, " webpage:") && i+1 < len(lines) {
			pkg.Metadata["webpage"] = strings.TrimSpace(lines[i+1])
		}

		if strings.Contains(line, " installed size:") && i+1 < len(lines) {
			pkg.Metadata["installed_size"] = strings.TrimSpace(lines[i+1])
		}
	}

	return pkg
}

// listInstalled lists installed apk packages
func (m *Manager) listInstalled(ctx context.Context, _ *manager.Options) ([]manager.PackageInfo, error) {
	output, err := m.GetRunner().Run(ctx, "apk", []string{"info", "-v"})
	if err != nil {
		return nil, manager.WrapCommandError("apk info -v failed", err)
	}

	var packages []manager.PackageInfo
	lines := strings.Split(string(output.Output), "\n")

	// Regex to parse: package-version arch {origin} (license) [installed]
	installedRegex := regexp.MustCompile(`^([^\s]+) ([^\s]+) \{([^}]+)\} \(([^)]+)\) \[installed\]`)

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		if match := installedRegex.FindStringSubmatch(line); match != nil {
			nameVersion := match[1]
			arch := match[2]
			origin := match[3]
			license := match[4]

			// Extract name and version
			var name, version string
			if strings.Contains(nameVersion, "-") {
				lastDash := strings.LastIndex(nameVersion, "-")
				if lastDash > 0 {
					name = nameVersion[:lastDash]
					version = nameVersion[lastDash+1:]
				}
			} else {
				name = nameVersion
			}

			pkg := manager.NewPackageInfo(name, version, manager.StatusInstalled, "apk")
			pkg.Metadata["arch"] = arch
			pkg.Metadata["origin"] = origin
			pkg.Metadata["license"] = license

			packages = append(packages, pkg)
		}
	}

	return packages, nil
}

// listUpgradable lists packages that can be upgraded
func (m *Manager) listUpgradable(ctx context.Context, _ *manager.Options) ([]manager.PackageInfo, error) {
	// APK doesn't have a direct "list upgradable" command
	// We would need to simulate upgrade to see what would be upgraded
	output, err := m.GetRunner().Run(ctx, "apk", []string{"upgrade", "--simulate"})
	if err != nil {
		return nil, manager.WrapCommandError("apk upgrade --simulate failed", err)
	}

	return m.parseUpgradeOutput(string(output.Output)), nil
}

// parseInstallOutput parses the output of apk add
func (m *Manager) parseInstallOutput(output string) []manager.PackageInfo {
	var packages []manager.PackageInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.Contains(line, "Installing") || strings.Contains(line, "Upgrading") {
			// Extract package names from installation lines
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.Contains(part, "-") && !strings.HasPrefix(part, "-") {
					// Looks like a package name
					pkg := manager.NewPackageInfo(part, "", manager.StatusInstalled, "apk")
					packages = append(packages, pkg)
				}
			}
		}
	}

	return packages
}

// parseRemoveOutput parses the output of apk del
func (m *Manager) parseRemoveOutput(output string) []manager.PackageInfo {
	var packages []manager.PackageInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.Contains(line, "Purging") {
			// Extract package names from removal lines
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.Contains(part, "-") && !strings.HasPrefix(part, "-") {
					// Looks like a package name
					pkg := manager.NewPackageInfo(part, "", "removed", "apk")
					packages = append(packages, pkg)
				}
			}
		}
	}

	return packages
}

// parseUpgradeOutput parses the output of apk upgrade
func (m *Manager) parseUpgradeOutput(output string) []manager.PackageInfo {
	var packages []manager.PackageInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.Contains(line, "Upgrading") {
			// Extract package names from upgrade lines
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.Contains(part, "-") && !strings.HasPrefix(part, "-") {
					// Looks like a package name
					pkg := manager.NewPackageInfo(part, "", "upgraded", "apk")
					packages = append(packages, pkg)
				}
			}
		}
	}

	return packages
}

// Plugin for registration
type Plugin struct{}

func (p *Plugin) CreateManager() manager.PackageManager { return NewManager() }
func (p *Plugin) GetPriority() int                      { return 60 } // Lower priority than APT/YUM

// Auto-register
func init() {
	_ = manager.Register("apk", &Plugin{})
}
