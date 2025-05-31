package testenv

import (
	"strconv"
	"strings"
	"testing"
)

func TestGetTestEnvironment(t *testing.T) {
	env, err := GetTestEnvironment()
	if err != nil {
		t.Fatalf("Failed to get test environment: %v", err)
	}

	if env.OS == "" {
		t.Error("OS should not be empty")
	}

	if env.Distribution == "" {
		t.Error("Distribution should not be empty")
	}

	if len(env.AvailableManagers) == 0 {
		t.Error("Should have at least one available package manager")
	}

	t.Logf("Test Environment: OS=%s, Distribution=%s, Version=%s",
		env.OS, env.Distribution, env.Version)
	t.Logf("Available Package Managers: %v", env.AvailableManagers)
	t.Logf("Recommended Test Tags: %v", env.TestTags)
	t.Logf("In Container: %v", env.InContainer)
}

func TestShouldSkipTest(t *testing.T) {
	env, err := GetTestEnvironment()
	if err != nil {
		t.Fatalf("Failed to get test environment: %v", err)
	}

	// Test with a package manager that should be available
	if len(env.AvailableManagers) > 0 {
		available := env.AvailableManagers[0]
		skip, reason := env.ShouldSkipTest(available)
		if skip {
			t.Errorf("Should not skip test for available package manager %s: %s", available, reason)
		}
	}

	// Test with a package manager that should not be available
	skip, reason := env.ShouldSkipTest("nonexistent-pm")
	if !skip {
		t.Error("Should skip test for nonexistent package manager")
	}
	if reason == "" {
		t.Error("Should provide reason for skipping")
	}
}

func TestGetFixturePath(t *testing.T) {
	env, err := GetTestEnvironment()
	if err != nil {
		t.Fatalf("Failed to get test environment: %v", err)
	}

	path := env.GetFixturePath("apt", "search-vim")
	if path == "" {
		t.Error("Fixture path should not be empty")
	}

	t.Logf("Fixture path for apt search-vim: %s", path)
}

// TestVersionParsing tests the version parsing logic for RHEL-based distributions
func TestVersionParsing(t *testing.T) {
	tests := []struct {
		version  string
		expected string // expected package manager
	}{
		{"8", "yum"},
		{"8.5", "yum"},
		{"8.5.2111", "yum"},
		{"9", "dnf"},
		{"9.0", "dnf"},
		{"9.1.2022", "dnf"},
		{"7.9", ""}, // No manager for version < 8
	}

	for _, tt := range tests {
		t.Run("version_"+tt.version, func(t *testing.T) {
			// Simulate version parsing logic
			versionParts := strings.Split(tt.version, ".")
			var manager string
			if len(versionParts) > 0 {
				majorVersion, err := strconv.Atoi(versionParts[0])
				if err == nil {
					if majorVersion >= 9 {
						manager = "dnf"
					} else if majorVersion >= 8 {
						manager = "yum"
					}
				}
			}

			if manager != tt.expected {
				t.Errorf("For version %s, expected %s but got %s", tt.version, tt.expected, manager)
			}
		})
	}
}
