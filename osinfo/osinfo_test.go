package osinfo

import (
	"testing"
)

func TestGetOSInfo(t *testing.T) {
	osInfo, err := GetOSInfo()
	if err != nil {
		t.Fatalf("GetOSInfo() failed with error: %v", err)
	}

	if osInfo.Name == "" {
		t.Errorf("OS name is empty")
	}

	if osInfo.Version == "" {
		t.Errorf("OS version is empty")
	}

	if osInfo.Distribution == "" {
		t.Errorf("OS distribution is empty")
	}

	if osInfo.Arch == "" {
		t.Errorf("OS architecture is empty")
	}

	t.Logf("OS Info: %+v", osInfo)
}
