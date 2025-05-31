# APT Exit Codes

This document details exit codes for APT (Advanced Package Tool) commands used in syspkg.

## Overview

APT uses a simple binary exit code system:
- **0**: Success
- **100**: Any error occurred
- **1**: Special case (apt run without options)

## Source Reference

From Debian APT source code (`cmdline/apt.cc`):
- Returns 100 on failure, 0 on success
- Multiple `exit(100)` statements throughout codebase for various errors
- No specific exit codes for different error types

## Verified Behavior

### apt search
```bash
# No packages found - returns SUCCESS
$ apt search nonexistentpackage123456
$ echo $?  # Returns: 0

# Invalid option - returns ERROR
$ apt search --invalid-option
$ echo $?  # Returns: 100
```

### apt-get commands
- `apt-get update` failures: **100**
- `apt-get install` failures: **100**
- Repository errors: **100**
- Network errors: **100**

## Current syspkg Bug

**BUG**: Our code incorrectly assumes:
```go
// WRONG: APT search does NOT return 100 for "no packages found"
if exitError.ExitCode() == 100 {
    // No packages found, return empty list
    return []manager.PackageInfo{}, nil
}
```

**Reality**:
- APT search returns **0** when no packages found
- APT search returns **100** only on actual errors (invalid options, etc.)

## dpkg Exit Codes

APT also uses `dpkg-query` which has different exit codes:

- **0**: Success
- **1**: Package not found (normal condition)
- **2**: Serious error

Example:
```go
// Current code in utils.go handles this correctly:
if exitErr.ExitCode() != 1 && !strings.Contains(string(out), "no packages found matching") {
    return nil, fmt.Errorf("command failed with output: %s", string(out))
}
```

## Recommendations

1. **Fix search exit code handling**: Remove incorrect 100 handling
2. **Test each command**: APT uses different tools with different codes
3. **No generic helpers**: APT behavior is unique

## Testing Commands

```bash
# Test in Ubuntu container
docker run --rm ubuntu:22.04 bash -c 'apt search test; echo $?'
docker run --rm ubuntu:22.04 bash -c 'apt search --invalid; echo $?'
```
