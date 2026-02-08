# melonDS

System(s): NDS
Opaque: no

## Installation

- [x] Enable system, apply succeeds
- [x] Emulator appears in application menu
- [x] Emulator launches from menu
- [x] No onboarding wizard or first-run prompts

Notes: Launches cleanly.

## BIOS/firmware

Optional files (melonDS has HLE for most games):
- `bios7.bin` (ARM7) - MD5: `df692a80a5b1bc90728bc3dfc76cd948` - optional
- `bios9.bin` (ARM9) - MD5: `a392174eb3e572fed6447e956bde4b25` - optional
- `firmware.bin` - MD5: `e45033d9b0fa6b0de071292bba7c9d13` - optional

Location: `~/Emulation/bios/nds/`

- [ ] `kyaraben doctor` shows correct status before adding
- [ ] Files in correct location are detected
- [ ] Hash verification works

Notes: Not tested (HLE works fine).

## ROM loading

- [ ] ROM in correct folder visible to emulator - NO, melonDS has no library view by default
- [x] ROM launches and runs (audio, video, input)

Test ROM used: Kirby Super Star Ultra (Europe).zip

### Extensions

Current: `.nds`, `.dsi`

- [x] Tested formats: `.zip` (contains .nds)
- [ ] Missing formats: `.zip` works but may not be in extension list
- [ ] Unnecessary formats:

## Path configuration

- [ ] Saves write to `~/Emulation/saves/nds/` - needs verification after TOML fix
- [ ] Save states write to `~/Emulation/states/melonds/` - needs verification after TOML fix
- [ ] Screenshots write to `~/Emulation/screenshots/nds/` - not tested

Notes: Fixed config format (now TOML). Paths configured in [Instance0] section. Launches correctly from application menu.

## Persistence

- [ ] Save file persists after closing
- [ ] Save loads correctly on re-launch

Notes:

## ES-DE integration

- [x] System appears in ES-DE
- [x] ROMs visible (including .zip)
- [ ] Scraping works
- [ ] Launching from ES-DE works - FAILS, uses RetroArch instead of standalone melonDS

Notes: Same `<loadExclusive/>` issue as mGBA. Bundled ES-DE config wins, defaults to RetroArch melonDS core.

### Multi-disc

Not applicable for NDS.

## Post-testing recon

After running the emulator, document what was created.

### Config location

Default config path: `~/.config/melonDS/`

Files found:

```
melonDS.toml (actual config - melonDS ignores .ini now)
melonDS.ini (kyaraben writes this - IGNORED)
rtc.bin
```

Managed by kyaraben: `melonDS.ini` - BUT THIS IS WRONG, melonDS uses TOML now

Not managed (should be?): `melonDS.toml` - this is the actual config file

### Data location

Default data path: `~/.local/share/melonDS/`

Files found:

```
```

### Cache location

Cache path: `~/.cache/melonDS/`

Files found:

```
```

### Other locations

Any other files created:

## Sync implications

Based on recon, what needs to sync for this emulator:

- Save data location: `~/Emulation/saves/nds/`
- Save state location: `~/Emulation/states/melonds/`
- Any emulator-specific considerations:

## Issues found

List any bugs or improvements needed:

- [x] Config format wrong: kyaraben writes `melonDS.ini` but melonDS now uses `melonDS.toml`. Need to switch config generator to TOML format. - FIXED, now uses TOML with [DS] and [Instance0] sections
- [ ] TOML paths are in `[Instance0]` section:
  ```toml
  [Instance0]
  SaveFilePath = "/home/fausto/Emulation/saves/nds"
  SavestatePath = "/home/fausto/Emulation/states/melonds"
  CheatFilePath = "/home/fausto/Emulation/cheats/nds"  # if we add cheats support
  ```
- [ ] No built-in game library - user must browse to ROM folder manually (like mGBA)
- [ ] melonDS does not support configurable screenshots directory

## Summary

| Device | Status | Tested by | Date |
|--------|--------|-----------|------|
| feanor | retesting after fix | fausto | 2026-02-08 |
| steamdeck | not started | | |
