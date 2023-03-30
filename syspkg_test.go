// syspkg/syspkg_test.go
package syspkg_test

import (
	"testing"

	"github.com/bluet/syspkg"
)

func TestNewPackageManager(t *testing.T) {
	manager, err := syspkg.NewPackageManager()
	if err != nil {
		t.Fatalf("NewPackageManager() error: %v", err)
	}

	if manager == nil {
		t.Fatal("NewPackageManager() returned a nil manager")
	}
}
