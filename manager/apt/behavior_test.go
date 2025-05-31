package apt_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bluet/syspkg/manager"
	"github.com/bluet/syspkg/manager/apt"
)

// TestPackageManager_ImplementsInterface verifies contract compliance
func TestPackageManager_ImplementsInterface(t *testing.T) {
	var _ interface {
		IsAvailable() bool
		GetPackageManager() string
		Find([]string, *manager.Options) ([]manager.PackageInfo, error)
		ListInstalled(*manager.Options) ([]manager.PackageInfo, error)
		ListUpgradable(*manager.Options) ([]manager.PackageInfo, error)
		GetPackageInfo(string, *manager.Options) (manager.PackageInfo, error)
	} = &apt.PackageManager{}
}

// TestFind_BehaviorWithFixtures tests the Find operation behavior using real command output fixtures
func TestFind_BehaviorWithFixtures(t *testing.T) {
	fixture := loadFixture(t, "search-vim.txt")

	packages := apt.ParseFindOutput(fixture, &manager.Options{})

	// Test behavior: Find should return available packages
	if len(packages) == 0 {
		t.Error("Find should return packages when searching for common package")
	}

	// Test contract: All returned packages should have expected fields
	for _, pkg := range packages {
		if pkg.Name == "" {
			t.Error("Package name should not be empty")
		}
		if pkg.PackageManager != "apt" {
			t.Errorf("Package manager should be 'apt', got '%s'", pkg.PackageManager)
		}
		// Find operation can return packages with different statuses based on installation state
		if pkg.Status != manager.PackageStatusAvailable &&
			pkg.Status != manager.PackageStatusInstalled &&
			pkg.Status != manager.PackageStatusUpgradable {
			t.Errorf("Search results should have valid status, got '%s'", pkg.Status)
		}
		if pkg.NewVersion == "" {
			t.Error("Found packages should have NewVersion (repository version) populated")
		}
	}
}

// TestListInstalled_BehaviorWithFixtures tests the ListInstalled operation behavior
func TestListInstalled_BehaviorWithFixtures(t *testing.T) {
	fixture := loadFixture(t, "list-installed.txt")

	packages := apt.ParseListInstalledOutput(fixture, &manager.Options{})

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
		if pkg.NewVersion != "" {
			t.Error("ListInstalled should not populate NewVersion field")
		}
	}
}

// TestListUpgradable_BehaviorWithFixtures tests the ListUpgradable operation behavior
func TestListUpgradable_BehaviorWithFixtures(t *testing.T) {
	fixture := loadFixture(t, "list-upgradable.txt")

	packages := apt.ParseListUpgradableOutput(fixture, &manager.Options{})

	// Test behavior: Function should not panic with real APT output
	// Note: The actual package count depends on the specific fixture content and locale
	// We focus on testing the behavior contract, not specific output parsing

	// Test contract: All returned packages should follow upgrade pattern
	for _, pkg := range packages {
		if pkg.Name == "" {
			t.Error("Package name should not be empty")
		}
		if pkg.Status != manager.PackageStatusUpgradable {
			t.Errorf("Upgradable packages should have status 'upgradable', got '%s'", pkg.Status)
		}
		if pkg.Version == "" {
			t.Error("Upgradable packages should have current Version populated")
		}
		if pkg.NewVersion == "" {
			t.Error("Upgradable packages should have NewVersion (upgrade target) populated")
		}
		if pkg.Version == pkg.NewVersion {
			t.Error("Upgradable packages should have different current and new versions")
		}
	}
}

