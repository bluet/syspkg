// yum/yum_test.go
package yum_test

import (
	"testing"

	"github.com/bluet/syspkg/manager"
	"github.com/bluet/syspkg/manager/yum"
)

func TestYumPackageManagerNotAvailable(t *testing.T) {
	yumManager := yum.PackageManager{}
	opts := manager.Options{}
	packages := []string{"nginx"}

	_, erri := yumManager.Install(packages, nil)
	if erri == nil {
		t.Fatal("YumPackageManager should not support installation")
	}
	_, errd := yumManager.Delete(packages, nil)
	if errd == nil {
		t.Fatal("YumPackageManager should not support removal")
	}
	_, errlu := yumManager.ListUpgradable(&opts)
	if errlu == nil {
		t.Fatal("YumPackageManager should not support list-upgradable")
	}
	_, erru := yumManager.Upgrade(packages, nil)
	if erru == nil {
		t.Fatal("YumPackageManager should not support upgrade")
	}
	_, errua := yumManager.UpgradeAll(&opts)
	if errua == nil {
		t.Fatal("YumPackageManager should not support upgrade-all")
	}
	_, errar := yumManager.AutoRemove(&opts)
	if errar == nil {
		t.Fatal("YumPackageManager should not support autoremove")
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

func TestParseFindOutput(t *testing.T) {
	msg := `
Last metadata expiration check: 4:56:00 ago on Thu 22 May 2025 04:30:18 PM UTC.
============================================================================= Name Exactly Matched: nginx ==============================================================================
nginx.x86_64 : A high performance web server and reverse proxy server
============================================================================ Name & Summary Matched: nginx =============================================================================
nginx-all-modules.noarch : A meta package that installs all available Nginx modules
nginx-core.x86_64 : nginx minimal core
nginx-filesystem.noarch : The basic directory layout for the Nginx server
nginx-mod-http-image-filter.x86_64 : Nginx HTTP image filter module
nginx-mod-http-perl.x86_64 : Nginx HTTP perl module
nginx-mod-http-xslt-filter.x86_64 : Nginx XSLT module
nginx-mod-mail.x86_64 : Nginx mail modules
nginx-mod-stream.x86_64 : Nginx stream modules
pcp-pmda-nginx.x86_64 : Performance Co-Pilot (PCP) metrics for the Nginx Webserver
`
	packages := yum.ParseFindOutput(msg, nil)
	if packages[0].Name != "nginx" || packages[0].Arch != "x86_64" {
		t.Errorf("Expected to find nginx, found %+v", packages[0])
	}
}

func TestParseListInstalledOutput(t *testing.T) {
	msg := `
	Installed Packages
NetworkManager.x86_64                                                                    1:1.48.10-2.el9_5                                                                    @baseos
rocky-release.noarch                                                                     9.5-1.2.el9                                                                          @baseos
rpm.x86_64                                                                               4.16.1.3-34.el9.0.1                                                                  @baseos
rsync.x86_64                                                                             3.2.3-20.el9                                                                         @baseos

`
	packages := yum.ParseListInstalledOutput(msg, nil)
	found := false
	for _, pack := range packages {
		if pack.Name == "rpm" || pack.Arch == "x86_64" || pack.Version == "4.16.1.3-34.el9.0.1" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected to find rpm, but not found. Found instead %+v", packages)
	}
}

func TestParsePackageInfoOutput(t *testing.T) {
	msg := `
	Last metadata expiration check: 5:16:10 ago on Thu 22 May 2025 04:30:18 PM UTC.
Available Packages
Name         : nginx
Epoch        : 2
Version      : 1.20.1
Release      : 20.el9.0.1
Architecture : x86_64
Size         : 36 k
Source       : nginx-1.20.1-20.el9.0.1.src.rpm
Repository   : appstream
Summary      : A high performance web server and reverse proxy server
URL          : https://nginx.org
License      : BSD
Description  : Nginx is a web server and a reverse proxy server for HTTP, SMTP, POP3 and
             : IMAP protocols, with a strong focus on high concurrency, performance and low
             : memory usage.



`
	packages := yum.ParsePackageInfoOutput(msg, nil)
	if packages.Name != "nginx" || packages.Arch != "x86_64" || packages.Version != "1.20.1" {
		t.Errorf("Expected to find nginx, found %+v", packages)
	}
}
