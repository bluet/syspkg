package flatpak

import (
	"fmt"
	"log"
	"strings"

	// "github.com/rs/zerolog"
	// "github.com/rs/zerolog/log"

	"github.com/bluet/syspkg/manager"
)

func ParseInstallOutput(msg string, opts *manager.Options) []manager.PackageInfo {
	var packages []manager.PackageInfo

	// cspell: disable
	// command:
	// flatpak install flathub net.davidotek.pupgui2 --noninteractive -y --verbose
	//
	// output:
	// F: Transaction: install flathub:app/net.davidotek.pupgui2/x86_64/stable[*]
	// F: Looking for remote metadata updates for flathub
	// F: marking op install:app/net.davidotek.pupgui2/x86_64/stable resolved to 8150b5ebfa488c4dc35fa52ceca13d403b8b1f0ce9021f0e9e69e67b9fbedc2b
	// F: Updating dependent runtime org.kde.Platform/x86_64/6.4
	// F: Transaction: update flathub:runtime/org.kde.Platform/x86_64/6.4[$old]
	// F: marking op update:runtime/org.kde.Platform/x86_64/6.4 resolved to 0762ee04666309b88870e809433f25bfc1392856a43dcc46fbeab31b50e64d4a
	// F: Transaction: install/update flathub:runtime/org.freedesktop.Platform.GL.default/x86_64/22.08[*]
	// F: Transaction: install/update flathub:runtime/org.freedesktop.Platform.GL.default/x86_64/22.08-extra[*]
	// F: Transaction: install/update flathub:runtime/org.freedesktop.Platform.VAAPI.Intel/x86_64/22.08[*]
	// F: Transaction: install/update flathub:runtime/org.freedesktop.Platform.openh264/x86_64/2.2.0[*]
	// F: Transaction: install/update flathub:runtime/org.gtk.Gtk3theme.Ambiance/x86_64/3.22[*]
	// F: Transaction: install/update flathub:runtime/org.gtk.Gtk3theme.Yaru/x86_64/3.22[*]
	// F: Transaction: install/update flathub:runtime/org.kde.PlatformTheme.QGnomePlatform/x86_64/6.4[*]
	// F: Transaction: install/update flathub:runtime/org.kde.WaylandDecoration.QGnomePlatform-decoration/x86_64/6.4[*]
	// F: Transaction: install/update flathub:runtime/org.kde.Platform.Locale/x86_64/6.4[/en, /zh]
	// F: marking op install/update:runtime/org.kde.Platform.Locale/x86_64/6.4 resolved to b606e8bb247fa5e06eea0adbac50d8d74052557e3c4407e29671a169b916b7df
	// F: marking op install/update:runtime/org.kde.WaylandDecoration.QGnomePlatform-decoration/x86_64/6.4 resolved to 7f13e35c40b0d8b100ce9b2009b9a433193888247f32d312e3ce0d520a24f9fd
	// F: marking op install/update:runtime/org.kde.PlatformTheme.QGnomePlatform/x86_64/6.4 resolved to 4d6563598e49461783273b6204f9e2571b4ecdc1291897a9e45847eff444b2e4
	// F: marking op install/update:runtime/org.gtk.Gtk3theme.Yaru/x86_64/3.22 resolved to 4b1e043544efb4a6d0278a3a2deaede9f6ff61034589bdb380054121aa098952
	// F: marking op install/update:runtime/org.gtk.Gtk3theme.Ambiance/x86_64/3.22 resolved to 73fed99df212c179f776452d1eb4f49d9e8fd87ba6f8434816db2e9c6c92ea19
	// F: marking op install/update:runtime/org.freedesktop.Platform.openh264/x86_64/2.2.0 resolved to bf24f23f3ba385f6e8c9215ed94d979db99814b0b614504a23a6d0751dc5f063
	// F: marking op install/update:runtime/org.freedesktop.Platform.VAAPI.Intel/x86_64/22.08 resolved to 601730e8e63a0687a10b388f5180f41ac1b3ed5a958f4b72b7aa61c334a54f18
	// F: marking op install/update:runtime/org.freedesktop.Platform.GL.default/x86_64/22.08-extra resolved to 87220a5fe19b979e65651eb6eb25719701392e1702f41d81980046a62ff527b3
	// F: marking op install/update:runtime/org.freedesktop.Platform.GL.default/x86_64/22.08 resolved to 1bbf632d2739210cb50bddcd1844c0ad33926068790b048546e9e8b983ba307a
	// Installing app/net.davidotek.pupgui2/x86_64/stable
	// cspell: enable



	// remove the last empty line
	msg = strings.TrimSuffix(msg, "\n")
	var lines []string = strings.Split(string(msg), "\n")

	for _, line := range lines {
		if opts.Verbose {
			log.Printf("%s: %s", pm, line)
		}
		if strings.HasPrefix(line, "marking op ") {
			var status manager.PackageStatus = manager.PackageStatusInstalled
			var msgParts []string = strings.Split(line, " ")
			var action string = strings.Split(msgParts[3], ":")[0]
			var pkgString string = strings.TrimPrefix(msgParts[3], action+":")
			var parts []string = strings.Split(pkgString, "/")
			var name string = parts[1]
			var arch string = parts[2]
			var version string = parts[3]
			var category string = parts[0]

			// if name is empty, it might be not what we want
			if name == "" {
				continue
			}

			if msgParts[4] != "resolved" {
				status = manager.PackageStatusUnknown
				// TODO: this might be an error
				fmt.Printf("package install/update unresolved: %s", line)
			} else if strings.HasPrefix(action, "install") || strings.HasPrefix(action, "update") {
				status = manager.PackageStatusInstalled
			} else if strings.HasPrefix(action, "uninstall") {
				status = manager.PackageStatusAvailable
			}


			packageInfo := manager.PackageInfo{
				Name:           name,
				Arch:           arch,
				Version:        version,
				NewVersion:     version,
				Category:       category,
				Status:         status,
				PackageManager: pm,
			}
			packages = append(packages, packageInfo)
		}
	}

	return packages
}

