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

echo "2. Check status..."
kyaraben status
echo ""

echo "3. Run doctor (SNES has no required provisions)..."
kyaraben doctor
echo ""

echo "4. Apply configuration..."
echo ""
kyaraben apply
echo ""

echo "5. Check status after apply..."
kyaraben status
echo ""

echo "6. Verify directory structure was created..."
for dir in roms/snes bios/snes saves/snes screenshots/snes; do
    if [ -d ~/Emulation/$dir ]; then
        echo "   OK: ~/Emulation/$dir exists"
    else
        echo "   FAIL: ~/Emulation/$dir missing"
        exit 1
    fi
done
echo ""

echo "7. Verify state directory exists..."
if [ -d ~/.local/state/kyaraben ]; then
    echo "   OK: state directory exists"
    ls -la ~/.local/state/kyaraben/ || true
else
    echo "   FAIL: state directory missing"
    exit 1
fi
echo ""

echo "8. Test uninstall..."
kyaraben uninstall -f
echo ""

echo "=== All CLI E2E tests passed! ==="
