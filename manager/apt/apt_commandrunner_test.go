package apt_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/bluet/syspkg/manager"
	"github.com/bluet/syspkg/manager/apt"
)

// TestAPT_WithMockedCommands demonstrates testing APT operations with mocked CommandRunner
// This shows the cross-platform testing capability enabled by CommandRunner pattern
func TestAPT_WithMockedCommands(t *testing.T) {
	t.Run("Find with mocked commands - cross-platform testing", func(t *testing.T) {
		// Create mock command runner
		mockRunner := manager.NewMockCommandRunner()

		// Mock apt search output (typical Ubuntu/Debian output)
		searchOutput := `Sorting...
Full Text Search...
vim/jammy 2:8.2.0716-3ubuntu2 amd64
  Vi IMproved - enhanced vi editor

vim-gtk3/jammy 2:8.2.0716-3ubuntu2 amd64
  Vi IMproved - enhanced vi editor - with GTK3 GUI
`
		mockRunner.AddCommand("apt", []string{"search", "vim"}, []byte(searchOutput), nil)

		// Mock dpkg-query for status detection
		dpkgOutput := `vim install ok installed 2:8.2.0716-3ubuntu2
vim-gtk3 deinstall ok config-files 2:8.2.0716-3ubuntu2
`
		mockRunner.AddCommand("dpkg-query", []string{"-W", "--showformat", "${binary:Package} ${Status} ${Version}\n", "vim", "vim-gtk3"}, []byte(dpkgOutput), nil)

		// Create APT package manager with mocked runner
		pm := apt.NewPackageManagerWithCustomRunner(mockRunner)

		// Execute Find operation
		packages, err := pm.Find([]string{"vim"}, &manager.Options{})
		if err != nil {
			t.Fatalf("Find failed: %v", err)
		}

		// Verify results
		if len(packages) != 2 {
			t.Fatalf("Expected 2 packages, got %d", len(packages))
		}

		// Check that vim package shows as installed
		found := false
		for _, pkg := range packages {
			if pkg.Name == "vim" && pkg.Status == "installed" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected vim package to be marked as installed")
		}

		// Check that vim-gtk3 shows as available (config-files normalized to available)
		found = false
		for _, pkg := range packages {
			if pkg.Name == "vim-gtk3" && pkg.Status == "available" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected vim-gtk3 package to be marked as available")
		}
	})

	t.Run("Install with mocked commands", func(t *testing.T) {
		// Create mock command runner
		mockRunner := manager.NewMockCommandRunner()

		// Mock successful install output
		installOutput := `Reading package lists...
Building dependency tree...
Reading state information...
The following additional packages will be installed:
  git-man liberror-perl
The following NEW packages will be installed:
  git git-man liberror-perl
0 upgraded, 3 newly installed, 0 to remove and 0 not upgraded.
After this operation, 25.9 MB of additional disk space will be used.
Selecting previously unselected package git.
(Reading database ... 195735 files and directories currently installed.)
Preparing to unpack .../git_1%3a2.34.1-1ubuntu1.6_amd64.deb ...
Unpacking git (1:2.34.1-1ubuntu1.6) ...
Setting up git (2.34.1-1ubuntu1.6) ...
Processing triggers for man-db (2.10.2-1) ...
`
		// The actual APT install command: install -f package -y
		mockRunner.AddCommand("apt", []string{"install", "-f", "git", "-y"}, []byte(installOutput), nil)

		// Create APT package manager with mocked runner
		pm := apt.NewPackageManagerWithCustomRunner(mockRunner)

		// Execute Install operation
		packages, err := pm.Install([]string{"git"}, &manager.Options{})
		if err != nil {
			t.Fatalf("Install failed: %v", err)
		}

		// Verify results - should contain newly installed packages
		if len(packages) < 1 {
			t.Fatalf("Expected at least 1 package in install results, got %d", len(packages))
		}

		// Verify environment variables were passed correctly
		env := mockRunner.GetEnvForCommand("apt", []string{"install", "-f", "git", "-y"})
		expectedEnv := []string{"DEBIAN_FRONTEND=noninteractive", "DEBCONF_NONINTERACTIVE_SEEN=true"}
		if len(env) != len(expectedEnv) {
			t.Errorf("Expected %d environment variables, got %d", len(expectedEnv), len(env))
		}
		for i, expected := range expectedEnv {
			if i >= len(env) || env[i] != expected {
				t.Errorf("Expected env[%d] = %q, got %q", i, expected, env[i])
			}
		}
	})

	t.Run("Install error handling", func(t *testing.T) {
		// Create mock command runner
		mockRunner := manager.NewMockCommandRunner()

		// Mock command failure with specific exit code
		mockRunner.AddCommand("apt", []string{"install", "-f", "nonexistent-package", "-y"}, nil, errors.New("E: Unable to locate package nonexistent-package"))

		// Create APT package manager with mocked runner
		pm := apt.NewPackageManagerWithCustomRunner(mockRunner)

		// Execute Install operation that should fail
		_, err := pm.Install([]string{"nonexistent-package"}, &manager.Options{})
		if err == nil {
			t.Fatal("Expected install to fail for nonexistent package")
		}

		// Verify error message contains expected information
		if !strings.Contains(err.Error(), "Unable to locate package") {
			t.Errorf("Expected error message about package not found, got: %v", err)
		}
	})

	t.Run("ListInstalled with mocked commands", func(t *testing.T) {
		// Create mock command runner
		mockRunner := manager.NewMockCommandRunner()

		// Mock dpkg-query output for installed packages
		installedOutput := `git 1:2.34.1-1ubuntu1.6
vim 2:8.2.0716-3ubuntu2
curl 7.81.0-1ubuntu1.4
`
		mockRunner.AddCommand("dpkg-query", []string{"-W", "-f", "${binary:Package} ${Version}\n"}, []byte(installedOutput), nil)

		// Create APT package manager with mocked runner
		pm := apt.NewPackageManagerWithCustomRunner(mockRunner)

		// Execute ListInstalled operation
		packages, err := pm.ListInstalled(&manager.Options{})
		if err != nil {
			t.Fatalf("ListInstalled failed: %v", err)
		}

		// Verify results
		if len(packages) != 3 {
			t.Fatalf("Expected 3 installed packages, got %d", len(packages))
		}

		// Check that all packages are marked as installed
		for _, pkg := range packages {
			if pkg.Status != "installed" {
				t.Errorf("Expected package %s to be marked as installed, got status: %s", pkg.Name, pkg.Status)
			}
		}
	})

	t.Run("Environment variable tracking", func(t *testing.T) {
		// Create mock command runner
		mockRunner := manager.NewMockCommandRunner()

		// Mock simple command
		mockRunner.AddCommand("apt", []string{"--version"}, []byte("apt 2.4.5 (amd64)\\n"), nil)

		// Create APT package manager with mocked runner
		pm := apt.NewPackageManagerWithCustomRunner(mockRunner)

		// Test IsAvailable (uses Run method)
		available := pm.IsAvailable()
		if !available {
			t.Error("Expected APT to be available with mocked commands")
		}

		// Verify that CommandRunner automatically added LC_ALL=C
		// Note: MockCommandRunner doesn't track this since Run() method doesn't accept extra env
		// But we can test methods that do use RunContext with env vars

		// Mock apt update command for Refresh
		mockRunner.AddCommand("apt", []string{"update"}, []byte("Hit:1 http://archive.ubuntu.com/ubuntu jammy InRelease\nReading package lists..."), nil)

		// Test Refresh (uses RunContext with environment variables)
		err := pm.Refresh(&manager.Options{})
		if err != nil {
			t.Fatalf("Refresh failed: %v", err)
		}

		// Verify environment variables were passed correctly
		env := mockRunner.GetEnvForCommand("apt", []string{"update"})
		expectedEnv := []string{"DEBIAN_FRONTEND=noninteractive", "DEBCONF_NONINTERACTIVE_SEEN=true"}
		if len(env) != len(expectedEnv) {
			t.Errorf("Expected %d environment variables, got %d", len(expectedEnv), len(env))
		}
	})

	t.Run("Delete with purge flag", func(t *testing.T) {
		// Create mock command runner
		mockRunner := manager.NewMockCommandRunner()

		removeOutputWithPurge := `Reading package lists... Done
Building dependency tree... Done
Reading state information... Done
The following packages will be REMOVED:
  ebtables*
0 upgraded, 0 newly installed, 1 to remove and 0 not upgraded.
After this operation, 123 kB disk space will be freed.
Do you want to continue? [Y/n] 
(Reading database ... 1234 files and directories currently installed.)
Removing ebtables (2.0.11-4ubuntu1) ...
Purging configuration files for ebtables (2.0.11-4ubuntu1) ...
`
		expectedArgs := []string{"remove", "-f", "--autoremove", "-y", "--purge", "ebtables"}
		mockRunner.AddCommand("apt", expectedArgs, []byte(removeOutputWithPurge), nil)

		// Create APT package manager with mocked runner
		pm := apt.NewPackageManagerWithCustomRunner(mockRunner)

		// Execute Delete operation with purge flag
		opts := &manager.Options{
			AssumeYes:         true,
			CustomCommandArgs: []string{apt.ArgsPurge},
		}
		packages, err := pm.Delete([]string{"ebtables"}, opts)
		if err != nil {
			t.Fatalf("Delete with purge flag failed: %v", err)
		}

		// Verify return value
		if packages == nil {
			t.Error("Delete should return package info, got nil")
		}
	})

	t.Run("Delete without purge flag", func(t *testing.T) {
		// Create mock command runner
		mockRunner := manager.NewMockCommandRunner()

		removeOutputWithoutPurge := `Reading package lists... Done
Building dependency tree... Done
Reading state information... Done
The following packages will be REMOVED:
  ebtables
0 upgraded, 0 newly installed, 1 to remove and 0 not upgraded.
After this operation, 123 kB disk space will be freed.
Do you want to continue? [Y/n] 
(Reading database ... 1234 files and directories currently installed.)
Removing ebtables (2.0.11-4ubuntu1) ...
`
		expectedArgs := []string{"remove", "-f", "--autoremove", "-y", "ebtables"}
		mockRunner.AddCommand("apt", expectedArgs, []byte(removeOutputWithoutPurge), nil)

		// Create APT package manager with mocked runner
		pm := apt.NewPackageManagerWithCustomRunner(mockRunner)

		// Execute Delete operation without purge flag
		opts := &manager.Options{
			AssumeYes: true,
		}
		packages, err := pm.Delete([]string{"ebtables"}, opts)
		if err != nil {
			t.Fatalf("Delete without purge flag failed: %v", err)
		}

		// Verify return value
		if packages == nil {
			t.Error("Delete should return package info, got nil")
		}
	})
}

