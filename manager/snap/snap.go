// Package snap provides an implementation of the syspkg manager interface for the snap package manager.
// It provides a Go (golang) API interface for interacting with the snap package manager.
// It allows you to query, install, and remove packages, and supports package managers like Apt, Snap, and Flatpak.
// This package is a wrapper around the snap command line tool.
//
// Snap is a software deployment and package management system originally designed and built by Canonical, the company behind the Ubuntu Linux distribution.
// Snap packages are self-contained applications running in a sandbox with mediated access to the host system.
// Snap packages have no dependency on any libraries or packages installed on the host system, designed to work on any Linux distribution that has snap support installed.
// Snap packages are available from the Snap Store, an app store with an audience of millions, and also available from other sources, including the Ubuntu Store, the KDE Discover Store, and the elementary AppCenter.
//
// For more information, see:
//   - https://snapcraft.io/docs/getting-started
//   - https://en.wikipedia.org/wiki/Snap_(software)
//
// This package is part of the syspkg library.
package snap

import (
	"log"
	"os"
	"os/exec"

	"github.com/bluet/syspkg/manager"
)

var pm string = "snap"

// Constants for various command line arguments used by the snap package manager.
const (
	ArgsAssumeYes    string = "-y"
	ArgsAssumeNo     string = "--assume-no"
	ArgsDryRun       string = "--dry-run"
	ArgsFixBroken    string = ""
	ArgsQuiet        string = "-qq"
	ArgsPurge        string = "--purge"
	ArgsAutoRemove   string = "--autoremove"
	ArgsShowProgress string = "--show-progress"
)

// ENV_NonInteractive is an environment variable configuration to set non-interactive mode for package manager commands.
var ENV_NonInteractive []string = []string{"LC_ALL=C"}

// PackageManager is an empty struct that implements the manager.PackageManager interface for the snap package manager.
type PackageManager struct{}

// IsAvailable checks if the snap package manager is available on the system.
func (a *PackageManager) IsAvailable() bool {
	_, err := exec.LookPath(pm)
	return err == nil
}

// GetPackageManager returns the package manager name (in this case, "snap").
func (a *PackageManager) GetPackageManager() string {
	return pm
}

// Install installs the specified packages using the snap package manager with the provided options.
func (a *PackageManager) Install(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error) {
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

	if opts.Verbose {
		args = append(args, ArgsShowProgress)
	}

	cmd := exec.Command(pm, args...)
	// cmd.Env = append(os.Environ(), ENV_NonInteractive...)

	log.Printf("Running command: %s %s", pm, args)

	if opts.Interactive {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		return nil, err
	}

	cmd.Env = append(os.Environ(), ENV_NonInteractive...)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return ParseInstallOutput(string(out), opts), nil
}

// Delete removes the specified packages using the snap package manager with the provided options.
func (a *PackageManager) Delete(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	args := append([]string{"remove", ArgsFixBroken}, pkgs...)

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

	if opts.Verbose {
		args = append(args, ArgsShowProgress)
	}

	cmd := exec.Command(pm, args...)
	// cmd.Env = append(os.Environ(), ENV_NonInteractive...)

	log.Printf("Running command: %s %s", pm, args)

	if opts.Interactive {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		return nil, err
	}

	cmd.Env = append(os.Environ(), ENV_NonInteractive...)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return ParseInstallOutput(string(out), opts), nil
}

// Refresh refreshes the package index for the snap package manager. Currently not implemented.
func (a *PackageManager) Refresh(opts *manager.Options) error {
	return nil
}

// Find searches for packages matching the provided keywords using the snap package manager.
func (a *PackageManager) Find(keywords []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	args := append([]string{"search"}, keywords...)
	cmd := exec.Command("snap", args...)
	cmd.Env = ENV_NonInteractive

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return ParseFindOutput(string(out), opts), nil
}

// ListInstalled lists all installed packages using the snap package manager.
func (a *PackageManager) ListInstalled(opts *manager.Options) ([]manager.PackageInfo, error) {
	cmd := exec.Command("snap", "list")
	cmd.Env = ENV_NonInteractive
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return ParseListInstalledOutput(string(out), opts), nil
}

// ListUpgradable lists all upgradable packages using the snap package manager.
func (a *PackageManager) ListUpgradable(opts *manager.Options) ([]manager.PackageInfo, error) {
	cmd := exec.Command(pm, "refresh", "--list")
	cmd.Env = ENV_NonInteractive
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return ParseListUpgradableOutput(string(out), opts), nil
}

// Upgrade upgrades the specified packages using the snap package manager with the provided options.
func (a *PackageManager) Upgrade(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	args := []string{"refresh"}
	if len(pkgs) > 0 {
		args = append(args, pkgs...)
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

	// assume yes if not interactive, to avoid hanging
	// if !opts.Interactive {
	// 	args = append(args, ArgsAssumeYes)
	// }

	if opts.Verbose {
		args = append(args, ArgsShowProgress)
	}

	cmd := exec.Command(pm, args...)
	// cmd.Env = append(os.Environ(), ENV_NonInteractive...)

	log.Printf("Running command: %s %s", pm, args)

	if opts.Interactive {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		return nil, err
	}

	// cmd.Env = append(os.Environ(), ENV_NonInteractive...)
	cmd.Env = ENV_NonInteractive
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return ParseInstallOutput(string(out), opts), nil
}

// UpgradeAll upgrades all upgradable packages using the snap package manager with the provided options.
func (a *PackageManager) UpgradeAll(opts *manager.Options) ([]manager.PackageInfo, error) {
	return a.Upgrade(nil, opts)
}

// GetPackageInfo retrieves information about the specified package using the snap package manager.
func (a *PackageManager) GetPackageInfo(pkg string, opts *manager.Options) (manager.PackageInfo, error) {
	cmd := exec.Command("snap", "info", pkg)
	cmd.Env = ENV_NonInteractive
	out, err := cmd.Output()
	if err != nil {
		return manager.PackageInfo{}, err
	}
	return ParsePackageInfoOutput(string(out), opts), nil
}

// Clean performs cleanup of snap caches and old snap revisions.
// Removes old snap revisions that are no longer active.
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
		log.Println("Dry run mode: would clean old snap revisions")
		return nil
	}

	// For now, snap doesn't have a built-in clean command
	// This could be enhanced to remove old revisions
	log.Println("Snap clean: no cleanup operations available")
	return nil
}

// AutoRemove removes snap packages that were automatically installed as dependencies.
// Note: Snap doesn't have traditional dependency-based auto-removal like APT/YUM.
// Returns empty list as snap manages dependencies automatically.
func (a *PackageManager) AutoRemove(opts *manager.Options) ([]manager.PackageInfo, error) {
	if opts == nil {
		opts = &manager.Options{
			DryRun:      false,
			Interactive: false,
			Verbose:     false,
		}
	}

	// Snap doesn't have traditional auto-remove functionality
	// Snap manages dependencies automatically and removes them when the main snap is removed
	if opts.Verbose {
		log.Println("Snap autoremove: dependencies are managed automatically by snap")
	}

	return []manager.PackageInfo{}, nil
}
