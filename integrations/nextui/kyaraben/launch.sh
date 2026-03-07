#!/bin/sh
# Kyaraben launcher for NextUI
# Based on upstream syncthing pak, with folder auto-configuration

PAK_DIR="$(cd "$(dirname "$0")" && pwd)"
PAK_NAME="$(basename "$PAK_DIR")"
PAK_NAME="${PAK_NAME%.*}"
set -x

rm -f "$LOGS_PATH/$PAK_NAME.txt"
exec >>"$LOGS_PATH/$PAK_NAME.txt"
exec 2>&1

echo "$0" "$@"
cd "$PAK_DIR" || exit 1
mkdir -p "$USERDATA_PATH/$PAK_NAME"

architecture=arm
if uname -m | grep -q '64'; then
    architecture=arm64
fi

export HOME="$USERDATA_PATH/$PAK_NAME"
export LD_LIBRARY_PATH="$PAK_DIR/lib:$LD_LIBRARY_PATH"
export PATH="$PAK_DIR/bin/$architecture:$PAK_DIR/bin/$PLATFORM:$PAK_DIR/bin:$PATH"
export SSL_CERT_FILE="$PAK_DIR/certs/ca-certificates.crt"

show_message() {
    message="$1"
    seconds="$2"
    killall minui-presenter >/dev/null 2>&1 || true
    if [ -z "$seconds" ] || [ "$seconds" = "forever" ]; then
        minui-presenter --message "$message" --timeout -1 &
    else
        minui-presenter --message "$message" --timeout "$seconds"
    fi
}

cleanup() {
    rm -f /tmp/stay_awake
    rm -f /tmp/kyaraben-*.json
    killall minui-presenter >/dev/null 2>&1 || true
}

service_is_running() {
    pgrep syncthing >/dev/null 2>&1
}

start_service() {
    show_message "Starting Kyaraben..." forever

    # Kill any existing syncthing to avoid duplicates
    killall syncthing >/dev/null 2>&1 || true
    sleep 1

    # Use upstream service-on to start Syncthing
    "$PAK_DIR/bin/service-on"

    # Wait for Syncthing to be running
    counter=0
    while ! service_is_running; do
        counter=$((counter + 1))
        if [ "$counter" -gt 10 ]; then
            killall minui-presenter >/dev/null 2>&1 || true
            show_message "Failed to start" 2
            return 1
        fi
        sleep 1
    done

    # Configure folders
    . "$PAK_DIR/kyaraben/setup.sh"
    setup_all_folders

    killall minui-presenter >/dev/null 2>&1 || true
    return 0
}

stop_service() {
    show_message "Stopping..." 2
    "$PAK_DIR/bin/service-off"
}

main() {
    echo "1" >/tmp/stay_awake
    trap "cleanup" EXIT INT TERM HUP QUIT

    allowed_platforms="miyoomini my282 my355 tg5040 rg35xxplus"
    if ! echo "$allowed_platforms" | grep -q "$PLATFORM"; then
        show_message "$PLATFORM not supported" 2
        return 1
    fi

    chmod +x "$PAK_DIR/bin/$architecture/syncthing"
    chmod +x "$PAK_DIR/bin/$architecture/jq"
    chmod +x "$PAK_DIR/bin/$PLATFORM/minui-list"
    chmod +x "$PAK_DIR/bin/$PLATFORM/minui-presenter"
    chmod +x "$PAK_DIR/bin/$PLATFORM/minui-keyboard"
    chmod +x "$PAK_DIR/bin/service-on"
    chmod +x "$PAK_DIR/bin/service-off"
    chmod +x "$PAK_DIR/kyaraben/"*.sh

    # Start Syncthing if not running
    if ! service_is_running; then
        if ! start_service; then
            return 1
        fi
    else
        # Syncthing already running, ensure folders are configured
        . "$PAK_DIR/kyaraben/setup.sh"
        setup_all_folders
    fi

    # Show main menu
    . "$PAK_DIR/kyaraben/menu.sh"
    main_menu
}

main "$@"
