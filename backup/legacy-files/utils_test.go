package apt_test

import (
	"testing"

	"github.com/bluet/syspkg/manager/apt"
)

// TestPackageManager_IsAvailable tests the basic availability check behavior
func TestPackageManager_IsAvailable(t *testing.T) {
	pm := &apt.PackageManager{}

	// Test behavior: IsAvailable should return a boolean
	available := pm.IsAvailable()

	// We don't assert the specific value since it depends on the system
	// We just test that the method doesn't panic and returns a boolean
	_ = available
}

// TestPackageManager_GetPackageManager tests the identifier behavior
func TestPackageManager_GetPackageManager(t *testing.T) {
	pm := &apt.PackageManager{}

	// Test contract: Should always return "apt"
	if pm.GetPackageManager() != "apt" {
		t.Errorf("GetPackageManager() should return 'apt', got '%s'", pm.GetPackageManager())
	}
}
