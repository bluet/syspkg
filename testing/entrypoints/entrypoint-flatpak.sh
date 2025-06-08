#!/bin/bash
# Flatpak fixture generation entrypoint for Ubuntu containers with Flatpak
# This script runs inside the container with root permissions and generates realistic Flatpak command outputs

set -e
export DEBIAN_FRONTEND=noninteractive
export LC_ALL=C

echo "=== Flatpak Fixture Generation Started ==="
echo "Container: $(cat /etc/os-release | grep PRETTY_NAME | cut -d'"' -f2)"
echo "Working directory: $(pwd)"
echo "Fixtures will be saved to: /fixtures"
echo ""

# Ensure fixtures directory exists
mkdir -p /fixtures

echo "🔧 Phase 0: Installing and setting up Flatpak..."
echo "  Installing Flatpak package..."
apt update -qq
apt install -y flatpak

echo "  Adding Flathub repository..."
# Use --user to avoid sudo issues and system-wide changes
flatpak remote-add --user --if-not-exists flathub https://flathub.org/repo/flathub.flatpakrepo >/dev/null 2>&1 || true

echo "  Updating app metadata..."
flatpak update --user --appstream >/dev/null 2>&1 || true

echo "  ✅ Flatpak setup complete"
echo ""

echo "🧹 Phase 1: Clean system capture (before installations)..."

echo "  Capturing version information..."
flatpak --version > /fixtures/version-ubuntu2204.txt 2>/dev/null || true

echo "  Capturing clean list operations..."
flatpak list --user --columns=name,version,origin > /fixtures/list-installed-clean-ubuntu2204.txt 2>/dev/null || true
flatpak list --user --updates --columns=name,version,origin > /fixtures/list-upgradable-clean-ubuntu2204.txt 2>/dev/null || true

echo "  Capturing search operations..."
flatpak search vim > /fixtures/search-vim-ubuntu2204.txt 2>/dev/null || true
flatpak search neovim > /fixtures/search-neovim-ubuntu2204.txt 2>/dev/null || true
flatpak search editor > /fixtures/search-editor-ubuntu2204.txt 2>/dev/null || true
flatpak search nonexistentapp12345 > /fixtures/search-empty-ubuntu2204.txt 2>/dev/null || true

echo "  ✅ Clean system state captured"
echo ""

echo "🔧 Phase 2: Installing sample applications..."
echo "  Installing test applications (this may take time)..."

# Install a lightweight app for testing - use --user to avoid system changes
echo "    Installing org.gnome.Calculator..."
flatpak install --user -y flathub org.gnome.Calculator >/dev/null 2>&1 || echo "    Calculator install failed (expected in CI)"

echo "    Installing io.github.shiftey.Desktop..."
flatpak install --user -y flathub io.github.shiftey.Desktop >/dev/null 2>&1 || echo "    Desktop install failed (expected in CI)"

echo "  ✅ Sample applications installed (where possible)"
echo ""

echo "🚫 Phase 3: Generating error scenarios..."

echo "  Capturing install-notfound scenario..."
flatpak install --user -y flathub nonexistent.app.NotReal > /fixtures/install-notfound-ubuntu2204.txt 2>&1 || true

echo "  Capturing remove-notfound scenario..."
flatpak uninstall --user -y nonexistent.app.NotReal > /fixtures/remove-notfound-ubuntu2204.txt 2>&1 || true

echo "  Capturing info-notfound scenario..."
flatpak info nonexistent.app.NotReal > /fixtures/info-notfound-ubuntu2204.txt 2>&1 || true

echo "  ✅ Error scenarios captured"
echo ""

echo "🔄 Phase 4: Generating system operations..."

echo "  Capturing update operations..."
flatpak update --user --appstream > /fixtures/update-appstream-ubuntu2204.txt 2>&1 || true
flatpak update --user --dry-run > /fixtures/update-dryrun-ubuntu2204.txt 2>&1 || true

echo "  Capturing clean operations..."
flatpak uninstall --user --unused --dry-run > /fixtures/clean-dryrun-ubuntu2204.txt 2>&1 || true

echo "  Capturing autoremove operations..."
flatpak uninstall --user --unused > /fixtures/autoremove-ubuntu2204.txt 2>&1 || true

echo "  ✅ System operations captured"
echo ""

echo "📋 Phase 5: Generating post-install scenarios..."

echo "  Capturing list operations after installs..."
flatpak list --user --columns=name,version,origin > /fixtures/list-installed-ubuntu2204.txt 2>/dev/null || true
flatpak list --user --updates --columns=name,version,origin > /fixtures/list-upgradable-ubuntu2204.txt 2>/dev/null || true

echo "  Capturing info operations..."
flatpak info org.gnome.Calculator > /fixtures/info-calculator-ubuntu2204.txt 2>/dev/null || true

echo "  Capturing install already-installed scenario..."
flatpak install --user -y flathub org.gnome.Calculator > /fixtures/install-already-ubuntu2204.txt 2>&1 || true

echo "  Capturing remove operations..."
flatpak uninstall --user --dry-run org.gnome.Calculator > /fixtures/remove-dryrun-ubuntu2204.txt 2>&1 || true

echo "  ✅ Post-install scenarios captured"
echo ""

echo "🔀 Phase 6: Generating complex scenarios..."

echo "  Capturing install multiple packages..."
flatpak install --user --dry-run flathub org.gnome.TextEditor org.gnome.Calculator > /fixtures/install-multiple-ubuntu2204.txt 2>&1 || true

echo "  Capturing upgrade scenarios..."
flatpak update --user --dry-run org.gnome.Calculator > /fixtures/upgrade-specific-ubuntu2204.txt 2>&1 || true
flatpak update --user --dry-run > /fixtures/upgrade-all-ubuntu2204.txt 2>&1 || true

echo "  Capturing remote information..."
flatpak remotes --user > /fixtures/remotes-ubuntu2204.txt 2>/dev/null || true

echo "  ✅ Complex scenarios captured"
echo ""

echo "🔧 Phase 7: Setting proper file ownership..."
echo "  Setting ownership to uid:gid 1000:1000 for host compatibility..."
chown -R 1000:1000 /fixtures
echo "  ✅ File ownership corrected"
echo ""

echo "=== Flatpak Fixture Generation Complete ==="
echo ""
echo "📊 Generated fixtures:"
find /fixtures -name "*.txt" -type f | sort | while read -r file; do
    lines=$(wc -l < "$file" 2>/dev/null || echo "0")
    size=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null || echo "0")
    echo "  $(basename "$file"): $lines lines, $size bytes"
done
echo ""
echo "💡 All fixtures generated with authentic real-world Flatpak command outputs"
echo "💡 Error scenarios captured with complete stderr+stdout"
echo "💡 Standard operations captured from actual Flatpak environment"
echo "💡 Note: Some install operations may fail in CI/containerized environments"
