package main

import (
	"fmt"
	"log"

	"github.com/bluet/syspkg"
	"github.com/bluet/syspkg/manager"
	"github.com/bluet/syspkg/manager/apt"
)

func main() {
	// Example 1: Using apt-fast via GetPackageManagerWithOptions
	fmt.Println("=== Example 1: Using apt-fast via GetPackageManagerWithOptions ===")
	includeOptions := syspkg.IncludeOptions{
		Apt: true, // Enable apt
	}

	sysPkg, err := syspkg.New(includeOptions)
	if err != nil {
		log.Fatalf("Failed to create SysPkg: %v", err)
	}

	// Get apt manager with apt-fast binary
	aptFastManager, err := sysPkg.GetPackageManagerWithOptions("apt", &syspkg.ManagerCreationOptions{
		BinaryPath: "apt-fast",
	})
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

	// Example 3: Fallback pattern - prefer apt-fast, fallback to apt
	fmt.Println("\n=== Example 3: Fallback pattern ===")
	autoOptions := syspkg.IncludeOptions{
		Apt: true,
	}

	autoSysPkg, err := syspkg.New(autoOptions)
	if err != nil {
		log.Fatalf("Failed to create SysPkg: %v", err)
	}

	// Try to get apt-fast first, fall back to apt
	var pkgManager syspkg.PackageManager
	pkgManager, err = autoSysPkg.GetPackageManagerWithOptions("apt", &syspkg.ManagerCreationOptions{
		BinaryPath: "apt-fast",
	})
	if err != nil || !pkgManager.IsAvailable() {
		fmt.Println("apt-fast not available, using standard apt")
		pkgManager, err = autoSysPkg.GetPackageManager("apt")
		if err != nil {
			log.Fatalf("apt is not available: %v", err)
		}
	} else {
		fmt.Println("Using apt-fast for faster downloads")
	}

	fmt.Printf("Using package manager: %s\n", pkgManager.GetPackageManager())

	// Example 4: Using custom binary path (not just name)
	fmt.Println("\n=== Example 4: Custom binary path ===")
	fmt.Println("To use a custom binary path:")
	fmt.Println("  customApt, _ := syspkg.GetPackageManagerWithOptions(\"apt\", &syspkg.ManagerCreationOptions{")
	fmt.Println("      BinaryPath: \"/usr/local/bin/custom-apt\",")
	fmt.Println("  })")
	fmt.Println()
	fmt.Println("Or for development/testing:")
	fmt.Println("  devApt, _ := syspkg.GetPackageManagerWithOptions(\"apt\", &syspkg.ManagerCreationOptions{")
	fmt.Println("      BinaryPath: \"./my-test-apt\",")
	fmt.Println("  })")
}
