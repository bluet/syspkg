package apt_test

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/bluet/syspkg/manager"
	"github.com/bluet/syspkg/manager/apt"
)

func TestParseInstallOutput(t *testing.T) {
	var inputParseInstallOutput string = strings.Join([]string{
		`Setting up libglib2.0-0:amd64 (2.56.4-0ubuntu0.18.04.4) ...`,
		`Setting up libglib2.0-data (2.56.4-0ubuntu0.18.04.4) ...`,
		`Setting up libglib2.0-bin (2.56.4-0ubuntu0.18.04.4) ...`,
		`Processing triggers for libc-bin (2.27-3ubuntu1) ...`,
	}, "\n")
	// `Setting up libglib2.0-0:amd64 (2.56.4-0ubuntu0.18.04.4) ...`
	// 	+ `Setting up libglib2.0-data (2.56.4-0ubuntu0.18.04.4) ...`
	// 	+ `Setting up libglib2.0-bin (2.56.4-0ubuntu0.18.04.4) ...`
	// 	+ `Processing triggers for libc-bin (2.27-3ubuntu1) ...`

	var expectedPackageInfo = []manager.PackageInfo{
		{
			Name:           "libglib2.0-0",
			Version:        "2.56.4-0ubuntu0.18.04.4",
			NewVersion:     "2.56.4-0ubuntu0.18.04.4",
			Status:         manager.PackageStatusInstalled,
			Category:       "",
			Arch:           "amd64",
			PackageManager: "apt",
		},
		{
			Name:           "libglib2.0-data",
			Version:        "2.56.4-0ubuntu0.18.04.4",
			NewVersion:     "2.56.4-0ubuntu0.18.04.4",
			Status:         manager.PackageStatusInstalled,
			Category:       "",
			Arch:           "",
			PackageManager: "apt",
		},
		{
			Name:           "libglib2.0-bin",
			Version:        "2.56.4-0ubuntu0.18.04.4",
			NewVersion:     "2.56.4-0ubuntu0.18.04.4",
			Status:         manager.PackageStatusInstalled,
			Category:       "",
			Arch:           "",
			PackageManager: "apt",
		},
	}

	actualPackageInfo := apt.ParseInstallOutput(inputParseInstallOutput, &manager.Options{})

	if !reflect.DeepEqual(expectedPackageInfo, actualPackageInfo) {
		t.Errorf("ParseInstallOutput() = %+v, want %+v", actualPackageInfo, expectedPackageInfo)
	}
}

func TestParseDeletedOutput(t *testing.T) {
	var inputParseDeletedeOutput string = strings.Join([]string{
		`Reading package lists...`,
		`Building dependency tree...`,
		`Reading state information...`,
		`The following packages were automatically installed and are no longer required:`,
		`  libglib2.0-0 libglib2.0-bin libglib2.0-data`,
		`Use 'sudo apt autoremove' to remove them.`,
		`The following packages will be REMOVED:`,
		`  libglib2.0-0:amd64 libglib2.0-bin libglib2.0-data`,
		`0 upgraded, 0 newly installed, 3 to remove and 0 not upgraded.`,
		`After this operation, 3,072 kB disk space will be freed.`,
		`Do you want to continue? [Y/n]`,
		`(Reading database ... 123456 files and directories currently installed.)`,
		`Removing pkg1.2-3:amd64 (1.2.3-0ubuntu0.18.04.4) ...`,
		`Removing pkg2.0-bin (v2) ...`,
		`Removing pkg3.0-data (22222A-A) ...)`,
	}, "\n")

	var expectedPackageInfo = []manager.PackageInfo{
		{
			Name:           "pkg1.2-3",
			Version:        "1.2.3-0ubuntu0.18.04.4",
			NewVersion:     "",
			Status:         manager.PackageStatusAvailable,
			Category:       "",
			Arch:           "amd64",
			PackageManager: "apt",
		},
		{
			Name:           "pkg2.0-bin",
			Version:        "v2",
			NewVersion:     "",
			Status:         manager.PackageStatusAvailable,
			Category:       "",
			Arch:           "",
			PackageManager: "apt",
		},
		{
			Name:           "pkg3.0-data",
			Version:        "22222A-A",
			NewVersion:     "",
			Status:         manager.PackageStatusAvailable,
			Category:       "",
			Arch:           "",
			PackageManager: "apt",
		},
	}

	actualPackageInfo := apt.ParseDeletedOutput(inputParseDeletedeOutput, &manager.Options{})

	if !reflect.DeepEqual(expectedPackageInfo, actualPackageInfo) {
		t.Errorf("ParseDeletedOutput() = %+v, want %+v", actualPackageInfo, expectedPackageInfo)
	}
}

