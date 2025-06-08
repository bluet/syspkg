#!/bin/bash
# Simple fixture generation orchestrator using entrypoint.sh approach
# This script uses the brilliant entrypoint.sh pattern for clean, maintainable fixture generation

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
FIXTURES_DIR="${SCRIPT_DIR}/fixtures"
ENTRYPOINTS_DIR="${SCRIPT_DIR}/entrypoints"

echo "=== SysPkg Fixture Generation (Entrypoint Approach) ==="
echo "Using persistent container approach with mounted entrypoints"
echo "Project root: $PROJECT_ROOT"
echo "Fixtures dir: $FIXTURES_DIR"
echo ""

# Function to check Docker availability
check_docker() {
    echo "🐳 Checking Docker availability..."
    if ! command -v docker &> /dev/null; then
        echo "❌ Docker is required but not installed"
        exit 1
    fi

    if ! docker info >/dev/null 2>&1; then
        echo "❌ Docker daemon is not running"
        exit 1
    fi

    echo "✅ Docker is available"
    echo ""
}

# Function to generate APT fixtures
generate_apt_fixtures() {
    echo "🔧 Generating APT fixtures (Ubuntu 22.04)..."
    echo "  Using entrypoint: $ENTRYPOINTS_DIR/entrypoint-apt.sh"

    mkdir -p "${FIXTURES_DIR}/apt"

    docker run --rm \
        -v "${FIXTURES_DIR}/apt:/fixtures" \
        -v "${ENTRYPOINTS_DIR}/entrypoint-apt.sh:/entrypoint.sh:ro" \
        ubuntu:22.04 \
        /entrypoint.sh

    echo "  ✅ APT fixtures generated"
    echo ""
}

# Function to generate YUM fixtures
generate_yum_fixtures() {
    echo "🔧 Generating YUM fixtures (Rocky Linux 8)..."
    echo "  Using entrypoint: $ENTRYPOINTS_DIR/entrypoint-yum.sh"

    mkdir -p "${FIXTURES_DIR}/yum"

    docker run --rm \
        -v "${FIXTURES_DIR}/yum:/fixtures" \
        -v "${ENTRYPOINTS_DIR}/entrypoint-yum.sh:/entrypoint.sh:ro" \
        rockylinux:8 \
        /entrypoint.sh

    echo "  ✅ YUM fixtures generated"
    echo ""
}

# Function to generate DNF fixtures
generate_dnf_fixtures() {
    echo "🔧 Generating DNF fixtures (Fedora 39)..."
    echo "  Using entrypoint: $ENTRYPOINTS_DIR/entrypoint-dnf.sh"

    mkdir -p "${FIXTURES_DIR}/dnf"

    docker run --rm \
        -v "${FIXTURES_DIR}/dnf:/fixtures" \
        -v "${ENTRYPOINTS_DIR}/entrypoint-dnf.sh:/entrypoint.sh:ro" \
        fedora:39 \
        /entrypoint.sh

    echo "  ✅ DNF fixtures generated"
    echo ""
}

# Function to generate APK fixtures
generate_apk_fixtures() {
    echo "🔧 Generating APK fixtures (Alpine 3.18)..."
    echo "  Using entrypoint: $ENTRYPOINTS_DIR/entrypoint-apk.sh"

    mkdir -p "${FIXTURES_DIR}/apk"

    docker run --rm \
        -v "${FIXTURES_DIR}/apk:/fixtures" \
        -v "${ENTRYPOINTS_DIR}/entrypoint-apk.sh:/entrypoint.sh:ro" \
        alpine:3.18 \
        /entrypoint.sh

    echo "  ✅ APK fixtures generated"
    echo ""
}

# Function to generate Flatpak fixtures
generate_flatpak_fixtures() {
    echo "🔧 Generating Flatpak fixtures (Ubuntu 22.04 with Flatpak)..."
    echo "  Using entrypoint: $ENTRYPOINTS_DIR/entrypoint-flatpak.sh"

    mkdir -p "${FIXTURES_DIR}/flatpak"

    docker run --rm \
        -v "${FIXTURES_DIR}/flatpak:/fixtures" \
        -v "${ENTRYPOINTS_DIR}/entrypoint-flatpak.sh:/entrypoint.sh:ro" \
        ubuntu:22.04 \
        /entrypoint.sh

    echo "  ✅ Flatpak fixtures generated"
    echo ""
}

# Main execution
main() {
    check_docker

    # Generate fixtures based on GENERATE_ONLY environment variable
    case "${GENERATE_ONLY:-all}" in
        "apt")
            generate_apt_fixtures
            ;;
        "yum")
            generate_yum_fixtures
            ;;
        "dnf")
            generate_dnf_fixtures
            ;;
        "apk")
            generate_apk_fixtures
            ;;
        "flatpak")
            generate_flatpak_fixtures
            ;;
        *)
            # Generate fixtures for all package managers
            generate_apt_fixtures
            generate_yum_fixtures
            generate_dnf_fixtures
            generate_apk_fixtures
            generate_flatpak_fixtures
            ;;
    esac

    echo "=== Fixture Generation Complete ==="
    echo ""
    echo "📋 Summary by package manager:"

    for pm in apt yum dnf apk flatpak; do
        if [ -d "${FIXTURES_DIR}/$pm" ]; then
            count=$(find "${FIXTURES_DIR}/$pm" -name "*.txt" -type f 2>/dev/null | wc -l)
            echo "  $pm: $count fixtures"

            # Show some examples
            find "${FIXTURES_DIR}/$pm" -name "*.txt" -type f 2>/dev/null | head -3 | while read -r file; do
                lines=$(wc -l < "$file" 2>/dev/null || echo "0")
                echo "    $(basename "$file"): $lines lines"
            done
        fi
    done

    echo ""
    echo "💡 Key Benefits of Entrypoint Approach:"
    echo "  • Each container builds realistic system state internally"
    echo "  • No complex external orchestration with docker exec"
    echo "  • Proper file ownership (uid:gid 1000:1000) automatically handled"
    echo "  • Self-contained logic - easy to debug and maintain"
    echo "  • Atomic operations - clean success or failure"
    echo ""
    echo "🔍 Next steps:"
    echo "  make test                    # Test with new fixtures"
    echo "  make test-fixtures-validate  # Validate fixture quality"
    echo "  head testing/fixtures/apt/autoremove-apt.txt  # View a fixture"
}

# Script entry point
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
