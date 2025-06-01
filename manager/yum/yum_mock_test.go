package yum_test

import (
	"errors"
	"testing"

	"github.com/bluet/syspkg/manager"
	"github.com/bluet/syspkg/manager/yum"
)

// TestYUM_WithMockedCommands demonstrates testing YUM operations with mocked commands
func TestYUM_WithMockedCommands(t *testing.T) {
	t.Run("Find with mocked commands", func(t *testing.T) {
		// Create mock command runner
		mock := manager.NewMockCommandRunner()

		// Mock yum search output
		searchOutput := `========================= Name & Summary Matched: vim ==========================
vim-enhanced.x86_64 : A version of the VIM editor which includes recent enhancements
vim-minimal.x86_64 : A minimal version of the VIM editor
`
		mock.AddCommand("yum", []string{"search", "vim"}, []byte(searchOutput), nil)

		// Mock rpm --version (for status check)
		mock.AddCommand("rpm", []string{"--version"}, []byte("RPM version 4.14.3\n"), nil)

		// Mock rpm -q for status detection
		mock.AddCommand("rpm", []string{"-q", "vim-enhanced"},
			[]byte("vim-enhanced-8.0.1763-19.el8_6.4.x86_64\n"), nil)
		mock.AddCommand("rpm", []string{"-q", "vim-minimal"}, nil,
			errors.New("package vim-minimal is not installed"))

		// Create YUM package manager with mocked runner
		pm := yum.NewPackageManagerWithCustomRunner(mock)

		// Execute Find operation
		packages, err := pm.Find([]string{"vim"}, &manager.Options{})
		if err != nil {
			t.Fatalf("Find failed: %v", err)
		}

		// Verify results
		if len(packages) != 2 {
			t.Fatalf("Expected 2 packages, got %d", len(packages))
		}

		// Check status detection worked correctly
		for _, pkg := range packages {
			switch pkg.Name {
			case "vim-enhanced":
				if pkg.Status != manager.PackageStatusInstalled {
					t.Errorf("vim-enhanced should be installed, got %s", pkg.Status)
				}
				if pkg.Version == "" {
					t.Error("vim-enhanced should have version")
				}
			case "vim-minimal":
				if pkg.Status != manager.PackageStatusAvailable {
					t.Errorf("vim-minimal should be available, got %s", pkg.Status)
				}
				if pkg.Version != "" {
					t.Error("vim-minimal should not have version")
				}
			default:
				t.Errorf("Unexpected package: %s", pkg.Name)
			}
		}
	})

	t.Run("Install with mocked commands", func(t *testing.T) {
		// Create mock command runner
		mock := manager.NewMockCommandRunner()

		// Mock yum install output
		installOutput := `Installing:
 vim-enhanced       x86_64       2:8.0.1763-19.el8_6.4       appstream       1.4 M

Transaction Summary
================================================================================
Install  1 Package

Installed:
  vim-enhanced-2:8.0.1763-19.el8_6.4.x86_64

Complete!
`
		mock.AddCommand("yum", []string{"install", "-y", "vim-enhanced"},
			[]byte(installOutput), nil)

		// Create YUM package manager with mocked runner
		pm := yum.NewPackageManagerWithCustomRunner(mock)

		// Execute Install operation
		packages, err := pm.Install([]string{"vim-enhanced"}, &manager.Options{})
		if err != nil {
			t.Fatalf("Install failed: %v", err)
		}

		// Verify results
		if len(packages) != 1 {
			t.Fatalf("Expected 1 package installed, got %d", len(packages))
		}

		pkg := packages[0]
		if pkg.Name != "vim-enhanced" {
			t.Errorf("Expected vim-enhanced, got %s", pkg.Name)
		}
		if pkg.Status != manager.PackageStatusInstalled {
			t.Errorf("Expected installed status, got %s", pkg.Status)
		}
		if pkg.Version == "" {
			t.Error("Installed package should have version")
		}
	})

	t.Run("Error handling with mocked commands", func(t *testing.T) {
		// Create mock command runner
		mock := manager.NewMockCommandRunner()

		// Mock command failure
		mock.AddCommand("yum", []string{"install", "-y", "nonexistent-package"}, nil,
			errors.New("No package nonexistent-package available"))

		// Create YUM package manager with mocked runner
		pm := yum.NewPackageManagerWithCustomRunner(mock)

		// Execute Install operation that should fail
		_, err := pm.Install([]string{"nonexistent-package"}, &manager.Options{})
		if err == nil {
			t.Error("Expected error for nonexistent package")
		}
	})
}
