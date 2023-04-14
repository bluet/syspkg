// Package snap provides an implementation of the syspkg manager interface for snap package manager
// https://snapcraft.io/docs/getting-started
// https://en.wikipedia.org/wiki/Snap_(software)

package snap

import (
	"log"
	"os"
	"os/exec"

	"github.com/bluet/syspkg/manager"
)

var pm string = "snap"

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

var ENV_NonInteractive []string = []string{"LC_ALL=C", "DEBIAN_FRONTEND=noninteractive", "DEBCONF_NONINTERACTIVE_SEEN=true"}

type PackageManager struct{}

func (a *PackageManager) IsAvailable() bool {
	_, err := exec.LookPath(pm)
	return err == nil
}

func (a *PackageManager) GetPackageManager() string {
	return pm
}

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

func (a *PackageManager) Refresh(opts *manager.Options) error {
	return nil
}

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


func (a *PackageManager) ListInstalled(opts *manager.Options) ([]manager.PackageInfo, error) {
	cmd := exec.Command("snap", "list")
	cmd.Env = ENV_NonInteractive
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return ParseListInstalledOutput(string(out), opts), nil
}

func (a *PackageManager) ListUpgradable(opts *manager.Options) ([]manager.PackageInfo, error) {
	cmd := exec.Command(pm, "refresh", "--list")
	cmd.Env = ENV_NonInteractive
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return ParseListUpgradableOutput(string(out), opts), nil
}

func (a *PackageManager) Upgrade(opts *manager.Options) ([]manager.PackageInfo, error) {
	args := []string{"refresh", ArgsFixBroken}

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