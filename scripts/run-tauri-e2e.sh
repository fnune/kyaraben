#!/usr/bin/env bash
set -euo pipefail

# Note: Nix is bundled as nix-portable within kyaraben, no system Nix needed

echo "=== Kyaraben Tauri E2E Tests ==="
echo ""

# Copy external binaries to target/release for E2E testing
# When running the raw binary (not AppImage), external binaries must be alongside the executable
echo "Setting up external binaries for E2E tests..."
BINARIES_DIR="/home/testuser/kyaraben/ui/src-tauri/binaries"
TARGET_DIR="/home/testuser/kyaraben/ui/src-tauri/target/release"
for binary in "$BINARIES_DIR"/*; do
    if [ -f "$binary" ]; then
        filename=$(basename "$binary")
        cp -v "$binary" "$TARGET_DIR/$filename"
    fi
done

echo "Starting Xvfb..."
Xvfb :99 -screen 0 1280x720x24 &
export DISPLAY=:99
sleep 2

echo "Running WebdriverIO tests..."
cd /home/testuser/kyaraben/ui
npm run test:e2e

echo ""
echo "=== Tauri E2E tests complete! ==="
