package syspkg

import (
	"errors"

	"github.com/bluet/syspkg/apt"
	"github.com/bluet/syspkg/snap"
	"github.com/bluet/syspkg/yum"

	"github.com/bluet/syspkg/internal"
)

type PackageManager interface {
	Install(pkg string) error
	Uninstall(pkg string) error
	Search(pkg string) ([]internal.PackageInfo, error)
	ListInstalled() ([]internal.PackageInfo, error)
	ListUpgradable() ([]internal.PackageInfo, error)
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

	yumManager := &yum.PackageManager{}
	if yumManager.IsAvailable() {
		pms = append(pms, yumManager)
	}

	snapManager := &snap.PackageManager{}
	if snapManager.IsAvailable() {
		pms = append(pms, snapManager)
	}

	if len(pms) == 0 {
		return nil, errors.New("no supported package manager found")
	}

	return pms, nil
}
