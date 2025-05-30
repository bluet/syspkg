# SysPkg

[![Test and Coverage](https://github.com/bluet/syspkg/actions/workflows/test-and-coverage.yml/badge.svg)](https://github.com/bluet/syspkg/actions/workflows/test-and-coverage.yml)
[![Build](https://github.com/bluet/syspkg/actions/workflows/build.yml/badge.svg)](https://github.com/bluet/syspkg/actions/workflows/build.yml)
[![Lint and Format](https://github.com/bluet/syspkg/actions/workflows/lint-and-format.yml/badge.svg)](https://github.com/bluet/syspkg/actions/workflows/lint-and-format.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/bluet/syspkg)](https://goreportcard.com/report/github.com/bluet/syspkg)
[![Go Reference](https://pkg.go.dev/badge/github.com/bluet/syspkg.svg)](https://pkg.go.dev/github.com/bluet/syspkg)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](https://github.com/bluet/syspkg/blob/main/LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/bluet/syspkg)](https://github.com/bluet/syspkg)
[![GitHub release](https://img.shields.io/github/v/release/bluet/syspkg)](https://github.com/bluet/syspkg/releases)

SysPkg is a unified CLI tool and Golang library for managing system packages across different package managers. Currently, it supports APT, Snap, and Flatpak, with plans for more. It simplifies package management by providing a consistent interface and API through an abstraction layer that focuses on package manager tools rather than specific operating systems.

## Features

- A unified package management interface for various package managers
- Supports popular package managers such as APT, Snap, Flatpak, and more
- Easy-to-use API for package installation, removal, search, listing, and system upgrades
- Expandable architecture to support more package managers in the future

## API Documentation

See the [Go Reference](https://pkg.go.dev/github.com/bluet/syspkg) for the full API documentation.

## Getting Started

### Prerequisites

- Go 1.23 or later

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

 // Get APT package manager (if available)
 aptManager, err := syspkgManager.GetPackageManager("apt")
 if err != nil {
  fmt.Printf("APT package manager not available: %v\n", err)
  return
 }

 // List installed packages using APT
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
| YUM             | ❓      | ❓    | ✅     | ❓     | ✅             | ❓             | ✅               |
| SNAP            | ✅      | ✅    | ✅     | ✅     | ✅             | ✅             | ✅               |
| Flatpak         | ✅      | ✅    | ✅     | ✅     | ✅             | ✅             | ✅               |
| DNF/YUM         | 🚧      | 🚧    | 🚧     | 🚧     | 🚧             | 🚧             | 🚧               |
| APK (Alpine)    | 🚧      | 🚧    | 🚧     | 🚧     | 🚧             | 🚧             | 🚧               |
| Zypper (openSUSE) | 🚧   | 🚧    | 🚧     | 🚧     | 🚧             | 🚧             | 🚧               |

**Legend:** ✅ Implemented, 🚧 Planned, ❌ Not supported

**Philosophy:** SysPkg focuses on supporting package manager tools wherever they work, regardless of the underlying operating system. If you have apt+dpkg working on macOS via Homebrew, or in a container, SysPkg will support it.

Please open an issue (or PR ❤️) if you'd like to see support for any unlisted specific package manager.

## Development

### Documentation
- **CLAUDE.md** - Development guidelines, architecture, and project roadmap
- **testing/** - Test fixtures and Docker testing infrastructure
- **.pre-commit-config.yaml** - Secure pre-commit hooks aligned with Go best practices
- **.github/workflows/** - CI/CD pipelines for testing, linting, building, and releases

### CI/CD Status

| Workflow | Status | Description |
| -------- | ------ | ----------- |
| **Test and Coverage** | ✅ | Go 1.23/1.24 testing with coverage reporting |
| **Lint and Format** | ✅ | golangci-lint, gofmt, go vet quality checks |
| **Build** | ✅ | Multi-version build verification |
| **Release Binaries** | ✅ | Cross-platform binary releases |

- ✅ **Pre-commit hooks**: Automated code quality and security checks
- ✅ **Go mod verification**: Dependency integrity validation
- 🚧 **Multi-platform testing**: macOS/Windows testing planned

### Development Setup

1. **Clone and setup**:
   ```bash
   git clone https://github.com/bluet/syspkg.git
   cd syspkg
   ```

2. **Install pre-commit hooks**:
   ```bash
   pre-commit install
   ```

3. **Run development commands**:
   ```bash
   make test          # Run tests
   make check         # Code quality checks
   make build         # Build binary
   ```

### Contributing
See [CLAUDE.md](CLAUDE.md) for detailed development guidelines and architecture overview.

### TODO

- [ ] Add brew support for macOS
- [ ] Add chocolatey/scoop/winget support for Windows
- [ ] Add support for more Linux package managers (dnf, apk, zypper)
- [ ] Implement Docker-based testing for multi-OS validation
- [ ] Improve error handling and status codes

## Contributing

We welcome contributions to Go-SysPkg! Please read our [CONTRIBUTING.md](CONTRIBUTING.md) for more information on how to contribute.

## License

This project is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for details.
