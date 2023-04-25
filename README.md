# Go-SysPkg

A unified Go library and CLI tool for managing system packages across different package managers (apt, yum, snap, flatpak, and more).

## Features

- Unified package management interface for various package managers
- Supports popular package managers such as APT, Snap, Flatpak, and more
- Easy-to-use API for package installation, removal, search, listing, and system upgrades
- Extendable to support additional package managers

## Getting Started

### Prerequisites

- Go 1.16 or later (1.20+ preferred)

### Installation

Install the library using the `go get` command:

```bash
go get github.com/bluet/syspkg
```

## Usage

Here's an example demonstrating how to use SysPkg as a Go library:
(for real use cases, see the [cmd/syspkg-cli/](cmd/syspkg-cli/) directory

```go
package main

import (
 "fmt"
 "log"
 "os"
 "github.com/bluet/syspkg"
)

func main() {
 // Initialize SysPkg with default package managers
 spkg, err := syspkg.New(syspkg.IncludeOptions{AllAvailable: true})
 if err != nil {
  fmt.Printf("Error while initializing package managers: %+v\n", err)
  os.Exit(1)
 }

 // Find available package managers
 packageManagers, err := spkg.FindPackageManagers(syspkg.IncludeOptions{AllAvailable: true})
 if err != nil {
  fmt.Printf("Error while finding package managers: %+v\n", err)
  os.Exit(1)
 }

 // list all upgradable packages for each package manager
 for pmName, pm := range packageManagers {
  log.Printf("Listing upgradable packages for %s...\n", pmName)
  upgradablePackages, err := pm.ListUpgradable(&manager.Options{})
  if err != nil {
   fmt.Printf("Error while listing upgradable packages for %s: %+v\n", pmName, err)
   continue
  }

  fmt.Printf("Upgradable packages for %s:\n", pmName)
  for _, pkg := range upgradablePackages {
   fmt.Printf("%s: %s %s -> %s (%s)\n", pkg.PackageManager, pkg.Name, pkg.Version, pkg.NewVersion, pkg.Status)
  }
 }
}
```

## Supported Package Managers

| Package Manager | Install | Remove | Search | Upgrade | List Installed | List Upgradable | Get Package Info |
| --------------- | ------- | ------ | ------ | ------- | -------------- | --------------- | ---------------- |
| APT             | ✅      | ✅    | ✅     | ✅     | ✅             | ✅             | ✅               |
| SNAP            | ✅      | ✅    | ✅     | ✅     | ✅             | ✅             | ✅               |
| Flatpak         | ❓      | ❓    | ✅     | ✅     | ✅             | ✅             | ✅               |
| DNF             | ❌      | ❌    | ❌     | ❌     | ❌             | ❌             | ❌               |
| Zypper          | ❌      | ❌    | ❌     | ❌     | ❌             | ❌             | ❌               |
| APK             | ❌      | ❌    | ❌     | ❌     | ❌             | ❌             | ❌               |


Please open an issue (or PR ❤️) if you'd like to see support for any unlisted specific package manager.

### TODO

- [ ] Add support for more package managers
- [ ] Better error handling
- [ ] Better return values and status codes

## Contributing

We welcome contributions to Go-SysPkg! Please read our [CONTRIBUTING.md](CONTRIBUTING.md) for more information on how to contribute.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.