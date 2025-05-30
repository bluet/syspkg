// Package yum provides a package manager implementation for RedHat-based systems using
// YUM as the underlying package management tool.
package yum

import (
	"regexp"
	"strings"

	"github.com/bluet/syspkg/manager"
)

// packageLineRegex matches package lines in yum search output (name.arch format)
var packageLineRegex = regexp.MustCompile(`^[\w\d-]+\.[\w\d_]+`)

// ParseFindOutput parses the output of `yum search packageName` command
// and returns a list of available packages that match the search query. It extracts package
// information such as name, architecture from the
// output, and stores them in a list of manager.PackageInfo objects.
//
// The output format is expected to be similar to the following example:
//
//Last metadata expiration check: 0:26:09 ago on Thu 22 May 2025 04:30:18 PM UTC.
// ==================================================Name Exactly Matched: nginx ====================================================
//nginx.x86_64 : A high performance web server and reverse proxy server
//====================================================Name & Summary Matched: nginx==================================================
//nginx-all-modules.noarch : A meta package that installs all available Nginx modules
//nginx-core.x86_64 : nginx minimal core

// The function first removes the "Last Metadata..." and the "========="
// lines, and then processes each package entry line to extract relevant
// information.
//
// The opts parameter is reserved for future parsing options and is currently unused.
func ParseFindOutput(msg string, opts *manager.Options) []manager.PackageInfo {
	var packages []manager.PackageInfo

	// remove the last empty line
	msg = strings.TrimSuffix(msg, "\n")

	// split output by lines
	lines := strings.Split(msg, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "=======") {
			continue
		}
		if strings.HasPrefix(line, "Last metadata") {
			continue
		}
		if packageLineRegex.MatchString(line) {
			parts := strings.Fields(line)

			// if name is empty, it might be not what we want
			if parts[0] == "" {
				continue
			}
			// Find the last dot to separate name and architecture
			lastDotIndex := strings.LastIndex(parts[0], ".")
			if lastDotIndex == -1 {
				// No dot found, skip this line
				continue
			}

			packageInfo := manager.PackageInfo{
				Name:           parts[0][:lastDotIndex],
				Arch:           parts[0][lastDotIndex+1:],
				PackageManager: pm,
			}

			packages = append(packages, packageInfo)
		}
	}

	return packages
}

// ParseListInstalledOutput parses the output of `yum list --installed` command
// and returns a list of installed packages. It extracts the package name, version,
// and architecture from the output and stores them in a list of manager.PackageInfo objects.
//
// The opts parameter is reserved for future parsing options and is currently unused.
func ParseListInstalledOutput(msg string, opts *manager.Options) []manager.PackageInfo {
	var packages []manager.PackageInfo

	// remove the last empty line
	msg = strings.TrimSuffix(msg, "\n")
	lines := strings.Split(msg, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "Installed Packages") {
			continue
		}

		if len(line) > 0 {
			parts := strings.Fields(line)

			// if it doesn't split correctly, or the name is empty, it might be not what we want
			if len(parts) < 2 || parts[0] == "" {
				continue
			}
			// Find the last dot to separate name and architecture
			lastDotIndex := strings.LastIndex(parts[0], ".")
			if lastDotIndex == -1 {
				// No dot found, skip this line
				continue
			}
			name := parts[0][:lastDotIndex]
			arch := parts[0][lastDotIndex+1:]

			packageInfo := manager.PackageInfo{
				Name:           name,
				Version:        parts[1],
				Status:         manager.PackageStatusInstalled,
				Arch:           arch,
				PackageManager: "yum",
			}
			packages = append(packages, packageInfo)
		}
	}

	return packages
}

// ParsePackageInfoOutput parses the output of `yum info packageName` command
// and returns a manager.PackageInfo object containing package information such as name, version,
// architecture, and category. This function is useful for getting detailed package information.
//
// The opts parameter is reserved for future parsing options and is currently unused.
func ParsePackageInfoOutput(msg string, opts *manager.Options) manager.PackageInfo {
	var pkg manager.PackageInfo

	// remove the last empty line
	msg = strings.TrimSuffix(msg, "\n")
	lines := strings.Split(msg, "\n")

	for _, line := range lines {
		if len(line) > 0 {
			parts := strings.SplitN(line, ":", 2)

			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "Name":
				pkg.Name = value
			case "Version":
				pkg.Version = value
			case "Architecture":
				pkg.Arch = value
			}
		}
	}

	pkg.PackageManager = "yum"

	return pkg
}
