package yum_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bluet/syspkg/manager"
	"github.com/bluet/syspkg/manager/yum"
)

// loadFixture loads test fixture data from testing/fixtures/yum/
func loadFixture(t *testing.T, filename string) string {
	t.Helper()
	// Get the module root by finding go.mod
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Walk up the directory tree to find go.mod
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("Could not find module root (go.mod)")
		}
		dir = parent
	}

	fixturePath := filepath.Join(dir, "testing", "fixtures", "yum", filename)
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("Failed to load fixture %s: %v", filename, err)
	}

	return string(data)
}

// TestPackageManager_ImplementsInterface verifies contract compliance
func TestPackageManager_ImplementsInterface(t *testing.T) {
	var _ interface {
		IsAvailable() bool
		GetPackageManager() string
		Find([]string, *manager.Options) ([]manager.PackageInfo, error)
		ListInstalled(*manager.Options) ([]manager.PackageInfo, error)
		GetPackageInfo(string, *manager.Options) (manager.PackageInfo, error)
		Clean(*manager.Options) error
		Refresh(*manager.Options) error
	} = &yum.PackageManager{}
}

// TestPackageManager_GetPackageManager tests the identifier behavior
func TestPackageManager_GetPackageManager(t *testing.T) {
	pm := &yum.PackageManager{}

	// Test contract: Should always return "yum"
	if pm.GetPackageManager() != "yum" {
		t.Errorf("GetPackageManager() should return 'yum', got '%s'", pm.GetPackageManager())
	}
}

// TestFind_BehaviorWithFixtures tests the Find operation behavior using real command output fixtures
func TestFind_BehaviorWithFixtures(t *testing.T) {
	testCases := []struct {
		name             string
		fixture          string
		expectPackages   bool
		expectVimPackage bool
	}{
		{
			name:             "vim search results",
			fixture:          "search-vim-rocky8.txt",
			expectPackages:   true,
			expectVimPackage: true,
		},
		{
			name:             "nginx search results",
			fixture:          "search-nginx-rocky8.txt",
			expectPackages:   true,
			expectVimPackage: false,
		},
		{
			name:             "empty search results",
			fixture:          "search-empty-rocky8.txt",
			expectPackages:   false,
			expectVimPackage: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := loadFixture(t, tc.fixture)
			packages := yum.ParseFindOutput(fixture, &manager.Options{})

			// Test behavior expectations
			if tc.expectPackages && len(packages) == 0 {
				t.Error("Expected packages but got none")
			}
			if !tc.expectPackages && len(packages) > 0 {
				t.Errorf("Expected no packages but got %d", len(packages))
			}

			// Test vim-specific expectations
			if tc.expectVimPackage {
				found := false
				for _, pkg := range packages {
					if pkg.Name == "vim-enhanced" {
						found = true
						break
					}
				}
				if !found {
					t.Error("Expected to find vim-enhanced package in vim search results")
				}
			}

			// Test contract: All returned packages should have expected fields
			for _, pkg := range packages {
				if pkg.Name == "" {
					t.Error("Package name should not be empty")
				}
				if pkg.PackageManager != "yum" {
					t.Errorf("Package manager should be 'yum', got '%s'", pkg.PackageManager)
				}
				if pkg.Arch == "" {
					t.Error("Package architecture should not be empty")
				}
			}
		})
	}
}

// TestListInstalled_BehaviorWithFixtures tests the ListInstalled operation behavior
func TestListInstalled_BehaviorWithFixtures(t *testing.T) {
	testCases := []struct {
		name    string
		fixture string
	}{
		{
			name:    "full system packages",
			fixture: "list-installed-rocky8.txt",
		},
		{
			name:    "minimal system packages",
			fixture: "list-installed-minimal-rocky8.txt",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := loadFixture(t, tc.fixture)
			packages := yum.ParseListInstalledOutput(fixture, &manager.Options{})

			// Test behavior: ListInstalled should return installed packages
			if len(packages) == 0 {
				t.Error("ListInstalled should return packages on a system with installed packages")
			}

			// Test contract: All returned packages should have expected fields
			for _, pkg := range packages {
				if pkg.Name == "" {
					t.Error("Package name should not be empty")
				}
				if pkg.Status != manager.PackageStatusInstalled {
					t.Errorf("Installed packages should have status 'installed', got '%s'", pkg.Status)
				}
				if pkg.Version == "" {
					t.Error("Installed packages should have Version populated")
				}
				if pkg.Arch == "" {
					t.Error("Installed packages should have Architecture populated")
				}
				if pkg.PackageManager != "yum" {
					t.Errorf("Package manager should be 'yum', got '%s'", pkg.PackageManager)
				}
			}

			// Test specific packages that should exist
			if tc.name == "minimal system packages" {
				foundRpm := false
				foundPerlDBD := false
				for _, pkg := range packages {
					if pkg.Name == "rpm" && pkg.Arch == "x86_64" {
						foundRpm = true
					}
					if pkg.Name == "perl-DBD-MySQL" && pkg.Arch == "x86_64" {
						foundPerlDBD = true
					}
				}
				if !foundRpm {
					t.Error("Expected to find rpm package in installed packages")
				}
				if !foundPerlDBD {
					t.Error("Expected to find perl-DBD-MySQL package (test for packages with dots in name)")
				}
			}
		})
	}
}

