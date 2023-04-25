// Package manager provides utilities for managing the application.
package manager

// PackageStatus represents the current status of a package in the system.
type PackageStatus string

// PackageStatus constants define possible statuses for packages.
const (
	// PackageStatusInstalled represents an installed package.
	PackageStatusInstalled PackageStatus = "installed"

	// PackageStatusUpgradable represents a package with a newer version available for upgrade.
	PackageStatusUpgradable PackageStatus = "upgradable"

	// PackageStatusAvailable represents a package that is available but not yet installed.
	// Note: In some cases, installed packages may also be marked as available.
	PackageStatusAvailable PackageStatus = "available"

	// PackageStatusUnknown represents a package with an unknown status.
	PackageStatusUnknown PackageStatus = "unknown"

	// PackageStatusConfigFiles represents a package that has only configuration files remaining on the system.
	PackageStatusConfigFiles PackageStatus = "config-files"
)

// PackageInfo contains information about a specific package.
type PackageInfo struct {
	// Name is the package name.
	Name string

	// Version is the currently installed version of the package.
	Version string

	// NewVersion is the latest available version of the package. This field can be empty for installed and available packages.
	NewVersion string

	// Status indicates the current PackageStatus of the package.
	Status PackageStatus

	// Category is the category the package belongs to, such as "utilities" or "development".
	Category string

	// Arch is the architecture the package is built for, such as "amd64" or "arm64".
	Arch string

	// PackageManager is the name of the package manager used to manage this package, such as "apt" or "yum".
	PackageManager string

	// AdditionalData is a map of key-value pairs that store any additional package-specific data.
	AdditionalData map[string]string
}
