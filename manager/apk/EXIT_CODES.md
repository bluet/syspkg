# APK Exit Codes Reference

This document contains comprehensive exit code mappings for APK package manager operations, generated through systematic Docker testing in Alpine Linux.

## Exit Code Summary

| Exit Code | Category | Description |
|-----------|----------|-------------|
| **0** | Success | Successful operations, package found, lists generated |
| **1** | General Error | Invalid commands, package not found, command parsing errors |
| **99** | Permission/Format Error | Database lock errors, permission denied, invalid package format |

## Detailed Exit Code Analysis

### Exit Code 0 - Success

APK returns exit code 0 for successful operations and valid commands that execute properly.

**Success Cases:**
- Version and help: `apk --version`
- Package operations: `apk add curl`, `apk del curl` (when successful)
- Information queries: `apk info busybox`, `apk list --installed`
- Repository operations: `apk update`, `apk upgrade`
- Search operations: `apk search vim` (even with warnings about missing repos)
- System operations: `apk stats`, `apk policy package`

**Success with Warnings (still exit code 0):**
- Repository warnings: `WARNING: Ignoring https://...` when repos unavailable
- Search/info operations when repository index is missing

### Exit Code 1 - General Errors

APK uses exit code 1 for command errors, package not found, and general operational failures.

**Command Errors:**
- Invalid commands: `apk invalidcommand`
- Invalid options: `apk --invalid-option`
- **Help command**: `apk --help` (unusually returns exit 1)

**Package Errors:**
- Package not found: `apk info nonexistent-package`
- Install non-existent package: `apk add nonexistent-package`
- Remove non-existent package: `apk del nonexistent-package`

### Exit Code 99 - Permission and Format Errors

APK uses exit code 99 for specific operational errors including permissions and invalid package formats.

**Permission Errors:**
- Database lock: `ERROR: Unable to lock database: Permission denied`
- Database access: `ERROR: Failed to open apk database: Permission denied`
- Operations requiring root: `apk update`, `apk add`, `apk del` (as non-root)

**Format Errors:**
- Invalid package names: `apk add 'package name with spaces'`
- Invalid world dependency format: `ERROR: ... is not a valid world dependency`

## Operation-Specific Behavior

### Search Operations
- **Command**: `apk search {query}`
- **Success**: Exit 0, results to stdout
- **No results**: Exit 0, no output
- **Repository issues**: Exit 0, warnings to stderr

### List Operations
- **Command**: `apk list [--installed]`
- **Success**: Exit 0, package list to stdout
- **Empty**: Exit 0, no output
- **Permission issues**: Exit 0, works for read operations

### Install Operations
- **Command**: `apk add {packages}`
- **Success**: Exit 0, installation details to stdout
- **Package not found**: Exit 1, "unable to select packages" to stderr
- **Permission denied**: Exit 99, "Permission denied" to stderr
- **Invalid format**: Exit 99, format error to stderr

### Remove Operations
- **Command**: `apk del {packages}`
- **Success**: Exit 0, removal details to stdout
- **Package not installed**: Exit 1, "No such package" to stderr
- **Permission denied**: Exit 99, "Permission denied" to stderr

### Update Operations
- **Command**: `apk update`
- **Success**: Exit 0, repository fetch progress to stdout
- **Permission denied**: Exit 99, "Permission denied" to stderr
- **Network issues**: Exit 0, continues with available repositories

### Upgrade Operations
- **Command**: `apk upgrade`
- **Success**: Exit 0, upgrade progress to stdout
- **Nothing to upgrade**: Exit 0, minimal output
- **Permission denied**: Exit 99, "Permission denied" to stderr

### Info Operations
- **Command**: `apk info {package}`
- **Success**: Exit 0, package details to stdout
- **Package not found**: Exit 1, no output (silent failure)
- **Repository warnings**: Exit 0, warnings to stderr

### Policy Operations
- **Command**: `apk policy {package}`
- **Success**: Exit 0, policy information to stdout
- **Package exists**: Exit 0, shows repository priorities

### Stats Operations
- **Command**: `apk stats`
- **Success**: Exit 0, statistics to stdout
- **Always available**: Works as any user

## Implementation Notes

### Current syspkg Implementation Analysis

**Exit Code Handling:**
- ✅ Exit code 0 correctly mapped to success
- ✅ Exit code 1 correctly mapped to general error
- ❌ Exit code 99 needs specific handling for permission/format errors

**Issues Found:**
- ❌ Exit code 99 not recognized as permission error
- ❌ Help command returns exit 1 but should be treated as success
- ❌ Package not found detection needs stderr parsing (exit 1)
- ❌ Permission errors (exit 99) need specific classification

### Recommended Error Detection Strategy

APK has cleaner exit code semantics than other package managers:

1. **Exit Code Analysis**:
   - 0 = Success (may have warnings)
   - 1 = General error (parse stderr for specifics)
   - 99 = Permission/format error

2. **Stderr Parsing Patterns**:
   - `unable to select packages` → Package not found
   - `No such package` → Package not installed
   - `Permission denied` → Permission error
   - `not a valid world dependency` → Invalid package format
   - `is not an apk command` → Invalid command

3. **Special Cases**:
   - `apk --help` returns exit 1 but should be treated as success
   - Repository warnings (exit 0) are informational, not errors

### Critical Implementation Requirements

**APK requires three-tier exit code handling:**

```go
// For APK - Enhanced approach needed:
if result.ExitCode == 99 {
    // Permission or format errors
    if strings.Contains(stderr, "Permission denied") {
        return StatusPermissionError
    }
    if strings.Contains(stderr, "not a valid world dependency") {
        return StatusUsageError
    }
    return StatusPermissionError // Default for exit 99
} else if result.ExitCode == 1 {
    // General errors - parse stderr for specifics
    if strings.Contains(stderr, "unable to select packages") {
        return StatusPackageNotFound
    }
    if strings.Contains(stderr, "No such package") {
        return StatusPackageNotFound
    }
    if strings.Contains(stderr, "is not an apk command") {
        return StatusUsageError
    }
    return StatusGeneralError
}
// Exit code 0 = success (even with warnings)
return success()
```

### Permission Model

**APK Permission Behavior:**
- Read operations (list, info, search, stats, policy) work as any user
- Write operations (add, del, update, upgrade) require root privileges
- **Specific exit code 99 for permission errors** - very helpful for classification
- Database lock conflicts also return exit 99

### Unique APK Characteristics

**Different from other package managers:**
- **Three distinct exit codes** (0, 1, 99) vs binary (APT, Flatpak) or always-0 (YUM)
- **Help command returns exit 1** - unusual but consistent behavior
- **Repository warnings don't cause failures** - robust network handling
- **Clear permission error detection** via exit code 99

## Testing Environment

All exit codes verified in clean Docker Alpine Linux environment:
```bash
docker run --rm alpine:latest sh -c "command_here"
```

**Test Date**: 2025-06-12
**APK Version**: 2.12.10-2.12.14 (Alpine Linux)
**Key Finding**: APK has clean three-tier exit codes (0=success, 1=general error, 99=permission/format error)
**Verification**: Systematic testing of all operations including permission scenarios, package errors, and invalid formats
