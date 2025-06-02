// Package apt provides a package manager implementation for Debian-based systems using
// Advanced Package Tool (APT) as the underlying package management tool.
package apt

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"

	// "github.com/rs/zerolog"
	// "github.com/rs/zerolog/log"

	"github.com/bluet/syspkg/manager"
)

// removeRegex matches APT remove output lines to extract package information
var removeRegex = regexp.MustCompile(`^Removing\s+(\S+?)(?::(\S+))?\s+\(([^)]+)\)`)

// ParseInstallOutput parses the output of `apt install packageName` command and returns a list of installed packages.
// It extracts the package name, package architecture, and version from the lines that start with "Setting up ".
// Example msg:
//
//	Preparing to unpack .../openssl_3.0.2-0ubuntu1.9_amd64.deb ...
//	Unpacking openssl (3.0.2-0ubuntu1.9) over (3.0.2-0ubuntu1.8) ...
//	Setting up libssl3:amd64 (3.0.2-0ubuntu1.9) ...
//	Setting up libssl3:i386 (3.0.2-0ubuntu1.9) ...
//	Setting up libssl-dev:amd64 (3.0.2-0ubuntu1.9) ...
//	Setting up openssl (3.0.2-0ubuntu1.9) ...
//	Processing triggers for man-db (2.10.2-1) ...
//	Processing triggers for libc-bin (2.35-0ubuntu3.1) ...
func ParseInstallOutput(msg string, opts *manager.Options) []manager.PackageInfo {
	var packages []manager.PackageInfo

	// remove the last empty line
	msg = strings.TrimSuffix(msg, "\n")
	var lines []string = strings.Split(string(msg), "\n")

	packageInfoPattern := regexp.MustCompile(`Setting up ([\w\d.-]+):?([\w\d]+)? \(([\w\d\.-]+)\)`)

	for _, line := range lines {
		if opts.Verbose {
			log.Printf("apt: %s", line)
		}

		match := packageInfoPattern.FindStringSubmatch(line)

		if len(match) == 4 {
			name := match[1]
			arch := strings.TrimPrefix(match[2], ":") // Remove the colon prefix from the architecture
			version := match[3]

			// if name is empty, it might be not what we want
			if name == "" {
				continue
			}

			packageInfo := manager.PackageInfo{
				Name:           name,
				Arch:           arch,
				Version:        version,
				NewVersion:     version,
				Status:         manager.PackageStatusInstalled,
				PackageManager: pm,
			}
			packages = append(packages, packageInfo)
		}
	}

	return packages
}

// ParseDeletedOutput parses the output of `apt remove packageName` command
// and returns a list of removed packages.
func ParseDeletedOutput(msg string, opts *manager.Options) []manager.PackageInfo {
	var packages []manager.PackageInfo

	// Normalize line endings (handle both Unix \n and Windows \r\n)
	msg = strings.ReplaceAll(msg, "\r\n", "\n")
	msg = strings.TrimSuffix(msg, "\n")
	var lines []string = strings.Split(msg, "\n")

	for _, rawLine := range lines {
		// Normalize whitespace (drops CR and any leading/trailing whitespace)
		line := strings.TrimSpace(rawLine)

		if opts.Verbose {
			log.Printf("apt: %s", line)
		}

		// Use regex for robust parsing of "Removing package:arch (version) ..." lines
		if strings.HasPrefix(line, "Removing") {
			// Regex handles both "package (version)" and "package:arch (version)" formats
			if match := removeRegex.FindStringSubmatch(line); match != nil {
				name := match[1]
				arch := match[2] // May be empty if no architecture specified
				version := match[3]

				if opts.Verbose {
					log.Printf("apt: parsed - name: %s, arch: %s, version: %s", name, arch, version)
				}

				// Skip if name is empty
				if name == "" {
					continue
				}

				packageInfo := manager.PackageInfo{
					Name:           name,
					Version:        version,
					NewVersion:     "",
					Category:       "",
					Arch:           arch,
					Status:         manager.PackageStatusAvailable,
					PackageManager: pm,
				}
				packages = append(packages, packageInfo)
			}
		}
	}

	return packages
}

