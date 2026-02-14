#!/usr/bin/env bash
# Test kyaraben in isolated containers using Podman
#
# Usage:
#   ./scripts/test-container.sh           # Quick tests
#   ./scripts/test-container.sh quick     # Quick tests
#   ./scripts/test-container.sh cli-e2e   # Full CLI E2E tests
#   ./scripts/test-container.sh shell     # Drop into quick test container
#   ./scripts/test-container.sh cli-shell # Drop into CLI E2E container

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

IMAGE_QUICK="kyaraben-test"
IMAGE_CLI="kyaraben-cli-e2e"

build_quick() {
    echo "Building quick test container..."
    podman build -t "$IMAGE_QUICK" -f Containerfile .
}

build_cli() {
    echo "Building CLI E2E test container..."
    podman build -t "$IMAGE_CLI" -f Containerfile.cli-e2e .
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

run_cli() {
    echo "Running CLI E2E tests in container..."
    echo ""
    podman run --rm "$IMAGE_CLI"
}

shell_quick() {
    echo "Starting shell in quick test container..."
    podman run -it --rm "$IMAGE_QUICK" /bin/bash
}

shell_cli() {
    echo "Starting shell in CLI E2E container..."
    podman run -it --rm "$IMAGE_CLI" /bin/bash
}

case "${1:-quick}" in
    quick|test)
        build_quick
        run_quick
        ;;
    cli-e2e|full)
        build_cli
        run_cli
        ;;
    shell)
        build_quick
        shell_quick
        ;;
    cli-shell)
        build_cli
        shell_cli
        ;;
    build-quick)
        build_quick
        ;;
    build-cli)
        build_cli
        ;;
    *)
        echo "Usage: $0 [quick|cli-e2e|shell|cli-shell|build-quick|build-cli]"
        echo ""
        echo "Commands:"
        echo "  quick      - Run quick CLI tests"
        echo "  cli-e2e    - Run full CLI E2E tests"
        echo "  shell      - Drop into quick test container"
        echo "  cli-shell  - Drop into CLI E2E container"
        exit 1
        ;;
esac
