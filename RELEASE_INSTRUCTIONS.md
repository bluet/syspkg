# Release Instructions for v0.1.6

This document provides step-by-step instructions for creating and publishing release tag v0.1.6.

## Prerequisites

- [ ] All changes have been merged to the `main` branch
- [ ] CHANGELOG.md has been updated with v0.1.6 release notes
- [ ] All tests pass on the main branch
- [ ] Build succeeds on all supported platforms

## Release Checklist

### 1. Verify Current State
```bash
# Ensure you're on the main branch and it's up to date
git checkout main
git pull origin main

# Verify the build and tests pass
make test
make build
```

### 2. Create the Release Tag

Create an annotated tag for version 0.1.6:

```bash
# Create annotated tag (recommended for releases)
git tag -a v0.1.6 -m "Release v0.1.6 - Add YUM package manager support"

# Verify the tag was created
git tag -l v0.1.6

# View tag details
git show v0.1.6
```

### 3. Push the Tag to GitHub

```bash
# Push the tag to the remote repository
git push origin v0.1.6
```

**Note**: Pushing the tag will automatically trigger the `release-binaries.yml` GitHub Actions workflow, which will build binaries for all supported platforms (Linux, Windows, macOS) and architectures (amd64, arm64, 386).

### 4. Create GitHub Release

After pushing the tag, create a GitHub Release:

1. Go to https://github.com/bluet/syspkg/releases/new
2. Select the tag: `v0.1.6`
3. Set the release title: **v0.1.6 - YUM Package Manager Support**
4. Copy the release notes from CHANGELOG.md section for v0.1.6
5. Check "Set as the latest release"
6. Click "Publish release"

The GitHub Actions workflow will automatically attach the compiled binaries to the release.

### 5. Verify the Release

After the release is published:

1. Check that the release appears at https://github.com/bluet/syspkg/releases
2. Verify that binaries are attached to the release (this may take a few minutes)
3. Test installation with `go install`:
   ```bash
   go install github.com/bluet/syspkg/cmd/syspkg@v0.1.6
   ```
4. Verify the installed version includes YUM support:
   ```bash
   syspkg --yum --help
   ```

## What's New in v0.1.6

### Major Features
- **YUM Package Manager Support**: Full implementation for RHEL/CentOS/Rocky Linux/AlmaLinux
- **ARM64 Architecture Support**: Docker testing on Apple Silicon
- **Enhanced Testing Infrastructure**: Multi-OS Docker testing
- **Comprehensive Documentation**: Architecture, exit codes, and contributing guides

### Package Manager Support Status
- ✅ APT (Debian/Ubuntu)
- ✅ YUM (RHEL/CentOS/Rocky/AlmaLinux) - **NEW in v0.1.6**
- ✅ Snap (Universal)
- ✅ Flatpak (Universal)

## Troubleshooting

### If the tag already exists
```bash
# Delete local tag
git tag -d v0.1.6

# Delete remote tag (use with caution!)
git push origin :refs/tags/v0.1.6
```

### If release binaries fail to build
Check the GitHub Actions workflow logs at:
https://github.com/bluet/syspkg/actions/workflows/release-binaries.yml

## Additional Notes

- The release follows [Semantic Versioning](https://semver.org/)
- The CHANGELOG follows [Keep a Changelog](https://keepachangelog.com/) format
- The Apache 2.0 license provides patent protection and enterprise clarity
- Users can lock to this version with `go get github.com/bluet/syspkg@v0.1.6`

## Post-Release Tasks

- [ ] Announce the release in project channels
- [ ] Update any external documentation referencing the version
- [ ] Close related GitHub issues (e.g., the issue requesting this release)
- [ ] Consider updating pkg.go.dev documentation if needed
