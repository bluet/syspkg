# Docker-Based Testing Strategy

## Overview

This directory contains Docker configurations for testing go-syspkg across multiple Linux distributions.

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

## Docker Test Structure (Planned)

**Note:** The docker-compose.test.yml file is planned for future implementation. This shows the intended structure.

```yaml
# Planned: docker-compose.test.yml
version: '3.8'

services:
  ubuntu-test:
    build:
      context: ../..
      dockerfile: testing/docker/ubuntu.Dockerfile
    environment:
      - IN_CONTAINER=true
      - TEST_TAGS=unit,parser
    volumes:
      - ../..:/workspace
    working_dir: /workspace
    command: go test -tags="unit parser" ./...

  debian-test:
    build:
      context: ../..
      dockerfile: testing/docker/debian.Dockerfile
    environment:
      - IN_CONTAINER=true
      - TEST_TAGS=unit,parser
    volumes:
      - ../..:/workspace
    working_dir: /workspace
    command: go test -tags="unit parser" ./...

  fedora-test:
    build:
      context: ../..
      dockerfile: testing/docker/fedora.Dockerfile
    environment:
      - IN_CONTAINER=true
      - TEST_TAGS=unit,parser
    volumes:
      - ../..:/workspace
    working_dir: /workspace
    command: go test -tags="unit parser" ./...

  alpine-test:
    build:
      context: ../..
      dockerfile: testing/docker/alpine.Dockerfile
    environment:
      - IN_CONTAINER=true
      - TEST_TAGS=unit,parser
    volumes:
      - ../..:/workspace
    working_dir: /workspace
    command: go test -tags="unit parser" ./...
```

## Usage

**Note:** These Docker testing targets are planned for future implementation. Currently, the project tests on Ubuntu in CI/CD.

### Planned: Run All Container Tests
```bash
# TODO: Implement when Makefile targets are added
make test-docker-all
```

### Planned: Run Specific OS Test
```bash
# TODO: Implement when docker-compose.test.yml is created
docker-compose -f docker-compose.test.yml run ubuntu-test
```

### Capture Test Fixtures
```bash
# Run container interactively to capture package manager outputs
docker-compose -f docker-compose.test.yml run --rm ubuntu-test bash
apt update
apt search vim > testing/fixtures/apt/search-vim.txt
```

## Test Implementation Example

```go
// +build parser

package apt_test

import (
    "os"
    "os/exec"
    "testing"
    
    "github.com/bluet/syspkg/manager/apt"
)

func TestParseSearchOutput_MultiOS(t *testing.T) {
    // Skip if not in container
    if os.Getenv("IN_CONTAINER") != "true" {
        t.Skip("Skipping container-only test")
    }
    
    // Test with real apt output from this OS version
    output, err := exec.Command("apt", "search", "vim").Output()
    if err != nil {
        t.Skip("apt not available in this container")
    }
    
    packages := apt.ParseSearchOutput(string(output))
    
    // Verify parsing works for this OS version
    if len(packages) == 0 {
        t.Error("Expected to parse some packages")
    }
}
```

## CI Integration (Planned)

**Note:** This is a planned feature for multi-OS Docker testing. Currently, the project uses `.github/workflows/test.yml` for Ubuntu-based testing.

```yaml
# Planned: .github/workflows/multi-os-test.yml
name: Multi-OS Docker Tests

on: [push, pull_request]

jobs:
  docker-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Run Docker Tests
        run: |
          # TODO: Create docker-compose.test.yml before enabling
          docker-compose -f testing/docker/docker-compose.test.yml up \
            --abort-on-container-exit \
            --exit-code-from ubuntu-test
            
      - name: Upload Test Results
        uses: actions/upload-artifact@v3
        with:
          name: test-results
          path: test-results/  # TODO: Ensure tests output to this directory
```

## Best Practices

1. **Keep Images Minimal**: Install only what's required for testing
2. **Cache Aggressively**: Use Docker layer caching
3. **Parallelize Tests**: Run different OS tests concurrently
4. **Mock External Calls**: Don't actually install packages in tests
5. **Capture Real Outputs**: Use containers to generate test fixtures