# Kyaraben design document

> This is an early-stage design exploration. Feature sets, UI mockups, and config examples are illustrative, not final.

## Overview

An emulation manager that lets users declaratively configure their emulation setup through a simple UI. Powered by Nix under the hood, but users never see or interact with Nix directly.

## Core concept

Ship a binary that:

1. Contains a bundled Nix runtime (portable, no system install required)
2. Presents a UI where users select what they want
3. Fetches and configures emulators on demand from a binary cache
4. Generates all configs, folder structures, and integrations automatically

Users pick systems, the tool does the rest.

## Delivery methods

### 1. AppImage (primary)

The main product: a self-contained AppImage that bundles Nix internally. No dependencies, no installation, runs on any Linux distro.

```bash
chmod +x Kyaraben-x86_64.AppImage
./Kyaraben-x86_64.AppImage
```

Why AppImage:

- Single file, download and run
- No sandbox restrictions to work around
- Full filesystem access for emulation folders, configs, controllers
- Can implement its own auto-update mechanism
- Works on Steam Deck (desktop mode)
- No store review process

The binary size will be roughly 50-80 MB (Nix runtime + GUI framework + initial assets). First run will download emulator packages from the binary cache, which can be several hundred MB depending on selections. This is fine since users will be transferring gigabytes of ROMs anyway.

### 2. Home-manager module (power users)

For users already running NixOS or home-manager, we expose the same functionality as a Nix module. No need for the standalone binary.

```nix
{
  imports = [ inputs.kyaraben.homeManagerModules.default ];

  programs.kyaraben = {
    enable = true;
    emulationDir = "~/Emulation";

    systems = {
      nes.enable = true;
      snes.enable = true;
      psx.enable = true;
      gamecube.enable = true;
    };
  };
}
```

Then just `home-manager switch`. Same result as using the UI, but declarative and version-controlled.

The standalone binary and the home-manager module share the same core logic.

---

## Target platforms

Primary: Steam Deck (desktop mode), desktop Linux (any distro)

Linux handhelds (Anbernic, AYN, etc.) are an interesting target but present a conflict: these devices typically run custom OS images (ROCKNIX, NextUI, etc.) that bundle their own emulator setups. Since kyaraben's goal is to provide a complete emulator setup, it would compete with what these images already do. Supporting these devices properly would mean shipping a trimmed-down OS image with kyaraben as the emulation layer, which is a separate project.

## User experience

### First run

1. User downloads the AppImage
2. Runs it
3. Sees a welcome screen, picks a location for their emulation folder (default: `~/Emulation`)
4. Selects which systems they want to emulate
5. Clicks "Apply"
6. Tool fetches emulators, generates configs, creates folder structure
7. Done

### Ongoing use

- Open kyaraben to add/remove systems
- Update emulators with one click
- Optional: sync saves across devices

## UI screens

### 1. System picker

The main screen. Shows available systems grouped by manufacturer.

```
┌──────────────────────────────────────────────────────────────┐
│  What do you want to emulate?                                │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  Nintendo                        Sony                        │
│  ─────────                       ────                        │
│  ☑ NES                           ☐ PlayStation               │
│  ☑ SNES                          ☐ PlayStation 2             │
│  ☐ Nintendo 64                   ☐ PlayStation 3             │
│  ☐ GameCube                      ☐ PSP                       │
│  ☐ Wii                           ☐ PS Vita                   │
│  ☐ Switch                                                    │
│  ☐ Game Boy / Color              Sega                        │
│  ☐ Game Boy Advance              ────                        │
│  ☐ DS                            ☐ Master System / Genesis   │
│  ☐ 3DS                           ☐ Saturn                    │
│                                  ☐ Dreamcast                 │
│  Other                                                       │
│  ─────                           Arcade                      │
│  ☐ Atari (2600, 7800, etc.)      ──────                      │
│  ☐ Neo Geo                       ☐ MAME                      │
│  ☐ TurboGrafx-16                 ☐ FBNeo                     │
│                                                              │
├──────────────────────────────────────────────────────────────┤
│  Selected: 2 systems                         [ Apply ]       │
└──────────────────────────────────────────────────────────────┘
```

### 2. Settings

