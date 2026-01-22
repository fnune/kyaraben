#!/usr/bin/env bash
# Download nix-portable binary for the current platform
# This script is called during build to bundle nix-portable with kyaraben

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# nix-portable releases: https://github.com/DavHau/nix-portable/releases
# Pin version for reproducibility. Update manually when needed.
NIX_PORTABLE_VERSION="v012"
NIX_PORTABLE_URL="https://github.com/DavHau/nix-portable/releases/download/${NIX_PORTABLE_VERSION}/nix-portable-$(uname -m)"

get_target_triple() {
    local arch=$(uname -m)
    local os=$(uname -s)

    case "$os" in
        Linux)
            case "$arch" in
                x86_64)  echo "x86_64-unknown-linux-gnu" ;;
                aarch64) echo "aarch64-unknown-linux-gnu" ;;
                *)       echo "unknown-unknown-linux-gnu" ;;
            esac
            ;;
        Darwin)
            # nix-portable doesn't support macOS directly
            # On macOS, users need system Nix
            echo ""
            ;;
        *)
            echo ""
            ;;
    esac
}

TARGET_TRIPLE=$(get_target_triple)

if [ -z "$TARGET_TRIPLE" ]; then
    echo "nix-portable is only supported on Linux. Skipping download."
    exit 0
fi

# Output location - same as sidecar binaries
OUTPUT_DIR="${1:-$PROJECT_ROOT/ui/binaries}"
OUTPUT_NAME="nix-portable-$TARGET_TRIPLE"

mkdir -p "$OUTPUT_DIR"

echo "Downloading nix-portable ${NIX_PORTABLE_VERSION} for $(uname -m)..."
curl -fsSL "$NIX_PORTABLE_URL" -o "$OUTPUT_DIR/$OUTPUT_NAME"
chmod +x "$OUTPUT_DIR/$OUTPUT_NAME"

echo "Downloaded: $OUTPUT_DIR/$OUTPUT_NAME"
