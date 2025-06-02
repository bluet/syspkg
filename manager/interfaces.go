// Package manager provides unified interfaces for all package management systems
package manager

import (
	"context"
	"errors"
	"fmt"
)

// Standard error for unsupported operations
var (
	ErrOperationNotSupported = errors.New("operation not supported by this package manager")
	ErrPackageNotFound       = errors.New("package not found")
	ErrInvalidPackageName    = errors.New("invalid package name")
)

// OperationError provides structured error information for better debugging
type OperationError struct {
	Manager   string   // Package manager name
	Operation string   // Operation being performed
	Packages  []string // Packages involved
	Cause     error    // Underlying error
}

func (e *OperationError) Error() string {
	if len(e.Packages) > 0 {
		return fmt.Sprintf("%s %s failed for packages %v: %v", e.Manager, e.Operation, e.Packages, e.Cause)
	}
	return fmt.Sprintf("%s %s failed: %v", e.Manager, e.Operation, e.Cause)
}

func (e *OperationError) Unwrap() error {
	return e.Cause
}

// PackageManager defines the unified interface that ALL package managers must implement.
// If a manager doesn't support an operation, it should return ErrOperationNotSupported
// with an appropriate message rather than panicking or failing silently.
//
// This ensures users have a consistent experience across all package managers,
// while plugin developers have clear expectations for what to implement.
type PackageManager interface {
	// === BASIC INFORMATION ===

	// GetName returns the human-readable name of the package manager (e.g., "APT", "npm", "Steam")
	GetName() string

	// GetType returns the category type (e.g., "system", "language", "game", "version")
	GetType() string

	// IsAvailable checks if this package manager is available on the current system
	IsAvailable() bool

	// GetVersion returns the version of the package manager itself (if applicable)
	GetVersion() (string, error)

	// === CORE PACKAGE OPERATIONS ===

	// Search finds packages matching the given query terms
	// Returns list of available packages (not necessarily installed)
	Search(ctx context.Context, query []string, opts *Options) ([]PackageInfo, error)

	// List returns packages based on the filter (installed, available, upgradable)
	List(ctx context.Context, filter ListFilter, opts *Options) ([]PackageInfo, error)

	// Install installs the specified packages
	// Returns info about successfully installed packages
	Install(ctx context.Context, packages []string, opts *Options) ([]PackageInfo, error)

	// Remove/uninstall the specified packages
	Remove(ctx context.Context, packages []string, opts *Options) ([]PackageInfo, error)

	// GetInfo returns detailed information about a specific package
	GetInfo(ctx context.Context, packageName string, opts *Options) (PackageInfo, error)

	// === UPDATE/UPGRADE OPERATIONS ===

	// Refresh updates the package database/repository information
	Refresh(ctx context.Context, opts *Options) error

	// Upgrade updates specified packages to their latest versions
	// If packages is empty, may upgrade all packages (manager-dependent)
	Upgrade(ctx context.Context, packages []string, opts *Options) ([]PackageInfo, error)

	// === CLEANUP OPERATIONS ===

	// Clean removes cached files, temporary data, etc.
	Clean(ctx context.Context, opts *Options) error

	// AutoRemove removes orphaned packages/dependencies no longer needed
	AutoRemove(ctx context.Context, opts *Options) ([]PackageInfo, error)

	// === HEALTH/STATUS OPERATIONS ===

	// Verify checks integrity of installed packages (if supported)
	Verify(ctx context.Context, packages []string, opts *Options) ([]PackageInfo, error)

	// Status returns overall status/health of the package manager
	Status(ctx context.Context, opts *Options) (ManagerStatus, error)
}

// ListFilter specifies what types of packages to list
type ListFilter string

const (
	FilterInstalled  ListFilter = "installed"  // Only installed packages
	FilterAvailable  ListFilter = "available"  // Available but not installed
	FilterUpgradable ListFilter = "upgradable" // Installed packages with updates available
	FilterAll        ListFilter = "all"        // All packages (installed + available)
)

