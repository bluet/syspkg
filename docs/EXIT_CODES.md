# Package Manager Exit Codes Overview

This document provides a high-level overview of exit code behavior across package managers.
**For detailed information, see the EXIT_CODES.md file in each package manager directory.**

## 📖 Related Documentation

- **[../README.md](../README.md)** - Project overview
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Technical design and interfaces
- **[../CONTRIBUTING.md](../CONTRIBUTING.md)** - Development workflow and testing guide
- **[../manager/apt/EXIT_CODES.md](../manager/apt/EXIT_CODES.md)** - APT-specific exit codes
- **[../manager/yum/EXIT_CODES.md](../manager/yum/EXIT_CODES.md)** - YUM-specific exit codes

## Critical Insights

### Exit Codes Are NOT Consistent
| Package Manager | "No Error" | "Updates Available" | "No Packages Found" | "Usage Error" |
|----------------|------------|-------------------|-------------------|---------------|
| **APT**        | 0          | N/A               | 0 (success)       | 100           |
| **YUM**        | 0          | 100 (success!)    | 0 (success)       | 1             |
| **Snap**       | 0          | N/A               | 0 (success)       | 64            |
| **Flatpak**    | 0          | N/A               | 1 (needs verify)  | 1             |

### Dangerous Assumptions

⚠️ **Same exit code, opposite meanings:**
- APT: 100 = error
- YUM: 100 = success (updates available)

⚠️ **Our code has bugs:**
- APT: Assumes 100 = "no packages found" (WRONG - it's an error!)
- Snap: Assumes 64 = "no packages found" (WRONG - it's usage error!)

## Key Principles

1. **Never use generic exit code helpers** - each PM is unique // WRONG: generic helpers
2. **Test actual behavior** - documentation can be wrong // BAD: trust docs only
3. **Each PM uses different tools** - APT uses both `apt` and `dpkg-query`
4. **Verify through testing** - not just documentation

## Documentation Structure

- **Central overview**: This file (cross-PM comparison)
- **Detailed docs**: `manager/{pm}/EXIT_CODES.md` (PM-specific behavior)

## For Package Manager Implementation

When implementing Option C (CommandBuilder), each package manager must:
1. Handle its own exit codes specifically
2. Document actual behavior (not assumptions)
3. Test thoroughly in real environments
4. Never rely on generic patterns // WRONG: generic patterns

## Bugs to Fix

1. **APT**: Remove incorrect handling of exit code 100 as "no packages found"
2. **Snap**: Remove incorrect handling of exit code 64 as "no packages found"
3. **All PMs**: Verify and document actual exit code behavior

## SysPkg CLI Exit Codes

The syspkg CLI tool uses standardized exit codes following POSIX and sysexits.h conventions:

### Exit Code Reference

| Code | Constant | Meaning | Example |
|------|----------|---------|---------|
| **0** | `ExitSuccess` | Success | `syspkg search vim` (found packages) |
| **1** | `ExitGeneralError` | General errors | Network issues, unknown failures |
| **2** | `ExitUsageError` | Invalid arguments (POSIX) | `syspkg` (no command), `syspkg invalid-cmd` |
| **77** | `ExitNoPermission` | Permission denied (sysexits.h) | `syspkg install vim` (needs sudo) |
| **69** | `ExitUnavailable` | Service unavailable (sysexits.h) | `syspkg --yum install` (on Ubuntu) |
| **130** | `ExitSignalInt` | SIGINT (Ctrl+C) | User interrupted operation |

### Usage Examples

```bash
# Script can handle specific error cases
syspkg install vim
case $? in
    0)   echo "Success" ;;
    1)   echo "General error - check logs" ;;
    2)   echo "Usage error - check command syntax" ;;
    77)  echo "Permission denied - try with sudo" ;;
    69)  echo "Package manager not available" ;;
    130) echo "Interrupted by user" ;;
esac
```

### Error Output

- **Usage errors** → stderr (with exit code 2)
- **Help output** → stdout (with exit code 0)
- **General errors** → stderr (with appropriate exit code)

### For Plugin Developers

When developing package manager plugins, ensure your error messages contain recognizable patterns for proper exit code classification:

- **Permission errors**: Include "permission denied", "are you root", "try with sudo", "access denied"
- **Unavailable errors**: Include "not found", "not available", "unavailable"
- **Usage errors**: Include "requires", "invalid", "usage"

See [PLUGIN_DEVELOPMENT.md](PLUGIN_DEVELOPMENT.md#error-handling) for detailed error handling guidelines.
