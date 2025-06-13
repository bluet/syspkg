# YUM Exit Codes Reference

This document contains comprehensive exit code mappings for YUM package manager operations, generated through systematic Docker testing in Rocky Linux 8.

## Exit Code Summary

| Exit Code | Category | Description |
|-----------|----------|-------------|
| **0** | Success | Successful operations |
| **1** | Error | ALL error conditions (invalid commands, package not found, permission denied) |

## Detailed Exit Code Analysis

### Exit Code 0 - Success Only

YUM returns exit code 0 only for successful operations.

**Success Cases:**
- Package found and installed/removed/upgraded successfully
- Search results returned (even if empty)
- Package lists updated successfully
- Repository metadata refreshed successfully
- Version information displayed
- Valid operations completed

### Exit Code 1 - All Error Conditions

**IMPORTANT**: YUM uses exit code 1 for ALL error conditions, making error differentiation dependent on stderr message parsing.

**Command Errors:**
- Invalid commands: `No such command: {command}. Please use /usr/bin/yum --help`
- Invalid options: Various option error messages

**Package Errors:**
- Package not found: `No match for argument: {package}` + `Error: Unable to find a match: {package}`
- Package not installed (for remove): `No Packages marked for removal`

**Permission Errors:**
- Root privileges required: `Error: This command has to be run with superuser privileges (under the root user on most systems).`

**Repository Errors:**
- Repository issues: Various repository and metadata error messages
- Network issues: Connection and download error messages

## Operation-Specific Behavior

### Search Operations
- **Command**: `yum search {query}`
- **Success**: Exit 0, results to stdout
- **No results**: Exit 0, may show "No matches found"
- **Invalid syntax**: Exit 1, usage message to stderr

### Install Operations
- **Command**: `yum install {packages} -y`
- **Success**: Exit 0, installation details to stdout
- **Package not found**: Exit 1, `No match for argument` + `Error: Unable to find a match` to stderr
- **Permission denied**: Exit 1, `Error: This command has to be run with superuser privileges` to stderr
- **Already installed**: Exit 0, shows "Nothing to do"

### Remove Operations
- **Command**: `yum remove {packages} -y`
- **Success**: Exit 0, removal details to stdout
- **Package not installed**: Exit 1, `No Packages marked for removal` to stderr
- **Permission denied**: Exit 1, `Error: This command has to be run with superuser privileges` to stderr

### Update Operations
- **Command**: `yum update -y` / `yum makecache`
- **Success**: Exit 0, update progress to stdout
- **Permission denied**: Exit 1, `Error: This command has to be run with superuser privileges` to stderr
- **Repository issues**: Exit 1, repository error messages to stderr

### Info Operations
- **Command**: `yum info {package}`
- **Success**: Exit 0, package details to stdout
- **Package not found**: Exit 1, downloads repo metadata, then shows `No matching Packages`

### List Operations
- **Command**: `yum list [installed]`
- **Success**: Exit 0, package list to stdout
- **Permission issues**: Exit 0, continues with available data

### Clean Operations
- **Command**: `yum clean all`
- **Success**: Exit 0, cleanup summary to stdout
- **Permission issues**: Exit 1, permission error messages

## Implementation Notes

### Current syspkg Implementation Analysis

**Correct Exit Code Handling:**
- ✅ Exit code 0 correctly mapped to success
- ✅ Exit code 1 correctly mapped to errors

**Error Detection Requirements:**

Since YUM uses exit code 1 for ALL errors, proper error detection requires **stderr pattern matching**:

1. **Exit Code Analysis**:
   - 0 = Success
   - 1 = Parse stderr for specific error types

2. **Stderr Parsing Patterns**:
   - `Error: This command has to be run with superuser privileges` → Permission error
   - `No match for argument:` + `Error: Unable to find a match:` → Package not found
   - `No Packages marked for removal` → Package not installed
   - `No such command:` → Invalid command

3. **Success Determination**:
   - Exit code 0 = Success
   - Exit code 1 + error patterns in stderr = Specific error type

### Critical Implementation Requirements

**YUM requires simple binary error handling with stderr parsing:**

```go
// For YUM - Correct approach:
if result.ExitCode == 1 {
    // Parse stderr for specific error types
    if strings.Contains(stderr, "superuser privileges") {
        return StatusPermissionError
    }
    if strings.Contains(stderr, "No match for argument") ||
       strings.Contains(stderr, "Unable to find a match") {
        return StatusPackageNotFound
    }
    if strings.Contains(stderr, "No Packages marked for removal") {
        return StatusPackageNotFound
    }
    if strings.Contains(stderr, "No such command") {
        return StatusUsageError
    }
    return StatusGeneralError // Default for exit 1
}
// Exit code 0 = success
return success()
```

## Testing Environment

All exit codes verified in clean Docker Rocky Linux 8 environment:
```bash
docker run --rm rockylinux:8 bash -c "command_here"
```

**Test Date**: 2025-06-12
**YUM Version**: 4.7.0 (Rocky Linux 8)
**Key Finding**: YUM uses simple binary exit codes (0=success, 1=all errors)
**Verification**: Systematic testing of all operations including permission errors, package not found, invalid commands
