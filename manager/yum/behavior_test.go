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

// TestParseFindOutput_BehaviorWithFixtures tests the ParseFindOutput parser behavior using real YUM search command output fixtures
// Note: This tests the parser function directly, not the Find() method which enhances results with rpm -q
func TestParseFindOutput_BehaviorWithFixtures(t *testing.T) {
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

				// Test ParseFindOutput limitation: Status is always available since YUM search output
				// doesn't include installation status (the Find() method enhances this with rpm -q)
				if pkg.Status != manager.PackageStatusAvailable {
					t.Errorf("ParseFindOutput should always return PackageStatusAvailable due to YUM search output limitation, got '%s'", pkg.Status)
				}

				// Test ParseFindOutput limitation: Version fields are not populated by YUM search output
				// (the Find() method enhances this for installed packages)
				if pkg.Version != "" {
					t.Errorf("ParseFindOutput should not populate Version field from search output, got '%s'", pkg.Version)
				}
				if pkg.NewVersion != "" {
					t.Errorf("ParseFindOutput should not populate NewVersion field from search output, got '%s'", pkg.NewVersion)
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
		name           string
		fixture        string
		expectError    bool
		expectedName   string
		expectedStatus manager.PackageStatus
	}{
		{
			name:           "available package info",
			fixture:        "info-nginx-rocky8.txt",
			expectError:    false,
			expectedName:   "nginx",
			expectedStatus: manager.PackageStatusAvailable,
		},
		{
			name:           "installed package info",
			fixture:        "info-vim-installed-rocky8.txt",
			expectError:    false,
			expectedName:   "vim-enhanced",
			expectedStatus: manager.PackageStatusInstalled,
		},
		{
			name:           "available vim package info",
			fixture:        "info-vim-rocky8.txt",
			expectError:    false,
			expectedName:   "vim-enhanced",
			expectedStatus: manager.PackageStatusAvailable,
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

				// Test behavior: Status should be determined by section header
				if pkg.Status != tc.expectedStatus {
					t.Errorf("Expected status '%s' based on section, got '%s'", tc.expectedStatus, pkg.Status)
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

// TestYUM_CrossPackageManagerCompatibility documents YUM status detection capabilities
// and ensures cross-package manager API consistency
func TestYUM_CrossPackageManagerCompatibility(t *testing.T) {
	t.Run("Find operation status detection", func(t *testing.T) {
		// Document YUM's enhanced status detection capability
		fixture := loadFixture(t, "search-vim-rocky8.txt")
		packages := yum.ParseFindOutput(fixture, &manager.Options{})

		if len(packages) == 0 {
			t.Fatal("Expected packages in search results")
		}

		// YUM now provides accurate status detection by using rpm -q
		// In test environment, packages should show 'available' since they're not installed
		for _, pkg := range packages {
			if pkg.Status != manager.PackageStatusAvailable && pkg.Status != manager.PackageStatusInstalled {
				t.Errorf("YUM Find should return installed or available status, got %s", pkg.Status)
			}

			// Verify that package has proper metadata
			if pkg.Name == "" {
				t.Error("Package name should not be empty")
			}
			if pkg.PackageManager != "yum" {
				t.Errorf("Package manager should be 'yum', got '%s'", pkg.PackageManager)
			}
		}

		t.Log("YUM Find() now provides accurate installation status via rpm -q, ensuring API consistency with APT")
	})

	t.Run("GetPackageInfo provides accurate status", func(t *testing.T) {
		// Test that GetPackageInfo correctly determines status from section headers
		installedFixture := loadFixture(t, "info-vim-installed-rocky8.txt")
		installedPkg := yum.ParsePackageInfoOutput(installedFixture, &manager.Options{})

		if installedPkg.Status != manager.PackageStatusInstalled {
			t.Errorf("GetPackageInfo should detect installed status from 'Installed Packages' section, got %s", installedPkg.Status)
		}

		availableFixture := loadFixture(t, "info-vim-rocky8.txt")
		availablePkg := yum.ParsePackageInfoOutput(availableFixture, &manager.Options{})

		if availablePkg.Status != manager.PackageStatusAvailable {
			t.Errorf("GetPackageInfo should detect available status from 'Available Packages' section, got %s", availablePkg.Status)
		}
	})
}

// TestInstall_BehaviorWithFixtures tests the Install operation behavior
func TestInstall_BehaviorWithFixtures(t *testing.T) {
	testCases := []struct {
		name           string
		fixture        string
		expectPackages bool
		expectError    bool
	}{
		{
			name:           "successful install with dependencies",
			fixture:        "install-vim-rocky8.txt",
			expectPackages: true,
			expectError:    false,
		},
		{
			name:           "install multiple packages",
			fixture:        "install-multiple-rocky8.txt",
			expectPackages: true,
			expectError:    false,
		},
		{
			name:           "install already installed package",
			fixture:        "install-already-installed-rocky8.txt",
			expectPackages: false, // "Nothing to do" case
			expectError:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := loadFixture(t, tc.fixture)
			packages := yum.ParseInstallOutput(fixture, &manager.Options{})

			// Test behavior expectations
			if tc.expectPackages && len(packages) == 0 {
				t.Error("Expected packages to be installed but got none")
			}
			if !tc.expectPackages && len(packages) > 0 {
				t.Errorf("Expected no packages to be installed but got %d", len(packages))
			}

			// Test contract: All returned packages should have expected fields
			for _, pkg := range packages {
				if pkg.Name == "" {
					t.Error("Installed package name should not be empty")
				}
				if pkg.Status != manager.PackageStatusInstalled {
					t.Errorf("Installed packages should have status 'installed', got '%s'", pkg.Status)
				}
				if pkg.Version == "" {
					t.Error("Installed packages should have Version populated")
				}
				if pkg.NewVersion != pkg.Version {
					t.Errorf("For install, NewVersion should equal Version, got Version='%s' NewVersion='%s'", pkg.Version, pkg.NewVersion)
				}
				if pkg.PackageManager != "yum" {
					t.Errorf("Package manager should be 'yum', got '%s'", pkg.PackageManager)
				}
			}
		})
	}
}

// TestDelete_BehaviorWithFixtures tests the Delete operation behavior
func TestDelete_BehaviorWithFixtures(t *testing.T) {
	testCases := []struct {
		name           string
		fixture        string
		expectPackages bool
	}{
		{
			name:           "successful remove single package",
			fixture:        "remove-tree-rocky8.txt",
			expectPackages: true,
		},
		{
			name:           "remove with dependencies",
			fixture:        "remove-nginx-rocky8.txt",
			expectPackages: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := loadFixture(t, tc.fixture)
			packages := yum.ParseDeleteOutput(fixture, &manager.Options{})

			// Test behavior expectations
			if tc.expectPackages && len(packages) == 0 {
				t.Error("Expected packages to be removed but got none")
			}

			// Test contract: All returned packages should have expected fields
			for _, pkg := range packages {
				if pkg.Name == "" {
					t.Error("Removed package name should not be empty")
				}
				if pkg.Status != manager.PackageStatusAvailable {
					t.Errorf("Removed packages should have status 'available', got '%s'", pkg.Status)
				}
				if pkg.Version == "" {
					t.Error("Removed packages should have Version populated")
				}
				if pkg.NewVersion != "" {
					t.Errorf("For delete, NewVersion should be empty, got '%s'", pkg.NewVersion)
				}
				if pkg.PackageManager != "yum" {
					t.Errorf("Package manager should be 'yum', got '%s'", pkg.PackageManager)
				}
			}
		})
	}
}

// TestListUpgradable_BehaviorWithFixtures tests the ListUpgradable operation behavior
func TestListUpgradable_BehaviorWithFixtures(t *testing.T) {
	testCases := []struct {
		name    string
		fixture string
	}{
		{
			name:    "check for available updates",
			fixture: "check-update-rocky8.txt",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := loadFixture(t, tc.fixture)
			packages := yum.ParseListUpgradableOutput(fixture, &manager.Options{})

			// Test behavior: ListUpgradable should not fail with real YUM output
			// Note: The actual package count depends on the specific fixture content

			// Test contract: All returned packages should have expected fields
			for _, pkg := range packages {
				if pkg.Name == "" {
					t.Error("Upgradable package name should not be empty")
				}
				if pkg.Status != manager.PackageStatusUpgradable {
					t.Errorf("Upgradable packages should have status 'upgradable', got '%s'", pkg.Status)
				}
				if pkg.NewVersion == "" {
					t.Error("Upgradable packages should have NewVersion populated")
				}
				// Note: Current version may not be provided by yum check-update
				if pkg.PackageManager != "yum" {
					t.Errorf("Package manager should be 'yum', got '%s'", pkg.PackageManager)
				}
			}
		})
	}
}

// TestAutoRemove_BehaviorWithFixtures tests the AutoRemove operation behavior
func TestAutoRemove_BehaviorWithFixtures(t *testing.T) {
	testCases := []struct {
		name    string
		fixture string
	}{
		{
			name:    "autoremove unused dependencies",
			fixture: "autoremove-rocky8.txt",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := loadFixture(t, tc.fixture)
			packages := yum.ParseAutoRemoveOutput(fixture, &manager.Options{})

			// Test behavior: AutoRemove should not fail with real YUM output
			// Note: May return empty list if no packages to remove

			// Test contract: All returned packages should have expected fields
			for _, pkg := range packages {
				if pkg.Name == "" {
					t.Error("Auto-removed package name should not be empty")
				}
				if pkg.Status != manager.PackageStatusAvailable {
					t.Errorf("Auto-removed packages should have status 'available', got '%s'", pkg.Status)
				}
				if pkg.NewVersion != "" {
					t.Errorf("For autoremove, NewVersion should be empty, got '%s'", pkg.NewVersion)
				}
				if pkg.PackageManager != "yum" {
					t.Errorf("Package manager should be 'yum', got '%s'", pkg.PackageManager)
				}
			}
		})
	}
}

