package yum

import (
	"errors"
	"strings"
	"testing"

	"github.com/bluet/syspkg/manager"
)

func TestCheckRpmInstallationStatus(t *testing.T) {
	tests := []struct {
		name             string
		packageNames     []string
		mockedCommands   map[string][]byte
		mockedErrors     map[string]error
		expectedPackages map[string]manager.PackageInfo
		expectedError    bool
	}{
		{
			name:         "single installed package",
			packageNames: []string{"vim-enhanced"},
			mockedCommands: map[string][]byte{
				"rpm --version":       []byte("RPM version 4.14.3\n"),
				"rpm -q vim-enhanced": []byte("vim-enhanced-8.0.1763-19.el8_6.4.x86_64\n"),
			},
			expectedPackages: map[string]manager.PackageInfo{
				"vim-enhanced": {
					Name:           "vim-enhanced",
					Version:        "8.0.1763-19.el8_6.4",
					Status:         manager.PackageStatusInstalled,
					PackageManager: "yum",
				},
			},
			expectedError: false,
		},
		{
			name:         "package not installed",
			packageNames: []string{"nonexistent"},
			mockedCommands: map[string][]byte{
				"rpm --version": []byte("RPM version 4.14.3\n"),
			},
			mockedErrors: map[string]error{
				"rpm -q nonexistent": errors.New("package nonexistent is not installed"),
			},
			expectedPackages: map[string]manager.PackageInfo{},
			expectedError:    false,
		},
		{
			name:         "multiple packages mixed status",
			packageNames: []string{"vim-enhanced", "nonexistent", "bash"},
			mockedCommands: map[string][]byte{
				"rpm --version":       []byte("RPM version 4.14.3\n"),
				"rpm -q vim-enhanced": []byte("vim-enhanced-8.0.1763-19.el8_6.4.x86_64\n"),
				"rpm -q bash":         []byte("bash-4.4.20-1.el8.x86_64\n"),
			},
			mockedErrors: map[string]error{
				"rpm -q nonexistent": errors.New("package nonexistent is not installed"),
			},
			expectedPackages: map[string]manager.PackageInfo{
				"vim-enhanced": {
					Name:           "vim-enhanced",
					Version:        "8.0.1763-19.el8_6.4",
					Status:         manager.PackageStatusInstalled,
					PackageManager: "yum",
				},
				"bash": {
					Name:           "bash",
					Version:        "4.4.20-1.el8",
					Status:         manager.PackageStatusInstalled,
					PackageManager: "yum",
				},
			},
			expectedError: false,
		},
		{
			name:         "rpm command not available",
			packageNames: []string{"vim-enhanced"},
			mockedErrors: map[string]error{
				"rpm --version": errors.New("rpm: command not found"),
			},
			expectedPackages: nil,
			expectedError:    true,
		},
		{
			name:             "empty package list",
			packageNames:     []string{},
			mockedCommands:   map[string][]byte{},
			expectedPackages: map[string]manager.PackageInfo{},
			expectedError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock command runner
			runner := manager.NewMockCommandRunner()

			// Set up mocked commands and errors using the proper methods
			for cmd, output := range tt.mockedCommands {
				// Parse the command string to extract name and args
				parts := strings.Fields(cmd)
				if len(parts) > 0 {
					name := parts[0]
					args := parts[1:]
					runner.AddCommand(name, args, output, nil)
				}
			}
			for cmd, err := range tt.mockedErrors {
				// Parse the command string to extract name and args
				parts := strings.Fields(cmd)
				if len(parts) > 0 {
					name := parts[0]
					args := parts[1:]
					runner.AddCommand(name, args, nil, err)
				}
			}

			// Test the method using PackageManager
			pm := NewPackageManagerWithCustomRunner(runner)
			result, err := pm.checkRpmInstallationStatus(tt.packageNames)

			// Check error expectation
			if tt.expectedError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check results if no error expected
			if !tt.expectedError {
				if len(result) != len(tt.expectedPackages) {
					t.Errorf("Expected %d packages, got %d", len(tt.expectedPackages), len(result))
				}

				for name, expectedPkg := range tt.expectedPackages {
					actualPkg, exists := result[name]
					if !exists {
						t.Errorf("Expected package %s not found in result", name)
						continue
					}

					if actualPkg.Name != expectedPkg.Name {
						t.Errorf("Package %s: expected name %s, got %s", name, expectedPkg.Name, actualPkg.Name)
					}
					if actualPkg.Version != expectedPkg.Version {
						t.Errorf("Package %s: expected version %s, got %s", name, expectedPkg.Version, actualPkg.Version)
					}
					if actualPkg.Status != expectedPkg.Status {
						t.Errorf("Package %s: expected status %s, got %s", name, expectedPkg.Status, actualPkg.Status)
					}
					if actualPkg.PackageManager != expectedPkg.PackageManager {
						t.Errorf("Package %s: expected package manager %s, got %s", name, expectedPkg.PackageManager, actualPkg.PackageManager)
					}
				}
			}
		})
	}
}

