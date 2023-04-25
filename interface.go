package syspkg

import "github.com/bluet/syspkg/manager"


type PackageManager interface {
	IsAvailable() bool
	GetPackageManager() string
	Install(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error)
	Delete(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error)
	Find(keywords []string, opts *manager.Options) ([]manager.PackageInfo, error)
	ListInstalled(opts *manager.Options) ([]manager.PackageInfo, error)
	ListUpgradable(opts *manager.Options) ([]manager.PackageInfo, error)
	Upgrade(opts *manager.Options) ([]manager.PackageInfo, error)
	Refresh(opts *manager.Options) error
	GetPackageInfo(pkg string, opts *manager.Options) (manager.PackageInfo, error)
}

type SysPkg interface {
	FindPackageManagers(include IncludeOptions) (map[string]PackageManager, error)
	GetPackageManager(name string) (PackageManager)
	// IsAvailable() bool
	// GetPackageManager() string
	// Install(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error)
	// Delete(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error)
	// Find(keywords []string, opts *manager.Options) ([]manager.PackageInfo, error)
	// ListInstalled(opts *manager.Options) ([]manager.PackageInfo, error)
	// ListUpgradable(opts *manager.Options) ([]manager.PackageInfo, error)
	// Upgrade(opts *manager.Options) ([]manager.PackageInfo, error)
	// Refresh(opts *manager.Options) error
	// GetPackageInfo(pkg string, opts *manager.Options) (manager.PackageInfo, error)
}
