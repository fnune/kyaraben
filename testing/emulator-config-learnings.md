# Emulator configuration learnings from EmuDeck and RetroDECK

This document compares how kyaraben, EmuDeck, and RetroDECK configure emulators and identifies potential improvements for kyaraben.

## Overview of approaches

| Aspect | Kyaraben | EmuDeck | RetroDECK |
|--------|----------|---------|-----------|
| Language | Go | Bash | Bash |
| Distribution | Native | AppImage/Flatpak/Binary | Flatpak |
| Config model | Interface-based generators | Rsync + variable substitution | Component-based + presets |
| Emulator count | 17 | 31+ | 13+ |

## Potential improvements for kyaraben

### 1. Cross-emulator presets

RetroDECK implements a preset system that applies settings across multiple emulators at once:
- `widescreen` - toggle widescreen for all supporting emulators
- `rewind` - enable rewind in RetroArch and compatible emulators
- `quick_resume` - auto-save/load state on exit/launch
- `discord_rich_presence` - enable Discord integration everywhere
- `borders` - enable bezels/borders across emulators

Kyaraben could benefit from a similar system. Instead of configuring each emulator individually, users could toggle high-level features that cascade to all compatible emulators.

**Example features that could be presets:**
- Widescreen mode
- Integer scaling
- Discord presence
- RetroAchievements
- Auto-save on exit

### 2. RetroAchievements integration

Both EmuDeck and RetroDECK have first-class RetroAchievements support:
- Global credential storage (username + token)
- Per-emulator login functions
- Hardcore mode toggle
- Support in: DuckStation, PCSX2, PPSSPP, RetroArch cores

Kyaraben doesn't appear to configure RetroAchievements. This is a significant feature for retro gaming enthusiasts.

**Emulators that support RetroAchievements:**
- DuckStation (PSX)
- PCSX2 (PS2)
- PPSSPP (PSP)
- RetroArch (many cores)
- Dolphin (experimental)

### 3. Default performance/graphics settings

EmuDeck ships with extensive default configurations that optimize each emulator:

**DuckStation:**
- Vulkan renderer
- 3x resolution scale (1080p) or 2x (720p)
- Memory card type: per-game title
- CPU recompiler mode
- Fast boot enabled

**PCSX2:**
- Vulkan renderer
- Setup wizard bypass
- Widescreen patches enabled
- EE/VU recompilers enabled

**Dolphin:**
- Vulkan backend
- 3x internal resolution
- Widescreen hack per-game
- Audio stretching enabled

Kyaraben currently focuses on paths but doesn't set performance defaults. Users have to manually configure each emulator's renderer, resolution, and optimizations.

### 4. Controller/hotkey configuration

EmuDeck pre-configures controller mappings and hotkeys for Steam Deck:
- Save state: Select + R1
- Load state: Select + L1
- Fast forward: Select + R2
- Screenshot: Select + L2
- Exit: Select + Start

Kyaraben doesn't manage controller or hotkey configuration. This is essential for a "set up once and play" experience, especially on handheld devices.

### 5. More ROM directory options

EmuDeck supports 200+ system directories with specific naming conventions:
- `psx`, `ps2`, `psp`, `ps3`, `ps4`
- `gc`, `wii`, `wiiu`, `switch`
- `n64`, `nds`, `3ds`
- `genesis`, `saturn`, `dreamcast`
- `arcade`, `mame`, `fbneo`

Kyaraben has a simpler structure. EmuDeck's comprehensive naming could help with ES-DE/Pegasus integration and multi-emulator setups.

### 6. Shader and overlay management

EmuDeck and RetroDECK both manage:
- CRT shaders for RetroArch
- Bezel overlays
- Per-core shader presets
- HD texture pack paths

This could enhance the visual experience in kyaraben, especially for users who want authentic CRT aesthetics.

### 7. Compression support

RetroDECK has a compression framework:
- CHD for disc-based games (PSX, PS2, Saturn, Dreamcast)
- RVZ for GameCube/Wii
- ZIP/7z for cartridge-based systems

Kyaraben could provide tools or guidance for ROM compression to save storage space.

### 8. Multi-user support

RetroDECK supports multiple user profiles with:
- Separate save directories per user
- Shared BIOS and ROM directories
- Per-user emulator configurations

This could be valuable for shared devices like a family Steam Deck.

## Specific emulator configurations to adopt

### DuckStation