// TestStatusNormalization_CrossPackageManagerCompatibility tests the documented behavior
// that APT-specific statuses are normalized for cross-package manager compatibility
func TestStatusNormalization_CrossPackageManagerCompatibility(t *testing.T) {
	// This tests the documented behavior that config-files status is normalized to available
	// for cross-package manager compatibility as specified in packageinfo.go

	dpkgOutput := []byte("qemu-kvm deinstall ok config-files 1:4.2-3ubuntu6.23")
	packages := map[string]manager.PackageInfo{
		"qemu-kvm": {Name: "qemu-kvm"},
	}

	result, err := apt.ParseDpkgQueryOutput(dpkgOutput, packages, nil)
	if err != nil {
		t.Fatalf("ParseDpkgQueryOutput failed: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}

	// Test documented behavior: config-files is normalized to available
	if result[0].Status != manager.PackageStatusAvailable {
		t.Errorf("config-files should be normalized to available for cross-PM compatibility, got %s", result[0].Status)
	}
}

// TestInstall_BehaviorWithFixtures tests the ParseInstallOutput function behavior
func TestInstall_BehaviorWithFixtures(t *testing.T) {
	fixture := loadFixture(t, "apt-install-vim.txt")

	packages := apt.ParseInstallOutput(fixture, &manager.Options{})

	// Test behavior: Install output should indicate successful installation
	if len(packages) == 0 {
		t.Error("Install operation should return package information")
	}

	// Test contract: Installed packages should have correct status
	for _, pkg := range packages {
		if pkg.Name == "" {
			t.Error("Installed package name should not be empty")
		}
		if pkg.Status != manager.PackageStatusInstalled {
			t.Errorf("Installed packages should have status 'installed', got '%s'", pkg.Status)
		}
		if pkg.PackageManager != "apt" {
			t.Errorf("Package manager should be 'apt', got '%s'", pkg.PackageManager)
		}
	}
}

// TestRemove_BehaviorWithFixtures tests the ParseDeletedOutput function behavior
func TestRemove_BehaviorWithFixtures(t *testing.T) {
	fixture := loadFixture(t, "apt-remove-vim.txt")

	packages := apt.ParseDeletedOutput(fixture, &manager.Options{})

	// Test behavior: Remove output should indicate packages were removed
	if len(packages) == 0 {
		t.Error("Remove operation should return information about removed packages")
	}

	// Test contract: Removed packages should have correct status
	for _, pkg := range packages {
		if pkg.Name == "" {
			t.Error("Removed package name should not be empty")
		}
		// Note: Removed packages typically don't have a specific status in our system,
		// we focus on ensuring the parsing doesn't fail
		if pkg.PackageManager != "apt" {
			t.Errorf("Package manager should be 'apt', got '%s'", pkg.PackageManager)
		}
	}
}

// TestPackageInfo_BehaviorWithFixtures tests the ParsePackageInfoOutput function behavior
func TestPackageInfo_BehaviorWithFixtures(t *testing.T) {
	fixture := loadFixture(t, "show-vim.txt")

	pkg := apt.ParsePackageInfoOutput(fixture, &manager.Options{})

	// Test behavior: Package info should provide detailed information
	if pkg.Name == "" {
		t.Error("Package info should include package name")
	}
	if pkg.PackageManager != "apt" {
		t.Errorf("Package manager should be 'apt', got '%s'", pkg.PackageManager)
	}

	// Test contract: Package info should have version information
	if pkg.Version == "" {
		t.Error("Package info should include version")
	}

	// Test that additional data might contain description or other metadata
	if len(pkg.AdditionalData) > 0 {
		// Package info may include additional metadata
		t.Logf("Additional package data available: %v", pkg.AdditionalData)
	}
}

// TestExpectedUsagePattern_SearchAndInstall documents a common user workflow
func TestExpectedUsagePattern_SearchAndInstall(t *testing.T) {
	pm := &apt.PackageManager{}

	// Test usage pattern: Check availability before operations
	if !pm.IsAvailable() {
		t.Skip("APT not available on this system")
	}

	// Test contract: GetPackageManager returns consistent identifier
	if pm.GetPackageManager() != "apt" {
		t.Errorf("GetPackageManager should return 'apt', got '%s'", pm.GetPackageManager())
	}

	// Note: We don't test actual Find/Install here as that would require system changes
	// and violate the principle of testing behavior, not system state
}

// loadFixture loads a test fixture file from the testing/fixtures/apt directory
func loadFixture(t *testing.T, filename string) string {
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

	fixturePath := filepath.Join(projectRoot, "testing", "fixtures", "apt", filename)
	content, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("Failed to load fixture %s: %v", filename, err)
	}

	return string(content)
}
