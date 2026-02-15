#!/usr/bin/env bash
set -euo pipefail

echo "=== Kyaraben CLI E2E Tests ==="
echo ""
echo "Kyaraben location: $(which kyaraben)"
echo "Using fake installer: ${KYARABEN_E2E_FAKE_INSTALLER:-0}"
echo ""

echo "1. Initialize with default systems..."
kyaraben init -u ~/Emulation -f
echo "   OK"
echo ""

echo "2. Remove systems with required provisions (psx, ps2)..."
CONFIG_PATH="$HOME/.config/kyaraben/config.toml"
sed -i '/^[[:space:]]*psx = /d' "$CONFIG_PATH"
sed -i '/^[[:space:]]*ps2 = /d' "$CONFIG_PATH"
echo "   OK"
echo ""

echo "3. Check status..."
kyaraben status
echo ""

echo "4. Run doctor (remaining systems have no required provisions)..."
kyaraben doctor
echo ""

echo "5. Apply configuration..."
echo ""
kyaraben apply
echo ""

echo "6. Check status after apply..."
kyaraben status
echo ""

echo "7. Verify directory structure was created..."
for dir in roms/snes saves/snes roms/gb bios/gb saves/gb; do
    if [ -d ~/Emulation/$dir ]; then
        echo "   OK: ~/Emulation/$dir exists"
    else
        echo "   FAIL: ~/Emulation/$dir missing"
        exit 1
    fi
done
echo ""

echo "8. Verify state directory exists..."
if [ -d ~/.local/state/kyaraben ]; then
    echo "   OK: state directory exists"
    ls -la ~/.local/state/kyaraben/ || true
else
    echo "   FAIL: state directory missing"
    exit 1
fi
echo ""

echo "9. Test uninstall..."
kyaraben uninstall -f
echo ""

echo "=== All CLI E2E tests passed! ==="
