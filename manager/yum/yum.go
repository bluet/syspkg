// Package yum provides an implementation of the syspkg manager interface for the yum package manager.
// It provides a Go (golang) API interface for interacting with the YUM package manager.
// This package is a wrapper around the yum command line tool.
//
// YUM was the default package manager on RedHat-based systems such as CentOS, it has been recently superseded by DNF (Dandified YUM)
//
// Behavior Contracts:
//
// Status Determination:
//   - Find() provides accurate installation status via rpm -q integration
//   - Cross-package manager API consistency with APT implementation
//   - GetPackageInfo() shows "Installed Packages" vs "Available Packages" sections
//
// Field Usage by Operation:
//   - Find: Status=installed/available/upgradable (via rpm -q), Version=installed_version, NewVersion=available_version
//   - ListInstalled: Status=installed, Version=installed_version, NewVersion=""
//   - GetPackageInfo: Status=available/installed (based on section), Version=package_version
//
// Cross-Package Manager Compatibility:
//   - Full API consistency with APT package manager
//   - Normalized status reporting across all package managers
//
// This package is part of the syspkg library.
package yum

import (
	"context"
	"log"
	"os/exec"
	"sync"
	"time"

	"github.com/bluet/syspkg/manager"
)

// Timeouts for different YUM operations
const (
	readTimeout  = 3 * time.Minute // For search, list, info operations
	cleanTimeout = 5 * time.Minute // For clean operations
)

var pm string = "yum"

// Constants used for yum commands
const (
	ArgsAssumeYes    string = "-y"
	ArgsAssumeNo     string = "--assumeno"
	ArgsQuiet        string = "-q"
	ArgsDryRun       string = "--setopt=tsflags=test" // Test transaction without executing
	ArgsFixBroken    string = "check"                 // Check for broken dependencies
	ArgsPurge        string = ""                      // YUM doesn't distinguish remove vs purge
	ArgsAutoRemove   string = "autoremove"            // Remove unneeded dependencies
	ArgsShowProgress string = "-v"                    // Verbose output shows progress
)

// PackageManager implements the manager.PackageManager interface for the yum package manager.
type PackageManager struct {
	// runner is the command execution interface (can be mocked for testing)
	runner manager.CommandRunner
	// runnerOnce protects lazy initialization for zero-value struct usage (e.g., &PackageManager{})
	// This enables defensive programming and backward compatibility with existing test patterns
	runnerOnce sync.Once
}

// NewPackageManager creates a new YUM package manager with default command runner
func NewPackageManager() *PackageManager {
	return &PackageManager{
		runner: manager.NewDefaultCommandRunner(),
	}
}

// NewPackageManagerWithCustomRunner creates a new YUM package manager with custom command runner
// This is primarily used for testing with mocked commands
func NewPackageManagerWithCustomRunner(runner manager.CommandRunner) *PackageManager {
	return &PackageManager{
		runner: runner,
	}
}

// getRunner returns the command runner, creating a default one if not set.
// Uses sync.Once for thread-safe lazy initialization to support zero-value struct usage:
//   - Production: NewPackageManager() pre-initializes runner
//   - Testing: &PackageManager{} uses lazy initialization here
//
// This pattern enables defensive programming and prevents panics on zero-value usage.
func (a *PackageManager) getRunner() manager.CommandRunner {
	a.runnerOnce.Do(func() {
		if a.runner == nil {
			a.runner = manager.NewDefaultCommandRunner()
		}
	})
	return a.runner
}

// executeCommand handles command execution with support for both interactive and non-interactive modes
// For interactive mode, it uses RunInteractive to handle stdin/stdout/stderr
// For non-interactive mode, it uses RunContext for testability
func (a *PackageManager) executeCommand(ctx context.Context, args []string, opts *manager.Options) ([]byte, error) {
	if opts != nil && opts.Interactive {
		// Interactive mode uses RunInteractive for stdin/stdout/stderr handling
		err := a.getRunner().RunInteractive(ctx, pm, args)
		return nil, err
	}

	// Use RunContext for non-interactive execution (automatically includes LC_ALL=C)
	return a.getRunner().RunContext(ctx, pm, args)
}

