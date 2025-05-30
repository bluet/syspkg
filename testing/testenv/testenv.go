// Package testenv provides utilities for detecting test environments and
// determining which package managers should be tested based on the current OS.
package testenv

import (
	"os"
	"strings"

	"github.com/bluet/syspkg/osinfo"
)

// TestEnvironment represents the current testing environment
type TestEnvironment struct {
	OS                string
	Distribution      string
	Version           string
	InContainer       bool
	AvailableManagers []string
	TestTags          []string
}

// GetTestEnvironment detects the current test environment and returns
// information about what should be tested
func GetTestEnvironment() (*TestEnvironment, error) {
	osInfo, err := osinfo.GetOSInfo()
	if err != nil {
		return nil, err
	}

	env := &TestEnvironment{
		OS:           osInfo.Name,
		Distribution: osInfo.Distribution,
		Version:      osInfo.Version,
		InContainer:  os.Getenv("IN_CONTAINER") == "true",
	}

	// Determine available package managers based on OS
	env.AvailableManagers = getAvailableManagers(osInfo)
	env.TestTags = getRecommendedTestTags(env)

	return env, nil
}

// getAvailableManagers returns the list of package managers that should
// be available on the given OS
func getAvailableManagers(osInfo *osinfo.OSInfo) []string {
	var managers []string

	switch osInfo.Name {
	case "linux":
		switch strings.ToLower(osInfo.Distribution) {
		case "ubuntu", "debian":
			managers = []string{"apt"}
			// Flatpak available but requires setup
			if os.Getenv("IN_CONTAINER") != "true" {
				managers = append(managers, "flatpak", "snap")
			}

		case "fedora":
			managers = []string{"dnf"}
			if os.Getenv("IN_CONTAINER") != "true" {
				managers = append(managers, "flatpak")
			}

		case "rocky", "almalinux", "centos":
			// Determine YUM vs DNF based on version
			if osInfo.Version >= "8" {
				managers = []string{"yum"}
			}
			if osInfo.Version >= "9" || osInfo.Distribution == "fedora" {
				managers = []string{"dnf"}
			}

		case "alpine":
			managers = []string{"apk"}

		case "arch":
			managers = []string{"pacman"}
		}

	case "darwin":
		managers = []string{"brew"}

	case "windows":
		managers = []string{"choco", "scoop", "winget"}
	}

	return managers
}

// getRecommendedTestTags returns the recommended test tags for the environment
func getRecommendedTestTags(env *TestEnvironment) []string {
	tags := []string{"unit"} // Always run unit tests

	if env.InContainer {
		tags = append(tags, "integration")
		// Add specific package manager tags
		tags = append(tags, env.AvailableManagers...)
	} else {
		// Native environment can run system tests
		tags = append(tags, "integration", "system")
	}

	return tags
}

// ShouldSkipTest determines if a test should be skipped based on environment
func (env *TestEnvironment) ShouldSkipTest(requiredPM string) (bool, string) {
	// Check if package manager is available in this environment
	for _, available := range env.AvailableManagers {
		if available == requiredPM {
			return false, ""
		}
	}

	return true, "Package manager " + requiredPM + " not available on " +
		env.OS + "/" + env.Distribution
}

// GetFixturePath returns the appropriate fixture path for the current OS
func (env *TestEnvironment) GetFixturePath(pm, operation string) string {
	base := "testing/fixtures/" + pm + "/"

	// Use OS-specific fixtures if available
	osSpecific := base + operation + "-" + env.Distribution + ".txt"
	if _, err := os.Stat(osSpecific); err == nil {
		return osSpecific
	}

	// Fall back to generic fixtures
	return base + operation + ".txt"
}

// IsContainerEnvironment returns true if running in a container
func IsContainerEnvironment() bool {
	return os.Getenv("IN_CONTAINER") == "true"
}

// GetTestPackageManager returns the package manager to test from environment
func GetTestPackageManager() string {
	return os.Getenv("TEST_PACKAGE_MANAGER")
}

// GetTestOS returns the OS being tested from environment
func GetTestOS() string {
	return os.Getenv("TEST_OS")
}
