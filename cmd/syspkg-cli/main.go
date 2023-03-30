package main

import (
	"fmt"
	"os"

	"github.com/bluet/syspkg"
)

func main() {
	pms, err := syspkg.NewPackageManager()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// list upgradable packages for all package managers
	for _, pm := range pms {
		upgradablePackages, err := pm.ListUpgradable()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		fmt.Printf("Upgradable packages for %T:\n", pm)
		for _, pkg := range upgradablePackages {
			fmt.Printf("%s: %s %s -> %s (%s)\n", pkg.PackageManager, pkg.Name, pkg.Version, pkg.NewVersion, pkg.Status)
		}
	}
}
