# {Emulator name}

System(s): {systems}
Opaque: yes/no

## Installation

- [ ] Enable system, apply succeeds
- [ ] Emulator appears in application menu
- [ ] Emulator launches from menu
- [ ] No onboarding wizard or first-run prompts

Notes:

## BIOS/firmware

Required files: {list or "none"}

- [ ] `kyaraben doctor` shows correct status before adding
- [ ] Files in correct location are detected
- [ ] Hash verification works

Notes:

## ROM loading

- [ ] ROM launches and runs (audio, video, input)

Test ROM used:

### Game library

Does the emulator have a built-in game list/library?

- [ ] Emulator supports game library: yes / no
- [ ] Library can be pre-configured via config: yes / no
- [ ] Games from ROM directory appear in library

Notes:

### Extensions

Current: {list from system definition}

- [ ] Tested formats:
- [ ] Missing formats:
- [ ] Unnecessary formats:

## Path configuration

- [ ] Saves write to `~/Emulation/saves/{system}/`
- [ ] Save states write to `~/Emulation/states/{emulator}/`
- [ ] Screenshots write to `~/Emulation/screenshots/{system}/`

Notes:

## Persistence

- [ ] Save file persists after closing
- [ ] Save loads correctly on re-launch

Notes:

## ES-DE integration

- [ ] System appears in ES-DE
- [ ] ROMs visible
- [ ] Scraping works
- [ ] Launching from ES-DE works

### Multi-disc (if applicable)

- [ ] `.m3u` extension supported
- [ ] Discs in `.hidden/` folder not shown
- [ ] `.m3u` shows as single entry
- [ ] Disc switching works in-game

Notes:

## Post-testing recon

After running the emulator, document what was created.

### Config location

Default config path: `~/.config/{emulator}/`

Files found:

```
(tree or ls output)
```

Managed by kyaraben:

Not managed (should be?):

### Data location

Default data path: `~/.local/share/{emulator}/`

Files found:

```
```

### Cache location

Cache path: `~/.cache/{emulator}/`

Files found:

```
```

### Opaque directory (if applicable)

Path: `~/Emulation/opaque/{emulator}/`

Structure:

```
```

What is machine-specific (should not sync):

What should sync:

### Other locations

Any other files created:

## Sync implications

Based on recon, what needs to sync for this emulator:

- Save data location:
- Save state location:
- Any emulator-specific considerations:

## Issues found

List any bugs or improvements needed:

## Summary

| Device | Status | Tested by | Date |
|--------|--------|-----------|------|
| feanor | not started | | |
| steamdeck | not started | | |
