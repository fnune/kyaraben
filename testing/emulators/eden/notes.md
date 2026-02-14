# Eden

System(s): Nintendo Switch
Opaque: no (uses symlinks for keys, firmware, saves, screenshots)

## Installation

- [x] Enable system, apply succeeds
- [x] Emulator appears in application menu
- [x] Emulator launches from menu
- [x] No onboarding wizard or first-run prompts

Notes: Shows "Decryption keys are missing" dialog on first launch if keys not present. This is expected and informative rather than a blocking wizard.

## BIOS/firmware

Required files:
- `prod.keys` (production keys, required)
- `title.keys` (title keys, optional - for DLC and updates)
- `*.nca` (firmware files, optional - enables system applets, placed directly in bios/switch/)

- [x] `kyaraben doctor` shows correct status before adding
- [x] Files in correct location are detected
- [x] Hash verification works (n/a for keys - filename only check)

Notes:
- Keys (`*.keys` files) and firmware (`*.nca` files) coexist in `~/Emulation/bios/switch/`
- Both `~/.local/share/eden/keys/` and `~/.local/share/eden/nand/system/Contents/registered/` symlink to `bios/switch/`
- Eden source confirms keys import is just a file copy (no transformation)

## ROM loading

- [x] ROM launches and runs (audio, video, input)

Test ROM used: Kirby and the Forgotten Land

### Game library

Does the emulator have a built-in game list/library?

- [x] Emulator supports game library: yes
- [x] Library can be pre-configured via config: yes (qt-config.ini Paths\gamedirs)
- [x] Games from ROM directory appear in library

Notes: kyaraben configures the game directory path in qt-config.ini.

### Extensions

Current: .nsp, .xci, .nca, .nro, .nso

- [x] Tested formats: .nsp
- [ ] Missing formats:
- [ ] Unnecessary formats:

## Path configuration

- [x] Saves write to `~/Emulation/saves/switch/` (via symlink from `nand/user/save/`)
- N/A Save states - Eden does not expose savestate hotkeys via CLI
- [x] Screenshots write to `~/Emulation/screenshots/eden/` (via symlink)

Notes:
- Symlinks created by kyaraben (keys and firmware both point to same dir since they use different extensions):
  - `~/.local/share/eden/keys/` → `~/Emulation/bios/switch/`
  - `~/.local/share/eden/nand/system/Contents/registered/` → `~/Emulation/bios/switch/`
  - `~/.local/share/eden/nand/user/save/` → `~/Emulation/saves/switch/`
  - `~/.local/share/eden/screenshots/` → `~/Emulation/screenshots/eden/`

## Persistence

- [x] Save file persists after closing
- [x] Save loads correctly on re-launch

Notes: Tested with Kirby - save persisted and loaded correctly after symlink setup.

## ES-DE integration

- [x] System appears in ES-DE
- [x] ROMs visible
- [ ] Scraping works
- [x] Launching from ES-DE works

### Multi-disc (if applicable)

N/A for Switch

## Updates and DLC

Eden requires updates and DLC to be installed via the UI "Install Files to NAND" feature. This extracts NSP containers into individual NCA files at `~/.local/share/eden/nand/user/Contents/registered/`. This cannot be automated or symlinked because:

1. NSP is a container format that must be extracted
2. NCAs are placed in a hash-based directory structure
3. Eden's CLI has no install option (only game launching)

Recommended workflow for users:
1. Create a directory `~/Emulation/roms/switch/Updates/` (or `Patches/`, `DLC/`, etc.)
2. Add a `noload.txt` file inside so ES-DE ignores the folder
3. Store NSP update/DLC files there for organization
4. Use Eden UI (Tools > Install Files to NAND) to import them

The installed content lives at `~/.local/share/eden/nand/user/Contents/registered/` and is not currently managed by kyaraben.

## Post-testing recon

After running the emulator, document what was created.

### Config location

Default config path: `~/.config/eden/`

Files found:

```
qt-config.ini (kyaraben manages UI section paths)
```

Managed by kyaraben: qt-config.ini (screenshot_path, gamedirs)

Not managed (should be?): None identified

### Data location

Default data path: `~/.local/share/eden/`

Symlinks created by kyaraben:

```
~/.local/share/eden/keys/                              → ~/Emulation/bios/switch/
~/.local/share/eden/nand/system/Contents/registered/   → ~/Emulation/bios/switch/
~/.local/share/eden/nand/user/save/                    → ~/Emulation/saves/switch/
~/.local/share/eden/screenshots/                       → ~/Emulation/screenshots/eden/
```

Other data (not symlinked):

```
nand/user/Contents/registered/ (installed updates/DLC - see "Updates and DLC" section)
log/                           (logs)
shader/                        (shader cache - machine-specific)
```

### Cache location

Cache path: `~/.cache/eden/` (if exists)

Files found: TBD

### Other locations

Any other files created: None outside standard XDG locations

## Sync implications

Based on recon, what needs to sync for this emulator:

- Save data location: `saves/switch/`
- Save state location: N/A
- Keys and firmware: `bios/switch/` (keys are `*.keys` files, firmware are `*.nca` files)
- Screenshots: `screenshots/eden/`
- Updates/DLC: Not synced (installed per-device via Eden UI)
- Shader cache: Should NOT sync (machine-specific)

## Issues found

- Savestates: Eden supports them but CLI doesn't expose hotkeys, so can't be configured via kyaraben
- Updates/DLC workflow requires manual UI import (documented above)

## Summary

| Device | Status | Tested by | Date |
|--------|--------|-----------|------|
| feanor | passed | fausto | 2026-02-10 |
| steamdeck | not started | | |
