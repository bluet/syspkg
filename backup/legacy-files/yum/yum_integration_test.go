//go:build integration
// +build integration

package yum_test

import (
	"os"
	"testing"

	"github.com/bluet/syspkg/manager"
	"github.com/bluet/syspkg/manager/yum"
	"github.com/bluet/syspkg/testing/testenv"
)

// TestYUMOperations_Integration tests YUM operations with real command execution
// These tests are skipped unless YUM is available on the system
func TestYUMOperations_Integration(t *testing.T) {
	// Check if we should run YUM tests
	env, err := testenv.GetTestEnvironment()
	if err != nil {
		t.Fatalf("Failed to get test environment: %v", err)
	}

	if skip, reason := env.ShouldSkipTest("yum"); skip {
		t.Skip(reason)
	}

	pm := &yum.PackageManager{}

	// Verify YUM is available
	if !pm.IsAvailable() {
		t.Skip("YUM is not available on this system")
	}

	t.Run("Find", func(t *testing.T) {
		// Search for a package that should exist in repos
		packages, err := pm.Find([]string{"bash"}, &manager.Options{})
		if err != nil {
			t.Fatalf("Find failed: %v", err)
		}

		if len(packages) == 0 {
			t.Error("Expected to find bash package")
		}

		// Verify the Find() enhancement works
		for _, pkg := range packages {
			if pkg.Name == "bash" {
				// bash should be installed on any Linux system
				if pkg.Status != manager.PackageStatusInstalled {
					t.Errorf("bash should be installed, got status: %s", pkg.Status)
				}
				if pkg.Version == "" {
					t.Error("Installed bash should have version")
				}
				return
			}
		}
		t.Error("bash package not found in results")
	})

	t.Run("ListInstalled", func(t *testing.T) {
		packages, err := pm.ListInstalled(&manager.Options{})
		if err != nil {
			t.Fatalf("ListInstalled failed: %v", err)
		}

		if len(packages) == 0 {
			t.Error("Expected installed packages on system")
		}

		// All packages should have installed status
		for _, pkg := range packages {
			if pkg.Status != manager.PackageStatusInstalled {
				t.Errorf("Package %s should have status installed, got %s",
					pkg.Name, pkg.Status)
			}
			if pkg.Version == "" {
				t.Errorf("Package %s should have version", pkg.Name)
			}
		}
	})

	t.Run("GetPackageInfo", func(t *testing.T) {
		// Test with a package that should be installed
		pkg, err := pm.GetPackageInfo("rpm", &manager.Options{})
		if err != nil {
			t.Fatalf("GetPackageInfo failed: %v", err)
		}

		if pkg.Name != "rpm" {
			t.Errorf("Expected package name rpm, got %s", pkg.Name)
		}

		// rpm should always be installed on YUM-based systems
		if pkg.Status != manager.PackageStatusInstalled {
			t.Errorf("rpm should be installed, got status: %s", pkg.Status)
		}
	})

	t.Run("Clean", func(t *testing.T) {
		// Test clean operation (safe to run)
		err := pm.Clean(&manager.Options{DryRun: true})
		if err != nil {
			t.Errorf("Clean (dry run) failed: %v", err)
		}
	})

	t.Run("Refresh", func(t *testing.T) {
		// Skip refresh in CI to avoid network operations
		if os.Getenv("CI") == "true" {
			t.Skip("Skipping refresh in CI environment")
		}

		err := pm.Refresh(&manager.Options{DryRun: true})
		if err != nil {
			t.Errorf("Refresh (dry run) failed: %v", err)
		}
	})
}

// TestYUMParsers_UnitWithFixtures tests parser functions with real YUM output fixtures
// These are pure unit tests that don't require YUM to be installed
func TestYUMParsers_UnitWithFixtures(t *testing.T) {
	t.Run("ParseFindOutput", func(t *testing.T) {
		fixture := loadFixture(t, "search-vim-rocky8.txt")
		pm := yum.NewPackageManager()
		packages := pm.ParseFindOutput(fixture, &manager.Options{})

		if len(packages) != 5 {
			t.Errorf("Expected 5 packages, got %d", len(packages))
		}

		// Verify parser limitations are documented
		for _, pkg := range packages {
			if pkg.Status != manager.PackageStatusAvailable {
				t.Errorf("Parser should return all as available, got %s", pkg.Status)
			}
			if pkg.Version != "" {
				t.Errorf("Parser should not set version from search output")
			}
		}
	})

	t.Run("ParseListInstalledOutput", func(t *testing.T) {
		fixture := loadFixture(t, "list-installed-minimal-rocky8.txt")
		packages := yum.ParseListInstalledOutput(fixture, &manager.Options{})

		if len(packages) == 0 {
			t.Error("Expected packages from list installed output")
		}

		// All should be installed with versions
		for _, pkg := range packages {
			if pkg.Status != manager.PackageStatusInstalled {
				t.Errorf("Package %s should be installed", pkg.Name)
			}
			if pkg.Version == "" {
				t.Errorf("Package %s should have version", pkg.Name)
			}
		}
	})

	t.Run("ParsePackageInfoOutput", func(t *testing.T) {
		// Test installed package
		fixture := loadFixture(t, "info-vim-installed-rocky8.txt")
		pkg := yum.ParsePackageInfoOutput(fixture, &manager.Options{})

		if pkg.Name != "vim-enhanced" {
			t.Errorf("Expected vim-enhanced, got %s", pkg.Name)
		}
		if pkg.Status != manager.PackageStatusInstalled {
			t.Errorf("Expected installed status, got %s", pkg.Status)
		}

		// Test available package
		fixture = loadFixture(t, "info-nginx-rocky8.txt")
		pkg = yum.ParsePackageInfoOutput(fixture, &manager.Options{})

		if pkg.Status != manager.PackageStatusAvailable {
			t.Errorf("Expected available status, got %s", pkg.Status)
		}
	})
}

// loadFixture is defined in behavior_test.go in the same package
