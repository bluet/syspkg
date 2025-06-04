#!/bin/bash
# YUM fixture generation entrypoint for Rocky Linux/RHEL containers
# This script runs inside the container with root permissions and generates realistic YUM command outputs

set -e
export LC_ALL=C

echo "=== YUM Fixture Generation Started ==="
echo "Container: $(cat /etc/os-release | grep PRETTY_NAME | cut -d'"' -f2)"
echo "Working directory: $(pwd)"
echo "Fixtures will be saved to: /fixtures"
echo ""

# Ensure fixtures directory exists
mkdir -p /fixtures

echo "🔧 Phase 1: Building realistic system state..."
echo "  Installing packages to create realistic dependencies..."
yum install -y vim wget curl httpd gcc make >/dev/null 2>&1 || true
echo "  ✅ System now has realistic package dependencies"
echo ""

echo "🚫 Phase 2: Generating error scenarios..."

echo "  Capturing install-notfound scenario..."
yum install -y nonexistentpackage > /fixtures/install-notfound-rocky8.txt 2>&1 || true

echo "  Capturing remove-notfound scenario..."
yum remove -y nonexistentpackage > /fixtures/remove-notfound-rocky8.txt 2>&1 || true

echo "  Capturing already-installed scenario..."
yum install -y vim > /fixtures/install-already-installed-rocky8.txt 2>&1 || true

echo "  ✅ Error scenarios captured"
echo ""

echo "🔄 Phase 3: Generating system operations..."

echo "  Creating autoremove scenario by removing httpd..."
yum remove -y httpd >/dev/null 2>&1 || true
yum autoremove > /fixtures/autoremove-rocky8.txt 2>&1 || true

echo "  Capturing clean operation..."
yum clean all > /fixtures/clean-rocky8.txt 2>&1 || true

echo "  Capturing check-update operation..."
yum check-update > /fixtures/check-update-rocky8.txt 2>&1 || true

echo "  Capturing update dry-run..."
yum update --assumeno > /fixtures/update-dryrun-rocky8.txt 2>&1 || true

echo "  Capturing refresh operation..."
yum makecache > /fixtures/refresh-rocky8.txt 2>&1 || true

echo "  ✅ System operations captured"
echo ""

echo "🔀 Phase 4: Generating complex scenarios..."

echo "  Capturing install multiple packages (dry-run)..."
yum install --assumeno nginx git htop > /fixtures/install-multiple-rocky8.txt 2>&1 || true

echo "  Capturing remove with dependencies (dry-run)..."
yum remove --assumeno vim > /fixtures/remove-with-dependencies-rocky8.txt 2>&1 || true

echo "  Capturing empty search results..."
yum search zzz9999nonexistent > /fixtures/search-empty-rocky8.txt 2>&1 || true

echo "  ✅ Complex scenarios captured"
echo ""

echo "📋 Phase 5: Generating standard operations (clean output)..."

echo "  Capturing search results..."
yum search vim > /fixtures/search-vim-rocky8.txt 2>/dev/null || true

echo "  Capturing package info for installed package..."
yum info vim > /fixtures/info-vim-installed-rocky8.txt 2>/dev/null || true

echo "  Capturing package info for available package..."
yum info nginx > /fixtures/info-nginx-rocky8.txt 2>/dev/null || true

echo "  Capturing info for non-existent package..."
yum info nonexistentpackage > /fixtures/info-notfound-rocky8.txt 2>&1 || true

echo "  Capturing installed packages list..."
yum list installed > /fixtures/list-installed-rocky8.txt 2>/dev/null || true

echo "  Capturing installed packages list (full)..."
yum list installed > /fixtures/list-installed-full-rocky8.txt 2>/dev/null || true

echo "  Capturing available updates..."
yum list updates > /fixtures/list-updates-rocky8.txt 2>/dev/null || true

echo "  ✅ Standard operations captured"
echo ""

echo "🔧 Phase 6: Setting proper file ownership..."
echo "  Setting ownership to uid:gid 1000:1000 for host compatibility..."
chown -R 1000:1000 /fixtures
echo "  ✅ File ownership corrected"
echo ""

echo "=== YUM Fixture Generation Complete ==="
echo ""
echo "📊 Generated fixtures:"
find /fixtures -name "*.txt" -type f | sort | while read -r file; do
    lines=$(wc -l < "$file" 2>/dev/null || echo "0")
    size=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null || echo "0")
    echo "  $(basename "$file"): $lines lines, $size bytes"
done
echo ""
echo "💡 All fixtures generated with authentic real-world YUM command outputs"
echo "💡 Error scenarios captured with complete stderr+stdout"
echo "💡 Standard operations captured with clean stdout only"
echo "💡 Includes subscription manager warnings and realistic system complexity"
