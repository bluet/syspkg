# Go-SysPkg

A unified Go library for managing system packages across different package managers (APT, YUM, SNAP, and more).

## Features

- Unified package management interface
- Supports popular package managers such as APT, YUM, and SNAP
- Easy-to-use API for package installation, removal, search, listing, and system upgrades

## Getting Started

### Prerequisites

- Go 1.16 or later

### Installation

Install the library using `go get`:

```bash
go get github.com/bluet/go-syspkg
```

## Usage

Here's an example demonstrating how to use Go-SysPkg:

```go
package main

import (
 "fmt"
 "log"

 "github.com/bluet/go-syspkg/pkg/syspkg"
)

func main() {
 manager, err := syspkg.NewPackageManager()
 if err != nil {
  log.Fatalf("Error initializing package manager: %v", err)
 }

 // List installed packages
 installedPackages, err := manager.ListInstalled()
 if err != nil {
  log.Fatalf("Error listing installed packages: %v", err)
 }

 fmt.Println("Installed packages:")
 for _, pkg := range installedPackages {
  fmt.Printf("- %s (%s)\n", pkg.Name, pkg.Version)
 }

 // List upgradable packages
 upgradablePackages, err := manager.ListUpgradable()
 if err != nil {
  log.Fatalf("Error listing upgradable packages: %v", err)
 }

 fmt.Println("Upgradable packages:")
 for _, pkg := range upgradablePackages {
  fmt.Printf("- %s (%s -> %s)\n", pkg.Name, pkg.Version, pkg.NewVersion)
 }

 // Upgrade system packages
 err = manager.UpgradeSystem(true)
 if err != nil {
  log.Fatalf("Error upgrading system packages: %v", err)
 }
}

```

## Status
- apt: Supported
- snap: Partially supported
- dnf: Partially supported (untested)
- zypper: Partially supported (untested)
- others: Please open an issue if you'd like to see support for a specific package manager

| Package Manager | Install | Remove | Search | List | Upgrade |
| --------------- | ------- | ------ | ------ | ---- | ------- |
| APT             | ✅      | ✅     | ✅     | ✅   | ✅      |
| SNAP            | ✅      | ✅     | ✅     | ✅   | ✅      |
| DNF             | ❓      | ❓   | ❓     | ❓   | ❓      |
| Zypper          | ❓      | ❓   | ❓     | ❓   | ❓      |

### TODO
- [ ] Add support for more package managers
- [ ] Better error handling
- [ ] Better return values and status codes


## Contributing
We welcome contributions to Go-SysPkg! Please read our CONTRIBUTING.md for more information on how to contribute.

## License
This project is licensed under the MIT License. See the LICENSE file for details.
