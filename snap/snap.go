package snap

import (
	"os/exec"
	"strings"

	"github.com/bluet/syspkg/internal"
)

var pm string = "snap"

type PackageManager struct{}

func (s *PackageManager) IsAvailable() bool {
	_, err := exec.LookPath(pm)
	return err == nil
}

func (s *PackageManager) Install(pkgs []string) error {
	// Snap package manager installs one package at a time
	for _, pkg := range pkgs {
		cmd := exec.Command(pm, "install", pkg)
		err := cmd.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *PackageManager) Uninstall(pkgs []string) error {
	// Snap package manager removes one package at a time
	for _, pkg := range pkgs {
		cmd := exec.Command(pm, "remove", pkg)
		err := cmd.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

// Implement Update function for Snap
func (s *PackageManager) Update() error {
	cmd := exec.Command(pm, "refresh")
	err := cmd.Run()
	return err
}

func (s *PackageManager) Search(keywords []string) ([]internal.PackageInfo, error) {
	args := append([]string{"find"}, keywords...)
	cmd := exec.Command(pm, args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseSearchOutput(string(out)), nil
}

func (s *PackageManager) ListInstalled() ([]internal.PackageInfo, error) {
	cmd := exec.Command(pm, "list", "--all")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseListInstalledOutput(string(out)), nil
}

func (s *PackageManager) ListUpgradable() ([]internal.PackageInfo, error) {
	cmd := exec.Command(pm, "list", "--all")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseListUpgradableOutput(string(out)), nil
}

func (s *PackageManager) Upgrade() error {
	cmd := exec.Command(pm, "refresh")
	err := cmd.Run()
	return err
}

func parseSearchOutput(output string) []internal.PackageInfo {
	lines := strings.Split(output, "\n")
	var packages []internal.PackageInfo

	for _, line := range lines {
		if len(line) > 0 {
			parts := strings.Fields(line)
			if parts[0] == "Name" {
				continue
			}
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
		if len(line) > 0 {
			parts := strings.Fields(line)
			if parts[3] == "upgrade" {
				packageInfo := internal.PackageInfo{
					Name:           parts[0],
					Version:        parts[1],
					NewVersion:     parts[2],
					Status:         internal.Upgradable,
					PackageManager: pm,
				}
				packages = append(packages, packageInfo)
			}
		}
	}

	return packages
}
