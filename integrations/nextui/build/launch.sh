#!/bin/sh
# Kyaraben PAK launcher for NextUI

PAK_DIR="$(dirname "$0")"
export PAK_PATH="$PAK_DIR"

# Start Syncthing in background if not already running
if ! pgrep -f "syncthing.*kyaraben" > /dev/null 2>&1; then
    "$PAK_DIR/syncthing" \
        --home="$USERDATA_PATH/$PLATFORM/kyaraben/syncthing" \
        --no-browser \
        --no-default-folder \
        --gui-address="127.0.0.1:8484" \
        > "$LOGS_PATH/kyaraben-syncthing.log" 2>&1 &
fi

# Run the main UI
exec "$PAK_DIR/kyaraben-nextui"