func TestParseFindOutput(t *testing.T) {
	var inputParseSearchOutput string = strings.Join([]string{
		`Sorting...`,
		`Full Text Search...`,
		`zutty/jammy 0.11.2.20220109.192032+dfsg1-1 amd64`,
		`Efficient full-featured X11 terminal emulator`,
		``,
		`zvbi/jammy 0.2.35-19 amd64`,
		`Vertical Blanking Interval (VBI) utilities`,
	}, "\n")

	var expectedPackageInfo = []manager.PackageInfo{
		{
			Name: "zutty",
			// Version:    "0.11.2.20220109.192032+dfsg1-1",
			// NewVersion: "",
			Version:        "",
			NewVersion:     "0.11.2.20220109.192032+dfsg1-1",
			Status:         manager.PackageStatusAvailable,
			Category:       "jammy",
			Arch:           "amd64",
			PackageManager: "apt",
		},
		{
			Name: "zvbi",
			// Version:    "0.2.35-19",
			// NewVersion: "",
			Version:        "",
			NewVersion:     "0.2.35-19",
			Status:         manager.PackageStatusAvailable,
			Category:       "jammy",
			Arch:           "amd64",
			PackageManager: "apt",
		},
	}

	actualPackageInfo := apt.ParseFindOutput(inputParseSearchOutput, &manager.Options{})

	if !reflect.DeepEqual(expectedPackageInfo, actualPackageInfo) {
		t.Errorf("ParseSearchOutput() = %+v, want %+v", actualPackageInfo, expectedPackageInfo)
	}
}

func TestParseInstalledOutput(t *testing.T) {
	var inputParseInstalledOutput = strings.Join([]string{
		`bind9-libs:amd64 1:9.18.12-0ubuntu0.22.04.1`,
		`binfmt-support 2.2.1-2`,
		`binutils 2.38-4ubuntu2.1`,
	}, "\n")

	var expectedPackageInfo = []manager.PackageInfo{
		{
			Name:           "bind9-libs",
			Version:        "1:9.18.12-0ubuntu0.22.04.1",
			NewVersion:     "",
			Status:         manager.PackageStatusInstalled,
			Category:       "",
			Arch:           "amd64",
			PackageManager: "apt",
		},
		{
			Name:           "binfmt-support",
			Version:        "2.2.1-2",
			NewVersion:     "",
			Status:         manager.PackageStatusInstalled,
			Category:       "",
			Arch:           "",
			PackageManager: "apt",
		},
		{
			Name:           "binutils",
			Version:        "2.38-4ubuntu2.1",
			NewVersion:     "",
			Status:         manager.PackageStatusInstalled,
			Category:       "",
			Arch:           "",
			PackageManager: "apt",
		},
	}

	actualPackageInfo := apt.ParseListInstalledOutput(inputParseInstalledOutput, &manager.Options{Verbose: true})

	if !reflect.DeepEqual(expectedPackageInfo, actualPackageInfo) {
		t.Errorf("ParseInstalledOutput() = %+v, want %+v", actualPackageInfo, expectedPackageInfo)
	}
}

