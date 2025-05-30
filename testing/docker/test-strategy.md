# Docker Testing Strategy for Package Managers

## What Works Well in Docker

### ✅ APT Testing
```bash
docker run --rm -v $(PWD):/workspace ubuntu:22.04 bash -c "
  apt update
  apt search vim | head -20 > /workspace/testing/fixtures/apt/search-output.txt
  apt show vim > /workspace/testing/fixtures/apt/show-output.txt
  dpkg -l | head -20 > /workspace/testing/fixtures/apt/list-output.txt
"
```

### ✅ DNF/YUM Testing (Fedora)
```bash
docker run --rm -v $(PWD):/workspace fedora:38 bash -c "
  dnf search vim | head -20 > /workspace/testing/fixtures/dnf/search-output.txt
  dnf info vim > /workspace/testing/fixtures/dnf/info-output.txt
"
```

### ✅ APK Testing (Alpine)
```bash
docker run --rm -v $(PWD):/workspace alpine:3.18 sh -c "
  apk search vim > /workspace/testing/fixtures/apk/search-output.txt
  apk info vim > /workspace/testing/fixtures/apk/info-output.txt
"
```

### ✅ Flatpak Testing (Limited)
```bash
# Flatpak can list/search without full functionality
docker run --rm -v $(PWD):/workspace ubuntu:22.04 bash -c "
  apt update && apt install -y flatpak
  flatpak remote-add --if-not-exists flathub https://flathub.org/repo/flathub.flatpakrepo
  flatpak search vim > /workspace/testing/fixtures/flatpak/search-output.txt || true
"
```

## What Doesn't Work in Docker

### ❌ Snap Operations
- `snap install/remove` - Requires snapd daemon
- `snap list` - Requires running snapd
- `snap find` - Requires snapd connection

### ❌ System Package Installation
- Installing actual packages (without --privileged)
- System-level operations
- Hardware access

## Recommended Test Approach

### 1. **Docker for Fixture Generation**
Use Docker to generate real command outputs:

```go
//go:generate docker run --rm -v $(PWD):/workspace ubuntu:22.04 bash -c "apt update && apt search vim > /workspace/testing/fixtures/apt/search-vim.txt"
```

### 2. **Mock Testing for CI**
```go
// +build !integration

func TestAptInstall(t *testing.T) {
    executor := &MockExecutor{
        outputs: map[string]string{
            "apt install -y vim": readFixture("apt/install-vim.txt"),
        },
    }
    
    pm := apt.NewWithExecutor(executor)
    packages, err := pm.Install([]string{"vim"}, &manager.Options{AssumeYes: true})
    
    assert.NoError(t, err)
    assert.Len(t, packages, 1)
}
```

### 3. **Integration Testing Matrix**
```yaml
# .github/workflows/integration.yml
jobs:
  apt-test:
    runs-on: ubuntu-latest
    steps:
      - run: |
          sudo apt update
          sudo apt install -y vim
          go test -tags=integration ./manager/apt
          
  snap-test:
    runs-on: ubuntu-latest
    steps:
      - run: |
          sudo snap install hello-world
          go test -tags=integration ./manager/snap
          
  flatpak-test:
    runs-on: ubuntu-latest
    steps:
      - run: |
          sudo apt install -y flatpak
          go test -tags=integration ./manager/flatpak
```

## Docker Test Implementation

### Base Test Runner
```go
package testing

type DockerTestRunner struct {
    Image string
    Cmd   string
}

func (d *DockerTestRunner) CaptureOutput(outputFile string) error {
    cmd := exec.Command("docker", "run", "--rm", 
        "-v", fmt.Sprintf("%s:/workspace", os.Getwd()),
        d.Image,
        "bash", "-c", d.Cmd + " > /workspace/" + outputFile)
    return cmd.Run()
}
```

### Usage in Tests
```go
func TestCaptureRealOutputs(t *testing.T) {
    if os.Getenv("CAPTURE_FIXTURES") != "true" {
        t.Skip("Skipping fixture capture")
    }
    
    runner := &DockerTestRunner{
        Image: "ubuntu:22.04",
        Cmd:   "apt update && apt search golang",
    }
    
    err := runner.CaptureOutput("testing/fixtures/apt/search-golang.txt")
    assert.NoError(t, err)
}
```

## Summary

- **Use Docker for**: Capturing real outputs, testing parsers, OS detection
- **Don't use Docker for**: snap operations, actual installations, privileged operations
- **Alternative for snap**: Mock data or native CI runners