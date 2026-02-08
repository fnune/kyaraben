# mGBA

System(s): GB, GBC, GBA
Opaque: no

## Installation

- [x] Enable system, apply succeeds
- [x] Emulator appears in application menu
- [x] Emulator launches from menu
- [x] No onboarding wizard or first-run prompts

Notes: Opens cleanly with no wizard.

## BIOS/firmware

Optional files (mGBA has HLE):
- `gba_bios.bin` - MD5: `a860e8c0b6d573d191e4ec7db1b1e4f6` - optional

Location: `~/Emulation/bios/gba/`

- [ ] `kyaraben doctor` shows correct status before adding
- [ ] Files in correct location are detected
- [ ] Hash verification works

Notes:

## ROM loading

- [x] ROM launches and runs (audio, video, input)

Test ROM used: Fire Emblem (Europe) (En,Fr,De).zip

### Game library

Does the emulator have a built-in game list/library?

- [x] Emulator supports game library: yes (Tools > Game Library)
- [ ] Library can be pre-configured via config: NO - mGBA has no config key for ROM directories
- [ ] Games from ROM directory appear in library

Notes: mGBA remembers last browsed directory but cannot be pre-configured. User must manually browse to ROM folder on first use. Library has "Show when no game open" option in settings.

### Extensions

Current:
- GB: `.gb`
- GBC: `.gbc`
- GBA: `.gba`

- [x] Tested formats: `.zip` (GBA)
- [x] Missing formats: `.zip` works but not in extension list
- [ ] Unnecessary formats:

## Path configuration

Note: mGBA config points saves/states/screenshots to GBA paths. GB/GBC saves go where?

- [x] Saves write to `~/Emulation/saves/gba/`
  - `Fire Emblem (Europe) (En,Fr,De).sav` (32768 bytes)
- [x] Save states write to `~/Emulation/states/mgba/`
  - `Fire Emblem (Europe) (En,Fr,De).ss1` (102321 bytes)
- [x] Screenshots write to `~/Emulation/screenshots/gba/`
  - `Fire Emblem (Europe) (En,Fr,De)-0.png` (36120 bytes)

Notes: Save filename matches ROM name with `.sav` extension. Savestate uses `.ss1` for slot 1. Screenshot uses `-0` suffix (increments).

## Persistence

- [x] Save file persists after closing
- [x] Save loads correctly on re-launch

Notes: Works as expected.

## ES-DE integration

ES-DE config location: `~/ES-DE/custom_systems/es_systems.xml`

### ~~Critical issue: ES-DE integration architecture~~ RESOLVED

Initial diagnosis was incorrect. ES-DE custom_systems DOES work for overriding bundled systems. The override mechanism is based on matching `<path>` tags and whether ROMs are found.

The actual issue was that our system definitions had:
- Missing extensions (e.g. .zip for N64)
- Wrong paths (e.g. SystemID "3ds" creating path `%ROMPATH%/3ds` when ROMs were in `n3ds/`)

When our custom definition doesn't find ROMs (wrong path or missing extension), ES-DE's bundled definition finds them instead, making it appear that bundled "wins".

Fix applied: ensure paths match folder conventions (renamed SystemID3DS to SystemIDN3DS) and extensions are complete (added .zip to N64).

TODO: audit all system extensions against ES-DE's bundled config.

Reference files saved to `testing/reference/`:
- esde-builtin-linux-systems.xml - source for system definitions
- esde-builtin-linux-find-rules.xml - emulator name mappings
- esde-userguide.md, esde-install.md, esde-faq.md

Reference: ES-DE built-in find rules downloaded to `testing/reference/esde-builtin-linux-systems.xml`

- [ ] GB system appears in ES-DE
- [ ] GBC system appears in ES-DE
- [ ] GBA system appears in ES-DE
- [ ] ROMs visible (`.zip` already in ES-DE built-in extensions)
- [ ] Scraping works
- [ ] Launching from ES-DE works - BLOCKED by find rules issue

### Multi-disc

Not applicable for GB/GBC/GBA.

## Post-testing recon

After running the emulator, document what was created.

### Config location

Default config path: `~/.config/mgba/`

Files found:

```
~/.config/mgba/config.ini
~/.config/mgba/qt.ini
~/.config/mgba/library.sqlite3
~/.config/mgba/nointro.sqlite3
```

