# Dolphin

System(s): GameCube, Wii
Opaque: yes

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

- [x] Saves write to opaque GC/ or Wii/ directories
- [x] Save states write to opaque StateSaves/
- [x] Screenshots write to `~/Emulation/screenshots/dolphin/` (via symlink)

Notes: Saves at `saves/gamecube/{region}/Card A/*.gci`. Savestates at `states/dolphin/`. Screenshots at `screenshots/dolphin/`. All via symlinks from opaque dir - see testing/plans/symlinks-over-opaque.md.

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

Default config path: `~/Emulation/opaque/dolphin/Config/`

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

Default data path: n/a (uses opaque dir)

### Cache location

Cache path: `~/Emulation/opaque/dolphin/Cache/`

Files found:

```
Shaders/OpenGL-specialized-pipeline-GFZP01-*.cache
Shaders/OpenGL-uber-pipeline-*.cache
GFZP01.uidcache
```

### Opaque directory

Path: `~/Emulation/opaque/dolphin/`

Structure:

```
Cache/
Config/
Dump/
GameSettings/
GBA/
GC/EUR/Card A/, GC/JAP/, GC/USA/
Load/
Maps/
ResourcePacks/
SavedAssembly/
ScreenShots/
Shaders/
StateSaves/
Styles/
Themes/
WFS/
Wii/shared2/sys/SYSCONF, Wii/fst.bin
```

What is machine-specific (should not sync): Cache/, Shaders/

What should sync: GC/ (memory cards), Wii/ (NAND), StateSaves/, Config/ (maybe), GameSettings/

### Other locations

Any other files created: None outside opaque dir (good!)

## Sync implications

Based on recon, what needs to sync for this emulator:

- Save data location: opaque/GC/ (memory cards), opaque/Wii/ (NAND)
- Save state location: opaque/StateSaves/
- Any emulator-specific considerations: Wii NAND can be large. Shader cache should NOT sync.

## Issues found

- Screenshots go to `opaque/ScreenShots/` instead of `~/Emulation/screenshots/dolphin/`. The `DumpPath` setting is for frame/texture/audio dumps, not screenshots. Dolphin may not have a configurable screenshot path - screenshots always go to `<user_dir>/ScreenShots/`. Workaround: symlink approach (see testing/plans/symlinks-over-opaque.md).
- When launched from CLI with a ROM path (including ES-DE), Dolphin does not show the game list - goes straight to the game. This is expected behavior when using `-e <rom>`.

## Summary

| Device | Status | Tested by | Date |
|--------|--------|-----------|------|
| feanor | passed | fausto | 2026-02-09 |
| steamdeck | not started | | |
