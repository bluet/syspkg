package apt

import (
	"context"
	"strings"
	"testing"

	"github.com/bluet/syspkg/manager"
)

// MockCommandRunner for testing
type MockCommandRunner struct {
	commands map[string]string // command -> output
	errors   map[string]error  // command -> error
}

func NewMockCommandRunner() *MockCommandRunner {
	return &MockCommandRunner{
		commands: make(map[string]string),
		errors:   make(map[string]error),
	}
}

func (m *MockCommandRunner) SetOutput(command, output string) {
	m.commands[command] = output
}

func (m *MockCommandRunner) SetError(command string, err error) {
	m.errors[command] = err
}

func (m *MockCommandRunner) Run(command string, args ...string) ([]byte, error) {
	fullCmd := command + " " + strings.Join(args, " ")
	if err, exists := m.errors[fullCmd]; exists {
		return nil, err
	}
	if output, exists := m.commands[fullCmd]; exists {
		return []byte(output), nil
	}
	return []byte(""), nil
}

func (m *MockCommandRunner) RunContext(ctx context.Context, command string, args []string, env ...string) ([]byte, error) {
	return m.Run(command, args...)
}

func (m *MockCommandRunner) RunInteractive(ctx context.Context, command string, args []string, env ...string) error {
	_, err := m.RunContext(ctx, command, args, env...)
	return err
}

func TestManagerBasicInfo(t *testing.T) {
	runner := NewMockCommandRunner()
	runner.SetOutput("apt --version", "apt 2.4.14 (amd64)")

	mgr := NewManagerWithRunner(runner)

	// Test basic info
	if mgr.GetName() != "apt" {
		t.Errorf("Expected name 'apt', got '%s'", mgr.GetName())
	}

	if mgr.GetType() != manager.TypeSystem {
		t.Errorf("Expected type '%s', got '%s'", manager.TypeSystem, mgr.GetType())
	}

	// Test version
	version, err := mgr.GetVersion()
	if err != nil {
		t.Fatalf("GetVersion failed: %v", err)
	}

	if !strings.Contains(version, "2.4.14") {
		t.Errorf("Expected version to contain '2.4.14', got '%s'", version)
	}
}

func TestSearch(t *testing.T) {
	runner := NewMockCommandRunner()
	searchOutput := `Sorting...
Full Text Search...
vim/jammy 2:8.2.3458-2ubuntu2.5 amd64
  Vi IMproved - enhanced vi editor

vim-common/jammy,jammy 2:8.2.3458-2ubuntu2.5 all
  Vi IMproved - Common files`

	runner.SetOutput("apt search vim", searchOutput)

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	packages, err := mgr.Search(ctx, []string{"vim"}, nil)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(packages) != 2 {
		t.Errorf("Expected 2 packages, got %d", len(packages))
	}

	// Check first package
	pkg := packages[0]
	if pkg.Name != "vim" {
		t.Errorf("Expected name 'vim', got '%s'", pkg.Name)
	}

	if pkg.NewVersion != "2:8.2.3458-2ubuntu2.5" {
		t.Errorf("Expected version '2:8.2.3458-2ubuntu2.5', got '%s'", pkg.NewVersion)
	}

	if pkg.Status != manager.StatusAvailable {
		t.Errorf("Expected status '%s', got '%s'", manager.StatusAvailable, pkg.Status)
	}

	if pkg.Category != "jammy" {
		t.Errorf("Expected category 'jammy', got '%s'", pkg.Category)
	}

	if pkg.ManagerType != manager.TypeSystem {
		t.Errorf("Expected manager type '%s', got '%s'", manager.TypeSystem, pkg.ManagerType)
	}
}

