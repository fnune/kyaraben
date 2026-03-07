#!/bin/sh
# Kyaraben device pairing via relay
# Uses 6-character codes instead of 56-character Syncthing device IDs

RELAY_URL="${KYARABEN_RELAY_URL:-https://kyaraben-relay-kyaraben-28e14310.koyeb.app}"
SYNCTHING_HOME="${USERDATA_PATH}/Syncthing/config"

get_local_device_id() {
    device_id=$(syncthing device-id --home="$SYNCTHING_HOME" 2>/dev/null)
    if [ -z "$device_id" ]; then
        echo "syncthing device-id --home=$SYNCTHING_HOME failed" >&2
        return 1
    fi
    echo "$device_id"
}

# Join an existing pairing session using a 6-char code
# Returns the initiator's device ID on success
join_session() {
    code="$1"
    if ! response=$(curl -sf "${RELAY_URL}/pair/${code}"); then
        echo "Failed to connect to relay" >&2
        return 1
    fi

    device_id=$(echo "$response" | jq -r '.deviceId // empty')
    if [ -z "$device_id" ]; then
        echo "Invalid or expired code" >&2
        return 1
    fi

    echo "$device_id"
}

# Submit our device ID to complete pairing
submit_response() {
    code="$1"
    local_device_id="$2"

    if ! curl -sf -X POST \
        -H "Content-Type: application/json" \
        -d "{\"deviceId\": \"${local_device_id}\"}" \
        "${RELAY_URL}/pair/${code}/response" >/dev/null; then
        echo "Failed to submit device ID" >&2
        return 1
    fi

    return 0
}

# Full pairing flow: get code from user, exchange device IDs
pair_with_code() {
    code="$1"

    local_device_id=$(get_local_device_id)
    if [ -z "$local_device_id" ]; then
        echo "Could not get local Syncthing device ID" >&2
        return 1
    fi

    if ! remote_device_id=$(join_session "$code"); then
        return 1
    fi

    if ! submit_response "$code" "$local_device_id"; then
        return 1
    fi

    echo "$remote_device_id"
}
