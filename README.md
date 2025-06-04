# SysPkg

[![Test and Coverage](https://github.com/bluet/syspkg/actions/workflows/test-and-coverage.yml/badge.svg)](https://github.com/bluet/syspkg/actions/workflows/test-and-coverage.yml)
[![Build](https://github.com/bluet/syspkg/actions/workflows/build.yml/badge.svg)](https://github.com/bluet/syspkg/actions/workflows/build.yml)
[![Lint and Format](https://github.com/bluet/syspkg/actions/workflows/lint-and-format.yml/badge.svg)](https://github.com/bluet/syspkg/actions/workflows/lint-and-format.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/bluet/syspkg)](https://goreportcard.com/report/github.com/bluet/syspkg)
[![Go Reference](https://pkg.go.dev/badge/github.com/bluet/syspkg.svg)](https://pkg.go.dev/github.com/bluet/syspkg)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](https://github.com/bluet/syspkg/blob/main/LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/bluet/syspkg)](https://github.com/bluet/syspkg)
[![GitHub release](https://img.shields.io/github/v/release/bluet/syspkg)](https://github.com/bluet/syspkg/releases)

SysPkg is a unified CLI tool and Golang library for managing system packages across different package managers. Currently, it supports APT, YUM, Snap, and Flatpak, with plans for more. It simplifies package management by providing a consistent interface and API through an abstraction layer that focuses on package manager tools rather than specific operating systems.

## Key Features

- **Cross-Package Manager Compatibility**: Normalized status reporting (e.g., APT's config-files state maps to available)
- **Consistent API**: Same interface across all supported package managers
- **Tool-Focused**: Works wherever package manager tools work (containers, cross-platform, etc.)
- **Production Ready**: Comprehensive testing across multiple OS distributions
- **Performance Optimized**: Efficient parsing with compiled regexes and robust error handling
- **Cross-Platform**: Handles different line endings (CRLF/LF) and whitespace variations
- **Hybrid Search**: Fast default search with optional `--status` flag for detailed package state information

## Features

- A unified package management interface for various package managers
- Supports popular package managers such as APT, YUM, Snap, Flatpak, and more
- Easy-to-use API for package installation, removal, search, listing, and system upgrades
- Expandable architecture to support more package managers in the future

## Documentation

- [API Documentation](https://pkg.go.dev/github.com/bluet/syspkg) - Full Go API reference
- [Architecture Overview](docs/ARCHITECTURE_OVERVIEW.md) - Technical architecture and design patterns
- [Contributing Guide](CONTRIBUTING.md) - Development workflow and testing strategy
- [Test Fixtures](testing/fixtures/README.md) - Comprehensive fixture generation using Docker entrypoints

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

# List upgradable packages using Flatpak
syspkg --flatpak list upgradable

# Install a package using YUM (on RHEL/CentOS/Rocky/AlmaLinux)
syspkg --yum install vim

# Show package information
syspkg --apt info vim

# List installed packages
syspkg --snap list installed

# List upgradable packages
syspkg --flatpak list upgradable
```

Or, you can do operations without knowing the package manager:

```bash
# Install a package using all available package managers
syspkg install vim

# Remove a package using all available package manager
syspkg remove vim

# Search for a package using all available package manager
syspkg search vim

# Search with installation status information (slightly slower)
syspkg search vim --status

# Upgrade all packages using all available package manager
syspkg upgrade

# Update package lists
syspkg update

# Show package information
syspkg info vim

# List installed packages
syspkg list installed

# List upgradable packages
syspkg list upgradable
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

| Package Manager | Install | Remove | Search | Upgrade | List Installed | List Upgradable | Get Package Info | AutoRemove | Clean | Refresh |
| --------------- | ------- | ------ | ------ | ------- | -------------- | --------------- | ---------------- | ---------- | ----- | ------- |
| APT             | ✅      | ✅    | ✅     | ✅     | ✅             | ✅             | ✅               | ✅         | ✅    | ✅      |
| YUM             | ✅      | ✅    | ✅     | ✅     | ✅             | ✅             | ✅               | ✅         | ✅    | ✅      |
| SNAP            | ✅      | ✅    | ✅     | ✅     | ✅             | ✅             | ✅               | ✅         | ✅    | ✅      |
| Flatpak         | ✅      | ✅    | ✅     | ✅     | ✅             | ✅             | ✅               | ✅         | ✅    | ✅      |
| DNF             | 🚧      | 🚧    | 🚧     | 🚧     | 🚧             | 🚧             | 🚧               | 🚧         | 🚧    | 🚧      |
| APK (Alpine)    | 🚧      | 🚧    | 🚧     | 🚧     | 🚧             | 🚧             | 🚧               | 🚧         | 🚧    | 🚧      |
| Zypper (openSUSE) | 🚧   | 🚧    | 🚧     | 🚧     | 🚧             | 🚧             | 🚧               | 🚧         | 🚧    | 🚧      |

**Legend:** ✅ Implemented, 🚧 Planned, ❌ Not supported

**Philosophy:** SysPkg focuses on supporting package manager tools wherever they work, regardless of the underlying operating system. If you have apt+dpkg working on macOS via Homebrew, or in a container, SysPkg will support it.

Please open an issue (or PR ❤️) if you'd like to see support for any unlisted specific package manager.

## Development

### Documentation

### For Users 👥
- **[README.md](README.md)** (you are here) - Project overview and quick start
- **[Go Reference](https://pkg.go.dev/github.com/bluet/syspkg)** - Complete API documentation
- **[CHANGELOG.md](CHANGELOG.md)** - Recent achievements and version history

### For Developers 🛠️
- **[CONTRIBUTING.md](CONTRIBUTING.md)** - Development workflow and testing guide
- **[docs/ARCHITECTURE_OVERVIEW.md](docs/ARCHITECTURE_OVERVIEW.md)** - Technical design and interfaces
- **[docs/EXIT_CODES.md](docs/EXIT_CODES.md)** - Package manager exit code behaviors
- **[testing/fixtures/README.md](testing/fixtures/README.md)** - Comprehensive fixture generation with Docker entrypoints
- **[testing/docker/README.md](testing/docker/README.md)** - Multi-OS testing strategy

### For AI Assistants 🤖
- **[CLAUDE.md](CLAUDE.md)** - Development guidelines and project rules

## Project Status

**Current Version**: [Latest Release](https://github.com/bluet/syspkg/releases)

**Stability**: Production ready with unified interface architecture and comprehensive fixture-based testing across multiple OS distributions

**Architecture**: V2.0 unified interface with plugin system - APT implementation complete, legacy managers in transition

**Active Development**: See [Issues](https://github.com/bluet/syspkg/issues) for roadmap and current work

### Current Priorities
- **Branch Integration**: Merge refactor-unified-interface to main
- **Legacy Cleanup**: Resolve backup directory compilation issues
- **Security Enhancement**: Add Snyk security scanning to CI/CD
- **Platform Expansion**: Snap and Flatpak plugin migration to unified interface
- **New Package Managers**: DNF and APK implementations using unified architecture

See [CHANGELOG.md](CHANGELOG.md) for recent achievements and [CLAUDE.md](CLAUDE.md) for detailed development roadmap.

## Contributing

We welcome contributions to SysPkg!

### For Users
- **Bug reports**: Open an issue with details about the problem
- **Feature requests**: Please let us know what package managers or features you'd like to see

### For Developers
- **Quick start**: See [CONTRIBUTING.md](CONTRIBUTING.md) for complete development workflow
- **Architecture**: See [docs/ARCHITECTURE_OVERVIEW.md](docs/ARCHITECTURE_OVERVIEW.md) for technical design details
- **Fixture generation**: See [testing/fixtures/README.md](testing/fixtures/README.md) for comprehensive approach

**Quick development setup:**
```bash
git clone https://github.com/bluet/syspkg.git
cd syspkg
make test          # Smart testing - detects your OS
make check         # Code quality checks
make help          # See all available targets
```

**Advanced fixture generation:**
```bash
make test-fixtures-apt    # Generate realistic APT fixtures using Docker entrypoints
make test-fixtures        # Generate fixtures for all package managers
make test-fixtures-validate  # Validate fixture quality
```

For complete development workflow and multi-OS testing, see [CONTRIBUTING.md](CONTRIBUTING.md).

## License

This project is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for details.
