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

// TestGetPackageManagerEmpty tests the GetPackageManager function with an empty string
func TestGetPackageManagerEmpty(t *testing.T) {
	t.Run("with available package managers", func(t *testing.T) {
		s, err := syspkg.New(syspkg.IncludeOptions{
			AllAvailable: true,
		})
		if err != nil {
			t.Fatalf("New() error: %+v", err)
		}

		// Get first available package manager with empty string
		pm, err := s.GetPackageManager("")
		if err != nil {
			// This might happen if no package managers are available on the system
			t.Logf("GetPackageManager(\"\") returned error (expected if no PMs available): %v", err)
			return
		}

		if pm == nil {
			t.Fatal("GetPackageManager(\"\") returned nil without error")
		}

		// Verify it returns a valid package manager
		if !pm.IsAvailable() {
			t.Fatal("GetPackageManager(\"\") returned an unavailable package manager")
		}

		t.Logf("GetPackageManager(\"\") successfully returned first available package manager")
	})

	t.Run("with no package managers", func(t *testing.T) {
		// Try to create instance with no package managers
		s, err := syspkg.New(syspkg.IncludeOptions{
			AllAvailable: false,
			// Don't specify any specific package managers
		})
		
		if err != nil {
			// This is expected - no package managers specified
			t.Logf("New() correctly returned error when no PMs specified: %v", err)
			return
		}

		// If New() didn't error, try GetPackageManager("")
		_, err = s.GetPackageManager("")
		if err == nil {
			t.Fatal("GetPackageManager(\"\") should return error when no PMs available")
		}
		
		t.Logf("GetPackageManager(\"\") correctly returned error: %v", err)
	})

	t.Run("error cases", func(t *testing.T) {
		s, err := syspkg.New(syspkg.IncludeOptions{
			AllAvailable: true,
		})
		if err != nil {
			t.Skipf("Skipping error cases test - no package managers available: %v", err)
			return
		}

		// Test non-existent package manager
		_, err = s.GetPackageManager("nonexistent-pm")
		if err == nil {
			t.Fatal("GetPackageManager(\"nonexistent-pm\") should return error")
		}
		t.Logf("GetPackageManager(\"nonexistent-pm\") correctly returned error: %v", err)
	})
}

// TestGetPackageManagerPanicRegression tests that GetPackageManager doesn't panic
// This is a regression test for the issue fixed in PR #12
func TestGetPackageManagerPanicRegression(t *testing.T) {
	// This test ensures GetPackageManager("") doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("GetPackageManager(\"\") panicked: %v", r)
		}
	}()

	s, err := syspkg.New(syspkg.IncludeOptions{
		AllAvailable: true,
	})
	if err != nil {
		t.Skipf("No package managers available: %v", err)
		return
	}

	// This used to panic before PR #12
	pm, err := s.GetPackageManager("")
	
	// We don't care about the result, just that it doesn't panic
	if err != nil {
		t.Logf("GetPackageManager(\"\") returned error (ok): %v", err)
	} else if pm != nil {
		t.Logf("GetPackageManager(\"\") returned a package manager (ok)")
	} else {
		t.Fatal("GetPackageManager(\"\") returned nil without error")
	}
}
