// Package apt provides an implementation of the syspkg manager interface for the apt package manager.
// It provides an Go (golang) API interface for interacting with the APT package manager.
// This package is a wrapper around the apt command line tool.
//
// APT is the default package manager on Debian-based systems such as Ubuntu, it's a set of core tools inside the Debian package management system.
// APT simplifies the process of managing software on Unix-like computer systems by automating the retrieval, configuration and installation of software packages, either from precompiled files or by compiling source code.
// APT was originally designed as a front-end for dpkg to work with Debian's .deb packages, but it has since been modified to also work with the RPM Package Manager system via APT-RPM.
// The Fink project has ported APT to Mac OS X for some of its own package management tasks.
// APT is also the upstream for Aptitude, another Debian package manager.
//
// For more information about apt, visit:
// - https://wiki.debian.org/Apt
// - https://ubuntu.com/server/docs/package-management
// This package is part of the syspkg library.
package apt

import (
	"context"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"

	// "github.com/rs/zerolog"
	// "github.com/rs/zerolog/log"

	"github.com/bluet/syspkg/manager"
)

var pm string = "apt"

// Constants used for apt commands
const (
	ArgsAssumeYes    string = "-y"
	ArgsAssumeNo     string = "--assume-no"
	ArgsDryRun       string = "--dry-run"
	ArgsFixBroken    string = "-f"
	ArgsQuiet        string = "-qq"
	ArgsPurge        string = "--purge"
	ArgsAutoRemove   string = "--autoremove"
	ArgsShowProgress string = "--show-progress"

	// dpkgQueryCmd is the command used to query package information
	dpkgQueryCmd string = "dpkg-query"
)

// Environment variables for non-interactive mode
var (
	aptNonInteractiveEnv = []string{
		"DEBIAN_FRONTEND=noninteractive",
		"DEBCONF_NONINTERACTIVE_SEEN=true",
	}
)

// NOTE: Environment variables for non-interactive mode are now handled automatically by CommandRunner
// LC_ALL=C is set automatically, and DEBIAN_FRONTEND=noninteractive, DEBCONF_NONINTERACTIVE_SEEN=true
// are passed as additional environment variables to each RunContext/RunInteractive call

// PackageManager implements the manager.PackageManager interface for the apt package manager.
type PackageManager struct {
	// runner is the command execution interface (can be mocked for testing)
	runner manager.CommandRunner
	// runnerOnce protects lazy initialization for zero-value struct usage (e.g., &PackageManager{})
	// This enables defensive programming and backward compatibility with existing test patterns
	runnerOnce sync.Once
	// binaryName is the name of the binary to use (e.g., "apt", "apt-fast")
	// Defaults to "apt" if not specified
	binaryName string
	// binaryOnce protects lazy initialization of binaryName
	binaryOnce sync.Once
}

// NewPackageManager creates a new APT package manager with default command runner
func NewPackageManager() *PackageManager {
	return &PackageManager{
		runner: manager.NewDefaultCommandRunner(),
	}
}

// NewPackageManagerWithCustomRunner creates a new APT package manager with custom command runner
// This is primarily used for testing with mocked commands
func NewPackageManagerWithCustomRunner(runner manager.CommandRunner) *PackageManager {
	return &PackageManager{
		runner:     runner,
		binaryName: pm,
	}
}

// NewPackageManagerWithBinary creates a new APT package manager with a custom binary name
// This allows using apt-compatible binaries like apt-fast as a drop-in replacement
func NewPackageManagerWithBinary(binaryName string) *PackageManager {
	return &PackageManager{
		runner:     manager.NewDefaultCommandRunner(),
		binaryName: binaryName,
	}
}

// getBinaryName returns the binary name, defaulting to "apt" if not set.
// Uses sync.Once for thread-safe lazy initialization to support zero-value struct usage.
func (a *PackageManager) getBinaryName() string {
	a.binaryOnce.Do(func() {
		if a.binaryName == "" {
			a.binaryName = pm
		}
	})
	return a.binaryName
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
		err := a.getRunner().RunInteractive(ctx, a.getBinaryName(), args, aptNonInteractiveEnv...)
		return nil, err
	}

	// Use RunContext for non-interactive execution (automatically includes LC_ALL=C)
	return a.getRunner().RunContext(ctx, a.getBinaryName(), args, aptNonInteractiveEnv...)
}

