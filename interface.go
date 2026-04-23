package syspkg

import "github.com/bluet/syspkg/manager"

// PackageManager is the interface that defines the methods for interacting with various package managers.
type PackageManager interface {
	// IsAvailable checks if the package manager is available on the current system.
	IsAvailable() bool

	// GetPackageManager returns the name of the package manager.
	GetPackageManager() string

	// Install installs the specified packages using the package manager.
	// Returns PackageInfo for each successfully installed package with Status=installed.
	// Version and NewVersion fields will contain the installed version.
	Install(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error)

	// Delete removes the specified packages using the package manager.
	// Returns PackageInfo for each successfully removed package with Status=available.
	// Version field contains the removed version, NewVersion will be empty.
	Delete(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error)

	// Find searches for packages using the specified keywords and checks their installation status.
	// Cross-package manager status normalization ensures consistent behavior:
	//   - Status=installed: Package is currently installed
	//   - Status=available: Package exists in repositories but is not installed
	//     (includes previously installed packages that have been removed, even with config files remaining)
	//   - Status=upgradable: Package is installed but newer version is available
	//
	// Version field contains installed version (empty if not installed).
	// NewVersion field contains available version from repositories.
	//
	// Implementation notes:
	//   - APT: Uses dpkg-query to check installation status
	//   - YUM: Uses rpm -q to check installation status
	//   - APT config-files state is normalized to available for cross-PM compatibility
	//   - Both package managers provide accurate status detection
	Find(keywords []string, opts *manager.Options) ([]manager.PackageInfo, error)

	// ListInstalled lists all currently installed packages.
	// Returns packages with Status=installed, Version set to installed version, NewVersion empty.
	ListInstalled(opts *manager.Options) ([]manager.PackageInfo, error)

	// ListUpgradable lists all packages that have newer versions available.
	// Returns packages with Status=upgradable, Version=current, NewVersion=available.
	ListUpgradable(opts *manager.Options) ([]manager.PackageInfo, error)

	// Upgrade upgrades the specified packages to their latest versions.
	// Returns PackageInfo for each upgraded package with new version information.
	Upgrade(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error)

	// UpgradeAll upgrades all packages or only the specified ones.
	// Returns PackageInfo for each upgraded package with new version information.
	UpgradeAll(opts *manager.Options) ([]manager.PackageInfo, error)

	// Refresh refreshes the package index/repositories.
	// This should be called before search operations to ensure up-to-date package information.
	Refresh(opts *manager.Options) error

	// GetPackageInfo returns detailed information about the specified package.
	// Returns package metadata including name, version, architecture, and category.
	GetPackageInfo(pkg string, opts *manager.Options) (manager.PackageInfo, error)

	// Clean performs cleanup of package manager caches and temporary files.
	// The specific behavior depends on the package manager implementation.
	Clean(opts *manager.Options) error

	// AutoRemove removes packages that were automatically installed as dependencies
	// but are no longer needed by any manually installed packages.
	// Returns PackageInfo for each removed package with Status=available.
	AutoRemove(opts *manager.Options) ([]manager.PackageInfo, error)
}

// ManagerCreationOptions specifies options for creating a package manager instance.
type ManagerCreationOptions struct {
	// BinaryPath specifies a custom binary path or name to use for the package manager.
	// This can be either just a binary name (e.g., "apt-fast") which will be searched in PATH,
	// or a full path (e.g., "/usr/local/bin/apt-fast").
	// If empty, the default binary for the package manager will be used.
	BinaryPath string
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
	// For custom binary paths, use GetPackageManagerWithOptions.
	GetPackageManager(name string) (PackageManager, error)

	// GetPackageManagerWithOptions returns a PackageManager instance with custom configuration options.
	// This allows specifying custom binary paths (e.g., "apt-fast" instead of "apt") and other creation-time options.
	// The returned manager instance will use the specified options for all operations.
	//
	// Example usage:
	//   // Use apt-fast instead of apt
	//   apt, _ := syspkg.GetPackageManagerWithOptions("apt", &ManagerCreationOptions{
	//       BinaryPath: "apt-fast",
	//   })
	//
	//   // Use custom yum path
	//   yum, _ := syspkg.GetPackageManagerWithOptions("yum", &ManagerCreationOptions{
	//       BinaryPath: "/opt/custom/yum",
	//   })
	GetPackageManagerWithOptions(name string, opts *ManagerCreationOptions) (PackageManager, error)

	// Install(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error)
	// Delete(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error)
	// Find(keywords []string, opts *manager.Options) ([]manager.PackageInfo, error)
	// ListInstalled(opts *manager.Options) ([]manager.PackageInfo, error)
	// ListUpgradable(opts *manager.Options) ([]manager.PackageInfo, error)
	// Upgrade(opts *manager.Options) ([]manager.PackageInfo, error)
	// Refresh(opts *manager.Options) error
	// GetPackageInfo(pkg string, opts *manager.Options) (manager.PackageInfo, error)
}
