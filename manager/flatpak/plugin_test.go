package flatpak

import (
	"context"
	"strings"
	"testing"

	"github.com/bluet/syspkg/manager"
	"github.com/bluet/syspkg/testing/testutil"
)

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

// loadFlatpakFixture loads a Flatpak fixture file
func loadFlatpakFixture(t *testing.T, filename string) string {
	t.Helper()
	return testutil.LoadFixture(t, "flatpak", filename)
}

func TestManagerBasicInfo(t *testing.T) {
	runner := NewMockCommandRunner()
	versionFixture := loadFlatpakFixture(t, "version-ubuntu2204.txt")
	runner.SetOutput("flatpak --version", versionFixture)

	mgr := NewManagerWithRunner(runner)

	// Test basic info
	if mgr.GetName() != "flatpak" {
		t.Errorf("Expected name 'flatpak', got '%s'", mgr.GetName())
	}

	if mgr.GetCategory() != manager.CategorySystem {
		t.Errorf("Expected category '%s', got '%s'", manager.CategorySystem, mgr.GetCategory())
	}

	// Test version
	version, err := mgr.GetVersion()
	if err != nil {
		t.Fatalf("GetVersion failed: %v", err)
	}

	if !strings.Contains(version, "1.12.7") {
		t.Errorf("Expected version to contain '1.12.7', got '%s'", version)
	}
}

func TestSearch(t *testing.T) {
	runner := NewMockCommandRunner()
	searchFixture := testutil.LoadFlatpakFixture(t, "search-neovim-ubuntu2204.txt")
	runner.SetOutput("flatpak search neovim", searchFixture)

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	packages, err := mgr.Search(ctx, []string{"neovim"}, nil)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(packages) == 0 {
		t.Fatalf("Expected packages from search fixture, got none")
	}

	// Check first package (Neovim)
	pkg := packages[0]
	if pkg.Name != "Neovim" {
		t.Errorf("Expected name 'Neovim', got '%s'", pkg.Name)
	}

	if pkg.Version != "0.11.2" {
		t.Errorf("Expected version '0.11.2', got '%s'", pkg.Version)
	}

	if pkg.Status != manager.StatusAvailable {
		t.Errorf("Expected status '%s', got '%s'", manager.StatusAvailable, pkg.Status)
	}

	if pkg.ManagerName != "flatpak" {
		t.Errorf("Expected manager name '%s', got '%s'", "flatpak", pkg.ManagerName)
	}

	// Check metadata
	if appId, ok := pkg.Metadata["app_id"]; !ok || appId != "io.neovim.nvim" {
		t.Errorf("Expected app_id 'io.neovim.nvim', got '%v'", appId)
	}

	if branch, ok := pkg.Metadata["branch"]; !ok || branch != "stable" {
		t.Errorf("Expected branch 'stable', got '%v'", branch)
	}

	if remotes, ok := pkg.Metadata["remotes"]; !ok || remotes != "flathub" {
		t.Errorf("Expected remotes 'flathub', got '%v'", remotes)
	}
}

func TestSearchNoResults(t *testing.T) {
	runner := NewMockCommandRunner()
	searchFixture := loadFlatpakFixture(t, "search-empty-ubuntu2204.txt")
	runner.SetOutput("flatpak search vim", searchFixture)

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	packages, err := mgr.Search(ctx, []string{"vim"}, nil)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// This fixture contains "No matches found"
	if len(packages) != 0 {
		t.Errorf("Expected no packages for 'vim' search, got %d", len(packages))
	}
}

