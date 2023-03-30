// apt/apt.go
package yum

import (
	"errors"
	"os/exec"
	"strings"

	"github.com/bluet/syspkg/internal"
)

var pm string = "yum"

type PackageManager struct{}

func (a *PackageManager) IsAvailable() bool {
	_, err := exec.LookPath(pm)
	return err == nil
}

func (a *PackageManager) Install(pkg string) error {
	// Implement the installation logic for apt
	return errors.New("Not implemented")
}

func (a *PackageManager) Uninstall(pkg string) error {
	// Implement the uninstallation logic for apt
	return errors.New("Not implemented")
}

func (a *PackageManager) Search(pkg string) ([]internal.PackageInfo, error) {
	// Implement the search logic for apt
	return nil, errors.New("Not implemented")
}

func (a *PackageManager) ListInstalled() ([]internal.PackageInfo, error) {
	// Implement the list installed packages logic for apt
	return nil, errors.New("Not implemented")
}

func (a *PackageManager) ListUpgradable() ([]internal.PackageInfo, error) {
	var cmd *exec.Cmd = exec.Command(pm, "upgrade", "-s")

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseOutput(string(out)), nil
}

func parseOutput(output string) []internal.PackageInfo {
	// raw string: Inst libpulse-dev [1:15.99.1+dfsg1-1ubuntu2] (1:15.99.1+dfsg1-1ubuntu2.1 Ubuntu:22.04/jammy-updates [amd64]) []
	// format: STATUS NAME [VERSION] (NEW_VERSION CATEGORY [ARCHITECTURES]) [OTHER_INFO]

	lines := strings.Split(output, "\n")
	var packages []internal.PackageInfo

	for _, line := range lines {
		if strings.HasPrefix(line, "Inst") {
			parts := strings.Fields(line)

			packageInfo := internal.PackageInfo{
				Name:       parts[1],
				Version:    strings.Trim(parts[2], "[]"),
				NewVersion: strings.Trim(parts[3], "()"),
				Category:   parts[4],
				Arch:       strings.Trim(parts[5], "[]"),
				Status:     internal.Upgradable,
			}

			packages = append(packages, packageInfo)
		}
	}

	return packages
}
