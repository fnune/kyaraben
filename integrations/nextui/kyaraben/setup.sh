#!/bin/sh
# Kyaraben folder setup - called after Syncthing starts
# Configures all system folders via Syncthing REST API (batch)

# Source systems.sh if not already loaded
# PAK_DIR is set by launch.sh before sourcing this
if ! command -v list_systems >/dev/null 2>&1; then
    . "${PAK_DIR}/kyaraben/systems.sh"
fi

SYNCTHING_CONFIG="${USERDATA_PATH}/Syncthing/config"
SYNCTHING_API="http://127.0.0.1:8384"

get_api_key() {
    grep -o '<apikey>[^<]*</apikey>' "$SYNCTHING_CONFIG/config.xml" | sed 's/<[^>]*>//g'
}

wait_for_api() {
    max_attempts="${1:-30}"
    attempt=0

    # Wait for config file to exist first
    while [ ! -f "$SYNCTHING_CONFIG/config.xml" ] && [ "$attempt" -lt "$max_attempts" ]; do
        attempt=$((attempt + 1))
        sleep 1
    done

    api_key=$(get_api_key)
    if [ -z "$api_key" ]; then
        echo "No API key found in config"
        return 1
    fi

    # Now wait for API to respond
    while [ "$attempt" -lt "$max_attempts" ]; do
        if curl -sf -H "X-API-Key: $api_key" "$SYNCTHING_API/rest/system/status" >/dev/null 2>&1; then
            return 0
        fi
        attempt=$((attempt + 1))
        sleep 1
    done
    return 1
}

# Build JSON array of all kyaraben folders
build_folders_json() {
    echo "Building folder JSON..." >&2

    # Collect all folder objects, then join with commas
    folder_objects=""

    for system_line in $(list_systems | tr ' ' '_'); do
        # Parse system_line (format: kyaraben_id:nextui_tag:display_name)
        kyaraben_id=$(echo "$system_line" | cut -d: -f1)
        nextui_tag=$(echo "$system_line" | cut -d: -f2)
        display_name=$(echo "$system_line" | cut -d: -f3 | tr '_' ' ')

        [ -z "$kyaraben_id" ] && continue

        for content_type in roms saves bios; do
            folder_id=$(get_folder_id "$kyaraben_id" "$content_type")
            folder_path=$(get_nextui_path "$kyaraben_id" "$content_type")
            label="$display_name - $content_type"

            echo "  Creating folder: $folder_id -> $folder_path" >&2
            mkdir -p "$folder_path"

            folder_obj="{\"id\": \"$folder_id\", \"path\": \"$folder_path\", \"label\": \"$label\", \"type\": \"sendreceive\", \"rescanIntervalS\": 60, \"fsWatcherEnabled\": false, \"ignorePerms\": true, \"devices\": []}"

            if [ -z "$folder_objects" ]; then
                folder_objects="$folder_obj"
            else
                folder_objects="$folder_objects, $folder_obj"
            fi
        done
    done

    echo "[$folder_objects]"
}

set_device_name() {
    api_key="$1"
    device_name="nextui-$(echo "$PLATFORM" | tr '[:upper:]' '[:lower:]')"

    # Get local device ID
    local_id=$(curl -sf -H "X-API-Key: $api_key" "$SYNCTHING_API/rest/system/status" | jq -r '.myID')
    if [ -z "$local_id" ]; then
        echo "ERROR: Could not get local device ID"
        return 1
    fi

    # Get current device config and update name
    device_config=$(curl -sf -H "X-API-Key: $api_key" \
        "$SYNCTHING_API/rest/config/devices/$local_id")
    if [ -z "$device_config" ]; then
        return 1
    fi

    updated=$(echo "$device_config" | jq --arg name "$device_name" '.name = $name')
    curl -sf -X PUT \
        -H "X-API-Key: $api_key" \
        -H "Content-Type: application/json" \
        -d "$updated" \
        "$SYNCTHING_API/rest/config/devices/$local_id" >/dev/null 2>&1

    echo "Device name set to: $device_name"
}

