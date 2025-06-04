#!/bin/bash
# APT fixture generation entrypoint for Ubuntu/Debian containers
# This script runs inside the container with root permissions and generates realistic APT command outputs

set -e
export DEBIAN_FRONTEND=noninteractive
export LC_ALL=C

echo "=== APT Fixture Generation Started ==="
echo "Container: $(cat /etc/os-release | grep PRETTY_NAME | cut -d'"' -f2)"
echo "Working directory: $(pwd)"
echo "Fixtures will be saved to: /fixtures"
echo ""

# Ensure fixtures directory exists
mkdir -p /fixtures

echo "🧹 Phase 0: Clean system capture..."
echo "  Updating package lists..."
apt update -qq

echo "  Capturing clean search results (before installs)..."
apt search vim > /fixtures/search-vim-clean-ubuntu2204.txt 2>/dev/null || true
apt search zzz9999nonexistent > /fixtures/search-empty-ubuntu2204.txt 2>&1 || true

echo "  Capturing clean list operations..."
apt list --installed > /fixtures/list-installed-clean-ubuntu2204.txt 2>/dev/null || true
apt list --upgradable > /fixtures/list-upgradable-clean-ubuntu2204.txt 2>/dev/null || true

echo "  ✅ Clean system state captured"
echo ""

echo "🔧 Phase 1: Building realistic system state..."
echo "  Installing packages to create realistic dependencies..."
apt install -y vim python3-pip curl build-essential >/dev/null 2>&1
echo "  ✅ System now has realistic package dependencies"
echo ""

echo "🚫 Phase 2: Generating error scenarios..."

echo "  Capturing install-notfound scenario..."
apt install -y nonexistentpackage > /fixtures/install-notfound-ubuntu2204.txt 2>&1 || true

echo "  Capturing remove-notfound scenario..."
apt remove -y nonexistentpackage > /fixtures/remove-notfound-ubuntu2204.txt 2>&1 || true

echo "  Capturing already-installed scenario..."
apt install -y vim > /fixtures/install-already-ubuntu2204.txt 2>&1 || true

echo "  ✅ Error scenarios captured"
echo ""

echo "🔄 Phase 3: Generating system operations..."

echo "  Creating autoremove scenario by removing python3-pip..."
apt remove -y python3-pip >/dev/null 2>&1
apt autoremove --dry-run > /fixtures/autoremove-ubuntu2204.txt 2>&1 || true

echo "  Capturing clean operation..."
apt clean > /fixtures/clean-ubuntu2204.txt 2>&1 || true

echo "  Capturing update operation..."
apt update > /fixtures/update-ubuntu2204.txt 2>&1 || true

echo "  Capturing upgrade dry-run..."
apt upgrade --dry-run > /fixtures/upgrade-dryrun-ubuntu2204.txt 2>&1 || true

echo "  ✅ System operations captured"
echo ""

echo "🔀 Phase 4: Generating complex scenarios..."

echo "  Capturing install multiple packages (dry-run)..."
apt install --dry-run curl wget htop nginx > /fixtures/install-multiple-ubuntu2204.txt 2>&1 || true

echo "  Capturing remove with dependencies (dry-run)..."
apt remove --dry-run vim > /fixtures/remove-with-dependencies-ubuntu2204.txt 2>&1 || true

echo "  ✅ Complex scenarios captured"
echo ""

echo "📋 Phase 5: Generating mixed-status operations (after installs)..."

echo "  Capturing mixed search results (with [installed] indicators)..."
apt search vim > /fixtures/search-vim-mixed-ubuntu2204.txt 2>/dev/null || true

echo "  Capturing package info..."
apt show vim > /fixtures/show-vim-ubuntu2204.txt 2>/dev/null || true

echo "  Capturing installed packages list (with new packages)..."
apt list --installed > /fixtures/list-installed-ubuntu2204.txt 2>/dev/null || true

echo "  Capturing upgradable packages (may include new package updates)..."
apt list --upgradable > /fixtures/list-upgradable-ubuntu2204.txt 2>/dev/null || true

echo "  ✅ Standard operations captured"
echo ""

echo "🔧 Phase 6: Setting proper file ownership..."
echo "  Setting ownership to uid:gid 1000:1000 for host compatibility..."
chown -R 1000:1000 /fixtures
echo "  ✅ File ownership corrected"
echo ""

echo "=== APT Fixture Generation Complete ==="
echo ""
echo "📊 Generated fixtures:"
find /fixtures -name "*.txt" -type f | sort | while read -r file; do
    lines=$(wc -l < "$file" 2>/dev/null || echo "0")
    size=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null || echo "0")
    echo "  $(basename "$file"): $lines lines, $size bytes"
done
echo ""
echo "💡 All fixtures generated with authentic real-world APT command outputs"
echo "💡 Error scenarios captured with complete stderr+stdout"
echo "💡 Standard operations captured with clean stdout only"
