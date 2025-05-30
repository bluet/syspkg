#!/bin/bash
# Script to capture real package manager outputs for test fixtures

set -e

FIXTURE_DIR="testing/fixtures"
mkdir -p "$FIXTURE_DIR"/{apt,snap,flatpak,dnf,apk}

echo "Capturing package manager outputs for test fixtures..."

# APT (if available)
if command -v apt &> /dev/null; then
    echo "Capturing APT outputs..."
    apt search vim 2>/dev/null | head -50 > "$FIXTURE_DIR/apt/search-vim.txt" || true
    apt show vim 2>/dev/null > "$FIXTURE_DIR/apt/show-vim.txt" || true
    dpkg -l | head -20 > "$FIXTURE_DIR/apt/list-installed.txt" || true
    apt list --upgradable 2>/dev/null | head -20 > "$FIXTURE_DIR/apt/list-upgradable.txt" || true
fi

# SNAP (if available) - requires real system with snapd running
if command -v snap &> /dev/null && systemctl is-active snapd &>/dev/null; then
    echo "Capturing SNAP outputs..."
    snap find vim 2>/dev/null | head -20 > "$FIXTURE_DIR/snap/find-vim.txt" || true
    snap list 2>/dev/null > "$FIXTURE_DIR/snap/list.txt" || true
    snap info core 2>/dev/null > "$FIXTURE_DIR/snap/info-core.txt" || true
else
    echo "SNAP not available or snapd not running - using Docker for mock data"
    # We can't get real snap outputs from Docker, so we'll document sample outputs
    cat > "$FIXTURE_DIR/snap/find-vim.txt" << 'EOF'
Name           Version          Publisher     Notes  Summary
vim-editor     8.2.3995         jonathonf     -      Vi IMproved - enhanced vi editor
nvim           0.7.2            neovim        -      Vim-fork focused on extensibility
EOF

    cat > "$FIXTURE_DIR/snap/list.txt" << 'EOF'
Name                     Version          Rev    Tracking       Publisher   Notes
core20                   20230801         2015   latest/stable  canonical✓  base
core22                   20230801         864    latest/stable  canonical✓  base
snapd                    2.60.3           20290  latest/stable  canonical✓  snapd
EOF
fi

# FLATPAK (if available)
if command -v flatpak &> /dev/null; then
    echo "Capturing FLATPAK outputs..."
    flatpak search vim 2>/dev/null | head -20 > "$FIXTURE_DIR/flatpak/search-vim.txt" || true
    flatpak list 2>/dev/null > "$FIXTURE_DIR/flatpak/list.txt" || true
fi

echo "Fixture capture complete!"
echo "You can now use these files in your tests:"
find "$FIXTURE_DIR" -type f -name "*.txt" | sort