func TestListInstalled(t *testing.T) {
	runner := NewMockCommandRunner()
	listFixture := loadFlatpakFixture(t, "list-installed-ubuntu2204.txt")
	runner.SetOutput("flatpak list --user --columns=name,version,origin", listFixture)

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	packages, err := mgr.List(ctx, manager.FilterInstalled, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	// The fixture has some installed packages
	if len(packages) == 0 {
		t.Error("Expected some installed packages from fixture")
	}

	// Verify first package details
	if len(packages) > 0 {
		pkg := packages[0]
		if pkg.Name != "shiftkey" {
			t.Errorf("Expected first package name 'shiftkey', got '%s'", pkg.Name)
		}
		if pkg.Status != manager.StatusInstalled {
			t.Errorf("Expected status 'installed', got '%s'", pkg.Status)
		}
	}
}

func TestListUpgradable(t *testing.T) {
	runner := NewMockCommandRunner()
	// Mock empty output for upgradable packages
	runner.SetOutput("flatpak list --user --updates --columns=name,version,origin", "")

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	packages, err := mgr.List(ctx, manager.FilterUpgradable, nil)
	if err != nil {
		t.Fatalf("List upgradable failed: %v", err)
	}

	// Should return empty list for no updates
	if len(packages) != 0 {
		t.Errorf("Expected no upgradable packages, got %d", len(packages))
	}
}

func TestGetInfo(t *testing.T) {
	runner := NewMockCommandRunner()
	// Mock flatpak info output
	infoOutput := `Neovim - Vim-fork focused on extensibility and usability

         ID: io.neovim.nvim
    Version: 0.11.2
     Branch: stable
       Arch: x86_64
     Origin: flathub
Collection: org.flathub.Stable
Installation: system
   Installed: 128.4 MB
    Runtime: org.freedesktop.Platform/x86_64/24.08`

	runner.SetOutput("flatpak info io.neovim.nvim", infoOutput)

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	pkg, err := mgr.GetInfo(ctx, "io.neovim.nvim", nil)
	if err != nil {
		t.Fatalf("GetInfo failed: %v", err)
	}

	if pkg.Name != "io.neovim.nvim" {
		t.Errorf("Expected name 'io.neovim.nvim', got '%s'", pkg.Name)
	}

	if pkg.Status != manager.StatusAvailable {
		t.Errorf("Expected status '%s', got '%s'", manager.StatusAvailable, pkg.Status)
	}

	if pkg.ManagerName != "flatpak" {
		t.Errorf("Expected manager name '%s', got '%s'", "flatpak", pkg.ManagerName)
	}
}

func TestStatus(t *testing.T) {
	runner := NewMockCommandRunner()
	runner.SetOutput("flatpak --version", "Flatpak 1.12.7")

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	status, err := mgr.Status(ctx, nil)
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	if !status.Available {
		t.Error("Expected Status.Available to be true")
	}

	if !status.Healthy {
		t.Error("Expected Status.Healthy to be true")
	}

	if status.Version != "1.12.7" {
		t.Errorf("Expected version '1.12.7', got '%s'", status.Version)
	}
}

func TestInstall(t *testing.T) {
	runner := NewMockCommandRunner()
	installOutput := `Installing: io.neovim.nvim/x86_64/stable from flathub
Required runtime for io.neovim.nvim/x86_64/stable (runtime/org.freedesktop.Platform/x86_64/24.08) found in remote flathub
Installing 1 new application...`

	runner.SetOutput("flatpak install -y io.neovim.nvim", installOutput)

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	packages, err := mgr.Install(ctx, []string{"io.neovim.nvim"}, nil)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if len(packages) != 1 {
		t.Fatalf("Expected 1 package installed, got %d", len(packages))
	}

	pkg := packages[0]
	if pkg.Name != "io.neovim.nvim" {
		t.Errorf("Expected name 'io.neovim.nvim', got '%s'", pkg.Name)
	}

	if pkg.Status != manager.StatusInstalled {
		t.Errorf("Expected status '%s', got '%s'", manager.StatusInstalled, pkg.Status)
	}
}

func TestInstallDryRun(t *testing.T) {
	runner := NewMockCommandRunner()
	// For dry run, the plugin calls GetInfo
	infoFixture := loadFlatpakFixture(t, "info-calculator-ubuntu2204.txt")
	runner.SetOutput("flatpak info test-package", infoFixture)

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()
	opts := &manager.Options{DryRun: true}

	packages, err := mgr.Install(ctx, []string{"test-package"}, opts)
	if err != nil {
		t.Fatalf("Install dry run failed: %v", err)
	}

	// Should return the package info with would-install status
	if len(packages) != 1 {
		t.Fatalf("Expected 1 package from dry run, got %d", len(packages))
	}

	pkg := packages[0]
	if pkg.Status != "would-install" {
		t.Errorf("Expected status 'would-install', got '%s'", pkg.Status)
	}
}

func TestRemove(t *testing.T) {
	runner := NewMockCommandRunner()
	removeOutput := `Uninstalling: io.neovim.nvim/x86_64/stable
Uninstalling 1 application...`

	runner.SetOutput("flatpak uninstall -y io.neovim.nvim", removeOutput)

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	packages, err := mgr.Remove(ctx, []string{"io.neovim.nvim"}, nil)
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if len(packages) != 1 {
		t.Fatalf("Expected 1 package removed, got %d", len(packages))
	}

	pkg := packages[0]
	if pkg.Name != "io.neovim.nvim" {
		t.Errorf("Expected name 'io.neovim.nvim', got '%s'", pkg.Name)
	}
}

func TestRemoveDryRun(t *testing.T) {
	runner := NewMockCommandRunner()
	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()
	opts := &manager.Options{DryRun: true}

	packages, err := mgr.Remove(ctx, []string{"test-package"}, opts)
	if err != nil {
		t.Fatalf("Remove dry run failed: %v", err)
	}

	// Should return packages with would-remove status for dry run
	if len(packages) != 1 {
		t.Fatalf("Expected 1 package from dry run, got %d", len(packages))
	}

	pkg := packages[0]
	if pkg.Name != "test-package" {
		t.Errorf("Expected name 'test-package', got '%s'", pkg.Name)
	}
	if pkg.Status != "would-remove" {
		t.Errorf("Expected status 'would-remove', got '%s'", pkg.Status)
	}
}

func TestRefresh(t *testing.T) {
	runner := NewMockCommandRunner()
	runner.SetOutput("flatpak update --appstream", "Updating appstream data for remote flathub\nUpdate complete.")

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	err := mgr.Refresh(ctx, nil)
	if err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}
}

