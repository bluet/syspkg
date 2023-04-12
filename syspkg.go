package syspkg

import (
	"errors"

	"github.com/bluet/syspkg/manager/apt"
	// "github.com/bluet/syspkg/snap"
	// "github.com/bluet/syspkg/dnf"
	// "github.com/bluet/syspkg/zypper"
	"github.com/bluet/syspkg/manager"
)

type PackageInfo = manager.PackageInfo
type Options = manager.Options

type PackageManager interface {
	IsAvailable() bool
	Install(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error)
	Delete(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error)
	Find(keywords []string, opts *manager.Options) ([]manager.PackageInfo, error)
	ListInstalled(opts *manager.Options) ([]manager.PackageInfo, error)
	ListUpgradable(opts *manager.Options) ([]manager.PackageInfo, error)
	GetPackageInfo(pkg string, opts *manager.Options) (manager.PackageInfo, error)
	Upgrade(opts *manager.Options) ([]manager.PackageInfo, error)
	Refresh(opts *manager.Options) error
}

func NewPackageManager() ([]PackageManager, error) {
	var pms []PackageManager

	// check if apt is available
	// call apt/apt.go IsAvailable()
	// if yes, return apt/apt.go PackageManager

	aptManager := &apt.PackageManager{}
	if aptManager.IsAvailable() {
		pms = append(pms, aptManager)
	}

	// snapManager := &snap.PackageManager{}
	// if snapManager.IsAvailable() {
	// 	pms = append(pms, snapManager)
	// }

	// dnfManager := &dnf.PackageManager{}
	// if dnfManager.IsAvailable() {
	// 	pms = append(pms, dnfManager)
	// }

	// zypperManager := &zypper.PackageManager{}
	// if zypperManager.IsAvailable() {
	// 	pms = append(pms, zypperManager)
	// }

	if len(pms) == 0 {
		return nil, errors.New("no supported package manager found")
	}

	return pms, nil
}
