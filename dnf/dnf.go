package dnf

import (
	"os/exec"
	"strings"

	"github.com/bluet/syspkg/internal"
)

var pm string = "dnf"

type PackageManager struct{}

func (d *PackageManager) IsAvailable() bool {
	_, err := exec.LookPath(pm)
	return err == nil
}

func (d *PackageManager) Install(pkgs []string) error {
	args := append([]string{"install", "-y"}, pkgs...)
	cmd := exec.Command(pm, args...)
	err := cmd.Run()
	return err
}

func (d *PackageManager) Uninstall(pkgs []string) error {
	args := append([]string{"remove", "-y"}, pkgs...)
	cmd := exec.Command(pm, args...)
	err := cmd.Run()
	return err
}

func (d *PackageManager) Update() error {
	cmd := exec.Command(pm, "check-update")
	err := cmd.Run()
	return err
}

func (d *PackageManager) Search(keywords []string) ([]internal.PackageInfo, error) {
	args := append([]string{"search"}, keywords...)
	cmd := exec.Command(pm, args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseSearchOutput(string(out)), nil
}

func (d *PackageManager) ListInstalled() ([]internal.PackageInfo, error) {
	cmd := exec.Command(pm, "list", "installed")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseListInstalledOutput(string(out)), nil
}

func (d *PackageManager) ListUpgradable() ([]internal.PackageInfo, error) {
	cmd := exec.Command(pm, "list", "updates")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseListUpgradableOutput(string(out)), nil
}

func (d *PackageManager) Upgrade() error {
	cmd := exec.Command(pm, "upgrade", "-y")
	err := cmd.Run()
	return err
}

func parseSearchOutput(output string) []internal.PackageInfo {
	lines := strings.Split(output, "\n")
	var packages []internal.PackageInfo

	for _, line := range lines {
		if len(line) > 0 {
			parts := strings.Fields(line)
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
		if len(line) > 0 && !strings.HasPrefix(line, "Installed") {
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
		if strings.HasPrefix(line, "Upgrades") {
			parts := strings.Fields(line)
			packageInfo := internal.PackageInfo{
				Name:           parts[1],
				Version:        strings.Trim(parts[3], "[]"),
				NewVersion:     strings.Trim(parts[5], "()"),
				Status:         internal.Upgradable,
				PackageManager: pm,
			}
			packages = append(packages, packageInfo)
		}
	}

	return packages
}
