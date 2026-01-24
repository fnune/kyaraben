#!/usr/bin/env bash
set -euo pipefail

# Note: Nix is bundled as nix-portable within kyaraben, no system Nix needed

echo "=== Kyaraben Electron E2E Tests ==="
echo ""

# Set up external binaries for E2E tests
# Electron looks for binaries in dist-electron/../src-tauri/binaries
echo "Setting up external binaries for E2E tests..."
BINARIES_DIR="/home/testuser/kyaraben/ui/src-tauri/binaries"
mkdir -p "$BINARIES_DIR"
# Binaries are already in place from build-sidecar.sh

echo "Starting Xvfb..."
Xvfb :99 -screen 0 1280x720x24 &
export DISPLAY=:99
sleep 2

echo "Running Playwright tests..."
cd /home/testuser/kyaraben/ui
npm run test:e2e

echo ""
echo "=== Electron E2E tests complete! ==="
