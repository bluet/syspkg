// Package manager provides plugin registration and management
package manager

import (
	"fmt"
	"sort"
	"sync"
)

// Plugin represents a package manager plugin that can be registered
type Plugin interface {
	// CreateManager creates a new instance of the package manager
	CreateManager() PackageManager

	// GetPriority returns the priority for this plugin (higher = preferred)
	// Used when multiple plugins can handle the same type
	GetPriority() int
}

// Registry manages registered package manager plugins
type Registry struct {
	plugins map[string]Plugin // name -> plugin
	mutex   sync.RWMutex
}

// globalRegistry is the default registry used by the package
var globalRegistry = NewRegistry()

// NewRegistry creates a new plugin registry
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]Plugin),
	}
}

// Register registers a package manager plugin
func (r *Registry) Register(name string, plugin Plugin) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	if plugin == nil {
		return fmt.Errorf("plugin cannot be nil")
	}

	// Check if a plugin with this name already exists
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin '%s' is already registered", name)
	}

	r.plugins[name] = plugin
	return nil
}

// Unregister removes a plugin from the registry
func (r *Registry) Unregister(name string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	delete(r.plugins, name)
}

// Get returns a plugin by name
func (r *Registry) Get(name string) (Plugin, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	plugin, exists := r.plugins[name]
	return plugin, exists
}

// List returns all registered plugin names, sorted alphabetically
func (r *Registry) List() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	names := make([]string, 0, len(r.plugins))
	for name := range r.plugins {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// GetAvailable returns all available package managers (where IsAvailable() returns true)
func (r *Registry) GetAvailable() map[string]PackageManager {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	available := make(map[string]PackageManager)
	for name, plugin := range r.plugins {
		manager := plugin.CreateManager()
		if manager.IsAvailable() {
			available[name] = manager
		}
	}
	return available
}

// GetByType returns all available package managers of a specific type
func (r *Registry) GetByType(managerType string) map[string]PackageManager {
	available := r.GetAvailable()
	result := make(map[string]PackageManager)

	for name, manager := range available {
		if manager.GetType() == managerType {
			result[name] = manager
		}
	}
	return result
}

// GetBestMatch returns the best available package manager for a given type
// based on plugin priority. Returns nil if no suitable manager is found.
func (r *Registry) GetBestMatch(managerType string) PackageManager {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var bestManager PackageManager
	var bestPriority int = -1

	for _, plugin := range r.plugins {
		manager := plugin.CreateManager()
		if manager.IsAvailable() && manager.GetType() == managerType {
			priority := plugin.GetPriority()
			if priority > bestPriority {
				bestPriority = priority
				bestManager = manager
			}
		}
	}

	return bestManager
}

// Count returns the number of registered plugins
func (r *Registry) Count() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return len(r.plugins)
}

// Clear removes all registered plugins
func (r *Registry) Clear() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.plugins = make(map[string]Plugin)
}

// === Global registry functions for convenience ===

// Register registers a plugin in the global registry
func Register(name string, plugin Plugin) error {
	return globalRegistry.Register(name, plugin)
}

// Unregister removes a plugin from the global registry
func Unregister(name string) {
	globalRegistry.Unregister(name)
}

// GetPlugin returns a plugin by name from the global registry
func GetPlugin(name string) (Plugin, bool) {
	return globalRegistry.Get(name)
}

// ListPlugins returns all registered plugin names from the global registry
func ListPlugins() []string {
	return globalRegistry.List()
}

// GetAvailableManagers returns all available package managers from the global registry
func GetAvailableManagers() map[string]PackageManager {
	return globalRegistry.GetAvailable()
}

// GetManagersByType returns all available package managers of a specific type
func GetManagersByType(managerType string) map[string]PackageManager {
	return globalRegistry.GetByType(managerType)
}

// GetBestManager returns the best available package manager for a given type
func GetBestManager(managerType string) PackageManager {
	return globalRegistry.GetBestMatch(managerType)
}

// GetGlobalRegistry returns the global registry (useful for testing)
func GetGlobalRegistry() *Registry {
	return globalRegistry
}

// === Common manager types ===
const (
	TypeSystem     = "system"     // OS package managers (apt, yum, etc.)
	TypeLanguage   = "language"   // Language-specific (npm, pip, cargo, etc.)
	TypeVersion    = "version"    // Version managers (nvm, asdf, pyenv, etc.)
	TypeContainer  = "container"  // Container management (docker, podman, etc.)
	TypeGame       = "game"       // Game managers (steam, lutris, etc.)
	TypeScientific = "scientific" // Scientific computing (conda, mamba, etc.)
	TypeBuild      = "build"      // Build tools (vcpkg, conan, etc.)
	TypeApp        = "app"        // Application stores (flatpak, snap, etc.)
)
