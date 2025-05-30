# Test Improvements Needed

## High Priority (Security & Reliability)

1. **Add snap/flatpak parse function tests**
   - Copy the pattern from apt/utils_test.go
   - Test with real command outputs
   - Cover error cases

2. **Create mock-based integration tests**
   - Mock exec.Command to avoid running real package operations
   - Test error handling (command not found, permission denied, etc.)
   - Test command construction with various options

3. **Simplify TestNewPackageManager**
   - Remove tests for unimplemented package managers
   - Focus only on apt, snap, flatpak

## Example Mock Test Structure

```go
// Mock command executor
type mockCommandExecutor struct {
    responses map[string]mockResponse
}

type mockResponse struct {
    stdout string
    stderr string
    err    error
}

func TestInstallWithMock(t *testing.T) {
    executor := &mockCommandExecutor{
        responses: map[string]mockResponse{
            "apt install -y vim": {
                stdout: "Reading package lists...\nInstalling vim...",
                err:    nil,
            },
        },
    }
    
    pm := apt.NewWithExecutor(executor)
    pkgs, err := pm.Install([]string{"vim"}, &manager.Options{AssumeYes: true})
    
    assert.NoError(t, err)
    assert.Len(t, pkgs, 1)
    assert.Equal(t, "vim", pkgs[0].Name)
}
```

## Test Coverage Goals

- apt: 80%+ (currently ~40%)
- snap: 80%+ (currently 0%)
- flatpak: 80%+ (currently 0%)
- syspkg: 70%+ (currently ~30%)
- CLI: 50%+ (currently 0%)

## What NOT to Test

1. Don't test third-party libraries (urfave/cli)
2. Don't test actual package installations in unit tests
3. Don't test OS-specific behavior that can't be mocked
4. Don't test for package managers that aren't implemented