EmuDeck configures these beyond paths:
```ini
[GPU]
Renderer = Vulkan
ResolutionScale = 3
Multisampling = Disabled
TextureFilter = Nearest

[Display]
AspectRatio = Auto
Alignment = Center
ShowOSDMessages = true

[BIOS]
PatchFastBoot = true

[MemoryCards]
Card1Type = PerGameTitle

[Cheevos]
Enabled = true
ChallengeMode = false
RichPresence = true
```

### PCSX2

EmuDeck configures:
```ini
[UI]
StartFullscreen = true
ConfirmShutdown = false
SetupWizardIncomplete = false

[Folders]
Bios = ${biosPath}/pcsx2
MemoryCards = ${savesPath}/ps2/pcsx2/memcards
Savestates = ${savesPath}/ps2/pcsx2/sstates
Screenshots = ${screenshotsPath}/pcsx2

[EmuCore/GS]
Renderer = 14  # Vulkan
upscale_multiplier = 3
enable_widescreen_patches = true

[Achievements]
Enabled = true
TestMode = false
HardcoreMode = false
RichPresence = true
```

### Dolphin

EmuDeck configures:
```ini
[General]
ISOPath0 = ${romsPath}/gc
ISOPath1 = ${romsPath}/wii
ISOPaths = 2
DumpPath = ${storagePath}/dolphin

[Core]
AudioStretch = True
OverclockEnable = False

[Display]
FullscreenDisplayRes = Auto
FullscreenResolution = Auto
Fullscreen = True
DisableScreenSaver = True
```

Plus `GFX.ini`:
```ini
[Settings]
AspectRatio = 0
wideScreenHack = True
InternalResolution = 3
ShowOSDMessages = True
```

### RetroArch

EmuDeck has 2600+ lines for RetroArch configuration:
- 50+ core downloads
- Per-core shader presets
- Per-core aspect ratio overrides
- Bezel/overlay configurations
- Controller autoconfig profiles

Key settings:
```cfg
video_driver = "vulkan"
audio_driver = "pulse"
menu_driver = "ozone"
input_joypad_driver = "sdl2"
savefile_directory = ${savesPath}/retroarch/saves
savestate_directory = ${savesPath}/retroarch/states
screenshot_directory = ${screenshotsPath}/retroarch
system_directory = ${biosPath}
video_shader_enable = "true"
```

## Architecture learnings

### 1. EmuDeck's declarative emulator metadata

```bash
declare -A CemuNative=(
    [emuName]="CemuNative"
    [emuType]="AppImage"
    [emuPath]="$emusFolder/Cemu.AppImage"
    [configDir]="${HOME}/.config/Cemu"
)
```

This pattern cleanly separates emulator metadata from configuration logic. Kyaraben's interface-based approach is similar but more type-safe.

### 2. RetroDECK's preset propagation

```bash
change_preset_config "widescreen" "dolphin" "true"
change_preset_config "widescreen" "duckstation" "true"
change_preset_config "widescreen" "pcsx2" "true"
```

The preset system maps high-level user intent to emulator-specific settings. This abstraction is valuable for user experience.

### 3. Format-agnostic configuration

RetroDECK's `set_setting_value()` handles multiple config formats:
```bash
set_setting_value "$config_file" "$setting" "$value" "$system_name" "$section"
```

This is similar to kyaraben's `ConfigWriter` but with format detection built into the function.

### 4. Version migration

Both projects track configuration versions and apply migrations when updating. This prevents breaking user setups when defaults change.

## Low-hanging fruit for kyaraben

1. **RetroAchievements support** - High user value, relatively simple to implement (add config entries for credentials)

2. **Widescreen preset** - Single toggle that enables widescreen in DuckStation, PCSX2, Dolphin, RetroArch

3. **Performance defaults** - Ship sensible defaults for renderer (Vulkan), resolution scale, recompilers

4. **Fast boot options** - Enable BIOS skip in DuckStation/PCSX2 for faster game loading

5. **Hotkey documentation** - Document recommended hotkey setups even if not auto-configured

## Questions to consider

1. Should kyaraben manage more than paths? Performance settings are valuable but risk breaking edge cases.

2. Is controller configuration in scope? This varies wildly by hardware (Steam Deck vs desktop vs arcade stick).

3. Should there be a "Steam Deck mode" with pre-tuned settings for that hardware?

4. How much should kyaraben overlap with frontend launchers like ES-DE that also configure emulators?