```
┌──────────────────────────────────────────────────────────────┐
│  Settings                                                    │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  Emulation folder         ~/Emulation               [Browse] │
│                                                              │
│  ─────────────────────────────────────────────────────────── │
│  Frontend                                                    │
│  ─────────────────────────────────────────────────────────── │
│  ☐ Install EmulationStation-DE                               │
│  ☐ Add frontend to Steam (for Steam Deck game mode)          │
│  ☐ Install Steam ROM Manager (add games directly to Steam)   │
│                                                              │
│  ─────────────────────────────────────────────────────────── │
│  Updates                                                     │
│  ─────────────────────────────────────────────────────────── │
│  [ Check for updates ]                                       │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

### 3. Progress view

```
┌──────────────────────────────────────────────────────────────┐
│  Applying changes...                                         │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ✓ Created folder structure                                  │
│  ✓ Fetched RetroArch                                         │
│  ● Fetching Dolphin...                       [=====>    ]    │
│  ○ Generating configs                                        │
│                                                              │
│                                              [ Cancel ]      │
└──────────────────────────────────────────────────────────────┘
```

## Folder structure created

```
~/Emulation/
├── roms/
│   ├── nes/
│   ├── snes/
│   ├── psx/
│   └── ...
├── bios/
│   ├── psx/
│   ├── ps2/
│   └── ...
├── saves/
│   └── (per-system)
├── states/
│   └── (per-system)
└── screenshots/
    └── (per-system)
```

## What happens under the hood

User perspective:

> "I checked PlayStation and GameCube, clicked Apply, now I have emulators configured"

What actually happens:

1. Kyaraben reads current selections
2. Generates a Nix expression representing desired state
3. Calls bundled Nix to build/fetch the closure from binary cache
4. Emulator binaries land in `~/.local/share/kyaraben/store/`
5. Generates emulator configs with correct paths
6. Creates wrapper scripts and .desktop files

User never sees Nix. It's an implementation detail.

## Emulator selection

Each system has a default emulator chosen for best out-of-the-box experience.

Philosophy: RetroArch cores when they're good enough. Standalones where RetroArch falls short.

| System                  | Default                      | Notes                                                  |
| ----------------------- | ---------------------------- | ------------------------------------------------------ |
| NES                     | RetroArch (Mesen)            |                                                        |
| SNES                    | RetroArch (bsnes)            |                                                        |
| Nintendo 64             | RetroArch (Mupen64Plus-Next) |                                                        |
| GameCube/Wii            | Dolphin                      | Standalone likely better than core, needs verification |
| Switch                  | Eden                         | Fetched from GitHub releases, user-initiated           |
| Game Boy / Color        | RetroArch (Gambatte)         |                                                        |
| GBA                     | RetroArch (mGBA)             |                                                        |
| DS                      | RetroArch (melonDS)          |                                                        |
| 3DS                     | Azahar                       |                                                        |
| PlayStation             | DuckStation                  |                                                        |
| PlayStation 2           | PCSX2                        | No viable RetroArch core                               |
| PlayStation 3           | RPCS3                        | No RetroArch core exists                               |
| PSP                     | RetroArch (PPSSPP)           |                                                        |
| PS Vita                 | Vita3K                       |                                                        |
| Master System / Genesis | RetroArch (Genesis Plus GX)  |                                                        |
| Saturn                  | RetroArch (Beetle Saturn)    |                                                        |
| Dreamcast               | RetroArch (Flycast)          |                                                        |
| Arcade                  | RetroArch (FBNeo)            |                                                        |

Notes on specific emulators:

- Switch emulators (Eden, etc.) are fetched from their GitHub releases as AppImages when the user explicitly enables Switch emulation
- 3DS: Citra is discontinued, Azahar is the active successor

## Design decisions

1. Emulator choice: default to RetroArch cores for unified experience. Standalones only where cores are significantly worse or unavailable.

1. Frontend: optional. EmulationStation-DE is offered as an integration, and can be added to Steam as a non-Steam game for Steam Deck game mode (like EmuDeck does). Off by default.

1. Controller config: we don't try to abstract this. Emulators have good auto-detection via SDL's gamecontrollerdb. Users configure controllers in RetroArch or standalone emulators directly. We just make sure hotkeys are sane (menu toggle, save state, etc.).

1. BIOS handling: the UI shows what's missing, what's found, and what's required vs optional. Verified by hash.

1. Settings abstraction: minimal for now. We manage paths and folder structure. Detailed emulator settings (resolution, shaders, etc.) are configured in the emulators themselves. This keeps the initial experience simple and avoids maintaining translation layers across emulator config formats.

---

## Configuration

### Philosophy

The UI is a friendly view into a text configuration. Everything the UI can do, the config file can do.

```
~/.config/kyaraben/config.toml
```

### Example config

```toml
[global]
emulation_dir = "~/Emulation"

