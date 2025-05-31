# YUM Exit Codes

This document details exit codes for YUM (Yellowdog Updater Modified) commands used in syspkg.

## Overview

YUM uses different exit codes depending on the operation:
- **0**: Success (general operations)
- **1**: Error occurred
- **100**: Special meaning for `check-update` (updates available - SUCCESS!)

## Source Reference

From Red Hat documentation:
- "yum check-update exit code 100" is documented as "packages available for update"
- Exit code 100 is **not an error** for check-update command
- Other commands follow standard 0=success, 1=error pattern

## Verified Behavior

### yum check-update (Special Case)
```bash
# No updates available
$ yum check-update
$ echo $?  # Returns: 0

# Updates available - THIS IS SUCCESS!
$ yum check-update
[... lists available updates ...]
$ echo $?  # Returns: 100

# Error occurred
$ yum check-update (with network issues)
$ echo $?  # Returns: 1
```

**Critical**: Exit code 100 means **SUCCESS** for check-update!

### Other YUM commands
- `yum search`: 0=success, 1=error
- `yum install`: 0=success, 1=error
- `yum info`: 0=success, 1=error

## Current syspkg Implementation

**CORRECT**: Our code properly handles check-update:
```go
// YUM check-update returns exit code 100 when updates are available
if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 100 {
    // Exit code 100 means updates are available, continue parsing
} else {
    // Other exit codes indicate real errors
    return nil, err
}
```

## Key Differences from APT

⚠️ **CRITICAL**: Same exit code, opposite meaning!
- **APT**: 100 = error
- **YUM**: 100 = success (for check-update)

This is why generic exit code helpers would be dangerous!

## rpm Integration

YUM also uses `rpm` commands for status detection:
- `rpm -q package`: 0=installed, 1=not installed
- Used in `Find()` method for accurate status detection

## Testing Commands

```bash
# Test in Rocky Linux container
docker run --rm rockylinux:8 bash -c 'yum check-update; echo $?'
docker run --rm rockylinux:8 bash -c 'yum search test; echo $?'
docker run --rm rockylinux:8 bash -c 'rpm -q bash; echo $?'
```

## Recommendations

1. **Document special behaviors**: check-update is unique
2. **Never generalize**: YUM exit codes are command-specific
3. **Test thoroughly**: Different containers have different update availability
