# NextUI integration notes

## Supported systems

14 systems can sync between kyaraben and NextUI:
- gb, gbc, gba, nes, snes, genesis, psx
- mastersystem, gamegear, pcengine
- atari2600, c64, arcade, ngp

## Systems that cannot sync

Kyaraben supports but NextUI doesn't:
- n64, nds, n3ds, gamecube, wii, wiiu, switch
- ps2, ps3, psp, psvita
- saturn, dreamcast
- xbox, xbox360

NextUI supports but kyaraben doesn't:
- Amiga (PUAE), Amstrad CPC, Atari 5200/7800/Lynx
- Colecovision, MSX, Pico-8, Pokemon mini
- Sega 32X/CD/SG-1000, Virtual Boy
- Various Commodore variants (C128, PET, Plus4, VIC20)

## How does this relate to kyaraben devices' enabled systems?

The pak's `systems.sh` defines a mapping between kyaraben system IDs and NextUI tags. This is currently a static list representing the intersection of what both support.

Open questions:
- Should the pak query the paired kyaraben device for its enabled systems?
- Or should users select systems locally on the NextUI device, and kyaraben devices just sync whatever folders exist?
- Current approach: user selects systems on NextUI device, pak creates folders, Syncthing syncs whatever exists

The practical limitation: if a kyaraben device doesn't have a system enabled, there's nothing in that folder to sync. The NextUI pak could still create the folder structure, but it would remain empty until the kyaraben side enables it.

## How does the user choose whether to sync states (opt in)?

Not yet implemented. States are risky because:
- Save states are emulator-specific (a RetroArch state won't work in standalone PCSX2)
- States can corrupt if the emulator version differs
- States are often large

Options to implement:
1. Global toggle in settings: "Sync save states (experimental)"
2. Per-system toggle in the system selection menu
3. Separate "states" content type with its own opt-in flow

Current `folders.sh` handles: roms, saves, bios. Adding states would need:
- A new content type in `setup_symlink`
- NextUI states path discovery (where does NextUI store states?)
- UI for opt-in

## Folder setup (no symlinks)

Syncthing folder IDs are arbitrary - `kyaraben-roms-gb` can point to `/Roms/Game Boy (GB)/` on NextUI and `~/kyaraben/collection/roms/gb/` on desktop.

We configure Syncthing to sync directly to NextUI paths:
- Folder ID `kyaraben-roms-gb` -> `/Roms/Game Boy (GB)/`
- Folder ID `kyaraben-saves-gb` -> `/Saves/GB/`
- Folder ID `kyaraben-bios-gb` -> `/Bios/GB/`

Syncthing handles existing content via its merge/conflict resolution. No symlinks needed.

## Folder discovery and matching

When two devices pair:

1. Each device has folders with kyaraben folder IDs (e.g., `kyaraben-roms-gb`)
2. Syncthing shows pending folder shares for matching IDs
3. Folders with the same ID auto-connect

The pak sets up all 14 supported systems x 3 content types = 42 folders on first run.

When a kyaraben device offers a new folder:
- Syncthing notifies of pending folder share (e.g., `kyaraben-roms-atari2600`)
- Pak extracts system ID from folder ID
- If system ID is in systems.sh mapping, pak auto-accepts and configures the local path
- If system ID is unknown, folder offer is ignored (or could prompt user)

Dynamic acceptance works for any system in the mapping. Adding new systems to kyaraben just requires updating systems.sh in the pak.

## Security notes

The Syncthing web UI on the device has no password by default and is accessible via SSH tunnel. This is a Syncthing default, not something we configure. Users who want to secure it can set a GUI password in Syncthing settings.

The pak disables Syncthing's usage reporting by setting `urAccepted: -1` on first run.