func TestParseFindOutput_PureFunction(t *testing.T) {
	// Test that ParseFindOutput is now a pure function (no system calls)
	searchOutput := `Last metadata expiration check: 0:26:09 ago on Thu 22 May 2025 04:30:18 PM UTC.
==================================================Name Exactly Matched: vim ====================================================
vim-enhanced.x86_64 : A highly configurable, improved version of the vi text editor
vim-common.x86_64 : Common files for vim
====================================================Name & Summary Matched: vim==================================================
vim-filesystem.noarch : VIM filesystem layout
vim-minimal.x86_64 : A minimal version of the VIM editor`

	packages := ParseFindOutput(searchOutput, nil)

	expectedPackages := []string{"vim-enhanced", "vim-common", "vim-filesystem", "vim-minimal"}

	if len(packages) != len(expectedPackages) {
		t.Errorf("Expected %d packages, got %d", len(expectedPackages), len(packages))
	}

	// All packages should have Status=Available by default (pure parsing)
	for i, pkg := range packages {
		if pkg.Status != manager.PackageStatusAvailable {
			t.Errorf("Package %d (%s): expected status %s, got %s", i, pkg.Name, manager.PackageStatusAvailable, pkg.Status)
		}
		if pkg.Version != "" {
			t.Errorf("Package %d (%s): expected empty version, got %s", i, pkg.Name, pkg.Version)
		}
		if pkg.PackageManager != "yum" {
			t.Errorf("Package %d (%s): expected package manager 'yum', got %s", i, pkg.Name, pkg.PackageManager)
		}
	}

	// Verify specific packages are found
	packageNames := make(map[string]bool)
	for _, pkg := range packages {
		packageNames[pkg.Name] = true
	}

	for _, expectedName := range expectedPackages {
		if !packageNames[expectedName] {
			t.Errorf("Expected package %s not found", expectedName)
		}
	}
}

func TestExtractVersionFromRpmOutput(t *testing.T) {
	tests := []struct {
		name        string
		rpmOutput   string
		packageName string
		expected    string
	}{
		{
			name:        "standard package with epoch",
			rpmOutput:   "vim-enhanced-2:8.0.1763-19.el8_6.4.x86_64",
			packageName: "vim-enhanced",
			expected:    "2:8.0.1763-19.el8_6.4",
		},
		{
			name:        "package without epoch",
			rpmOutput:   "bash-4.4.20-1.el8.x86_64",
			packageName: "bash",
			expected:    "4.4.20-1.el8",
		},
		{
			name:        "package with complex name",
			rpmOutput:   "python3-pip-9.0.3-22.el8.noarch",
			packageName: "python3-pip",
			expected:    "9.0.3-22.el8",
		},
		{
			name:        "malformed output fallback",
			rpmOutput:   "malformed-output",
			packageName: "package",
			expected:    "malformed-output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractVersionFromRpmOutput(tt.rpmOutput, tt.packageName)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