setup_all_folders() {
    echo "=== setup_all_folders starting ==="

    echo "Waiting for Syncthing API..."
    if ! wait_for_api 30; then
        echo "ERROR: Syncthing API not available after 30 seconds"
        return 1
    fi
    echo "Syncthing API is ready"

    api_key=$(get_api_key)
    if [ -z "$api_key" ]; then
        echo "ERROR: Could not get API key from config.xml"
        return 1
    fi
    echo "Got API key"

    # Set device name so other devices see us as "nextui-<platform>"
    set_device_name "$api_key"

    # Disable usage reporting
    echo "Disabling usage reporting..."
    options=$(curl -sf -H "X-API-Key: $api_key" "$SYNCTHING_API/rest/config/options")
    if [ -n "$options" ]; then
        updated_options=$(echo "$options" | jq '.urAccepted = -1')
        curl -sf -X PUT \
            -H "X-API-Key: $api_key" \
            -H "Content-Type: application/json" \
            -d "$updated_options" \
            "$SYNCTHING_API/rest/config/options" >/dev/null 2>&1
        echo "Usage reporting disabled"
    fi

    # Bind GUI to localhost only (dismisses auth warning since it's not remote-accessible)
    echo "Configuring GUI to localhost only..."
    gui_config=$(curl -sf -H "X-API-Key: $api_key" "$SYNCTHING_API/rest/config/gui")
    if [ -n "$gui_config" ]; then
        updated_gui=$(echo "$gui_config" | jq '.address = "127.0.0.1:8384"')
        curl -sf -X PUT \
            -H "X-API-Key: $api_key" \
            -H "Content-Type: application/json" \
            -d "$updated_gui" \
            "$SYNCTHING_API/rest/config/gui" >/dev/null 2>&1
        echo "GUI bound to localhost"
    fi

    echo "Fetching current folders..."
    existing=$(curl -sf -H "X-API-Key: $api_key" "$SYNCTHING_API/rest/config/folders")
    if [ -z "$existing" ]; then
        existing="[]"
    fi
    existing_count=$(echo "$existing" | jq 'length')
    echo "Found $existing_count existing folders"

    # Check if kyaraben folders already configured
    if echo "$existing" | grep -q "kyaraben-roms-gb"; then
        echo "Kyaraben folders already configured, skipping"
        return 0
    fi

    echo "Configuring Kyaraben folders..."

    # Build kyaraben folders JSON
    kyaraben_folders=$(build_folders_json)
    kyaraben_count=$(echo "$kyaraben_folders" | jq 'length')
    echo "Built $kyaraben_count kyaraben folders"

    # Merge existing folders with kyaraben folders
    merged=$(echo "$existing" "$kyaraben_folders" | jq -s 'add')
    merged_count=$(echo "$merged" | jq 'length')
    echo "Merged total: $merged_count folders"

    # Batch update all folders
    echo "Sending batch update to Syncthing API..."
    if curl -sf -X PUT \
        -H "X-API-Key: $api_key" \
        -H "Content-Type: application/json" \
        -d "$merged" \
        "$SYNCTHING_API/rest/config/folders" >/dev/null; then
        echo "=== Folder setup complete ($kyaraben_count folders) ==="
    else
        echo "ERROR: Folder setup API call failed"
        return 1
    fi
}

# Share all kyaraben folders with a device
share_folders_with_device() {
    device_id="$1"
    api_key=$(get_api_key)

    if [ -z "$api_key" ] || [ -z "$device_id" ]; then
        echo "Missing API key or device ID"
        return 1
    fi

    echo "Sharing folders with device..."

    # Get current folders
    folders=$(curl -sf -H "X-API-Key: $api_key" "$SYNCTHING_API/rest/config/folders")

    # Add device to all kyaraben folders
    updated=$(echo "$folders" | jq --arg dev "$device_id" '
        map(
            if .id | startswith("kyaraben-") then
                .devices += [{"deviceID": $dev}] | .devices |= unique_by(.deviceID)
            else
                .
            end
        )
    ')

    # Batch update
    curl -sf -X PUT \
        -H "X-API-Key: $api_key" \
        -H "Content-Type: application/json" \
        -d "$updated" \
        "$SYNCTHING_API/rest/config/folders" >/dev/null

    echo "Shared all kyaraben folders with device"
}

# Run if executed directly
if [ "$(basename "$0")" = "setup.sh" ]; then
    setup_all_folders
fi
