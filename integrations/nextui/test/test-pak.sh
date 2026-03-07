#!/bin/sh
set -e

PAK_DIR="/pak"
SDCARD="/mnt/SDCARD"

export SDCARD_PATH="$SDCARD"
export USERDATA_PATH="/tmp/userdata"
export SYNCTHING_API="http://127.0.0.1:8384"

# Create fake config dir with API key
mkdir -p "$USERDATA_PATH/Syncthing/config"
cat > "$USERDATA_PATH/Syncthing/config/config.xml" << 'EOF'
<configuration>
    <gui><apikey>test-api-key-12345</apikey></gui>
</configuration>
EOF

echo "=== Kyaraben.pak test suite ==="
echo

# --- Unit tests (no API needed) ---

echo "--- Test: systems.sh ---"
. "$PAK_DIR/kyaraben/systems.sh"

system_count=$(list_systems | wc -l)
if [ "$system_count" -eq 14 ]; then
    echo "System count: OK ($system_count)"
else
    echo "System count: FAIL (expected 14, got $system_count)"
    exit 1
fi

tag=$(get_nextui_tag "gb")
folder_id=$(get_folder_id "gb" "roms")
if [ "$tag" = "GB" ] && [ "$folder_id" = "kyaraben-roms-gb" ]; then
    echo "Mapping functions: OK"
else
    echo "Mapping functions: FAIL"
    exit 1
fi
echo

echo "--- Test: build_folders_json ---"
. "$PAK_DIR/kyaraben/setup.sh"

folders_json=$(build_folders_json)

if ! echo "$folders_json" | jq empty 2>/dev/null; then
    echo "JSON validity: FAIL"
    exit 1
fi
echo "JSON validity: OK"

json_folder_count=$(echo "$folders_json" | jq 'length')
if [ "$json_folder_count" -eq 42 ]; then
    echo "Folder count: OK ($json_folder_count)"
else
    echo "Folder count: FAIL (expected 42, got $json_folder_count)"
    exit 1
fi

gb_path=$(echo "$folders_json" | jq -r '.[] | select(.id == "kyaraben-roms-gb") | .path')
if [ "$gb_path" = "/mnt/SDCARD/Roms/Game Boy (GB)" ]; then
    echo "Folder paths: OK"
else
    echo "Folder paths: FAIL"
    exit 1
fi
echo

echo "--- Test: jq device sharing logic ---"
test_folders='[{"id": "kyaraben-roms-gb", "devices": []}, {"id": "other", "devices": []}]'
updated=$(echo "$test_folders" | jq --arg dev "NEW-DEV" '
    map(if .id | startswith("kyaraben-") then .devices += [{"deviceID": $dev}] else . end)
')
kb_devs=$(echo "$updated" | jq '[.[] | select(.id | startswith("kyaraben-")) | .devices | length] | add')
other_devs=$(echo "$updated" | jq '[.[] | select(.id == "other") | .devices | length] | add')
if [ "$kb_devs" -eq 1 ] && [ "$other_devs" -eq 0 ]; then
    echo "Device sharing logic: OK"
else
    echo "Device sharing logic: FAIL"
    exit 1
fi
echo

# --- Integration tests (with fake API) ---

echo "--- Test: setup_all_folders (with fake API) ---"

# Start fake API in background
python3 /test/fake-syncthing-api.py &
api_pid=$!
sleep 1

# Run the actual setup
if setup_all_folders 2>/dev/null; then
    echo "setup_all_folders: OK"
else
    echo "setup_all_folders: FAIL"
    kill $api_pid 2>/dev/null
    exit 1
fi

# Test sharing with a device
share_folders_with_device "TEST-DEVICE-123" 2>/dev/null
echo "share_folders_with_device: OK"

# Stop API
kill $api_pid 2>/dev/null
echo

echo "=== All tests passed ==="
