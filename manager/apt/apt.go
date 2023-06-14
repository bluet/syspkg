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
	"log"
	"os"
	"os/exec"

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
)

// ENV_NonInteractive contains environment variables used to set non-interactive mode for apt and dpkg.
var ENV_NonInteractive []string = []string{"LC_ALL=C", "DEBIAN_FRONTEND=noninteractive", "DEBCONF_NONINTERACTIVE_SEEN=true"}

// PackageManager implements the manager.PackageManager interface for the apt package manager.
type PackageManager struct{}

// IsAvailable checks if the apt package manager is available on the system.
func (a *PackageManager) IsAvailable() bool {
	_, err := exec.LookPath(pm)
	return err == nil
}

// GetPackageManager returns the name of the apt package manager.
func (a *PackageManager) GetPackageManager() string {
	return pm
}

// Install installs the provided packages using the apt package manager.
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

// Delete removes the provided packages using the apt package manager.
func (a *PackageManager) Delete(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	// args := append([]string{"remove", ArgsFixBroken, ArgsPurge, ArgsAutoRemove}, pkgs...)
	args := append([]string{"remove", ArgsFixBroken, ArgsAutoRemove}, pkgs...)
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
		return ParseDeletedOutput(string(out), opts), nil
	}
}

// Refresh updates the package list using the apt package manager.
func (a *PackageManager) Refresh(opts *manager.Options) error {
	cmd := exec.Command(pm, "update")
	cmd.Env = ENV_NonInteractive

	if opts == nil {
		opts = &manager.Options{
			DryRun:      false,
			Interactive: false,
			Verbose:     false,
		}
	}
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

// Find searches for packages matching the provided keywords using the apt package manager.
func (a *PackageManager) Find(keywords []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	args := append([]string{"search"}, keywords...)
	cmd := exec.Command("apt", args...)
	cmd.Env = ENV_NonInteractive

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return ParseFindOutput(string(out), opts), nil
}

// ListInstalled lists all installed packages using the apt package manager.
func (a *PackageManager) ListInstalled(opts *manager.Options) ([]manager.PackageInfo, error) {
	cmd := exec.Command("dpkg-query", "-W", "-f", "${binary:Package} ${Version}\n")
	// NOTE: can also use `apt list --installed`, but it's slower
	cmd.Env = ENV_NonInteractive
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return ParseListInstalledOutput(string(out), opts), nil
}

// ListUpgradable lists all upgradable packages using the apt package manager.
func (a *PackageManager) ListUpgradable(opts *manager.Options) ([]manager.PackageInfo, error) {
	cmd := exec.Command(pm, "list", "--upgradable")
	cmd.Env = ENV_NonInteractive
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return ParseListUpgradableOutput(string(out), opts), nil
}

// Upgrade upgrades the provided packages using the apt package manager.
func (a *PackageManager) Upgrade(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	args := []string{"upgrade"}
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

// UpgradeAll upgrades all installed packages using the apt package manager.
func (a *PackageManager) UpgradeAll(opts *manager.Options) ([]manager.PackageInfo, error) {
	// TODO: add support for upgrade specific packages
	return a.Upgrade(nil, opts)
}

// Clean cleans the local package cache used by the apt package manager.
func (a *PackageManager) Clean(opts *manager.Options) error {
	cmd := exec.Command(pm, "autoclean")
	cmd.Env = ENV_NonInteractive

	if opts == nil {
		opts = &manager.Options{
			DryRun:      false,
			Interactive: false,
			Verbose:     false,
		}
	}
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

// GetPackageInfo retrieves package information for the specified package using the apt package manager.
func (a *PackageManager) GetPackageInfo(pkg string, opts *manager.Options) (manager.PackageInfo, error) {
	cmd := exec.Command("apt-cache", "show", pkg)
	cmd.Env = ENV_NonInteractive
	out, err := cmd.Output()
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
		return ParseDeletedOutput(string(out), opts), nil
	}
}
