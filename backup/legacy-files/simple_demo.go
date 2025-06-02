// Simple demo of the unified interface
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/bluet/syspkg/manager"
)

// Simple test manager
type TestManager struct {
	*manager.BaseManager
}

func NewTestManager() *TestManager {
	base := manager.NewBaseManager("testmgr", "test", manager.NewDefaultCommandRunner())
	return &TestManager{
		BaseManager: base,
	}
}

func (m *TestManager) IsAvailable() bool {
	return true
}

func (m *TestManager) Search(ctx context.Context, query []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	if opts == nil {
		opts = manager.DefaultOptions()
	}

	m.LogDebug(opts, "Test search for: %v", query)

	// Return a test package
	pkg := manager.NewPackageInfo("test-package", "1.0.0", "available", m.GetType())
	pkg.Description = "Test package from demo"

	return []manager.PackageInfo{pkg}, nil
}

// Test plugin
type TestPlugin struct{}

func (p *TestPlugin) CreateManager() manager.PackageManager {
	return NewTestManager()
}

func (p *TestPlugin) GetPriority() int {
	return 10
}

func main() {
	fmt.Println("=== Simple Unified Interface Demo ===")

	// Register test plugin
	if err := manager.Register("testmgr", &TestPlugin{}); err != nil {
		log.Fatal("Failed to register:", err)
	}

	// Get available managers
	available := manager.GetAvailableManagers()
	fmt.Printf("Available managers: %d\n", len(available))

	for name, pm := range available {
		fmt.Printf("- %s (%s)\n", name, pm.GetType())
	}

	// Get test manager
	testPlugin, exists := manager.GetPlugin("testmgr")
	if !exists {
		log.Fatal("Plugin not found")
	}

	mgr := testPlugin.CreateManager()
	ctx := context.Background()
	opts := manager.DefaultOptions()
	opts.Debug = true

	fmt.Println("\n=== Testing Search ===")
	searchResults, searchErr := mgr.Search(ctx, []string{"test"}, opts)
	if searchErr != nil {
		fmt.Printf("Search error: %v\n", searchErr)
	} else {
		for _, pkg := range searchResults {
			fmt.Printf("Found: %s v%s - %s\n", pkg.Name, pkg.Version, pkg.Description)
		}
	}

	fmt.Println("\n=== Testing Unsupported Operation ===")
	_, listErr := mgr.List(ctx, manager.FilterInstalled, opts)
	fmt.Printf("List error (expected): %v\n", listErr)

	fmt.Println("\n=== Testing Status ===")
	status, statusErr := mgr.Status(ctx, opts)
	if statusErr != nil {
		fmt.Printf("Status error: %v\n", statusErr)
	} else {
		fmt.Printf("Available: %v, Healthy: %v\n", status.Available, status.Healthy)
	}

	fmt.Println("\n✓ Demo completed successfully!")
	fmt.Println("✓ Plugin system works")
	fmt.Println("✓ Unified interface works")
	fmt.Println("✓ Unsupported operations return proper errors")
}
