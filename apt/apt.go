package apt

import (
	"log"
	"os"
	"os/exec"

	"github.com/bluet/syspkg/internal"
)

var pm string = "apt"

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

func (a *PackageManager) Install(pkgs []string, opts *internal.Options) ([]internal.PackageInfo, error) {
	args := append([]string{"install", ArgsFixBroken}, pkgs...)

	if opts == nil {
		opts = &internal.Options{
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

func (a *PackageManager) Delete(pkgs []string, opts *internal.Options) ([]internal.PackageInfo, error) {
	args := append([]string{"remove", ArgsFixBroken, ArgsPurge, ArgsAutoRemove}, pkgs...)
	if opts == nil {
		opts = &internal.Options{
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

func (a *PackageManager) Refresh(opts *internal.Options) error {
	cmd := exec.Command(pm, "update")
	cmd.Env = ENV_NonInteractive

	if opts == nil {
		opts = &internal.Options{
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

func (a *PackageManager) Find(keywords []string, opts *internal.Options) ([]internal.PackageInfo, error) {
	args := append([]string{"search"}, keywords...)
	cmd := exec.Command("apt", args...)
	cmd.Env = ENV_NonInteractive

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return ParseFindOutput(string(out), opts), nil
}

func (a *PackageManager) ListInstalled(opts *internal.Options) ([]internal.PackageInfo, error) {
	cmd := exec.Command("dpkg-query", "-W", "-f", "${binary:Package} ${Version}\n")
	cmd.Env = ENV_NonInteractive
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return ParseListInstalledOutput(string(out), opts), nil
}

func (a *PackageManager) ListUpgradable(opts *internal.Options) ([]internal.PackageInfo, error) {
	cmd := exec.Command(pm, "list", "--upgradable")
	cmd.Env = ENV_NonInteractive
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return ParseListUpgradableOutput(string(out), opts), nil
}

func (a *PackageManager) Upgrade(opts *internal.Options) ([]internal.PackageInfo, error) {
	args := []string{"upgrade"}
	if opts == nil {
		opts = &internal.Options{
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

func (a *PackageManager) Clean(opts *internal.Options) error {
	cmd := exec.Command(pm, "autoclean")
	cmd.Env = ENV_NonInteractive

	if opts == nil {
		opts = &internal.Options{
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

func (a *PackageManager) GetPackageInfo(pkg string, opts *internal.Options) (internal.PackageInfo, error) {
	cmd := exec.Command("apt-cache", "show", pkg)
	cmd.Env = ENV_NonInteractive
	out, err := cmd.Output()
	if err != nil {
		return internal.PackageInfo{}, err
	}
	return ParsePackageInfoOutput(string(out), opts), nil
}

func (a *PackageManager) AutoRemove(opts *internal.Options) ([]internal.PackageInfo, error) {
	args := []string{"autoremove"}
	if opts == nil {
		opts = &internal.Options{
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
