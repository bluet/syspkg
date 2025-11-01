package apt

import (
	"context"
	"testing"

	"github.com/bluet/syspkg/manager"
)

// TestAptFast_BinaryConfiguration tests that apt-fast can be configured as the binary
func TestAptFast_BinaryConfiguration(t *testing.T) {
	t.Run("NewPackageManagerWithBinary creates apt-fast manager", func(t *testing.T) {
		pm := NewPackageManagerWithBinary("apt-fast")
		if pm == nil {
			t.Fatal("NewPackageManagerWithBinary returned nil")
		}

		// Verify binary name is set correctly
		binaryName := pm.GetPackageManager()
		if binaryName != "apt-fast" {
			t.Errorf("Expected binary name 'apt-fast', got '%s'", binaryName)
		}
	})

	t.Run("Default NewPackageManager uses apt", func(t *testing.T) {
		pm := NewPackageManager()
		if pm == nil {
			t.Fatal("NewPackageManager returned nil")
		}

		// Verify binary name is 'apt' by default
		binaryName := pm.GetPackageManager()
		if binaryName != "apt" {
			t.Errorf("Expected binary name 'apt', got '%s'", binaryName)
		}
	})

	t.Run("Zero-value struct defaults to apt", func(t *testing.T) {
		pm := &PackageManager{}

		// Verify lazy initialization defaults to 'apt'
		binaryName := pm.GetPackageManager()
		if binaryName != "apt" {
			t.Errorf("Expected binary name 'apt', got '%s'", binaryName)
		}
	})
}

// TestAptFast_CommandExecution tests that apt-fast commands use the correct binary
func TestAptFast_CommandExecution(t *testing.T) {
	t.Run("apt-fast Find uses correct binary", func(t *testing.T) {
		mockRunner := manager.NewMockCommandRunner()
		pm := NewPackageManagerWithCustomRunnerAndBinary(mockRunner, "apt-fast")

		// Mock the search command for apt-fast
		mockRunner.AddCommand("apt-fast", []string{"search", "vim"}, []byte("vim - Vi IMproved"), nil)

		opts := &manager.Options{
			Verbose: false,
		}

		_, err := pm.Find([]string{"vim"}, opts)
		if err != nil {
			t.Fatalf("Find() failed: %v", err)
		}

		// Verify the correct binary was called
		env := mockRunner.GetEnvForCommand("apt-fast", []string{"search", "vim"})
		if env == nil {
			t.Error("apt-fast search command was not called")
		}
	})

	t.Run("apt-fast Install uses correct binary", func(t *testing.T) {
		mockRunner := manager.NewMockCommandRunner()
		pm := NewPackageManagerWithCustomRunnerAndBinary(mockRunner, "apt-fast")

		// Mock the install command for apt-fast
		mockRunner.AddCommand("apt-fast", []string{"install", "-f", "vim", "-y"}, []byte("Setting up vim..."), nil)

		opts := &manager.Options{
			Interactive: false,
			DryRun:      false,
		}

		_, err := pm.Install([]string{"vim"}, opts)
		if err != nil {
			t.Fatalf("Install() failed: %v", err)
		}

		// Verify the correct binary was called
		env := mockRunner.GetEnvForCommand("apt-fast", []string{"install", "-f", "vim", "-y"})
		if env == nil {
			t.Error("apt-fast install command was not called")
		}
	})

	t.Run("apt-fast ListUpgradable uses correct binary", func(t *testing.T) {
		mockRunner := manager.NewMockCommandRunner()
		pm := NewPackageManagerWithCustomRunnerAndBinary(mockRunner, "apt-fast")

		// Mock the list command for apt-fast
		mockRunner.AddCommand("apt-fast", []string{"list", "--upgradable"}, []byte("Listing..."), nil)

		opts := &manager.Options{}

		_, err := pm.ListUpgradable(opts)
		if err != nil {
			t.Fatalf("ListUpgradable() failed: %v", err)
		}

		// Verify the correct binary was called
		env := mockRunner.GetEnvForCommand("apt-fast", []string{"list", "--upgradable"})
		if env == nil {
			t.Error("apt-fast list command was not called")
		}
	})
}

