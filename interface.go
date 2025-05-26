package syspkg

import "github.com/bluet/syspkg/manager"

// PackageManager is the interface that defines the methods for interacting with various package managers.
type PackageManager interface {
	// IsAvailable checks if the package manager is available on the current system.
	IsAvailable() bool

	// GetPackageManager returns the name of the package manager.
	GetPackageManager() string

	// Install installs the specified packages using the package manager.
	Install(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error)

	// Delete removes the specified packages using the package manager.
	Delete(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error)

	// Find searches for packages using the specified keywords.
	Find(keywords []string, opts *manager.Options) ([]manager.PackageInfo, error)

	// ListInstalled lists all installed packages.
	ListInstalled(opts *manager.Options) ([]manager.PackageInfo, error)

	// ListUpgradable lists all upgradable packages.
	ListUpgradable(opts *manager.Options) ([]manager.PackageInfo, error)

	// Upgrade upgrades all packages or only the specified ones.
	UpgradeAll(opts *manager.Options) ([]manager.PackageInfo, error)

	// Refresh refreshes the package index.
	Refresh(opts *manager.Options) error

	// GetPackageInfo returns information about the specified package.
	GetPackageInfo(pkg string, opts *manager.Options) (manager.PackageInfo, error)
}

// SysPkg is the interface that defines the methods for interacting with the SysPkg library.
type SysPkg interface {
	// FindPackageManagers returns a map of available package managers based on the specified IncludeOptions.
	// If the AllAvailable option is set to true, all available package managers will be returned.
	// Otherwise, only the specified package managers will be returned.
	// If no suitable package managers are found, an error is returned.
	FindPackageManagers(include IncludeOptions) (map[string]PackageManager, error)

	// RefreshPackageManagers refreshes the internal package manager list based on the specified IncludeOptions, and returns the new list.
	// If the AllAvailable option is set to true, all available package managers will be included.
	// Otherwise, only the specified package managers will be included.
	// If no suitable package managers are found, an error is returned.
	RefreshPackageManagers(include IncludeOptions) (map[string]PackageManager, error)

	// GetPackageManager returns a PackageManager instance based on the specified name, from the list of available package managers specified in the IncludeOptions.
	// If the name is empty, the first available package manager will be returned.
	// If no suitable package manager is found, an error is returned.
	// Note: only package managers that are specified in the IncludeOptions when creating the SysPkg instance (with New() method) will be returned. If you want to use package managers that are not specified in the IncludeOptions, you should use the FindPackageManagers() method to get a list of all available package managers, or use RefreshPackageManagers() with the IncludeOptions parameter to refresh the package manager list.
	GetPackageManager(name string) (PackageManager, error)

	// Install(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error)
	// Delete(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error)
	// Find(keywords []string, opts *manager.Options) ([]manager.PackageInfo, error)
	// ListInstalled(opts *manager.Options) ([]manager.PackageInfo, error)
	// ListUpgradable(opts *manager.Options) ([]manager.PackageInfo, error)
	// Upgrade(opts *manager.Options) ([]manager.PackageInfo, error)
	// Refresh(opts *manager.Options) error
	// GetPackageInfo(pkg string, opts *manager.Options) (manager.PackageInfo, error)
}
