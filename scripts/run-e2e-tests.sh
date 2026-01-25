#!/bin/bash
# E2E test runner script for Docker container
# This extracts the AppImage and runs Playwright tests

set -ex

echo '=== Setting up directories ==='
mkdir -p /home/kyaraben/.local/state/kyaraben /home/kyaraben/.local/share/kyaraben
rm -rf /home/kyaraben/.local/share/kyaraben/nix-portable/.nix-portable

echo '=== Checking release directory ==='
ls -la /home/kyaraben/app/ui/release/

APPIMAGE=$(ls /home/kyaraben/app/ui/release/*.AppImage | head -1)
echo "=== Found AppImage: $APPIMAGE ==="

cd /home/kyaraben/app/ui/release

echo '=== Running extraction ==='
"$APPIMAGE" --appimage-extract
echo '=== Extraction exit code:' $? '==='

echo '=== Checking if squashfs-root exists ==='
if [ -d squashfs-root ]; then
    echo 'squashfs-root directory EXISTS'
    ls -la squashfs-root/ | head -20
else
    echo 'ERROR: squashfs-root directory DOES NOT EXIST'
    echo 'Current directory contents:'
    ls -la
    exit 1
fi

echo '=== Looking for kyaraben executable ==='
ls -la squashfs-root/kyaraben* 2>&1 || true

echo '=== Starting Xvfb ==='
Xvfb :99 -ac -screen 0 1280x1024x24 >/dev/null 2>&1 &
sleep 2

: > /home/kyaraben/.local/state/kyaraben/kyaraben.log
tail -f /home/kyaraben/.local/state/kyaraben/kyaraben.log &

echo '=== Running tests ==='
env KYARABEN_APPIMAGE=/home/kyaraben/app/ui/release/squashfs-root/kyaraben-ui \
    APPDIR=/home/kyaraben/app/ui/release/squashfs-root \
    DISPLAY=:99 \
    npm run test:e2e --prefix /home/kyaraben/app/ui