Managed by kyaraben: `config.ini` (bios, savegamePath, savestatePath, screenshotPath)

Not managed (should be?):
- `qt.ini` - Qt window state, recent files, etc. Probably fine to leave unmanaged.
- `library.sqlite3` - game library database. Could be useful for sync if user curates library.
- `nointro.sqlite3` - No-Intro database for game identification. Machine-generated, no need to sync.

### Data location

Default data path: `~/.local/share/mgba/`

Files found:

```
(none)
```

### Cache location

Cache path: `~/.cache/mgba/`

Files found:

```
(none)
```

### Other locations

Any other files created: None found.

## Sync implications

Based on recon, what needs to sync for this emulator:

- Save data location: `~/Emulation/saves/gba/*.sav`
- Save state location: `~/Emulation/states/mgba/*.ss{1-9}`
- Screenshots: `~/Emulation/screenshots/gba/*.png` (optional sync)
- Config sync: `~/.config/mgba/library.sqlite3` (optional, if user curates library)
- Skip: `~/.config/mgba/nointro.sqlite3` (machine-generated)
- Skip: `~/.config/mgba/qt.ini` (window state, machine-specific)

Any emulator-specific considerations: GB/GBC saves currently go to same `gba/` dir. Acceptable since mGBA handles all three systems.

## Issues found

List any bugs or improvements needed:

- [x] Add `showLibrary=1` to managed config keys (improves first-run UX) - FIXED
- [x] Library ROM directories: CONFIRMED cannot pre-configure (mGBA stores file paths, not directories; inserting directory paths doesn't work)
- [ ] Multi-system save paths: mGBA handles GB, GBC, GBA but kyaraben config points all saves to `gba/` directory. Should saves be separated by system or is a shared folder acceptable?
- [x] Missing `.zip` extension: works in mGBA but not in kyaraben system extension list - FIXED (added to GB, GBC, GBA)
- [x] ES-DE integration BROKEN: kyaraben's find rules don't define per-emulator paths, so ES-DE can't find standalone emulators. See ES-DE integration section for details. - FIXED (custom es_systems.xml approach)

## Architectural considerations

### Multi-system emulator save paths

mGBA supports GB, GBC, and GBA but kyaraben currently points all saves to a single system directory (`gba/`). This happens because `ConfigGenerator.Generate()` can only specify one path.

Research confirmed: mGBA has a single `savegamePath` config option with no per-system settings.

Options:
1. Use emulator-specific saves dir: `~/Emulation/saves/mgba/` - honest about the limitation
2. Accept shared `gba/` directory - saves are named after ROM so no collision risk
3. Switch to RetroArch cores: use `retroarch:gambatte` for GB/GBC and `retroarch:mgba` for GBA. Each core gets its own per-core override config, so paths can be properly separated by system through the existing RetroArch path configuration system

Option 3 is cleanest for per-system organization but requires users to use RetroArch instead of standalone mGBA.

### Game library pre-configuration

mGBA has a library feature. Two parts to configure:

**1. Show library on startup (config.ini):**
```ini
[ports.qt]
showLibrary=1
```
Kyaraben should set this as a default managed key.

**2. Library ROM directories (library.sqlite3):**

mGBA stores each ROM file as a separate "root" entry (unusual design):
```sql
-- roots table (stores full file paths, not directories):
2|/home/fausto/Emulation/roms/gba/Fire Emblem.zip|0
1|/home/fausto/Emulation/roms/gba/Metroid Fusion.zip|0

-- paths table (stores filename inside archive + rootid FK):
1|1|Fire Emblem.gba|0|1|
2|2|Metroid Fusion.gba|0|2|
```

Tested: Inserting directory paths does NOT work. mGBA only recognizes individual file paths in `roots`. When user adds a directory via UI, mGBA scans and inserts each file separately.

Conclusion: Library cannot be pre-configured by kyaraben. User must add directories manually via mGBA UI. Setting `showLibrary=1` at least ensures the library is visible on startup.

**qt.ini** only stores window state and MRU, not library config.

## Summary

| Device | Status | Tested by | Date |
|--------|--------|-----------|------|
| feanor | passed with issues | fausto | 2026-02-08 |
| steamdeck | not started | | |
