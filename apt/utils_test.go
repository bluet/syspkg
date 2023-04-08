package apt_test

import (
	"reflect"
	"testing"

	"github.com/bluet/syspkg/apt"
	"github.com/bluet/syspkg/internal"
)

func TestParseInstallOutput(t *testing.T) {
	var inputParseInstallOutput string = `Setting up libglib2.0-0:amd64 (2.56.4-0ubuntu0.18.04.4) ...
Setting up libglib2.0-data (2.56.4-0ubuntu0.18.04.4) ...
Setting up libglib2.0-bin (2.56.4-0ubuntu0.18.04.4) ...
Processing triggers for libc-bin (2.27-3ubuntu1) ...`

	var expectedPackageInfo = []internal.PackageInfo{
		{
			Name:           "libglib2.0-0",
			Version:        "2.56.4-0ubuntu0.18.04.4",
			NewVersion:     "2.56.4-0ubuntu0.18.04.4",
			Status:         internal.PackageStatusInstalled,
			Category:       "",
			Arch:           "amd64",
			PackageManager: "apt",
		},
		{
			Name:           "libglib2.0-data",
			Version:        "2.56.4-0ubuntu0.18.04.4",
			NewVersion:     "2.56.4-0ubuntu0.18.04.4",
			Status:         internal.PackageStatusInstalled,
			Category:       "",
			Arch:           "",
			PackageManager: "apt",
		},
		{
			Name:           "libglib2.0-bin",
			Version:        "2.56.4-0ubuntu0.18.04.4",
			NewVersion:     "2.56.4-0ubuntu0.18.04.4",
			Status:         internal.PackageStatusInstalled,
			Category:       "",
			Arch:           "",
			PackageManager: "apt",
		},
	}

	actualPackageInfo := apt.ParseInstallOutput(inputParseInstallOutput, &internal.Options{Verbose: true})

	if !reflect.DeepEqual(expectedPackageInfo, actualPackageInfo) {
		t.Errorf("ParseInstallOutput() = %v, want %v", actualPackageInfo, expectedPackageInfo)
	}
}

func TestParseDeletedOutput(t *testing.T) {
	var inputParseDeletedeOutput string = `Reading package lists...
Building dependency tree...
Reading state information...
The following packages were automatically installed and are no longer required:
  libglib2.0-0 libglib2.0-bin libglib2.0-data
Use 'sudo apt autoremove' to remove them.
The following packages will be REMOVED:
  libglib2.0-0:amd64 libglib2.0-bin libglib2.0-data
0 upgraded, 0 newly installed, 3 to remove and 0 not upgraded.
After this operation, 3,072 kB disk space will be freed.
Do you want to continue? [Y/n]
(Reading database ... 123456 files and directories currently installed.)
Removing pkg1.2-3:amd64 (1.2.3-0ubuntu0.18.04.4) ...
Removing pkg2.0-bin (v2) ...
Removing pkg3.0-data (22222A-A) ...`

	var expectedPackageInfo = []internal.PackageInfo{
		{
			Name:           "pkg1.2-3",
			Version:        "1.2.3-0ubuntu0.18.04.4",
			NewVersion:     "",
			Status:         internal.PackageStatusAvailable,
			Category:       "",
			Arch:           "amd64",
			PackageManager: "apt",
		},
		{
			Name:           "pkg2.0-bin",
			Version:        "v2",
			NewVersion:     "",
			Status:         internal.PackageStatusAvailable,
			Category:       "",
			Arch:           "",
			PackageManager: "apt",
		},
		{
			Name:           "pkg3.0-data",
			Version:        "22222A-A",
			NewVersion:     "",
			Status:         internal.PackageStatusAvailable,
			Category:       "",
			Arch:           "",
			PackageManager: "apt",
		},
	}

	actualPackageInfo := apt.ParseDeletedOutput(inputParseDeletedeOutput, &internal.Options{Verbose: true})

	if !reflect.DeepEqual(expectedPackageInfo, actualPackageInfo) {
		t.Errorf("ParseDeletedOutput() = %v, want %v", actualPackageInfo, expectedPackageInfo)
	}
}
