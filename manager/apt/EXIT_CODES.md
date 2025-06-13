# APT Exit Codes Reference

This document contains comprehensive exit code mappings for APT package manager operations, generated through systematic Docker testing.

## Exit Code Summary

| Exit Code | Category | Description |
|-----------|----------|-------------|
| **0** | Success | Command executed successfully |
| **100** | Error | ALL error conditions (package not found, permission denied, invalid commands) |

## Detailed Exit Code Analysis

### Exit Code 0 - Success Only

APT returns exit code 0 only for successful operations.

**Success Cases:**
- Package found and installed/removed/upgraded successfully
- Search results returned (even if empty)
- Package lists updated successfully
- Cache cleaned successfully
- Version information displayed
- Help displayed

### Exit Code 100 - All Error Conditions

**IMPORTANT**: APT uses exit code 100 for ALL error conditions, making error differentiation dependent on stderr message parsing.

**Package Errors:**
- Package not found: `E: Unable to locate package {name}`
- Package not installed (for remove): Various error messages

**Permission Errors:**
- Lock file access: `E: Could not open lock file /var/lib/dpkg/lock-frontend - open (13: Permission denied)`
- Root privileges: `E: Unable to acquire the dpkg frontend lock (/var/lib/dpkg/lock-frontend), are you root?`

**Command Errors:**
- Invalid commands: `E: Invalid operation {command}`
- Invalid options: `E: Command line option --invalid is not understood`

**System Errors:**
- Repository issues: Various network/repository error messages
- Dependency conflicts: Dependency resolution error messages

## Operation-Specific Behavior

### Search Operations
- **Command**: `apt search {query}`
- **Success**: Exit 0, results to stdout
- **No results**: Exit 0, empty results
- **Invalid syntax**: Exit 100, error message to stderr

### List Operations
- **Command**: `apt list [--installed|--upgradable]`
- **Success**: Exit 0, package list to stdout
- **Invalid options**: Exit 100, error message to stderr

### Install Operations
- **Command**: `apt install {packages} -y`
- **Success**: Exit 0, installation details to stdout
- **Package not found**: Exit 100, `E: Unable to locate package` to stderr
- **Permission denied**: Exit 100, lock file error to stderr
- **Already installed**: Exit 0, "already newest version" message

### Remove Operations
- **Command**: `apt remove {packages} -y`
- **Success**: Exit 0, removal details to stdout
- **Package not installed**: Exit 100, error message to stderr
- **Permission denied**: Exit 100, lock file error to stderr

### Update Operations
- **Command**: `apt update`
- **Success**: Exit 0, update progress to stdout
- **Permission denied**: Exit 100, lock file error to stderr
- **Network issues**: Exit 100, network error messages to stderr

### Upgrade Operations
- **Command**: `apt upgrade -y`
- **Success**: Exit 0, upgrade details to stdout
- **No packages to upgrade**: Exit 0, "0 upgraded" message
- **Permission denied**: Exit 100, lock file error to stderr

### Clean Operations
- **Command**: `apt clean` / `apt autoclean`
- **Success**: Exit 0, minimal output
- **Permission denied**: Exit 100, lock file error to stderr

### AutoRemove Operations
- **Command**: `apt autoremove -y`
- **Success**: Exit 0, removal details to stdout
- **Nothing to remove**: Exit 0, "0 to remove" message
- **Permission denied**: Exit 100, lock file error to stderr

### Package Info Operations
- **Command**: `apt-cache show {package}`
- **Success**: Exit 0, package details to stdout
- **Package not found**: Exit 100, error message to stderr

## Implementation Notes

### Current syspkg Implementation Analysis

**Correct Exit Code Handling:**
- ✅ Exit code 100 correctly mapped to errors
- ✅ Exit code 0 correctly mapped to success

**Error Detection Requirements:**

Since APT uses exit code 100 for ALL errors, proper error detection requires **stderr pattern matching**:

1. **Exit Code Analysis**:
   - 0 = Success
   - 100 = Parse stderr for specific error types

2. **Stderr Parsing Patterns**:
   - `E: Unable to locate package` → Package not found
   - `E: Could not open lock file` → Permission error
   - `E: Unable to acquire.*lock` → Permission error
   - `E: Invalid operation` → Invalid command
   - `E: Command line option.*not understood` → Invalid option

3. **Success Determination**:
   - Exit code 0 = Success
   - Exit code 100 + error patterns in stderr = Specific error type

### Critical Implementation Requirements

**APT requires simple binary error handling with stderr parsing:**

```go
// For APT - Correct approach:
if result.ExitCode == 100 {
    // Parse stderr for specific error types
    if strings.Contains(stderr, "Unable to locate package") {
        return StatusPackageNotFound
    }
    if strings.Contains(stderr, "Could not open lock file") ||
       strings.Contains(stderr, "Unable to acquire.*lock") {
        return StatusPermissionError
    }
    if strings.Contains(stderr, "Invalid operation") {
        return StatusUsageError
    }
    return StatusGeneralError // Default for exit 100
}
// Exit code 0 = success
return success()
```

## Testing Environment

All exit codes verified in clean Docker Ubuntu 22.04 environment:
```bash
docker run --rm ubuntu:22.04 bash -c "command_here"
```

**Test Date**: 2025-06-12
**APT Version**: 2.4.13 (Ubuntu 22.04)
**Key Finding**: APT uses simple binary exit codes (0=success, 100=all errors)
**Verification**: Systematic testing of all operations including permission scenarios, package errors, and invalid commands