// IsAvailable checks if the yum package manager is available on the system.
func (a *PackageManager) IsAvailable() bool {
	_, err := exec.LookPath(pm)
	return err == nil
}

// GetPackageManager returns the name of the yum package manager.
func (a *PackageManager) GetPackageManager() string {
	return pm
}

// Install installs the specified packages using the yum package manager.
// Returns PackageInfo for each successfully installed package with Status=installed.
// Version and NewVersion fields will contain the installed version.
//
// Behavior:
//   - Automatically installs dependencies
//   - Returns all installed packages (main packages + dependencies)
//   - Uses -y flag to automatically answer yes to prompts
//   - Respects DryRun, Interactive, and Verbose options
func (a *PackageManager) Install(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	// Validate package names to prevent command injection
	if err := manager.ValidatePackageNames(pkgs); err != nil {
		return nil, err
	}

	if opts == nil {
		opts = &manager.Options{
			DryRun:      false,
			Interactive: false,
			Verbose:     false,
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	args := []string{"install"}

	// Handle options
	if opts.DryRun {
		args = append(args, "--assumeno")
	} else if !opts.Interactive {
		args = append(args, "-y")
	}

	if opts.Verbose {
		args = append(args, "-v")
	}

	args = append(args, pkgs...)

	out, err := a.executeCommand(ctx, args, opts)
	if err != nil {
		return nil, err
	}

	// For interactive mode, we can't parse output, return empty list
	if opts.Interactive {
		return []manager.PackageInfo{}, nil
	}

	if opts.Verbose {
		log.Println(string(out))
	}

	return ParseInstallOutput(string(out), opts), nil
}

// Delete removes the specified packages using the yum package manager.
// Returns PackageInfo for each successfully removed package with Status=available.
// Version field contains the removed version, NewVersion will be empty.
//
// Behavior:
//   - Does not remove dependencies by default (use AutoRemove for that)
//   - Uses -y flag to automatically answer yes to prompts
//   - Respects DryRun, Interactive, and Verbose options
func (a *PackageManager) Delete(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	// Validate package names to prevent command injection
	if err := manager.ValidatePackageNames(pkgs); err != nil {
		return nil, err
	}

	if opts == nil {
		opts = &manager.Options{
			DryRun:      false,
			Interactive: false,
			Verbose:     false,
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	args := []string{"remove"}

	// Handle options
	if opts.DryRun {
		args = append(args, "--assumeno")
	} else if !opts.Interactive {
		args = append(args, "-y")
	}

	if opts.Verbose {
		args = append(args, "-v")
	}

	args = append(args, pkgs...)

	out, err := a.executeCommand(ctx, args, opts)
	if err != nil {
		return nil, err
	}

	// For interactive mode, we can't parse output, return empty list
	if opts.Interactive {
		return []manager.PackageInfo{}, nil
	}

	if opts.Verbose {
		log.Println(string(out))
	}

	return ParseDeleteOutput(string(out), opts), nil
}

// Refresh updates the package list using the yum package manager.
// Uses 'yum clean expire-cache' which efficiently refreshes metadata without
// aggressive cache clearing. This preserves valid cache files while ensuring
// up-to-date repository information.
func (a *PackageManager) Refresh(opts *manager.Options) error {
	if opts == nil {
		opts = &manager.Options{
			DryRun:      false,
			Interactive: false,
			Verbose:     false,
		}
	}

	// Handle dry run mode
	if opts.DryRun {
		log.Println("Dry run mode: would execute 'yum clean expire-cache'")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), cleanTimeout)
	defer cancel()

	args := []string{"clean", "expire-cache"}

	out, err := a.executeCommand(ctx, args, opts)
	if err != nil {
		return err
	}

	if !opts.Interactive && opts.Verbose {
		log.Println(string(out))
	}
	return nil
}

// Find searches for packages matching the provided keywords using the yum package manager.
// Returns packages with accurate installation status detection via rpm -q integration.
//
// Status Detection:
//   - Uses rpm -q to check installation status for each found package
//   - Status=installed: Package is currently installed
//   - Status=available: Package exists in repositories but is not installed
//   - Status=upgradable: Package is installed but newer version is available
//
// Returned fields:
//   - Name: Package name
//   - Arch: Package architecture
//   - Status: Accurate status via rpm -q integration (installed/available/upgradable)
//   - Version: Installed version (if installed) or empty
//   - NewVersion: Available version from repositories
func (a *PackageManager) Find(keywords []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	// Validate keywords to prevent command injection
	if err := manager.ValidatePackageNames(keywords); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	defer cancel()

	args := append([]string{"search"}, keywords...)

	// Use CommandRunner for search operation (automatically includes LC_ALL=C)
	out, err := a.getRunner().RunContext(ctx, pm, args)
	if err != nil {
		return nil, err
	}

	// Parse the search output to get basic package info
	packages := ParseFindOutput(string(out), opts)

	// Enhance with accurate status information using rpm -q
	// This provides cross-package manager API consistency
	enhancedPackages, err := a.enhancePackagesWithStatus(packages, opts)
	if err != nil {
		// If status enhancement fails, return basic packages (backward compatibility)
		return packages, nil
	}

	return enhancedPackages, nil
}

// ListInstalled lists all installed packages using the yum package manager.
func (a *PackageManager) ListInstalled(opts *manager.Options) ([]manager.PackageInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	defer cancel()

	args := []string{"list", "--installed"}
	// Use CommandRunner for list operation (automatically includes LC_ALL=C)
	out, err := a.getRunner().RunContext(ctx, pm, args)
	if err != nil {
		return nil, err
	}
	return ParseListInstalledOutput(string(out), opts), nil
}

// ListUpgradable lists all packages that have newer versions available.
// Returns packages with Status=upgradable, Version=current, NewVersion=available.
//
// Uses 'yum check-update' which returns exit code 100 when updates are available,
// exit code 0 when no updates are available.
func (a *PackageManager) ListUpgradable(opts *manager.Options) ([]manager.PackageInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	defer cancel()

	args := []string{"check-update"}

	// Use CommandRunner for check-update operation (automatically includes LC_ALL=C)
	out, err := a.getRunner().RunContext(ctx, pm, args)
	// YUM check-update returns exit code 100 when updates are available
	// This is normal behavior, not an error
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 100 {
			// Exit code 100 means updates are available, continue parsing
		} else {
			// Other exit codes indicate real errors
			return nil, err
		}
	}

	return ParseListUpgradableOutput(string(out), opts), nil
}

// Upgrade upgrades the specified packages using the yum package manager.
// Returns PackageInfo for each successfully upgraded package with new version information.
func (a *PackageManager) Upgrade(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	// Validate package names to prevent command injection
	if len(pkgs) > 0 {
		if err := manager.ValidatePackageNames(pkgs); err != nil {
			return nil, err
		}
	}

	if opts == nil {
		opts = &manager.Options{
			DryRun:      false,
			Interactive: false,
			Verbose:     false,
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	args := []string{"update"}

	// Handle options
	if opts.DryRun {
		args = append(args, "--assumeno")
	} else if !opts.Interactive {
		args = append(args, "-y")
	}

	if opts.Verbose {
		args = append(args, "-v")
	}

	args = append(args, pkgs...)

	out, err := a.executeCommand(ctx, args, opts)
	if err != nil {
		return nil, err
	}

	// For interactive mode, we can't parse output, return empty list
	if opts.Interactive {
		return []manager.PackageInfo{}, nil
	}

	if opts.Verbose {
		log.Println(string(out))
	}

	return ParseUpgradeOutput(string(out), opts), nil
}

// UpgradeAll upgrades all packages that have newer versions available.
// Returns PackageInfo for each upgraded package with new version information.
func (a *PackageManager) UpgradeAll(opts *manager.Options) ([]manager.PackageInfo, error) {
	if opts == nil {
		opts = &manager.Options{
			DryRun:      false,
			Interactive: false,
			Verbose:     false,
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	args := []string{"update"}

	// Handle options
	if opts.DryRun {
		args = append(args, "--assumeno")
	} else if !opts.Interactive {
		args = append(args, "-y")
	}

	if opts.Verbose {
		args = append(args, "-v")
	}

	out, err := a.executeCommand(ctx, args, opts)
	if err != nil {
		return nil, err
	}

	// For interactive mode, we can't parse output, return empty list
	if opts.Interactive {
		return []manager.PackageInfo{}, nil
	}

	if opts.Verbose {
		log.Println(string(out))
	}

	return ParseUpgradeOutput(string(out), opts), nil
}

// Clean performs comprehensive cleanup of YUM caches.
// Uses 'yum clean all' which removes all cached packages, metadata, and headers.
// This is what administrators typically expect from a clean operation.
func (a *PackageManager) Clean(opts *manager.Options) error {
	if opts == nil {
		opts = &manager.Options{
			DryRun:      false,
			Interactive: false,
			Verbose:     false,
		}
	}

	// Handle dry run mode
	if opts.DryRun {
		log.Println("Dry run mode: would execute 'yum clean all'")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), cleanTimeout)
	defer cancel()

	args := []string{"clean", "all"}

	out, err := a.executeCommand(ctx, args, opts)
	if err != nil {
		return err
	}

	if !opts.Interactive && opts.Verbose {
		log.Println(string(out))
	}
	return nil
}

// GetPackageInfo retrieves package information for the specified package using the yum package manager.
//
// This method can determine accurate installation status by checking whether the package
// appears under "Installed Packages" or "Available Packages" in the yum info output.
//
// Returned fields:
//   - Name: Package name
//   - Version: Package version
//   - Arch: Package architecture
//   - Status: PackageStatusInstalled if under "Installed Packages" section,
//     PackageStatusAvailable if under "Available Packages" section
//   - PackageManager: "yum"
func (a *PackageManager) GetPackageInfo(pkg string, opts *manager.Options) (manager.PackageInfo, error) {
	// Validate package name to prevent command injection
	if err := manager.ValidatePackageName(pkg); err != nil {
		return manager.PackageInfo{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	defer cancel()

	// Use CommandRunner for package info query (automatically includes LC_ALL=C)
	out, err := a.getRunner().RunContext(ctx, pm, []string{"info", pkg})
	if err != nil {
		return manager.PackageInfo{}, err
	}
	return ParsePackageInfoOutput(string(out), opts), nil
}

// AutoRemove removes unneeded dependencies using the yum package manager.
// Returns PackageInfo for each successfully removed package with Status=available.
func (a *PackageManager) AutoRemove(opts *manager.Options) ([]manager.PackageInfo, error) {
	if opts == nil {
		opts = &manager.Options{
			DryRun:      false,
			Interactive: false,
			Verbose:     false,
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	args := []string{"autoremove"}

	// Handle options
	if opts.DryRun {
		args = append(args, "--assumeno")
	} else if !opts.Interactive {
		args = append(args, "-y")
	}

	if opts.Verbose {
		args = append(args, "-v")
	}

	out, err := a.executeCommand(ctx, args, opts)
	if err != nil {
		return nil, err
	}

	// For interactive mode, we can't parse output, return empty list
	if opts.Interactive {
		return []manager.PackageInfo{}, nil
	}

	if opts.Verbose {
		log.Println(string(out))
	}

	return ParseAutoRemoveOutput(string(out), opts), nil
}

// enhancePackagesWithStatus adds accurate installation status to packages using rpm -q
// This method performs the system calls that were previously embedded in parsing functions
func (a *PackageManager) enhancePackagesWithStatus(packages []manager.PackageInfo, _ *manager.Options) ([]manager.PackageInfo, error) {
	if len(packages) == 0 {
		return packages, nil
	}

	// Build map for faster lookup
	packageMap := make(map[string]manager.PackageInfo)
	packageNames := make([]string, 0, len(packages))
	for _, pkg := range packages {
		packageMap[pkg.Name] = pkg
		packageNames = append(packageNames, pkg.Name)
	}

	// Check installation status using rpm -q
	installedPackages, err := a.checkRpmInstallationStatus(packageNames)
	if err != nil {
		return nil, err
	}

	// Build result list with updated status
	result := make([]manager.PackageInfo, 0, len(packages))
	for _, pkg := range packages {
		// Check if package is installed
		if installedInfo, isInstalled := installedPackages[pkg.Name]; isInstalled {
			pkg.Status = manager.PackageStatusInstalled
			pkg.Version = installedInfo.Version
		} else {
			pkg.Status = manager.PackageStatusAvailable
			pkg.Version = ""
		}
		result = append(result, pkg)
	}

	return result, nil
}
