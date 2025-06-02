// Working demo of the unified interface architecture
// This demonstrates how the new plugin system works without dependencies on legacy code
package main

import (
	"context"
	"fmt"
	"strings"
)

// === STANDALONE UNIFIED INTERFACE (simplified for demo) ===

// PackageManager defines the unified interface
type PackageManager interface {
	GetName() string
	GetType() string
	IsAvailable() bool
	Search(ctx context.Context, query []string) ([]PackageInfo, error)
	Install(ctx context.Context, packages []string) ([]PackageInfo, error)
}

// PackageInfo represents package information
type PackageInfo struct {
	Name        string
	Version     string
	Status      string
	Description string
	Metadata    map[string]interface{}
}

// Plugin interface for registration
type Plugin interface {
	CreateManager() PackageManager
	GetPriority() int
}

// Simple registry for demo
var registry = make(map[string]Plugin)

func Register(name string, plugin Plugin) {
	registry[name] = plugin
}

// === BASE MANAGER (demonstrates "Less is more") ===

type BaseManager struct {
	name        string
	managerType string
}

func NewBaseManager(name, managerType string) *BaseManager {
	return &BaseManager{
		name:        name,
		managerType: managerType,
	}
}

func (b *BaseManager) GetName() string   { return b.name }
func (b *BaseManager) GetType() string   { return b.managerType }
func (b *BaseManager) IsAvailable() bool { return true } // Default: available

// Default implementations return "not supported"
func (b *BaseManager) Search(ctx context.Context, query []string) ([]PackageInfo, error) {
	return nil, fmt.Errorf("%s: search operation not supported", b.name)
}

func (b *BaseManager) Install(ctx context.Context, packages []string) ([]PackageInfo, error) {
	return nil, fmt.Errorf("%s: install operation not supported", b.name)
}

// === EXAMPLE PLUGIN 1: Mock APT Manager ===

type MockAPTManager struct {
	*BaseManager
	packages map[string]string // name -> version
}

func NewMockAPTManager() *MockAPTManager {
	return &MockAPTManager{
		BaseManager: NewBaseManager("mock-apt", "system"),
		packages: map[string]string{
			"vim":     "8.2.0716",
			"curl":    "7.68.0",
			"git":     "2.25.1",
			"firefox": "109.0.1",
		},
	}
}

// Override only the operations APT supports
func (m *MockAPTManager) Search(ctx context.Context, query []string) ([]PackageInfo, error) {
	var results []PackageInfo

	for name, version := range m.packages {
		for _, q := range query {
			if strings.Contains(name, strings.ToLower(q)) {
				pkg := PackageInfo{
					Name:        name,
					Version:     version,
					Status:      "available",
					Description: fmt.Sprintf("APT package %s", name),
					Metadata:    map[string]interface{}{"source": "apt"},
				}
				results = append(results, pkg)
				break
			}
		}
	}

	return results, nil
}

func (m *MockAPTManager) Install(ctx context.Context, packages []string) ([]PackageInfo, error) {
	var results []PackageInfo

	for _, pkg := range packages {
		if version, exists := m.packages[pkg]; exists {
			result := PackageInfo{
				Name:        pkg,
				Version:     version,
				Status:      "installed",
				Description: fmt.Sprintf("Successfully installed %s", pkg),
				Metadata:    map[string]interface{}{"method": "apt install"},
			}
			results = append(results, result)
		} else {
			return nil, fmt.Errorf("package '%s' not found in APT repositories", pkg)
		}
	}

	return results, nil
}

// APT Plugin
type APTPlugin struct{}

func (p *APTPlugin) CreateManager() PackageManager { return NewMockAPTManager() }
func (p *APTPlugin) GetPriority() int              { return 90 } // High priority for system packages

// === EXAMPLE PLUGIN 2: Mock npm Manager ===

type MockNPMManager struct {
	*BaseManager
	packages map[string]string
}

func NewMockNPMManager() *MockNPMManager {
	return &MockNPMManager{
		BaseManager: NewBaseManager("mock-npm", "language"),
		packages: map[string]string{
			"react":      "18.2.0",
			"lodash":     "4.17.21",
			"express":    "4.18.2",
			"typescript": "4.9.5",
		},
	}
}