func TestListInstalled(t *testing.T) {
	runner := NewMockCommandRunner()
	listOutput := `accountsservice 22.07.5-2ubuntu1.5 amd64
adduser 3.118ubuntu5 all
apt 2.4.14 amd64`

	runner.SetOutput("dpkg-query -W -f ${binary:Package} ${Version} ${Architecture}\n", listOutput)

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	packages, err := mgr.List(ctx, manager.FilterInstalled, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(packages) != 3 {
		t.Errorf("Expected 3 packages, got %d", len(packages))
	}

	// Check first package
	pkg := packages[0]
	if pkg.Name != "accountsservice" {
		t.Errorf("Expected name 'accountsservice', got '%s'", pkg.Name)
	}

	if pkg.Version != "22.07.5-2ubuntu1.5" {
		t.Errorf("Expected version '22.07.5-2ubuntu1.5', got '%s'", pkg.Version)
	}

	if pkg.Status != manager.StatusInstalled {
		t.Errorf("Expected status '%s', got '%s'", manager.StatusInstalled, pkg.Status)
	}

	if arch, ok := pkg.Metadata["arch"]; !ok || arch != "amd64" {
		t.Errorf("Expected arch 'amd64', got '%v'", arch)
	}
}

func TestListUpgradable(t *testing.T) {
	runner := NewMockCommandRunner()
	upgradeOutput := `Listing...
vim/jammy 2:8.2.3458-2ubuntu2.6 amd64 [upgradable from: 2:8.2.3458-2ubuntu2.5]
curl/jammy 7.81.0-1ubuntu1.20 amd64 [upgradable from: 7.81.0-1ubuntu1.19]`

	runner.SetOutput("apt list --upgradable", upgradeOutput)

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	packages, err := mgr.List(ctx, manager.FilterUpgradable, nil)
	if err != nil {
		t.Fatalf("List upgradable failed: %v", err)
	}

	if len(packages) != 2 {
		t.Errorf("Expected 2 packages, got %d", len(packages))
	}

	// Check first package
	pkg := packages[0]
	if pkg.Name != "vim" {
		t.Errorf("Expected name 'vim', got '%s'", pkg.Name)
	}

	if pkg.Version != "2:8.2.3458-2ubuntu2.5" {
		t.Errorf("Expected current version '2:8.2.3458-2ubuntu2.5', got '%s'", pkg.Version)
	}

	if pkg.NewVersion != "2:8.2.3458-2ubuntu2.6" {
		t.Errorf("Expected new version '2:8.2.3458-2ubuntu2.6', got '%s'", pkg.NewVersion)
	}

	if pkg.Status != manager.StatusUpgradable {
		t.Errorf("Expected status '%s', got '%s'", manager.StatusUpgradable, pkg.Status)
	}
}

func TestInstall(t *testing.T) {
	runner := NewMockCommandRunner()
	installOutput := `Reading package lists...
Building dependency tree...
Reading state information...
The following NEW packages will be installed:
  vim
0 upgraded, 1 newly installed, 0 to remove and 0 not upgraded.
Need to get 1,315 kB of archives.
After this operation, 3,031 kB of additional disk space will be used.
Get:1 http://archive.ubuntu.com/ubuntu jammy/main amd64 vim amd64 2:8.2.3458-2ubuntu2.5 [1,315 kB]
Fetched 1,315 kB in 1s (1,234 kB/s)
Selecting previously unselected package vim.
(Reading database ... 123456 files and directories currently installed.)
Preparing to unpack .../vim_2%3a8.2.3458-2ubuntu2.5_amd64.deb ...
Unpacking vim (2:8.2.3458-2ubuntu2.5) ...
Setting up vim (2:8.2.3458-2ubuntu2.5) ...
Processing triggers for man-db (2.10.2-1) ...`

	runner.SetOutput("apt install -y vim", installOutput)

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	packages, err := mgr.Install(ctx, []string{"vim"}, nil)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if len(packages) != 1 {
		t.Errorf("Expected 1 package, got %d", len(packages))
	}

	// Check installed package
	pkg := packages[0]
	if pkg.Name != "vim" {
		t.Errorf("Expected name 'vim', got '%s'", pkg.Name)
	}

	if pkg.Version != "2:8.2.3458-2ubuntu2.5" {
		t.Errorf("Expected version '2:8.2.3458-2ubuntu2.5', got '%s'", pkg.Version)
	}

	if pkg.Status != manager.StatusInstalled {
		t.Errorf("Expected status '%s', got '%s'", manager.StatusInstalled, pkg.Status)
	}
}

func TestRemove(t *testing.T) {
	runner := NewMockCommandRunner()
	removeOutput := `Reading package lists...
Building dependency tree...
Reading state information...
The following packages will be REMOVED:
  vim
0 upgraded, 0 newly installed, 1 to remove and 0 not upgraded.
After this operation, 3,031 kB disk space will be freed.
(Reading database ... 123456 files and directories currently installed.)
Removing vim (2:8.2.3458-2ubuntu2.5) ...
Processing triggers for man-db (2.10.2-1) ...`

	runner.SetOutput("apt remove -y --autoremove vim", removeOutput)

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	packages, err := mgr.Remove(ctx, []string{"vim"}, nil)
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if len(packages) != 1 {
		t.Errorf("Expected 1 package, got %d", len(packages))
	}

	// Check removed package
	pkg := packages[0]
	if pkg.Name != "vim" {
		t.Errorf("Expected name 'vim', got '%s'", pkg.Name)
	}

	if pkg.Version != "2:8.2.3458-2ubuntu2.5" {
		t.Errorf("Expected version '2:8.2.3458-2ubuntu2.5', got '%s'", pkg.Version)
	}

	if pkg.Status != manager.StatusAvailable {
		t.Errorf("Expected status '%s', got '%s'", manager.StatusAvailable, pkg.Status)
	}
}

func TestGetInfo(t *testing.T) {
	runner := NewMockCommandRunner()
	infoOutput := `Package: vim
Version: 2:8.2.3458-2ubuntu2.5
Architecture: amd64
Maintainer: Ubuntu Developers <ubuntu-devel-discuss@lists.ubuntu.com>
Installed-Size: 3031
Depends: vim-common (= 2:8.2.3458-2ubuntu2.5), vim-runtime (= 2:8.2.3458-2ubuntu2.5), libacl1 (>= 2.2.23)
Section: editors
Priority: optional
Homepage: https://www.vim.org/
Description: Vi IMproved - enhanced vi editor
 Vim is an almost compatible version of the UNIX editor Vi.`

	runner.SetOutput("apt-cache show vim", infoOutput)

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	pkg, err := mgr.GetInfo(ctx, "vim", nil)
	if err != nil {
		t.Fatalf("GetInfo failed: %v", err)
	}

	if pkg.Name != "vim" {
		t.Errorf("Expected name 'vim', got '%s'", pkg.Name)
	}

	if pkg.Version != "2:8.2.3458-2ubuntu2.5" {
		t.Errorf("Expected version '2:8.2.3458-2ubuntu2.5', got '%s'", pkg.Version)
	}

	if pkg.Category != "editors" {
		t.Errorf("Expected category 'editors', got '%s'", pkg.Category)
	}

	if !strings.Contains(pkg.Description, "Vi IMproved") {
		t.Errorf("Expected description to contain 'Vi IMproved', got '%s'", pkg.Description)
	}

	if arch, ok := pkg.Metadata["arch"]; !ok || arch != "amd64" {
		t.Errorf("Expected arch 'amd64', got '%v'", arch)
	}
}

func TestRefresh(t *testing.T) {
	runner := NewMockCommandRunner()
	runner.SetOutput("apt update", "Hit:1 http://archive.ubuntu.com/ubuntu jammy InRelease\nReading package lists... Done")

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	err := mgr.Refresh(ctx, nil)
	if err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}
}

