#!/usr/bin/env bash
set -euo pipefail

# Note: Nix is bundled as nix-portable within kyaraben, no system Nix needed

echo "=== Kyaraben Electron E2E Tests ==="
echo ""

# Debug info
echo "Node version: $(node --version)"
echo "NPM version: $(npm --version)"
echo "Working directory: $(pwd)"

# Set up external binaries for E2E tests
echo "Setting up external binaries for E2E tests..."
BINARIES_DIR="/home/testuser/kyaraben/ui/src-tauri/binaries"
mkdir -p "$BINARIES_DIR"
echo "Binaries in $BINARIES_DIR:"
ls -la "$BINARIES_DIR" || echo "No binaries found"

# Check Electron installation
echo ""
echo "Checking Electron installation..."
if [ -f "/home/testuser/kyaraben/ui/node_modules/electron/dist/electron" ]; then
    echo "Electron binary found"
    ls -la /home/testuser/kyaraben/ui/node_modules/electron/dist/electron
else
    echo "ERROR: Electron binary not found!"
    echo "Contents of node_modules/electron/:"
    ls -la /home/testuser/kyaraben/ui/node_modules/electron/ || echo "electron dir missing"
fi

# Check built files
echo ""
echo "Checking built files..."
ls -la /home/testuser/kyaraben/ui/dist-electron/ || echo "dist-electron missing"
ls -la /home/testuser/kyaraben/ui/dist/ || echo "dist missing"

echo ""
echo "Starting D-Bus..."
mkdir -p /run/dbus
dbus-daemon --system --fork 2>/dev/null || echo "D-Bus system daemon already running or not available"
export DBUS_SESSION_BUS_ADDRESS="unix:path=/tmp/dbus-session-$$"
dbus-daemon --session --address="$DBUS_SESSION_BUS_ADDRESS" --fork 2>/dev/null || echo "D-Bus session daemon failed"
export NO_AT_BRIDGE=1

echo ""
echo "Starting Xvfb..."
Xvfb :99 -screen 0 1280x720x24 &
XVFB_PID=$!
export DISPLAY=:99
sleep 2

# Verify display is working
echo "DISPLAY=$DISPLAY"
echo "DBUS_SESSION_BUS_ADDRESS=$DBUS_SESSION_BUS_ADDRESS"
xdpyinfo -display :99 >/dev/null 2>&1 && echo "Xvfb is running" || echo "WARNING: xdpyinfo failed"

echo ""
echo "Running Playwright tests..."
cd /home/testuser/kyaraben/ui
npm run test:e2e || TEST_EXIT=$?

# Cleanup
kill $XVFB_PID 2>/dev/null || true

echo ""
echo "=== Electron E2E tests complete! ==="
exit ${TEST_EXIT:-0}
