# Flycast

System(s): Dreamcast
Opaque: no

## Installation

- [x] Enable system, apply succeeds
- [x] Emulator appears in application menu
- [x] Emulator launches from menu
- [ ] No onboarding wizard or first-run prompts - shows "game list is empty" on first launch

Notes: User must manually add ROM directory on first launch.

## BIOS/firmware

Optional files:
- `dc_boot.bin` (Boot ROM) - MD5: `e10c53c2f8b90bab96ead2d368858623` - optional
- `dc_flash.bin` (Flash ROM) - MD5: `0a93f7940c455905bea6e392dfde92a4` - optional

Location: `~/Emulation/bios/dreamcast/`

- [ ] `kyaraben doctor` shows correct status before adding
- [ ] Files in correct location are detected
- [ ] Hash verification works

Notes:

## ROM loading

- [x] ROM in correct folder visible to emulator (after manually adding dir)
- [x] ROM launches and runs (audio, video, input)

Test ROM used: Sturmwind

### Extensions

Current: `.gdi`, `.cdi`, `.chd`, `.cue`

Possible gaps: `.m3u` for multi-disc? (rare for DC but exists)

- [ ] Tested formats:
- [ ] Missing formats:
- [ ] Unnecessary formats:

## Path configuration

- [x] Saves write to `~/Emulation/saves/dreamcast/` - WORKS (Game Save Folder configured)
- [ ] Save states write to `~/Emulation/states/flycast/` - FAILS (Savestate Folders shows "Add" in UI, config key may be wrong)
- [ ] Screenshots write to `~/Emulation/screenshots/dreamcast/` - N/A, flycast doesn't support configurable screenshots dir

Notes: INI format fix working. Content and Game Save paths configured correctly. Savestate path config key needs investigation.

## Persistence

- [ ] Save file persists after closing
- [ ] Save loads correctly on re-launch

Notes:

## ES-DE integration

- [x] System appears in ES-DE
- [x] ROMs visible
- [ ] Scraping works
- [x] Launching from ES-DE works (used standalone flycast, not RetroArch core)

Notes: ES-DE launched standalone flycast despite bundled config defaulting to RetroArch core. Possibly fell back because RetroArch flycast core not installed.

### Multi-disc

Rare for Dreamcast but exists (e.g., Shenmue, D2).

- [ ] `.m3u` extension supported (needs verification)
- [ ] Discs in `.hidden/` folder not shown
- [ ] `.m3u` shows as single entry
- [ ] Disc switching works in-game

Notes:

## Post-testing recon

After running the emulator, document what was created.

### Config location

Default config path: `~/.config/flycast/`

Files found:

```
```

Managed by kyaraben: `emu.cfg` (DataPath, ContentPath, SavePath, SavestatesPath, ScreenshotsPath)

Not managed (should be?):

### Data location

Default data path: `~/.local/share/flycast/`

Files found:

```
```

### Cache location

Cache path: `~/.cache/flycast/`

Files found:

```
```

### Other locations

Any other files created:

## Sync implications

Based on recon, what needs to sync for this emulator:

- Save data location: `~/Emulation/saves/dreamcast/` (VMU saves)
- Save state location: `~/Emulation/states/flycast/`
- Any emulator-specific considerations:

## Issues found

List any bugs or improvements needed:

- [x] Config format wrong: kyaraben writes top-level keys but flycast uses `[config]` section - FIXED, switched to INI format
- [x] Savestate path key wrong: kyaraben sets `SavestatesPath` but Flycast uses `Dreamcast.SavestatePath` (with prefix) - FIXED
- [x] VMU path separate from save path: need to set `Dreamcast.VMUPath` in addition to `Dreamcast.SavePath` - FIXED
- [x] BIOS path: should also set `Dreamcast.BiosPath` (Custom Paths) in addition to `Flycast.DataPath` - FIXED
- [ ] Additional Flycast Custom Paths keys (optional):
  - `Dreamcast.CheatPath` - cheat files
  - `Dreamcast.BoxartPath` - box art
  - `Dreamcast.MappingsPath` - controller mappings
  - `Dreamcast.TexturePath` - texture packs
  - `Dreamcast.TextureDumpPath` - texture dumps
- [ ] Flycast does not support configurable screenshots directory (confirmed)
- [x] First-launch shows "game list is empty" - FIXED, ContentPath now configured
- [ ] Save state hotkeys don't work when launching via CLI (known flycast issue)

### Correct Flycast config keys (from UI testing)

```ini
[config]
Dreamcast.BiosPath = ~/Emulation/bios/dreamcast
Dreamcast.ContentPath = ~/Emulation/roms/dreamcast
Dreamcast.SavePath = ~/Emulation/saves/dreamcast
Dreamcast.VMUPath = ~/Emulation/saves/dreamcast
Dreamcast.SavestatePath = ~/Emulation/states/flycast
Dreamcast.CheatPath = ~/Emulation/cheats/dreamcast  # for future cheats support
Flycast.DataPath = ~/Emulation/bios/dreamcast
```

## Summary

| Device | Status | Tested by | Date |
|--------|--------|-----------|------|
| feanor | retesting after fix | fausto | 2026-02-08 |
| steamdeck | not started | | |
