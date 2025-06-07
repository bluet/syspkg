// Package yum provides utility functions for parsing YUM command outputs
package yum

import (
	"context"
	"regexp"
	"strings"

	"github.com/bluet/syspkg/manager"
)

// epochRegex matches RPM epoch pattern: -digit:
var epochRegex = regexp.MustCompile(`-(\d+):`)

// packageLineRegex matches package lines in yum search output (name.arch format)
var packageLineRegex = regexp.MustCompile(`^[\w\d-]+\.[\w\d_]+`)

// parseSearchOutput parses the output of `yum search packageName` command
// and returns a list of packages that match the search query.
func parseSearchOutput(output string) []manager.PackageInfo {
	var packagesDict = make(map[string]manager.PackageInfo)

	// remove the last empty line
	output = strings.TrimSuffix(output, "\n")

	// split output by lines
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "=======") {
			continue
		}
		if strings.HasPrefix(line, "Last metadata") {
			continue
		}
		if packageLineRegex.MatchString(line) {
			// Parse format: "package.arch : description"
			colonIndex := strings.Index(line, " : ")
			if colonIndex == -1 {
				// No description separator, skip
				continue
			}

			packagePart := strings.TrimSpace(line[:colonIndex])
			descriptionPart := strings.TrimSpace(line[colonIndex+3:])

			// Find the last dot to separate name and architecture
			lastDotIndex := strings.LastIndex(packagePart, ".")
			if lastDotIndex == -1 {
				// No dot found, skip this line
				continue
			}

			packageName := packagePart[:lastDotIndex]
			arch := packagePart[lastDotIndex+1:]

			packageInfo := manager.NewPackageInfo(
				packageName,             // name
				"",                      // version (empty for search)
				manager.StatusAvailable, // status (will be enhanced later)
				"yum",                   // manager type
			)
			packageInfo.Description = descriptionPart
			packageInfo.Metadata = make(map[string]interface{})
			packageInfo.Metadata["arch"] = arch

			packagesDict[packageName] = packageInfo
		}
	}

	// Convert map to slice and return
	result := make([]manager.PackageInfo, 0, len(packagesDict))
	for _, pkg := range packagesDict {
		result = append(result, pkg)
	}
	return result
}

// parseListOutput parses the output of `yum list` command (installed or updates)
// and returns a list of packages.
func parseListOutput(output string) []manager.PackageInfo {
	var packages []manager.PackageInfo

	// remove the last empty line
	output = strings.TrimSuffix(output, "\n")
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "Installed Packages") {
			continue
		}
		if strings.HasPrefix(line, "Available Upgrades") {
			continue
		}
		if strings.HasPrefix(line, "Last metadata") {
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

			status := manager.StatusInstalled
			if strings.Contains(output, "Available Upgrades") {
				status = manager.StatusUpgradable
			}

			packageInfo := manager.NewPackageInfo(
				name,     // name
				parts[1], // version
				status,   // status
				"yum",    // manager type
			)
			packageInfo.Metadata = make(map[string]interface{})
			packageInfo.Metadata["arch"] = arch
			packages = append(packages, packageInfo)
		}
	}

	return packages
}

// parseInfoOutput parses the output of `yum info packageName` command
// and returns a manager.PackageInfo object containing package information.
func parseInfoOutput(output string, packageName string) (manager.PackageInfo, error) {
	var isInstalled bool

	// remove the last empty line
	output = strings.TrimSuffix(output, "\n")
	lines := strings.Split(output, "\n")

	// Check for error conditions first
	for _, line := range lines {
		if strings.Contains(line, "Error: No matching Packages") ||
			strings.Contains(line, "No matching Packages to list") {
			return manager.PackageInfo{}, manager.ErrPackageNotFound
		}
	}

	name := packageName
	version := ""
	arch := ""
	description := ""
	hasValidData := false

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
				name = value
				hasValidData = true
			case "Version":
				version = value
				hasValidData = true
			case "Architecture":
				arch = value
			case "Summary":
				description = value
			}
		}
	}

	// If we didn't find any valid package data, return not found error
	if !hasValidData || version == "" {
		return manager.PackageInfo{}, manager.ErrPackageNotFound
	}

	status := manager.StatusAvailable
	if isInstalled {
		status = manager.StatusInstalled
	}

	pkg := manager.NewPackageInfo(name, version, status, "yum")
	pkg.Description = description
	pkg.Metadata = make(map[string]interface{})
	pkg.Metadata["arch"] = arch
	return pkg, nil
}

