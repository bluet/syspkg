//go:build integration
// +build integration

package yum

import (
	"os"
	"testing"

	"github.com/bluet/syspkg/manager"
	"github.com/bluet/syspkg/testing/testenv"
)

// TestYumIntegrationEnvironmentAware demonstrates environment-aware testing
func TestYumIntegrationEnvironmentAware(t *testing.T) {
	env, err := testenv.GetTestEnvironment()
	if err != nil {
		t.Fatalf("Failed to get test environment: %v", err)
	}

	// Skip if YUM not available in this environment
	if skip, reason := env.ShouldSkipTest("yum"); skip {
		t.Skip(reason)
	}

	yumManager := PackageManager{}

	// Test availability
	if !yumManager.IsAvailable() {
		t.Skip("YUM not available in this environment")
	}

	t.Run("ListInstalled", func(t *testing.T) {
		opts := &manager.Options{}
		packages, err := yumManager.ListInstalled(opts)

		if err != nil {
			t.Errorf("ListInstalled failed: %v", err)
			return
		}

		if len(packages) == 0 {
			t.Log("No packages found (expected in minimal containers)")
		} else {
			t.Logf("Found %d installed packages", len(packages))

			// Log first few packages for debugging
			for i, pkg := range packages {
				if i >= 3 {
					break
				}
				t.Logf("Package: %s, Version: %s, Arch: %s",
					pkg.Name, pkg.Version, pkg.Arch)
			}
		}
	})

	t.Run("SearchVim", func(t *testing.T) {
		if env.InContainer {
			t.Log("Running search in container environment")
		}

		opts := &manager.Options{}
		packages, err := yumManager.Find([]string{"vim"}, opts)

		if err != nil {
			t.Errorf("Find failed: %v", err)
			return
		}

		if len(packages) == 0 {
			t.Log("No vim packages found (may be expected in some environments)")
		} else {
			t.Logf("Found %d vim-related packages", len(packages))

			// Verify at least one package has "vim" in the name
			found := false
			for _, pkg := range packages {
				if pkg.Name == "vim" {
					found = true
					t.Logf("Found vim package: %s, Arch: %s", pkg.Name, pkg.Arch)
					break
				}
			}

			if !found {
				t.Log("No exact 'vim' match found, but related packages exist")
			}
		}
	})

	t.Run("GetPackageInfo", func(t *testing.T) {
		// Test with a package that should exist in most RHEL-based systems
		testPackage := "bash"

		opts := &manager.Options{}
		pkg, err := yumManager.GetPackageInfo(testPackage, opts)

		if err != nil {
			t.Logf("GetPackageInfo for %s failed: %v (may be expected in containers)", testPackage, err)
			return
		}

		if pkg.Name == "" {
			t.Error("Package info returned empty name")
		} else {
			t.Logf("Package info: Name=%s, Version=%s, Arch=%s",
				pkg.Name, pkg.Version, pkg.Arch)
		}
	})

	t.Run("Clean", func(t *testing.T) {
		opts := &manager.Options{Verbose: env.InContainer} // Verbose in containers for debugging

		err := yumManager.Clean(opts)
		if err != nil {
			t.Errorf("Clean failed: %v", err)
		} else {
			t.Log("Clean operation completed successfully")
		}
	})

	t.Run("Refresh", func(t *testing.T) {
		opts := &manager.Options{}

		err := yumManager.Refresh(opts)
		if err != nil {
			t.Errorf("Refresh failed: %v", err)
		} else {
			t.Log("Refresh operation completed successfully")
		}
	})
}

// TestYumParsingWithRealOutput tests parsing with real YUM output
func TestYumParsingWithRealOutput(t *testing.T) {
	env, err := testenv.GetTestEnvironment()
	if err != nil {
		t.Fatalf("Failed to get test environment: %v", err)
	}

	// Only run if we can capture real output
	if !env.InContainer {
		t.Skip("Real output parsing test only runs in containers")
	}

	// Test parsing with fixtures appropriate to current environment
	t.Run("ParseSearchOutput", func(t *testing.T) {
		fixturePath := env.GetFixturePath("yum", "search-vim")

		if data, err := os.ReadFile(fixturePath); err == nil {
			pm := NewPackageManager()
			packages := pm.ParseFindOutput(string(data), nil)

			if len(packages) == 0 {
				t.Error("Failed to parse any packages from fixture")
			} else {
				t.Logf("Parsed %d packages from %s", len(packages), fixturePath)
			}
		} else {
			t.Logf("No fixture available at %s, skipping", fixturePath)
		}
	})
}