// PackageInfo represents information about a package in a unified way
type PackageInfo struct {
	// Core identification
	Name       string `json:"name"`
	Version    string `json:"version"`     // Current/installed version
	NewVersion string `json:"new_version"` // Available version (for upgrades)

	// Status and metadata
	Status      string `json:"status"` // installed, available, upgradable, etc.
	Description string `json:"description"`
	Category    string `json:"category"` // Package category/section

	// Manager-specific data
	ManagerType string                 `json:"manager_type"` // Which manager this came from
	Metadata    map[string]interface{} `json:"metadata"`     // Flexible manager-specific data
}

// ManagerStatus represents the overall health/status of a package manager
type ManagerStatus struct {
	Available      bool                   `json:"available"`       // Is the manager available?
	Healthy        bool                   `json:"healthy"`         // Is it working properly?
	Version        string                 `json:"version"`         // Manager version
	LastRefresh    string                 `json:"last_refresh"`    // When was DB last updated
	CacheSize      int64                  `json:"cache_size"`      // Cache size in bytes
	PackageCount   int                    `json:"package_count"`   // Total packages available
	InstalledCount int                    `json:"installed_count"` // Packages installed
	Issues         []string               `json:"issues"`          // Any problems detected
	Metadata       map[string]interface{} `json:"metadata"`        // Manager-specific status info
}

// Options provides common configuration for all operations
type Options struct {
	// Execution mode
	DryRun      bool `json:"dry_run"`     // Don't actually perform operations
	Interactive bool `json:"interactive"` // Allow interactive prompts
	Verbose     bool `json:"verbose"`     // Show detailed output
	Debug       bool `json:"debug"`       // Show debug information
	Quiet       bool `json:"quiet"`       // Minimal output

	// Authorization
	AssumeYes bool `json:"assume_yes"` // Automatically answer yes to prompts
	NoConfirm bool `json:"no_confirm"` // Skip confirmation prompts

	// Scope and filtering
	GlobalScope bool     `json:"global_scope"` // System-wide vs user-local (where applicable)
	SkipBroken  bool     `json:"skip_broken"`  // Skip packages with problems
	OnlyEnabled bool     `json:"only_enabled"` // Only show enabled packages/repos
	Arch        string   `json:"arch"`         // Target architecture
	Tags        []string `json:"tags"`         // Filter by tags/categories

	// Manager-specific options
	CustomArgs []string               `json:"custom_args"` // Additional command-line arguments
	Metadata   map[string]interface{} `json:"metadata"`    // Manager-specific options

	// Timeout and retries
	TimeoutSecs int `json:"timeout_secs"` // Operation timeout (0 = use default)
	Retries     int `json:"retries"`      // Number of retries on failure
}

// DefaultOptions returns a new Options struct with sensible defaults
func DefaultOptions() *Options {
	return &Options{
		DryRun:      false,
		Interactive: false,
		Verbose:     false,
		Debug:       false,
		Quiet:       false,
		AssumeYes:   false,
		NoConfirm:   false,
		GlobalScope: true,
		SkipBroken:  false,
		OnlyEnabled: true,
		TimeoutSecs: 0, // Use manager's default
		Retries:     0, // No retries by default
		CustomArgs:  []string{},
		Metadata:    make(map[string]interface{}),
		Tags:        []string{},
	}
}

// Standard package status values
const (
	StatusInstalled  = "installed"
	StatusAvailable  = "available"
	StatusUpgradable = "upgradable"
	StatusUnknown    = "unknown"
)

// Standard manager types
const (
	TypeSystem     = "system"     // APT, YUM, DNF, Pacman, etc.
	TypeLanguage   = "language"   // npm, pip, cargo, gem, etc.
	TypeVersion    = "version"    // nvm, asdf, pyenv, rbenv, etc.
	TypeContainer  = "container"  // docker, podman, helm, etc.
	TypeGame       = "game"       // steam, lutris, gog, etc.
	TypeScientific = "scientific" // conda, mamba, bioconda, etc.
	TypeBuild      = "build"      // vcpkg, conan, cmake, etc.
	TypeApp        = "app"        // flatpak, snap, appimage, etc.
)
