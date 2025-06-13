# SysPkg

[![Test and Coverage](https://github.com/bluet/syspkg/actions/workflows/test-and-coverage.yml/badge.svg)](https://github.com/bluet/syspkg/actions/workflows/test-and-coverage.yml)
[![Build](https://github.com/bluet/syspkg/actions/workflows/build.yml/badge.svg)](https://github.com/bluet/syspkg/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/bluet/syspkg)](https://goreportcard.com/report/github.com/bluet/syspkg)
[![Go Reference](https://pkg.go.dev/badge/github.com/bluet/syspkg.svg)](https://pkg.go.dev/github.com/bluet/syspkg)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](https://github.com/bluet/syspkg/blob/main/LICENSE)

**A unified package management tool for Linux systems** - One interface for APT, YUM, Snap, Flatpak, and more.

SysPkg provides a consistent CLI and Go library interface across different package managers, making it easy to manage packages regardless of your Linux distribution or containerized environment.

## Features

- **🔧 Unified Interface**: Same commands work with APT, YUM, Snap, and Flatpak
- **⚡ Concurrent Operations**: 3x faster multi-manager operations with automatic parallelization
- **🛡️ Secure by Design**: Input validation and command injection prevention
- **🐳 Container Ready**: Works in Docker, LXC, and other containerized environments
- **📊 Rich Output**: JSON, table, and human-readable formats
- **🚀 Production Ready**: Thread-safe concurrent execution with comprehensive error handling
- **🔍 Smart Search**: Automatically searches across available package managers

**Supported Package Managers:**

| Package Manager | Status | Distributions |
|-----------------|--------|---------------|
| **APT** | ✅ Production | Ubuntu, Debian, derivatives |
| **YUM** | ✅ Production | RHEL, CentOS, Rocky Linux 8 |
| **APK** | ✅ Production | Alpine Linux |
| **Snap** | 🚧 Beta | Universal Linux packages |
| **Flatpak** | 🚧 Beta | Universal Linux applications |
| **DNF** | 🚧 Planned | Fedora, RHEL 9+ |

*More package managers coming soon: Pacman, Zypper, and more.*

**Quick Documentation Reference:**

- **👥 New users?** → Continue reading this README
- **🔧 Want to contribute?** → [CONTRIBUTING.md](CONTRIBUTING.md)
- **⚙️ Technical details?** → [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
- **🧪 Testing & fixtures?** → [docs/TESTING.md](docs/TESTING.md)
- **🔌 Building plugins?** → [docs/PLUGIN_DEVELOPMENT.md](docs/PLUGIN_DEVELOPMENT.md)
- **🏗️ Production integration?** → [docs/INTEGRATION_GUIDE.md](docs/INTEGRATION_GUIDE.md)
- **🔢 Exit codes & automation?** → [docs/EXIT_CODES.md](docs/EXIT_CODES.md)

## Getting Started

**Prerequisites:** Go 1.23 or later

**Install CLI tool:**
```bash
go install github.com/bluet/syspkg/cmd/syspkg@latest
```

**Install Go library:**
```bash
go get github.com/bluet/syspkg
```

## Usage

**CLI Tool Examples:**

```bash
# Use specific package managers
syspkg -m apt install vim       # Install using APT
syspkg -m snap search vim       # Search using Snap
syspkg -m flatpak list upgradable  # List using Flatpak
syspkg -m yum install vim       # Install using YUM

# Use by manager category
syspkg -c system install vim    # Use system package manager (apt/yum/apk)
syspkg -c app search vim        # Use app managers (snap/flatpak)
```

**Multi-Manager Operations** (use `--all` for 3x performance with concurrent execution):

```bash
# Read-only operations (concurrent, safe, fast)
syspkg search vim --all             # Search across all managers (parallel execution)
syspkg list installed --all         # List from all managers (concurrent)
syspkg info vim --all               # Get info from all managers (parallel)
syspkg status --all                 # Show status of all managers (concurrent)

# Write operations (concurrent with safety prompts)
syspkg update --all                 # Update package lists (parallel across managers)
syspkg upgrade --all                # Upgrade packages (concurrent execution)
syspkg clean --all                  # Clean caches (parallel cleaning)
syspkg autoremove --all             # Remove orphaned packages (concurrent)
syspkg install vim curl --all       # Install across managers (parallel)
syspkg remove package --all         # Remove from managers (concurrent)

# Performance and safety options
syspkg upgrade --all --yes          # Skip confirmation prompt
syspkg clean --all --dry-run        # See what would be done
```

**Auto-Selection:** Without `--all` or `-m`, syspkg automatically selects the best available system package manager:

```bash
syspkg install vim                  # Uses best system manager (apt/yum/apk)
syspkg search vim                   # Searches all managers (default behavior)
syspkg list installed               # Uses best system manager
syspkg upgrade                      # Uses best system manager
```

**Manager Selection Logic** - SysPkg uses priority-based automatic selection:

| Manager | Priority | Use Case |
|---------|----------|----------|
| **APT** | 90 | Ubuntu/Debian systems |
| **YUM** | 80 | RHEL/CentOS systems |
| **Snap** | 80 | Universal packages |
| **Flatpak** | 70 | Universal applications |
| **APK** | 60 | Alpine Linux |

```bash
# Automatic selection (no flags)
syspkg install vim                  # Uses APT on Ubuntu, YUM on RHEL, etc.
syspkg search vim                   # Searches ALL managers (multi-manager default)

# Category-based selection
syspkg install vim -c system        # Uses highest priority system manager
syspkg install discord -c app       # Uses highest priority app manager (Snap > Flatpak)

# Explicit manager selection
syspkg install vim -m yum           # Force specific manager regardless of priority
```

**Pipeline Support** - Use `-` to read package names from stdin:

```bash
# Install packages from a file
cat packages.txt | syspkg install -

# Install multiple packages from command output
echo "vim curl git" | syspkg install -

# Verify all installed packages (with new tab-separated format)
syspkg list installed -q | cut -f1 | syspkg verify -

# Advanced pipeline with manager information
syspkg list installed --all -q | awk '$2=="apt"' | cut -f1 | syspkg verify -
```

**Output Formats** - Multiple formats optimized for different use cases:

```bash
# Human-readable format (default)
syspkg list installed
# Output: packagename [manager] [version] status - description

# Quiet mode - tab-separated for parsing
syspkg list installed -q
# Output: packagename<TAB>manager<TAB>version<TAB>status

# JSON output for programmatic use
syspkg list installed -j
# Output: [{"name":"vim","manager":"apt","version":"8.2","status":"installed"}]

# With status information (slower but accurate)
syspkg search vim --status
# Shows real installation status across managers
```

**Go Library** - Use SysPkg programmatically with concurrent operations:

```go
package main

import (
 "context"
 "fmt"
 "github.com/bluet/syspkg/manager"

 // Import package managers you want to use
 _ "github.com/bluet/syspkg/manager/apt"
 _ "github.com/bluet/syspkg/manager/yum"
)

func main() {
 // Get the global registry of package managers
 registry := manager.GetGlobalRegistry()

 // Single manager usage
 aptManager, exists := registry.GetAvailable()["apt"]
 if exists {
  ctx := context.Background()
  packages, err := aptManager.List(ctx, manager.FilterInstalled, nil)
  if err == nil {
   fmt.Printf("APT: %d packages installed\n", len(packages))
  }
 }

 // Concurrent multi-manager operations (3x faster!)
 ctx := context.Background()
 opts := manager.DefaultOptions()

 // Search across all managers concurrently
 searchResults := registry.SearchAllConcurrent(ctx, []string{"vim"}, opts)
 for managerName, packages := range searchResults {
  fmt.Printf("%s found %d packages\n", managerName, len(packages))
 }

 // List installed packages from all managers concurrently
 allInstalled := registry.ListInstalledAllConcurrent(ctx, opts)
 totalInstalled := 0
 for managerName, packages := range allInstalled {
  fmt.Printf("%s: %d installed\n", managerName, len(packages))
  totalInstalled += len(packages)
 }
 fmt.Printf("Total: %d packages across all managers\n", totalInstalled)
}
```

**Learning Resources:**
- **New to the API?** → Start with [examples/](examples/) for clean integration patterns
- **Building production services?** → See [docs/INTEGRATION_GUIDE.md](docs/INTEGRATION_GUIDE.md) for advanced patterns
- **Need CLI reference?** → Check [cmd/syspkg/](cmd/syspkg/) for complete implementation

## Supported Package Managers

| Package Manager | Install | Remove | Search | Upgrade | List Installed | List Upgradable | Get Package Info | AutoRemove | Clean | Refresh |
| --------------- | ------- | ------ | ------ | ------- | -------------- | --------------- | ---------------- | ---------- | ----- | ------- |
| APT             | ✅      | ✅    | ✅     | ✅     | ✅             | ✅             | ✅               | ✅         | ✅    | ✅      |
| YUM             | ✅      | ✅    | ✅     | ✅     | ✅             | ✅             | ✅               | ✅         | ✅    | ✅      |
| SNAP            | ✅      | ✅    | ✅     | ✅     | ✅             | ✅             | ✅               | ✅         | ✅    | ✅      |
| Flatpak         | ✅      | ✅    | ✅     | ✅     | ✅             | ✅             | ✅               | ✅         | ✅    | ✅      |
| DNF             | 🚧      | 🚧    | 🚧     | 🚧     | 🚧             | 🚧             | 🚧               | 🚧         | 🚧    | 🚧      |
| APK (Alpine)    | ✅      | ✅    | ✅     | ✅     | ✅             | ✅             | ✅               | ✅         | ✅    | ✅      |
| Zypper (openSUSE) | 🚧   | 🚧    | 🚧     | 🚧     | 🚧             | 🚧             | 🚧               | 🚧         | 🚧    | 🚧      |

**Legend:** ✅ Implemented, 🚧 Planned, ❌ Not supported

**Philosophy:** SysPkg focuses on supporting package manager tools wherever they work, regardless of the underlying operating system. If you have apt+dpkg working on macOS via Homebrew, or in a container, SysPkg will support it.

Please open an issue (or PR ❤️) if you'd like to see support for any unlisted specific package manager.

## Development

**Documentation by Audience:**

**For Users 👥**
- **[README.md](README.md)** (you are here) - Project overview and quick start
- **[Go Reference](https://pkg.go.dev/github.com/bluet/syspkg)** - Complete API documentation
- **[CHANGELOG.md](CHANGELOG.md)** - Recent achievements and version history

**For Developers 🛠️**
- **[CONTRIBUTING.md](CONTRIBUTING.md)** - Development workflow and testing guide
- **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** - Technical design and interfaces
- **[docs/EXIT_CODES.md](docs/EXIT_CODES.md)** - Package manager exit code behaviors
- **[testing/fixtures/README.md](testing/fixtures/README.md)** - Comprehensive fixture generation with Docker entrypoints
- **[testing/docker/README.md](testing/docker/README.md)** - Multi-OS testing strategy

**For AI Assistants 🤖**
- **[CLAUDE.md](CLAUDE.md)** - Development guidelines and project rules

**Project Status:** [Latest Release](https://github.com/bluet/syspkg/releases) • Production ready with unified interface architecture and comprehensive fixture-based testing across multiple OS distributions • V2.0 unified interface with plugin system - APT implementation complete, legacy managers in transition • Active development at [Issues](https://github.com/bluet/syspkg/issues)

**Current Priorities:**
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
- **Architecture**: See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for technical design details
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
