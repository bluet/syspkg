#!/bin/bash
# YUM fixture generation entrypoint for Rocky Linux/RHEL containers
# This script runs inside the container with root permissions and generates realistic YUM command outputs
# CRITICAL: Execution order matters - each operation changes system state!

set -e
export LC_ALL=C

echo "=== YUM Fixture Generation Started ==="
echo "Container: $(cat /etc/os-release | grep PRETTY_NAME | cut -d'"' -f2)"
echo "Working directory: $(pwd)"
echo "Fixtures will be saved to: /fixtures"
echo ""

# Ensure fixtures directory exists
mkdir -p /fixtures

echo "📋 Phase 1: CLEAN SYSTEM BASELINE (FIRST!) ..."
echo "  Capturing clean system state before ANY modifications..."

echo "  Version information..."
yum --version > /fixtures/version.clean-system.rocky-8.txt 2>/dev/null || true

echo "  Clean system search results..."
yum search vim > /fixtures/search-vim.clean-system.rocky-8.txt 2>/dev/null || true

echo "  Package info for available package..."
yum info vim > /fixtures/info-vim.clean-system.rocky-8.txt 2>/dev/null || true

echo "  Package info for non-existent package..."
yum info zzz9999nonexistent > /fixtures/info-zzz9999nonexistent.clean-system.rocky-8.txt 2>&1 || true

echo "  Installed packages list (minimal clean system)..."
yum list installed > /fixtures/list-installed.clean-system.rocky-8.txt 2>/dev/null || true

echo "  Available updates check..."
yum list updates > /fixtures/list-updates.clean-system.rocky-8.txt 2>/dev/null || true

echo "  Search for non-existent package..."
yum search zzz9999nonexistent > /fixtures/search-zzz9999nonexistent.clean-system.rocky-8.txt 2>&1 || true

echo "  Install dry-run for vim..."
yum install --setopt=tsflags=test vim > /fixtures/install-vim.dry-run.clean-system.rocky-8.txt 2>&1 || true

echo "  Install dry-run for multiple packages..."
yum install --setopt=tsflags=test nginx git htop > /fixtures/install-multiple.dry-run.clean-system.rocky-8.txt 2>&1 || true

echo "  Install dry-run for non-existent package..."
yum install --setopt=tsflags=test zzz9999nonexistent > /fixtures/install-zzz9999nonexistent.dry-run.clean-system.rocky-8.txt 2>&1 || true

echo "  ✅ Clean system baseline captured"
echo ""

echo "📦 Phase 2: INSTALL VIM AND VERIFY STATE CHANGE..."
echo "  Installing vim to create vim-installed state..."

echo "  Installing vim..."
yum install -y vim > /fixtures/install-vim.clean-system.rocky-8.txt 2>&1 || true

echo "  Verifying vim installation..."
yum list installed > /fixtures/list-installed.vim-installed.rocky-8.txt 2>/dev/null || true

echo "  Package info for installed vim..."
yum info vim > /fixtures/info-vim.vim-installed.rocky-8.txt 2>/dev/null || true

echo "  Attempting to install vim again (already installed)..."
yum install -y vim > /fixtures/install-vim.vim-already-installed.rocky-8.txt 2>&1 || true

echo "  Remove dry-run for vim..."
yum remove --setopt=tsflags=test vim > /fixtures/remove-vim.dry-run.vim-installed.rocky-8.txt 2>&1 || true

echo "  ✅ Vim installed state captured"
echo ""

echo "🔗 Phase 3: COMPLEX DEPENDENCIES SETUP..."
echo "  Installing packages with dependencies to test autoremove..."

echo "  Installing gcc-c++ (has many dependencies)..."
yum install -y gcc-c++ > /fixtures/install-gcc-cpp.vim-installed.rocky-8.txt 2>&1 || true

echo "  Installing curl for testing..."
yum install -y curl > /fixtures/install-curl.packages-installed.rocky-8.txt 2>&1 || true

echo "  Current packages state after multiple installs..."
yum list installed > /fixtures/list-installed.packages-installed.rocky-8.txt 2>/dev/null || true

echo "  Removing gcc-c++ but keeping dependencies (creates orphans)..."
yum remove -y --noautoremove gcc-c++ > /fixtures/remove-gcc-cpp.packages-installed.rocky-8.txt 2>&1 || true

echo "  ✅ Dependencies setup complete"
echo ""

echo "🧹 Phase 4: AUTOREMOVE WITH ACTUAL ORPHANS..."
echo "  Now we should have orphaned dependencies from gcc-c++..."

echo "  AutoRemove dry-run (should find orphans)..."
yum autoremove --setopt=tsflags=test > /fixtures/autoremove.dry-run.orphaned-packages.rocky-8.txt 2>&1 || true

echo "  AutoRemove execution (removes orphans)..."
yum autoremove -y > /fixtures/autoremove.orphaned-packages.rocky-8.txt 2>&1 || true

echo "  ✅ AutoRemove scenarios captured"
echo ""

echo "🔄 Phase 5: REMOVAL SCENARIOS..."
echo "  Testing various removal scenarios..."

echo "  Remove curl (actual removal)..."
yum remove -y curl > /fixtures/remove-curl.curl-installed.rocky-8.txt 2>&1 || true

echo "  Remove vim (actual removal)..."
yum remove -y vim > /fixtures/remove-vim.vim-installed.rocky-8.txt 2>&1 || true

echo "  Attempt to remove non-existent package..."
yum remove -y zzz9999nonexistent > /fixtures/remove-zzz9999nonexistent.clean-system.rocky-8.txt 2>&1 || true

echo "  ✅ Removal scenarios captured"
echo ""

echo "🚫 Phase 6: ERROR AND EDGE CASES..."
echo "  Capturing error scenarios and edge cases..."

echo "  Install non-existent package..."
yum install -y zzz9999nonexistent > /fixtures/install-zzz9999nonexistent.clean-system.rocky-8.txt 2>&1 || true

echo "  Update dry-run check..."
yum update --setopt=tsflags=test > /fixtures/update.dry-run.clean-system.rocky-8.txt 2>&1 || true

echo "  System maintenance operations..."
yum clean all > /fixtures/clean.clean-system.rocky-8.txt 2>&1 || true
yum makecache > /fixtures/makecache.clean-system.rocky-8.txt 2>&1 || true
yum check-update > /fixtures/check-update.clean-system.rocky-8.txt 2>&1 || true

echo "  ✅ Error scenarios captured"
echo ""

echo "🔧 Phase 7: Setting proper file ownership..."
echo "  Setting ownership to uid:gid 1000:1000 for host compatibility..."
chown -R 1000:1000 /fixtures
echo "  ✅ File ownership corrected"
echo ""

echo "=== YUM Fixture Generation Complete ==="
echo ""
echo "📊 Generated fixtures:"
find /fixtures -name "*.txt" -type f | sort | while read -r file; do
    lines=$(wc -l < "$file" 2>/dev/null || echo "0")
    size=$(stat -c%s "$file" 2>/dev/null || echo "0")
    echo "  $(basename "$file"): $lines lines, $size bytes"
done
echo ""
echo "💡 Fixtures generated with correct execution order:"
echo "💡 1. Clean system baseline (FIRST!)"
echo "💡 2. Systematic state transitions (install → verify → remove)"
echo "💡 3. Proper dependency setup for autoremove testing"
echo "💡 4. Authentic real-world YUM command outputs"
echo "💡 5. New naming convention: {operation}.{execution-mode}.{system-status}.{distro}-{version}.txt"
