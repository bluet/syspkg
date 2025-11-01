package syspkg

import (
	"testing"
)

func TestGetPackageManagerWithOptions(t *testing.T) {
	// Create a SysPkg instance with APT
	includeOptions := IncludeOptions{
		Apt: true,
	}

	sysPkg, err := New(includeOptions)
	if err != nil {
		t.Skipf("Skipping test: %v", err)
	}

	t.Run("GetPackageManager returns default apt", func(t *testing.T) {
		pm, err := sysPkg.GetPackageManager("apt")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if pm.GetPackageManager() != "apt" {
			t.Errorf("Expected package manager name 'apt', got '%s'", pm.GetPackageManager())
		}
	})

	t.Run("GetPackageManagerWithOptions with nil opts returns default apt", func(t *testing.T) {
		pm, err := sysPkg.GetPackageManagerWithOptions("apt", nil)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if pm.GetPackageManager() != "apt" {
			t.Errorf("Expected package manager name 'apt', got '%s'", pm.GetPackageManager())
		}
	})

	t.Run("GetPackageManagerWithOptions with empty BinaryPath returns default apt", func(t *testing.T) {
		pm, err := sysPkg.GetPackageManagerWithOptions("apt", &ManagerCreationOptions{
			BinaryPath: "",
		})
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if pm.GetPackageManager() != "apt" {
			t.Errorf("Expected package manager name 'apt', got '%s'", pm.GetPackageManager())
		}
	})

	t.Run("GetPackageManagerWithOptions with custom binary name", func(t *testing.T) {
		pm, err := sysPkg.GetPackageManagerWithOptions("apt", &ManagerCreationOptions{
			BinaryPath: "apt-fast",
		})
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if pm.GetPackageManager() != "apt-fast" {
			t.Errorf("Expected package manager name 'apt-fast', got '%s'", pm.GetPackageManager())
		}
	})

	t.Run("GetPackageManagerWithOptions with custom binary path", func(t *testing.T) {
		pm, err := sysPkg.GetPackageManagerWithOptions("apt", &ManagerCreationOptions{
			BinaryPath: "/usr/local/bin/apt-custom",
		})
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if pm.GetPackageManager() != "/usr/local/bin/apt-custom" {
			t.Errorf("Expected package manager name '/usr/local/bin/apt-custom', got '%s'", pm.GetPackageManager())
		}
	})

	t.Run("GetPackageManagerWithOptions for unsupported PM returns error", func(t *testing.T) {
		_, err := sysPkg.GetPackageManagerWithOptions("nonexistent", &ManagerCreationOptions{
			BinaryPath: "some-binary",
		})
		if err == nil {
			t.Error("Expected error for nonexistent package manager, got nil")
		}
	})

	t.Run("GetPackageManagerWithOptions for yum returns not yet supported error", func(t *testing.T) {
		_, err := sysPkg.GetPackageManagerWithOptions("yum", &ManagerCreationOptions{
			BinaryPath: "custom-yum",
		})
		if err == nil {
			t.Error("Expected error for yum custom binary (not yet supported), got nil")
		}
	})
}

func TestGetPackageManagerCompatibility(t *testing.T) {
	// Test that GetPackageManager is a wrapper around GetPackageManagerWithOptions
	includeOptions := IncludeOptions{
		Apt: true,
	}

	sysPkg, err := New(includeOptions)
	if err != nil {
		t.Skipf("Skipping test: %v", err)
	}

	t.Run("GetPackageManager behaves same as GetPackageManagerWithOptions(nil)", func(t *testing.T) {
		pm1, err1 := sysPkg.GetPackageManager("apt")
		pm2, err2 := sysPkg.GetPackageManagerWithOptions("apt", nil)

		if (err1 == nil) != (err2 == nil) {
			t.Errorf("Error mismatch: GetPackageManager err=%v, GetPackageManagerWithOptions err=%v", err1, err2)
		}

		if err1 == nil && err2 == nil {
			if pm1.GetPackageManager() != pm2.GetPackageManager() {
				t.Errorf("Package manager name mismatch: %s != %s", pm1.GetPackageManager(), pm2.GetPackageManager())
			}
		}
	})
}

func TestManagerCreationOptions(t *testing.T) {
	t.Run("ManagerCreationOptions struct has BinaryPath field", func(t *testing.T) {
		opts := &ManagerCreationOptions{
			BinaryPath: "test-binary",
		}

		if opts.BinaryPath != "test-binary" {
			t.Errorf("Expected BinaryPath 'test-binary', got '%s'", opts.BinaryPath)
		}
	})

	t.Run("ManagerCreationOptions zero value", func(t *testing.T) {
		opts := &ManagerCreationOptions{}

		if opts.BinaryPath != "" {
			t.Errorf("Expected empty BinaryPath, got '%s'", opts.BinaryPath)
		}
	})
}