func TestUpgrade(t *testing.T) {
	runner := NewMockCommandRunner()
	upgradeOutput := `Looking for updates...
Nothing to do.`

	runner.SetOutput("flatpak update -y", upgradeOutput)

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	packages, err := mgr.Upgrade(ctx, []string{}, nil)
	if err != nil {
		t.Fatalf("Upgrade failed: %v", err)
	}

	// Should return empty list when nothing to upgrade
	if len(packages) != 0 {
		t.Errorf("Expected no packages to upgrade, got %d", len(packages))
	}
}

func TestUpgradeSpecificPackages(t *testing.T) {
	runner := NewMockCommandRunner()
	upgradeOutput := `Looking for updates...
Updated: io.neovim.nvim/x86_64/stable from flathub
1 application updated.`

	runner.SetOutput("flatpak update -y io.neovim.nvim", upgradeOutput)

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	packages, err := mgr.Upgrade(ctx, []string{"io.neovim.nvim"}, nil)
	if err != nil {
		t.Fatalf("Upgrade specific packages failed: %v", err)
	}

	if len(packages) != 1 {
		t.Fatalf("Expected 1 package upgraded, got %d", len(packages))
	}

	pkg := packages[0]
	if pkg.Name != "io.neovim.nvim" {
		t.Errorf("Expected name 'io.neovim.nvim', got '%s'", pkg.Name)
	}
}

func TestClean(t *testing.T) {
	runner := NewMockCommandRunner()
	runner.SetOutput("flatpak uninstall --unused -y", "Nothing unused to uninstall")

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	err := mgr.Clean(ctx, nil)
	if err != nil {
		t.Fatalf("Clean failed: %v", err)
	}
}

func TestCleanDryRun(t *testing.T) {
	runner := NewMockCommandRunner()
	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()
	opts := &manager.Options{DryRun: true}

	err := mgr.Clean(ctx, opts)
	if err != nil {
		t.Fatalf("Clean dry run failed: %v", err)
	}

	// Test completed successfully - dry run should not execute commands
}