func TestClean(t *testing.T) {
	runner := NewMockCommandRunner()
	runner.SetOutput("apt autoclean", "Reading package lists... Done\nBuilding dependency tree... Done")

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	err := mgr.Clean(ctx, nil)
	if err != nil {
		t.Fatalf("Clean failed: %v", err)
	}
}

func TestAutoRemove(t *testing.T) {
	runner := NewMockCommandRunner()
	autoRemoveOutput := `Reading package lists...
Building dependency tree...
Reading state information...
The following packages will be REMOVED:
  libvim-dev
0 upgraded, 0 newly installed, 1 to remove and 0 not upgraded.
After this operation, 1,234 kB disk space will be freed.
Removing libvim-dev (1.0.0) ...`

	runner.SetOutput("apt autoremove -y", autoRemoveOutput)

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	packages, err := mgr.AutoRemove(ctx, nil)
	if err != nil {
		t.Fatalf("AutoRemove failed: %v", err)
	}

	if len(packages) != 1 {
		t.Errorf("Expected 1 package, got %d", len(packages))
	}

	pkg := packages[0]
	if pkg.Name != "libvim-dev" {
		t.Errorf("Expected name 'libvim-dev', got '%s'", pkg.Name)
	}
}

func TestVerify(t *testing.T) {
	runner := NewMockCommandRunner()
	runner.SetOutput("dpkg -s vim", "Package: vim\nStatus: install ok installed\nVersion: 2:8.2.3458-2ubuntu2.5")
	runner.SetError("dpkg -s nonexistent", &ExitError{ExitCode: 1})

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	// Test successful verification
	packages, err := mgr.Verify(ctx, []string{"vim"}, nil)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if len(packages) != 1 {
		t.Errorf("Expected 1 package, got %d", len(packages))
	}

	pkg := packages[0]
	if pkg.Name != "vim" {
		t.Errorf("Expected name 'vim', got '%s'", pkg.Name)
	}

	if pkg.Status != manager.StatusInstalled {
		t.Errorf("Expected status '%s', got '%s'", manager.StatusInstalled, pkg.Status)
	}

	// Test failed verification
	packages, err = mgr.Verify(ctx, []string{"nonexistent"}, nil)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if len(packages) != 1 {
		t.Errorf("Expected 1 package, got %d", len(packages))
	}

	pkg = packages[0]
	if pkg.Status != "broken" {
		t.Errorf("Expected status 'broken', got '%s'", pkg.Status)
	}
}

