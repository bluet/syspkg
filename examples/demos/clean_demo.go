// Clean unified interface demo - ideal design
package main

import (
	"context"
	"fmt"

	"github.com/bluet/syspkg/manager"

	// Import clean plugins
	_ "github.com/bluet/syspkg/manager/apt"
)

func main() {
	fmt.Println("🎯 Clean Unified Interface Demo")
	fmt.Println("===============================")

	// Get available managers
	registry := manager.GetGlobalRegistry()
	managers := registry.GetAvailable()

	fmt.Printf("📦 Available: %d managers\n", len(managers))
	for name, mgr := range managers {
		fmt.Printf("   • %s (%s)\n", name, mgr.GetCategory())
	}

	// Get system package manager
	systemPM := registry.GetBestMatch(manager.CategorySystem)
	if systemPM == nil {
		fmt.Println("❌ No system package manager available")
		return
	}

	fmt.Printf("\n🔧 Using: %s\n", systemPM.GetName())

	// Test basic operations
	ctx := context.Background()
	opts := manager.DefaultOptions()

	// Search
	fmt.Printf("\n🔍 Searching for 'vim'...\n")
	packages, err := systemPM.Search(ctx, []string{"vim"}, opts)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Found: %d packages\n", len(packages))
		if len(packages) > 0 {
			pkg := packages[0]
			fmt.Printf("   Example: %s v%s (%s)\n", pkg.Name, pkg.NewVersion, pkg.Status)
		}
	}

	// Status
	fmt.Printf("\n📊 Manager Status:\n")
	status, err := systemPM.Status(ctx, opts)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Available: %v, Healthy: %v\n", status.Available, status.Healthy)
		if status.Version != "" {
			fmt.Printf("   Version: %s\n", status.Version)
		}
	}

	fmt.Println("\n✅ Clean design principles demonstrated:")
	fmt.Println("   • ~50 lines per plugin")
	fmt.Println("   • No legacy compatibility")
	fmt.Println("   • Pure unified interface")
	fmt.Println("   • Minimal, focused code")
}
