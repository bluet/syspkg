// Package osinfo provides a simple way to gather information about the
// operating system, including the name, distribution, version, and architecture.
// Description: Get OS information. Including: OS Type, Distribution, Version, CPU Architecture.
//  * Name: OS name (ex, linux, darwin, windows)
//  * Distribution: OS distribution (ex, Ubuntu)
//  * Version: OS version (ex, 20.04)
//  * Arch: OS architecture (ex, amd64)
//
// Author: BlueT - Matthew Lien - 練喆明 <bluet@bluet.org>

package osinfo

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// OSInfo represents the operating system information, including its name,
// distribution, version, and architecture.
type OSInfo struct {
	Name         string
	Distribution string
	Version      string
	Arch         string
}

// GetOSInfo returns the OS information, as a pointer to an OSInfo struct.
// The OSInfo struct contains the following fields:
//  * Name: OS name (ex, linux, darwin, windows)
//  * Distribution: OS distribution (ex, Ubuntu)
//  * Version: OS version (ex, 20.04)
//  * Arch: OS architecture (ex, amd64)
func GetOSInfo() (*OSInfo, error) {
	osName := runtime.GOOS
	osArch := runtime.GOARCH
	var osDist, osVersion string
	var err error

	switch osName {
	case "linux":
		dist, ver, err := getLinuxDistribution()
		if err != nil {
			return nil, err
		}
		osDist = dist
		osVersion = ver
	case "darwin":
		osDist = "macOS"
		osVersion, err = getMacOSVersion()
		if err != nil {
			return nil, fmt.Errorf("failed to get macOS version: %v", err)
		}
	case "windows":
		osDist = "Windows"
		osVersion, err = getWindowsVersion()
		if err != nil {
			return nil, fmt.Errorf("failed to get Windows version: %v", err)
		}
	default:
		osDist = "N/A"
		osVersion = "N/A"
	}

	return &OSInfo{
		Name:         osName,
		Version:      osVersion,
		Distribution: osDist,
		Arch:         osArch,
	}, nil
}

// getLinuxDistribution returns the Linux distribution name and version.
// Parse the content of /etc/os-release to get the distribution name and version.
func getLinuxDistribution() (string, string, error) {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	var dist, distVersion string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID=") {
			dist = strings.TrimPrefix(line, "ID=")
		} else if strings.HasPrefix(line, "VERSION_ID=") {
			distVersion = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), "\"")
		}
	}

	if err := scanner.Err(); err != nil {
		return "", "", err
	}

	return dist, distVersion, nil
}


// getMacOSVersion returns the macOS version as a string.
func getMacOSVersion() (string, error) {
	out, err := exec.Command("sw_vers", "-productVersion").Output()
	if err != nil {
		return "", err
	}

	macOSVersion := strings.TrimSpace(string(out))
	return macOSVersion, nil
}

// getWindowsVersion returns the Windows version as a string.
func getWindowsVersion() (string, error) {
	out, err := exec.Command("cmd", "/C", "ver").Output()
	if err != nil {
		return "", err
	}

	output := strings.TrimSpace(string(out))
	// The output will be in the format: "Microsoft Windows [Version X.Y.Z]".
	// We need to extract X.Y.Z from the output.
	start := strings.Index(output, "[")
	end := strings.Index(output, "]")

	if start == -1 || end == -1 {
		return "", errors.New("failed to parse Windows version")
	}

	// Get the version string and remove the "Version" prefix.
	version := strings.TrimPrefix(strings.TrimSpace(output[start+1:end]), "Version ")

	return version, nil
}
