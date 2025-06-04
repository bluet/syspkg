# Docker-Based Testing Strategy

## Overview

This directory contains Docker configurations for testing go-syspkg across multiple Linux distributions. The multi-OS Docker testing system is **fully implemented and actively used** for comprehensive cross-platform validation.

## 📖 Related Documentation

- **[../../README.md](../../README.md)** - Project overview
- **[../../CONTRIBUTING.md](../../CONTRIBUTING.md)** - Complete development workflow and testing guide
- **[../../docs/ARCHITECTURE_OVERVIEW.md](../../docs/ARCHITECTURE_OVERVIEW.md)** - Technical design and interfaces
- **[../../docs/EXIT_CODES.md](../../docs/EXIT_CODES.md)** - Package manager exit code behavior

## Test Categories

### 1. Unit Tests (Run Everywhere)
- Parser functions with captured outputs
- OS detection logic
- Command construction
- No actual package manager execution

### 2. Integration Tests (Container-Specific)
- Real package manager availability checks
- Command output capture for test fixtures
- Limited package operations (list, search)

### 3. Full System Tests (Native CI Only)
- Actual package installation/removal
- Privileged operations
- Snap/systemd dependent features

## Current Docker Test Implementation

**Note:** The Docker testing system is fully operational with comprehensive multi-OS support.

**Supported Operating Systems:**
- **Ubuntu 22.04**: APT, Snap, Flatpak testing
- **Rocky Linux 8**: YUM testing with real RPM packages
- **AlmaLinux 8**: YUM testing and validation
- **Fedora 39**: DNF testing (implementation in progress)
- **Alpine Linux**: APK testing (implementation in progress)

## Usage

### Run All Container Tests
```bash
make test-docker-all        # Test all OS in parallel
```

### Run Specific OS Tests
```bash
make test-docker-ubuntu     # Test APT/Snap/Flatpak on Ubuntu
make test-docker-rocky      # Test YUM on Rocky Linux 8
make test-docker-alma       # Test YUM on AlmaLinux 8
make test-docker-fedora     # Test DNF on Fedora 39
make test-docker-alpine     # Test APK on Alpine Linux
```

### Cleanup Docker Resources
```bash
make test-docker-clean      # Remove test containers and images
```

### Generate Test Fixtures
```bash
# Generate fresh fixtures from real package managers
make test-fixtures           # Capture outputs from all supported OS

# Manual fixture generation for specific OS
docker exec -it syspkg-rocky-test bash
yum search vim > /workspace/testing/fixtures/yum/search-vim-rocky8.txt
```

## Current CI Integration

The Docker testing system is fully integrated into CI/CD:

**GitHub Actions Workflows:**
- **test-and-coverage.yml**: Standard Ubuntu testing with APT/Snap/Flatpak
- **multi-os-test.yml**: Docker matrix testing across all supported OS
- **build.yml**: Multi-version Go build verification

**Multi-OS Testing Matrix:**
```yaml
# .github/workflows/multi-os-test.yml (active)
strategy:
  matrix:
    include:
      - os: ubuntu-22.04
        pm: apt
        dockerfile: ubuntu.Dockerfile
      - os: rockylinux-8
        pm: yum
        dockerfile: rockylinux.Dockerfile
      - os: almalinux-8
        pm: yum
        dockerfile: almalinux.Dockerfile
      - os: fedora-39
        pm: dnf
        dockerfile: fedora.Dockerfile
      - os: alpine-3.18
        pm: apk
        dockerfile: alpine.Dockerfile
```

## Best Practices

1. **Keep Images Minimal**: Install only what's required for testing
2. **Cache Aggressively**: Use Docker layer caching
3. **Parallelize Tests**: Run different OS tests concurrently
4. **Mock External Calls**: Don't actually install packages in tests
5. **Capture Real Outputs**: Use containers to generate test fixtures