func ParseFindOutput(msg string, opts *manager.Options) []manager.PackageInfo {
	// FreeRDP Remote Desktop Client	FreeRDP (Remote Desktop Protocol) Client for Linux.	com.freerdp.FreeRDP	2.10.0	stable	flathub
	// Fightcade	Play arcade games online.	com.fightcade.Fightcade	2.2	stable	flathub
	// White House	Using the magic of CSS, hack your world into a unique burst of color and light revealing hidden objects and clues.	com.endlessnetwork.whitehouse	1.175	stable	flathub
	var packages []manager.PackageInfo

	msg = strings.TrimSuffix(msg, "\n")
	var lines []string = strings.Split(string(msg), "\n")

	for _, line := range lines {
		if opts.Verbose {
			log.Printf("%s: %s", pm, line)
		}

		if len(line) == 0 || strings.HasPrefix(line, "Name\t") {
			continue
		}

		var parts []string = strings.Split(line, "\t")
		var name string = parts[2]
		// var arch string = ""
		var version string = parts[3]
		// var category string = parts[5]

		packageInfo := manager.PackageInfo{
			Name:           name,
			// Arch:           arch,
			Version:        version,
			NewVersion:     version,
			// Category:       category,
			Status:         manager.PackageStatusAvailable,
			PackageManager: pm,
		}
		packages = append(packages, packageInfo)
	}

	return packages
}

func ParseListInstalledOutput(msg string, opts *manager.Options) []manager.PackageInfo {
	var packages []manager.PackageInfo

	msg = strings.TrimSuffix(msg, "\n")
	var lines []string = strings.Split(string(msg), "\n")

	for _, line := range lines {
		if opts.Verbose {
			log.Printf("%s: %s", pm, line)
		}

		if len(line) == 0 || strings.HasPrefix(line, "Name\t") {
			continue
		}

		var parts []string = strings.Split(line, "\t")
		var name string = parts[1]
		// var arch string = ""
		var version string = parts[2]
		// var category string = parts[5]

		packageInfo := manager.PackageInfo{
			Name:           name,
			// Arch:           arch,
			Version:        version,
			// NewVersion:     version,
			// Category:       category,
			Status:         manager.PackageStatusInstalled,
			PackageManager: pm,
		}
		packages = append(packages, packageInfo)
	}

	return packages
}


func ParseListUpgradableOutput(msg string, opts *manager.Options) []manager.PackageInfo {
	var packages []manager.PackageInfo

	msg = strings.TrimSuffix(msg, "\n")
	var lines []string = strings.Split(string(msg), "\n")

	for _, line := range lines {
		if opts.Verbose {
			log.Printf("%s: %s", pm, line)
		}

		if len(line) == 0 || strings.HasPrefix(line, "Name\t") {
			continue
		}

		var parts []string = strings.Split(line, "\t")
		var name string = parts[1]
		var arch string = parts[4]
		var version string = parts[2]
		if version == "" {
			version = "unknown"
		}
		// var category string = parts[5]

		packageInfo := manager.PackageInfo{
			Name:           name,
			Arch:           arch,
			// Version:        version,
			NewVersion:     version,
			// Category:       category,
			Status:         manager.PackageStatusInstalled,
			PackageManager: pm,
		}
		packages = append(packages, packageInfo)
	}

	return packages
}


func ParsePackageInfoOutput(msg string, opts *manager.Options) manager.PackageInfo {
	var pkg manager.PackageInfo

	// remove the last empty line
	msg = strings.TrimSuffix(msg, "\n")
	lines := strings.Split(string(msg), "\n")

	for _, line := range lines {
		// remove all leading and trailing spaces
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			parts := strings.SplitN(line, ":", 2)

			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "ID":
				pkg.Name = value
			case "Version":
				pkg.Version = value
			case "Arch":
				pkg.Arch = value
			// case "Section":
			// 	pkg.Category = value
			}
		}
	}

	pkg.PackageManager = "apt"

	return pkg
}
