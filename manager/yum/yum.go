// Package yum provides an implementation of the syspkg manager interface for the yum package manager.
// It provides a Go (golang) API interface for interacting with the YUM package manager.
// This package is a wrapper around the yum command line tool.
//
// YUM was the default package manager on RedHat-based systems such as CentOS, it has been recently superseded by DNF (Dandified YUM)
//
// This package is part of the syspkg library.
package yum

import (
	"context"
	"errors"
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

func (a *PackageManager) Install(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	return nil, errors.New("not implemented")
}

func (a *PackageManager) Delete(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	return nil, errors.New("not implemented")
}

// Refresh updates the package list using the yum package manager.
// Uses 'yum clean expire-cache' which efficiently refreshes metadata without
// aggressive cache clearing. This preserves valid cache files while ensuring
// up-to-date repository information.
func (a *PackageManager) Refresh(opts *manager.Options) error {
	ctx, cancel := context.WithTimeout(context.Background(), cleanTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, pm, "clean", "expire-cache")

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

// Find searches for packages matching the provided keywords using the yum package manager.
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

func (a *PackageManager) ListUpgradable(opts *manager.Options) ([]manager.PackageInfo, error) {
	return nil, errors.New("not implemented")
}
func (a *PackageManager) Upgrade(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	return nil, errors.New("not implemented")
}
func (a *PackageManager) UpgradeAll(opts *manager.Options) ([]manager.PackageInfo, error) {
	return nil, errors.New("not implemented")
}

// Clean performs comprehensive cleanup of YUM caches.
// Uses 'yum clean all' which removes all cached packages, metadata, and headers.
// This is what administrators typically expect from a clean operation.
func (a *PackageManager) Clean(opts *manager.Options) error {
	ctx, cancel := context.WithTimeout(context.Background(), cleanTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, pm, "clean", "all")

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

func (a *PackageManager) AutoRemove(opts *manager.Options) ([]manager.PackageInfo, error) {
	return nil, errors.New("not implemented")
}
