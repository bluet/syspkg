// Package yum provides an implementation of the syspkg manager interface for the yum package manager.
// It provides an Go (golang) API interface for interacting with the YUM package manager.
// This package is a wrapper around the yum command line tool.
//
// YUM was the default package manager on RedHat-based systems such as Centos, it has been recently superseded by DNF (Dandified YUM)
//
// This package is part of the syspkg library.
package yum

import (
	"errors"
	"log"
	"os"
	"os/exec"

	"github.com/bluet/syspkg/manager"
)

var pm string = "yum"

// Constants used for yum commands
const (
	ArgsAssumeYes    string = "-y"
	ArgsAssumeNo     string = "--assumeno"
	ArgsQuiet        string = "-q"
	ArgsDryRun       string = ""
	ArgsFixBroken    string = ""
	ArgsPurge        string = ""
	ArgsAutoRemove   string = ""
	ArgsShowProgress string = ""
)

// ENV_NonInteractive contains environment variables used to set non-interactive mode for yum and dpkg.
var ENV_NonInteractive []string = []string{}

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
func (a *PackageManager) Refresh(opts *manager.Options) error {
	cmd := exec.Command(pm, "clean", "expire-cache")
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

// Find searches for packages matching the provided keywords using the yum package manager.
func (a *PackageManager) Find(keywords []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	args := append([]string{"search"}, keywords...)
	cmd := exec.Command(pm, args...)
	cmd.Env = ENV_NonInteractive

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return ParseFindOutput(string(out), opts), nil
}

// ListInstalled lists all installed packages using the yum package manager.
func (a *PackageManager) ListInstalled(opts *manager.Options) ([]manager.PackageInfo, error) {
	args := []string{"list", "--installed"}
	cmd := exec.Command(pm, args...)
	cmd.Env = ENV_NonInteractive
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
func (a *PackageManager) Clean(opts *manager.Options) error {
	return a.Refresh(nil)
}

// GetPackageInfo retrieves package information for the specified package using the yum package manager.
func (a *PackageManager) GetPackageInfo(pkg string, opts *manager.Options) (manager.PackageInfo, error) {
	cmd := exec.Command(pm, "info", pkg)
	cmd.Env = ENV_NonInteractive
	out, err := cmd.Output()
	if err != nil {
		return manager.PackageInfo{}, err
	}
	return ParsePackageInfoOutput(string(out), opts), nil
}

func (a *PackageManager) AutoRemove(opts *manager.Options) ([]manager.PackageInfo, error) {
	return nil, errors.New("not implemented")
}