func (m *MockNPMManager) Search(ctx context.Context, query []string) ([]PackageInfo, error) {
	var results []PackageInfo

	for name, version := range m.packages {
		for _, q := range query {
			if strings.Contains(name, strings.ToLower(q)) {
				pkg := PackageInfo{
					Name:        name,
					Version:     version,
					Status:      "available",
					Description: fmt.Sprintf("npm package %s", name),
					Metadata: map[string]interface{}{
						"source":   "npm",
						"registry": "https://registry.npmjs.org",
					},
				}
				results = append(results, pkg)
				break
			}
		}
	}

	return results, nil
}

func (m *MockNPMManager) Install(ctx context.Context, packages []string) ([]PackageInfo, error) {
	var results []PackageInfo

	for _, pkg := range packages {
		if version, exists := m.packages[pkg]; exists {
			result := PackageInfo{
				Name:        pkg,
				Version:     version,
				Status:      "installed",
				Description: fmt.Sprintf("Successfully installed %s via npm", pkg),
				Metadata: map[string]interface{}{
					"method": "npm install",
					"global": false,
				},
			}
			results = append(results, result)
		} else {
			return nil, fmt.Errorf("package '%s' not found in npm registry", pkg)
		}
	}

	return results, nil
}

// npm Plugin
type NPMPlugin struct{}

func (p *NPMPlugin) CreateManager() PackageManager { return NewMockNPMManager() }
func (p *NPMPlugin) GetPriority() int              { return 70 } // Medium priority for language packages

// === EXAMPLE PLUGIN 3: Mock Steam Manager (demonstrates different operations) ===

type MockSteamManager struct {
	*BaseManager
	games map[string]string // appid -> name
}

func NewMockSteamManager() *MockSteamManager {
	return &MockSteamManager{
		BaseManager: NewBaseManager("mock-steam", "game"),
		games: map[string]string{
			"730":    "Counter-Strike: Global Offensive",
			"440":    "Team Fortress 2",
			"570":    "Dota 2",
			"271590": "Grand Theft Auto V",
		},
	}
}

func (m *MockSteamManager) Search(ctx context.Context, query []string) ([]PackageInfo, error) {
	var results []PackageInfo

	for appid, name := range m.games {
		for _, q := range query {
			if strings.Contains(strings.ToLower(name), strings.ToLower(q)) {
				pkg := PackageInfo{
					Name:        name,
					Version:     "latest",
					Status:      "available",
					Description: fmt.Sprintf("Steam game: %s", name),
					Metadata: map[string]interface{}{
						"appid":    appid,
						"platform": "steam",
						"type":     "game",
					},
				}
				results = append(results, pkg)
				break
			}
		}
	}

	return results, nil
}

func (m *MockSteamManager) Install(ctx context.Context, packages []string) ([]PackageInfo, error) {
	var results []PackageInfo

	for _, pkg := range packages {
		// Steam uses app IDs, but for demo we'll accept game names
		var appid, gameName string
		var found bool

		// Try to find by name or appid
		for id, gameNameLoop := range m.games {
			if gameNameLoop == pkg || id == pkg {
				appid = id
				gameName = gameNameLoop
				found = true
				break
			}
		}

		if found {
			result := PackageInfo{
				Name:        gameName,
				Version:     "latest",
				Status:      "installed",
				Description: fmt.Sprintf("Successfully installed %s via Steam", gameName),
				Metadata: map[string]interface{}{
					"appid":           appid,
					"download_method": "steam",
					"install_size":    "15.2 GB",
				},
			}
			results = append(results, result)
		} else {
			return nil, fmt.Errorf("game '%s' not found in Steam library", pkg)
		}
	}

	return results, nil
}

// Steam Plugin
type SteamPlugin struct{}

func (p *SteamPlugin) CreateManager() PackageManager { return NewMockSteamManager() }
func (p *SteamPlugin) GetPriority() int              { return 60 } // Lower priority (specialized)

// === DEMO MAIN FUNCTION ===

