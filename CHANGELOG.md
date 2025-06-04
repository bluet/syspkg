# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **CRITICAL**: Complete unified interface architecture implementation (Issue #20)
- **Real fixture generation**: Docker entrypoint approach for authentic test data
- **Comprehensive APT plugin**: All 13 operations with real functionality
- **Security enhancements**: Input validation and command injection prevention
- **Architecture documentation**: Complete technical guides and development workflow
- Technical debt cleanup and APT Upgrade method fix
- Comprehensive documentation reorganization with proper cross-references

### Fixed
- **CRITICAL**: APT Upgrade method now correctly uses `apt install` for specific packages instead of upgrading all packages
- **Build system**: Excluded legacy backup files from compilation
- **Test reliability**: Real fixture-based testing with Docker-generated outputs

## Recent Achievements ✅

### Architecture & Code Quality
- ✅ **Unified Interface Architecture**: Complete implementation with APT plugin (Issue #20)
- ✅ **CommandRunner Architecture**: Complete architectural consistency (Issue #20, PR #26)
- ✅ **Plugin System**: Registry with auto-registration and priority-based selection
- ✅ **Real Fixture Testing**: Docker entrypoint approach for comprehensive test coverage
- ✅ **APT & YUM executeCommand Pattern**: Centralized command execution, eliminated code duplication
- ✅ **Technical Debt Cleanup**: Fixed APT Upgrade method bug, removed misleading TODOs, verified no resource leaks

### Security Enhancements
- ✅ **Security Enhancements**: Input validation for package names (Issue #23, PR #25)
- ✅ **Command Injection Prevention**: Comprehensive ValidatePackageName implementation across all package managers

### Bug Fixes
- ✅ **APT Exit Code Bug**: Fixed in commit 3751f45 - now properly propagates errors (Issue #21)
- ✅ **Snap Exit Code Bug**: Fixed in commit 3751f45 - now properly handles usage errors (Issue #22)
- ✅ **Flatpak Exit Code Bug**: Fixed in commit 3751f45 - now properly handles general errors (Issue #24)

## CI/CD Status

| Workflow | Status | Description |
| -------- | ------ | ----------- |
| **Test and Coverage** | ✅ | Go 1.23/1.24 testing with coverage reporting |
| **Lint and Format** | ✅ | golangci-lint, gofmt, go vet quality checks |
| **Build** | ✅ | Multi-version build verification |
| **Multi-OS Tests** | ✅ | Docker-based testing across Ubuntu, Rocky Linux, Alpine |
| **Release Binaries** | ✅ | Cross-platform binary releases |

### Infrastructure Status
- ✅ **Pre-commit hooks**: Automated code quality and security checks
- ✅ **Go mod verification**: Dependency integrity validation
- ✅ **Multi-OS compatibility**: Docker testing with Go 1.23.4 across distributions
- ✅ **Fixture-based testing**: Real package manager output validation

## Active Development

Current development focus areas (see [GitHub Issues](https://github.com/bluet/syspkg/issues) and [CLAUDE.md](CLAUDE.md) for detailed tracking):

### High Priority Pending
- **Security scanning with Snyk** - Add to CI/CD pipeline
- **Legacy code cleanup** - Resolve backup directory compilation issues
- **Branch merging** - Merge refactor-unified-interface to main

### Medium Priority Pending
- **Test coverage improvements** - YUM gaps (Issue #32), Snap & Flatpak comprehensive suites
- **CLI improvements** - Upgrade display (Issue #3), macOS apt conflict (Issue #2)
- **Code quality** - Context support, custom error types, DRY principle improvements

### Low Priority Pending
- **New package managers** - DNF, APK, Homebrew, Windows package managers
- **Bug fixes** - APT multi-arch parsing (Issue #15), version parsing improvements

## Platform Support Status

### Currently Supported ✅
- **APT** (Ubuntu/Debian) - Full feature support
- **YUM** (Rocky Linux/AlmaLinux/RHEL) - Full feature support
- **Snap** (Universal packages) - Full feature support
- **Flatpak** (Universal packages) - Full feature support

### In Development 🚧
- **DNF** (Fedora/RHEL 9+) - Implementation in progress
- **APK** (Alpine Linux) - Implementation in progress

### Planned 📋
- **Homebrew** (macOS) - Planned for cross-platform expansion
- **Chocolatey/Scoop/winget** (Windows) - Planned for Windows support
- **Zypper** (openSUSE) - Lower priority
