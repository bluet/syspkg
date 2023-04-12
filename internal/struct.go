package internal

type PackageStatus string

const (
	PackageStatusInstalled   PackageStatus = "installed"
	PackageStatusUpgradable  PackageStatus = "upgradable"
	PackageStatusAvailable   PackageStatus = "available"
	PackageStatusUnknown     PackageStatus = "unknown"
	PackageStatusConfigFiles PackageStatus = "config-files"
)

type PackageInfo struct {
	Name           string
	Version        string
	NewVersion     string // This field can be empty for installed and available packages.
	Status         PackageStatus
	Category       string
	Arch           string
	PackageManager string
	AdditionalData map[string]string
}

type Options struct {
	Interactive bool
	DryRun      bool
	Verbose     bool
	AssumeYes   bool
	Debug       bool
}
