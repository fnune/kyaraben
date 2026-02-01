# Emulator Configuration Reference

This document provides a comprehensive reference for configuring emulators supported by kyaraben via text files. It serves as a guide for implementing configuration management within kyaraben to integrate with its UserStore structure and syncing capabilities.

## Table of Contents

1. [DuckStation (PSX)](#duckstation-psx)
2. [PCSX2 (PS2)](#pcsx2-ps2)
3. [RPCS3 (PS3)](#rpcs3-ps3)
4. [PPSSPP (PSP)](#ppsspp-psp)
5. [Vita3K (PS Vita)](#vita3k-ps-vita)
6. [mGBA (GB/GBC/GBA)](#mgba-gbgbcgba)
7. [melonDS (NDS)](#melonds-nds)
8. [Azahar (3DS)](#azahar-3ds)
9. [Dolphin (GameCube/Wii)](#dolphin-gamecubewii)
10. [Cemu (Wii U)](#cemu-wii-u)
11. [Eden (Switch)](#eden-switch)
12. [RetroArch (SNES, NES, Genesis, Saturn, N64)](#retroarch-snes-nes-genesis-saturn-n64)
13. [Flycast (Dreamcast)](#flycast-dreamcast)
14. [Controller Configuration](#controller-configuration)
15. [Summary Matrix](#summary-matrix)

---

## DuckStation (PSX)

### Config File

| Property | Value |
|----------|-------|
| Format | INI |
| Location | `~/.config/duckstation/settings.ini` |
| Base Dir | `$XDG_CONFIG_HOME/duckstation` or `~/.config/duckstation` |

### Path Configuration

| Setting | INI Section | INI Key | Description |
|---------|-------------|---------|-------------|
| BIOS Directory | `[BIOS]` | `SearchDirectory` | Directory containing PSX BIOS files |
| Memory Cards | `[MemoryCards]` | `Directory` | Directory for `.mcd` memory card files |
| Save States | `[Folders]` | `SaveStates` | Directory for save states |
| Screenshots | `[Folders]` | `Screenshots` | Directory for screenshots |
| Game List | `[GameList]` | `RecursivePaths` | ROM directories (semicolon-separated) |

### Example Configuration

```ini
[BIOS]
SearchDirectory = /home/user/Emulation/bios/psx

[MemoryCards]
Directory = /home/user/Emulation/saves/psx

[Folders]
SaveStates = /home/user/Emulation/states/duckstation
Screenshots = /home/user/Emulation/screenshots/psx

[GameList]
RecursivePaths = /home/user/Emulation/roms/psx
```

### Provisions Required

- **BIOS**: `scph5501.bin` (USA), `scph5500.bin` (Japan), `scph5502.bin` (Europe)
- BIOS files must be placed directly in the BIOS directory (not in subdirectories)

### Controller Configuration

- Controller profiles stored in `inputprofiles/` subdirectory
- Config key: `[Controller1]` section with `Type` and button mappings
- Supports SDL, XInput, and DInput backends
- SDL controller database: place `gamecontrollerdb.txt` in user directory

### Portable Mode

Create an empty `portable.txt` file alongside the executable to use portable mode.

### Manual Configuration Required

None - all paths configurable via text files.

### Sources

- [DuckStation GitHub](https://github.com/stenzek/duckstation)
- [EmuDeck DuckStation Documentation](https://emudeck.github.io/emulators/steamos/duckstation/)

---

## PCSX2 (PS2)

### Config File

| Property | Value |
|----------|-------|
| Format | INI |
| Location | `~/.config/PCSX2/inis/PCSX2.ini` |
| Base Dir | `$XDG_CONFIG_HOME/PCSX2` or `~/.config/PCSX2` |

### Path Configuration

| Setting | INI Section | INI Key | Description |
|---------|-------------|---------|-------------|
| BIOS Directory | `[Folders]` | `Bios` | Directory containing PS2 BIOS files |
| Memory Cards | `[Folders]` | `MemoryCards` | Directory for memory card files |
| Save States | `[Folders]` | `Savestates` | Directory for save states |
| Screenshots | `[Folders]` | `Screenshots` | Directory for screenshots |
| Game List | `[GameList]` | `RecursivePaths` | ROM directories |

### Example Configuration

```ini
[Folders]
Bios = /home/user/Emulation/bios/ps2
MemoryCards = /home/user/Emulation/saves/ps2
Savestates = /home/user/Emulation/states/pcsx2
Screenshots = /home/user/Emulation/screenshots/ps2

[GameList]
RecursivePaths = /home/user/Emulation/roms/ps2
```

### Directory Structure

```
PCSX2/
├── cheats/
├── gamesettings/
├── inis/
│   └── PCSX2.ini
├── inputprofiles/
├── logs/
├── memcards/
├── patches/
├── sstates/
└── videos/
```

### Provisions Required

- **BIOS**: `scph39001.bin` (USA), `scph70004.bin` (Europe), `scph10000.bin` (Japan)

### Controller Configuration

- Input profiles in `inputprofiles/` directory
- Per-game input profiles assignable via game properties
- Section: `[Pad]` with controller mappings

### Portable Mode

Create `portable.txt` or `portable.ini` in the PCSX2 directory, or use `-portable` launch argument.

### Manual Configuration Required

None - all paths configurable via text files.

### Sources

- [PCSX2 Configuration Guide](https://github.com/PCSX2/pcsx2/blob/master/pcsx2/Docs/Configuration_Guide/Configuration_Guide.md)
- [PCSX2 Official Documentation](https://pcsx2.net/docs/configuration/general/)
- [EmuDeck PCSX2 Documentation](https://emudeck.github.io/emulators/steamos/pcsx2/)

---

## RPCS3 (PS3)

### Config Files

| Property | Value |
|----------|-------|
| Format | YAML |
| Main Config | `~/.config/rpcs3/config.yml` |
| VFS Config | `~/.config/rpcs3/vfs.yml` |
| Base Dir | `$XDG_CONFIG_HOME/rpcs3` or `~/.config/rpcs3` |

### Path Configuration (vfs.yml)

RPCS3 uses a Virtual File System (VFS) that maps PS3 paths to host paths.

| VFS Path | Default Host Path | Description |
|----------|------------------|-------------|
| `$(EmulatorDir)` | `` (config directory) | Base variable for all paths |
| `/dev_hdd0/` | `$(EmulatorDir)dev_hdd0/` | Main PS3 storage (games, saves) |
| `/dev_hdd1/` | `$(EmulatorDir)dev_hdd1/` | Secondary storage |
| `/dev_flash/` | `$(EmulatorDir)dev_flash/` | Firmware location |
| `/dev_bdvd/` | `$(EmulatorDir)dev_bdvd/` | Disc mount point |
| `/games/` | `$(EmulatorDir)games/` | Games directory |

### Example vfs.yml

```yaml
$(EmulatorDir): /home/user/Emulation/opaque/rpcs3/
/dev_hdd0/: $(EmulatorDir)dev_hdd0/
/dev_hdd1/: $(EmulatorDir)dev_hdd1/
/dev_flash/: $(EmulatorDir)dev_flash/
/dev_flash2/: $(EmulatorDir)dev_flash2/
/dev_flash3/: $(EmulatorDir)dev_flash3/
/dev_bdvd/: $(EmulatorDir)dev_bdvd/
/games/: $(EmulatorDir)games/
```

### Directory Structure

```
rpcs3/
├── config.yml
├── vfs.yml
├── dev_bdvd/
├── dev_flash/
├── dev_hdd0/
│   ├── game/          # Installed games
│   ├── home/          # Save data
│   └── savedata/      # Additional saves
├── dev_hdd1/
├── dev_usb000/
└── games/
```

### Opaque Directory Pattern

RPCS3 is best managed as an opaque directory. The internal structure (dev_hdd0, dev_flash, etc.) should be kept together and synced as a unit. Individual path configuration is limited by RPCS3's architecture.

### Provisions Required

- **Firmware**: Installed through emulator GUI from official PS3 firmware file (PS3UPDAT.PUP)
- Firmware installs to `dev_flash/`

### Controller Configuration

- Config in `config.yml` under `Input` section
- Supports various input handlers: DualShock 3, DualShock 4, DualSense, XInput, evdev

### Manual Configuration Required

- **Firmware installation**: Must be done through the RPCS3 GUI (File > Install Firmware)
- VFS path modifications via config file may be overwritten by the UI

### Limitations

- The UI does not allow setting relative paths for VFS
- Some manual vfs.yml edits may be "fixed" by explicit code
- dev_flash must typically be in the same location as the executable for firmware recognition

### Sources

- [RPCS3 GitHub](https://github.com/RPCS3/rpcs3)
- [RPCS3 VFS Forum Discussion](https://forums.rpcs3.net/archive/index.php/thread-201512.html)
- [EmuDeck RPCS3 Documentation](https://emudeck.github.io/emulators/steamos/rpcs3/)

---

## PPSSPP (PSP)

### Config File

| Property | Value |
|----------|-------|
| Format | INI |
| Location | `~/.config/ppsspp/PSP/SYSTEM/ppsspp.ini` |
| Base Dir | `$XDG_CONFIG_HOME/ppsspp` or `~/.config/ppsspp` |

### Path Configuration

| Setting | INI Section | INI Key | Description |
|---------|-------------|---------|-------------|
| MemStick Directory | `[General]` | `MemStickDirectory` | Root directory for PSP virtual memory stick |
| Screenshots | `[General]` | `ScreenshotsPath` | Directory for screenshots |

### Memstick Directory Structure

PPSSPP uses a "memstick" directory that mimics the PSP's internal structure:

```
memstick/
├── PSP/
│   ├── CHEATS/
│   ├── GAME/           # Homebrew and game data
│   ├── PPSSPP_STATE/   # Save states
│   ├── SAVEDATA/       # Game saves
│   ├── SYSTEM/         # Config files
│   │   ├── ppsspp.ini
│   │   └── controls.ini
│   └── TEXTURES/       # Texture replacements
└── ...
```

### Example Configuration

```ini
[General]
MemStickDirectory = /home/user/Emulation/opaque/ppsspp
ScreenshotsPath = /home/user/Emulation/screenshots/psp
```

### Opaque Directory Pattern

PPSSPP is best managed as an opaque directory since:
- Saves are stored relative to memstick at `PSP/SAVEDATA/`
- Save states are at `PSP/PPSSPP_STATE/`
- The internal structure mimics original PSP hardware

### Provisions Required

None - PPSSPP uses High-Level Emulation (HLE), no BIOS required.

### Controller Configuration

- Controller config in `PSP/SYSTEM/controls.ini`
- Main config has input settings in `[Control]` section

### Hidden Settings

Some settings can only be set via ppsspp.ini and not through the GUI, such as:
- `[General] TopMost = True` - keeps PPSSPP window in front

### Manual Configuration Required

None - all paths configurable via text files.

### Sources

- [PPSSPP Official FAQ](https://www.ppsspp.org/faq.html)
- [PPSSPP Hidden Settings](https://www.ppsspp.org/docs/settings/hidden/)
- [EmuDeck PPSSPP Documentation](https://emudeck.github.io/emulators/steamos/ppsspp/)

---

## Vita3K (PS Vita)

### Config File

| Property | Value |
|----------|-------|
| Format | YAML |
| Location | `<Vita3K>/config.yml` or `~/.config/Vita3K/config.yml` |
| Base Dir | `~/.local/share/Vita3K/Vita3K` (Linux) |

### Path Configuration

| Setting | YAML Key | Description |
|---------|----------|-------------|
| Data Directory | `pref-path` | Root directory for all Vita3K data |

### Example Configuration

```yaml
pref-path: /home/user/Emulation/opaque/vita3k
log-level: 1
pstv-mode: false
```

### Directory Structure (within pref-path)

```
vita3k/
├── ux0/
│   ├── app/          # Installed games
│   ├── data/         # Game data
│   └── user/         # Save data
├── ur0/
├── uma0/
└── ...
```

### Opaque Directory Pattern

Vita3K must be managed as an opaque directory. The pref-path setting redirects the entire data directory, and Vita3K manages the internal structure (ux0, ur0, etc.) which mirrors the PS Vita's filesystem.

### Provisions Required

- **Firmware**: Downloaded and installed through the emulator GUI during first-time setup

### Controller Configuration

- Settings in config.yml
- Basic controller support configurable via GUI

### Manual Configuration Required

- **Firmware installation**: Must be done through the Vita3K GUI on first run

### Sources

- [Vita3K Quickstart Guide](https://vita3k.org/quickstart.html)
- [Vita3K Configuration Wiki](https://github.com/Vita3K/Vita3K/wiki/Configuration)
- [EmuDeck Vita3K Documentation](https://emudeck.github.io/emulators/steamos/vita3k/)

---

## mGBA (GB/GBC/GBA)

### Config File

| Property | Value |
|----------|-------|
| Format | INI |
| Location | `~/.config/mgba/config.ini` |
| Base Dir | `$XDG_CONFIG_HOME/mgba` or `~/.config/mgba` |

### Path Configuration

| Setting | INI Section | INI Key | Description |
|---------|-------------|---------|-------------|
| Save Games | `[ports.qt]` | `savegamePath` | Directory for battery saves (.sav) |
| Save States | `[ports.qt]` | `savestatePath` | Directory for save states |
| Screenshots | `[ports.qt]` | `screenshotPath` | Directory for screenshots |
| Patches | `[ports.qt]` | `patchPath` | Directory for patches |

### Example Configuration

```ini
[ports.qt]
savegamePath=/home/user/Emulation/saves/gba
savestatePath=/home/user/Emulation/states/mgba
screenshotPath=/home/user/Emulation/screenshots/gba
patchPath=/home/user/Emulation/patches/gba
```

### Provisions Required

- **BIOS**: `gba_bios.bin` (optional - mGBA has HLE)
- BIOS improves compatibility but is not required

### Controller Configuration

- Controller mappings in config.ini under relevant sections
- Supports keyboard and gamepad input

### Portable Mode

Click "Portable" option in the emulator to keep configuration alongside the executable instead of in `~/.config/mgba`.

### Manual Configuration Required

None - all paths configurable via text files.

### Notes

- Default screenshot behavior saves to ROM directory if no path configured
- Screenshots saved as PNG

### Sources

- [mGBA FAQ](https://mgba.io/faq.html)
- [EmuDeck mGBA Documentation](https://emudeck.github.io/emulators/steamos/mgba/)

---

## melonDS (NDS)

### Config File

| Property | Value |
|----------|-------|
| Format | INI |
| Location | `~/.config/melonDS/melonDS.ini` |
| Base Dir | `$XDG_CONFIG_HOME/melonDS` or `~/.config/melonDS` |

### Path Configuration

| Setting | INI Key | Description |
|---------|---------|-------------|
| ARM9 BIOS | `BIOS9Path` | Full path to bios9.bin |
| ARM7 BIOS | `BIOS7Path` | Full path to bios7.bin |
| Firmware | `FirmwarePath` | Full path to firmware.bin |
| Save Files | `SaveFilePath` | Directory for save files |
| Save States | `SavestatePath` | Directory for save states |
| Screenshots | `ScreenshotPath` | Directory for screenshots |
| Last ROM Folder | `LastROMFolder` | Default ROM browser location |

### Example Configuration

```ini
BIOS9Path=/home/user/Emulation/bios/nds/bios9.bin
BIOS7Path=/home/user/Emulation/bios/nds/bios7.bin
FirmwarePath=/home/user/Emulation/bios/nds/firmware.bin
SaveFilePath=/home/user/Emulation/saves/nds
SavestatePath=/home/user/Emulation/states/melonds
ScreenshotPath=/home/user/Emulation/screenshots/nds
LastROMFolder=/home/user/Emulation/roms/nds
```

### DSi Configuration (Additional Keys)

```ini
DSiBIOS9Path=/path/to/biosdsi9.bin
DSiBIOS7Path=/path/to/biosdsi7.bin
DSiFirmwarePath=/path/to/dsifirmware.bin
DSiNANDPath=/path/to/dsinand.bin
```

### Provisions Required

- **ARM7 BIOS**: `bios7.bin` (optional with HLE)
- **ARM9 BIOS**: `bios9.bin` (optional with HLE)
- **Firmware**: `firmware.bin` (optional with HLE)

Note: melonDS has an open-source BIOS that is enabled by default, making real BIOS files optional.

### Controller Configuration

Settings in melonDS.ini for input mappings.

### Manual Configuration Required

- To use console BIOS, must enable "Use external BIOS/firmware files" in the melonDS GUI

### Important Notes

- Do not close melonDS with the X button while changing settings; use File > Quit to save config
- BIOS file paths are full paths, not directory paths

### Sources

- [melonDS FAQ](https://melonds.kuribo64.net/faq.php)
- [melonDS GitHub](https://github.com/melonDS-emu/melonDS)
- [EmuDeck melonDS Documentation](https://emudeck.github.io/emulators/steamos/melonds/)

---

## Azahar (3DS)

### Config File

| Property | Value |
|----------|-------|
| Format | INI (Qt INI format) |
| Location | `~/.config/azahar/qt-config.ini` |
| Base Dir | `$XDG_CONFIG_HOME/azahar` or `~/.config/azahar` |
| Data Dir | `~/.local/share/azahar-emu` |

### Path Configuration

| Setting | INI Section | INI Key | Description |
|---------|-------------|---------|-------------|
| NAND Directory | `[Data%20Storage]` | `nand_directory` | Emulated NAND storage |
| NAND Default | `[Data%20Storage]` | `nand_directory\default` | Set to `false` to use custom path |
| SDMC Directory | `[Data%20Storage]` | `sdmc_directory` | Emulated SD card |
| SDMC Default | `[Data%20Storage]` | `sdmc_directory\default` | Set to `false` to use custom path |
| Screenshots | `[UI]` | `Screenshots\screenshot_path` | Screenshot directory |
| Game Directories | `[UI]` | `Paths\gamedirs\*` | ROM directory configuration |

### Example Configuration

```ini
[Data%20Storage]
nand_directory=/home/user/Emulation/opaque/azahar/nand
nand_directory\default=false
sdmc_directory=/home/user/Emulation/opaque/azahar/sdmc
sdmc_directory\default=false

[UI]
Screenshots\screenshot_path=/home/user/Emulation/screenshots/3ds
Paths\gamedirs\size=1
Paths\gamedirs\1\deep_scan=false
Paths\gamedirs\1\expanded=true
Paths\gamedirs\1\path=/home/user/Emulation/roms/3ds
```

### Directory Structure

```
azahar/
├── cheats/
├── nand/
│   └── data/
│       └── 00000000000000000000000000000000/
│           ├── extdata/
│           └── sysdata/
├── sdmc/
│   └── Nintendo 3DS/
│       └── 00000000000000000000000000000000/
│           └── 00000000000000000000000000000000/
│               ├── title/     # Save data
│               └── extdata/
├── screenshots/
└── textures/
```

### Opaque Directory Pattern

Azahar is best managed with an opaque directory pattern for the nand and sdmc directories. Save data is stored within the sdmc directory structure, following the 3DS's organization.

### Provisions Required

- **AES Keys**: May be required for certain encrypted content
- Keys placed in `sysdata/aes_keys.txt`

### Controller Configuration

Settings in qt-config.ini under input sections.

### Portable Mode

Create a `user` directory alongside the executable before first launch.

### Manual Configuration Required

- Close Azahar completely before editing qt-config.ini (changes will be overwritten otherwise)

### Sources

- [Azahar GitHub](https://github.com/azahar-emu/azahar)
- [Citra User Directory Documentation](https://citra-emulator.com/wiki/user-directory/)
- [EmuDeck Azahar Documentation](https://manual.emudeck.com/tricks/azahar/)

---

## Dolphin (GameCube/Wii)

### Config Files

| Property | Value |
|----------|-------|
| Format | INI |
| Main Config | `~/.local/share/dolphin-emu/Dolphin.ini` |
| Graphics | `~/.local/share/dolphin-emu/GFX.ini` |
| GC Controller | `~/.local/share/dolphin-emu/GCPadNew.ini` |
| Wii Controller | `~/.local/share/dolphin-emu/WiimoteNew.ini` |
| Hotkeys | `~/.local/share/dolphin-emu/Hotkeys.ini` |
| Base Dir | `$XDG_DATA_HOME/dolphin-emu` or `~/.local/share/dolphin-emu` |

### Path Configuration (Dolphin.ini)

| Setting | INI Section | INI Key | Description |
|---------|-------------|---------|-------------|
| ISO Paths | `[General]` | `ISOPath0`, `ISOPath1`, etc. | ROM directories |
| ISO Path Count | `[General]` | `ISOPaths` | Number of ISO paths |
| Skip BIOS | `[Core]` | `SkipIPL` | Skip GameCube BIOS animation |

### Directory Structure

```
dolphin-emu/
├── Config/
│   ├── Dolphin.ini
│   ├── GFX.ini
│   ├── GCPadNew.ini
│   └── WiimoteNew.ini
├── GC/
│   ├── USA/           # GameCube memory cards and BIOS by region
│   ├── EUR/
│   └── JAP/
├── Wii/
│   ├── title/         # Wii save data
│   └── sd.raw         # Virtual SD card
├── StateSaves/        # Save states
├── ScreenShots/       # Screenshots
├── GameSettings/      # Per-game settings
└── Profiles/
    ├── GCPad/
    └── Wiimote/
```

### Example Configuration

```ini
[General]
ISOPaths = 2
ISOPath0 = /home/user/Emulation/roms/gamecube
ISOPath1 = /home/user/Emulation/roms/wii

[Core]
SkipIPL = True
```

### Save Data Locations

- **GameCube Memory Cards**: `GC/<region>/` as `.raw` or `.gci` files
- **Wii Saves**: `Wii/title/` directory structure
- **Save States**: `StateSaves/` directory

### Provisions Required

- **GameCube BIOS** (optional): `GC/<region>/IPL.bin` for boot animation
- **Wii System Menu** (optional): Installable through Dolphin

### Controller Configuration

- `GCPadNew.ini` for GameCube controller mappings
- `WiimoteNew.ini` for Wii Remote mappings
- Supports profiles in `Profiles/GCPad/` and `Profiles/Wiimote/`

### Virtual SD Card

- Located at `Wii/sd.raw`
- Configurable in Options > Configuration > Wii

### Manual Configuration Required

- Memory card path changes require editing through the GUI (Options > Configuration > GameCube)
- Some paths not easily configurable via INI alone

### Limitations

- Memory card directory not directly configurable via INI (uses slot-based system)
- Wii NAND location tied to data directory structure

### Sources

- [Dolphin Emulator FAQ](https://dolphin-emu.org/docs/faq/)
- [Dolphin Virtual SD Card Guide](https://dolphin-emu.org/docs/guides/virtual-sd-card-guide/)
- [EmuDeck Dolphin Documentation](https://emudeck.github.io/emulators/steamos/dolphin/)

---

## Cemu (Wii U)

### Config File

| Property | Value |
|----------|-------|
| Format | XML |
| Location | `~/.config/Cemu/settings.xml` |
| Base Dir | `$XDG_CONFIG_HOME/Cemu` or `~/.config/Cemu` |

### Path Configuration

| Setting | XML Path | Description |
|---------|----------|-------------|
| MLC Path | `<content><mlc_path>` | Wii U internal storage emulation |
| Game Paths | `<content><GamePaths><Entry>` | ROM directories |

### Example Configuration

```xml
<?xml version="1.0" encoding="utf-8"?>
<content>
    <mlc_path>/home/user/Emulation/opaque/cemu</mlc_path>
    <GamePaths>
        <Entry>/home/user/Emulation/roms/wiiu</Entry>
    </GamePaths>
</content>
```

### MLC Directory Structure

The mlc01 directory emulates the Wii U's internal storage:

```
mlc01/
├── sys/
│   └── title/        # System titles
├── usr/
│   ├── boss/         # SpotPass data
│   ├── save/         # Save data
│   └── title/        # Installed games, updates, DLC
└── ...
```

### Opaque Directory Pattern

Cemu is best managed with an opaque directory for the mlc_path. This directory contains:
- Save data
- Installed updates and DLC
- System files

### Provisions Required

- **Wii U Keys**: Required for decryption
- Keys file location depends on Cemu setup

### Controller Configuration

- Controller profiles in `controllerProfiles/` directory
- Configurable through GUI: Options > Input settings

### Manual Configuration Required

- Controller configuration typically done through GUI
- Flatpak users may need `flatpak override` for custom MLC paths

### Limitations

- Editing settings.xml directly may not work well (Cemu overwrites symlinks)
- Relative game paths may cause issues
- Some settings best changed through GUI

### Sources

- [Cemu Installation Guide](https://cemu.cfw.guide/installing-cemu)
- [Cemu Wiki - Folder Structure](https://wiki.cemu.info/wiki/Folder_structure)
- [Cemu FAQ](https://cemu.cfw.guide/faq.html)

---

## Eden (Switch)

### Config File

| Property | Value |
|----------|-------|
| Format | INI (Qt INI format) |
| Location | `~/.config/eden/qt-config.ini` |
| Base Dir | `$XDG_CONFIG_HOME/eden` or `~/.config/eden` |

### Path Configuration

| Setting | INI Section | INI Key | Description |
|---------|-------------|---------|-------------|
| NAND Directory | `[Data%20Storage]` | `nand_directory` | Emulated Switch NAND |
| NAND Default | `[Data%20Storage]` | `nand_directory\default` | Set to `false` for custom path |
| SDMC Directory | `[Data%20Storage]` | `sdmc_directory` | Emulated SD card |
| SDMC Default | `[Data%20Storage]` | `sdmc_directory\default` | Set to `false` for custom path |
| Screenshots | `[UI]` | `Screenshots\screenshot_path` | Screenshot directory |
| Game Directories | `[UI]` | `Paths\gamedirs\*` | ROM directory configuration |

### Example Configuration

```ini
[Data%20Storage]
nand_directory=/home/user/Emulation/opaque/eden/nand
nand_directory\default=false
sdmc_directory=/home/user/Emulation/opaque/eden/sdmc
sdmc_directory\default=false

[UI]
Screenshots\screenshot_path=/home/user/Emulation/screenshots/switch
Paths\gamedirs\size=1
Paths\gamedirs\1\deep_scan=false
Paths\gamedirs\1\expanded=true
Paths\gamedirs\1\path=/home/user/Emulation/roms/switch
```

### Directory Structure

```
eden/
├── nand/
│   └── user/
│       └── save/         # Save data
├── sdmc/
├── keys/
│   └── prod.keys         # Decryption keys
├── shader/               # Shader cache
└── ...
```

### Opaque Directory Pattern

Eden is strongly suited for the opaque directory pattern. The NAND contains:
- Save data
- User profiles
- Installed updates/DLC

### Provisions Required

- **prod.keys**: Required for game decryption
- **Firmware**: Optional, installed through Eden GUI

### Controller Configuration

Settings in qt-config.ini under input sections.

### Manual Configuration Required

- **Keys**: Must be placed in the keys directory
- **Firmware installation**: Done through Eden GUI if needed

### Important Notes

- If NAND is in a non-standard location, ensure configuration is correct to avoid "orphaned profiles" issues
- Save data location depends on NAND path

### Sources

- [Eden Emulator Website](https://eden-emu.dev/)
- [Eden GitHub Releases](https://github.com/eden-emulator/Releases)
- [Eden Installation Guide](https://wiki.axekin.com/2025/06/02/Eden/)

---

## RetroArch (SNES, NES, Genesis, Saturn, N64)

### Config File

| Property | Value |
|----------|-------|
| Format | CFG (key = "value" format) |
| Location | `~/.config/retroarch/retroarch.cfg` |
| Base Dir | `$XDG_CONFIG_HOME/retroarch` or `~/.config/retroarch` |

### Main Path Configuration (retroarch.cfg)

| Setting | Config Key | Description |
|---------|------------|-------------|
| System/BIOS | `system_directory` | BIOS and system files |
| Save Files | `savefile_directory` | Battery saves (.srm) |
| Save States | `savestate_directory` | Save states (.state) |
| Screenshots | `screenshot_directory` | Screenshots |
| Cores | `libretro_directory` | Core libraries (.so) |
| Assets | `assets_directory` | Menu assets |
| Core Assets | `core_assets_directory` | Downloaded content |
| ROM Browser | `rgui_browser_directory` | Default file browser location |
| Playlists | `playlist_directory` | Playlist files |
| Thumbnails | `thumbnails_directory` | Game thumbnails |
| Joypad Autoconfig | `joypad_autoconfig_dir` | Controller autoconfig files |
| Input Remapping | `input_remapping_directory` | Remapped controls |

### Example Configuration

```cfg
system_directory = "/home/user/Emulation/bios"
savefile_directory = "/home/user/Emulation/saves"
savestate_directory = "/home/user/Emulation/states"
screenshot_directory = "/home/user/Emulation/screenshots"
libretro_directory = "/home/user/.local/share/kyaraben/retroarch/cores"
assets_directory = "/home/user/.local/share/kyaraben/retroarch/assets"
rgui_browser_directory = "/home/user/Emulation/roms"
```

### Per-Core Configuration

RetroArch supports per-core overrides in `~/.config/retroarch/config/<core_name>/<core_name>.cfg`:

```cfg
savefile_directory = "/home/user/Emulation/saves/snes"
savestate_directory = "/home/user/Emulation/states/retroarch-bsnes"
screenshot_directory = "/home/user/Emulation/screenshots/snes"
rgui_browser_directory = "/home/user/Emulation/roms/snes"
```

### Supported Cores for kyaraben

| System | Core | Notes |
|--------|------|-------|
| SNES | `bsnes_libretro` | High accuracy |
| NES | `mesen_libretro` | High accuracy |
| Genesis | `genesis_plus_gx_libretro` | Sega Genesis/Mega Drive |
| Saturn | `beetle_saturn_libretro` | Sega Saturn |
| N64 | `mupen64plus_next_libretro` | Nintendo 64 |

### Provisions Required

Different cores have different BIOS requirements:

| System | BIOS Files | Required |
|--------|------------|----------|
| SNES | None | - |
| NES | None | - |
| Genesis | `bios_CD_U.bin`, `bios_CD_E.bin`, `bios_CD_J.bin` | Only for Sega CD |
| Saturn | `sega_101.bin`, `mpr-17933.bin` | Yes |
| N64 | None | - |

### Controller Configuration

RetroArch uses an autoconfig system:
- Autoconfig files in `joypad_autoconfig_dir`
- Per-controller profiles matched by Vendor ID, Product ID, and Device Name
- Input remapping saved to `input_remapping_directory`

### Important Configuration Options

```cfg
# Save configuration
config_save_on_exit = "true"
sort_savefiles_by_content_enable = "true"
sort_savestates_by_content_enable = "true"

# Per-core saves (recommended for kyaraben)
sort_savefiles_enable = "false"
sort_savestates_enable = "false"
savefiles_in_content_dir = "false"
savestates_in_content_dir = "false"
```

### Manual Configuration Required

None - all paths fully configurable via text files.

### Sources

- [RetroArch GitHub - retroarch.cfg](https://github.com/libretro/RetroArch/blob/master/retroarch.cfg)
- [Libretro Docs - Directory Configuration](https://docs.libretro.com/guides/change-directories/)
- [Libretro Docs - Controller Autoconfiguration](https://docs.libretro.com/guides/controller-autoconfiguration/)

---

## Flycast (Dreamcast)

### Config File

| Property | Value |
|----------|-------|
| Format | CFG (INI-like) |
| Location | `~/.config/flycast/emu.cfg` |
| Base Dir | `$XDG_CONFIG_HOME/flycast` or `~/.config/flycast` |

### Path Configuration

| Setting | CFG Section | CFG Key | Description |
|---------|-------------|---------|-------------|
| Content Path | `[config]` | `Dreamcast.ContentPath` | ROM directories (semicolon-separated) |
| Save States | `[config]` | `SavestatesPath` | Save state directory |
| Screenshots | `[config]` | `ScreenshotsPath` | Screenshot directory |

### Example Configuration

```cfg
[config]
Dreamcast.ContentPath = /home/user/Emulation/roms/dreamcast
SavestatesPath = /home/user/Emulation/states/flycast
ScreenshotsPath = /home/user/Emulation/screenshots/dreamcast
```

### Directory Structure

```
flycast/
├── emu.cfg
├── data/
│   ├── dc_boot.bin      # BIOS
│   ├── dc_flash.bin     # Flash memory
│   └── ...
├── mappings/            # Controller mappings
└── ...
```

### Provisions Required

- **dc_boot.bin**: Dreamcast BIOS (optional - Flycast has HLE)
- **dc_flash.bin**: Dreamcast flash memory (optional)

### Controller Configuration

- Mappings stored in `mappings/` directory
- Configurable through GUI and command line

### Command Line Configuration

Settings can be passed via command line: `-config section:key=value`

### Manual Configuration Required

None - all paths configurable via text files.

### Sources

- [Flycast GitHub](https://github.com/flyinghead/flycast)
- [Flycast Wiki - Configuration](https://github.com/TheArcadeStriker/flycast-wiki/wiki/Configuration-files-and-command-line-parameters)
- [EmuDeck Flycast Documentation](https://emudeck.github.io/emulators/steamos/flycast/)

---

## Controller Configuration

### Steam Input Considerations

When running emulators through Steam (especially on Steam Deck), Steam Input can intercept controller signals. Key considerations:

1. **Steam Input Translation**: Steam may translate controller input to appear as an Xbox 360 controller
2. **Disabling Steam Input**: May be necessary for proper emulator controller detection
3. **Per-Application Settings**: Configure in Steam > Properties > Controller

### EmuDeck Controller Integration

EmuDeck provides standardized controller configurations across emulators with consistent hotkeys:

| Hotkey | Action |
|--------|--------|
| Select + Start | Exit emulator |
| Select + L1 | Load state |
| Select + R1 | Save state |
| Select + L2 | Rewind |
| Select + R2 | Fast forward |

### SDL Controller Database

Many emulators use the SDL GameControllerDB for controller mapping:
- Repository: https://github.com/mdqinc/SDL_GameControllerDB
- Custom mappings can be added to `gamecontrollerdb.txt` in the emulator's user directory

### Per-Emulator Controller Notes

| Emulator | Controller Config Location | Notes |
|----------|---------------------------|-------|
| DuckStation | `inputprofiles/` | SDL/XInput/DInput backends |
| PCSX2 | `inputprofiles/` | Per-game profiles supported |
| RetroArch | `joypad_autoconfig_dir` | Extensive autoconfig database |
| Dolphin | `GCPadNew.ini`, `WiimoteNew.ini` | Profiles supported |
| mGBA | `config.ini` | Basic mapping |

---

## Summary Matrix

### Path Configurability

| Emulator | ROMs | BIOS | Saves | States | Screenshots | Config Format |
|----------|------|------|-------|--------|-------------|---------------|
| DuckStation | ✅ | ✅ | ✅ | ✅ | ✅ | INI |
| PCSX2 | ✅ | ✅ | ✅ | ✅ | ✅ | INI |
| RPCS3 | ✅ | ⚠️ | ⚠️ | ⚠️ | ⚠️ | YAML (opaque) |
| PPSSPP | ✅ | N/A | ⚠️ | ⚠️ | ✅ | INI (opaque) |
| Vita3K | ⚠️ | ⚠️ | ⚠️ | N/A | ⚠️ | YAML (opaque) |
| mGBA | ⚠️ | ⚠️ | ✅ | ✅ | ✅ | INI |
| melonDS | ✅ | ✅ | ✅ | ✅ | ✅ | INI |
| Azahar | ✅ | ⚠️ | ⚠️ | ⚠️ | ✅ | INI (opaque) |
| Dolphin | ✅ | ⚠️ | ⚠️ | ✅ | ✅ | INI |
| Cemu | ✅ | ⚠️ | ⚠️ | ⚠️ | ⚠️ | XML (opaque) |
| Eden | ✅ | ⚠️ | ⚠️ | ⚠️ | ✅ | INI (opaque) |
| RetroArch | ✅ | ✅ | ✅ | ✅ | ✅ | CFG |
| Flycast | ✅ | ⚠️ | ⚠️ | ✅ | ✅ | CFG |

**Legend**:
- ✅ = Fully configurable via text files
- ⚠️ = Partially configurable or requires opaque directory pattern
- N/A = Not applicable

### Opaque Directory Recommendations

Emulators recommended for the opaque directory pattern (managed as a single unit):

| Emulator | Reason |
|----------|--------|
| RPCS3 | Complex VFS structure, firmware requirements |
| PPSSPP | PSP memstick directory structure |
| Vita3K | PS Vita filesystem emulation |
| Azahar | 3DS NAND/SDMC structure |
| Cemu | Wii U MLC storage structure |
| Eden | Switch NAND/SDMC structure |

### Manual Setup Requirements

| Emulator | Manual Steps Required |
|----------|----------------------|
| RPCS3 | Firmware installation via GUI |
| Vita3K | Firmware installation via GUI |
| Cemu | Initial setup wizard |
| Eden | Keys placement, optional firmware |
| Azahar | Optional AES keys |

---

## Implementation Notes for kyaraben

### Config Generation Strategy

1. **Simple Path Emulators** (DuckStation, PCSX2, mGBA, melonDS, RetroArch, Flycast):
   - Generate config patches with explicit paths to UserStore directories
   - Use system-specific subdirectories for saves/screenshots
   - Use emulator-specific subdirectories for states

2. **Opaque Directory Emulators** (RPCS3, PPSSPP, Vita3K, Azahar, Cemu, Eden):
   - Configure the emulator's root data path to `UserStore/opaque/<emulator>/`
   - Let the emulator manage internal structure
   - Sync entire opaque directory

### Config File Preservation

When writing config files:
- Preserve user customizations where possible
- Only modify kyaraben-managed paths
- Consider three-way merge for user settings

### Portable vs Installed Mode

Most emulators support portable mode via:
- `portable.txt` file (DuckStation, PCSX2)
- `portable.ini` file (PCSX2)
- `user/` directory (Azahar, Eden)

For kyaraben's AppImage-based installation, portable mode is generally not used since configs are managed separately.

### Sync Considerations

For the Synchronizer:
- Ignore shader caches and regenerable data
- Consider per-emulator ignore patterns for large caches
- Opaque directories may include unwanted data (shader cache)

Suggested ignore patterns:
```
**/shader_cache/**
**/cache/**
**/shaders/**
**/*.tmp
**/logs/**
```
