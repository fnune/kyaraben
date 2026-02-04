#!/bin/bash
# Downloads emulator icons from versions.toml and places them in src/assets/emulators/
# Run from the ui/ directory: ./scripts/download-emulator-icons.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
UI_DIR="$(dirname "$SCRIPT_DIR")"
VERSIONS_TOML="$UI_DIR/../internal/versions/versions.toml"
ASSETS_DIR="$UI_DIR/src/assets/emulators"

mkdir -p "$ASSETS_DIR"

# Parse versions.toml and extract emulator icon URLs
# Format in TOML: icon_url = "https://..."
# The section name [emulator] precedes it

current_emulator=""
while IFS= read -r line; do
  # Match section headers like [eden] or [duckstation]
  if [[ "$line" =~ ^\[([a-z0-9_-]+)\]$ ]]; then
    section="${BASH_REMATCH[1]}"
    # Skip special sections
    if [[ "$section" != "nixpkgs" && "$section" != "retroarch-cores" && "$section" != "nix-portable" && ! "$section" =~ \. ]]; then
      current_emulator="$section"
    else
      current_emulator=""
    fi
  fi

  # Match icon_url lines
  if [[ -n "$current_emulator" && "$line" =~ ^icon_url[[:space:]]*=[[:space:]]*\"(.+)\"$ ]]; then
    url="${BASH_REMATCH[1]}"

    # Determine file extension from URL
    ext="${url##*.}"
    # Handle URLs with query params
    ext="${ext%%\?*}"

    output_file="$ASSETS_DIR/$current_emulator.$ext"

    echo "Downloading $current_emulator icon..."
    echo "  URL: $url"
    echo "  Output: $output_file"

    if curl -fsSL "$url" -o "$output_file"; then
      echo "  Done"
    else
      echo "  Failed to download $current_emulator icon"
    fi
    echo
  fi
done < "$VERSIONS_TOML"

echo "All icons downloaded to $ASSETS_DIR"
echo
echo "Files:"
ls -la "$ASSETS_DIR"
