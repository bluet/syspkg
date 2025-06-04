#!/bin/bash

# Generate APT fixtures with native status indicators
echo "Generating APT fixtures with native status..."

FIXTURES_DIR="$(dirname "$0")/apt"
mkdir -p "$FIXTURES_DIR"

# Generate search fixture with mixed status (installed and available)
docker run --rm ubuntu:22.04 bash -c "
apt update -qq > /dev/null 2>&1
apt install -qq -y vim vim-common vim-runtime > /dev/null 2>&1
apt search vim 2>/dev/null | head -50
" > "$FIXTURES_DIR/search-vim-with-status.txt"

# Generate upgradable list fixture
docker run --rm ubuntu:22.04 bash -c "
apt update -qq > /dev/null 2>&1
apt install -qq -y vim=2:8.2.3995-1ubuntu2.23 vim-common=2:8.2.3995-1ubuntu2.23 vim-runtime=2:8.2.3995-1ubuntu2.23 --allow-downgrades > /dev/null 2>&1 || true
apt list --upgradable 2>/dev/null | grep -E 'vim|apt|gpgv' | head -10
" > "$FIXTURES_DIR/list-upgradable-with-status.txt"

# Generate install output fixture
docker run --rm ubuntu:22.04 bash -c "
apt update -qq > /dev/null 2>&1
apt install -y vim 2>&1 | grep -E 'Setting up|Reading|Unpacking|Get:|The following|Processing'
" > "$FIXTURES_DIR/install-vim-verbose.txt"

echo "APT fixtures generated:"
echo "  - $FIXTURES_DIR/search-vim-with-status.txt"
echo "  - $FIXTURES_DIR/list-upgradable-with-status.txt"
echo "  - $FIXTURES_DIR/install-vim-verbose.txt"