func TestAutoRemove(t *testing.T) {
	runner := NewMockCommandRunner()
	autoremoveOutput := `These runtimes are pinned and won't be removed:
	org.freedesktop.Platform/x86_64/24.08
Nothing unused to uninstall`

	runner.SetOutput("flatpak uninstall --unused", autoremoveOutput)

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	packages, err := mgr.AutoRemove(ctx, nil)
	if err != nil {
		t.Fatalf("AutoRemove failed: %v", err)
	}

	// Should return empty list when nothing to remove
	if len(packages) != 0 {
		t.Errorf("Expected no packages to autoremove, got %d", len(packages))
	}
}

func TestAutoRemoveDryRun(t *testing.T) {
	runner := NewMockCommandRunner()
	// Mock the dry run command (without -y)
	runner.SetOutput("flatpak uninstall --unused", "Nothing unused to uninstall")

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()
	opts := &manager.Options{DryRun: true}

	packages, err := mgr.AutoRemove(ctx, opts)
	if err != nil {
		t.Fatalf("AutoRemove dry run failed: %v", err)
	}

	// Should return empty results for dry run
	if len(packages) != 0 {
		t.Errorf("Expected no packages from dry run, got %d", len(packages))
	}
}

func TestVerify(t *testing.T) {
	runner := NewMockCommandRunner()
	// Mock verification output - Verify runs 'flatpak info package_name'
	infoOutput := `Neovim - Vim-fork focused on extensibility and usability

         ID: io.neovim.nvim
    Version: 0.11.2
     Branch: stable
       Arch: x86_64
     Origin: flathub
Collection: org.flathub.Stable
Installation: system
   Installed: 128.4 MB
    Runtime: org.freedesktop.Platform/x86_64/24.08`
	runner.SetOutput("flatpak info io.neovim.nvim", infoOutput)

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	packages, err := mgr.Verify(ctx, []string{"io.neovim.nvim"}, nil)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if len(packages) != 1 {
		t.Fatalf("Expected 1 package verified, got %d", len(packages))
	}

	pkg := packages[0]
	if pkg.Name != "io.neovim.nvim" {
		t.Errorf("Expected name 'io.neovim.nvim', got '%s'", pkg.Name)
	}

	if pkg.Status != manager.StatusInstalled {
		t.Errorf("Expected status '%s', got '%s'", manager.StatusInstalled, pkg.Status)
	}
}

func TestParseSearchOutput(t *testing.T) {
	mgr := NewManager()

	// Test parsing search output with multiple packages
	searchOutput := `Neovim	Vim-fork focused on extensibility and usability	io.neovim.nvim	0.11.2	stable	flathub
neovide	Advanced graphical interface for Neovim	dev.neovide.neovide	0.15.0	stable	flathub
Helix	A post-modern text editor	com.helix_editor.Helix	25.01	stable	flathub`

	packages := mgr.parseSearchOutput(searchOutput)

	if len(packages) != 3 {
		t.Fatalf("Expected 3 packages, got %d", len(packages))
	}

	// Test first package
	pkg := packages[0]
	if pkg.Name != "Neovim" {
		t.Errorf("Expected name 'Neovim', got '%s'", pkg.Name)
	}
	if pkg.Version != "0.11.2" {
		t.Errorf("Expected version '0.11.2', got '%s'", pkg.Version)
	}
	if pkg.Metadata["app_id"] != "io.neovim.nvim" {
		t.Errorf("Expected app_id 'io.neovim.nvim', got '%v'", pkg.Metadata["app_id"])
	}
}

func TestParseSearchOutputEmpty(t *testing.T) {
	mgr := NewManager()

	// Test parsing empty search output
	packages := mgr.parseSearchOutput("")

	if len(packages) != 0 {
		t.Errorf("Expected 0 packages from empty output, got %d", len(packages))
	}

	// Test parsing "No matches found"
	packages = mgr.parseSearchOutput("Flatpak 1.12.7\nNo matches found")

	if len(packages) != 0 {
		t.Errorf("Expected 0 packages from 'No matches found', got %d", len(packages))
	}
}

func TestIsAvailable(t *testing.T) {
	runner := NewMockCommandRunner()
	runner.SetOutput("flatpak --version", "Flatpak 1.12.7")

	mgr := NewManagerWithRunner(runner)

	if !mgr.IsAvailable() {
		t.Error("Expected IsAvailable to return true with mocked flatpak")
	}
}