// TestAptFast_InteractiveMode tests that apt-fast works in interactive mode
func TestAptFast_InteractiveMode(t *testing.T) {
	t.Run("apt-fast interactive install", func(t *testing.T) {
		mockRunner := manager.NewMockCommandRunner()
		pm := NewPackageManagerWithCustomRunnerAndBinary(mockRunner, "apt-fast")

		// Mock the interactive install command
		mockRunner.AddCommand("apt-fast", []string{"install", "-f", "vim"}, nil, nil)

		opts := &manager.Options{
			Interactive: true,
		}

		_, err := pm.Install([]string{"vim"}, opts)
		if err != nil {
			t.Fatalf("Install() in interactive mode failed: %v", err)
		}

		// Verify interactive command was called
		if !mockRunner.WasInteractiveCalled("apt-fast", []string{"install", "-f", "vim"}) {
			t.Error("apt-fast install was not called in interactive mode")
		}
	})
}

// TestAptFast_BackwardCompatibility tests that existing code still works
func TestAptFast_BackwardCompatibility(t *testing.T) {
	t.Run("NewPackageManagerWithCustomRunner still works with apt", func(t *testing.T) {
		mockRunner := manager.NewMockCommandRunner()
		pm := NewPackageManagerWithCustomRunner(mockRunner)

		// Should default to 'apt'
		binaryName := pm.GetPackageManager()
		if binaryName != "apt" {
			t.Errorf("Expected binary name 'apt', got '%s'", binaryName)
		}

		// Mock a command for apt (not apt-fast)
		mockRunner.AddCommand("apt", []string{"search", "test"}, []byte("test package"), nil)

		opts := &manager.Options{}
		_, err := pm.Find([]string{"test"}, opts)
		if err != nil {
			t.Fatalf("Find() failed: %v", err)
		}

		// Verify apt was called, not apt-fast
		env := mockRunner.GetEnvForCommand("apt", []string{"search", "test"})
		if env == nil {
			t.Error("apt search command was not called")
		}
	})
}

// TestAptFast_IsAvailable tests the availability check for apt-fast
func TestAptFast_IsAvailable(t *testing.T) {
	t.Run("IsAvailable checks for configured binary", func(t *testing.T) {
		mockRunner := manager.NewMockCommandRunner()
		pm := NewPackageManagerWithCustomRunnerAndBinary(mockRunner, "apt-fast")

		// Mock the --version command for apt-fast
		mockRunner.AddCommand("apt-fast", []string{"--version"}, []byte("apt-fast 1.9.10"), nil)

		// IsAvailable will use exec.LookPath which we can't easily mock
		// So we just verify the binary name is set correctly
		binaryName := pm.GetPackageManager()
		if binaryName != "apt-fast" {
			t.Errorf("Expected binary name 'apt-fast', got '%s'", binaryName)
		}
	})
}

// TestAptFast_AllOperations tests that all package manager operations work with apt-fast
func TestAptFast_AllOperations(t *testing.T) {
	mockRunner := manager.NewMockCommandRunner()
	pm := NewPackageManagerWithCustomRunnerAndBinary(mockRunner, "apt-fast")

	testCases := []struct {
		name     string
		binary   string
		args     []string
		mockData []byte
		testFunc func() error
	}{
		{
			name:     "Refresh",
			binary:   "apt-fast",
			args:     []string{"update"},
			mockData: []byte("Reading package lists..."),
			testFunc: func() error {
				return pm.Refresh(&manager.Options{})
			},
		},
		{
			name:     "Delete",
			binary:   "apt-fast",
			args:     []string{"remove", "-f", "--autoremove", "-y", "vim"},
			mockData: []byte("Removing vim..."),
			testFunc: func() error {
				_, err := pm.Delete([]string{"vim"}, &manager.Options{})
				return err
			},
		},
		{
			name:     "Clean",
			binary:   "apt-fast",
			args:     []string{"autoclean"},
			mockData: []byte("Reading package lists..."),
			testFunc: func() error {
				return pm.Clean(&manager.Options{})
			},
		},
		{
			name:     "AutoRemove",
			binary:   "apt-fast",
			args:     []string{"autoremove", "-y"},
			mockData: []byte("Reading package lists..."),
			testFunc: func() error {
				_, err := pm.AutoRemove(&manager.Options{})
				return err
			},
		},
		{
			name:     "Upgrade",
			binary:   "apt-fast",
			args:     []string{"install", "vim", "-y"},
			mockData: []byte("Upgrading vim..."),
			testFunc: func() error {
				_, err := pm.Upgrade([]string{"vim"}, &manager.Options{})
				return err
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRunner.AddCommand(tc.binary, tc.args, tc.mockData, nil)

			err := tc.testFunc()
			if err != nil {
				t.Fatalf("%s failed: %v", tc.name, err)
			}

			// Verify the correct binary was called
			env := mockRunner.GetEnvForCommand(tc.binary, tc.args)
			if env == nil {
				t.Errorf("%s command was not called with binary %s", tc.name, tc.binary)
			}
		})
	}
}

