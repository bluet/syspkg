# Test Fixtures

This directory contains real command outputs from various package managers for testing purposes. These fixtures ensure that our parsers work correctly with actual package manager outputs across different operating systems.

## Directory Structure

```
testing/fixtures/
├── apt/           # APT package manager (Ubuntu/Debian)
├── yum/           # YUM package manager (RHEL/Rocky/CentOS)
├── dnf/           # DNF package manager (Fedora)
├── snap/          # Snap package manager (Universal)
├── flatpak/       # Flatpak package manager (Universal)
├── apk/           # APK package manager (Alpine Linux)
└── README.md      # This file
```

## Testing Principles

1. **Fixtures are Primary Test Data**
   - Use real fixture files from this directory in unit tests
   - Fixtures contain full raw outputs from actual package managers
   - Inline mocks are for quick tests and edge cases only

2. **Fixture Content**
   - All fixtures contain unmodified, authentic command outputs
   - They reflect real-world package states and formats
   - Include various scenarios: available, installed, upgradable packages

3. **Usage in Tests**
   ```go
   // Use testutil helper functions
   fixture := testutil.LoadAPTFixture(t, "search-vim.txt")
   packages := parseSearchOutput(fixture)
   ```

## Package Manager Coverage

### APT (Ubuntu/Debian) - ✅ Complete
- **search-vim.txt** - Basic search output (no native status)
- **search-vim-with-status.txt** - Search with native status indicators (`[installed]`)
- **list-upgradable-with-status.txt** - Upgradable packages with version info
- **list-installed.txt** - Installed packages listing
- **show-vim.txt** - Package information details
- **apt-install-vim.txt** - Installation command output
- **apt-remove-vim.txt** - Removal command output
- **dpkg-query-mixed-status.txt** - Mixed package states

### YUM (RHEL/Rocky/CentOS) - ✅ Complete Implementation with Comprehensive Testing
26+ fixtures with full parser testing coverage (100% parser coverage achieved):
- Search operations (vim, nginx, empty results) ✅ Tested
- Info operations (installed, available, not found) ✅ Tested  
- Install operations (new, multiple, already installed, not found) ✅ Tested
- Remove operations (success, not found, dependencies) ✅ Tested
- List operations (installed, updates, minimal) ✅ Tested
- System operations (check-update, refresh, clean, autoremove) ✅ Tested
- Security validation (input validation, command injection prevention) ✅ Complete

### DNF (Fedora) - ⚠️ Basic Coverage
- **search-vim-fedora39.txt** - Basic search
- **info-vim-fedora39.txt** - Package info
- **list-installed-fedora39.txt** - Installed packages
- *Missing*: Install/remove operations, error scenarios

### Snap (Universal) - ✅ Good Coverage
- **find-vim.txt** - Search results with multiple vim-related snaps
- **info-core.txt** - Package information
- **list.txt** - Installed snaps listing

### Flatpak (Universal) - ⚠️ Limited Coverage
- **search-vim.txt** - Empty search results (legacy)
- **search-neovim.txt** - Successful search results
- **list.txt** - Installed flatpaks
- *Missing*: Install/remove operations, detailed info

### APK (Alpine Linux) - ⚠️ Basic Coverage
- **search-vim-alpine.txt** - Search results
- **info-vim-alpine.txt** - Package information
- **list-installed-alpine.txt** - Installed packages
- *Missing*: Install/remove operations, error scenarios

## Fixture Generation

### Automated Generation
```bash
# Generate all fixtures using Docker
make test-fixtures

# Generate specific APT fixtures
./testing/fixtures/generate-apt-fixtures.sh
```

### Manual Generation Commands

#### APT (Ubuntu/Debian)
```bash
docker run --rm ubuntu:22.04 bash -c "
apt update -qq && apt search vim 2>/dev/null
" > testing/fixtures/apt/search-vim.txt
```

#### YUM (Rocky Linux)
```bash
docker run --rm rockylinux:8 bash -c "
yum search vim 2>/dev/null
" > testing/fixtures/yum/search-vim-rocky8.txt
```

#### DNF (Fedora)
```bash
docker run --rm fedora:39 bash -c "
dnf search vim 2>/dev/null
" > testing/fixtures/dnf/search-vim-fedora39.txt
```

## Fixture Quality Standards

