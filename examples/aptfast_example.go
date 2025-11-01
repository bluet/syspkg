package main

import (
	"fmt"
	"log"

	"github.com/bluet/syspkg"
	"github.com/bluet/syspkg/manager"
	"github.com/bluet/syspkg/manager/apt"
)

func main() {
	// Example 1: Using apt-fast via the SysPkg interface
	fmt.Println("=== Example 1: Using apt-fast via SysPkg ===")
	includeOptions := syspkg.IncludeOptions{
		AptFast: true, // Enable apt-fast
	}

	sysPkg, err := syspkg.New(includeOptions)
	if err != nil {
		log.Fatalf("Failed to create SysPkg: %v", err)
	}

	// Get the apt-fast package manager
	aptFastManager, err := sysPkg.GetPackageManager("apt-fast")
	if err != nil {
		log.Fatalf("Failed to get apt-fast manager: %v", err)
	}

	fmt.Printf("Package Manager: %s\n", aptFastManager.GetPackageManager())

	// Search for a package using apt-fast
	opts := &manager.Options{
		Verbose: true,
	}

	packages, err := aptFastManager.Find([]string{"vim"}, opts)
	if err != nil {
		log.Fatalf("Failed to search for packages: %v", err)
	}

	fmt.Printf("Found %d packages matching 'vim'\n", len(packages))
	for _, pkg := range packages {
		fmt.Printf("  - %s (%s) [%s]\n", pkg.Name, pkg.Version, pkg.Status)
	}

	// Example 2: Using apt-fast directly
	fmt.Println("\n=== Example 2: Using apt-fast directly ===")
	aptFastDirect := apt.NewPackageManagerWithBinary("apt-fast")

	if aptFastDirect.IsAvailable() {
		fmt.Println("apt-fast is available on this system")

		// List installed packages
		installed, err := aptFastDirect.ListInstalled(&manager.Options{})
		if err != nil {
			log.Fatalf("Failed to list installed packages: %v", err)
		}
		fmt.Printf("Total installed packages: %d\n", len(installed))

		// Check for upgradable packages
		upgradable, err := aptFastDirect.ListUpgradable(&manager.Options{})
		if err != nil {
			log.Fatalf("Failed to list upgradable packages: %v", err)
		}
		fmt.Printf("Upgradable packages: %d\n", len(upgradable))
	} else {
		fmt.Println("apt-fast is not available on this system")
		fmt.Println("Falling back to regular apt...")

		// Fall back to regular apt if apt-fast is not available
		aptManager := apt.NewPackageManager()
		if aptManager.IsAvailable() {
			fmt.Println("Using regular apt")
			packages, err := aptManager.Find([]string{"vim"}, opts)
			if err != nil {
				log.Fatalf("Failed to search for packages: %v", err)
			}
			fmt.Printf("Found %d packages matching 'vim' using apt\n", len(packages))
		}
	}

	// Example 3: Auto-detection - prefer apt-fast over apt
	fmt.Println("\n=== Example 3: Auto-detection ===")
	autoOptions := syspkg.IncludeOptions{
		AllAvailable: true, // This will detect both apt and apt-fast if available
	}

	autoSysPkg, err := syspkg.New(autoOptions)
	if err != nil {
		log.Fatalf("Failed to create SysPkg: %v", err)
	}

	// Try to get apt-fast first, fall back to apt
	var pkgManager syspkg.PackageManager
	pkgManager, err = autoSysPkg.GetPackageManager("apt-fast")
	if err != nil {
		fmt.Println("apt-fast not available, using apt")
		pkgManager, err = autoSysPkg.GetPackageManager("apt")
		if err != nil {
			log.Fatalf("Neither apt-fast nor apt is available: %v", err)
		}
	}

	fmt.Printf("Using package manager: %s\n", pkgManager.GetPackageManager())

	// Example 4: Using apt-fast for package installation (requires root)
	fmt.Println("\n=== Example 4: Package Installation (demonstration only) ===")
	fmt.Println("Note: The following operations require root privileges")

	// Refresh package list
	fmt.Println("Refreshing package list...")
	// err = aptFastDirect.Refresh(&manager.Options{})
	// if err != nil {
	//     log.Fatalf("Failed to refresh: %v", err)
	// }

	// Install a package (commented out to prevent actual installation)
	fmt.Println("To install a package:")
	fmt.Println("  packages, err := aptFastDirect.Install([]string{\"package-name\"}, &manager.Options{})")

	// Upgrade packages
	fmt.Println("To upgrade packages:")
	fmt.Println("  packages, err := aptFastDirect.Upgrade([]string{\"package-name\"}, &manager.Options{})")

	// Upgrade all packages
	fmt.Println("To upgrade all packages:")
	fmt.Println("  packages, err := aptFastDirect.UpgradeAll(&manager.Options{})")
}
