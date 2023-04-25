// syspkg/syspkg_test.go
package syspkg_test

import (
	"log"
	"testing"

	"github.com/bluet/syspkg"
	"github.com/bluet/syspkg/osinfo"
)

func TestNewPackageManager(t *testing.T) {

	// get system type
	OSInfo, err := osinfo.GetOSInfo()
	if err != nil {
		t.Fatalf("GetOSInfo() error: %+v", err)
	}

	log.Printf("OSInfo: %+v", OSInfo)

	s, err := syspkg.New(syspkg.IncludeOptions{
		AllAvailable: true,
	})
	if err != nil {
		t.Fatalf("NewPackageManager() error: %+v", err)
	}

	pms, err := s.FindPackageManagers(syspkg.IncludeOptions{
		AllAvailable: true,
	})
	if err != nil {
		t.Fatalf("GetPackageManagers() error: %+v", err)
	}

	log.Printf("pms: %+v", pms)

	// if we are on ubuntu, debian, mint, PopOS, elementary, Zorin, ChromeOS or any other debian-based distro, we should have apt, snap, or flatpak
	// if we are on fedora, centos, rhel, rockylinux, almalinux, amazon linux, oracle linux, scientific linux, or cloudlinux, we should have dnf or yum
	// if we are on opensuse, we should have zypper
	// if we are on alpine, we should have apk
	// if we are on arch, we should have pacman
	// if we are on gentoo, we should have emerge
	// if we are on slackware, we should have slackpkg
	// if we are on void, we should have xbps
	// if we are on solus, we should have eopkg
	// if we are on freebsd, dragonfly, or termux, we should have pkg
	// if we are on openbsd or netbsd, we should have pkg_add
	// if we are on macos, we should have brew
	// if we are on windows, we should have chocolatey or scoop or winget
	// if we are on android, we should have f-droid
	// if we are on ios, we should have cydia
	// if we are on any other distro, we should have nothing

	if OSInfo.Distribution == "ubuntu" || OSInfo.Distribution == "debian" || OSInfo.Distribution == "mint" || OSInfo.Distribution == "PopOS" || OSInfo.Distribution == "elementary" || OSInfo.Distribution == "Zorin" || OSInfo.Distribution == "ChromeOS" {
		if _, ok := pms["apt"]; !ok && s.GetPackageManager("apt") == nil {
			if _, ok := pms["snap"]; !ok && s.GetPackageManager("snap") == nil {
				if _, ok := pms["flatpak"]; !ok && s.GetPackageManager("flatpak") == nil {
					t.Fatalf("apt, snap, or flatpak package manager not found")
				}
			}
		}
	} else if OSInfo.Distribution == "fedora" || OSInfo.Distribution == "centos" || OSInfo.Distribution == "rhel" || OSInfo.Distribution == "rockylinux" || OSInfo.Distribution == "almalinux" || OSInfo.Distribution == "amazon linux" || OSInfo.Distribution == "oracle linux" || OSInfo.Distribution == "scientific linux" || OSInfo.Distribution == "cloudlinux" {
		if _, ok := pms["dnf"]; !ok && s.GetPackageManager("dnf") == nil {
			if _, ok := pms["yum"]; !ok && s.GetPackageManager("yum") == nil {
				t.Fatalf("dnf or yum package manager not found")
			}
		}
	} else if OSInfo.Distribution == "opensuse" {
		if _, ok := pms["zypper"]; !ok && s.GetPackageManager("zypper") == nil {
			t.Fatalf("zypper package manager not found")
		}
	} else if OSInfo.Distribution == "alpine" {
		if _, ok := pms["apk"]; !ok && s.GetPackageManager("apk") == nil {
			t.Fatalf("apk package manager not found")
		}
	} else if OSInfo.Distribution == "arch" {
		if _, ok := pms["pacman"]; !ok && s.GetPackageManager("pacman") == nil {
			t.Fatalf("pacman package manager not found")
		}
	} else if OSInfo.Distribution == "gentoo" {
		if _, ok := pms["emerge"]; !ok && s.GetPackageManager("emerge") == nil {
			t.Fatalf("emerge package manager not found")
		}
	} else if OSInfo.Distribution == "slackware" {
		if _, ok := pms["slackpkg"]; !ok && s.GetPackageManager("slackpkg") == nil {
			t.Fatalf("slackpkg package manager not found")
		}
	} else if OSInfo.Distribution == "void" {
		if _, ok := pms["xbps"]; !ok && s.GetPackageManager("xbps") == nil {
			t.Fatalf("xbps package manager not found")
		}
	} else if OSInfo.Distribution == "solus" {
		if _, ok := pms["eopkg"]; !ok && s.GetPackageManager("eopkg") == nil {
			t.Fatalf("eopkg package manager not found")
		}
	} else if OSInfo.Distribution == "freebsd" || OSInfo.Distribution == "dragonfly" || OSInfo.Distribution == "termux" {
		if _, ok := pms["pkg"]; !ok && s.GetPackageManager("pkg") == nil {
			t.Fatalf("pkg package manager not found")
		}
	} else if OSInfo.Distribution == "openbsd" || OSInfo.Distribution == "netbsd" {
		if _, ok := pms["pkg_add"]; !ok && s.GetPackageManager("pkg_add") == nil {
			t.Fatalf("pkg_add package manager not found")
		}
	} else if OSInfo.Distribution == "macos" {
		if _, ok := pms["brew"]; !ok && s.GetPackageManager("brew") == nil {
			t.Fatalf("brew package manager not found")
		}
	} else if OSInfo.Distribution == "windows" {
		if _, ok := pms["chocolatey"]; !ok && s.GetPackageManager("chocolatey") == nil {
			if _, ok := pms["scoop"]; !ok && s.GetPackageManager("scoop") == nil {
				if _, ok := pms["winget"]; !ok && s.GetPackageManager("winget") == nil {
					t.Fatalf("chocolatey, scoop, or winget package manager not found")
				}
			}
		}
	} else if OSInfo.Distribution == "android" {
		if _, ok := pms["f-droid"]; !ok && s.GetPackageManager("f-droid") == nil {
			t.Fatalf("f-droid package manager not found")
		}
	} else if OSInfo.Distribution == "ios" {
		if _, ok := pms["cydia"]; !ok && s.GetPackageManager("cydia") == nil {
			t.Fatalf("cydia package manager not found")
		}
	} else {
		if len(pms) > 0 {
			t.Fatalf("package manager found when none should be")
		} else {
			log.Printf("no package manager found, as expected")
		}
	}


	// if manager == nil {
	// 	t.Fatal("NewPackageManager() returned a nil manager")
	// }
}
