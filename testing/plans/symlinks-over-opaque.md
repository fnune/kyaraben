# Symlinks over opaque directories

## Status: implemented (Dolphin, Cemu)

## The idea

Replace the "opaque directory" model with a symlink-based approach. Instead of moving all emulator data into `opaque/{emulator}/` and treating it as a black box, we:

1. Let the emulator use its default data location (or a kyaraben-managed location via CLI flags)
2. Create symlinks FROM the emulator's expected subdirectories TO kyaraben's standard locations

Example for Dolphin:

```
~/.local/share/dolphin-emu/GC/          → ~/Emulation/saves/gamecube/
~/.local/share/dolphin-emu/Wii/         → ~/Emulation/saves/wii/
~/.local/share/dolphin-emu/StateSaves/  → ~/Emulation/states/dolphin/
~/.local/share/dolphin-emu/ScreenShots/ → ~/Emulation/screenshots/dolphin/
```

Or if using `-u` to isolate from system installs:

```
~/Emulation/emulators/dolphin/GC/          → ~/Emulation/saves/gamecube/
~/Emulation/emulators/dolphin/Wii/         → ~/Emulation/saves/wii/
~/Emulation/emulators/dolphin/StateSaves/  → ~/Emulation/states/dolphin/
~/Emulation/emulators/dolphin/ScreenShots/ → ~/Emulation/screenshots/dolphin/
```

## Why this is better

1. No more "opaque" concept - we're explicit about what we manage
2. All saves in `saves/`, all states in `states/`, all screenshots in `screenshots/`
3. Sync configuration becomes uniform across all emulators
4. Same mental model as non-opaque emulators, just with symlinks for paths that aren't configurable
5. Simpler docs - no need to explain opaque directories

## Concerns

### Emulator-dependent subdirectory structure

With symlinks, the emulator's internal structure leaks into the standard directories. For example, Dolphin creates regional subdirs:

```
~/Emulation/saves/gamecube/EUR/Card A/game.gci
~/Emulation/saves/gamecube/USA/Card A/game.gci
```

This is different from other emulators that might just put saves directly in the system folder.

This may be acceptable:
- Users rarely browse save directories manually
- The structure is deterministic and documented
- Sync still works correctly
- Each emulator's saves remain isolated by system folder

### Symlink compatibility

- Does the emulator follow symlinks correctly?
- What if the emulator tries to create the directory and it already exists as a symlink?
- What if the symlink target doesn't exist yet?

### Migration

If a user already has data in the emulator's directories, we need to:
1. Move the data to the kyaraben location
2. Replace the directory with a symlink
3. Handle failures gracefully

### Wii NAND complexity

Dolphin's Wii/ directory contains the full Wii NAND, which includes:
- Save data (what we want to sync)
- System settings (region-specific, maybe shouldn't sync)
- Downloaded channels/titles (large, maybe shouldn't sync)

May need to symlink deeper, e.g., `Wii/title/` only.

## Test plan

Test with Dolphin first since we have it set up.

### Preparation

1. Close Dolphin
2. Move existing saves from `opaque/dolphin/GC/` to `~/Emulation/saves/gamecube/`
3. Move existing savestates from `opaque/dolphin/StateSaves/` to `~/Emulation/states/dolphin/`
4. Remove the original directories
5. Create symlinks from original locations to new locations

### Verification

1. Launch Dolphin
2. Load a game, verify existing save appears
3. Create a new save, verify it writes to the symlinked location
4. Load a savestate, verify it works
5. Create a new savestate, verify it writes to the symlinked location
6. Take a screenshot, verify it writes to `~/Emulation/screenshots/dolphin/`

### Results

Tested 2026-02-09 on feanor:

1. Created symlinks:
   - `opaque/dolphin/GC` → `saves/gamecube/`
   - `opaque/dolphin/StateSaves` → `states/dolphin/`
   - `opaque/dolphin/ScreenShots` → `screenshots/dolphin/`

2. Launched Dolphin - existing saves appeared correctly

3. Created new savestate - wrote to `states/dolphin/GSAE01.s02`

4. Took screenshot - wrote to `screenshots/dolphin/GSAE01/*.png`

Conclusion: symlink approach works for Dolphin.

## If successful

- Remove the opaque directory model from kyaraben
- Update emulator definitions to use symlinks instead
- Update documentation to remove opaque directory explanations
- Simplify sync configuration

## Emulators to test

Current opaque emulators:
- [x] dolphin (GameCube/Wii)
- [x] cemu (Wii U) - game saves at usr/save/00050000/, screenshots at ~/.local/share/Cemu/screenshots/
- [x] azahar (3DS)
- [x] eden (Switch)
- [x] ppsspp (PSP)
- [ ] rpcs3 (PS3)
- [ ] vita3k (PS Vita)

Each may have different subdirectories that need symlinking.