// ParseFindOutput parses the output of `apt search packageName` command
// and returns a list of packages that match the search query with their installation status.
//
// This method performs two operations:
// 1. Parses APT search output to extract package information
// 2. Checks installation status via dpkg-query for each found package
//
// Expected APT search output format:
//
//	Sorting...
//	Full Text Search...
//	zutty/jammy 0.11.2.20220109.192032+dfsg1-1 amd64
//	Efficient full-featured X11 terminal emulator
//	zvbi/jammy 0.2.35-19 amd64
//	Vertical Blanking Interval (VBI) utilities
//
// Returned PackageInfo status will be:
//   - installed: Package is currently installed (dpkg-query returns "installed")
//   - available: Package exists in repos but not installed (dpkg-query not found or "not-installed")
//   - upgradable: Package installed but newer version available (handled elsewhere)
//
// Version field usage:
//   - installed packages: Version=installed_version, NewVersion=repo_version
//   - available packages: Version="", NewVersion=repo_version
func (a *PackageManager) ParseFindOutput(msg string, opts *manager.Options) []manager.PackageInfo {
	var packages []manager.PackageInfo
	var packagesDict = make(map[string]manager.PackageInfo)

	msg = strings.TrimPrefix(msg, "Sorting...\nFull Text Search...\n")

	// remove the last empty line
	msg = strings.TrimSuffix(msg, "\n")

	// split output by empty lines
	var lines []string = strings.Split(msg, "\n\n")

	for _, line := range lines {
		if regexp.MustCompile(`^[^\s]+/[^\s]+`).MatchString(line) {
			parts := strings.Fields(line)

			// if name is empty, it might be not what we want
			if parts[0] == "" || len(parts) < 3 {
				continue
			}

			// Parse package name and category safely
			nameParts := strings.Split(parts[0], "/")
			name := nameParts[0]
			var category string
			if len(nameParts) > 1 {
				category = nameParts[1]
			}

			packageInfo := manager.PackageInfo{
				Name:           name,
				Version:        "",
				NewVersion:     parts[1],
				Category:       category,
				Arch:           parts[2],
				PackageManager: pm,
			}

			packagesDict[packageInfo.Name] = packageInfo
		}
	}

	if len(packagesDict) == 0 {
		return packages
	}

	packages, err := a.getPackageStatus(packagesDict, opts)
	if err != nil {
		log.Printf("apt: getPackageStatus error: %s\n", err)
	}

	return packages
}

// ParseListInstalledOutput parses the output of `dpkg-query -W -f '${binary:Package} ${Version}\n'` command
// and returns a list of installed packages. It extracts the package name, version,
// and architecture from the output and stores them in a list of manager.PackageInfo objects.
func ParseListInstalledOutput(msg string, opts *manager.Options) []manager.PackageInfo {
	var packages []manager.PackageInfo

	// remove the last empty line
	msg = strings.TrimSuffix(msg, "\n")
	lines := strings.Split(string(msg), "\n")

	for _, line := range lines {
		if len(line) > 0 {
			parts := strings.Fields(line)

			// Validate minimum required fields
			if len(parts) < 2 || parts[0] == "" {
				continue
			}
			var name, arch string
			if strings.Contains(parts[0], ":") {
				archParts := strings.Split(parts[0], ":")
				name = archParts[0]
				if len(archParts) > 1 {
					arch = archParts[1]
				}
			} else {
				name = parts[0]
			}

			packageInfo := manager.PackageInfo{
				Name:           name,
				Version:        parts[1],
				Status:         manager.PackageStatusInstalled,
				Arch:           arch,
				PackageManager: pm,
			}
			packages = append(packages, packageInfo)
		}
	}

	return packages
}