func TestStatus(t *testing.T) {
	runner := NewMockCommandRunner()
	runner.SetOutput("apt --version", "apt 2.4.14 (amd64)")
	runner.SetOutput("apt list --installed", "Listing...\nvim/jammy 2:8.2.3458-2ubuntu2.5 amd64 [installed]")

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()

	status, err := mgr.Status(ctx, nil)
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	if !status.Available {
		t.Error("Expected status.Available to be true")
	}

	if !status.Healthy {
		t.Error("Expected status.Healthy to be true")
	}

	if !strings.Contains(status.Version, "2.4.14") {
		t.Errorf("Expected version to contain '2.4.14', got '%s'", status.Version)
	}
}

func TestDryRun(t *testing.T) {
	runner := NewMockCommandRunner()
	runner.SetOutput("apt install -y --dry-run vim", "NOTE: This is only a simulation!")

	mgr := NewManagerWithRunner(runner)
	ctx := context.Background()
	opts := &manager.Options{DryRun: true}

	// Should not fail in dry run mode
	packages, err := mgr.Install(ctx, []string{"vim"}, opts)
	if err != nil {
		t.Fatalf("Install dry run failed: %v", err)
	}

	// Should return empty packages in dry run (no actual installation)
	if len(packages) != 0 {
		t.Errorf("Expected 0 packages in dry run, got %d", len(packages))
	}
}

func TestPluginRegistration(t *testing.T) {
	// Test that plugin auto-registers
	registry := manager.GetGlobalRegistry()
	managers := registry.GetAvailable()

	aptMgr, exists := managers["apt"]
	if !exists {
		t.Fatal("APT plugin not registered")
	}

	if aptMgr.GetName() != "apt" {
		t.Errorf("Expected name 'apt', got '%s'", aptMgr.GetName())
	}

	if aptMgr.GetType() != manager.TypeSystem {
		t.Errorf("Expected type '%s', got '%s'", manager.TypeSystem, aptMgr.GetType())
	}

	// Test plugin priority
	plugin := &Plugin{}
	if plugin.GetPriority() != 90 {
		t.Errorf("Expected priority 90, got %d", plugin.GetPriority())
	}
}

// ExitError simulates command execution errors
type ExitError struct {
	ExitCode int
}

func (e *ExitError) Error() string {
	return "exit status " + string(rune(e.ExitCode))
}

func TestInputValidation(t *testing.T) {
	mgr := NewManager()
	ctx := context.Background()

	// Test invalid package names
	invalidNames := []string{"invalid;package", "bad&&name", "package`injection`"}

	for _, name := range invalidNames {
		_, err := mgr.Search(ctx, []string{name}, nil)
		if err == nil {
			t.Errorf("Expected error for invalid package name '%s'", name)
		}

		_, err = mgr.Install(ctx, []string{name}, nil)
		if err == nil {
			t.Errorf("Expected error for invalid package name '%s'", name)
		}

		_, err = mgr.Remove(ctx, []string{name}, nil)
		if err == nil {
			t.Errorf("Expected error for invalid package name '%s'", name)
		}

		_, err = mgr.GetInfo(ctx, name, nil)
		if err == nil {
			t.Errorf("Expected error for invalid package name '%s'", name)
		}
	}
}
