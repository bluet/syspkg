// Package manager provides utilities for managing the application.
package manager

// PackageStatus represents the current status of a package in the system.
type PackageStatus string

// PackageStatus constants define possible statuses for packages across all package managers.
// These statuses are normalized for cross-package manager compatibility.
const (
	// PackageStatusInstalled represents a package that is currently installed and functional.
	// Used by: All package managers
	PackageStatusInstalled PackageStatus = "installed"

	// PackageStatusUpgradable represents an installed package that has a newer version available.
	// Used by: All package managers
	PackageStatusUpgradable PackageStatus = "upgradable"

	// PackageStatusAvailable represents a package that exists in repositories but is not installed.
	// This includes packages that were previously installed but removed (including config-files state).
	// For cross-package manager compatibility, APT's config-files state is normalized to this status.
	// Used by: All package managers
	PackageStatusAvailable PackageStatus = "available"

	// PackageStatusUnknown represents a package with an unknown or error state.
	// This is rare and typically indicates system errors or corrupted package databases.
	// Used by: All package managers (rare cases)
	PackageStatusUnknown PackageStatus = "unknown"

	// PackageStatusConfigFiles represents a package with only configuration files remaining.
	// Note: This is deprecated and normalized to PackageStatusAvailable for cross-PM compatibility.
	// Only kept for internal use by APT implementation.
	PackageStatusConfigFiles PackageStatus = "config-files"
)

// PackageInfo contains information about a specific package.
// Field usage varies by operation and package status:
//
// Field Usage by Operation:
//
//	Install:     Version=installed_version, NewVersion=installed_version, Status=installed
//	Delete:      Version=removed_version, NewVersion="", Status=available
//	Find:        Version=installed_version (or ""), NewVersion=repo_version, Status=installed/available/upgradable
//	ListInstalled: Version=installed_version, NewVersion="", Status=installed
//	ListUpgradable: Version=current_version, NewVersion=available_version, Status=upgradable
//	GetPackageInfo: Version=available_version, NewVersion="", Status varies
//
// Field Usage by Status:
//
//	installed:   Version=installed_version, NewVersion=installed_version (Install) or "" (ListInstalled)
//	available:   Version="" (not installed) or removed_version (Delete), NewVersion=repo_version
//	upgradable:  Version=current_version, NewVersion=newer_version
//	unknown:     Version="", NewVersion may contain repo_version
type PackageInfo struct {
	// Name is the package name.
	Name string

	// Version is the currently installed version of the package.
	// Empty if package is not installed (Status=available).
	// Contains removed version for Delete operations.
	Version string

	// NewVersion is the latest available version from repositories.
	// Used for available versions in Find operations and upgrade targets.
	// Empty for ListInstalled operations.
	// Same as Version for Install operations.
	NewVersion string

	// Status indicates the current PackageStatus of the package.
	// See PackageStatus constants for detailed descriptions.
	Status PackageStatus

	// Category is the category/section the package belongs to.
	// Examples: "utilities", "development", "web", "jammy", "main"
	// May represent repository sections or package categories depending on package manager.
	Category string

	// Arch is the architecture the package is built for.
	// Examples: "amd64", "arm64", "i386", "all"
	// Empty if architecture is not specified or not applicable.
	Arch string

	// PackageManager is the name of the package manager used to manage this package.
	// Examples: "apt", "yum", "dnf", "snap", "flatpak"
	PackageManager string

	// AdditionalData is a map of key-value pairs for additional package-specific metadata.
	// Used for package manager specific information that doesn't fit standard fields.
	AdditionalData map[string]string
}
