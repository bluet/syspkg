package syspkg

import (
	"errors"

	"github.com/bluet/syspkg/manager"
	"github.com/bluet/syspkg/manager/apt"
	"github.com/bluet/syspkg/manager/flatpak"
	// "github.com/bluet/syspkg/snap"
	// "github.com/bluet/syspkg/dnf"
	// "github.com/bluet/syspkg/zypper"
)

type PackageInfo = manager.PackageInfo
type Options = manager.Options

type PackageManager interface {
	IsAvailable() bool
	GetPackageManager() string
	Install(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error)
	Delete(pkgs []string, opts *manager.Options) ([]manager.PackageInfo, error)
	Find(keywords []string, opts *manager.Options) ([]manager.PackageInfo, error)
	// ListInstalled(opts *manager.Options) ([]manager.PackageInfo, error)
	ListUpgradable(opts *manager.Options) ([]manager.PackageInfo, error)
	Upgrade(opts *manager.Options) ([]manager.PackageInfo, error)
	Refresh(opts *manager.Options) error
	GetPackageInfo(pkg string, opts *manager.Options) (manager.PackageInfo, error)
}

func NewPackageManager(wants []string) (map[string]PackageManager, error) {
	var pms map[string]PackageManager = make(map[string]PackageManager)

	// check if apt is available
	// call apt/apt.go IsAvailable()
	// if yes, return apt/apt.go PackageManager

	aptManager := &apt.PackageManager{}
	if aptManager.IsAvailable() {
		pms["apt"] = aptManager
	}

	flatpakManager := &flatpak.PackageManager{}
	if flatpakManager.IsAvailable() {
		pms["flatpak"] = flatpakManager
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

	if len(wants) == 0 {
		return pms, nil
	}

	var ret map[string]PackageManager = make(map[string]PackageManager)
	for _, pm := range pms {
		// for i, want := range wants {
			for _, want := range wants {
			if want == pm.GetPackageManager() {
				ret[want] = pm
				// wants = append(wants[:i], wants[i+1:]...)
			}
		}
	}

	// if len(wants) > 0 {
	// 	return nil, errors.New("unsupported package manager: " + wants[0])
	// }

	return ret, nil
}
