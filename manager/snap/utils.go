package snap

import (
	"fmt"
	"strings"

	"github.com/bluet/syspkg/manager"
)

// ParseInstallOutput parses the output of `snap install` command
// and returns a list of PackageInfo
//
// Example output:
// snap "deja-dup" is already installed, see 'snap help refresh'
// blablaland-desktop (edge) 1.0.1 from AdeDev installed
func ParseInstallOutput(msg string, opts *manager.Options) []manager.PackageInfo {
	var packages []manager.PackageInfo

	// remove the last empty line
	msg = strings.TrimSuffix(msg, "\n")
	var lines []string = strings.Split(string(msg), "\n")

	for _, line := range lines {
		if opts.Verbose {
			fmt.Printf("snap: %s", line)
		}
		if strings.HasPrefix(line, "snap \"") {
			parts := strings.Fields(line)
			name := strings.Trim(parts[1], "\"")
			// if name is empty, it might be not what we want
			if name == "" {
				continue
			}

			packageInfo := manager.PackageInfo{
				Name:           name,
				Status:         manager.PackageStatusInstalled,
				PackageManager: pm,
			}
			packages = append(packages, packageInfo)
		} else if strings.HasSuffix(line, "installed") {
			parts := strings.Fields(line)
			name := parts[0]
			version := parts[2]
			// if name is empty, it might be not what we want
			if name == "" {
				continue
			}

			packageInfo := manager.PackageInfo{
				Name:           name,
				Version:        version,
				Status:         manager.PackageStatusInstalled,
				PackageManager: pm,
			}
			packages = append(packages, packageInfo)
		}
	}

	return packages
}

// ParseDeletedOutput parses the output of `snap search` command
// and returns a list of PackageInfo
//
// Example output:
// Name                Version  Publisher  Notes  Summary
// blablaland-desktop  1.0.1    adedev     -      Blablaland Desktop
func ParseSearchOutput(msg string, opts *manager.Options) []manager.PackageInfo {
	var packages []manager.PackageInfo

	// remove the last empty line
	msg = strings.TrimSuffix(msg, "\n")
	var lines []string = strings.Split(string(msg), "\n")

	// skip the first line
	for _, line := range lines[1:] {
		if opts.Verbose {
			fmt.Printf("%s: %s", pm, line)
		}
		parts := strings.Fields(line)
		if len(parts) < 5 {
			continue
		}

		// skip the first line (header/title)
		if parts[0] == "Name" {
			continue
		}

		packageInfo := manager.PackageInfo{
			Name:           parts[0],
			Version:        parts[1],
			Status:         manager.PackageStatusAvailable,
			PackageManager: pm,
		}
		packages = append(packages, packageInfo)
	}

	return packages

}

// cspell: disable
// ParsePackageInfoOutput parses the output of `snap info` command
// and returns a list of PackageInfo
//
// Example msg:
// name:      blablaland-desktop
// summary:   Blablaland Desktop
// publisher: AdeDev
// store-url: https://snapcraft.io/blablaland-desktop
// license:   unset
// description: |
//
//	Version bureau du jeu Blablaland (inclus Flash Player)
//
// snap-id: yEfmuhiQDVy5B2rxNLaPyUYOE6iJakwr
// channels:
//
//	latest/stable:    –
//	latest/candidate: –
//	latest/beta:      –
//	latest/edge:      1.0.1 2021-06-08 (3) 112MB -
//
// cspell: enable
func ParsePackageInfoOutput(msg string, opts *manager.Options) manager.PackageInfo {
	var pkg manager.PackageInfo

	// remove the last empty line
	msg = strings.TrimSuffix(msg, "\n")
	lines := strings.Split(string(msg), "\n")

	for _, line := range lines {
		// remove all leading and trailing spaces
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			parts := strings.SplitN(line, ":", 2)

			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			if key == "name" {
				pkg.Name = value
			} else if strings.HasPrefix(key, "latest/") {
				version := strings.Fields(value)[0]
				if pkg.Version == "" {
					pkg.Version = version
				}
			}
		}
	}

	pkg.PackageManager = "apt"

	return pkg
}

// ParseListUpgradableOutput parses the output of `snap refresh --list` command
// and returns a list of PackageInfo
//
// Example msg:
// bluet@ocisly:~/workspace/go-syspkg$ snap refresh --list
// Name             Version                     Rev   Size   Publisher   Notes
// firefox          112.0.1-1                   2579  253MB  mozilla✓    -
// gnome-3-28-1804  3.28.0-19-g98f9e67.98f9e67  198   172MB  canonical✓  -
// bluet@ocisly:~/workspace/go-syspkg$ snap list|grep firefox
// firefox                         112.0-2                     2559   latest/stable    mozilla**               -
func ParseListUpgradableOutput(msg string, opts *manager.Options) []manager.PackageInfo {
	return ParseListOutput(msg, opts)
}

// ParseFindOutput parses the output of `snap search` command
// and returns a list of PackageInfo
//
// Example output:
// Name                Version  Publisher  Notes  Summary
// blablaland-desktop  1.0.1    adedev     -      Blablaland Desktop
func ParseFindOutput(msg string, opts *manager.Options) []manager.PackageInfo {
	return ParseListOutput(msg, opts)
}

// ParseListInstalledOutput parses the output of `snap list` command
// and returns a list of PackageInfo
//
// Example output:
// Name                            Version                     Rev    Tracking         Publisher               Notes
// bare                            1.0                         5      latest/stable    canonical✓              base
// blablaland-desktop              1.0.1                       3      latest/edge      adedev                  -
// canonical-livepatch             10.5.3                      196    latest/stable    canonical✓              -
// caprine                         2.57.0                      53     latest/stable    sindresorhus            -
func ParseListInstalledOutput(msg string, opts *manager.Options) []manager.PackageInfo {
	return ParseListOutput(msg, opts)
}

func ParseListOutput(msg string, opts *manager.Options) []manager.PackageInfo {
	var packages []manager.PackageInfo

	// remove the last empty line
	msg = strings.TrimSuffix(msg, "\n")
	var lines []string = strings.Split(string(msg), "\n")

	for _, line := range lines {
		if opts.Verbose {
			fmt.Printf("%s: %s", pm, line)
		}
		parts := strings.Fields(line)
		if len(parts) < 5 {
			continue
		}

		// skip the first line (header/title)
		if parts[0] == "Name" {
			continue
		}

		packageInfo := manager.PackageInfo{
			Name:           parts[0],
			Version:        parts[1],
			Status:         manager.PackageStatusAvailable,
			PackageManager: pm,
		}
		packages = append(packages, packageInfo)
	}

	return packages
}