package syspkg

import (
	"errors"
	"log"

	"github.com/bluet/syspkg/manager"
	"github.com/bluet/syspkg/manager/apt"
	"github.com/bluet/syspkg/manager/flatpak"
	"github.com/bluet/syspkg/manager/snap"
	// "github.com/bluet/syspkg/zypper"
	// "github.com/bluet/syspkg/dnf"
	// "github.com/bluet/syspkg/apk"
)

type PackageInfo = manager.PackageInfo

type IncludeOptions struct {
	AllAvailable bool
	Apk          bool
	Apt          bool
	Dnf          bool
	Flatpak      bool
	Snap         bool
	Zypper       bool
}

type sysPkgImpl struct {
	pms map[string]PackageManager
}

// make sure sysPkgImpl implements SysPkg
var _ SysPkg = (*sysPkgImpl)(nil)

func New(include IncludeOptions) (SysPkg, error) {
	impl := &sysPkgImpl{}
	pms, err := impl.FindPackageManagers(include)
	if err != nil {
		return nil, err
	}

	return &sysPkgImpl{
		pms: pms,
	}, nil
}

func (s *sysPkgImpl) FindPackageManagers(include IncludeOptions) (map[string]PackageManager, error) {
	var pms = make(map[string]PackageManager)
	managerList := []struct {
		managerName string
		manager     PackageManager
		include     bool
	}{
		{"apt", &apt.PackageManager{}, include.Apt},
		{"flatpak", &flatpak.PackageManager{}, include.Flatpak},
		{"snap", &snap.PackageManager{}, include.Snap},
		// {"apk", &apk.PackageManager{}, include.Apk},
		// {"dnf", &dnf.PackageManager{}, include.Dnf},
		// {"zypper", &zypper.PackageManager{}, include.Zypper},
	}

	for _, m := range managerList {
		if include.AllAvailable || m.include {
			if m.manager.IsAvailable() {
				pms[m.managerName] = m.manager
				log.Printf("%s manager is available", m.managerName)
			}
		}
	}

	if len(pms) == 0 {
		return nil, errors.New("no supported package manager found")
	}

	return pms, nil
}

func (s *sysPkgImpl) GetPackageManager(name string) (PackageManager) {
	return s.pms[name]
}
