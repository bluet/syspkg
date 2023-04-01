package apt

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"regexp"
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
				Status:         internal.PackageStatusInstalled,
				PackageManager: pm,
			}
			packages = append(packages, packageInfo)
		}
	}

	return packages
}

func parseSearchOutput(output string, opts *internal.Options) []internal.PackageInfo {
	// apt-cache search output is like this:
	// package-name/category version\n\s\sdescription
	// ---------------------- --------------------
	// zutty/jammy 0.11.2.20220109.192032+dfsg1-1 amd64
	//   Efficient full-featured X11 terminal emulator
	//
	// zvbi/jammy 0.2.35-19 amd64
	//   Vertical Blanking Interval (VBI) utilities

	// split output by empty lines
	lines := strings.Split(output, "\n\n")

	var packages []internal.PackageInfo
	var packagesDict = make(map[string]internal.PackageInfo)

	for _, line := range lines {
		if regexp.MustCompile(`^[\w\d-]+/[\w\d-]+`).MatchString(line) {
			parts := strings.Fields(line)
			// names = append(names, strings.Split(parts[0], "/")[0])
			// name := strings.Split(parts[0], "/")[0]
			// status, _ := getPackageStatus(name)
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

	packages, err := getPackageStatus(&packagesDict)
	if err != nil {
		log.Printf("apt: %s", err)
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
				Status:         internal.PackageStatusInstalled,
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

	for name := range *packages {
		packageNames = append(packageNames, name)
	}

	args := []string{"-W", "--showformat", "${binary:Package} ${Status} ${Version}\n"}
	args = append(args, packageNames...)
	cmd := exec.Command("dpkg-query", args...)
	cmd.Env = ENV_NonInteractive
	// out, _ := cmd.Output()
	out, _ := cmd.CombinedOutput()
	// out, err := cmd.Output()

	// FIXME: apt: exit status 1
	// if err != nil {
	// 	return nil, err
	// }

	// fmt.Printf("apt: dpkg-query output: %s\n", string(out))
	out = bytes.TrimRight(out, "\n")
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		// fmt.Printf("apt: line (%s):\t%s\n", fmt.Sprint(len(line)), line)
		if len(line) > 0 {
			parts := strings.Fields(line)

			name := parts[0]
			// when package not found, message format will be:
			// dpkg-query: no packages found matching PACKAGE_NAME
			if strings.HasPrefix(name, "dpkg-query:") {
				name = parts[len(parts)-1]
			}
			// sometimes the package name contains ":", and the name is the first part before ":"
			if strings.Contains(name, ":") {
				name = strings.Split(name, ":")[0]
			}

			version := parts[len(parts)-1]
			// if version not starts with number, it's not a version
			if !regexp.MustCompile(`^\d`).MatchString(version) {
				version = ""
			}

			// pkg, ok := (*packages)[name] // get the package value from the map
			// delete(*packages, name) and save it to the list
			pkg := (*packages)[name]
			delete(*packages, name)

			// if !ok {                     // if the package is not found, initialize its status and add it to the map
			// 	pkg = internal.PackageInfo{Status: internal.PackageStatusUnknown}
			// 	(*packages)[name] = pkg
			// }

			// if ok { // if the package is found, update its version
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
			// }
			// (*packages)[name] = pkg
			packagesList = append(packagesList, pkg)
		} else {
			fmt.Printf("apt: line is empty\n")
		}
	}

	// for all the packages that are not found, set their status to unknown, if any
	for _, pkg := range *packages {
		fmt.Printf("apt: package not found by dpkg-query: %s", pkg.Name)
		pkg.Status = internal.PackageStatusUnknown
		packagesList = append(packagesList, pkg)
	}

	return packagesList, nil
}
