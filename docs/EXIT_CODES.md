# Package Manager Exit Codes Overview

This document provides a high-level overview of exit code behavior across package managers.
**For detailed information, see the EXIT_CODES.md file in each package manager directory.**

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

1. **Never use generic exit code helpers** - each PM is unique
2. **Test actual behavior** - documentation can be wrong
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
4. Never rely on generic patterns

## Bugs to Fix

1. **APT**: Remove incorrect handling of exit code 100 as "no packages found"
2. **Snap**: Remove incorrect handling of exit code 64 as "no packages found"
3. **All PMs**: Verify and document actual exit code behavior
