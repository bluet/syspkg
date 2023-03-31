package apt

import (
	"os"
	"os/exec"
	"strings"

	"github.com/bluet/syspkg/internal"
)

var pm string = "apt-get"

type PackageManager struct{}

func (a *PackageManager) IsAvailable() bool {
	_, err := exec.LookPath(pm)
	return err == nil
}

func (a *PackageManager) Install(pkgs []string) error {
	args := append([]string{"install", "-y"}, pkgs...)
	cmd := exec.Command(pm, args...)
	err := cmd.Run()
	return err
}

func (a *PackageManager) Uninstall(pkgs []string) error {
	args := append([]string{"remove", "-y", "--purge"}, pkgs...)
	cmd := exec.Command(pm, args...)
	err := cmd.Run()
	return err
}

func (a *PackageManager) Update() error {
	cmd := exec.Command(pm, "update", "-qq")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

func (a *PackageManager) Search(keywords []string) ([]internal.PackageInfo, error) {
	args := append([]string{"search"}, keywords...)
	cmd := exec.Command("apt-cache", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseSearchOutput(string(out)), nil
}

func (a *PackageManager) ListInstalled() ([]internal.PackageInfo, error) {
	cmd := exec.Command("dpkg-query", "-W", "-f", "${binary:Package} ${Version}\\n")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseListInstalledOutput(string(out)), nil
}

func (a *PackageManager) ListUpgradable() ([]internal.PackageInfo, error) {
	cmd := exec.Command(pm, "upgrade", "-s")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseListUpgradableOutput(string(out)), nil
}

func (a *PackageManager) Upgrade() error {
	cmd := exec.Command(pm, "upgrade", "-y")
	err := cmd.Run()

	return err
}

func parseSearchOutput(output string) []internal.PackageInfo {
	lines := strings.Split(output, "\n")
	var packages []internal.PackageInfo

	for _, line := range lines {
		if len(line) > 0 {
			parts := strings.SplitN(line, " - ", 2)
			packageInfo := internal.PackageInfo{
				Name:           parts[0],
				Status:         internal.Available,
				PackageManager: pm,
			}
			packages = append(packages, packageInfo)
		}
	}

	return packages
}

func parseListInstalledOutput(output string) []internal.PackageInfo {
	lines := strings.Split(output, "\n")
	var packages []internal.PackageInfo

	for _, line := range lines {
		if len(line) > 0 {
			parts := strings.Fields(line)
			packageInfo := internal.PackageInfo{
				Name:           parts[0],
				Version:        parts[1],
				Status:         internal.Installed,
				PackageManager: pm,
			}
			packages = append(packages, packageInfo)
		}
	}

	return packages
}

func parseListUpgradableOutput(output string) []internal.PackageInfo {
	lines := strings.Split(output, "\n")
	var packages []internal.PackageInfo

	for _, line := range lines {
		if strings.HasPrefix(line, "Inst") {
			parts := strings.Fields(line)
			packageInfo := internal.PackageInfo{
				Name:           parts[1],
				Version:        strings.Trim(parts[2], "[]"),
				NewVersion:     strings.Trim(parts[3], "()"),
				Category:       parts[4],
				Arch:           strings.Trim(parts[5], "[]"),
				Status:         internal.Upgradable,
				PackageManager: pm,
			}
			packages = append(packages, packageInfo)
		}
	}

	return packages
}