// TestUpgrade_BehaviorWithFixtures tests the Upgrade operation behavior
func TestUpgrade_BehaviorWithFixtures(t *testing.T) {
	testCases := []struct {
		name           string
		fixture        string
		expectPackages bool
		expectError    bool
	}{
		{
			name:           "no packages to upgrade (dry run)",
			fixture:        "update-all-dryrun-rocky8.txt",
			expectPackages: false, // "Nothing to do" case
			expectError:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := loadFixture(t, tc.fixture)
			packages := yum.ParseUpgradeOutput(fixture, &manager.Options{})

			// Test behavior expectations
			if tc.expectPackages && len(packages) == 0 {
				t.Error("Expected packages to be upgraded but got none")
			}
			if !tc.expectPackages && len(packages) > 0 {
				t.Errorf("Expected no packages to be upgraded but got %d", len(packages))
			}

			// Test contract: All returned packages should have expected fields
			for _, pkg := range packages {
				if pkg.Name == "" {
					t.Error("Upgraded package name should not be empty")
				}
				if pkg.Status != manager.PackageStatusInstalled {
					t.Errorf("Upgraded packages should have status 'installed', got '%s'", pkg.Status)
				}
				if pkg.Version == "" {
					t.Error("Upgraded packages should have Version populated")
				}
				if pkg.PackageManager != "yum" {
					t.Errorf("Package manager should be 'yum', got '%s'", pkg.PackageManager)
				}

				// Note: Current implementation limitation - Upgrade uses same parser as Install
				// Future enhancement: Should show Version=old_version, NewVersion=new_version for upgrades
				// Currently: Version=NewVersion=final_installed_version (same as Install behavior)
				if pkg.NewVersion != pkg.Version {
					t.Logf("INFO: Upgrade shows version transition - Version='%s' NewVersion='%s'", pkg.Version, pkg.NewVersion)
				}
			}
		})
	}
}