// IsAvailable checks if the apt package manager is available on the system.
// It verifies both that apt exists and that it's the Debian apt package manager
// (not the Java Annotation Processing Tool with the same name on some systems).
func (a *PackageManager) IsAvailable() bool {
	// First check if apt command exists
	_, err := exec.LookPath(a.getBinaryName())
	if err != nil {
		return false
	}

	// Verify it's the Debian apt by checking for dpkg (Debian package manager)
	_, dpkgErr := exec.LookPath("dpkg")
	if dpkgErr != nil {
		return false
	}

	// Test if this is actually functional Debian apt by trying a safe command
	// This approach: if apt+dpkg work together, support them regardless of platform
	output, err := a.getRunner().Run(a.getBinaryName(), "--version")
	if err != nil {
		return false
	}

	// Verify the output looks like Debian apt (not Java apt)
	outputStr := string(output)
	// Debian apt version output typically contains "apt" and version info
	// Java apt would have very different output
	return len(outputStr) > 0 &&
		strings.Contains(strings.ToLower(outputStr), "apt") &&
		!strings.Contains(strings.ToLower(outputStr), "java")
}

// GetPackageManager returns the name of the apt package manager.
func (a *PackageManager) GetPackageManager() string {
	return a.getBinaryName()
}

// Install installs the provided packages using the apt package manager.
func (a *PackageManager) Install(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	// Validate package names to prevent command injection
	if err := manager.ValidatePackageNames(pkgs); err != nil {
		return nil, err
	}

	args := append([]string{"install", ArgsFixBroken}, pkgs...)

	if opts == nil {
		opts = &manager.Options{
			DryRun:      false,
			Interactive: false,
			Verbose:     false,
		}
	}

	if opts.DryRun {
		args = append(args, ArgsDryRun)
	}

	// assume yes if not interactive, to avoid hanging
	if !opts.Interactive {
		args = append(args, ArgsAssumeYes)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	out, err := a.executeCommand(ctx, args, opts)
	if err != nil {
		return nil, err
	}

	// Interactive mode returns empty slice (output goes directly to user)
	if opts != nil && opts.Interactive {
		return []manager.PackageInfo{}, nil
	}

	return ParseInstallOutput(string(out), opts), nil
}

// Delete removes the provided packages using the apt package manager.
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

	// Start with base remove command
	args := []string{"remove"}

	// always fix broken and auto-remove unused dependencies
	args = append(args, ArgsFixBroken, ArgsAutoRemove)

	// Add dry-run if requested
	if opts.DryRun {
		args = append(args, ArgsDryRun)
	}

	// assume yes if not interactive, to avoid hanging
	if !opts.Interactive {
		args = append(args, ArgsAssumeYes)
	}

	// Check if purge is requested in CustomCommandArgs
	for _, arg := range opts.CustomCommandArgs {
		if arg == ArgsPurge {
			args = append(args, ArgsPurge)
			break
		}
	}

	// Append package names to the command arguments
	args = append(args, pkgs...)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	out, err := a.executeCommand(ctx, args, opts)
	if err != nil {
		return nil, err
	}

	// Interactive mode returns empty slice (output goes directly to user)
	if opts != nil && opts.Interactive {
		return []manager.PackageInfo{}, nil
	}

	return ParseDeletedOutput(string(out), opts), nil
}

// Refresh updates the package list using the apt package manager.
func (a *PackageManager) Refresh(opts *manager.Options) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if opts == nil {
		opts = &manager.Options{
			DryRun:      false,
			Interactive: false,
			Verbose:     false,
		}
	}
	args := []string{"update"}
	out, err := a.executeCommand(ctx, args, opts)
	if err != nil {
		return err
	}

	// Interactive mode output goes directly to user, no need to process
	if opts != nil && opts.Interactive {
		return nil
	}

	if opts.Verbose {
		log.Println(string(out))
	}
	return nil
}

// Find searches for packages matching the provided keywords using the apt package manager.
func (a *PackageManager) Find(keywords []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	// Validate keywords to prevent command injection
	if err := manager.ValidatePackageNames(keywords); err != nil {
		return nil, err
	}

	args := append([]string{"search"}, keywords...)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	out, err := a.getRunner().RunContext(ctx, pm, args, aptNonInteractiveEnv...)
	if err != nil {
		return nil, err
	}

	return a.ParseFindOutput(string(out), opts), nil
}

