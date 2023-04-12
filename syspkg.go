package syspkg

import (
	"errors"

	"github.com/bluet/syspkg/apt"
	// "github.com/bluet/syspkg/snap"
	// "github.com/bluet/syspkg/dnf"
	// "github.com/bluet/syspkg/zypper"
	"github.com/bluet/syspkg/internal"
)

type PackageInfo = internal.PackageInfo
type Options = internal.Options

type PackageManager interface {
	IsAvailable() bool
	Install(pkgs []string, opts *internal.Options) ([]internal.PackageInfo, error)
	Delete(pkgs []string, opts *internal.Options) ([]internal.PackageInfo, error)
	Find(keywords []string, opts *internal.Options) ([]internal.PackageInfo, error)
	ListInstalled(opts *internal.Options) ([]internal.PackageInfo, error)
	ListUpgradable(opts *internal.Options) ([]internal.PackageInfo, error)
	GetPackageInfo(pkg string, opts *internal.Options) (internal.PackageInfo, error)
	Upgrade(opts *internal.Options) ([]internal.PackageInfo, error)
	Refresh(opts *internal.Options) error
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