[systems]
nes = true
snes = true
psx = true
gamecube = true
```

That's it for a basic setup. Advanced options can be added later as the project matures.

---

## BIOS doctor

A CLI command and UI panel that tells users exactly what they need.

```
$ kyaraben doctor

Checking BIOS and firmware files...

PlayStation
  ✓ scph5501.bin (USA) - found, verified
  ✗ scph5500.bin (Japan) - missing (optional)
  ✗ scph5502.bin (Europe) - missing (optional)

PlayStation 2
  ✗ ps2-bios.bin - MISSING (required)
    Expected location: ~/Emulation/bios/ps2/

Summary: 1 required file missing, 2 optional files missing
```

---

## Configuration management

We generate emulator config files but users will also change settings via emulator UIs. We need to preserve their changes.

### Approach

On each `apply`, we maintain three versions of managed config files:

- base: what we last wrote
- current: what's on disk (may include user changes)
- new: what we want to write

We use three-way merge. For structured formats (INI, TOML), we can normalize with formatters before diffing to handle whitespace/ordering differences.

Most changes won't conflict since users typically change display/input settings while we manage paths.

If there's a real conflict (both changed the same key), we show the user and ask what to do.

---

## Observability

Users should always know what kyaraben is doing and what changed. Inspired by tools like `nh` that show clear diffs during system updates.

### Version changes

When updating emulators, show what's changing:

```
$ kyaraben update

Checking for updates...

Updates available:
  RetroArch      1.17.0  →  1.18.0
  Dolphin        5.0-21088  →  5.0-21456
  DuckStation    0.1-6722  →  0.1-6891

Apply updates? [y/N]
```

### Config changes

Before applying, show what config changes will be made:

```
$ kyaraben apply

Config changes:
  ~/.config/retroarch/retroarch.cfg
    + system_directory = "/home/user/Emulation/bios"
    + savefile_directory = "/home/user/Emulation/saves"

  ~/.config/duckstation/settings.ini
    ~ BIOS.SearchDirectory: "" → "/home/user/Emulation/bios/psx"

Apply? [y/N]
```

### Status command

Show current state at a glance:

```
$ kyaraben status

Emulation folder: ~/Emulation

Enabled systems: NES, SNES, PlayStation, GameCube
Managed emulators:
  RetroArch      1.18.0    (4 cores)
  Dolphin        5.0-21456
  DuckStation    0.1-6891

Config status: clean (no pending changes)
BIOS status: 1 required file missing (run 'kyaraben doctor')
```

---

## Updates

Kyaraben can check for and apply updates to:

- Itself (the AppImage)
- Managed emulators

```
$ kyaraben update
```

The AppImage can implement self-updating by downloading a new version and replacing itself.

---

## Uninstall

```
$ kyaraben uninstall

This will:
  ✓ Remove emulator binaries from ~/.local/share/kyaraben/
  ✓ Restore original emulator configs from backups

This will NOT touch:
  • ~/Emulation/ (your ROMs, saves, BIOS)
  • ~/.config/kyaraben/ (your kyaraben settings)

Proceed? [y/N]
```

---

## Non-goals

- Windows support
- iOS support
- ROM management / scraping (use ES-DE, Skraper, etc.)
- Netplay configuration
- Deep controller abstraction layer
- Abstracting every emulator setting

## Success criteria

A user with zero emulation experience should be able to:

1. Download kyaraben
2. Select "SNES" and "PlayStation"
3. Click Apply
4. Drop ROMs into the created folders
5. Launch games

---

## Future possibilities (not v1)

- Save sync via Syncthing
- Per-game overrides
- Android companion mode
- Emulator version channels (stable/beta/git)