// ListInstalled lists all installed packages using the apt package manager.
func (a *PackageManager) ListInstalled(opts *manager.Options) ([]manager.PackageInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// NOTE: can also use `apt list --installed`, but it's slower
	out, err := a.getRunner().RunContext(ctx, dpkgQueryCmd, []string{"-W", "-f", "${binary:Package} ${Version}\n"}, aptNonInteractiveEnv...)
	if err != nil {
		return nil, err
	}
	return ParseListInstalledOutput(string(out), opts), nil
}

// ListUpgradable lists all upgradable packages using the apt package manager.
func (a *PackageManager) ListUpgradable(opts *manager.Options) ([]manager.PackageInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	out, err := a.getRunner().RunContext(ctx, pm, []string{"list", "--upgradable"}, aptNonInteractiveEnv...)
	if err != nil {
		return nil, err
	}
	return ParseListUpgradableOutput(string(out), opts), nil
}

// Upgrade upgrades the provided packages using the apt package manager.
func (a *PackageManager) Upgrade(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	// Validate package names to prevent command injection
	if len(pkgs) > 0 {
		if err := manager.ValidatePackageNames(pkgs); err != nil {
			return nil, err
		}
	}

	// For specific packages, use 'apt install' since 'apt upgrade package' upgrades all packages
	// For all packages upgrade, use 'apt upgrade' (when pkgs is empty)
	var args []string
	if len(pkgs) > 0 {
		args = append([]string{"install"}, pkgs...)
	} else {
		args = []string{"upgrade"}
	}

	if opts == nil {
		opts = &manager.Options{
			DryRun:      false,
			Interactive: false,
			Verbose:     false,
		}
	}

	if opts.DryRun {
		args = append(args, ArgsDryRun)
	}
	if !opts.Interactive {
		args = append(args, ArgsAssumeYes)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	log.Printf("Running command: %s %s", pm, args)

	out, err := a.executeCommand(ctx, args, opts)
	if err != nil {
		return nil, err
	}

	// Interactive mode returns empty slice (output goes directly to user)
	if opts != nil && opts.Interactive {
		return []manager.PackageInfo{}, nil
	}

	return ParseInstallOutput(string(out), opts), nil
}

// UpgradeAll upgrades all installed packages using the apt package manager.
func (a *PackageManager) UpgradeAll(opts *manager.Options) ([]manager.PackageInfo, error) {
	return a.Upgrade(nil, opts)
}

// Clean cleans the local package cache used by the apt package manager.
func (a *PackageManager) Clean(opts *manager.Options) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if opts == nil {
		opts = &manager.Options{
			DryRun:      false,
			Interactive: false,
			Verbose:     false,
		}
	}
	args := []string{"autoclean"}
	out, err := a.executeCommand(ctx, args, opts)
	if err != nil {
		return err
	}

	// Interactive mode output goes directly to user, no need to process
	if opts != nil && opts.Interactive {
		return nil
	}

	if opts.Verbose {
		log.Println(string(out))
	}
	return nil
}

// GetPackageInfo retrieves package information for the specified package using the apt package manager.
func (a *PackageManager) GetPackageInfo(pkg string, opts *manager.Options) (manager.PackageInfo, error) {
	// Validate package name to prevent command injection
	if err := manager.ValidatePackageName(pkg); err != nil {
		return manager.PackageInfo{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	out, err := a.getRunner().RunContext(ctx, "apt-cache", []string{"show", pkg}, aptNonInteractiveEnv...)
	if err != nil {
		return manager.PackageInfo{}, err
	}
	return ParsePackageInfoOutput(string(out), opts), nil
}

// AutoRemove removes unused packages and dependencies using the apt package manager.
func (a *PackageManager) AutoRemove(opts *manager.Options) ([]manager.PackageInfo, error) {
	args := []string{"autoremove"}
	if opts == nil {
		opts = &manager.Options{
			Verbose:     false,
			DryRun:      false,
			Interactive: false,
		}
	}

	if opts.DryRun {
		args = append(args, ArgsDryRun)
	}
	if !opts.Interactive {
		args = append(args, ArgsAssumeYes)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	out, err := a.executeCommand(ctx, args, opts)
	if err != nil {
		return nil, err
	}

	// Interactive mode returns empty slice (output goes directly to user)
	if opts != nil && opts.Interactive {
		return []manager.PackageInfo{}, nil
	}

	return ParseDeletedOutput(string(out), opts), nil
}
