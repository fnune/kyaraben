# RetroArch bsnes

System(s): SNES
Opaque: no

## Installation

- [x] Enable system, apply succeeds
- [x] Emulator appears in application menu
- [x] Emulator launches from menu
- [x] No onboarding wizard or first-run prompts

Notes: RetroArch launches cleanly.

## BIOS/firmware

No BIOS required for SNES.

## ROM loading

- [x] ROM in correct folder visible to emulator
- [x] ROM launches and runs (audio, video, input)

Test ROM used: Legend of Zelda - A Link to the Past (USA).zip, Kirby Super Star (USA).sfc

### Extensions

Current: `.sfc`, `.smc`, `.bs`

- [x] Tested formats: `.zip`, `.sfc`
- [x] Missing formats: `.zip` not in kyaraben's extension list (works in RetroArch but ES-DE won't find them)
- [ ] Unnecessary formats:

## Path configuration

- [x] Saves write to `~/Emulation/saves/snes/` - WORKS via symlink
- [x] Save states write to `~/Emulation/states/retroarch:bsnes/` - WORKS via symlink
- [x] Screenshots write to `~/Emulation/screenshots/retroarch/` - WORKS via screenshot_directory setting

### Investigation findings

1. Override file naming: RetroArch uses short core name (`bsnes`), not full name (`bsnes_libretro`)
   - Wrong: `~/.config/retroarch/config/bsnes_libretro/bsnes_libretro.cfg`
   - Correct: `~/.config/retroarch/config/bsnes/bsnes.cfg`
   - With correct name, Quick Menu -> Overrides shows "configuration override loaded"

2. Partial override support: Even with correct override file:
   - `savestate_directory` WORKS - savestates go to override path
   - `savefile_directory` BROKEN - saves still go to main config path
   - This is a known RetroArch bug: https://github.com/libretro/RetroArch/issues/13686

3. Empty string approach fails: Setting `savefile_directory = ""` in main config causes RetroArch to reset it to default on launch

## Persistence

- [ ] Save file persists after closing
- [ ] Save loads correctly on re-launch

Notes:

## ES-DE integration

- [x] System appears in ES-DE
- [ ] ROMs visible - only `.sfc` files shown, `.zip` files missing due to extension list
- [ ] Scraping works
- [x] Launching from ES-DE works (Kirby loaded fine)

Notes: ES-DE only finds Kirby Super Star (.sfc), not the .zip ROMs. Launching works.

### Multi-disc

Not applicable for SNES.

## Post-testing recon

After running the emulator, document what was created.

### Config location

Default config path: `~/.config/retroarch/`

Files found:

```
retroarch.cfg (main config)
config/bsnes_libretro/bsnes_libretro.cfg (per-core config)
states/*.state (save states)
saves/*.srm (SRAM saves)
```

Managed by kyaraben: `config/bsnes_libretro/bsnes_libretro.cfg`

Not managed (should be?): `retroarch.cfg` main config paths

### Data location

Default data path: `~/.local/share/retroarch/`

Files found: not checked

### Cache location

Cache path: `~/.cache/retroarch/`

Files found: not checked

### Other locations

Any other files created:

## Sync implications

Based on recon, what needs to sync for this emulator:

- Save data location: currently `~/.config/retroarch/saves/` (should be `~/Emulation/saves/snes/`)
- Save state location: currently `~/.config/retroarch/states/` (should be `~/Emulation/states/retroarch:bsnes/`)
- Any emulator-specific considerations: RetroArch main config overrides per-core paths

## Issues found

List any bugs or improvements needed:

- [x] Missing `.zip` extension in SNES system definition - FIXED
- [x] Override file uses wrong name: should be `bsnes/bsnes.cfg` not `bsnes_libretro/bsnes_libretro.cfg` - FIXED
- [x] `savefile_directory` override broken in RetroArch (known bug) - WORKAROUND via symlinks

## Proposed solution: symlink approach

Since RetroArch's per-core override system is buggy for directory settings, use symlinks instead:

1. Enable sorting in main `retroarch.cfg`:
   ```
   sort_savefiles_enable = "true"
   sort_savestates_enable = "true"
   ```

2. This makes RetroArch create per-core subdirectories:
   - `~/.config/retroarch/saves/bsnes/`
   - `~/.config/retroarch/states/bsnes/`

3. Kyaraben creates symlinks during `apply`:
   - `~/.config/retroarch/saves/bsnes/` → `~/Emulation/saves/snes/`
   - `~/.config/retroarch/states/bsnes/` → `~/Emulation/states/retroarch:bsnes/`

4. Remove per-core override files since they don't work reliably

Benefits:
- Sidesteps broken RetroArch override system entirely
- RetroArch does its thing, symlinks redirect to kyaraben dirs
- Works for both saves and savestates
- Per-system organization preserved via symlinks

Implementation:
- Update `SharedConfig()` to enable sorting
- Add symlink creation to apply process for each RetroArch core
- Change core name from `bsnes_libretro` to `bsnes` (short name)
- Remove per-core `.cfg` files (or keep only for non-directory settings like `rgui_browser_directory`)

## Summary

| Device | Status | Tested by | Date |
|--------|--------|-----------|------|
| feanor | symlink approach implemented | fausto | 2026-02-08 |
| steamdeck | not started | | |
