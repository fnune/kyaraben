# Dolphin

System(s): GameCube, Wii
Opaque: no (uses symlinks)

## Installation

- [x] Enable system, apply succeeds
- [x] Emulator appears in application menu
- [x] Emulator launches from menu
- [x] No onboarding wizard or first-run prompts

Notes: Clean launch, no wizard.

## BIOS/firmware

Required files: none (GameCube IPL optional for boot animation/fonts)

- [ ] `kyaraben doctor` shows correct status before adding
- [ ] Files in correct location are detected
- [ ] Hash verification works

Notes: BIOS is optional. Not tested yet.

## ROM loading

- [x] ROM launches and runs (audio, video, input)

Test ROM used: F-Zero GX (EUR)

### Game library

Does the emulator have a built-in game list/library?

- [x] Emulator supports game library: yes
- [x] Library can be pre-configured via config: yes (ISOPath0, ISOPath1)
- [x] Games from ROM directory appear in library (6 GameCube ROMs visible)

Notes: kyaraben configures ISOPath0 (gamecube) and ISOPath1 (wii).

### Extensions

Current: .ciso, .dff, .dol, .elf, .gcm, .gcz, .iso, .json, .m3u, .rvz, .tgc, .wad, .wbfs, .wia

- [ ] Tested formats: .iso, .rvz
- [ ] Missing formats:
- [ ] Unnecessary formats:

## Path configuration

- [x] Saves write to `~/Emulation/saves/gamecube/` and `~/Emulation/saves/wii/` (via symlinks)
- [x] Save states write to `~/Emulation/states/dolphin/` (via symlink)
- [x] Screenshots write to `~/Emulation/screenshots/dolphin/` (via symlink)

Notes: kyaraben creates symlinks from `~/.local/share/dolphin-emu/` to the user store. Config at standard XDG location.

## Persistence

- [ ] Save file persists after closing
- [ ] Save loads correctly on re-launch

Notes:

## ES-DE integration

- [x] System appears in ES-DE (gamecube)
- [ ] System appears in ES-DE (wii) - no ROMs to test
- [x] ROMs visible
- [ ] Scraping works - not tested
- [x] Launching from ES-DE works

### Multi-disc (if applicable)

- [ ] `.m3u` extension supported
- [ ] Discs in `.hidden/` folder not shown
- [ ] `.m3u` shows as single entry
- [ ] Disc switching works in-game

Notes: No multi-disc ROMs available for testing.

## Post-testing recon

After running the emulator, document what was created.

### Config location

Default config path: `~/.config/dolphin-emu/`

Files found:

```
Dolphin.ini (kyaraben-managed)
WiimoteNew.ini
GCPadNew.ini
GBA.ini
GCKeyNew.ini
FreeLookController.ini
Qt.ini
TimePlayed.ini
```

Managed by kyaraben: Dolphin.ini (DumpPath, ISOPath0, ISOPath1, ISOPaths)

Not managed (should be?): Controller configs, Qt.ini

### Data location

Default data path: `~/.local/share/dolphin-emu/`

Symlinks created by kyaraben:

```
~/.local/share/dolphin-emu/GC/          → ~/Emulation/saves/gamecube/
~/.local/share/dolphin-emu/Wii/         → ~/Emulation/saves/wii/
~/.local/share/dolphin-emu/StateSaves/  → ~/Emulation/states/dolphin/
~/.local/share/dolphin-emu/ScreenShots/ → ~/Emulation/screenshots/dolphin/
```

### Cache location

Cache path: `~/.local/share/dolphin-emu/Cache/`

Files found:

```
Shaders/OpenGL-specialized-pipeline-GFZP01-*.cache
Shaders/OpenGL-uber-pipeline-*.cache
GFZP01.uidcache
```

### Other locations

Any other files created: None outside standard XDG locations (good!)

## Sync implications

Based on recon, what needs to sync for this emulator:

- Save data location: `saves/gamecube/`, `saves/wii/`
- Save state location: `states/dolphin/`
- Any emulator-specific considerations: Wii NAND can be large. Shader cache (in XDG data dir) should NOT sync.

## Issues found

- When launched from CLI with a ROM path (including ES-DE), Dolphin does not show the game list - goes straight to the game. This is expected behavior when using `-e <rom>`.

Resolved:
- Screenshots now go to correct location via symlink.

## Summary

| Device | Status | Tested by | Date |
|--------|--------|-----------|------|
| feanor | passed | fausto | 2026-02-09 |
| steamdeck | not started | | |
