# Emulator testing

This directory tracks manual testing of each emulator to verify:

1. Kyaraben's configuration works correctly
2. Extensions are complete
3. ES-DE integration works (including multi-disc)
4. We understand what files each emulator creates

## Devices

Test devices are documented in `devices/`. Each emulator should be tested on each device:

- `feanor` - desktop (CachyOS, Ryzen 9, RX 6650 XT)
- `steamdeck` - Steam Deck (SteamOS)

## Process

For each emulator:

1. Copy `emulators/_template.md` to `emulators/{emulator}/notes.md`
2. Enable the system in kyaraben and apply
3. Work through the checklist
4. Document findings, especially in the recon section
5. Note any issues or improvements needed

## Emulators to test

Standalone:

- [ ] duckstation (PSX)
- [ ] pcsx2 (PS2)
- [ ] rpcs3 (PS3) - opaque
- [ ] ppsspp (PSP) - opaque
- [ ] vita3k (Vita) - opaque
- [ ] mgba (GB, GBC, GBA)
- [ ] melonds (NDS)
- [ ] azahar (N3DS)
- [ ] dolphin (GameCube, Wii)
- [ ] cemu (Wii U)
- [ ] eden (Switch)
- [ ] flycast (Dreamcast)

RetroArch cores:

- [ ] retroarch-mesen (NES)
- [ ] retroarch-bsnes (SNES)
- [ ] retroarch-genesis-plus-gx (Genesis)
- [ ] retroarch-mupen64plus-next (N64)
- [ ] retroarch-beetle-saturn (Saturn)

## Multi-disc systems

These systems commonly have multi-disc games and need `.m3u` testing:

- PSX (already has .m3u extension)
- PS2 (needs .m3u extension added)
- Saturn (rare but possible)
- Dreamcast (rare but possible)

Multi-disc approach for ES-DE:

1. Put individual discs in `roms/{system}/.hidden/`
2. Create `.m3u` file in `roms/{system}/` pointing to discs
3. ES-DE shows only the `.m3u` entry

## Sync implications

Recon findings feed into the sync strategy. For each emulator, we need to know:

- Where save data actually lives
- Whether the emulator uses opaque directory structure
- What can be safely synced vs. what is machine-specific (caches, shaders)

## Cross-cutting issues found

### ~~ES-DE integration broken~~ RESOLVED

Initial diagnosis was wrong. ES-DE custom_systems DOES work, but the override is based on matching `<path>` tags and ROM extensions. If our custom definition doesn't find ROMs (wrong path or missing extension), ES-DE's bundled definition finds them instead.

Root causes found:
- N64: user ROMs were .zip but we didn't include .zip in extensions
- N3DS: user ROMs were in `n3ds/` but our SystemID was "3ds" causing path `%ROMPATH%/3ds`

Fix: ensure kyaraben's system definitions have correct paths (matching SystemID to ES-DE folder conventions) and complete extensions (including .zip where appropriate).

Extensions audit complete - vendored from ES-DE bundled config (commit 1ae0a21).

### ~~BIOS detection too strict~~ RESOLVED

Implemented ProvisionGroup model with:
- Case-insensitive filename matching
- Multiple hash alternatives per BIOS file
- "At least N of these" semantics via MinRequired
- Expanded BIOS data for PSX, PS2, Saturn, Dreamcast, NDS, GameCube, PS3, PS Vita, Wii U, 3DS

See `plans/bios-detection-improvements.md` for design details.

### Cheats directory needed (found during melonDS testing)

Some emulators support cheats with a configurable path (e.g. melonDS `CheatFilePath`). Need to decide:

- Per-emulator cheats dir: `~/Emulation/cheats/{emulator}/`
- Per-system cheats dir: `~/Emulation/cheats/{system}/`

TBD which makes more sense.
