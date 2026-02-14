# Azahar (3DS emulator)

## Status: implemented

## Config location

Azahar uses `~/.config/azahar-emu/qt-config.ini`.

## Data directories

Default data location: `~/.local/share/azahar-emu/`

Structure:
- `nand/` - NAND filesystem
- `sdmc/` - SD card filesystem (contains saves)
- `states/` - savestates
- `screenshots/` - screenshots
- `cheats/` - cheat files
- `shaders/` - shader cache
- `sysdata/` - system data (aes_keys.txt goes here)
- `load/textures/` - custom textures

## Critical setting: use_custom_storage

Azahar ignores custom nand/sdmc paths unless `use_custom_storage=true` is set.

Without this, all data goes to the default `~/.local/share/azahar-emu/` location.

## Required config entries

```ini
[Data%20Storage]
nand_directory=/path/to/nand/
nand_directory\default=false
sdmc_directory=/path/to/sdmc/
sdmc_directory\default=false
use_custom_storage=true
use_custom_storage\default=false

[UI]
Paths\gamedirs\1\path=INSTALLED
Paths\gamedirs\2\path=SYSTEM
Paths\gamedirs\3\path=/path/to/roms/n3ds
Paths\gamedirs\size=3
Paths\screenshotPath=/path/to/screenshots/
Paths\screenshotPath\default=false
```

Note: game directory uses index 3 because 1=INSTALLED and 2=SYSTEM are built-in virtual directories.

## Implementation

Kyaraben now:
- Configures sdmc directly to `saves/n3ds/`
- Sets `use_custom_storage=true` so paths are respected
- Configures screenshot path via `Paths\screenshotPath`
- Symlinks `~/.local/share/azahar-emu/states/` to `states/azahar/` (no config option for states)
- Uses game directory index 3 to preserve INSTALLED/SYSTEM virtual directories
- No provisions: Azahar removed encrypted game support in 2025, ROMs must be pre-decrypted

## Reference

EmuDeck config: `configs/azahar/qt-config.ini`
EmuDeck script: `functions/EmuScripts/emuDeckAzahar.sh`
