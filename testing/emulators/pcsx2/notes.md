# PCSX2

System(s): PS2
Opaque: no

## Installation

- [x] Enable system, apply succeeds
- [x] Emulator appears in application menu
- [x] Emulator launches from menu
- [x] No onboarding wizard or first-run prompts - WORKS after SettingsVersion fix

Notes: Fixed by adding `[UI] SettingsVersion = 1` and `[UI] SetupWizardIncomplete = false` to managed config.

## BIOS/firmware

Required files:
- `scph39001.bin` (USA) - MD5: `d5ce2c7d119f563ce04bc04dbc3a323e` - required
- `scph70004.bin` (Europe) - MD5: `d333558cc14561c1fdc334c75d5f37b7` - optional
- `scph10000.bin` (Japan) - MD5: `2e6e6db3a66e65e86ad75389cd7fb4b6` - optional

Location: `~/Emulation/bios/ps2/`

- [ ] `kyaraben doctor` shows correct status before adding
- [ ] Files in correct location are detected
- [ ] Hash verification works

Notes:

## ROM loading

- [x] ROM in correct folder visible to emulator (5 games found after wizard)
- [x] ROM launches and runs (audio, video, input)

Test ROM used: Shadow of the Colossus

### Extensions

Current: `.bin`, `.chd`, `.cso`, `.iso`, `.gz`

Known gap: `.m3u` not included (PS2 has multi-disc games)

- [ ] Tested formats:
- [ ] Missing formats:
- [ ] Unnecessary formats:

## Path configuration

- [ ] Saves write to `~/Emulation/saves/ps2/` - not yet verified after fix
- [x] Save states write to `~/Emulation/states/pcsx2/` - WORKS after SettingsVersion fix
- [ ] Screenshots write to `~/Emulation/screenshots/ps2/` - FAILS (goes to `~/.config/PCSX2/snaps/`)

Notes: After adding SettingsVersion = 1, save states now go to correct location. Screenshots use "Snapshots Directory" config key (not Screenshots).

## Persistence

- [ ] Save file persists after closing
- [ ] Save loads correctly on re-launch

Notes:

## ES-DE integration

- [x] System appears in ES-DE
- [x] ROMs visible (all ROMs found)
- [ ] Scraping works
- [x] Launching from ES-DE works

### Multi-disc

PS2 has multi-disc games. Currently `.m3u` is NOT in extension list.

- [ ] `.m3u` extension supported (needs adding to system definition)
- [ ] Discs in `.hidden/` folder not shown
- [ ] `.m3u` shows as single entry
- [ ] Disc switching works in-game

Notes:

## Post-testing recon

After running the emulator, document what was created.

### Config location

Default config path: `~/.config/PCSX2/`

Files found:

```
```

Managed by kyaraben: `inis/PCSX2.ini` (Bios, MemoryCards, Savestates, Screenshots, GameList paths)

Not managed (should be?):

### Data location

Default data path: `~/.local/share/PCSX2/`

Files found:

```
```

### Cache location

Cache path: `~/.cache/PCSX2/`

Files found:

```
```

### Other locations

Any other files created:

## Sync implications

Based on recon, what needs to sync for this emulator:

- Save data location: `~/Emulation/saves/ps2/` (memory cards)
- Save state location: `~/Emulation/states/pcsx2/`
- Any emulator-specific considerations:

## Issues found

List any bugs or improvements needed:

- [x] Add `.m3u` to PS2 extensions for multi-disc support - DONE
- [x] Config missing required version: need `[UI] SettingsVersion = 1` or PCSX2 rejects config as invalid - DONE
- [x] Config missing wizard flag: need `[UI] SetupWizardIncomplete = false` to skip wizard - DONE
- [x] Screenshots config key is `Snapshots` not `Screenshots` - FIXED
- [ ] Additional PCSX2 directories kyaraben could configure:
  - `Snapshots` - screenshots and GS dumps (currently `~/.config/PCSX2/snaps`)
  - `Cheats` - .pnach cheat files (currently `~/.config/PCSX2/cheats`)
  - `Covers` - game grid artwork (currently `~/.config/PCSX2/covers`)
  - `Videos` - video captures (currently `~/.config/PCSX2/videos`)
  - `Cache` - shaders, game list, achievements (currently `~/.config/PCSX2/cache`)

## Summary

| Device | Status | Tested by | Date |
|--------|--------|-----------|------|
| feanor | passed with issues | fausto | 2026-02-08 |
| steamdeck | not started | | |
