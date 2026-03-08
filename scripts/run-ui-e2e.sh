#!/usr/bin/env bash
set -euo pipefail

E2E_ROOT="${E2E_ROOT:-/tmp/kyaraben-e2e}"

mkdir -p "$E2E_ROOT/config" "$E2E_ROOT/state" "$E2E_ROOT/data" "$E2E_ROOT/home"

export HOME="$E2E_ROOT/home"
export XDG_CONFIG_HOME="$E2E_ROOT/config"
export XDG_STATE_HOME="$E2E_ROOT/state"
export XDG_DATA_HOME="$E2E_ROOT/data"
export KYARABEN_E2E_FAKE_INSTALLER=1
export ELECTRON_OZONE_PLATFORM_HINT="${ELECTRON_OZONE_PLATFORM_HINT:-headless}"

exec "$@"
