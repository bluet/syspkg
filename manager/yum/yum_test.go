// yum/yum_test.go
package yum_test

import (
	"testing"

	"github.com/bluet/syspkg/manager"
	"github.com/bluet/syspkg/manager/yum"
)

func TestYumPackageManagerNotAvailable(t *testing.T) {
	yumManager := yum.PackageManager{}

	// Skip this test if YUM is actually available (like in Rocky Linux CI environment)
	if yumManager.IsAvailable() {
		t.Skip("Skipping 'not available' test because YUM is available in this environment")
	}

	// This test validates behavior when YUM is not available on the system
	opts := manager.Options{}
	packages := []string{"nginx"}

	_, erri := yumManager.Install(packages, nil)
	if erri == nil {
		t.Fatal("YumPackageManager should not support installation when not available")
	}
	_, errd := yumManager.Delete(packages, nil)
	if errd == nil {
		t.Fatal("YumPackageManager should not support removal when not available")
	}
	_, errlu := yumManager.ListUpgradable(&opts)
	if errlu == nil {
		t.Fatal("YumPackageManager should not support list-upgradable when not available")
	}
	_, erru := yumManager.Upgrade(packages, nil)
	if erru == nil {
		t.Fatal("YumPackageManager should not support upgrade when not available")
	}
	_, errua := yumManager.UpgradeAll(&opts)
	if errua == nil {
		t.Fatal("YumPackageManager should not support upgrade-all when not available")
	}
	_, errar := yumManager.AutoRemove(&opts)
	if errar == nil {
		t.Fatal("YumPackageManager should not support autoremove when not available")
	}
}

/*
// these e2e tests work only under a RHEL derived Linux distro that yum installed

func TestYumPackageManagerIsAvailable(t *testing.T) {
	yumManager := yum.PackageManager{}
	if !yumManager.IsAvailable() {
		t.Fatal("YumPackageManager is not available")
	}
}

func TestYumPackageManagerListPackages(t *testing.T) {
	yumManager := yum.PackageManager{}
	opts:=manager.Options{}
	result,err:=yumManager.ListInstalled(&opts)
	if err!=nil{
		t.Errorf("Should have been able to list correctly, %s",err)
	}
	if len(result)==0{
		t.Fatal("Zero packages detected, there should have been at least one")
	}
}

func TestYumPackageManagerGetPackageInfo(t *testing.T) {
	yumManager := yum.PackageManager{}
	opts:=manager.Options{}
	packages:="rpm"

	result,err:=yumManager.GetPackageInfo(packages, &opts)
	if err!=nil{
		t.Errorf("Should have been able to get info correctly, %s",err)
	}
	if result.Name != packages {

		t.Errorf("rpm should be present, found %+v", result)
	}
}

func TestYumPackageManagerFind(t *testing.T) {
	yumManager := yum.PackageManager{}
	opts:=manager.Options{}
	packages:=[]string{"nginx"}

	result,err:=yumManager.Find(packages, &opts)
	if err!=nil{
		t.Errorf("Should have been able to search correctly, %s",err)
	}
	if len(result)==0 {
		t.Errorf("nginx should be present, found %+v", result)
	}
	if result[0].Name != packages[0] {
		t.Errorf("nginx should be available, found %+v", result)
	}
}
*/

// TestParseFindOutput_DeprecatedInlineData is kept for backwards compatibility
// New tests should use behavior_test.go with fixtures instead
func TestParseFindOutput_DeprecatedInlineData(t *testing.T) {
	t.Skip("Deprecated: inline test data replaced with fixture-based tests in behavior_test.go")
}

// TestParseListInstalledOutput_DeprecatedInlineData is kept for backwards compatibility
// New tests should use behavior_test.go with fixtures instead
func TestParseListInstalledOutput_DeprecatedInlineData(t *testing.T) {
	t.Skip("Deprecated: inline test data replaced with fixture-based tests in behavior_test.go")
}

// TestParsePackageInfoOutput_DeprecatedInlineData is kept for backwards compatibility
// New tests should use behavior_test.go with fixtures instead
func TestParsePackageInfoOutput_DeprecatedInlineData(t *testing.T) {
	t.Skip("Deprecated: inline test data replaced with fixture-based tests in behavior_test.go")
}
