package apt

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"

	"github.com/bluet/syspkg/internal"
)

func ParseInstallOutput(output string, opts *internal.Options) []internal.PackageInfo {
	// remove the last empty line
	output = strings.TrimSuffix(output, "\n")
	lines := strings.Split(string(output), "\n")
	// lines := strings.Split(output, "\n")
	var packages []internal.PackageInfo

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

			packageInfo := internal.PackageInfo{
				Name:           name,
				Arch:           arch,
				Version:        strings.Trim(parts[3], "()"),
				NewVersion:     strings.Trim(parts[3], "()"),
				Status:         internal.PackageStatusInstalled,
				PackageManager: pm,
			}
			packages = append(packages, packageInfo)
		}
	}

	return packages
}

func ParseDeletedOutput(output string, opts *internal.Options) []internal.PackageInfo {
	var packages []internal.PackageInfo
	// remove the last empty line
	output = strings.TrimSuffix(output, "\n")
	lines := strings.Split(string(output), "\n")

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

			packageInfo := internal.PackageInfo{
				Name:           name,
				Version:        strings.Trim(parts[2], "()"),
				NewVersion:     "",
				Category:       "",
				Arch:           arch,
				Status:         internal.PackageStatusAvailable,
				PackageManager: pm,
			}
			packages = append(packages, packageInfo)
		}
	}

	return packages
}

func ParseSearchOutput(output string, opts *internal.Options) []internal.PackageInfo {
	// apt-cache search output is like this:
	// package-name/category version\n\s\sdescription
	// ---------------------- --------------------
	// zutty/jammy 0.11.2.20220109.192032+dfsg1-1 amd64
	//   Efficient full-featured X11 terminal emulator
	//
	// zvbi/jammy 0.2.35-19 amd64
	//   Vertical Blanking Interval (VBI) utilities

	// fmt.Printf("apt: output: %s", output)
	// output starts with the following lines, which we need to remove:
	// "Sorting...\nFull Text Search...\n"
	output = strings.TrimPrefix(output, "Sorting...\nFull Text Search...\n")
	// remove the last empty line
	output = strings.TrimSuffix(output, "\n")

	// split output by empty lines
	lines := strings.Split(output, "\n\n")
	// fmt.Printf("apt: lines: %v (%d)", lines, len(lines))

	var packages []internal.PackageInfo
	var packagesDict = make(map[string]internal.PackageInfo)

	for _, line := range lines {
		if regexp.MustCompile(`^[\w\d-]+/[\w\d-,]+`).MatchString(line) {
			parts := strings.Fields(line)
			// names = append(names, strings.Split(parts[0], "/")[0])
			// name := strings.Split(parts[0], "/")[0]
			// status, _ := getPackageStatus(name)

			// debug a package with version "1.4p5-50build1"
			if strings.Contains(parts[1], "1.4p5-50build1") || strings.Contains(parts[1], "1.2.6-1") {
				fmt.Printf("apt: debug line: %s\n", line)
			}

			// if name is empty, it might be not what we want
			if parts[0] == "" {
				continue
			}

			packageInfo := internal.PackageInfo{
				Name:           strings.Split(parts[0], "/")[0],
				Version:        parts[1],
				NewVersion:     parts[1],
				Category:       strings.Split(parts[0], "/")[1],
				Arch:           parts[2],
				PackageManager: pm,
			}

			// fmt.Printf("apt: %v\n", packageInfo)

			// packages = append(packages, packageInfo)
			packagesDict[packageInfo.Name] = packageInfo
		}
	}

	if len(packagesDict) == 0 {
		return packages
	}

	packages, err := getPackageStatus(&packagesDict)
	if err != nil {
		log.Printf("apt: getPackageStatus error: %s\n", err)
	}

	return packages
}

func ParseListInstalledOutput(output string, opts *internal.Options) []internal.PackageInfo {
	var packages []internal.PackageInfo
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

			packageInfo := internal.PackageInfo{
				Name:           parts[0],
				Version:        parts[1],
				Status:         internal.PackageStatusInstalled,
				PackageManager: pm,
			}
			packages = append(packages, packageInfo)
		}
	}

	return packages
}

func ParseListUpgradableOutput(output string, opts *internal.Options) []internal.PackageInfo {
	var packages []internal.PackageInfo
	// remove the last empty line
	output = strings.TrimSuffix(output, "\n")
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "Inst") {
			parts := strings.Fields(line)

			// if name is empty, it might be not what we want
			if parts[1] == "" {
				continue
			}

			packageInfo := internal.PackageInfo{
				Name:           parts[1],
				Version:        strings.Trim(parts[2], "[]"),
				NewVersion:     strings.Trim(parts[3], "()"),
				Category:       parts[4],
				Arch:           strings.Trim(parts[5], "[]"),
				Status:         internal.PackageStatusUpgradable,
				PackageManager: pm,
			}
			packages = append(packages, packageInfo)
		}
	}

	return packages
}