// parseInstallOutput parses the output of `yum install` command
// and returns a list of packages that were installed.
func parseInstallOutput(output string) []manager.PackageInfo {
	var packages []manager.PackageInfo

	lines := strings.Split(output, "\n")
	inInstalledSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for the "Installed:" section
		if line == "Installed:" {
			inInstalledSection = true
			continue
		}

		// Stop parsing when we hit "Complete!" or empty line after packages
		if strings.Contains(line, "Complete!") || (inInstalledSection && line == "") {
			break
		}

		// Parse package lines in the Installed section
		if inInstalledSection && line != "" {
			// Format: package-version.arch
			parts := strings.Fields(line)
			if len(parts) > 0 {
				packageLine := parts[0]

				// Find the last dot to separate arch
				lastDotIndex := strings.LastIndex(packageLine, ".")
				if lastDotIndex == -1 {
					continue
				}

				packageWithVersion := packageLine[:lastDotIndex]
				arch := packageLine[lastDotIndex+1:]

				// Find package name and version by looking for version pattern
				// Try to split on common version separators
				name, version := parsePackageNameVersion(packageWithVersion)

				if name != "" {
					pkg := manager.NewPackageInfo(name, version, manager.StatusInstalled, "yum")
					pkg.Metadata = make(map[string]interface{})
					pkg.Metadata["arch"] = arch
					packages = append(packages, pkg)
				}
			}
		}
	}

	return packages
}

// parseRemoveOutput parses the output of `yum remove` command
// and returns a list of packages that were removed.
// Handles both table format and colon format output.
func parseRemoveOutput(output string) []manager.PackageInfo {
	var packages []manager.PackageInfo

	lines := strings.Split(output, "\n")
	inRemovingSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for "Removing:" header (table format)
		if line == "Removing:" {
			inRemovingSection = true
			continue
		}

		// Stop when we hit section boundaries or transaction summary
		if strings.HasPrefix(line, "Transaction Summary") ||
			strings.HasPrefix(line, "Complete!") ||
			strings.HasPrefix(line, "=====") ||
			(inRemovingSection && line == "") {
			inRemovingSection = false
			continue
		}

		// Parse table format lines in Removing section
		if inRemovingSection && line != "" {
			// Format: " vim-enhanced       x86_64     2:8.0.1763-19.el8_6.4       @appstream     2.9 M"
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				packageName := fields[0]
				arch := fields[1]
				version := fields[2]

				// Clean up package name and version
				name, cleanVersion := parsePackageNameVersion(packageName + "-" + version)
				if name == "" {
					name = packageName
					cleanVersion = version
				}

				pkg := manager.NewPackageInfo(name, cleanVersion, manager.StatusAvailable, "yum")
				pkg.Metadata = make(map[string]interface{})
				pkg.Metadata["arch"] = arch
				packages = append(packages, pkg)
			}
			continue
		}

		// Fallback: Look for "Removing" or "Erasing" lines with colon (legacy format)
		if (strings.HasPrefix(line, "Removing") || strings.HasPrefix(line, "Erasing")) && strings.Contains(line, ":") {
			// Format: "Removing       : package-version.arch"
			colonIndex := strings.Index(line, ":")
			if colonIndex != -1 {
				packagePart := strings.TrimSpace(line[colonIndex+1:])

				// Find the last dot to separate arch
				lastDotIndex := strings.LastIndex(packagePart, ".")
				if lastDotIndex == -1 {
					continue
				}

				packageWithVersion := packagePart[:lastDotIndex]
				arch := packagePart[lastDotIndex+1:]

				name, version := parsePackageNameVersion(packageWithVersion)

				if name != "" {
					pkg := manager.NewPackageInfo(name, version, manager.StatusAvailable, "yum")
					pkg.Metadata = make(map[string]interface{})
					pkg.Metadata["arch"] = arch
					packages = append(packages, pkg)
				}
			}
		}
	}

	return packages
}

