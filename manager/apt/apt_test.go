// apt/apt_test.go
package apt_test

import (
	"testing"

	"github.com/bluet/syspkg/manager/apt"
)

func TestAptPackageManager(t *testing.T) {
	// Implement test cases for AptPackageManager
	aptManager := &apt.PackageManager{}
	if aptManager.IsAvailable() {
		t.Log("AptPackageManager is available")
	} else {
		t.Fatal("AptPackageManager is not available")
	}
}