func TestParseListUpgradable(t *testing.T) {
	var inputParseListUpgradable = strings.Join([]string{
		`Listing... Done`,
		`cloudflared/unknown 2023.4.0 amd64 [upgradable from: 2023.3.1]`,
		`libllvm15/jammy-updates 1:15.0.7-0ubuntu0.22.04.1 amd64 [upgradable from: 1:15.0.6-3~ubuntu0.22.04.2]`,
		`libllvm15/jammy-updates 1:15.0.7-0ubuntu0.22.04.1 i386 [upgradable from: 1:15.0.6-3~ubuntu0.22.04.2]`,
	}, "\n")

	var expectedPackageInfo = []manager.PackageInfo{
		{
			Name:           "cloudflared",
			Version:        "2023.3.1",
			NewVersion:     "2023.4.0",
			Status:         manager.PackageStatusUpgradable,
			Category:       "unknown",
			Arch:           "amd64",
			PackageManager: "apt",
		},
		{
			Name:           "libllvm15",
			Version:        "1:15.0.6-3~ubuntu0.22.04.2",
			NewVersion:     "1:15.0.7-0ubuntu0.22.04.1",
			Status:         manager.PackageStatusUpgradable,
			Category:       "jammy-updates",
			Arch:           "amd64",
			PackageManager: "apt",
		},
		{
			Name:           "libllvm15",
			Version:        "1:15.0.6-3~ubuntu0.22.04.2",
			NewVersion:     "1:15.0.7-0ubuntu0.22.04.1",
			Status:         manager.PackageStatusUpgradable,
			Category:       "jammy-updates",
			Arch:           "i386",
			PackageManager: "apt",
		},
	}

	actualPackageInfo := apt.ParseListUpgradableOutput(inputParseListUpgradable, &manager.Options{Verbose: true})

	if !reflect.DeepEqual(expectedPackageInfo, actualPackageInfo) {
		t.Errorf("ParseListUpgradable() = %+v, want %+v", actualPackageInfo, expectedPackageInfo)
	}
}

func TestParsePackageInfoOutput(t *testing.T) {
	var inputParsePackageInfoOutput = strings.Join([]string{
		`Package: cloudflared`,
		`Version: 2023.4.0`,
		`Priority: optional`,
		`Section: default`,
		`Maintainer: Cloudflare <support@cloudflare.com>`,
		`Installed-Size: 36.1 MB`,
		`Homepage: https://github.com/cloudflare/cloudflared`,
		`License: Apache License Version 2.0`,
		`Vendor: Cloudflare`,
		`Download-Size: 17.5 MB`,
		`APT-Sources: https://pkg.cloudflare.com/cloudflared jammy/main amd64 Packages`,
		`Description: Cloudflare Tunnel daemon`,
	}, "\n")

	var expectedPackageInfo = manager.PackageInfo{
		Name:           "cloudflared",
		Version:        "2023.4.0",
		NewVersion:     "",
		Status:         "",
		Category:       "default",
		Arch:           "",
		PackageManager: "apt",
	}

	actualPackageInfo := apt.ParsePackageInfoOutput(inputParsePackageInfoOutput, &manager.Options{})

	if !reflect.DeepEqual(expectedPackageInfo, actualPackageInfo) {
		t.Errorf("ParsePackageInfoOutput() = %+v, want %+v", actualPackageInfo, expectedPackageInfo)
	}
}

func TestParseDpkgQueryOutput(t *testing.T) {
	type args struct {
		output   []byte
		packages map[string]manager.PackageInfo
	}
	tests := []struct {
		name    string
		args    args
		want    []manager.PackageInfo
		wantErr bool
	}{
		{
			name: "ParseDpkgQueryOutput",
			args: args{
				output: bytes.Join(
					[][]byte{
						[]byte(`bash install ok installed 5.1-6ubuntu1`),
						[]byte(`cloudflared install ok installed 2023.3.1`),
						[]byte(`qemu-kvm deinstall ok config-files 1:4.2-3ubuntu6.23`),
						[]byte(`dpkg-query: no packages found matching ajsdjsks`),
						[]byte(`dpkg-query: no packages found matching byobu`),
					},
					[]byte("\n"),
				),

				packages: map[string]manager.PackageInfo{
					"bash":        {Name: "bash"},
					"cloudflared": {Name: "cloudflared"},
					"qemu-kvm":    {Name: "qemu-kvm"},
					"ajsdjsks":    {Name: "ajsdjsks"},
					"byobu":       {Name: "byobu"},
				},
			},
			want: []manager.PackageInfo{
				{Name: "bash", Status: manager.PackageStatusInstalled, Version: "5.1-6ubuntu1"},
				{Name: "cloudflared", Status: manager.PackageStatusInstalled, Version: "2023.3.1"},
				{Name: "qemu-kvm", Status: manager.PackageStatusConfigFiles, Version: "1:4.2-3ubuntu6.23"},
				{Name: "ajsdjsks", Status: manager.PackageStatusUnknown, Version: ""},
				{Name: "byobu", Status: manager.PackageStatusUnknown, Version: ""},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := apt.ParseDpkgQueryOutput(tt.args.output, tt.args.packages, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDpkgQueryOutput() error = %+v, wantErr %+v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseDpkgQueryOutput() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
