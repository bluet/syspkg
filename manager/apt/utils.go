// Package apt provides utility functions for parsing APT command outputs
package apt

import (
	"regexp"
	"strings"

	"github.com/bluet/syspkg/manager"
)

// parseSearchOutput parses the output of `apt search` command
func parseSearchOutput(output string) []manager.PackageInfo {
	var packages []manager.PackageInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.Contains(line, "/") && !strings.HasPrefix(line, "Sorting") && !strings.HasPrefix(line, "Full Text") && !strings.HasPrefix(line, "WARNING") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				nameRepo := strings.Split(parts[0], "/")
				if len(nameRepo) >= 2 {
					// Default to available
					status := manager.StatusAvailable

					// Check if package is installed (indicated by "now" in repository list after "/")
					repoList := nameRepo[1] // repository list after the "/"
					if strings.Contains(repoList, "now") {
						status = manager.StatusInstalled
					}

					// Check for additional status indicators in brackets
					if len(parts) >= 4 && strings.Contains(line, "[") {
						statusField := line[strings.Index(line, "["):]
						if strings.Contains(statusField, "upgradable") {
							status = manager.StatusUpgradable
						}
					}

					pkg := manager.NewPackageInfo(nameRepo[0], parts[1], status, manager.TypeSystem)
					pkg.NewVersion = parts[1]
					pkg.Category = strings.Split(nameRepo[1], ",")[0] // Remove ",now" suffix
					pkg.Metadata["arch"] = parts[2]

					// Add description if available on next line
					if len(lines) > 0 {
						// Look ahead for description line
						for i, currentLine := range lines {
							if currentLine == line && i+1 < len(lines) {
								nextLine := strings.TrimSpace(lines[i+1])
								if nextLine != "" && !strings.Contains(nextLine, "/") {
									pkg.Description = nextLine
								}
								break
							}
						}
					}

					packages = append(packages, pkg)
				}
			}
		}
	}

	return packages
}

// parseInstallOutput parses the output of `apt install` command
func parseInstallOutput(output string) []manager.PackageInfo {
	var packages []manager.PackageInfo
	lines := strings.Split(output, "\n")

	// Look for "Setting up" lines which indicate successful installation
	versionRegex := regexp.MustCompile(`Setting up ([^\s]+) \(([^)]+)\)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if matches := versionRegex.FindStringSubmatch(line); matches != nil {
			packageName := matches[1]
			version := matches[2]

			pkg := manager.NewPackageInfo(packageName, version, manager.StatusInstalled, manager.TypeSystem)
			pkg.Metadata = make(map[string]interface{})
			packages = append(packages, pkg)
		}
	}

	return packages
}

// parseRemoveOutput parses the output of `apt remove` command
func parseRemoveOutput(output string) []manager.PackageInfo {
	var packages []manager.PackageInfo
	lines := strings.Split(output, "\n")

	// Look for "Removing" lines which indicate successful removal
	versionRegex := regexp.MustCompile(`Removing ([^\s]+) \(([^)]+)\)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if matches := versionRegex.FindStringSubmatch(line); matches != nil {
			packageName := matches[1]
			version := matches[2]

			pkg := manager.NewPackageInfo(packageName, version, manager.StatusAvailable, manager.TypeSystem)
			pkg.Metadata = make(map[string]interface{})
			packages = append(packages, pkg)
		}
	}

	return packages
}

// parseListUpgradableOutput parses the output of `apt list --upgradable` command
func parseListUpgradableOutput(output string) []manager.PackageInfo {
	var packages []manager.PackageInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip header and empty lines
		if strings.HasPrefix(line, "Listing") || strings.HasPrefix(line, "WARNING") || line == "" {
			continue
		}

		// Parse format: "package/repo version arch [upgradable from: old_version]"
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		// Extract package name
		nameRepo := strings.Split(parts[0], "/")
		if len(nameRepo) < 1 {
			continue
		}
		packageName := nameRepo[0]

		// Extract new version and architecture
		newVersion := parts[1]
		arch := parts[2]

		// Extract old version from [upgradable from: old_version]
		upgradeMatch := regexp.MustCompile(`\[upgradable from: ([^\]]+)\]`).FindStringSubmatch(line)
		oldVersion := ""
		if len(upgradeMatch) > 1 {
			oldVersion = upgradeMatch[1]
		}

		pkg := manager.NewPackageInfo(packageName, oldVersion, manager.StatusUpgradable, manager.TypeSystem)
		pkg.NewVersion = newVersion
		pkg.Metadata = make(map[string]interface{})
		pkg.Metadata["arch"] = arch

		packages = append(packages, pkg)
	}

	return packages
}

// ParsePackageInfo parses the output of `apt show` or `apt-cache show` command
func ParsePackageInfo(output string) manager.PackageInfo {
	pkg := manager.NewPackageInfo("", "", manager.StatusUnknown, manager.TypeSystem)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				switch key {
				case "Package":
					pkg.Name = value
				case "Version":
					pkg.Version = value
				case "Architecture":
					pkg.Metadata["arch"] = value
				case "Section":
					pkg.Category = value
				case "Description":
					pkg.Description = value
				case "Installed-Size":
					pkg.Metadata["installed_size"] = value
				case "Maintainer":
					pkg.Metadata["maintainer"] = value
				}
			}
		}
	}

	// Determine status based on whether we have version info
	if pkg.Version != "" {
		pkg.Status = manager.StatusInstalled
	} else {
		pkg.Status = manager.StatusAvailable
	}

	return pkg
}
