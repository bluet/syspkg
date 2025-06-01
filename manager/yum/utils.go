// Package yum provides a package manager implementation for RedHat-based systems using
// YUM as the underlying package management tool.
package yum

import (
	"regexp"
	"strings"

	"github.com/bluet/syspkg/manager"
)

// epochRegex matches RPM epoch pattern: -digit:
var epochRegex = regexp.MustCompile(`-(\d+):`)

// packageLineRegex matches package lines in yum search output (name.arch format)
var packageLineRegex = regexp.MustCompile(`^[\w\d-]+\.[\w\d_]+`)

// ParseFindOutput parses the output of `yum search packageName` command
// and returns a list of packages that match the search query.
//
// This function performs pure parsing of yum search output without making any system calls.
// Status detection is handled separately by the calling function using rpm -q integration.
//
// The output format is expected to be similar to the following example:
//
//	Last metadata expiration check: 0:26:09 ago on Thu 22 May 2025 04:30:18 PM UTC.
//	==================================================Name Exactly Matched: nginx ====================================================
//	nginx.x86_64 : A high performance web server and reverse proxy server
//	====================================================Name & Summary Matched: nginx==================================================
//	nginx-all-modules.noarch : A meta package that installs all available Nginx modules
//	nginx-core.x86_64 : nginx minimal core
//
// Returned PackageInfo fields:
//   - Name: Package name (e.g., "nginx")
//   - Arch: Architecture (e.g., "x86_64")
//   - Status: PackageStatusAvailable (default - enhanced by calling function)
//   - Version: Empty (will be set by status enhancement if installed)
//   - NewVersion: Empty (search doesn't provide repo version)
//   - PackageManager: "yum"
//
// The opts parameter is reserved for future parsing options and is currently unused.
func ParseFindOutput(msg string, opts *manager.Options) []manager.PackageInfo {
	var packagesDict = make(map[string]manager.PackageInfo)

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
				Status:         manager.PackageStatusAvailable, // Will be updated by getYumPackageStatus
				PackageManager: pm,
			}

			packagesDict[packageInfo.Name] = packageInfo
		}
	}

	// Convert map to slice and return
	result := make([]manager.PackageInfo, 0, len(packagesDict))
	for _, pkg := range packagesDict {
		result = append(result, pkg)
	}
	return result
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
// architecture, and status. This function determines installation status based on whether
// the package appears under "Installed Packages" or "Available Packages" section.
//
// Expected output format:
//
//	Installed Packages
//	Name         : package-name
//	Version      : 1.0.0
//	...
//
// OR:
//
//	Available Packages
//	Name         : package-name
//	Version      : 1.0.0
//	...
//
// The opts parameter is reserved for future parsing options and is currently unused.
func ParsePackageInfoOutput(msg string, opts *manager.Options) manager.PackageInfo {
	var pkg manager.PackageInfo
	var isInstalled bool

	// remove the last empty line
	msg = strings.TrimSuffix(msg, "\n")
	lines := strings.Split(msg, "\n")

	for _, line := range lines {
		// Check for section headers to determine status
		if strings.HasPrefix(line, "Installed Packages") {
			isInstalled = true
			continue
		}
		if strings.HasPrefix(line, "Available Packages") {
			isInstalled = false
			continue
		}

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

	// Set status based on which section the package was found in
	if isInstalled {
		pkg.Status = manager.PackageStatusInstalled
	} else {
		pkg.Status = manager.PackageStatusAvailable
	}

	return pkg
}

// ParseInstallOutput parses the output of `yum install -y packageName` command
// and returns a list of successfully installed packages. It extracts package
// information from both the "Installing:" and "Installing dependencies:" sections.
//
// Expected output format:
//
//	Installing:
//	 vim-enhanced       x86_64     2:8.0.1763-19.el8_6.4        appstream     1.4 M
//	Installing dependencies:
//	 vim-common         x86_64     2:8.0.1763-19.el8_6.4        appstream     6.3 M
//	...
//	Installed:
//	  vim-enhanced-2:8.0.1763-19.el8_6.4.x86_64
//	  vim-common-2:8.0.1763-19.el8_6.4.x86_64
//
// Returns all installed packages with Status=installed.
// The opts parameter is reserved for future parsing options and is currently unused.
func ParseInstallOutput(msg string, opts *manager.Options) []manager.PackageInfo {
	var packages []manager.PackageInfo

	// remove the last empty line
	msg = strings.TrimSuffix(msg, "\n")
	lines := strings.Split(msg, "\n")

	installedSection := false
	for _, line := range lines {
		// Look for the "Installed:" section which contains the final list
		if strings.HasPrefix(line, "Installed:") {
			installedSection = true
			continue
		}

		// Parse packages from the Installed: section
		if installedSection {
			line = strings.TrimSpace(line)
			if line == "" || line == "Complete!" {
				break
			}

			// Parse package-version-arch format: vim-enhanced-2:8.0.1763-19.el8_6.4.x86_64
			if strings.Contains(line, "-") && strings.Contains(line, ".") {
				// Find the last dot to separate arch
				lastDotIndex := strings.LastIndex(line, ".")
				if lastDotIndex == -1 {
					continue
				}

				arch := line[lastDotIndex+1:]
				nameVersion := line[:lastDotIndex]

				// Find package name by looking for version pattern (epoch or version)
				var name, version string
				if matches := epochRegex.FindStringIndex(nameVersion); matches != nil {
					name = nameVersion[:matches[0]]
					version = nameVersion[matches[0]+1:]
				} else if versionIndex := strings.LastIndex(nameVersion, "-"); versionIndex != -1 {
					name = nameVersion[:versionIndex]
					version = nameVersion[versionIndex+1:]
				} else {
					// Fallback: treat entire thing as name
					name = nameVersion
					version = ""
				}

				packageInfo := manager.PackageInfo{
					Name:           name,
					Version:        version,
					NewVersion:     version, // For install, new version equals installed version
					Status:         manager.PackageStatusInstalled,
					Arch:           arch,
					PackageManager: "yum",
				}
				packages = append(packages, packageInfo)
			}
		}
	}

	return packages
}

// ParseDeleteOutput parses the output of `yum remove -y packageName` command
// and returns a list of successfully removed packages.
//
// Expected output format:
//
//	Removing:
//	 tree           x86_64           1.7.0-15.el8           @baseos           106 k
//	...
//	Removed:
//	  tree-1.7.0-15.el8.x86_64
//
// Returns all removed packages with Status=available.
// The opts parameter is reserved for future parsing options and is currently unused.
func ParseDeleteOutput(msg string, opts *manager.Options) []manager.PackageInfo {
	var packages []manager.PackageInfo

	// remove the last empty line
	msg = strings.TrimSuffix(msg, "\n")
	lines := strings.Split(msg, "\n")

	removedSection := false
	for _, line := range lines {
		// Look for the "Removed:" section which contains the final list
		if strings.HasPrefix(line, "Removed:") {
			removedSection = true
			continue
		}

		// Parse packages from the Removed: section
		if removedSection {
			line = strings.TrimSpace(line)
			if line == "" || line == "Complete!" {
				break
			}

			// Parse package-version-arch format: tree-1.7.0-15.el8.x86_64
			if strings.Contains(line, "-") && strings.Contains(line, ".") {
				// Find the last dot to separate arch
				lastDotIndex := strings.LastIndex(line, ".")
				if lastDotIndex == -1 {
					continue
				}

				arch := line[lastDotIndex+1:]
				nameVersion := line[:lastDotIndex]

				// Find package name by looking for version pattern
				var name, version string
				if versionIndex := strings.LastIndex(nameVersion, "-"); versionIndex != -1 {
					name = nameVersion[:versionIndex]
					version = nameVersion[versionIndex+1:]
				} else {
					// Fallback: treat entire thing as name
					name = nameVersion
					version = ""
				}

				packageInfo := manager.PackageInfo{
					Name:           name,
					Version:        version,
					NewVersion:     "", // For delete, new version is empty
					Status:         manager.PackageStatusAvailable,
					Arch:           arch,
					PackageManager: "yum",
				}
				packages = append(packages, packageInfo)
			}
		}
	}

	return packages
}

// ParseListUpgradableOutput parses the output of `yum check-update` command
// and returns a list of packages that have updates available.
//
// Expected output format:
//
//	Last metadata expiration check: 0:05:23 ago on Sat May 31 10:00:00 2025.
//
//	kernel.x86_64                    4.18.0-477.27.1.el8_8      baseos
//	vim-common.x86_64                2:8.0.1763-19.el8_6.4      appstream
//
// Returns packages with Status=upgradable, NewVersion=available version.
// The opts parameter is reserved for future parsing options and is currently unused.
func ParseListUpgradableOutput(msg string, opts *manager.Options) []manager.PackageInfo {
	var packages []manager.PackageInfo

	// remove the last empty line
	msg = strings.TrimSuffix(msg, "\n")
	lines := strings.Split(msg, "\n")

	for _, line := range lines {
		// Skip metadata expiration and empty lines
		if strings.HasPrefix(line, "Last metadata") || strings.TrimSpace(line) == "" {
			continue
		}

		// Parse package lines: package.arch version repository
		parts := strings.Fields(line)
		if len(parts) >= 3 && strings.Contains(parts[0], ".") {
			// Find the last dot to separate name and architecture
			lastDotIndex := strings.LastIndex(parts[0], ".")
			if lastDotIndex == -1 {
				continue
			}

			name := parts[0][:lastDotIndex]
			arch := parts[0][lastDotIndex+1:]
			newVersion := parts[1]

			packageInfo := manager.PackageInfo{
				Name:           name,
				Version:        "", // Current version not provided by yum check-update
				NewVersion:     newVersion,
				Status:         manager.PackageStatusUpgradable,
				Arch:           arch,
				PackageManager: "yum",
			}
			packages = append(packages, packageInfo)
		}
	}

	return packages
}

// ParseUpgradeOutput parses the output of `yum update -y packageName` command
// and returns a list of successfully upgraded packages.
//
// Expected output format is similar to install, but with "Upgrading:" section:
//
//	Upgrading:
//	 vim-common      x86_64    2:8.0.1763-19.el8_6.4  appstream              6.3 M
//	...
//	Upgraded:
//	  vim-common-2:8.0.1763-19.el8_6.4.x86_64
//
// Returns all upgraded packages with new version information.
// The opts parameter is reserved for future parsing options and is currently unused.
func ParseUpgradeOutput(msg string, opts *manager.Options) []manager.PackageInfo {
	// TODO: YUM upgrade operations currently reuse install output parsing.
	// Limitation: Version transitions (old→new) are not captured from upgrade output.
	// Future enhancement: Parse 'Upgrading:' section to capture version transitions.
	// Current behavior: Shows final installed version only (like Install operation).

	// Upgrade output format is very similar to install output,
	// we can reuse the same parser logic
	return ParseInstallOutput(msg, opts)
}

// ParseAutoRemoveOutput parses the output of `yum autoremove -y` command
// and returns a list of successfully removed packages.
//
// Expected output format:
//
//	Removing:
//	 perl-IO-Socket-IP      noarch    0.39-5.el8         @appstream           99 k
//	Removing unused dependencies:
//	 perl-IO-Socket-SSL     noarch    2.066-4.module     @appstream          618 k
//	...
//	Removed:
//	  perl-IO-Socket-IP-0.39-5.el8.noarch
//	  perl-IO-Socket-SSL-2.066-4.module.noarch
//
// Returns all removed packages with Status=available.
// The opts parameter is reserved for future parsing options and is currently unused.
func ParseAutoRemoveOutput(msg string, opts *manager.Options) []manager.PackageInfo {
	// AutoRemove output format is the same as regular remove output,
	// we can reuse the same parser logic
	return ParseDeleteOutput(msg, opts)
}

// checkRpmInstallationStatus uses rpm -q to check which packages are installed
// Returns a map of installed package names to their PackageInfo using the provided CommandRunner
func checkRpmInstallationStatus(packageNames []string, runner manager.CommandRunner) (map[string]manager.PackageInfo, error) {
	installedPackages := make(map[string]manager.PackageInfo)

	// Check if rpm command is available by trying to run rpm --version
	_, err := runner.Output("rpm", "--version")
	if err != nil {
		return nil, err
	}

	// Check each package individually with rpm -q
	// Using individual queries because rpm -q with multiple packages can be unreliable
	for _, pkgName := range packageNames {
		out, err := runner.Output("rpm", "-q", pkgName)

		if err != nil {
			// rpm -q returns exit code 1 for packages that are not installed
			// This is normal and not an error - continue to next package
			continue
		}

		// Parse rpm output to extract version
		// Format: package-version-release.arch
		output := strings.TrimSpace(string(out))
		if output != "" {
			// Extract version information from rpm output
			version := extractVersionFromRpmOutput(output, pkgName)
			installedPackages[pkgName] = manager.PackageInfo{
				Name:           pkgName,
				Version:        version,
				Status:         manager.PackageStatusInstalled,
				PackageManager: "yum",
			}
		}
	}

	return installedPackages, nil
}

// extractVersionFromRpmOutput extracts version from rpm -q output
// Input format: package-version-release.arch (e.g., "vim-enhanced-8.0.1763-19.el8_6.4.x86_64")
// Returns: version-release (e.g., "8.0.1763-19.el8_6.4")
func extractVersionFromRpmOutput(rpmOutput, packageName string) string {
	// Remove package name prefix and arch suffix
	if strings.HasPrefix(rpmOutput, packageName+"-") {
		withoutPrefix := rpmOutput[len(packageName)+1:]

		// Find last dot to remove architecture
		lastDot := strings.LastIndex(withoutPrefix, ".")
		if lastDot > 0 {
			return withoutPrefix[:lastDot]
		}
		return withoutPrefix
	}

	// Fallback: return the whole output if parsing fails
	return rpmOutput
}