// TestAPTCommandRunnerMigration verifies that the migration from CommandBuilder to CommandRunner works correctly
func TestAPTCommandRunnerMigration(t *testing.T) {
	t.Run("NewPackageManagerWithCustomRunner accepts custom runner", func(t *testing.T) {
		mockRunner := manager.NewMockCommandRunner()

		pm := apt.NewPackageManagerWithCustomRunner(mockRunner)

		if pm == nil {
			t.Fatal("Expected non-nil package manager")
		}

		// Verify that the runner is used by testing a simple operation
		mockRunner.AddCommand("apt", []string{"--version"}, []byte("apt 2.4.5"), nil)
		available := pm.IsAvailable()
		if !available {
			t.Error("Expected package manager to be available with mocked runner")
		}
	})

	t.Run("CommandRunner interface compliance", func(t *testing.T) {
		// Verify that APT package manager works with the CommandRunner interface
		mockRunner := manager.NewMockCommandRunner()

		// Test with real Ubuntu-like output to ensure compatibility
		realUbuntuSearchOutput := `Sorting...
Full Text Search...
nginx/jammy-updates,jammy-security 1.18.0-6ubuntu14.4 amd64
  small, powerful, scalable web/proxy server

nginx-common/jammy-updates,jammy-security 1.18.0-6ubuntu14.4 all
  small, powerful, scalable web/proxy server - common files

nginx-core/jammy-updates,jammy-security 1.18.0-6ubuntu14.4 amd64
  nginx web/proxy server (standard version)
`
		mockRunner.AddCommand("apt", []string{"search", "nginx"}, []byte(realUbuntuSearchOutput), nil)

		dpkgOutput := `nginx deinstall ok config-files 1.18.0-6ubuntu14.4
nginx-common install ok installed 1.18.0-6ubuntu14.4
nginx-core deinstall ok config-files 1.18.0-6ubuntu14.4
`
		mockRunner.AddCommand("dpkg-query", []string{"-W", "--showformat", "${binary:Package} ${Status} ${Version}\n", "nginx", "nginx-common", "nginx-core"}, []byte(dpkgOutput), nil)

		// Create APT package manager with mocked runner
		pm := apt.NewPackageManagerWithCustomRunner(mockRunner)

		// Execute search operation
		packages, err := pm.Find([]string{"nginx"}, &manager.Options{})
		if err != nil {
			t.Fatalf("Find operation failed: %v", err)
		}

		if len(packages) != 3 {
			t.Fatalf("Expected 3 packages in search results, got %d", len(packages))
		}

		// ✅ Status detection now works with mocked CommandRunner!
		// Verify that status normalization works (config-files → available)
		for _, pkg := range packages {
			if pkg.Name == "nginx-common" && pkg.Status != "installed" {
				t.Errorf("Expected nginx-common to be installed, got status: %s", pkg.Status)
			}
			if (pkg.Name == "nginx" || pkg.Name == "nginx-core") && pkg.Status != "available" {
				t.Errorf("Expected %s to be available (config-files normalized), got status: %s", pkg.Name, pkg.Status)
			}
		}
	})
}
