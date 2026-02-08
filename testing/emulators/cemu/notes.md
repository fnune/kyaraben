# Cemu

System(s): Wii U
Opaque: yes (MLC directory for saves, updates, DLC)

## Installation

- [x] Enable system, apply succeeds
- [x] Emulator appears in application menu
- [x] Emulator launches from menu
- [x] No onboarding wizard or first-run prompts

Notes: Controller profiles were empty, had to configure manually (expected, see testing/plans/controller-support-plan.md).

## BIOS/firmware

Required files: keys.txt (encryption keys, imported via UI)

- [ ] `kyaraben doctor` shows correct status before adding
- [ ] Files in correct location are detected
- [ ] Hash verification works

Notes: Provision marked as ImportViaUI. Feedback item notes games launch fine even when doctor says keys missing - investigate.

## ROM loading

- [x] ROM launches and runs (audio, video, input)

Test ROM used: The Legend of Zelda - Breath of the Wild (USA) (DLC) (v208).wua

### Game library

Does the emulator have a built-in game list/library?

- [x] Emulator supports game library: yes
- [x] Library can be pre-configured via config: yes (GamePaths in settings.xml)
- [x] Games from ROM directory appear in library

Notes: kyaraben configures GamePaths. 3 games visible (BotW, Wind Waker HD, Twilight Princess HD).

### Extensions

Current: .wua, .wud, .wux, .rpx, .elf

- [ ] Tested formats: .wua
- [ ] Missing formats:
- [ ] Unnecessary formats:

## Path configuration

- [x] Saves write to kyaraben location (via symlink)
- N/A Save states - Cemu does not support savestates
- [x] Screenshots write to kyaraben location (via symlink)

Notes:
- Saves: symlink `opaque/cemu/usr/save/00050000/` → `saves/wiiu/` (game title IDs inside)
- Screenshots: symlink `~/.local/share/Cemu/screenshots/` → `screenshots/cemu/`
- Only symlink 00050000 (games), not 00050010 (system) or system/ (account data)
- No savestate support due to Wii U hardware complexity

## Persistence

- [x] Save file persists after closing
- [x] Save loads correctly on re-launch

Notes: Tested with symlinked saves - BotW save loaded correctly after symlink migration.

## ES-DE integration

- [ ] System appears in ES-DE
- [ ] ROMs visible
- [ ] Scraping works
- [ ] Launching from ES-DE works

### Multi-disc (if applicable)

N/A for Wii U

## Post-testing recon

After running the emulator, document what was created.

### Config location

Default config path: `~/.config/Cemu/`

Files found:

```
settings.xml (kyaraben manages GamePaths)
```

Managed by kyaraben: settings.xml (GamePaths only)

Not managed (should be?):

### Data location

Default data path: n/a (uses MLC via -mlc flag)

### Cache location

Cache path: TBD

### Opaque directory (MLC)

Path: `~/Emulation/opaque/cemu/`

Structure:

```
sys/title/... (system titles)
usr/save/system/... (system save data - account, play diary)
usr/save/{title_id}/... (game saves)
```

What is machine-specific (should not sync): possibly account.dat?

What should sync: usr/save/{title_id}/ (game saves)

### Other locations

Any other files created:

## Sync implications

Based on recon, what needs to sync for this emulator:

- Save data location: opaque/cemu/usr/save/{title_id}/
- Save state location: TBD
- Any emulator-specific considerations: MLC structure is complex, may need selective sync

## Issues found

- Investigate: kyaraben doctor says keys.txt required but games launch fine?
- Update prompt on launch: fixed, kyaraben now sets check_update=false
- Version outdated: fixed, updated to 2.6

## Summary

| Device | Status | Tested by | Date |
|--------|--------|-----------|------|
| feanor | passed | fausto | 2026-02-09 |
| steamdeck | not started | | |
