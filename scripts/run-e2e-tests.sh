#!/bin/bash
# E2E test runner script for Docker container
# Extracts the AppImage and runs Playwright tests

set -e

# Setup
mkdir -p /root/.local/state/kyaraben /root/.local/share/kyaraben

# Extract AppImage (FUSE not available in Docker)
cd /app/ui/release
APPIMAGE=$(ls *.AppImage | head -1)
echo "Extracting $APPIMAGE..."
./"$APPIMAGE" --appimage-extract >/dev/null

# Verify extraction
if [ ! -f squashfs-root/kyaraben-ui ]; then
    echo "ERROR: Extraction failed - kyaraben-ui not found"
    ls -la squashfs-root/ 2>/dev/null || echo "squashfs-root does not exist"
    exit 1
fi

# Start virtual display
Xvfb :99 -ac -screen 0 1280x1024x24 >/dev/null 2>&1 &
sleep 2

# Tail app log in background
tail -f /root/.local/state/kyaraben/kyaraben.log 2>/dev/null &

# Run tests
echo "Running Playwright tests..."
env KYARABEN_APPIMAGE=/app/ui/release/squashfs-root/kyaraben-ui \
    APPDIR=/app/ui/release/squashfs-root \
    DISPLAY=:99 \
    /app/scripts/run-ui-e2e.sh npm run test:e2e --prefix /app/ui
