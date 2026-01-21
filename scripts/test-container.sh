#!/usr/bin/env bash
# Test kyaraben in an isolated container using Podman
#
# Usage:
#   ./scripts/test-container.sh        # Build and run basic tests
#   ./scripts/test-container.sh shell  # Build and drop into a shell
#   ./scripts/test-container.sh build  # Just build the image

set -euo pipefail

IMAGE_NAME="kyaraben-test"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

build_image() {
    echo "Building container image..."
    podman build -t "$IMAGE_NAME" -f Containerfile .
}

run_tests() {
    echo "Running tests in container..."
    podman run --rm "$IMAGE_NAME" sh -c '
        set -e
        echo "=== Testing kyaraben CLI ==="

        echo "1. Check status..."
        kyaraben status

        echo ""
        echo "2. Check doctor..."
        kyaraben doctor

        echo ""
        echo "3. Test dry-run apply..."
        kyaraben apply --dry-run

        echo ""
        echo "4. Re-init with different systems..."
        kyaraben init -u ~/Emulation -s snes -s psx -f

        echo ""
        echo "5. Check doctor with PSX (will show missing BIOS)..."
        kyaraben doctor || true

        echo ""
        echo "=== All CLI tests completed ==="
    '
}

run_shell() {
    echo "Starting interactive shell in container..."
    podman run -it --rm "$IMAGE_NAME" /bin/bash
}

case "${1:-test}" in
    build)
        build_image
        ;;
    shell)
        build_image
        run_shell
        ;;
    test|*)
        build_image
        run_tests
        ;;
esac
