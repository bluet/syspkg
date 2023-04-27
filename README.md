# SysPkg

[![Go Reference](https://pkg.go.dev/badge/github.com/bluet/syspkg.svg)](https://pkg.go.dev/github.com/bluet/syspkg)
[![Go Report Card](https://goreportcard.com/badge/github.com/bluet/syspkg)](https://goreportcard.com/report/github.com/bluet/syspkg)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)]

SysPkg is a unified CLI tool and Golang library for managing system packages across different package managers (apt, snap, flatpak, yum, dnf, and more). It simplifies the process of working with various package managers by providing a consistent interface and API through an abstraction layer.

## Features

- A unified package management interface for various package managers
- Supports popular package managers such as APT, Snap, Flatpak, and more
- Easy-to-use API for package installation, removal, search, listing, and system upgrades
- Expandable architecture to support more package managers in the future

## API Documentation

See the [Go Reference](https://pkg.go.dev/github.com/bluet/syspkg) for the full API documentation.

## Getting Started

### Prerequisites

- Go 1.16 or later (1.20+ preferred)

### Installation (as CLI tool)

Install the CLI tool using the `go install` command:

```bash
go install github.com/bluet/syspkg/cmd/syspkg@latest
```

### Installation (as Go library)

Install the library using the `go get` command:

```bash
go get github.com/bluet/syspkg
```

## Usage

### CLI Tool

SysPkg provides a unified CLI tool for managing system packages across different package managers. It simplifies the process of working with various package managers by providing a consistent interface through an abstraction layer.

Here's an example demonstrating how to use SysPkg as a CLI tool:

```bash
# Install a package using APT
syspkg --apt install vim

# Remove a package using APT
syspkg --apt remove vim

# Search for a package using Snap
syspkg --snap search vim

# Show all upgradable packages using Flatpak
syspkg --flatpak show upgradable
```

Or, you can do operations without knowing the package manager:

```bash
# Install a package using all available package managers
syspkg install vim

# Remove a package using all available package manager
syspkg remove vim

# Search for a package using all available package manager
syspkg search vim

# Upgrade all packages using all available package manager
syspkg upgrade
```

For more examples and real use cases, see the [cmd/syspkg/](cmd/syspkg/) directory.

### Go Library

Here's an example demonstrating how to use SysPkg as a Go library:

```go
package main

import (
 "fmt"
 "github.com/bluet/syspkg"
)

func main() {
 // Initialize SysPkg with all available package managers on current system
 includeOptions := syspkg.IncludeOptions{
  AllAvailable: true,
 }
 syspkgManager, err := syspkg.New(includeOptions)
 if err != nil {
  fmt.Printf("Error initializing SysPkg: %v\n", err)
  return
 }

 // List installed packages using APT
 aptManager := syspkgManager.GetPackageManager("apt")
 installedPackages, err := aptManager.ListInstalled(nil)
 if err != nil {
  fmt.Printf("Error listing installed packages: %v\n", err)
  return
 }

 fmt.Println("Installed packages:")
 for _, pkg := range installedPackages {
  fmt.Printf("- %s (%s)\n", pkg.Name, pkg.Version)
 }
}
```

For more examples and real use cases, see the [cmd/syspkg/](cmd/syspkg/) directory.

## Supported Package Managers

| Package Manager | Install | Remove | Search | Upgrade | List Installed | List Upgradable | Get Package Info |
| --------------- | ------- | ------ | ------ | ------- | -------------- | --------------- | ---------------- |
| APT             | ✅      | ✅    | ✅     | ✅     | ✅             | ✅             | ✅               |
| SNAP            | ✅      | ✅    | ✅     | ✅     | ✅             | ✅             | ✅               |
| Flatpak         | ❓      | ❓    | ✅     | ✅     | ✅             | ✅             | ✅               |
| Your favorite package manager here! | 🚀 | 🚀 | 🚀 | 🚀 | 🚀 | 🚀 | 🚀 |

Please open an issue (or PR ❤️) if you'd like to see support for any unlisted specific package manager.

### TODO

- [ ] Add support for more package managers
- [ ] Improve error handling
- [ ] Enhance return values and status codes

## Contributing

We welcome contributions to Go-SysPkg! Please read our [CONTRIBUTING.md](CONTRIBUTING.md) for more information on how to contribute.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