// ParseListUpgradableOutput parses the output of `apt list --upgradable` command
// and returns a list of upgradable packages. It extracts the package name, version, new version,
// category, and architecture from the output and stores them in a list of manager.PackageInfo objects.
func ParseListUpgradableOutput(msg string, opts *manager.Options) []manager.PackageInfo {
	var packages []manager.PackageInfo

	// Listing...
	// cloudflared/unknown 2023.4.0 amd64 [upgradable from: 2023.3.1]
	// libllvm15/jammy-updates 1:15.0.7-0ubuntu0.22.04.1 amd64 [upgradable from: 1:15.0.6-3~ubuntu0.22.04.2]
	// libllvm15/jammy-updates 1:15.0.7-0ubuntu0.22.04.1 i386 [upgradable from: 1:15.0.6-3~ubuntu0.22.04.2]

	// remove the last empty line
	msg = strings.TrimSuffix(msg, "\n")
	lines := strings.Split(string(msg), "\n")

	for _, line := range lines {
		// skip if line starts with "Listing..."
		if strings.HasPrefix(line, "Listing...") {
			continue
		}

		if len(line) > 0 {
			parts := strings.Fields(line)
			// log.Printf("apt: parts: %+v", parts)

			// Validate minimum required fields for upgradable format
			if len(parts) < 6 {
				continue // Skip malformed lines
			}

			name := strings.Split(parts[0], "/")[0]
			category := strings.Split(parts[0], "/")[1]
			newVersion := parts[1]
			arch := parts[2]
			version := parts[5]
			version = strings.TrimSuffix(version, "]")

			packageInfo := manager.PackageInfo{
				Name:           name,
				Version:        version,
				NewVersion:     newVersion,
				Category:       category,
				Arch:           arch,
				Status:         manager.PackageStatusUpgradable,
				PackageManager: pm,
			}
			packages = append(packages, packageInfo)
		}
	}

	return packages
}

// logDebugPackages logs debug information about input packages
func logDebugPackages(packages map[string]manager.PackageInfo, opts *manager.Options) {
	if opts != nil && opts.Debug {
		log.Printf("getPackageStatus: received %d packages", len(packages))
		for name, pkg := range packages {
			log.Printf("Input package: %s -> %+v", name, pkg)
		}
	}
}

// runDpkgQuery executes dpkg-query command and handles errors appropriately
func (a *PackageManager) runDpkgQuery(packageNames []string, opts *manager.Options) ([]byte, error) {
	// Validate package names to prevent command injection
	if err := manager.ValidatePackageNames(packageNames); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	args := []string{"-W", "--showformat", "${binary:Package} ${Status} ${Version}\n"}
	args = append(args, packageNames...)

	// Use CommandRunner with automatic LC_ALL=C and additional APT env vars

	if opts != nil && opts.Debug {
		log.Printf("Running dpkg-query with args: %v", args)
	}

	out, err := a.getRunner().RunContext(ctx, dpkgQueryCmd, args, aptNonInteractiveEnv...)
	if err != nil {
		if opts != nil && opts.Debug {
			log.Printf("dpkg-query error: %v, output: %q", err, string(out))
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() != 1 && !strings.Contains(string(out), "no packages found matching") {
				return nil, fmt.Errorf("command failed with output: %s", string(out))
			}
		}
	}

	if opts != nil && opts.Debug {
		log.Printf("dpkg-query output: %q", string(out))
	}

	return out, nil
}

// addUnprocessedPackages adds packages that weren't found by dpkg-query with status available
func addUnprocessedPackages(packagesList []manager.PackageInfo, packages map[string]manager.PackageInfo, opts *manager.Options) []manager.PackageInfo {
	for _, pkg := range packages {
		// These are packages that weren't processed by dpkg-query (not installed)
		// They were found in APT search, so they are available for installation
		pkg.Status = manager.PackageStatusAvailable
		if opts != nil && opts.Debug {
			log.Printf("Adding unprocessed package: %+v", pkg)
		}
		packagesList = append(packagesList, pkg)
	}
	return packagesList
}

