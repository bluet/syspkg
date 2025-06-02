// Package syspkg provides a unified interface for package management across different systems
package syspkg

import (
	"context"
	"fmt"

	"github.com/bluet/syspkg/manager"

	// Import package manager plugins to auto-register them
	_ "github.com/bluet/syspkg/manager/apt"
	// Add more imports here as new plugins are created
)

// UnifiedSysPkg implements a unified package management interface using the plugin system
type UnifiedSysPkg struct {
	registry *manager.Registry
	managers map[string]manager.PackageManager
}

// NewUnifiedSysPkg creates a new unified package management instance
func NewUnifiedSysPkg() *UnifiedSysPkg {
	return &UnifiedSysPkg{
		registry: manager.GetGlobalRegistry(),
		managers: make(map[string]manager.PackageManager),
	}
}

// DiscoverManagers discovers and initializes all available package managers
func (s *UnifiedSysPkg) DiscoverManagers() map[string]manager.PackageManager {
	s.managers = s.registry.GetAvailable()
	return s.managers
}

// GetAvailableManagers returns all currently available package managers
func (s *UnifiedSysPkg) GetAvailableManagers() map[string]manager.PackageManager {
	if len(s.managers) == 0 {
		s.DiscoverManagers()
	}
	return s.managers
}

// GetManagerByName returns a specific package manager by name
func (s *UnifiedSysPkg) GetManagerByName(name string) (manager.PackageManager, error) {
	managers := s.GetAvailableManagers()
	if pm, exists := managers[name]; exists {
		return pm, nil
	}
	return nil, fmt.Errorf("package manager '%s' not found or not available", name)
}

// GetManagersByType returns all package managers of a specific type
func (s *UnifiedSysPkg) GetManagersByType(managerType string) map[string]manager.PackageManager {
	return s.registry.GetByType(managerType)
}

// GetBestManager returns the best available package manager for a given type
func (s *UnifiedSysPkg) GetBestManager(managerType string) (manager.PackageManager, error) {
	pm := s.registry.GetBestMatch(managerType)
	if pm == nil {
		return nil, fmt.Errorf("no available package manager found for type '%s'", managerType)
	}
	return pm, nil
}

// ListManagerTypes returns all available manager types
func (s *UnifiedSysPkg) ListManagerTypes() []string {
	managers := s.GetAvailableManagers()
	typeSet := make(map[string]bool)

	for _, pm := range managers {
		typeSet[pm.GetType()] = true
	}

	types := make([]string, 0, len(typeSet))
	for t := range typeSet {
		types = append(types, t)
	}

	return types
}

// === UNIFIED OPERATIONS ===

// Search searches for packages across all available package managers
func (s *UnifiedSysPkg) Search(ctx context.Context, query []string, opts *manager.Options) (map[string][]manager.PackageInfo, error) {
	managers := s.GetAvailableManagers()
	results := make(map[string][]manager.PackageInfo)

	for name, pm := range managers {
		packages, err := pm.Search(ctx, query, opts)
		if err != nil {
			// Log error but continue with other managers
			if opts != nil && opts.Verbose {
				fmt.Printf("Search failed for %s: %v\n", name, err)
			}
			continue
		}
		if len(packages) > 0 {
			results[name] = packages
		}
	}

	return results, nil
}

// SearchByType searches for packages using package managers of a specific type
func (s *UnifiedSysPkg) SearchByType(ctx context.Context, managerType string, query []string, opts *manager.Options) (map[string][]manager.PackageInfo, error) {
	managers := s.GetManagersByType(managerType)
	results := make(map[string][]manager.PackageInfo)

	for name, pm := range managers {
		packages, err := pm.Search(ctx, query, opts)
		if err != nil {
			if opts != nil && opts.Verbose {
				fmt.Printf("Search failed for %s: %v\n", name, err)
			}
			continue
		}
		if len(packages) > 0 {
			results[name] = packages
		}
	}

	return results, nil
}

// Install installs packages using the specified package manager
func (s *UnifiedSysPkg) Install(ctx context.Context, managerName string, packages []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	pm, err := s.GetManagerByName(managerName)
	if err != nil {
		return nil, err
	}

	return pm.Install(ctx, packages, opts)
}

// InstallWithBestManager installs packages using the best available manager for the given type
func (s *UnifiedSysPkg) InstallWithBestManager(ctx context.Context, managerType string, packages []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	pm, err := s.GetBestManager(managerType)
	if err != nil {
		return nil, err
	}

	return pm.Install(ctx, packages, opts)
}

// Remove removes packages using the specified package manager
func (s *UnifiedSysPkg) Remove(ctx context.Context, managerName string, packages []string, opts *manager.Options) ([]manager.PackageInfo, error) {
	pm, err := s.GetManagerByName(managerName)
	if err != nil {
		return nil, err
	}

	return pm.Remove(ctx, packages, opts)
}

// ListInstalled lists all installed packages across all managers
func (s *UnifiedSysPkg) ListInstalled(ctx context.Context, opts *manager.Options) (map[string][]manager.PackageInfo, error) {
	managers := s.GetAvailableManagers()
	results := make(map[string][]manager.PackageInfo)

	for name, pm := range managers {
		packages, err := pm.List(ctx, manager.FilterInstalled, opts)
		if err != nil {
			if opts != nil && opts.Verbose {
				fmt.Printf("List failed for %s: %v\n", name, err)
			}
			continue
		}
		if len(packages) > 0 {
			results[name] = packages
		}
	}

	return results, nil
}

// GetManagerStatus returns status for all available managers
func (s *UnifiedSysPkg) GetManagerStatus(ctx context.Context, opts *manager.Options) (map[string]manager.ManagerStatus, error) {
	managers := s.GetAvailableManagers()
	results := make(map[string]manager.ManagerStatus)

	for name, pm := range managers {
		status, err := pm.Status(ctx, opts)
		if err != nil {
			// Create a basic status indicating the error
			status = manager.ManagerStatus{
				Available: false,
				Healthy:   false,
				Issues:    []string{err.Error()},
				Metadata:  make(map[string]interface{}),
			}
		}
		results[name] = status
	}

	return results, nil
}

// === CONVENIENCE FUNCTIONS ===

// DefaultUnifiedSysPkg provides a default instance with auto-discovery
var DefaultUnifiedSysPkg = NewUnifiedSysPkg()

// GetAvailableManagers returns all available package managers from the default instance
func GetAvailableManagers() map[string]manager.PackageManager {
	return DefaultUnifiedSysPkg.GetAvailableManagers()
}

// GetManagerByName returns a package manager by name from the default instance
func GetManagerByName(name string) (manager.PackageManager, error) {
	return DefaultUnifiedSysPkg.GetManagerByName(name)
}

// SearchAll searches across all package managers using the default instance
func SearchAll(ctx context.Context, query []string, opts *manager.Options) (map[string][]manager.PackageInfo, error) {
	return DefaultUnifiedSysPkg.Search(ctx, query, opts)
}

// ListAllInstalled lists installed packages from all managers using the default instance
func ListAllInstalled(ctx context.Context, opts *manager.Options) (map[string][]manager.PackageInfo, error) {
	return DefaultUnifiedSysPkg.ListInstalled(ctx, opts)
}
