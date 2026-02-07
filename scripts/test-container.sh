#!/usr/bin/env bash
# Test kyaraben in isolated containers using Podman
#
# Usage:
#   ./scripts/test-container.sh           # Quick tests (no Nix)
#   ./scripts/test-container.sh quick     # Quick tests (no Nix)
#   ./scripts/test-container.sh nix       # Full Nix E2E tests (slower)
#   ./scripts/test-container.sh shell     # Drop into quick test container
#   ./scripts/test-container.sh nix-shell # Drop into Nix E2E container

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

IMAGE_QUICK="kyaraben-test"
IMAGE_NIX="kyaraben-nix-e2e"

build_quick() {
    echo "Building quick test container..."
    podman build -t "$IMAGE_QUICK" -f Containerfile .
}

build_nix() {
    echo "Building Nix E2E test container (this may take a while)..."
    podman build -t "$IMAGE_NIX" -f Containerfile.nix-e2e .
}

run_quick() {
    echo "Running quick tests in container..."
    podman run --rm "$IMAGE_QUICK" sh -c '
        set -e
        echo "=== Quick CLI Tests ==="

        echo "1. Check status..."
        kyaraben status

        echo ""
        echo "2. Check doctor..."
        kyaraben doctor

        echo ""
        echo "3. Test dry-run apply..."
        kyaraben apply --dry-run

        echo ""
        echo "4. Re-init with force..."
        kyaraben init -u ~/Emulation -f

        echo ""
        echo "5. Check doctor with PSX (will show missing BIOS)..."
        kyaraben doctor || true

        echo ""
        echo "6. Test uninstall..."
        kyaraben uninstall -f

        echo ""
        echo "=== Quick tests completed ==="
    '
}

run_nix() {
    echo "Running Nix E2E tests in container..."
    echo "This will actually build emulators via Nix - may take 5-15 minutes on first run."
    echo ""
    podman run --rm "$IMAGE_NIX"
}

shell_quick() {
    echo "Starting shell in quick test container..."
    podman run -it --rm "$IMAGE_QUICK" /bin/bash
}

shell_nix() {
    echo "Starting shell in Nix E2E container..."
    podman run -it --rm "$IMAGE_NIX" /bin/bash
}

case "${1:-quick}" in
    quick|test)
        build_quick
        run_quick
        ;;
    nix|full)
        build_nix
        run_nix
        ;;
    shell)
        build_quick
        shell_quick
        ;;
    nix-shell)
        build_nix
        shell_nix
        ;;
    build-quick)
        build_quick
        ;;
    build-nix)
        build_nix
        ;;
    *)
        echo "Usage: $0 [quick|nix|shell|nix-shell|build-quick|build-nix]"
        echo ""
        echo "Commands:"
        echo "  quick      - Run quick CLI tests (no Nix, fast)"
        echo "  nix        - Run full Nix E2E tests (builds emulators, slow)"
        echo "  shell      - Drop into quick test container"
        echo "  nix-shell  - Drop into Nix E2E container"
        exit 1
        ;;
esac
