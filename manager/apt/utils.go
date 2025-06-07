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

					pkg := manager.NewPackageInfo(nameRepo[0], parts[1], status, "apt")
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

			pkg := manager.NewPackageInfo(packageName, version, manager.StatusInstalled, "apt")
			pkg.Metadata = make(map[string]interface{})
			packages = append(packages, pkg)
		}
	}

	return packages
}

// parseRemoveOutput parses the output of `apt remove` command
// Handles both actual removal output ("Removing package (version)") and
// dry-run output ("The following packages will be REMOVED:")
func parseRemoveOutput(output string) []manager.PackageInfo {
	var packages []manager.PackageInfo
	lines := strings.Split(output, "\n")

	// Look for "Removing package:arch (version) ..." lines (actual removal)
	removingRegex := regexp.MustCompile(`Removing ([^:]+)(?::([^:]+))? \(([^)]+)\)`)

	for _, line := range lines {
		if match := removingRegex.FindStringSubmatch(line); match != nil {
			name := match[1]
			arch := match[2]
			version := match[3]

			pkg := manager.NewPackageInfo(name, version, manager.StatusAvailable, "apt")

			if arch != "" {
				pkg.Metadata["arch"] = arch
			}

			packages = append(packages, pkg)
		}
	}

	// If no "Removing" lines found, try to parse from "The following packages will be REMOVED:" (dry-run)
	if len(packages) == 0 {
		inRemoveSection := false
		for _, line := range lines {
			if strings.Contains(line, "The following packages will be REMOVED:") {
				inRemoveSection = true
				continue
			}

			if inRemoveSection && strings.TrimSpace(line) != "" && !strings.Contains(line, "upgraded") {
				// Parse package names from the removal list
				packageNames := strings.Fields(line)
				for _, name := range packageNames {
					// Clean up package name (remove any special characters)
					cleanName := strings.Trim(name, " \t")
					if cleanName != "" && !strings.Contains(cleanName, "operation") && !strings.Contains(cleanName, "newly") {
						pkg := manager.NewPackageInfo(cleanName, "", manager.StatusAvailable, "apt")
						packages = append(packages, pkg)
					}
				}
			}

			// Stop at summary lines
			if strings.Contains(line, "upgraded") || strings.Contains(line, "After this operation") {
				break
			}
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

		pkg := manager.NewPackageInfo(packageName, oldVersion, manager.StatusUpgradable, "apt")
		pkg.NewVersion = newVersion
		pkg.Metadata = make(map[string]interface{})
		pkg.Metadata["arch"] = arch

		packages = append(packages, pkg)
	}

	return packages
}

// ParsePackageInfo parses the output of `apt show` or `apt-cache show` command
func ParsePackageInfo(output string) manager.PackageInfo {
	pkg := manager.NewPackageInfo("", "", manager.StatusUnknown, "apt")

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