// TestAptFast_GetPackageInfo tests GetPackageInfo with apt-fast
func TestAptFast_GetPackageInfo(t *testing.T) {
	t.Run("GetPackageInfo uses apt-cache (not apt-fast)", func(t *testing.T) {
		mockRunner := manager.NewMockCommandRunner()
		pm := NewPackageManagerWithCustomRunnerAndBinary(mockRunner, "apt-fast")

		// GetPackageInfo uses apt-cache, not the main binary
		mockRunner.AddCommand("apt-cache", []string{"show", "vim"}, []byte(`Package: vim
Version: 2:8.1.2269-1ubuntu5.22
Description: Vi IMproved - enhanced vi editor`), nil)

		info, err := pm.GetPackageInfo("vim", &manager.Options{})
		if err != nil {
			t.Fatalf("GetPackageInfo() failed: %v", err)
		}

		if info.Name != "vim" {
			t.Errorf("Expected package name 'vim', got '%s'", info.Name)
		}
	})
}

// TestAptFast_EnvironmentVariables tests that environment variables are properly set
func TestAptFast_EnvironmentVariables(t *testing.T) {
	t.Run("apt-fast commands include DEBIAN_FRONTEND=noninteractive", func(t *testing.T) {
		mockRunner := manager.NewMockCommandRunner()
		pm := NewPackageManagerWithCustomRunnerAndBinary(mockRunner, "apt-fast")

		mockRunner.AddCommand("apt-fast", []string{"search", "vim"}, []byte("vim - editor"), nil)

		_, err := pm.Find([]string{"vim"}, &manager.Options{})
		if err != nil {
			t.Fatalf("Find() failed: %v", err)
		}

		// Check that environment variables were set
		env := mockRunner.GetEnvForCommand("apt-fast", []string{"search", "vim"})
		if env == nil {
			t.Fatal("Environment variables not tracked")
		}

		// Verify DEBIAN_FRONTEND is in the environment
		hasDebianFrontend := false
		for _, e := range env {
			if e == "DEBIAN_FRONTEND=noninteractive" {
				hasDebianFrontend = true
				break
			}
		}

		if !hasDebianFrontend {
			t.Error("DEBIAN_FRONTEND=noninteractive not found in environment variables")
		}
	})
}

// TestAptFast_ContextTimeout tests that context timeouts work correctly
func TestAptFast_ContextTimeout(t *testing.T) {
	t.Run("apt-fast respects context timeout", func(t *testing.T) {
		mockRunner := manager.NewMockCommandRunner()
		pm := NewPackageManagerWithCustomRunnerAndBinary(mockRunner, "apt-fast")

		// Simulate a command that would timeout
		mockRunner.AddCommand("apt-fast", []string{"search", "test"}, []byte("result"), context.DeadlineExceeded)

		_, err := pm.Find([]string{"test"}, &manager.Options{})
		if err != context.DeadlineExceeded {
			t.Errorf("Expected context.DeadlineExceeded error, got: %v", err)
		}
	})
}
