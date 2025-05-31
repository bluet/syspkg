// Package flatpak provides an implementation of the syspkg manager interface for the Flatpak package manager.
// It provides a unified interface for interacting with the Flatpak package manager.
// This package is a wrapper around the Flatpak command-line interface.
//
// Flatpak is a software utility for software deployment, package management, and application virtualization for Linux desktop computers.
// The Flatpak command-line interface aims to be a complete tool for installing, managing, and running Flatpak software.
// For more information about Flatpak, visit:
// - https://flatpak.org/
// - https://docs.flatpak.org/en/latest/flatpak-command-reference.html
//
// This package is part of the syspkg library.
package flatpak

import (
	"log"
	"os"
	"os/exec"

	// "github.com/rs/zerolog"
	// "github.com/rs/zerolog/log"

	"github.com/bluet/syspkg/manager"
)

var pm string = "flatpak"

// Constants representing Flatpak command arguments.
const (
	ArgsAssumeYes string = "-y"
	ArgsAssumeNo  string = ""
	ArgsDryRun    string = "--no-deploy"
	ArgsFixBroken string = ""
	ArgsQuiet     string = ""
	ArgsPurge     string = "--delete-data" // https://docs.flatpak.org/en/latest/flatpak-command-reference.html#flatpak-uninstall
	// When --delete-data is specified while removing an app, its data directory in ~/.var/app and any permissions it might have are removed. When --delete-data is used without a REF , all 'unowned' app data is removed.
	ArgsAutoRemove     string = "--unused" // Remove unused refs on the system.
	ArgsShowProgress   string = ""
	ArgsNonInteractive string = "--noninteractive"
	ArgsVerbose        string = "--verbose"
	ArgsUpsert         string = "--or-update"
)

// ENV_NonInteractive is an environment variable that sets the locale to C for non-interactive mode.
var ENV_NonInteractive []string = []string{"LC_ALL=C"}

// PackageManager implements the syspkg manager interface for Flatpak.
type PackageManager struct{}

// IsAvailable checks if the Flatpak package manager is available on the system.
func (a *PackageManager) IsAvailable() bool {
	_, err := exec.LookPath(pm)
	return err == nil
}

// GetPackageManager returns the name of the Flatpak package manager.
func (a *PackageManager) GetPackageManager() string {
	return pm
}

// Install installs the given packages using Flatpak with the provided options.
func (a *PackageManager) Install(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	args := append([]string{"install", ArgsFixBroken, ArgsUpsert, ArgsVerbose}, pkgs...)

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
		args = append(args, ArgsAssumeYes, ArgsNonInteractive)
	}

	if opts.Verbose {
		args = append(args, ArgsVerbose)
	}

	cmd := exec.Command(pm, args...)

	if opts.Interactive {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		return nil, err
	} else {
		cmd.Env = ENV_NonInteractive
		out, err := cmd.Output()
		if err != nil {
			return nil, err
		}
		return ParseInstallOutput(string(out), opts), nil
	}
}

// Delete removes the given packages using Flatpak with the provided options.
func (a *PackageManager) Delete(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	args := append([]string{"uninstall", ArgsFixBroken, ArgsVerbose}, pkgs...)

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
		args = append(args, ArgsAssumeYes, ArgsNonInteractive)
	}

	if opts.Verbose {
		args = append(args, ArgsVerbose)
	}

	cmd := exec.Command(pm, args...)

	if opts.Interactive {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		return nil, err
	} else {
		cmd.Env = ENV_NonInteractive
		out, err := cmd.Output()
		if err != nil {
			return nil, err
		}
		return ParseInstallOutput(string(out), opts), nil
	}
}

// Refresh updates the package metadata for Flatpak. Not currently implemented.
func (a *PackageManager) Refresh(opts *manager.Options) error {
	// not sure if this is needed

	return nil
}

// Find searches for packages matching the given keywords using Flatpak with the provided options.
func (a *PackageManager) Find(keywords []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	args := append([]string{"search", ArgsVerbose}, keywords...)

	if opts == nil {
		opts = &manager.Options{
			DryRun:      false,
			Interactive: false,
			Verbose:     false,
		}
	}

	if opts.Verbose {
		args = append(args, ArgsVerbose)
	}

	cmd := exec.Command(pm, args...)

	if opts.Interactive {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		return nil, err
	} else {
		cmd.Env = ENV_NonInteractive
		out, err := cmd.Output()
		if err != nil {
			// Flatpak search returns exit code 1 when no packages found - this is not an error
			if exitError, ok := err.(*exec.ExitError); ok {
				if exitError.ExitCode() == 1 {
					// No packages found, return empty list
					return []manager.PackageInfo{}, nil
				}
			}
			return nil, err
		}
		return ParseFindOutput(string(out), opts), nil
	}
}

// ListInstalled lists installed packages using Flatpak with the provided options.
func (a *PackageManager) ListInstalled(opts *manager.Options) ([]manager.PackageInfo, error) {
	cmd := exec.Command("flatpak", "list")
	cmd.Env = ENV_NonInteractive
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return ParseListInstalledOutput(string(out), opts), nil
}

// ListUpgradable lists upgradable packages using Flatpak with the provided options.
func (a *PackageManager) ListUpgradable(opts *manager.Options) ([]manager.PackageInfo, error) {
	cmd := exec.Command(pm, "remote-ls", "--updates")
	cmd.Env = ENV_NonInteractive
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return ParseListUpgradableOutput(string(out), opts), nil
}

// UpgradeAll upgrades all packages using Flatpak with the provided options.
func (a *PackageManager) UpgradeAll(opts *manager.Options) ([]manager.PackageInfo, error) {
	args := []string{"update"}
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

	cmd := exec.Command(pm, args...)

	log.Printf("Running command: %s %s", pm, args)

	if opts.Interactive {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		return nil, err
	}

	cmd.Env = ENV_NonInteractive
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return ParseInstallOutput(string(out), opts), nil
}

// GetPackageInfo retrieves package information for a single package using Flatpak with the provided options.
func (a *PackageManager) GetPackageInfo(pkg string, opts *manager.Options) (manager.PackageInfo, error) {
	cmd := exec.Command(pm, "info", pkg)
	cmd.Env = ENV_NonInteractive
	out, err := cmd.Output()
	if err != nil {
		return manager.PackageInfo{}, err
	}
	return ParsePackageInfoOutput(string(out), opts), nil
}