### High Quality Fixtures
- **YUM**: 26+ fixtures with 100% parser coverage and comprehensive testing (✅ Complete)
- **APT**: Complete coverage with native status indicators and security validation (✅ Complete)

### Good Quality
- **Snap**: Covers main operations, realistic content
- **APT**: Basic operations well covered

### Needs Improvement
- **DNF**: Missing install/remove operations and error scenarios
- **APK**: Missing install/remove operations and error scenarios  
- **Flatpak**: Very limited, mostly empty results

## Contributing New Fixtures

1. **Use Docker for Safety**
   ```bash
   # ALWAYS use Docker to avoid affecting your system
   docker run --rm <image> bash -c "<command>" > fixture.txt
   ```

2. **Follow Naming Convention**
   - `{operation}.{execution-mode}.{system-status}.{distro}-{version}.txt`
   - Examples: 
     - `install-vim.dry-run.clean-system.ubuntu-2204.txt`
     - `dpkg-status-vim.vim-installed.ubuntu-2204.txt`
     - `search-vim.clean-system.rocky-8.txt` (normal mode, execution-mode omitted)

3. **Include Various Scenarios**
   - Successful operations
   - Error cases (package not found, already installed)
   - Edge cases (empty results, special characters)

4. **Test Your Fixtures**
   ```go
   func TestNewFixture(t *testing.T) {
       fixture := testutil.LoadAPTFixture(t, "your-new-fixture.txt")
       // Test parsing logic
   }
   ```

## Docker Safety and Realistic Fixture Generation

**IMPORTANT**: Always use Docker for fixture generation to protect your development system from package manager operations.

### ❌ Wrong Approach: Clean Container Per Command
```bash
# WRONG - each command runs in fresh container, missing realistic scenarios
docker run --rm ubuntu:22.04 bash -c "apt install vim" > fixture.txt
docker run --rm ubuntu:22.04 bash -c "apt autoremove --dry-run" > fixture.txt  # Empty result!
```

**Problems with `--rm` approach:**
- No accumulated packages to autoremove
- No realistic "already installed" scenarios
- No dependency conflicts or complex system states
- Missing real-world messiness

### ✅ Best Approach: Entrypoint Scripts with Mounted Volumes
```bash
# BEST - use entrypoint scripts that build realistic state internally
docker run --rm \
  -v ./fixtures/apt:/fixtures \
  -v ./entrypoints/entrypoint-apt.sh:/entrypoint.sh:ro \
  ubuntu:22.04 /entrypoint.sh
```

**Benefits of Entrypoint Approach:**
- **Self-contained logic** - all operations happen inside container
- **No external orchestration** - no complex `docker exec` chains  
- **Atomic operations** - succeed completely or fail cleanly
- **Easy debugging** - simple, linear script execution
- **Proper file ownership** - automatically handled with `chown 1000:1000`
- **Maintainable** - one script per package manager, easy to modify

### ⚠️ Alternative: Persistent Container (More Complex)
```bash
# ACCEPTABLE but more complex - external orchestration approach
docker run -d --name fixture-gen ubuntu:22.04 sleep 3600

# Build realistic system with dependencies  
docker exec fixture-gen bash -c "apt update -qq"
docker exec fixture-gen bash -c "apt install -y vim python3-pip curl build-essential"

# Now capture authentic scenarios
docker exec fixture-gen bash -c "apt install -y vim" > install-already-installed.txt 2>&1
docker exec fixture-gen bash -c "apt remove -y python3-pip && apt autoremove --dry-run" > autoremove.txt 2>&1
docker exec fixture-gen bash -c "apt remove --dry-run vim" > remove-with-deps.txt 2>&1

# Clean up
docker stop fixture-gen && docker rm fixture-gen
```

### Stream Redirection Considerations

**Avoid `2>&1` unless needed:**
- `2>&1` mixes stderr with stdout
- Warnings and errors contaminate normal output
- Can affect parser behavior

**Use `2>&1` only when:**
- You want to capture error scenarios
- Testing error handling paths
- Need complete command output including warnings

## Maintenance

1. **Regular Updates**: Update fixtures when package managers change output formats
2. **Version Tracking**: Include OS version in fixture names for clarity
3. **Completeness**: Ensure each package manager has fixtures for all supported operations
4. **Authentication**: All fixtures contain real, unmodified command outputs

For more information about testing strategy, see [CONTRIBUTING.md](../../CONTRIBUTING.md).