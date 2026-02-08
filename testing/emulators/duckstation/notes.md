# DuckStation

System(s): PSX
Opaque: no

## Installation

- [x] Enable system, apply succeeds
- [x] Emulator appears in application menu
- [x] Emulator launches from menu
- [ ] No onboarding wizard or first-run prompts - FAILS, see issues

Notes: Launches fine but shows first-run wizard. Wizard overwrites kyaraben's absolute paths with relative paths.

## BIOS/firmware

Required files:
- `scph5501.bin` (USA) - MD5: `490f666e1afb15b7362b406ed1cea246` - required
- `scph5500.bin` (Japan) - MD5: `8dd7d5296a650fac7319bce665a6a53c` - optional
- `scph5502.bin` (Europe) - MD5: `32736f17079d0b2b7024407c39bd3050` - optional

Location: `~/Emulation/bios/psx/`

- [x] `kyaraben doctor` shows correct status before adding
- [ ] Files in correct location are detected - FAILS, see issues
- [ ] Hash verification works - N/A due to above

Notes: Doctor correctly shows required BIOS missing. User's BIOS (`SCPH1001.BIN`) works in DuckStation but kyaraben doesn't recognize it due to filename/hash mismatch.

## ROM loading

- [x] ROM in correct folder visible to emulator
- [ ] ROM launches and runs (audio, video, input)

Test ROMs: DoDonPachi (Japan).chd, Breath of Fire III (Europe).chd
Note: Castlevania SOTN has .bin but no .cue file - not detected

### Extensions

Current: `.bin`, `.cue`, `.chd`, `.iso`, `.img`, `.m3u`

- [x] Tested formats: `.chd`
- [ ] Missing formats: none identified
- [ ] Unnecessary formats: none identified

## Path configuration

- [ ] Saves write to `~/Emulation/saves/psx/` - FAILS (goes to `~/.config/duckstation/memcards/`)
- [ ] Save states write to `~/Emulation/states/duckstation/` - FAILS (goes to `~/.config/duckstation/savestates/`)
- [ ] Screenshots write to `~/Emulation/screenshots/psx/` - not tested

Notes: Wizard overwrites kyaraben's absolute paths with relative paths. Fix requires `SetupWizardIncomplete = false`.

## Persistence

- [ ] Save file persists after closing
- [ ] Save loads correctly on re-launch

Notes:

## ES-DE integration

- [x] System appears in ES-DE
- [x] ROMs visible (all 3 games, including .bin-only Castlevania)
- [ ] Scraping works
- [x] Launching from ES-DE works

### Multi-disc

PSX commonly has multi-disc games.

- [ ] `.m3u` extension supported
- [ ] Discs in `.hidden/` folder not shown
- [ ] `.m3u` shows as single entry
- [ ] Disc switching works in-game

Notes:

## Post-testing recon

After running the emulator, document what was created.

### Config location

Default config path: `~/.config/duckstation/`

Files found:

```
```

Managed by kyaraben: `settings.ini` (BIOS, MemoryCards, SaveStates, Screenshots, GameList paths)

Not managed (should be?):

### Data location

Default data path: `~/.local/share/duckstation/`

Files found:

```
```

### Cache location

Cache path: `~/.cache/duckstation/`

Files found:

```
```

### Other locations

Any other files created:

## Sync implications

Based on recon, what needs to sync for this emulator:

- Save data location: `~/Emulation/saves/psx/` (memory cards)
- Save state location: `~/Emulation/states/duckstation/`
- Any emulator-specific considerations:

## Issues found

List any bugs or improvements needed:

- [x] First-run wizard appears and overwrites kyaraben paths - FIXED using SettingsVersion = 3 (same as EmuDeck). DuckStation checks for either SetupWizardIncomplete OR SettingsVersion to determine if setup is needed.
- [ ] BIOS detection too strict (cross-cutting): user's `SCPH1001.BIN` works in DuckStation but kyaraben rejects it because:
  - Case mismatch (expected lowercase)
  - Different version/hash (SCPH1001 vs SCPH5501)
  - Need: case-insensitive matching, multiple accepted hashes per provision, "one of required" semantics
  - Valid alternative: `SCPH1001.BIN` (US/Canada) MD5: `924e392ed05558ffdb115408c263dccf`
  - Valid alternative: `ps-41e.bin` (Europe) MD5: `b9d9a0286c33dc6b7237bb13cd46fdee`

## Summary

| Device | Status | Tested by | Date |
|--------|--------|-----------|------|
| feanor | passed with issues | fausto | 2026-02-08 |
| steamdeck | not started | | |
