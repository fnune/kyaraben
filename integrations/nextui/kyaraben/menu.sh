#!/bin/sh
# Kyaraben main menu UI
# PAK_DIR is set by launch.sh before sourcing this

. "${PAK_DIR}/kyaraben/pair.sh"

show_message() {
    message="$1"
    timeout="${2:-2}"
    killall minui-presenter >/dev/null 2>&1 || true
    if [ "$timeout" = "forever" ]; then
        minui-presenter --message "$message" --timeout -1 &
    else
        minui-presenter --message "$message" --timeout "$timeout"
    fi
}

prompt_pairing_code() {
    code_file="/tmp/kyaraben-code.txt"
    minui-keyboard --title "Enter pairing code" --write-location "$code_file"
    exit_code=$?

    if [ $exit_code -ne 0 ]; then
        rm -f "$code_file"
        return 1
    fi

    code=$(cat "$code_file" 2>/dev/null)
    rm -f "$code_file"

    if [ -z "$code" ]; then
        return 1
    fi

    code=$(echo "$code" | tr '[:lower:]' '[:upper:]' | tr -d ' ')

    if [ ${#code} -ne 6 ]; then
        show_message "Code must be 6 characters" 2
        return 1
    fi

    echo "$code"
}

do_pairing() {
    if ! code=$(prompt_pairing_code); then
        return 1
    fi

    show_message "Connecting..." forever

    if ! remote_device_id=$(pair_with_code "$code"); then
        killall minui-presenter >/dev/null 2>&1 || true
        show_message "Pairing failed" 2
        return 1
    fi

    killall minui-presenter >/dev/null 2>&1 || true

    # Add device to Syncthing and share folders
    . "${PAK_DIR}/kyaraben/setup.sh"
    api_key=$(get_api_key)

    if [ -n "$api_key" ] && [ -n "$remote_device_id" ]; then
        # Add the remote device (don't set name - let Syncthing use its advertised name)
        curl -sf -X PUT \
            -H "X-API-Key: $api_key" \
            -H "Content-Type: application/json" \
            -d "{\"deviceID\": \"$remote_device_id\"}" \
            "$SYNCTHING_API/rest/config/devices/$remote_device_id" >/dev/null 2>&1

        # Share all folders with the device (batch)
        share_folders_with_device "$remote_device_id"
    fi

    show_message "Paired successfully!" 2
    return 0
}

main_menu() {
    while true; do
        menu_file="/tmp/kyaraben-menu.json"
        selection_file="/tmp/kyaraben-selection.txt"
        echo '["Pair with device", "Sync status", "Settings"]' > "$menu_file"

        minui-list \
            --disable-auto-sleep \
            --file "$menu_file" \
            --format json \
            --write-location "$selection_file" \
            --title "Kyaraben"

        exit_code=$?
        rm -f "$menu_file"

        if [ $exit_code -eq 2 ] || [ $exit_code -eq 3 ]; then
            rm -f "$selection_file"
            break
        fi

        selection=$(cat "$selection_file" 2>/dev/null)
        rm -f "$selection_file"

        case "$selection" in
            "Pair with device")
                do_pairing
                ;;
            "Sync status")
                show_message "Not implemented yet" 2
                ;;
            "Settings")
                show_message "Not implemented yet" 2
                ;;
        esac
    done
}