// TestUpgradeAll_BehaviorWithFixtures tests the UpgradeAll operation behavior
func TestUpgradeAll_BehaviorWithFixtures(t *testing.T) {
	testCases := []struct {
		name           string
		fixture        string
		expectPackages bool
	}{
		{
			name:           "no system updates available",
			fixture:        "update-all-dryrun-rocky8.txt",
			expectPackages: false, // "Nothing to do" case
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := loadFixture(t, tc.fixture)
			packages := yum.ParseUpgradeOutput(fixture, &manager.Options{})

			// Test behavior expectations
			if tc.expectPackages && len(packages) == 0 {
				t.Error("Expected packages to be upgraded but got none")
			}
			if !tc.expectPackages && len(packages) > 0 {
				t.Errorf("Expected no packages to be upgraded but got %d", len(packages))
			}

			// Test contract: All returned packages should have expected fields
			for _, pkg := range packages {
				if pkg.Name == "" {
					t.Error("Upgraded package name should not be empty")
				}
				if pkg.Status != manager.PackageStatusInstalled {
					t.Errorf("Upgraded packages should have status 'installed', got '%s'", pkg.Status)
				}
				if pkg.Version == "" {
					t.Error("Upgraded packages should have Version populated")
				}
				if pkg.PackageManager != "yum" {
					t.Errorf("Package manager should be 'yum', got '%s'", pkg.PackageManager)
				}

				// Note: Current implementation limitation - UpgradeAll uses same parser as Install
				// Future enhancement: Should show Version=old_version, NewVersion=new_version for upgrades
				// Currently: Version=NewVersion=final_installed_version (same as Install behavior)
				if pkg.NewVersion != pkg.Version {
					t.Logf("INFO: UpgradeAll shows version transition - Version='%s' NewVersion='%s'", pkg.Version, pkg.NewVersion)
				}
			}
		})
	}
}

// TestUpgrade_CrossOperationCompatibility documents behavioral differences between operations
func TestUpgrade_CrossOperationCompatibility(t *testing.T) {
	t.Run("upgrade vs install behavior documentation", func(t *testing.T) {
		// This test documents the current implementation behavior and expected differences

		// Current behavior (implementation limitation):
		// - Install: Version=NewVersion=installed_version
		// - Upgrade: Version=NewVersion=installed_version (same as Install)
		//
		// Expected behavior (future enhancement):
		// - Install: Version=NewVersion=installed_version (correct)
		// - Upgrade: Version=old_version, NewVersion=new_version (shows transition)

		t.Log("Current Implementation: Upgrade operations use ParseInstallOutput")
		t.Log("Limitation: Version transitions (old→new) are not captured")
		t.Log("Behavior: Upgrade currently behaves like Install (shows final installed version only)")
		t.Log("Future Enhancement: Should parse 'Upgrading:' section to capture version transitions")

		// Test that both operations exist and follow same interface
		// This ensures they can be enhanced independently in the future
		fixture := loadFixture(t, "update-all-dryrun-rocky8.txt")

		installResult := yum.ParseInstallOutput(fixture, &manager.Options{})
		upgradeResult := yum.ParseUpgradeOutput(fixture, &manager.Options{})

		// Currently they behave identically (both return empty for "Nothing to do")
		if len(installResult) != len(upgradeResult) {
			t.Errorf("Expected Install and Upgrade to behave identically with current implementation, got Install=%d Upgrade=%d",
				len(installResult), len(upgradeResult))
		}
	})
}
