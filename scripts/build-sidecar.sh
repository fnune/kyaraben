#!/usr/bin/env bash
# Build the Go CLI as a Tauri sidecar binary
# Tauri requires sidecars to be named with target triple suffix
# Also downloads nix-portable for bundling

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BINARIES_DIR="$PROJECT_ROOT/ui/src-tauri/binaries"

get_target_triple() {
    local arch=$(uname -m)
    local os=$(uname -s)

    case "$os" in
        Linux)
            case "$arch" in
                x86_64)  echo "x86_64-unknown-linux-gnu" ;;
                aarch64) echo "aarch64-unknown-linux-gnu" ;;
                armv7l)  echo "armv7-unknown-linux-gnueabihf" ;;
                *)       echo "unknown-unknown-linux-gnu" ;;
            esac
            ;;
        Darwin)
            case "$arch" in
                x86_64)  echo "x86_64-apple-darwin" ;;
                arm64)   echo "aarch64-apple-darwin" ;;
                *)       echo "unknown-apple-darwin" ;;
            esac
            ;;
        MINGW*|MSYS*|CYGWIN*)
            case "$arch" in
                x86_64)  echo "x86_64-pc-windows-msvc" ;;
                *)       echo "unknown-pc-windows-msvc" ;;
            esac
            ;;
        *)
            echo "unknown-unknown-unknown"
            ;;
    esac
}

TARGET_TRIPLE=$(get_target_triple)
OUTPUT_NAME="kyaraben-$TARGET_TRIPLE"

echo "Building kyaraben sidecar for $TARGET_TRIPLE..."

mkdir -p "$BINARIES_DIR"
cd "$PROJECT_ROOT"
go build -o "$BINARIES_DIR/$OUTPUT_NAME" ./cmd/kyaraben

echo "Built: $BINARIES_DIR/$OUTPUT_NAME"

# Download nix-portable for bundling (Linux only)
"$SCRIPT_DIR/download-nix-portable.sh" "$BINARIES_DIR"
