#!/usr/bin/env bash
set -euo pipefail

echo "=== Kyaraben Nix E2E Tests ==="
echo ""
echo "Kyaraben location: $(which kyaraben)"
echo "Note: Nix is bundled as nix-portable within kyaraben"
echo ""

echo "1. Initialize with TIC-80..."
kyaraben init -u ~/Emulation -s tic80 -f
echo "   OK"
echo ""

echo "2. Check status..."
kyaraben status
echo ""

echo "3. Run doctor (should show no provisions needed for TIC-80)..."
kyaraben doctor
echo ""

echo "4. Apply configuration (this builds TIC-80 via Nix - may take a while)..."
echo "   This invokes 'nix build' to fetch/build the TIC-80 emulator."
echo ""
kyaraben apply
echo ""

echo "5. Check status after apply..."
kyaraben status
echo ""

echo "6. Verify directory structure was created..."
# Note: states are per-emulator (states/tic-80), not per-system (states/tic80)
for dir in roms/tic80 bios/tic80 saves/tic80 screenshots/tic80; do
    if [ -d ~/Emulation/$dir ]; then
        echo "   OK: ~/Emulation/$dir exists"
    else
        echo "   FAIL: ~/Emulation/$dir missing"
        exit 1
    fi
done
echo ""

echo "7. Verify Nix store has content..."
if [ -d ~/.local/share/kyaraben/store ] || [ -d ~/.local/share/kyaraben/flake ]; then
    echo "   OK: kyaraben data directory exists"
    ls -la ~/.local/share/kyaraben/ || true
else
    echo "   Note: kyaraben data directory not found (may be using system nix store)"
fi
echo ""

echo "8. Test uninstall..."
kyaraben uninstall -f
echo ""

echo "=== All Nix E2E tests passed! ==="
