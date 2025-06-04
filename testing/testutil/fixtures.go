// Package testutil provides testing utilities for go-syspkg
package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// LoadFixture loads a test fixture file from the testing/fixtures directory
// packageManager should be "apt", "yum", "dnf", "snap", "flatpak", "apk"
func LoadFixture(t *testing.T, packageManager, filename string) string {
	t.Helper()

	// Get the project root by walking up from the current test file
	testDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Walk up to find the project root (where go.mod exists)
	projectRoot := testDir
	for {
		if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			t.Fatalf("Could not find project root with go.mod")
		}
		projectRoot = parent
	}

	fixturePath := filepath.Join(projectRoot, "testing", "fixtures", packageManager, filename)
	content, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("Failed to load fixture %s/%s: %v", packageManager, filename, err)
	}

	return string(content)
}

// LoadAPTFixture loads an APT fixture file
func LoadAPTFixture(t *testing.T, filename string) string {
	t.Helper()
	return LoadFixture(t, "apt", filename)
}

// LoadYUMFixture loads a YUM fixture file
func LoadYUMFixture(t *testing.T, filename string) string {
	t.Helper()
	return LoadFixture(t, "yum", filename)
}

// LoadSnapFixture loads a Snap fixture file
func LoadSnapFixture(t *testing.T, filename string) string {
	t.Helper()
	return LoadFixture(t, "snap", filename)
}

// LoadFlatpakFixture loads a Flatpak fixture file
func LoadFlatpakFixture(t *testing.T, filename string) string {
	t.Helper()
	return LoadFixture(t, "flatpak", filename)
}

// LoadDNFFixture loads a DNF fixture file
func LoadDNFFixture(t *testing.T, filename string) string {
	t.Helper()
	return LoadFixture(t, "dnf", filename)
}

// LoadAPKFixture loads an APK fixture file
func LoadAPKFixture(t *testing.T, filename string) string {
	t.Helper()
	return LoadFixture(t, "apk", filename)
}

// LoadZypperFixture loads a Zypper fixture file
func LoadZypperFixture(t *testing.T, filename string) string {
	t.Helper()
	return LoadFixture(t, "zypper", filename)
}
