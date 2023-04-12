package apt

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"

	// "github.com/rs/zerolog"
	// "github.com/rs/zerolog/log"

	"github.com/bluet/syspkg/manager"
)

func ParseInstallOutput(output string, opts *manager.Options) []manager.PackageInfo {
	var packages []manager.PackageInfo

	// remove the last empty line
	output = strings.TrimSuffix(output, "\n")
	var lines []string = strings.Split(string(output), "\n")

	for _, line := range lines {
		if opts.Verbose {
			log.Printf("apt: %s", line)
		}
		if strings.HasPrefix(line, "Setting up") {
			parts := strings.Fields(line)
			var name, arch string
			if strings.Contains(parts[2], ":") {
				name = strings.Split(parts[2], ":")[0]
				arch = strings.Split(parts[2], ":")[1]
			} else {
				name = parts[2]
			}

			// if name is empty, it might be not what we want
			if name == "" {
				continue
			}

			packageInfo := manager.PackageInfo{
				Name:           name,
				Arch:           arch,
				Version:        strings.Trim(parts[3], "()"),
				NewVersion:     strings.Trim(parts[3], "()"),
				Status:         manager.PackageStatusInstalled,
				PackageManager: pm,
			}
			packages = append(packages, packageInfo)
		}
	}

	return packages
}

func ParseDeletedOutput(output string, opts *manager.Options) []manager.PackageInfo {
	var packages []manager.PackageInfo

	// remove the last empty line
	output = strings.TrimSuffix(output, "\n")
	var lines []string = strings.Split(string(output), "\n")

	for _, line := range lines {
		if opts.Verbose {
			log.Printf("apt: %s", line)
		}
		if strings.HasPrefix(line, "Removing") {
			parts := strings.Fields(line)
			if opts.Verbose {
				log.Printf("apt: parts: %s", parts)
			}
			var name, arch string
			if strings.Contains(parts[1], ":") {
				name = strings.Split(parts[1], ":")[0]
				arch = strings.Split(parts[1], ":")[1]
			} else {
				name = parts[1]
			}

			// if name is empty, it might be not what we want
			if name == "" {
				continue
			}

			packageInfo := manager.PackageInfo{
				Name:           name,
				Version:        strings.Trim(parts[2], "()"),
				NewVersion:     "",
				Category:       "",
				Arch:           arch,
				Status:         manager.PackageStatusAvailable,
				PackageManager: pm,
			}
			packages = append(packages, packageInfo)
		}
	}

	return packages
}

func ParseFindOutput(output string, opts *manager.Options) []manager.PackageInfo {
	var packages []manager.PackageInfo
	var packagesDict = make(map[string]manager.PackageInfo)

	// Sorting...
	// Full Text Search...
	// zutty/jammy 0.11.2.20220109.192032+dfsg1-1 amd64
	//   Efficient full-featured X11 terminal emulator
	//
	// zvbi/jammy 0.2.35-19 amd64
	//   Vertical Blanking Interval (VBI) utilities

	output = strings.TrimPrefix(output, "Sorting...\nFull Text Search...\n")

	// remove the last empty line
	output = strings.TrimSuffix(output, "\n")

	// split output by empty lines
	var lines []string = strings.Split(output, "\n\n")

	for _, line := range lines {
		if regexp.MustCompile(`^[\w\d-]+/[\w\d-,]+`).MatchString(line) {
			parts := strings.Fields(line)

			// debug a package with version "1.4p5-50build1"
			if strings.Contains(parts[1], "1.4p5-50build1") || strings.Contains(parts[1], "1.2.6-1") {
				fmt.Printf("apt: debug line: %s\n", line)
			}

			// if name is empty, it might be not what we want
			if parts[0] == "" {
				continue
			}

			packageInfo := manager.PackageInfo{
				Name:           strings.Split(parts[0], "/")[0],
				Version:        parts[1],
				NewVersion:     parts[1],
				Category:       strings.Split(parts[0], "/")[1],
				Arch:           parts[2],
				PackageManager: pm,
			}

			packagesDict[packageInfo.Name] = packageInfo
		}
	}

	if len(packagesDict) == 0 {
		return packages
	}

	packages, err := getPackageStatus(packagesDict)
	if err != nil {
		log.Printf("apt: getPackageStatus error: %s\n", err)
	}

	return packages
}

