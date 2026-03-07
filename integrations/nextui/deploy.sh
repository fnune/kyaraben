#!/usr/bin/env bash
set -euo pipefail

HOST="${1:-root@192.168.178.81}"
PAK_PATH="/mnt/SDCARD/Tools/tg5040/Kyaraben.pak"
SYNCTHING_PATH="/mnt/SDCARD/.userdata/tg5040/kyaraben/syncthing"
KYARABEN_PATH="/mnt/SDCARD/.userdata/tg5040/kyaraben"

echo "Building..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/kyaraben-nextui ./cmd/kyaraben-nextui

echo "Killing processes and clearing config..."
ssh -o StrictHostKeyChecking=no "$HOST" "pkill -9 kyaraben-nextui 2>/dev/null || true; pkill -9 syncthing 2>/dev/null || true; rm -rf $SYNCTHING_PATH $KYARABEN_PATH/kyaraben.json $KYARABEN_PATH/syncthing.pid $KYARABEN_PATH/config.toml; sleep 1"

echo "Deploying..."
scp -o StrictHostKeyChecking=no dist/kyaraben-nextui "$HOST:$PAK_PATH/kyaraben-nextui"

echo "Done"
