// Package flatpack provides an implementation of the syspkg manager interface for flatpack package manager

package flatpak

import (
	"os"
	"os/exec"

	// "github.com/rs/zerolog"
	// "github.com/rs/zerolog/log"

	"github.com/bluet/syspkg/manager"
)

var pm string = "flatpack"

const (
	ArgsAssumeYes    string = "-y"
	ArgsAssumeNo     string = ""
	ArgsDryRun       string = "--no-deploy"
	ArgsFixBroken    string = ""
	ArgsQuiet        string = ""
	ArgsPurge        string = "--delete-data"	// https://docs.flatpak.org/en/latest/flatpak-command-reference.html#flatpak-uninstall
							// When --delete-data is specified while removing an app, its data directory in ~/.var/app and any permissions it might have are removed. When --delete-data is used without a REF , all 'unowned' app data is removed.
	ArgsAutoRemove   string = "--unused"	// Remove unused refs on the system.
	ArgsShowProgress string = ""
	ArgsNonInteractive string = "--noninteractive"
	ArgsVerbose string = "--verbose"
	ArgsUpsert string = "--or-update"
)

var ENV_NonInteractive []string = []string{"LC_ALL=C"}

type PackageManager struct{}

func (a *PackageManager) IsAvailable() bool {
	_, err := exec.LookPath(pm)
	return err == nil
}

func (a *PackageManager) GetPackageManager() string {
	return pm
}

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

func (a *PackageManager) Refresh(opts *manager.Options) (error) {
	// not sure if this is needed

	return nil
}

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
			return nil, err
		}
		return ParseFindOutput(string(out), opts), nil
	}
}

func (a *PackageManager) ListInstalled(opts *manager.Options) ([]manager.PackageInfo, error) {
	cmd := exec.Command("flatpak", "list")
	cmd.Env = ENV_NonInteractive
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return ParseListInstalledOutput(string(out), opts), nil
}

func (a *PackageManager) ListUpgradable(opts *manager.Options) ([]manager.PackageInfo, error) {
	cmd := exec.Command(pm, "remote-ls", "--update")
	cmd.Env = ENV_NonInteractive
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return ParseListUpgradableOutput(string(out), opts), nil
}

func (a *PackageManager) Upgrade(opts *manager.Options) ([]manager.PackageInfo, error) {
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

func (a *PackageManager) GetPackageInfo(pkg string, opts *manager.Options) (manager.PackageInfo, error) {
	cmd := exec.Command(pm, "info", pkg)
	cmd.Env = ENV_NonInteractive
	out, err := cmd.Output()
	if err != nil {
		return manager.PackageInfo{}, err
	}
	return ParsePackageInfoOutput(string(out), opts), nil
}