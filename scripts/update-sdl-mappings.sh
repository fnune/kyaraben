#!/usr/bin/env bash
# Regenerates internal/sdlmappings/gamecontrollerdb.txt from SDL_GameControllerDB.
# Usage: scripts/update-sdl-mappings.sh <commit>
#
# RetroArch's AppImage bundles an SDL2 whose built-in controller table predates
# the hid-playstation kernel driver, so some pads come out scrambled. Kyaraben
# feeds SDL extra mappings, but only for pads that table does not cover and
# never for Valve devices, so the Steam Deck and pads that already work keep
# exactly the mappings they have today.
#
# retroarch-builtin-guids.txt lists the GUIDs in that built-in table. Refresh it
# when the RetroArch version changes:
#   retroarch --appimage-extract && strings squashfs-root/usr/lib/libSDL2-2.0.so.0 \
#     | grep -E '^[0-9a-f]{32},' | cut -d, -f1 | sort -u

set -euo pipefail

commit="${1:?usage: $0 <SDL_GameControllerDB commit>}"
dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/internal/sdlmappings"

curl -fsSL "https://raw.githubusercontent.com/mdqinc/SDL_GameControllerDB/${commit}/gamecontrollerdb.txt" |
    awk -v builtin="$dir/retroarch-builtin-guids.txt" -v commit="$commit" '
        BEGIN {
            while ((getline g < builtin) > 0) skip[g] = 1
            print "# SDL_GameControllerDB " commit " (zlib license, https://github.com/mdqinc/SDL_GameControllerDB)"
            print "# Linux entries, minus Valve devices and GUIDs RetroArch'"'"'s SDL already maps."
        }
        /platform:Linux/ {
            guid = substr($0, 1, 32)
            if (substr(guid, 9, 4) == "de28" || guid in skip) next
            print
        }
    ' > "$dir/gamecontrollerdb.txt"

echo "Wrote $(grep -vc '^#' "$dir/gamecontrollerdb.txt") mappings to $dir/gamecontrollerdb.txt"
