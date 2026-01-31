#!/bin/bash

BIN_DIR="${HOME}/.local/state/kyaraben/bin"
TIMEOUT_SECS=5

echo "=== Emulator wrapper test script ==="
echo "Bin directory: $BIN_DIR"
echo ""

if [[ ! -d "$BIN_DIR" ]]; then
    echo "ERROR: Bin directory does not exist. Run 'kyaraben apply' first."
    exit 1
fi

echo "=== Available wrappers ==="
ls -la "$BIN_DIR"
echo ""

echo "=== Wrapper contents ==="
for wrapper in "$BIN_DIR"/*; do
    [[ -f "$wrapper" ]] || continue
    name=$(basename "$wrapper")
    echo "--- $name ---"
    cat "$wrapper"
    echo ""
done

echo "=== Current symlink target ==="
ls -la "${HOME}/.local/state/kyaraben/current"
echo ""

echo "=== Profile bin directory ==="
ls -la "${HOME}/.local/state/kyaraben/current/bin/" 2>/dev/null || echo "No bin directory in profile"
echo ""

test_emulator() {
    local name=$1
    local wrapper="$BIN_DIR/$name"

    echo "=== Testing: $name ==="

    if [[ ! -x "$wrapper" ]]; then
        echo "ERROR: $wrapper is not executable"
        return 1
    fi

    echo "Running with bash -x (timeout ${TIMEOUT_SECS}s)..."
    timeout "$TIMEOUT_SECS" bash -x "$wrapper" --help 2>&1 || true
    echo ""
    echo "Exit code: $?"
    echo ""
}

if [[ $# -gt 0 ]]; then
    for emu in "$@"; do
        test_emulator "$emu"
    done
else
    echo "=== Testing all emulators ==="
    echo "Usage: $0 [emulator_name...] to test specific ones"
    echo "Testing with --help flag and ${TIMEOUT_SECS}s timeout..."
    echo ""

    for wrapper in "$BIN_DIR"/*; do
        [[ -f "$wrapper" ]] || continue
        name=$(basename "$wrapper")
        test_emulator "$name"
    done
fi
