#!/bin/bash
# APK fixture generation entrypoint for Alpine Linux containers
# This script runs inside the container with root permissions and generates realistic APK command outputs

set -e
export LC_ALL=C

echo "=== APK Fixture Generation Started ==="
echo "Container: $(cat /etc/os-release | grep PRETTY_NAME | cut -d'"' -f2)"
echo "Working directory: $(pwd)"
echo "Fixtures will be saved to: /fixtures"
echo ""

# Ensure fixtures directory exists
mkdir -p /fixtures

echo "🔧 Phase 1: Building realistic system state..."
echo "  Updating package index..."
apk update >/dev/null 2>&1

echo "  Installing packages to create realistic dependencies..."
apk add vim wget curl nginx gcc make >/dev/null 2>&1 || true
echo "  ✅ System now has realistic package dependencies"
echo ""

echo "🚫 Phase 2: Generating error scenarios..."

echo "  Capturing add-notfound scenario..."
apk add nonexistentpackage > /fixtures/add-notfound-alpine318.txt 2>&1 || true

echo "  Capturing del-notfound scenario..."
apk del nonexistentpackage > /fixtures/del-notfound-alpine318.txt 2>&1 || true

echo "  Capturing already-installed scenario..."
apk add vim > /fixtures/add-already-alpine318.txt 2>&1 || true

echo "  ✅ Error scenarios captured"
echo ""

echo "🔄 Phase 3: Generating system operations..."

echo "  Creating orphan cleanup scenario by removing nginx..."
apk del nginx >/dev/null 2>&1 || true
# APK doesn't have autoremove, but we can show normal del operation
apk del --dry-run > /fixtures/autoremove-alpine318.txt 2>&1 || true

echo "  Capturing cache clean operation..."
apk cache clean > /fixtures/clean-alpine318.txt 2>&1 || true

echo "  Capturing update operation..."
apk update > /fixtures/update-alpine318.txt 2>&1 || true

echo "  Capturing upgrade dry-run..."
apk upgrade --simulate > /fixtures/upgrade-dryrun-alpine318.txt 2>&1 || true

echo "  ✅ System operations captured"
echo ""

echo "🔀 Phase 4: Generating complex scenarios..."

echo "  Capturing add multiple packages (dry-run)..."
apk add --simulate git htop python3 > /fixtures/add-multiple-alpine318.txt 2>&1 || true

echo "  Capturing del with dependencies (dry-run)..."
apk del --simulate vim > /fixtures/del-with-dependencies-alpine318.txt 2>&1 || true

echo "  Capturing empty search results..."
apk search zzz9999nonexistent > /fixtures/search-empty-alpine318.txt 2>&1 || true

echo "  ✅ Complex scenarios captured"
echo ""

echo "📋 Phase 5: Generating standard operations (clean output)..."

echo "  Capturing search results..."
apk search vim > /fixtures/search-vim-alpine318.txt 2>/dev/null || true

echo "  Capturing package info for installed package..."
apk info vim > /fixtures/info-vim-installed-alpine318.txt 2>/dev/null || true

echo "  Capturing package info for available package..."
apk info git > /fixtures/info-git-alpine318.txt 2>/dev/null || true

echo "  Capturing info for non-existent package..."
apk info nonexistentpackage > /fixtures/info-notfound-alpine318.txt 2>&1 || true

echo "  Capturing installed packages list..."
apk list --installed > /fixtures/list-installed-alpine318.txt 2>/dev/null || true

echo "  Capturing available packages..."
apk list --available > /fixtures/list-available-alpine318.txt 2>/dev/null || true

echo "  Capturing upgradable packages..."
apk list --upgradable > /fixtures/list-upgradable-alpine318.txt 2>/dev/null || true

echo "  ✅ Standard operations captured"
echo ""

echo "🔧 Phase 6: Setting proper file ownership..."
echo "  Setting ownership to uid:gid 1000:1000 for host compatibility..."
chown -R 1000:1000 /fixtures
echo "  ✅ File ownership corrected"
echo ""

echo "=== APK Fixture Generation Complete ==="
echo ""
echo "📊 Generated fixtures:"
ls /fixtures/*.txt 2>/dev/null | sort | while read -r file; do
    lines=$(wc -l < "$file" 2>/dev/null || echo "0")
    size=$(stat -c%s "$file" 2>/dev/null || echo "0")
    echo "  $(basename "$file"): $lines lines, $size bytes"
done
echo ""
echo "💡 All fixtures generated with authentic real-world APK command outputs"
echo "💡 Error scenarios captured with complete stderr+stdout"
echo "💡 Standard operations captured with clean stdout only"
echo "💡 APK uses 'add/del' instead of 'install/remove' terminology"
