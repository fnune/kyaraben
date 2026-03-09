#!/usr/bin/env bash
# Build standalone CLI binary for release (mirrors CI release-cli job)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DIST_DIR="$PROJECT_ROOT/dist"

VERSION=$(node -p "require('$PROJECT_ROOT/ui/package.json').version" 2>/dev/null || echo "0.0.0")

mkdir -p "$DIST_DIR"
cd "$PROJECT_ROOT"

echo "Building kyaraben CLI v$VERSION..."

CC=musl-gcc CGO_ENABLED=1 go build \
    -ldflags="-s -w -X github.com/fnune/kyaraben/internal/version.Version=$VERSION -linkmode external -extldflags '-static'" \
    -o "$DIST_DIR/Kyaraben-CLI-linux-amd64" \
    ./cmd/kyaraben

(cd "$DIST_DIR" && sha256sum Kyaraben-CLI-linux-amd64 > checksums.txt)

echo "Built: $DIST_DIR/Kyaraben-CLI-linux-amd64"
echo "Checksums: $DIST_DIR/checksums.txt"
