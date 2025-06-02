# Flatpak Exit Codes

This document details exit codes for Flatpak commands used in syspkg.

## Overview

Flatpak uses standard Unix exit codes with some special cases:
- **0**: Success
- **1**: General error (most common failure)
- **42**: Special case for flatpak-builder (json unchanged)
- **256**: Script execution failures

## Source Reference

No comprehensive official documentation found. Information gathered from:
- GitHub issues in flatpak/flatpak repository
- Community reports and bug reports
- Error messages in production systems

## Known Exit Codes

### Standard Codes
- **0**: Success (operation completed successfully)
- **1**: General error (installation failed, package not found, etc.)

### Special Codes
- **42**: Used by flatpak-builder when "json is unchanged since last build"
- **256**: Script failures (apply_extra script failed, ldconfig failed)

### Observed Behavior

From GitHub issues and community reports:

```bash
# Package not found (expected)
$ flatpak search nonexistentpackage123456
$ echo $?  # Expected: 0 or 1 (not well documented)

# Installation error (common)
$ flatpak install nonexistent-app
$ echo $?  # Returns: 1

# Script failures (ChromeOS specific)
# "ldconfig failed, exit status 256"
# "apply_extra script failed, exit status 256"
```

## Current syspkg Implementation

Our code assumes exit code 1 means "no packages found":
```go
if exitError.ExitCode() == 1 {
    // No packages found, return empty list
    return []manager.PackageInfo{}, nil
}
```

This may be correct, but needs verification through testing.

## Known Issues

1. **Missing Documentation**: Flatpak lacks comprehensive exit code documentation
2. **Platform Variations**: Different behavior on different systems (ChromeOS, Linux distributions)
3. **Script Dependencies**: Exit codes may vary based on external script failures

## Testing Challenges

Flatpak testing can be complex due to:
- Repository configuration requirements
- Runtime dependencies
- Platform-specific behaviors
- Service dependencies

## Recommendations

1. **Test thoroughly**: Verify exit code 1 behavior for search
2. **Document actual behavior**: Don't assume based on other package managers
3. **Handle script errors**: Be prepared for exit code 256 scenarios
4. **Monitor GitHub issues**: Watch flatpak/flatpak for exit code discussions

## Testing Commands

```bash
# Test in system with Flatpak configured
flatpak search test; echo $?
flatpak search nonexistentpackage123456; echo $?
flatpak install --dry-run nonexistent; echo $?

# Test in Docker (if Flatpak available)
docker run --rm fedora:latest bash -c 'dnf install -y flatpak && flatpak search test; echo $?'
```

**Note**: Flatpak behavior may vary significantly between different Linux distributions and configurations.
