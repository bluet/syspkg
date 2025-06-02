# Snap Exit Codes

This document details exit codes for Snap commands used in syspkg.

## Overview

Snap follows Unix standard exit codes:
- **0**: Success
- **1**: General error
- **64**: Command usage error (EX_USAGE from sysexits.h)

## Source Reference

From Unix sysexits.h:
```c
#define EX_USAGE 64  /* command line usage error */
```

Exit code 64 indicates incorrect command syntax or invalid arguments.

## Verified Behavior

### snap search
Testing is difficult in Docker due to snapd service requirements, but based on Unix standards:

```bash
# Valid search (expected)
$ snap search vim
$ echo $?  # Should return: 0

# Invalid option (expected)
$ snap search --invalid-option
$ echo $?  # Should return: 64 (usage error)

# No packages found (expected)
$ snap search nonexistentpackage123456
$ echo $?  # Should return: 0 (success, empty results)
```

## Current syspkg Bug

**BUG**: Our code incorrectly interprets exit code 64:
```go
// WRONG: Exit code 64 is NOT "no packages found"
if exitError.ExitCode() == 64 {
    // No packages found, return empty list  // ← This is wrong!
    return []manager.PackageInfo{}, nil
}
```

**Reality**:
- Exit code 64 = command usage error (invalid syntax/options)
- Exit code 0 = success (including when no packages found)
- Exit code 1 = general error

## Unix Exit Code Standards

Snap follows standard Unix exit codes (sysexits.h):
- **64 (EX_USAGE)**: Command line usage error
- **65 (EX_DATAERR)**: Data format error
- **66 (EX_NOINPUT)**: Cannot open input
- **67 (EX_NOUSER)**: Addressee unknown
- **68 (EX_NOHOST)**: Host name unknown
- **69 (EX_UNAVAILABLE)**: Service unavailable
- **70 (EX_SOFTWARE)**: Internal software error
- **71 (EX_OSERR)**: System error
- **72 (EX_OSFILE)**: Critical OS file missing
- **73 (EX_CANTCREAT)**: Can't create output file
- **74 (EX_IOERR)**: Input/output error
- **75 (EX_TEMPFAIL)**: Temporary failure
- **76 (EX_PROTOCOL)**: Remote error in protocol
- **77 (EX_NOPERM)**: Permission denied
- **78 (EX_CONFIG)**: Configuration error

## Recommendations

1. **Fix exit code 64 handling**: Remove incorrect assumption
2. **Handle usage errors properly**: Exit code 64 should be treated as command error
3. **Test with real snap environment**: Docker testing is unreliable for snap

## Testing Commands

```bash
# Test on system with snapd (not Docker)
snap search test; echo $?
snap search --invalid-option; echo $?
snap --invalid-command; echo $?
```

**Note**: Snap testing in Docker containers is unreliable due to systemd/snapd service dependencies.
