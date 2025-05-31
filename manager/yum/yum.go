// Package yum provides an implementation of the syspkg manager interface for the yum package manager.
// It provides a Go (golang) API interface for interacting with the YUM package manager.
// This package is a wrapper around the yum command line tool.
//
// YUM was the default package manager on RedHat-based systems such as CentOS, it has been recently superseded by DNF (Dandified YUM)
//
// Behavior Contracts:
//
// Status Determination:
//   - YUM search output does not indicate installation status
//   - Find() always returns PackageStatusAvailable (limitation of YUM output format)
//   - To determine actual installation status, use ListInstalled() or GetPackageInfo()
//   - GetPackageInfo() shows "Installed Packages" vs "Available Packages" sections
//
// Field Usage by Operation:
//   - Find: Status=available, Version="", NewVersion="" (YUM limitation)
//   - ListInstalled: Status=installed, Version=installed_version, NewVersion=""
//   - GetPackageInfo: Status=available/installed (based on section), Version=package_version
//
// Cross-Package Manager Compatibility Note:
//   - Unlike APT, YUM cannot determine installation status from search results
//   - Users should use GetPackageInfo() for accurate status determination
//
// This package is part of the syspkg library.
package yum

import (
	"context"
	"log"
	"os"
	"os/exec"
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
type PackageManager struct{}

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
	cmd := exec.CommandContext(ctx, pm, args...)

	if opts.Interactive {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		if err != nil {
			return nil, err
		}
		// For interactive mode, we can't parse output, return empty list
		return []manager.PackageInfo{}, nil
	}

	out, err := cmd.Output()
	if err != nil {
		return nil, err
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
	cmd := exec.CommandContext(ctx, pm, args...)

	if opts.Interactive {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		if err != nil {
			return nil, err
		}
		// For interactive mode, we can't parse output, return empty list
		return []manager.PackageInfo{}, nil
	}

	out, err := cmd.Output()
	if err != nil {
		return nil, err
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

	cmd := exec.CommandContext(ctx, pm, "clean", "expire-cache")

	if opts.Interactive {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		return err
	} else {
		out, err := cmd.Output()
		if err != nil {
			return err
		}
		if opts.Verbose {
			log.Println(string(out))
		}
		return nil
	}
}

// Find searches for packages matching the provided keywords using the yum package manager.
//
// IMPORTANT: Due to YUM output limitations, this method always returns PackageStatusAvailable
// regardless of actual installation status. YUM search output does not indicate whether
// packages are installed or not.
//
// To determine accurate installation status:
//   - Use GetPackageInfo() which shows "Installed Packages" vs "Available Packages"
//   - Use ListInstalled() and check if the package appears in the list
//
// Returned fields:
//   - Name: Package name
//   - Arch: Package architecture
//   - Status: Always PackageStatusAvailable (YUM limitation)
//   - Version: Always empty (not provided by yum search)
//   - NewVersion: Always empty (not provided by yum search)
func (a *PackageManager) Find(keywords []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	defer cancel()

	args := append([]string{"search"}, keywords...)
	cmd := exec.CommandContext(ctx, pm, args...)

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return ParseFindOutput(string(out), opts), nil
}

// ListInstalled lists all installed packages using the yum package manager.
func (a *PackageManager) ListInstalled(opts *manager.Options) ([]manager.PackageInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	defer cancel()

	args := []string{"list", "--installed"}
	cmd := exec.CommandContext(ctx, pm, args...)
	out, err := cmd.Output()
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
	cmd := exec.CommandContext(ctx, pm, args...)

	out, err := cmd.Output()
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
	cmd := exec.CommandContext(ctx, pm, args...)

	if opts.Interactive {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		if err != nil {
			return nil, err
		}
		// For interactive mode, we can't parse output, return empty list
		return []manager.PackageInfo{}, nil
	}

	out, err := cmd.Output()
	if err != nil {
		return nil, err
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

	cmd := exec.CommandContext(ctx, pm, args...)

	if opts.Interactive {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		if err != nil {
			return nil, err
		}
		// For interactive mode, we can't parse output, return empty list
		return []manager.PackageInfo{}, nil
	}

	out, err := cmd.Output()
	if err != nil {
		return nil, err
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

	cmd := exec.CommandContext(ctx, pm, "clean", "all")

	if opts.Interactive {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		return cmd.Run()
	}

	out, err := cmd.Output()
	if err != nil {
		return err
	}
	if opts.Verbose {
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
	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, pm, "info", pkg)
	out, err := cmd.Output()
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

	cmd := exec.CommandContext(ctx, pm, args...)

	if opts.Interactive {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		if err != nil {
			return nil, err
		}
		// For interactive mode, we can't parse output, return empty list
		return []manager.PackageInfo{}, nil
	}

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	if opts.Verbose {
		log.Println(string(out))
	}

	return ParseAutoRemoveOutput(string(out), opts), nil
}
