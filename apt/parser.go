package apt

import (
	"log"
	"strings"

	"github.com/bluet/syspkg/internal"
)

func parseInstallOutput(output string, opts *internal.Options) []internal.PackageInfo {
	lines := strings.Split(output, "\n")
	var packages []internal.PackageInfo

	for _, line := range lines {
		if opts.Verbose {
			log.Printf("apt: %s", line)
		}
		if strings.HasPrefix(line, "Setting up") {
			parts := strings.Fields(line)
			packageInfo := internal.PackageInfo{
				Name:           strings.Split(parts[2], ":")[0],
				Arch:           strings.Split(parts[2], ":")[1],
				Version:        strings.Trim(parts[3], "()"),
				NewVersion:     strings.Trim(parts[3], "()"),
				Status:         internal.Installed,
				PackageManager: pm,
			}
			packages = append(packages, packageInfo)
		}
	}

	return packages
}

func parseSearchOutput(output string, opts *internal.Options) []internal.PackageInfo {
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

func parseListInstalledOutput(output string, opts *internal.Options) []internal.PackageInfo {
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

func parseListUpgradableOutput(output string, opts *internal.Options) []internal.PackageInfo {
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
