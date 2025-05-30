package testenv

import (
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