// parsePackageNameVersion attempts to separate package name from version
// in the format "package-version-release" or "package-epoch:version-release"
func parsePackageNameVersion(packageWithVersion string) (name, version string) {
	// RPM package format: package-[epoch:]version-release
	// Examples:
	// - vim-enhanced-2:8.0.1763-19.el8_6.4
	// - gpm-libs-1.20.7-17.el8
	// - which-2.21-20.el8

	// Look for epoch pattern: -digit(s):
	epochMatch := epochRegex.FindStringIndex(packageWithVersion)
	if epochMatch != nil {
		// Found epoch - extract components
		epochStart := epochMatch[0] // position of "-" before epoch
		epochEnd := epochMatch[1]   // position after ":"

		// Package name is everything before the epoch
		name = packageWithVersion[:epochStart]

		// Version is epoch + everything after ":"
		epoch := packageWithVersion[epochStart+1 : epochEnd-1] // extract just the digits
		versionPart := packageWithVersion[epochEnd:]
		version = epoch + ":" + versionPart

		return name, version
	}

	// No epoch - split package name from version-release
	// Strategy: find the first part that looks like a version number
	parts := strings.Split(packageWithVersion, "-")

	// Look for version pattern (starts with digit, contains dots/digits)
	versionRegex := regexp.MustCompile(`^\d+(\.\d+)*`)
	for i := 1; i < len(parts); i++ {
		if versionRegex.MatchString(parts[i]) {
			// Found likely version start
			name = strings.Join(parts[:i], "-")
			version = strings.Join(parts[i:], "-")
			return name, version
		}
	}

	// Fallback: use last hyphen if no clear version pattern found
	lastHyphenIndex := strings.LastIndex(packageWithVersion, "-")
	if lastHyphenIndex != -1 {
		name = packageWithVersion[:lastHyphenIndex]
		version = packageWithVersion[lastHyphenIndex+1:]
		return name, version
	}

	// No hyphen found - entire string is name
	return packageWithVersion, ""
}

// parseRpmVersion extracts version from rpm -q output
func parseRpmVersion(output string) string {
	// rpm -q output format: packagename-version-release.arch
	// Example: vim-enhanced-8.0.1763-19.el8_6.4.x86_64
	output = strings.TrimSpace(output)
	lines := strings.Split(output, "\n")
	if len(lines) == 0 {
		return ""
	}

	line := lines[0]

	// First, remove the .arch suffix
	if dotIndex := strings.LastIndex(line, "."); dotIndex != -1 {
		line = line[:dotIndex]
	}

	// Now use parsePackageNameVersion to extract version
	_, version := parsePackageNameVersion(line)
	return version
}

// enhanceWithDetailedStatus enhances packages with detailed status using rpm -q
func (m *Manager) enhanceWithDetailedStatus(packages []manager.PackageInfo) []manager.PackageInfo {
	enhanced := make([]manager.PackageInfo, len(packages))

	for i, pkg := range packages {
		enhanced[i] = pkg

		// Check if package is installed using rpm -q
		result, err := m.GetRunner().Run(context.Background(), "rpm", []string{"-q", pkg.Name})
		if err == nil && result.ExitCode == 0 {
			// Package is installed
			version := parseRpmVersion(string(result.Output))
			enhanced[i].Status = manager.StatusInstalled
			enhanced[i].Version = version

			// Check if there's an available update
			updateResult, updateErr := m.GetRunner().Run(context.Background(), "yum", []string{"list", "updates", pkg.Name})
			if updateErr == nil && updateResult.ExitCode == 0 && strings.Contains(string(updateResult.Output), pkg.Name) {
				enhanced[i].Status = manager.StatusUpgradable
				// Parse available version would require more complex parsing
			}
		}
		// If rpm -q fails, package is not installed (status remains Available)
	}

	return enhanced
}