// TestGetPackageInfo_BehaviorWithFixtures tests the GetPackageInfo operation behavior
func TestGetPackageInfo_BehaviorWithFixtures(t *testing.T) {
	testCases := []struct {
		name         string
		fixture      string
		expectError  bool
		expectedName string
	}{
		{
			name:         "available package info",
			fixture:      "info-nginx-rocky8.txt",
			expectError:  false,
			expectedName: "nginx",
		},
		{
			name:         "installed package info",
			fixture:      "info-vim-installed-rocky8.txt",
			expectError:  false,
			expectedName: "vim-enhanced",
		},
		{
			name:         "available vim package info",
			fixture:      "info-vim-rocky8.txt",
			expectError:  false,
			expectedName: "vim-enhanced",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := loadFixture(t, tc.fixture)
			pkg := yum.ParsePackageInfoOutput(fixture, &manager.Options{})

			if !tc.expectError {
				// Test contract: Package info should have expected fields
				if pkg.Name != tc.expectedName {
					t.Errorf("Expected package name '%s', got '%s'", tc.expectedName, pkg.Name)
				}
				if pkg.Version == "" {
					t.Error("Package version should not be empty")
				}
				if pkg.Arch == "" {
					t.Error("Package architecture should not be empty")
				}
				if pkg.PackageManager != "yum" {
					t.Errorf("Package manager should be 'yum', got '%s'", pkg.PackageManager)
				}
			}
		})
	}
}

// TestParseEdgeCases_BehaviorWithFixtures tests parser behavior with edge cases
func TestParseEdgeCases_BehaviorWithFixtures(t *testing.T) {
	t.Run("search no results", func(t *testing.T) {
		fixture := loadFixture(t, "search-empty-rocky8.txt")
		packages := yum.ParseFindOutput(fixture, &manager.Options{})

		// Test behavior: Should handle empty results gracefully
		if len(packages) != 0 {
			t.Errorf("Expected no packages for empty search, got %d", len(packages))
		}
	})

	t.Run("package info not found", func(t *testing.T) {
		fixture := loadFixture(t, "info-notfound-rocky8.txt")
		pkg := yum.ParsePackageInfoOutput(fixture, &manager.Options{})

		// Test behavior: Should handle not found gracefully
		if pkg.Name != "" {
			t.Errorf("Expected empty package for not found, got name '%s'", pkg.Name)
		}
	})
}

// TestComplexPackageNames tests parsing of packages with dots, dashes, and complex names
func TestComplexPackageNames(t *testing.T) {
	t.Run("packages with dots in names", func(t *testing.T) {
		fixture := loadFixture(t, "search-nginx-rocky8.txt")
		packages := yum.ParseFindOutput(fixture, &manager.Options{})

		// Test critical parsing: packages with dots should be handled correctly
		foundPerlDBD := false
		foundLibreOffice := false
		for _, pkg := range packages {
			if pkg.Name == "perl-DBD-MySQL" && pkg.Arch == "x86_64" {
				foundPerlDBD = true
			}
			if pkg.Name == "libreoffice-langpack-en" && pkg.Arch == "x86_64" {
				foundLibreOffice = true
			}
		}
		if !foundPerlDBD {
			t.Error("Failed to correctly parse package with dots: perl-DBD-MySQL.x86_64")
		}
		if !foundLibreOffice {
			t.Error("Failed to correctly parse package with dots: libreoffice-langpack-en.x86_64")
		}
	})
}
