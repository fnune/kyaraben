#!/usr/bin/env bash
set -euo pipefail

. ~/.nix-profile/etc/profile.d/nix.sh

echo "=== Kyaraben Tauri E2E Tests ==="
echo ""
echo "Starting Xvfb..."
Xvfb :99 -screen 0 1280x720x24 &
export DISPLAY=:99
sleep 2

echo "Running WebdriverIO tests..."
cd /home/testuser/kyaraben/ui
npm run test:e2e

echo ""
echo "=== Tauri E2E tests complete! ==="
