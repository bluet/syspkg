#!/bin/bash
# Zypper fixture generation entrypoint for openSUSE containers
# This script runs inside the container with root permissions and generates realistic Zypper command outputs

set -e
export LC_ALL=C

echo "=== Zypper Fixture Generation Started ==="
echo "Container: $(cat /etc/os-release | grep PRETTY_NAME | cut -d'"' -f2)"
echo "Working directory: $(pwd)"
echo "Fixtures will be saved to: /fixtures"
echo ""

# Ensure fixtures directory exists
mkdir -p /fixtures

echo "🧹 Phase 0: Clean system capture..."
echo "  Refreshing repositories..."
zypper refresh >/dev/null 2>&1

echo "  Capturing clean search results (before installs)..."
zypper search vim > /fixtures/search-vim-clean-opensuse154.txt 2>/dev/null || true
zypper search zzz9999nonexistent > /fixtures/search-empty-opensuse154.txt 2>&1 || true

echo "  Capturing clean list operations..."
zypper list-installed > /fixtures/list-installed-clean-opensuse154.txt 2>/dev/null || true
zypper list-updates > /fixtures/list-updates-clean-opensuse154.txt 2>/dev/null || true

echo "  ✅ Clean system state captured"
echo ""

echo "🔧 Phase 1: Building realistic system state..."
echo "  Installing packages to create realistic dependencies..."
zypper install -y vim wget curl httpd gcc make >/dev/null 2>&1 || true
echo "  ✅ System now has realistic package dependencies"
echo ""

echo "🚫 Phase 2: Generating error scenarios..."

echo "  Capturing install-notfound scenario..."
zypper install -y nonexistentpackage > /fixtures/install-notfound-opensuse154.txt 2>&1 || true

echo "  Capturing remove-notfound scenario..."
zypper remove -y nonexistentpackage > /fixtures/remove-notfound-opensuse154.txt 2>&1 || true

echo "  Capturing already-installed scenario..."
zypper install -y vim > /fixtures/install-already-opensuse154.txt 2>&1 || true

echo "  ✅ Error scenarios captured"
echo ""

echo "🔄 Phase 3: Generating system operations..."

echo "  Creating orphan cleanup scenario by removing httpd..."
zypper remove -y httpd >/dev/null 2>&1 || true
# Zypper doesn't have autoremove, but we can show package cleanup
zypper packages --orphaned > /fixtures/autoremove-opensuse154.txt 2>&1 || true

echo "  Capturing clean operation..."
zypper clean -a > /fixtures/clean-opensuse154.txt 2>&1 || true

echo "  Capturing check-update operation..."
zypper list-updates > /fixtures/check-update-opensuse154.txt 2>&1 || true

echo "  Capturing update dry-run..."
zypper update --dry-run > /fixtures/update-dryrun-opensuse154.txt 2>&1 || true

echo "  Capturing refresh operation..."
zypper refresh > /fixtures/refresh-opensuse154.txt 2>&1 || true

echo "  ✅ System operations captured"
echo ""

echo "🔀 Phase 4: Generating complex scenarios..."

echo "  Capturing install multiple packages (dry-run)..."
zypper install --dry-run nginx git htop > /fixtures/install-multiple-opensuse154.txt 2>&1 || true

echo "  Capturing remove with dependencies (dry-run)..."
zypper remove --dry-run vim > /fixtures/remove-with-dependencies-opensuse154.txt 2>&1 || true

echo "  Capturing nginx search results..."
zypper search nginx > /fixtures/search-nginx-opensuse154.txt 2>/dev/null || true

echo "  Capturing actual package installs/removes..."
zypper install -y nginx > /fixtures/install-nginx-opensuse154.txt 2>&1 || true
zypper install -y vim > /fixtures/install-vim-opensuse154.txt 2>&1 || true
zypper remove -y nginx > /fixtures/remove-nginx-opensuse154.txt 2>&1 || true
zypper install -y tree > /dev/null 2>&1
zypper remove -y tree > /fixtures/remove-tree-opensuse154.txt 2>&1 || true

echo "  Capturing update all dry-run..."
zypper update --dry-run > /fixtures/update-all-dryrun-opensuse154.txt 2>&1 || true

echo "  ✅ Complex scenarios captured"
echo ""

echo "📋 Phase 5: Generating mixed-status operations (after installs)..."

echo "  Capturing mixed search results (with installed indicators)..."
zypper search vim > /fixtures/search-vim-mixed-opensuse154.txt 2>/dev/null || true

echo "  Capturing package info for installed package..."
zypper info vim > /fixtures/info-vim-installed-opensuse154.txt 2>/dev/null || true

echo "  Capturing package info for available package..."
zypper info nginx > /fixtures/info-nginx-opensuse154.txt 2>/dev/null || true

echo "  Capturing package info (general vim info)..."
zypper info vim > /fixtures/info-vim-opensuse154.txt 2>/dev/null || true

echo "  Capturing info for non-existent package..."
zypper info nonexistentpackage > /fixtures/info-notfound-opensuse154.txt 2>&1 || true

echo "  Capturing installed packages list (with new packages)..."
zypper list-installed > /fixtures/list-installed-opensuse154.txt 2>/dev/null || true

echo "  Capturing installed packages list (full)..."
zypper list-installed > /fixtures/list-installed-full-opensuse154.txt 2>/dev/null || true

echo "  Capturing available updates (may include new package updates)..."
zypper list-updates > /fixtures/list-updates-opensuse154.txt 2>/dev/null || true

echo "  ✅ Standard operations captured"
echo ""

echo "🔧 Phase 6: Setting proper file ownership..."
echo "  Setting ownership to uid:gid 1000:1000 for host compatibility..."
chown -R 1000:1000 /fixtures
echo "  ✅ File ownership corrected"
echo ""

echo "=== Zypper Fixture Generation Complete ==="
echo ""
echo "📊 Generated fixtures:"
find /fixtures -name "*.txt" -type f | sort | while read -r file; do
    lines=$(wc -l < "$file" 2>/dev/null || echo "0")
    size=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null || echo "0")
    echo "  $(basename "$file"): $lines lines, $size bytes"
done
echo ""
echo "💡 All fixtures generated with authentic real-world Zypper command outputs"
echo "💡 Error scenarios captured with complete stderr+stdout"
echo "💡 Standard operations captured with clean stdout only"
echo "💡 Zypper uses different command syntax than YUM/DNF (list-installed vs list installed)"
