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
		t.Fatalf("FindPackageManagers() error: %+v", err)
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
		pm, err := s.GetPackageManager("apt")

		if err != nil && pm == nil {
			pm, err := s.GetPackageManager("snap")

			if err != nil && pm == nil {
				pm, err := s.GetPackageManager("flatpak")

				if err != nil && pm == nil {
					t.Fatalf("apt, snap, or flatpak package manager not found")
				}
			}
		}
	} else if OSInfo.Distribution == "fedora" || OSInfo.Distribution == "centos" || OSInfo.Distribution == "rhel" || OSInfo.Distribution == "rockylinux" || OSInfo.Distribution == "almalinux" || OSInfo.Distribution == "amazon linux" || OSInfo.Distribution == "oracle linux" || OSInfo.Distribution == "scientific linux" || OSInfo.Distribution == "cloudlinux" {
		pm, err := s.GetPackageManager("dnf")
		if err != nil && pm == nil {
			pm, err := s.GetPackageManager("yum")
			if err != nil && pm == nil {
				t.Fatalf("dnf or yum package manager not found")
			}
		}
	} else if OSInfo.Distribution == "opensuse" {
		pm, err := s.GetPackageManager("zypper")
		if err != nil && pm == nil {
			t.Fatalf("zypper package manager not found")
		}
	} else if OSInfo.Distribution == "alpine" {
		pm, err := s.GetPackageManager("apk")
		if err != nil && pm == nil {
			t.Fatalf("apk package manager not found")
		}
	} else if OSInfo.Distribution == "arch" {
		pm, err := s.GetPackageManager("pacman")
		if err != nil && pm == nil {
			t.Fatalf("pacman package manager not found")
		}
	} else if OSInfo.Distribution == "gentoo" {
		pm, err := s.GetPackageManager("emerge")
		if err != nil && pm == nil {
			t.Fatalf("emerge package manager not found")
		}
	} else if OSInfo.Distribution == "slackware" {
		pm, err := s.GetPackageManager("slackpkg")
		if err != nil && pm == nil {
			t.Fatalf("slackpkg package manager not found")
		}
	} else if OSInfo.Distribution == "void" {
		pm, err := s.GetPackageManager("xbps")
		if err != nil && pm == nil {
			t.Fatalf("xbps package manager not found")
		}
	} else if OSInfo.Distribution == "solus" {
		pm, err := s.GetPackageManager("eopkg")
		if err != nil && pm == nil {
			t.Fatalf("eopkg package manager not found")
		}
	} else if OSInfo.Distribution == "freebsd" || OSInfo.Distribution == "dragonfly" || OSInfo.Distribution == "termux" {
		pm, err := s.GetPackageManager("pkg")
		if err != nil && pm == nil {
			t.Fatalf("pkg package manager not found")
		}
	} else if OSInfo.Distribution == "openbsd" || OSInfo.Distribution == "netbsd" {
		pm, err := s.GetPackageManager("pkg_add")
		if err != nil && pm == nil {
			t.Fatalf("pkg_add package manager not found")
		}
	} else if OSInfo.Distribution == "macos" {
		pm, err := s.GetPackageManager("brew")
		if err != nil && pm == nil {
			t.Fatalf("brew package manager not found")
		}
	} else if OSInfo.Distribution == "windows" {
		pm, err := s.GetPackageManager("chocolatey")
		if err != nil && pm == nil {
			pm, err := s.GetPackageManager("scoop")
			if err != nil && pm == nil {
				pm, err := s.GetPackageManager("winget")
				if err != nil && pm == nil {
					t.Fatalf("chocolatey, scoop, or winget package manager not found")
				}
			}
		}
	} else if OSInfo.Distribution == "android" {
		pm, err := s.GetPackageManager("f-droid")
		if err != nil && pm == nil {
			t.Fatalf("f-droid package manager not found")
		}
	} else if OSInfo.Distribution == "ios" {
		pm, err := s.GetPackageManager("cydia")
		if err != nil && pm == nil {
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