func main() {
	fmt.Println("🚀 go-syspkg Unified Interface Architecture Demo")
	fmt.Println("================================================")

	// Register plugins (in real implementation this would be via init())
	Register("mock-apt", &APTPlugin{})
	Register("mock-npm", &NPMPlugin{})
	Register("mock-steam", &SteamPlugin{})

	// Get available managers
	fmt.Println("\n📦 Available Package Managers:")
	for _, plugin := range registry {
		mgr := plugin.CreateManager()
		fmt.Printf("   • %s (%s) - Priority: %d\n",
			mgr.GetName(), mgr.GetType(), plugin.GetPriority())
	}

	ctx := context.Background()

	// Demo 1: Search across all managers
	fmt.Println("\n🔍 Demo 1: Search for 'git' across all managers")
	fmt.Println("----------------------------------------------")
	searchTerm := []string{"git"}

	for _, plugin := range registry {
		mgr := plugin.CreateManager()
		results, err := mgr.Search(ctx, searchTerm)

		if err != nil {
			fmt.Printf("   %s: %v\n", mgr.GetName(), err)
		} else if len(results) > 0 {
			fmt.Printf("   %s found %d packages:\n", mgr.GetName(), len(results))
			for _, pkg := range results {
				fmt.Printf("     - %s v%s (%s)\n", pkg.Name, pkg.Version, pkg.Status)
			}
		} else {
			fmt.Printf("   %s: no packages found\n", mgr.GetName())
		}
	}

	// Demo 2: Install packages using specific managers
	fmt.Println("\n💾 Demo 2: Install packages using different managers")
	fmt.Println("---------------------------------------------------")

	// Install system package via APT
	aptMgr := registry["mock-apt"].CreateManager()
	aptResults, err := aptMgr.Install(ctx, []string{"vim"})
	if err != nil {
		fmt.Printf("   APT install failed: %v\n", err)
	} else {
		fmt.Printf("   APT installed: %s v%s\n", aptResults[0].Name, aptResults[0].Version)
	}

	// Install language package via npm
	npmMgr := registry["mock-npm"].CreateManager()
	npmResults, err := npmMgr.Install(ctx, []string{"react"})
	if err != nil {
		fmt.Printf("   npm install failed: %v\n", err)
	} else {
		fmt.Printf("   npm installed: %s v%s\n", npmResults[0].Name, npmResults[0].Version)
	}

	// Install game via Steam
	steamMgr := registry["mock-steam"].CreateManager()
	steamResults, err := steamMgr.Install(ctx, []string{"Team Fortress 2"})
	if err != nil {
		fmt.Printf("   Steam install failed: %v\n", err)
	} else {
		fmt.Printf("   Steam installed: %s (%s)\n",
			steamResults[0].Name, steamResults[0].Metadata["appid"])
	}

	// Demo 3: Show how unsupported operations work
	fmt.Println("\n❌ Demo 3: Unsupported operations return clear errors")
	fmt.Println("----------------------------------------------------")

	// Create a minimal manager that only supports search
	type MinimalManager struct {
		*BaseManager
	}

	minimalMgr := &MinimalManager{
		BaseManager: NewBaseManager("minimal", "demo"),
	}

	// Search works (inherited from BaseManager default)
	_, err = minimalMgr.Search(ctx, []string{"test"})
	fmt.Printf("   Minimal manager search: %v\n", err)

	// Install doesn't work (returns clear error)
	_, err = minimalMgr.Install(ctx, []string{"test"})
	fmt.Printf("   Minimal manager install: %v\n", err)

	// Demo 4: Show metadata flexibility
	fmt.Println("\n📋 Demo 4: Package-specific metadata")
	fmt.Println("------------------------------------")

	reactResults, _ := npmMgr.Search(ctx, []string{"react"})
	if len(reactResults) > 0 {
		pkg := reactResults[0]
		fmt.Printf("   Package: %s\n", pkg.Name)
		fmt.Printf("   Metadata:\n")
		for key, value := range pkg.Metadata {
			fmt.Printf("     %s: %v\n", key, value)
		}
	}

	fmt.Println("\n✅ Demo Complete!")
	fmt.Println("==================")
	fmt.Println("Key Benefits Demonstrated:")
	fmt.Println("• Unified interface across different package types")
	fmt.Println("• Easy plugin registration and discovery")
	fmt.Println("• Graceful handling of unsupported operations")
	fmt.Println("• Flexible metadata for package-specific data")
	fmt.Println("• 'Less is more' - BaseManager provides defaults")
	fmt.Println("• Type safety with Go interfaces")
	fmt.Println("\nAdding a new package manager requires ~50 lines!")
}
