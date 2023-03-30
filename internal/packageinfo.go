package internal

type PackageStatus string

const (
	Installed  PackageStatus = "installed"
	Upgradable PackageStatus = "upgradable"
	Available  PackageStatus = "available"
)

type PackageInfo struct {
	Name           string
	Version        string
	NewVersion     string // This field can be empty for installed and available packages.
	Status         PackageStatus
	Category       string
	Arch           string
	PackageManager string
}
