// Package manager provides plugin registration and management
package manager

import (
	"context"
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

// Registry manages registered package manager plugins.
//
// Thread Safety:
// Registry is fully thread-safe for concurrent use. All public methods use
// sync.RWMutex for coordinated access:
//   - Read operations (Get, List, GetAvailable, etc.) use RLock for concurrent reads
//   - Write operations (Register, Unregister, Clear) use Lock for exclusive access
//   - Package managers created by plugins have no shared state and are safe for
//     concurrent use across different goroutines
type Registry struct {
	plugins map[string]Plugin // name -> plugin
	mutex   sync.RWMutex      // protects plugins map for thread-safe access
}

// globalRegistry is the default registry used by the package
var globalRegistry = NewRegistry()

// NewRegistry creates a new plugin registry.
// The returned registry is safe for concurrent use by multiple goroutines.
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
		if manager.GetCategory() == managerType {
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
		if manager.IsAvailable() && manager.GetCategory() == managerType {
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

// === CONCURRENT OPERATIONS ===

// SearchAllConcurrent performs concurrent search across all available package managers.
// This is significantly faster than sequential search when multiple managers are available.
// Returns a map of manager name to search results.
func (r *Registry) SearchAllConcurrent(ctx context.Context, query []string, opts *Options) map[string][]PackageInfo {
	managers := r.GetAvailable()
	if len(managers) == 0 {
		return make(map[string][]PackageInfo)
	}

	type searchResult struct {
		managerName string
		packages    []PackageInfo
		err         error
	}

	results := make(chan searchResult, len(managers))
	var wg sync.WaitGroup

	// Start concurrent searches
	for name, manager := range managers {
		wg.Add(1)
		go func(name string, manager PackageManager) {
			defer wg.Done()
			packages, err := manager.Search(ctx, query, opts)
			if err != nil {
				packages = []PackageInfo{} // Return empty slice on error
			}
			results <- searchResult{name, packages, err}
		}(name, manager)
	}

	// Wait for all searches to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	searchResults := make(map[string][]PackageInfo)
	for result := range results {
		searchResults[result.managerName] = result.packages
	}

	return searchResults
}

// ListInstalledAllConcurrent performs concurrent list of installed packages across all available managers.
// Returns a map of manager name to installed packages.
func (r *Registry) ListInstalledAllConcurrent(ctx context.Context, opts *Options) map[string][]PackageInfo {
	managers := r.GetAvailable()
	if len(managers) == 0 {
		return make(map[string][]PackageInfo)
	}

	type listResult struct {
		managerName string
		packages    []PackageInfo
		err         error
	}

	results := make(chan listResult, len(managers))
	var wg sync.WaitGroup

	// Start concurrent list operations
	for name, manager := range managers {
		wg.Add(1)
		go func(name string, manager PackageManager) {
			defer wg.Done()
			packages, err := manager.List(ctx, FilterInstalled, opts)
			if err != nil {
				packages = []PackageInfo{} // Return empty slice on error
			}
			results <- listResult{name, packages, err}
		}(name, manager)
	}

	// Wait for all operations to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	listResults := make(map[string][]PackageInfo)
	for result := range results {
		listResults[result.managerName] = result.packages
	}

	return listResults
}

// StatusAllConcurrent performs concurrent status checks across all available managers.
// Returns a map of manager name to status information.
func (r *Registry) StatusAllConcurrent(ctx context.Context, opts *Options) map[string]ManagerStatus {
	managers := r.GetAvailable()
	if len(managers) == 0 {
		return make(map[string]ManagerStatus)
	}

	type statusResult struct {
		managerName string
		status      ManagerStatus
		err         error
	}

	results := make(chan statusResult, len(managers))
	var wg sync.WaitGroup

	// Start concurrent status checks
	for name, manager := range managers {
		wg.Add(1)
		go func(name string, manager PackageManager) {
			defer wg.Done()
			status, err := manager.Status(ctx, opts)
			if err != nil {
				// Create a status indicating the manager is unhealthy
				status = ManagerStatus{
					Available: false,
					Healthy:   false,
					Version:   "",
					Issues:    []string{err.Error()},
				}
			}
			results <- statusResult{name, status, err}
		}(name, manager)
	}

	// Wait for all status checks to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	statusResults := make(map[string]ManagerStatus)
	for result := range results {
		statusResults[result.managerName] = result.status
	}

	return statusResults
}

// InstallAllConcurrent performs concurrent package installation across all available managers.
// Returns a map of manager name to installation results.
func (r *Registry) InstallAllConcurrent(ctx context.Context, packages []string, opts *Options) map[string][]PackageInfo {
	managers := r.GetAvailable()
	if len(managers) == 0 {
		return make(map[string][]PackageInfo)
	}

	type installResult struct {
		managerName string
		packages    []PackageInfo
		err         error
	}

	results := make(chan installResult, len(managers))
	var wg sync.WaitGroup

	// Start concurrent install operations
	for name, manager := range managers {
		wg.Add(1)
		go func(name string, manager PackageManager) {
			defer wg.Done()
			installedPackages, err := manager.Install(ctx, packages, opts)
			if err != nil {
				installedPackages = []PackageInfo{} // Return empty slice on error
			}
			results <- installResult{name, installedPackages, err}
		}(name, manager)
	}

	// Wait for all operations to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	installResults := make(map[string][]PackageInfo)
	for result := range results {
		installResults[result.managerName] = result.packages
	}

	return installResults
}

// RemoveAllConcurrent performs concurrent package removal across all available managers.
// Returns a map of manager name to removal results.
func (r *Registry) RemoveAllConcurrent(ctx context.Context, packages []string, opts *Options) map[string][]PackageInfo {
	managers := r.GetAvailable()
	if len(managers) == 0 {
		return make(map[string][]PackageInfo)
	}

	type removeResult struct {
		managerName string
		packages    []PackageInfo
		err         error
	}

	results := make(chan removeResult, len(managers))
	var wg sync.WaitGroup

	// Start concurrent remove operations
	for name, manager := range managers {
		wg.Add(1)
		go func(name string, manager PackageManager) {
			defer wg.Done()
			removedPackages, err := manager.Remove(ctx, packages, opts)
			if err != nil {
				removedPackages = []PackageInfo{} // Return empty slice on error
			}
			results <- removeResult{name, removedPackages, err}
		}(name, manager)
	}

	// Wait for all operations to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	removeResults := make(map[string][]PackageInfo)
	for result := range results {
		removeResults[result.managerName] = result.packages
	}

	return removeResults
}

// VerifyAllConcurrent performs concurrent package verification across all available managers.
// Returns a map of manager name to verification results.
func (r *Registry) VerifyAllConcurrent(ctx context.Context, packages []string, opts *Options) map[string][]PackageInfo {
	managers := r.GetAvailable()
	if len(managers) == 0 {
		return make(map[string][]PackageInfo)
	}

	type verifyResult struct {
		managerName string
		packages    []PackageInfo
		err         error
	}

	results := make(chan verifyResult, len(managers))
	var wg sync.WaitGroup

	// Start concurrent verify operations
	for name, manager := range managers {
		wg.Add(1)
		go func(name string, manager PackageManager) {
			defer wg.Done()
			verifiedPackages, err := manager.Verify(ctx, packages, opts)
			if err != nil {
				verifiedPackages = []PackageInfo{} // Return empty slice on error
			}
			results <- verifyResult{name, verifiedPackages, err}
		}(name, manager)
	}

	// Wait for all operations to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	verifyResults := make(map[string][]PackageInfo)
	for result := range results {
		verifyResults[result.managerName] = result.packages
	}

	return verifyResults
}

// RefreshAllConcurrent performs concurrent package list refresh across all available managers.
// Returns a map of manager name to success status.
func (r *Registry) RefreshAllConcurrent(ctx context.Context, opts *Options) map[string]error {
	managers := r.GetAvailable()
	if len(managers) == 0 {
		return make(map[string]error)
	}

	type refreshResult struct {
		managerName string
		err         error
	}

	results := make(chan refreshResult, len(managers))
	var wg sync.WaitGroup

	// Start concurrent refresh operations
	for name, manager := range managers {
		wg.Add(1)
		go func(name string, manager PackageManager) {
			defer wg.Done()
			err := manager.Refresh(ctx, opts)
			results <- refreshResult{name, err}
		}(name, manager)
	}

	// Wait for all operations to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	refreshResults := make(map[string]error)
	for result := range results {
		refreshResults[result.managerName] = result.err
	}

	return refreshResults
}

// UpgradeAllConcurrentWithErrors performs concurrent package upgrades across all available managers.
// Returns a map of manager name to operation results (including errors).
func (r *Registry) UpgradeAllConcurrentWithErrors(ctx context.Context, packages []string, opts *Options) map[string]OperationResult {
	managers := r.GetAvailable()
	if len(managers) == 0 {
		return make(map[string]OperationResult)
	}

	type upgradeResult struct {
		managerName string
		result      OperationResult
	}

	results := make(chan upgradeResult, len(managers))
	var wg sync.WaitGroup

	// Start concurrent upgrade operations
	for name, manager := range managers {
		wg.Add(1)
		go func(name string, manager PackageManager) {
			defer wg.Done()
			packages, err := manager.Upgrade(ctx, packages, opts)
			result := OperationResult{
				Packages: packages,
				Error:    err,
			}
			// Don't set empty packages on error - preserve the actual result
			results <- upgradeResult{name, result}
		}(name, manager)
	}

	// Wait for all operations to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	upgradeResults := make(map[string]OperationResult)
	for result := range results {
		upgradeResults[result.managerName] = result.result
	}

	return upgradeResults
}

// UpgradeAllConcurrent performs concurrent package upgrades across all available managers.
// Returns a map of manager name to upgrade results.
// DEPRECATED: Use UpgradeAllConcurrentWithErrors for proper error handling.
func (r *Registry) UpgradeAllConcurrent(ctx context.Context, packages []string, opts *Options) map[string][]PackageInfo {
	managers := r.GetAvailable()
	if len(managers) == 0 {
		return make(map[string][]PackageInfo)
	}

	type upgradeResult struct {
		managerName string
		packages    []PackageInfo
		err         error
	}

	results := make(chan upgradeResult, len(managers))
	var wg sync.WaitGroup

	// Start concurrent upgrade operations
	for name, manager := range managers {
		wg.Add(1)
		go func(name string, manager PackageManager) {
			defer wg.Done()
			upgradedPackages, err := manager.Upgrade(ctx, packages, opts)
			if err != nil {
				upgradedPackages = []PackageInfo{} // Return empty slice on error
			}
			results <- upgradeResult{name, upgradedPackages, err}
		}(name, manager)
	}

	// Wait for all operations to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	upgradeResults := make(map[string][]PackageInfo)
	for result := range results {
		upgradeResults[result.managerName] = result.packages
	}

	return upgradeResults
}

// CleanAllConcurrent performs concurrent cache cleaning across all available managers.
// Returns a map of manager name to success status.
func (r *Registry) CleanAllConcurrent(ctx context.Context, opts *Options) map[string]error {
	managers := r.GetAvailable()
	if len(managers) == 0 {
		return make(map[string]error)
	}

	type cleanResult struct {
		managerName string
		err         error
	}

	results := make(chan cleanResult, len(managers))
	var wg sync.WaitGroup

	// Start concurrent clean operations
	for name, manager := range managers {
		wg.Add(1)
		go func(name string, manager PackageManager) {
			defer wg.Done()
			err := manager.Clean(ctx, opts)
			results <- cleanResult{name, err}
		}(name, manager)
	}

	// Wait for all operations to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	cleanResults := make(map[string]error)
	for result := range results {
		cleanResults[result.managerName] = result.err
	}

	return cleanResults
}

// AutoRemoveAllConcurrent performs concurrent orphaned package removal across all available managers.
// Returns a map of manager name to removal results.
func (r *Registry) AutoRemoveAllConcurrent(ctx context.Context, opts *Options) map[string][]PackageInfo {
	managers := r.GetAvailable()
	if len(managers) == 0 {
		return make(map[string][]PackageInfo)
	}

	type autoRemoveResult struct {
		managerName string
		packages    []PackageInfo
		err         error
	}

	results := make(chan autoRemoveResult, len(managers))
	var wg sync.WaitGroup

	// Start concurrent autoremove operations
	for name, manager := range managers {
		wg.Add(1)
		go func(name string, manager PackageManager) {
			defer wg.Done()
			removedPackages, err := manager.AutoRemove(ctx, opts)
			if err != nil {
				removedPackages = []PackageInfo{} // Return empty slice on error
			}
			results <- autoRemoveResult{name, removedPackages, err}
		}(name, manager)
	}

	// Wait for all operations to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	autoRemoveResults := make(map[string][]PackageInfo)
	for result := range results {
		autoRemoveResults[result.managerName] = result.packages
	}

	return autoRemoveResults
}

// === GLOBAL CONCURRENT FUNCTIONS ===

// SearchAllManagersConcurrent performs concurrent search across all available managers in the global registry
func SearchAllManagersConcurrent(ctx context.Context, query []string, opts *Options) map[string][]PackageInfo {
	return globalRegistry.SearchAllConcurrent(ctx, query, opts)
}

// ListInstalledAllManagersConcurrent lists installed packages across all available managers concurrently
func ListInstalledAllManagersConcurrent(ctx context.Context, opts *Options) map[string][]PackageInfo {
	return globalRegistry.ListInstalledAllConcurrent(ctx, opts)
}

// StatusAllManagersConcurrent checks status of all available managers concurrently
func StatusAllManagersConcurrent(ctx context.Context, opts *Options) map[string]ManagerStatus {
	return globalRegistry.StatusAllConcurrent(ctx, opts)
}

// InstallAllManagersConcurrent installs packages across all available managers concurrently
func InstallAllManagersConcurrent(ctx context.Context, packages []string, opts *Options) map[string][]PackageInfo {
	return globalRegistry.InstallAllConcurrent(ctx, packages, opts)
}

// RemoveAllManagersConcurrent removes packages across all available managers concurrently
func RemoveAllManagersConcurrent(ctx context.Context, packages []string, opts *Options) map[string][]PackageInfo {
	return globalRegistry.RemoveAllConcurrent(ctx, packages, opts)
}

// VerifyAllManagersConcurrent verifies packages across all available managers concurrently
func VerifyAllManagersConcurrent(ctx context.Context, packages []string, opts *Options) map[string][]PackageInfo {
	return globalRegistry.VerifyAllConcurrent(ctx, packages, opts)
}

// RefreshAllManagersConcurrent refreshes package lists across all available managers concurrently
func RefreshAllManagersConcurrent(ctx context.Context, opts *Options) map[string]error {
	return globalRegistry.RefreshAllConcurrent(ctx, opts)
}

// UpgradeAllManagersConcurrent upgrades packages across all available managers concurrently
func UpgradeAllManagersConcurrent(ctx context.Context, packages []string, opts *Options) map[string][]PackageInfo {
	return globalRegistry.UpgradeAllConcurrent(ctx, packages, opts)
}

// CleanAllManagersConcurrent cleans caches across all available managers concurrently
func CleanAllManagersConcurrent(ctx context.Context, opts *Options) map[string]error {
	return globalRegistry.CleanAllConcurrent(ctx, opts)
}

// AutoRemoveAllManagersConcurrent removes orphaned packages across all available managers concurrently
func AutoRemoveAllManagersConcurrent(ctx context.Context, opts *Options) map[string][]PackageInfo {
	return globalRegistry.AutoRemoveAllConcurrent(ctx, opts)
}

// Manager types are now defined in interfaces.go