func ParseListInstalledOutput(output string, opts *manager.Options) []manager.PackageInfo {
	var packages []manager.PackageInfo

	// remove the last empty line
	output = strings.TrimSuffix(output, "\n")
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if len(line) > 0 {
			parts := strings.Fields(line)

			// if name is empty, it might be not what we want
			if parts[0] == "" {
				continue
			}
			var name, arch string
			if strings.Contains(parts[0], ":") {
				name = strings.Split(parts[0], ":")[0]
				arch = strings.Split(parts[0], ":")[1]
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

func ParseListUpgradableOutput(output string, opts *manager.Options) []manager.PackageInfo {
	var packages []manager.PackageInfo

	// Listing...
	// cloudflared/unknown 2023.4.0 amd64 [upgradable from: 2023.3.1]
	// libllvm15/jammy-updates 1:15.0.7-0ubuntu0.22.04.1 amd64 [upgradable from: 1:15.0.6-3~ubuntu0.22.04.2]
	// libllvm15/jammy-updates 1:15.0.7-0ubuntu0.22.04.1 i386 [upgradable from: 1:15.0.6-3~ubuntu0.22.04.2]

	// remove the last empty line
	output = strings.TrimSuffix(output, "\n")
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		// skip if line starts with "Listing..."
		if strings.HasPrefix(line, "Listing...") {
			continue
		}

		if len(line) > 0 {
			parts := strings.Fields(line)
			// log.Printf("apt: parts: %+v", parts)

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

func getPackageStatus(packages map[string]manager.PackageInfo) ([]manager.PackageInfo, error) {
	var packageNames []string
	var packagesList []manager.PackageInfo

	if len(packages) == 0 {
		return packagesList, nil
	}

	for name := range packages {
		packageNames = append(packageNames, name)
	}

	args := []string{"-W", "--showformat", "${binary:Package} ${Status} ${Version}\n"}
	args = append(args, packageNames...)
	cmd := exec.Command("dpkg-query", args...)
	cmd.Env = ENV_NonInteractive

	// dpkg-query might exit with status 1, which is not an error when some packages are not found
	out, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() != 1 && !strings.Contains(string(out), "no packages found matching") {
				return nil, fmt.Errorf("command failed with output: %s", string(out))
			}
		}
	}

	packagesList, err = ParseDpkgQueryOutput(out, packages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dpkg-query output: %+v", err)
	}

	// for all the packages that are not found, set their status to unknown, if any
	for _, pkg := range packages {
		fmt.Printf("apt: package not found by dpkg-query: %s", pkg.Name)
		pkg.Status = manager.PackageStatusUnknown
		packagesList = append(packagesList, pkg)
	}

	return packagesList, nil
}

func ParseDpkgQueryOutput(output []byte, packages map[string]manager.PackageInfo) ([]manager.PackageInfo, error) {
	var packagesList []manager.PackageInfo

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

				if strings.Contains(name, ":") {
					name = strings.Split(name, ":")[0]
				}
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
				pkg.Version = ""
			case string(parts[len(parts)-2]) == "installed":
				pkg.Status = manager.PackageStatusInstalled
				if version != "" {
					pkg.Version = version
				}
			case string(parts[len(parts)-2]) == "config-files":
				pkg.Status = manager.PackageStatusConfigFiles
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

func ParsePackageInfoOutput(output string, opts *manager.Options) manager.PackageInfo {
	var pkg manager.PackageInfo

	// remove the last empty line
	output = strings.TrimSuffix(output, "\n")
	lines := strings.Split(string(output), "\n")

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
