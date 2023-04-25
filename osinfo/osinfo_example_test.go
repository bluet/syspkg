// osinfo_example_test.go
package osinfo_test

import (
	"fmt"

	"github.com/bluet/syspkg/osinfo"
)

// ExampleGetOSInfo demonstrates how to use the GetOSInfo function to obtain
// information about the operating system.
func ExampleGetOSInfo() {
	osInfo, err := osinfo.GetOSInfo()
	if err != nil {
		fmt.Println("Error getting OS info:", err)
		return
	}

	fmt.Println("Name:", osInfo.Name)
	fmt.Println("Distribution:", osInfo.Distribution)
	fmt.Println("Version:", osInfo.Version)
	fmt.Println("Architecture:", osInfo.Arch)
}