// getPackageStatus takes a map of package names and manager.PackageInfo objects, and returns a list
// of manager.PackageInfo objects with their statuses updated using the output of `dpkg-query` command.
// It also adds any packages not found by dpkg-query to the list; their status is initially set to unknown,
// but then converted to available for cross-package manager compatibility.
func (a *PackageManager) getPackageStatus(packages map[string]manager.PackageInfo, opts *manager.Options) ([]manager.PackageInfo, error) {
	var packageNames []string
	var packagesList []manager.PackageInfo

	if len(packages) == 0 {
		return packagesList, nil
	}

	logDebugPackages(packages, opts)

	for name := range packages {
		packageNames = append(packageNames, name)
	}

	// Sort package names to ensure deterministic output order
	sort.Strings(packageNames)

	out, err := a.runDpkgQuery(packageNames, opts)
	if err != nil {
		return nil, err
	}

	packagesList, err = ParseDpkgQueryOutput(out, packages, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dpkg-query output: %+v", err)
	}

	// For packages found in APT search but not installed, change status from unknown to available
	for i := range packagesList {
		if packagesList[i].Status == manager.PackageStatusUnknown {
			packagesList[i].Status = manager.PackageStatusAvailable
		}
	}

	if opts != nil && opts.Debug {
		log.Printf("After ParseDpkgQueryOutput: packagesList=%+v, remaining packages=%+v", packagesList, packages)
	}

	packagesList = addUnprocessedPackages(packagesList, packages, opts)

	return packagesList, nil
}

// ParseDpkgQueryOutput parses the output of `dpkg-query` command and updates the status
// and version of the packages in the provided map of package names and manager.PackageInfo objects.
// It returns a list of manager.PackageInfo objects with their statuses and versions updated.
func ParseDpkgQueryOutput(output []byte, packages map[string]manager.PackageInfo, opts *manager.Options) ([]manager.PackageInfo, error) {
	packagesList := make([]manager.PackageInfo, 0)

	// Handle nil packages map
	if packages == nil {
		packages = make(map[string]manager.PackageInfo)
	}

	// remove the last empty line
	output = bytes.TrimSuffix(output, []byte("\n"))
	lines := bytes.Split(output, []byte("\n"))

	for _, line := range lines {
		if len(line) > 0 {
			parts := bytes.Fields(line)

			if len(parts) < 2 {
				continue
			}

			name := string(parts[0])

			if strings.HasPrefix(name, "dpkg-query:") {
				name = string(parts[len(parts)-1])
			}

			if strings.Contains(name, ":") {
				name = strings.Split(name, ":")[0]
			}

			// if name is empty, it might not be what we want
			if name == "" {
				continue
			}

			version := string(parts[len(parts)-1])
			if !regexp.MustCompile(`^\d`).MatchString(version) {
				version = ""
			}

			pkg, ok := packages[name]

			if !ok {
				pkg = manager.PackageInfo{}
				packages[name] = pkg
			}

			delete(packages, name)

			switch {
			case bytes.HasPrefix(line, []byte("dpkg-query: ")):
				pkg.Status = manager.PackageStatusUnknown
				// Keep the version from search results, don't overwrite with empty
			case string(parts[len(parts)-2]) == "installed":
				pkg.Status = manager.PackageStatusInstalled
				if version != "" {
					pkg.Version = version
				}
			case string(parts[len(parts)-2]) == "config-files":
				// Cross-package manager compatibility: normalize config-files state to available.
				// APT's config-files state (package removed but config files remain) maps to
				// the same semantic meaning as "available" in other package managers.
				pkg.Status = manager.PackageStatusAvailable
				if version != "" {
					pkg.Version = version
				}
			default:
				pkg.Status = manager.PackageStatusAvailable
				if version != "" {
					pkg.Version = version
				}
			}

			packagesList = append(packagesList, pkg)
		} else {
			log.Println("apt: line is empty")
		}
	}

	return packagesList, nil
}

// ParsePackageInfoOutput parses the output of `apt-cache show packageName` command
// and returns a manager.PackageInfo object containing package information such as name, version,
// architecture, and category. This function is useful for getting detailed package information.
func ParsePackageInfoOutput(msg string, opts *manager.Options) manager.PackageInfo {
	var pkg manager.PackageInfo

	// remove the last empty line
	msg = strings.TrimSuffix(msg, "\n")
	lines := strings.Split(string(msg), "\n")

	for _, line := range lines {
		if len(line) > 0 {
			parts := strings.SplitN(line, ":", 2)

			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "Package":
				pkg.Name = value
			case "Version":
				pkg.Version = value
			case "Architecture":
				pkg.Arch = value
			case "Section":
				pkg.Category = value
			}
		}
	}

	pkg.PackageManager = "apt"

	return pkg
}