func getPackageStatus(packages *map[string]internal.PackageInfo) ([]internal.PackageInfo, error) {
	var packageNames []string
	var packagesList []internal.PackageInfo

	if len(*packages) == 0 {
		return packagesList, nil
	}

	for name := range *packages {
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

	packagesList, err = ParseDpkgQueryOutput(string(out), packages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dpkg-query output: %v", err)
	}

	// for all the packages that are not found, set their status to unknown, if any
	for _, pkg := range *packages {
		fmt.Printf("apt: package not found by dpkg-query: %s", pkg.Name)
		pkg.Status = internal.PackageStatusUnknown
		packagesList = append(packagesList, pkg)
	}

	return packagesList, nil
}

func ParseDpkgQueryOutput(output string, packages *map[string]internal.PackageInfo) ([]internal.PackageInfo, error) {
	var packagesList []internal.PackageInfo

	// remove the last empty line
	output = strings.TrimSuffix(output, "\n")
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if len(line) > 0 {
			parts := strings.Fields(line)
			name := parts[0]

			if strings.HasPrefix(name, "dpkg-query:") {
				name = parts[len(parts)-1]

				if strings.Contains(name, ":") {
					name = strings.Split(name, ":")[0]
				}
			}

			// if name is empty, it might be not what we want
			if name == "" {
				continue
			}

			version := parts[len(parts)-1]

			if !regexp.MustCompile(`^\d`).MatchString(version) {
				version = ""
			}

			pkg := (*packages)[name]
			delete(*packages, name)

			if strings.HasPrefix(line, "dpkg-query: ") {
				pkg.Status = internal.PackageStatusUnknown
				pkg.Version = ""
			} else if parts[len(parts)-2] == "installed" {
				pkg.Status = internal.PackageStatusInstalled
				pkg.Version = version
			} else {
				pkg.Status = internal.PackageStatusAvailable
				pkg.Version = version
			}

			packagesList = append(packagesList, pkg)
		} else {
			fmt.Printf("apt: line is empty\n")
		}
	}

	return packagesList, nil
}

func ParsePackageInfoOutput(output string, opts *internal.Options) internal.PackageInfo {
	// var packages []internal.PackageInfo

	// Package: libslf4j-java
	// Architecture: all
	// Version: 1.7.32-1
	// Priority: optional
	// Section: universe/libs
	// Origin: Ubuntu
	// Maintainer: Ubuntu Developers <ubuntu-devel-discuss@lists.ubuntu.com>
	// Original-Maintainer: Debian Java Maintainers <pkg-java-maintainers@lists.alioth.debian.org>
	// Bugs: https://bugs.launchpad.net/ubuntu/+filebug
	// Installed-Size: 281
	// Suggests: libcommons-logging-java, liblog4j1.2-java
	// Filename: pool/universe/libs/libslf4j-java/libslf4j-java_1.7.32-1_all.deb
	// Size: 140954
	// MD5sum: a2cf48b9ba36005e2a052fcdaeed88c6
	// SHA1: e4a1999eae0cb1706d6ba56b914562f67891e99c
	// SHA256: 963b9764cfaa16ebd851dd8625193ad25d1b32767a6eafb73f6125d081cb5bf3
	// SHA512: 3f3e5ad0cd31d060b16e51c71fb7becb1b7184b018077fcd0520465fd15baf74c722c753ef196b4019e8a1b4e33478b83804e89faf9c1d447aa7276fc07a064b
	// Homepage: http://www.slf4j.org
	// Description-en: Simple Logging Facade for Java
	//  The Simple Logging Facade for Java (or SLF4J) is intended to serve as
	//  a simple facade for various logging APIs allowing to the end-user to
	//  plug in the desired implementation at deployment time. SLF4J also
	//  allows for a gradual migration path away from Apache Commons
	//  Logging (CL)
	//  .
	//  Logging API implementations can either choose to implement the SLF4J
	//  interfaces directly, e.g. logback or SimpleLogger. Alternatively, it
	//  is possible (and rather easy) to write SLF4J adapters for the given
	//  API implementation, e.g. Log4jLoggerAdapter or JDK14LoggerAdapter.
	// Description-md5: 307af13d2db4d50e6f124f83f84006d9

	// remove the last empty line
	output = strings.TrimSuffix(output, "\n")
	lines := strings.Split(string(output), "\n")

	var pkg internal.PackageInfo
	var packagesDict = make(map[string]internal.PackageInfo)
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

			pkg.PackageManager = "apt"

			pkg.AdditionalData = make(map[string]string)
			pkg.AdditionalData["output"] = output

			packagesDict[pkg.Name] = pkg
		}
	}

	var packages, err = getPackageStatus(&packagesDict)
	if err != nil {
		fmt.Printf("apt: failed to get package status: %v", err)
	}

	return packages[0]
}
