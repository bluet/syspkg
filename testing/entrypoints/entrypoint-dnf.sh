#!/bin/bash
# DNF fixture generation entrypoint for Fedora containers
# This script runs inside the container with root permissions and generates realistic DNF command outputs

set -e
export LC_ALL=C

echo "=== DNF Fixture Generation Started ==="
echo "Container: $(cat /etc/os-release | grep PRETTY_NAME | cut -d'"' -f2)"
echo "Working directory: $(pwd)"
echo "Fixtures will be saved to: /fixtures"
echo ""

# Ensure fixtures directory exists
mkdir -p /fixtures

echo "🔧 Phase 1: Building realistic system state..."
echo "  Installing packages to create realistic dependencies..."
dnf install -y vim wget curl httpd gcc make >/dev/null 2>&1 || true
echo "  ✅ System now has realistic package dependencies"
echo ""

echo "🚫 Phase 2: Generating error scenarios..."

echo "  Capturing install-notfound scenario..."
dnf install -y nonexistentpackage > /fixtures/install-notfound-fedora39.txt 2>&1 || true

echo "  Capturing remove-notfound scenario..."
dnf remove -y nonexistentpackage > /fixtures/remove-notfound-fedora39.txt 2>&1 || true

echo "  Capturing already-installed scenario..."
dnf install -y vim > /fixtures/install-already-installed-fedora39.txt 2>&1 || true

echo "  ✅ Error scenarios captured"
echo ""

echo "🔄 Phase 3: Generating system operations..."

echo "  Creating autoremove scenario by removing httpd..."
dnf remove -y httpd >/dev/null 2>&1 || true
dnf autoremove > /fixtures/autoremove-fedora39.txt 2>&1 || true

echo "  Capturing clean operation..."
dnf clean all > /fixtures/clean-fedora39.txt 2>&1 || true

echo "  Capturing check-update operation..."
dnf check-update > /fixtures/check-update-fedora39.txt 2>&1 || true

echo "  Capturing update dry-run..."
dnf update --assumeno > /fixtures/update-dryrun-fedora39.txt 2>&1 || true

echo "  Capturing refresh operation..."
dnf makecache > /fixtures/refresh-fedora39.txt 2>&1 || true

echo "  ✅ System operations captured"
echo ""

echo "🔀 Phase 4: Generating complex scenarios..."

echo "  Capturing install multiple packages (dry-run)..."
dnf install --assumeno nginx git htop > /fixtures/install-multiple-fedora39.txt 2>&1 || true

echo "  Capturing remove with dependencies (dry-run)..."
dnf remove --assumeno vim > /fixtures/remove-with-dependencies-fedora39.txt 2>&1 || true

echo "  Capturing empty search results..."
dnf search zzz9999nonexistent > /fixtures/search-empty-fedora39.txt 2>&1 || true

echo "  Capturing nginx search results..."
dnf search nginx > /fixtures/search-nginx-fedora39.txt 2>/dev/null || true

echo "  Capturing actual package installs/removes..."
dnf install -y nginx > /fixtures/install-nginx-fedora39.txt 2>&1 || true
dnf install -y vim > /fixtures/install-vim-fedora39.txt 2>&1 || true
dnf remove -y nginx > /fixtures/remove-nginx-fedora39.txt 2>&1 || true
dnf install -y tree > /dev/null 2>&1
dnf remove -y tree > /fixtures/remove-tree-fedora39.txt 2>&1 || true

echo "  Capturing update all dry-run..."
dnf update --assumeno > /fixtures/update-all-dryrun-fedora39.txt 2>&1 || true

echo "  ✅ Complex scenarios captured"
echo ""

echo "📋 Phase 5: Generating standard operations (clean output)..."

echo "  Capturing search results..."
dnf search vim > /fixtures/search-vim-fedora39.txt 2>/dev/null || true

echo "  Capturing package info for installed package..."
dnf info vim > /fixtures/info-vim-installed-fedora39.txt 2>/dev/null || true

echo "  Capturing package info for available package..."
dnf info nginx > /fixtures/info-nginx-fedora39.txt 2>/dev/null || true

echo "  Capturing package info (general vim info)..."
dnf info vim > /fixtures/info-vim-fedora39.txt 2>/dev/null || true

echo "  Capturing info for non-existent package..."
dnf info nonexistentpackage > /fixtures/info-notfound-fedora39.txt 2>&1 || true

echo "  Capturing installed packages list..."
dnf list installed > /fixtures/list-installed-fedora39.txt 2>/dev/null || true

echo "  Capturing installed packages list (full)..."
dnf list installed > /fixtures/list-installed-full-fedora39.txt 2>/dev/null || true

echo "  Capturing available updates..."
dnf list updates > /fixtures/list-updates-fedora39.txt 2>/dev/null || true

echo "  ✅ Standard operations captured"
echo ""

echo "🔧 Phase 6: Setting proper file ownership..."
echo "  Setting ownership to uid:gid 1000:1000 for host compatibility..."
chown -R 1000:1000 /fixtures
echo "  ✅ File ownership corrected"
echo ""

echo "=== DNF Fixture Generation Complete ==="
echo ""
echo "📊 Generated fixtures:"
find /fixtures -name "*.txt" -type f | sort | while read -r file; do
    lines=$(wc -l < "$file" 2>/dev/null || echo "0")
    size=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null || echo "0")
    echo "  $(basename "$file"): $lines lines, $size bytes"
done
echo ""
echo "💡 All fixtures generated with authentic real-world DNF command outputs"
echo "💡 Error scenarios captured with complete stderr+stdout"
echo "💡 Standard operations captured with clean stdout only"
