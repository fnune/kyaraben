#!/bin/sh
# Mapping between Kyaraben system IDs and NextUI paths
# Only includes systems supported by both kyaraben and NextUI

# Format: kyaraben_id:nextui_tag:display_name
# - kyaraben_id: used in Syncthing folder IDs (e.g., kyaraben-roms-gb)
# - nextui_tag: used in NextUI paths (e.g., /Saves/GB/)
# - display_name: used in ROM folder names (e.g., /Roms/Game Boy (GB)/)
SYSTEMS="
gb:GB:Game Boy
gbc:GBC:Game Boy Color
gba:GBA:Game Boy Advance
nes:FC:Nintendo Entertainment System
snes:SFC:Super Nintendo Entertainment System
genesis:MD:Sega Genesis
psx:PS:Sony PlayStation
mastersystem:SMS:Sega Master System
gamegear:GG:Sega Game Gear
pcengine:PCE:TurboGrafx-16
atari2600:A2600:Atari 2600
c64:C64:Commodore 64
arcade:FBN:Arcade
ngp:NGP:Neo Geo Pocket
"

get_nextui_tag() {
    kyaraben_id="$1"
    echo "$SYSTEMS" | grep "^${kyaraben_id}:" | cut -d: -f2
}

get_display_name() {
    kyaraben_id="$1"
    echo "$SYSTEMS" | grep "^${kyaraben_id}:" | cut -d: -f3
}

list_systems() {
    echo "$SYSTEMS" | grep -v '^$'
}

# Get NextUI path for a system and content type
get_nextui_path() {
    kyaraben_id="$1"
    content_type="$2"

    nextui_tag=$(get_nextui_tag "$kyaraben_id")
    display_name=$(get_display_name "$kyaraben_id")

    case "$content_type" in
        roms)
            echo "$SDCARD_PATH/Roms/$display_name ($nextui_tag)"
            ;;
        saves)
            echo "$SDCARD_PATH/Saves/$nextui_tag"
            ;;
        bios)
            echo "$SDCARD_PATH/Bios/$nextui_tag"
            ;;
    esac
}

# Get Syncthing folder ID (must match kyaraben desktop)
get_folder_id() {
    kyaraben_id="$1"
    content_type="$2"
    echo "kyaraben-${content_type}-${kyaraben_id}"
}
