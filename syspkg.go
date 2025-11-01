// Package syspkg provides a unified interface for interacting with multiple package management systems.
// It allows you to query, install, and remove packages, and supports package managers like Apt, Snap, and Flatpak.
//
// To get started, create a new SysPkg instance by calling the New() function with the desired IncludeOptions.
// After obtaining a SysPkg instance, you can use the FindPackageManagers() function to find available package managers
// on the system, and GetPackageManager() to get a specific package manager.
//
// Example:
//
//	includeOptions := syspkg.IncludeOptions{
//	    AllAvailable: true,
//	}
//	sysPkg, err := syspkg.New(includeOptions)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	aptManager, err := sysPkg.GetPackageManager("apt")
package syspkg

import (
	"errors"
	"log"
	"sort"

	"github.com/bluet/syspkg/manager"
	"github.com/bluet/syspkg/manager/apt"
	"github.com/bluet/syspkg/manager/flatpak"
	"github.com/bluet/syspkg/manager/snap"
	"github.com/bluet/syspkg/manager/yum"
	// "github.com/bluet/syspkg/zypper"
	// "github.com/bluet/syspkg/dnf"
	// "github.com/bluet/syspkg/apk"
)

// PackageInfo represents a package's information.
type PackageInfo = manager.PackageInfo

// IncludeOptions specifies which package managers to include when creating a SysPkg instance.
type IncludeOptions struct {
	AllAvailable bool
	Apk          bool
	Apt          bool
	AptFast      bool
	Dnf          bool
	Flatpak      bool
	Snap         bool
	Yum          bool
	Zypper       bool
}

type sysPkgImpl struct {
	pms map[string]PackageManager
}

// make sure sysPkgImpl implements SysPkg
var _ SysPkg = (*sysPkgImpl)(nil)

// New creates a new SysPkg instance with the specified IncludeOptions.
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

// FindPackageManagers returns a map of available package managers based on the specified IncludeOptions.
func (s *sysPkgImpl) FindPackageManagers(include IncludeOptions) (map[string]PackageManager, error) {
	var pms = make(map[string]PackageManager)
	managerList := []struct {
		managerName string
		manager     PackageManager
		include     bool
	}{
		{"apt", apt.NewPackageManager(), include.Apt},
		{"apt-fast", apt.NewPackageManagerWithBinary("apt-fast"), include.AptFast},
		{"flatpak", &flatpak.PackageManager{}, include.Flatpak},
		{"snap", &snap.PackageManager{}, include.Snap},
		{"yum", yum.NewPackageManager(), include.Yum},
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

// GetPackageManager returns a PackageManager instance by its name (e.g., "apt", "snap", "flatpak", etc.).
// if name is empty, return the first available
func (s *sysPkgImpl) GetPackageManager(name string) (PackageManager, error) {
	// if there are no package managers, return before accessing non existing properties
	if len(s.pms) == 0 {
		return nil, errors.New("no supported package manager detected")
	}

	if name == "" {
		// get first pm available, lexicographically sorted
		keys := make([]string, 0, len(s.pms))
		for k := range s.pms {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		return s.pms[keys[0]], nil
	}

	pm, found := s.pms[name]
	if !found {
		return nil, errors.New("no such package manager")
	}
	return pm, nil
}

// RefreshPackageManagers refreshes the internal list of available package managers, and returns the new list.
func (s *sysPkgImpl) RefreshPackageManagers(include IncludeOptions) (map[string]PackageManager, error) {
	pms, err := s.FindPackageManagers(include)
	if err != nil {
		return nil, err
	}

	s.pms = pms
	return pms, nil
}
