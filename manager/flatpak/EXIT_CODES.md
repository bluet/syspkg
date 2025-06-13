# Flatpak Exit Codes Reference

This document contains comprehensive exit code mappings for Flatpak package manager operations, generated through systematic Docker testing in Ubuntu 22.04.

## Exit Code Summary

| Exit Code | Category | Description |
|-----------|----------|-------------|
| **0** | Success | Successful operations, no results found, or warnings only |
| **1** | General Error | Invalid commands, package not found, network errors, application errors |

## Detailed Exit Code Analysis

### Exit Code 0 - Success and Warnings

Flatpak returns exit code 0 for successful operations and scenarios where no error occurred, even if no results were found.

**Success Cases:**
- Command executed successfully: `flatpak --version`, `flatpak --help`
- Package operations completed: remote add, install, remove (when successful)
- Empty results (not errors): `flatpak list` (empty), `flatpak search` (no matches)
- Update operations: `flatpak update` (even when nothing to update)
- List operations: `flatpak remotes` (even when empty)

**Warning Cases (still exit code 0):**
- Invalid remote URLs with warnings: `flatpak remote-add invalid-remote not-a-url`
- Network warnings during remote operations

### Exit Code 1 - General Errors

Flatpak uses exit code 1 for all error conditions, making error differentiation dependent on stderr message parsing.

**Command Errors:**
- Invalid commands: `flatpak invalidcommand`
- Invalid options: `flatpak --invalid-option`

**Package Errors:**
- Package not found: `flatpak info org.nonexistent.App`
- Package not installed: `flatpak remove org.nonexistent.App`, `flatpak uninstall org.nonexistent.App`
- Application not installed: `flatpak run org.nonexistent.App`

**Installation Errors:**
- No remotes configured: `flatpak install org.vim.Vim`
- Package not found in remotes: `flatpak install org.nonexistent.App`

**Network/Remote Errors:**
- Invalid remote URLs: `flatpak remote-add test https://example.com/repo.flatpakrepo`
- Server errors (404, etc.)

## Operation-Specific Behavior

### Search Operations
- **Command**: `flatpak search {query}`
- **Success**: Exit 0, results to stdout
- **No results**: Exit 0, "No matches found"
- **No remotes**: Exit 0, "No matches found"

### List Operations
- **Command**: `flatpak list [--app]`
- **Success**: Exit 0, package list to stdout
- **Empty**: Exit 0, no output
- **Invalid options**: Exit 1, error to stderr

### Install Operations
- **Command**: `flatpak install {package}`
- **Success**: Exit 0, installation details to stdout
- **Package not found**: Exit 1, "No remote refs found" to stderr
- **No remotes**: Exit 1, "No remote refs found" to stderr
- **Network issues**: Exit 1, specific error to stderr

### Remove/Uninstall Operations
- **Command**: `flatpak remove {package}` / `flatpak uninstall {package}`
- **Success**: Exit 0, removal details to stdout
- **Package not installed**: Exit 1, "not installed" to stderr

### Update Operations
- **Command**: `flatpak update`
- **Success**: Exit 0, update progress to stdout
- **Nothing to update**: Exit 0, "Nothing to do."
- **Network issues**: Exit 1, specific error to stderr

### Remote Management
- **Command**: `flatpak remote-add {name} {url}`
- **Success**: Exit 0, minimal output
- **Network errors**: Exit 1, HTTP error to stderr
- **Invalid URL format**: Exit 0, warning to stderr (surprisingly!)

### Application Execution
- **Command**: `flatpak run {app}`
- **Success**: Exit 0, application runs
- **App not installed**: Exit 1, "not installed" to stderr

### Info Operations
- **Command**: `flatpak info {package}`
- **Success**: Exit 0, package details to stdout
- **Package not found**: Exit 1, "not installed" to stderr

## Implementation Notes

### Current syspkg Implementation Analysis

**Exit Code Handling:**
- ✅ Exit code 0 correctly mapped to success
- ✅ Exit code 1 correctly mapped to general error
- ❌ No differentiation between error types (all use exit code 1)

**Issues Found:**
- ❌ Package not found (exit 1) needs specific detection via stderr parsing
- ❌ Network errors (exit 1) need differentiation from package errors
- ❌ Empty results (exit 0) vs actual errors (exit 1) properly handled
- ❌ Warning cases (exit 0 with stderr) need proper classification

### Recommended Error Detection Strategy

Since Flatpak uses only exit codes 0 and 1, proper error detection requires **stderr pattern matching**:

1. **Exit Code Analysis**:
   - 0 = Success or warnings
   - 1 = Parse stderr for specific error types

2. **Stderr Parsing Patterns**:
   - `not installed` → Package not found/installed
   - `No remote refs found` → Package not available in remotes
   - `Can't load uri` → Network/remote error
   - `Server returned status` → HTTP/network error
   - `Unknown option` → Invalid command option
   - `is not a flatpak command` → Invalid command

3. **Success Determination**:
   - Exit code 0 + no error patterns = Success
   - Exit code 0 + warning patterns = Success with warnings
   - Exit code 1 + error patterns = Specific error type

### Critical Implementation Requirements

**Flatpak requires simple binary error handling:**

```go
// For Flatpak - Current approach is mostly correct:
if result.ExitCode != 0 {
    // Parse stderr for specific error types
    if strings.Contains(stderr, "not installed") {
        return StatusPackageNotFound
    }
    if strings.Contains(stderr, "No remote refs found") {
        return StatusPackageNotFound
    }
    if strings.Contains(stderr, "Can't load uri") {
        return StatusNetworkError
    }
    return StatusGeneralError
}
// Exit code 0 = success (even with warnings)
return success()
```

### Permission Model

**Flatpak Permission Behavior:**
- Most read operations work as any user (list, search, info)
- Install/remove can work with `--user` flag for user installations
- System-wide installations typically require root or proper permissions
- **No specific exit codes for permission errors** - treated as general errors

## Testing Environment

All exit codes verified in clean Docker Ubuntu 22.04 environment:
```bash
docker run --rm ubuntu:22.04 bash -c "command_here"
```

**Test Date**: 2025-06-12
**Flatpak Version**: 1.12.7 (Ubuntu 22.04)
**Key Finding**: Flatpak uses simple binary exit codes (0=success, 1=error)
**Verification**: Systematic testing of all operations including network errors, package not found, invalid commands
