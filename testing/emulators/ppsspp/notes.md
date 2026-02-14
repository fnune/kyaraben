# PPSSPP (PSP emulator)

## Status: investigating

## Key finding: config paths are ignored

PPSSPP ignores `MemStickDirectory` and `ScreenshotsPath` config settings. Both EmuDeck and our testing confirm this - saves/states/screenshots go to the default location regardless of config.

EmuDeck's solution: symlinks (not config).

## PPSSPP directory structure

Default location: `~/.config/ppsspp/PSP/`

```
PSP/
├── SAVEDATA/      # in-game saves
├── PPSSPP_STATE/  # savestates (.ppst + .jpg thumbnail)
├── SCREENSHOT/    # screenshots
├── SYSTEM/        # config (ppsspp.ini)
├── GAME/          # homebrew, plugins
├── TEXTURES/      # texture packs
├── CHEATS/        # cheat files
└── PLUGINS/       # plugins
```

## EmuDeck approach

From `emuDeckPPSSPP.sh`:
- Only configures `CurrentDirectory` (ROMs path) via INI
- Uses `linkToSaveFolder` to create symlinks for saves/states
- Symlink direction: `Emulation/saves/ppsspp/saves` → `~/.config/ppsspp/PSP/SAVEDATA`

Note: EmuDeck's symlink direction is reversed from kyaraben's - they link TO the emulator dir, we link FROM it.

## Config that works

- `General.CurrentDirectory` - ROMs path (works)

## Config that doesn't work

- `General.MemStickDirectory` - ignored
- `General.ScreenshotsPath` - ignored

## Symlinks needed for kyaraben

Since config paths are ignored, we need symlinks FROM emulator TO standard dirs:

| Source | Target |
|--------|--------|
| `~/.config/ppsspp/PSP/SAVEDATA` | `saves/psp/` |
| `~/.config/ppsspp/PSP/PPSSPP_STATE` | `states/ppsspp/` |
| `~/.config/ppsspp/PSP/SCREENSHOT` | `screenshots/ppsspp/` |

## Implementation plan

1. Remove opaque dir usage
2. Remove MemStickDirectory/ScreenshotsPath config (doesn't work)
3. Keep CurrentDirectory config for ROMs path
4. Add Symlinks method for SAVEDATA, PPSSPP_STATE, SCREENSHOT

## ROM formats

PPSSPP can play directly from ZIP files (no extraction needed).

Also supports: ISO, CSO, EBOOT.PBP

## Testing

Games tested:
- Final Fantasy Tactics: War of the Lions (ULES00850)
- Lumines (ULES00043)
