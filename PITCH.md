# Emulation Manager — Design Document

> **Note**: This is an early-stage elevator pitch / design exploration. Feature sets, product decisions, UI mockups, and config examples are illustrative, not final. Everything is subject to change as the project develops.

## Overview

A lightweight emulation manager that lets users declaratively configure their emulation setup through a simple UI. Powered by Nix under the hood, but users never see or interact with Nix directly.

## Core Concept

Ship a small (~25-30 MB) binary that:

1. Contains a bundled Nix runtime (portable, no system install required)
2. Presents a UI where users select what they want
3. Fetches and configures emulators on demand from a binary cache
4. Generates all configs, folder structures, and integrations automatically

Users pick systems, the tool does the rest.

## Delivery Methods

### 1. Flatpak (Primary)

The main product: a Flatpak that bundles Nix internally. Users never see or interact with Nix directly.

```bash
flatpak install flathub dev.[project].[Project]
flatpak run dev.[project].[Project]
```

**Why Flatpak:**
- No `curl | sh` trust issues
- Sandboxed, auditable, familiar install flow
- Auto-updates via Flatpak/Flathub
- Works on Steam Deck out of the box (Discover store)
- Single package for all distros

**Alternatively:** AppImage for users who prefer a single portable binary, but Flatpak is the recommended default.

### 2. Home-Manager Module (Power Users)

For users already running NixOS or home-manager, we expose the same functionality as a proper Nix module. No need for the standalone binary — they can declare their emulation setup directly in their existing config.

```nix
# flake.nix or home.nix
{
  imports = [ inputs.[project].homeManagerModules.default ];

  programs.[project] = {
    enable = true;
    
    emulationDir = "~/Emulation";
    
    systems = {
      nes.enable = true;
      snes.enable = true;
      
      psx = {
        enable = true;
        emulator = "retroarch";  # or "duckstation"
        settings = {
          internalResolution = "4x";
          renderer = "vulkan";
        };
      };
      
      gamecube = {
        enable = true;
        emulator = "dolphin";
        channel = "beta";
      };
    };
    
    sync = {
      enable = true;
      # Syncthing integration
    };
    
    frontend = {
      emulationStationDe.enable = true;
    };
  };
}
```

Then just `home-manager switch` — same result as using the UI, but declarative and version-controlled.

**Benefits for Nix users:**
- Integrates with existing dotfiles/config repos
- Reproducible across machines via flakes
- No separate tool to learn — it's just home-manager options
- Can mix with other home-manager modules

**The standalone binary and the home-manager module share the same core logic** — the binary is essentially a wrapper that generates and applies a Nix expression under the hood.

---

## Target Platforms

**Primary:** Steam Deck, desktop Linux (any distro)

**Secondary:** Could extend to dedicated handheld images later

## User Experience

### First Run

1. User downloads single binary (~25 MB)
2. Runs it
3. Sees a welcome screen, picks a location for their emulation folder (default: `~/Emulation`)
4. Selects which systems they want to emulate
5. Clicks "Apply"
6. Tool fetches emulators, generates configs, creates folder structure
7. Done — user has a working emulation setup

### Ongoing Use

- Open the manager to add/remove systems
- Update emulators with one click
- Configure per-system settings (resolution, shaders, etc.)
- Optional: sync saves across devices

## UI Screens

### 1. System Picker

The main screen. Shows available systems grouped by generation/manufacturer.

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
│  Selected: 2 systems (~50 MB)              [ Apply Changes ] │
└──────────────────────────────────────────────────────────────┘
```

**Notes:**
- Show estimated download size based on selection
- Group logically (by company, or by era — TBD)
- Expand/collapse groups for cleaner look on smaller screens

### 2. System Detail / Settings

When user clicks on a system name (not the checkbox), show settings for that system.

```
┌──────────────────────────────────────────────────────────────┐
│  ← Back                         PlayStation                  │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  Emulator: DuckStation (recommended)              [Change]   │
│                                                              │
│  ─────────────────────────────────────────────────────────── │
│  Graphics                                                    │
│  ─────────────────────────────────────────────────────────── │
│  Internal Resolution    [ 4x Native (1440p)        ▼]        │
│  Renderer               [ Vulkan                   ▼]        │
│  VSync                  [✓]                                  │
│                                                              │
│  ─────────────────────────────────────────────────────────── │
│  Paths                                                       │
│  ─────────────────────────────────────────────────────────── │
│  ROMs folder            ~/Emulation/roms/psx      [Browse]   │
│  BIOS folder            ~/Emulation/bios/psx      [Browse]   │
│  Saves folder           ~/Emulation/saves/psx     [Browse]   │
│                                                              │
│  ─────────────────────────────────────────────────────────── │
│  BIOS Status                                                 │
│  ─────────────────────────────────────────────────────────── │
│  ✓ scph5501.bin (USA) — found                                │
│  ✗ scph5500.bin (Japan) — missing                            │
│  ✗ scph5502.bin (Europe) — missing                           │
│                                                              │
├──────────────────────────────────────────────────────────────┤
│                                            [ Save Settings ] │
└──────────────────────────────────────────────────────────────┘
```

**Notes:**
- Sane defaults for everything
- BIOS checker tells users what's missing
- Advanced users can change emulator (e.g., Beetle PSX vs DuckStation)

### 3. Global Settings

```
┌──────────────────────────────────────────────────────────────┐
│  Settings                                                    │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  Emulation Folder       ~/Emulation               [Browse]   │
│                                                              │
│  ─────────────────────────────────────────────────────────── │
│  Integrations                                                │
│  ─────────────────────────────────────────────────────────── │
│  ☐ Add games to Steam (as non-Steam games)                   │
│  ☐ Enable save sync (via Syncthing)                          │
│                                                              │
│  ─────────────────────────────────────────────────────────── │
│  Frontend                                                    │
│  ─────────────────────────────────────────────────────────── │
│  ☐ Install EmulationStation-DE                               │
│  ☐ Install Pegasus                                           │
│                                                              │
│  ─────────────────────────────────────────────────────────── │
│  Updates                                                     │
│  ─────────────────────────────────────────────────────────── │
│  [ Check for Updates ]                                       │
│  Auto-update emulators: [ Weekly ▼ ]                         │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

### 4. Progress / Activity View

Shown when fetching emulators or applying changes.

```
┌──────────────────────────────────────────────────────────────┐
│  Applying changes...                                         │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ✓ Created folder structure                                  │
│  ✓ Fetched RetroArch                                         │
│  ● Fetching DuckStation...                    [=====>    ]   │
│  ○ Generating configs                                        │
│  ○ Setting up Steam integration                              │
│                                                              │
│                                              [ Cancel ]      │
└──────────────────────────────────────────────────────────────┘
```

## Folder Structure Created

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
│   ├── nes/
│   ├── psx/
│   └── ...
├── states/
│   └── (same structure)
└── screenshots/
    └── (same structure)
```

## What Happens Under the Hood

User perspective:
> "I checked PlayStation, clicked Apply, now I have DuckStation configured"

What actually happens:

1. Manager reads current selections
2. Generates a Nix expression representing desired state
3. Calls bundled `nix-portable` to build/fetch the closure
4. Emulator binaries land in `~/.local/share/[app-name]/store/`
5. Manager generates emulator configs (INI, XML, etc.) with correct paths
6. Creates symlinks or wrapper scripts in `~/.local/bin/` or similar
7. Optionally registers with Steam, creates .desktop files, etc.

User never sees Nix. It's an implementation detail.

## Emulator Selection Logic

Each system has:
- A **default emulator** (recommended, works out of the box)
- **Alternatives** (for users who prefer something else)

Philosophy: **RetroArch cores first** when they're good enough. Standalones only where RetroArch falls short (newer consoles, accuracy-critical cases).

| System | Default | Alternatives |
|--------|---------|--------------|
| NES | RetroArch (Mesen) | RetroArch (FCEUmm) |
| SNES | RetroArch (bsnes) | RetroArch (Snes9x) |
| Nintendo 64 | RetroArch (Mupen64Plus-Next) | RetroArch (ParaLLEl) |
| GameCube/Wii | Dolphin | — |
| Switch | Eden (community fork) | — |
| Game Boy / Color | RetroArch (Gambatte) | RetroArch (SameBoy) |
| GBA | RetroArch (mGBA) | mGBA standalone |
| DS | RetroArch (melonDS) | melonDS standalone |
| 3DS | Azahar | — |
| PlayStation | RetroArch (Beetle PSX HW) | DuckStation |
| PlayStation 2 | PCSX2 | — |
| PlayStation 3 | RPCS3 | — |
| PSP | RetroArch (PPSSPP) | PPSSPP standalone |
| PS Vita | Vita3K | — |
| Master System / Genesis | RetroArch (Genesis Plus GX) | RetroArch (PicoDrive) |
| Saturn | RetroArch (Beetle Saturn) | — |
| Dreamcast | RetroArch (Flycast) | Flycast standalone |
| Arcade | RetroArch (FBNeo) | RetroArch (MAME), MAME standalone |
| Atari (2600, 7800, etc.) | RetroArch (Stella, ProSystem) | — |
| TurboGrafx-16 / PC Engine | RetroArch (Beetle PCE) | — |
| Neo Geo | RetroArch (FBNeo) | — |

**Notes:**
- GameCube/Wii: Dolphin standalone is significantly better than the core
- PS2: No viable RetroArch core, PCSX2 standalone required
- PS3: No RetroArch core exists, RPCS3 only option
- Switch: Legal grey area, community forks only (Eden, Sudachi, etc.)
- 3DS: Citra discontinued, Azahar is the merged successor (PabloMK7's fork + Lime3DS)

## Design Decisions

1. **Emulator choice philosophy**: Default to RetroArch cores when available. This is what most users are familiar with, provides a unified interface, and simplifies controller/shader configuration. Standalone emulators only for systems where RetroArch cores are significantly worse or unavailable (PS3, Switch, etc.).

2. **Frontend**: Optional. EmulationStation-DE is the recommended choice (opinionated). Steam integration is a separate optional feature. Both are off by default but easy to enable.

3. **Controller config**: Ship good defaults with auto-detection. Don't lock users out — configs are editable. Details TBD but the goal is "plug in controller, it works".

4. **BIOS handling**: First-class feature. The UI shows exactly what's missing, what's found, and what's optional vs required. A CLI `doctor` command provides a checklist for headless/scripting use.

5. **Multi-device sync**: Core feature, not optional. We're uniquely positioned here — declarative config + Nix + Syncthing integration. Saves sync automatically between devices running this tool.

---

## Configuration & Tinkering

### Philosophy

The UI is a friendly view into a text configuration. Everything the UI can do, the config file can do, and vice versa. Power users can ignore the UI entirely.

```
~/.config/[app-name]/config.toml    # or .yaml, .nix — TBD
```

### Example Config

```toml
[global]
emulation_dir = "~/Emulation"
sync_saves = true

[systems.psx]
enabled = true
emulator = "retroarch"
core = "beetle_psx_hw"

[systems.psx.settings]
internal_resolution = "4x"
renderer = "vulkan"

[systems.psx.paths]
roms = "~/Emulation/roms/psx"
bios = "~/Emulation/bios/psx"

[systems.gamecube]
enabled = true
emulator = "dolphin"
channel = "stable"  # or "beta", "git"

[systems.gamecube.settings]
# Dolphin-specific settings here

[controllers.default]
# Global controller mapping
# TBD: format for this

[games."SLUS-00594"]  # Castlevania: SotN
system = "psx"
[games."SLUS-00594".settings]
internal_resolution = "2x"  # Override for this specific game
```

### Controller Config

TBD, but goals:
- Auto-detect common controllers (Xbox, PlayStation, 8BitDo, etc.)
- Ship good default mappings
- Per-system overrides possible
- Per-game overrides possible
- Config is human-readable and editable
- Don't prevent users from also tweaking RetroArch/emulator configs directly

Open question: Do we abstract controller config across all emulators, or just set up RetroArch and let standalones use their own systems?

### Per-Game Overrides

Users can override settings for specific games. Identification by:
- Serial/ID (e.g., `SLUS-00594`)
- Filename (fallback)

```toml
[games."SLUS-00594"]
system = "psx"
name = "Castlevania: Symphony of the Night"  # Optional, for readability

[games."SLUS-00594".settings]
internal_resolution = "2x"
pgxp = false  # Game has issues with PGXP

[games."Final Fantasy VII (USA) (Disc 1)"]
system = "psx"
[games."Final Fantasy VII (USA) (Disc 1)".settings]
# ...
```

The UI would have a "Game Settings" section where users can search their library and add overrides. Under the hood, it writes to this config.

Challenge: Abstracting settings across different emulators. "internal_resolution = 4x" means different things to DuckStation vs RetroArch vs PCSX2. We need a translation layer.

---

## Emulator Versions & Channels

Users may want newer (or older) versions than we ship by default.

### Channels

Each emulator can have multiple channels:

| Channel | Description |
|---------|-------------|
| `stable` | Default. Well-tested, recommended. |
| `beta` | Newer features, might have bugs. |
| `git` | Latest commit. For enthusiasts. |

```toml
[systems.switch]
emulator = "eden"
channel = "git"  # I want the latest
```

Under the hood, each channel maps to a different Nix derivation (pinned version, or fetched from a flake input).

### Manual Override / Ejecting

For users who download a binary themselves (e.g., AppImage from GitHub):

```toml
[systems.switch]
emulator = "eden"
binary_override = "~/Downloads/Eden-0.4.0.AppImage"
```

When `binary_override` is set:
- We skip fetching that emulator via Nix
- We still generate configs pointing to the user's binary
- Updates for that emulator are the user's responsibility

This lets users "eject" from our management for specific emulators while keeping everything else managed.

### Updating

```
$ [app-name] update              # Update all emulators to latest in their channel
$ [app-name] update --system psx # Update just PSX emulator
$ [app-name] update --channel git --system switch  # Switch to git channel and update
```

UI equivalent: Settings → Updates → [Check for Updates] with per-system channel dropdowns.

---

## BIOS & Firmware Doctor

A CLI command and UI panel that tells users exactly what they need.

### CLI

```
$ [app-name] doctor

Checking BIOS and firmware files...

PlayStation
  ✓ scph5501.bin (USA) — found, verified
  ✗ scph5500.bin (Japan) — missing (optional, needed for JP games)
  ✗ scph5502.bin (Europe) — missing (optional, needed for EU games)

PlayStation 2
  ✗ ps2-bios.bin — MISSING (required)
    Expected location: ~/Emulation/bios/ps2/
    Expected hash: 9a9e8ed5...
    
Nintendo DS
  ✓ bios7.bin — found, verified
  ✓ bios9.bin — found, verified
  ✗ firmware.bin — missing (optional, needed for some features)

Summary: 2 required files missing, 3 optional files missing

Run '[app-name] doctor --help-bios' for guidance on obtaining BIOS files.
```

### UI

The System Detail screen shows BIOS status inline. A dedicated "BIOS Status" panel shows everything at once with:
- Which files are found/missing
- Which are required vs optional
- Expected filenames and hashes
- A button to open the BIOS folder

We verify by hash, not just filename — catches wrong/corrupt files.

---

## Android (Future)

Many handheld devices run Android (Retroid, Odin, AYN, etc.). Supporting Android would extend reach significantly.

### What It Would Look Like

Android can't run Nix the same way. Options:

1. **Companion mode**: Android device syncs with a Linux "server" running the full tool. Saves sync via Syncthing. ROMs sync one-way. Config is managed on the Linux side, Android just consumes it.

2. **Config generator only**: The tool runs on Linux but generates configs for Android emulators (RetroArch Android, standalone Android emulators). User transfers configs manually or via sync.

3. **Native Android app**: A separate Android app that speaks the same config format. Syncs config and saves with the Linux version. Would require significant development effort.

Most realistic path: **Option 1 (companion mode)**. The tool already supports Syncthing. We'd add device profiles that understand Android paths (`/storage/emulated/0/RetroArch/`, etc.) and sync appropriately.

This is not a v1 feature, but the architecture should not preclude it.

---

## Non-Goals

- Windows support (WSL2 is technically possible but not worth the complexity)
- iOS support (closed ecosystem, not happening)
- ROM management / scraping (use ES-DE, Skraper, etc.)
- Netplay configuration (complex, emulator-specific, low priority)

---

## Configuration Management

### The Problem

We generate emulator config files, but:
- Users will change settings via the emulator's own UI (DuckStation settings, RetroArch menus, etc.)
- Those UIs write directly to the config file
- We need to not destroy user changes when we regenerate

### Approach: Generate + Track + Merge

**On first run / adding a system:**
1. Check if config file exists
2. If yes, back it up: `retroarch.cfg` → `retroarch.cfg.bak.2026-01-21`
3. Generate our config (or merge into existing)
4. Record in manifest: file path, hash, backup location

**When user runs `apply` again:**
1. Check manifest — do we manage this file?
2. Compare current file hash to what we wrote
3. If unchanged: overwrite freely
4. If changed: **selective merge** — update only the keys we care about, preserve everything else

**On uninstall:**
1. Read manifest
2. Delete files we created (where no backup exists)
3. Restore backups (where they exist)
4. Remove manifest

### Three-Way Merge Strategy

We treat config files like source code and use proper merge tooling. On each `apply`, we maintain three versions:

- **base**: What we last wrote (stored in manifest or separate file)
- **current**: What's on disk now (includes user's UI changes)
- **new**: What we want to write

```
     base (last apply)
         │
    ┌────┴────┐
    │         │
    ▼         ▼
 current    new
 (user's   (our new
 changes)  settings)
    │         │
    └────┬────┘
         │
         ▼
      merged
```

We use standard merge tools:

```bash
# Three-way merge (POSIX)
diff3 -m current base new > merged

# Or git's merge-file (doesn't need a repo, handles conflicts nicely)
git merge-file -p current base new > merged
```

**Clean merge (no conflicts):**
- User changed `video_shader = "crt.glsl"`
- We changed `system_directory = "/new/path"`
- Both changes preserved automatically, apply silently

**Conflict (same key changed by both):**
- User changed `system_directory = "/their/path"`
- We want `system_directory = "/our/path"`
- Warn user, show diff, ask what to do

Most changes won't conflict because users typically change display/input settings while we manage paths and system settings.

This means users can freely:
- Change resolution in DuckStation's UI
- Configure shaders in RetroArch
- Remap controls in Dolphin

And we won't clobber any of it — their changes merge cleanly with ours.

### Manifest Format

```json
{
  "version": 1,
  "managed_files": [
    {
      "path": "~/.config/retroarch/retroarch.cfg",
      "base_snapshot": "~/.local/share/[app]/baselines/retroarch.cfg",
      "backup": "~/.config/retroarch/retroarch.cfg.bak.2026-01-21"
    },
    {
      "path": "~/.config/duckstation/settings.ini",
      "base_snapshot": "~/.local/share/[app]/baselines/duckstation-settings.ini",
      "backup": null
    }
  ]
}
```

The `base_snapshot` is a copy of the config as we last wrote it — this is the "base" for three-way merge. Updated after each successful `apply`.

### Edge Cases

**User deletes our config**: Fine. We regenerate it on next `apply`, create new baseline.

**User moves config elsewhere**: We detect missing file, regenerate, warn user.

**Config format changes in emulator update**: We parse conservatively. If merge fails due to format changes, we warn and offer to back up + regenerate fresh.

**Merge conflict**: Both user and us changed the same key. We:
1. Show the conflict clearly
2. Offer choices: keep theirs, use ours, or edit manually
3. Optionally open `$EDITOR` or a merge tool for complex cases

```
$ [app] apply

Merging ~/.config/retroarch/retroarch.cfg...

CONFLICT: system_directory
  Base:    /home/user/.local/share/[app]/bios
  Theirs:  /home/user/my-custom-bios-folder
  Ours:    /home/user/Emulation/bios

What do you want to do?
  [t] Keep theirs (/home/user/my-custom-bios-folder)
  [o] Use ours (/home/user/Emulation/bios)
  [e] Edit manually
  [d] Show full diff
> 
```

### What We Don't Do

- We don't try to use `#include` or drop-in directories (most emulators don't support them reliably)
- We don't symlink configs to the Nix store (breaks emulator UI saves)
- We don't prevent users from editing configs directly

### CLI Commands

```
$ [app] doctor --configs

Checking managed configuration files...

~/.config/retroarch/retroarch.cfg
  Status: managed
  User changes since last apply: 12 lines (video settings, shaders)
  Our pending changes: 0
  Merge prediction: clean ✓

~/.config/duckstation/settings.ini  
  Status: managed
  User changes since last apply: 3 lines
  Our pending changes: 2 lines (BIOS path update)
  Merge prediction: clean ✓

~/.config/dolphin-emu/Dolphin.ini
  Status: managed
  User changes since last apply: 5 lines
  Our pending changes: 1 line
  Merge prediction: CONFLICT on ISOPath0
  Run '[app] apply' to resolve
```

---

## Uninstall

```
$ [app] uninstall

This will:
  ✓ Remove emulator binaries from ~/.local/share/[app]/
  ✓ Remove our config manifest
  ✓ Restore original configs from backups (3 files)
  
This will NOT touch:
  • ~/Emulation/ (your ROMs, saves, BIOS)
  • ~/.config/[app]/config.toml (your [app] settings)
  • Emulator configs we didn't back up (none modified before we arrived)

Proceed? [y/N]
```

After uninstall:
- System is as if we were never installed (config-wise)
- User's games, saves, BIOS remain
- User can reinstall fresh or keep their config.toml for next time

---

## Open Questions (Remaining)

1. **Config format**: TOML for our config. Emulator configs stay in their native formats (INI, YAML, etc.).

2. **Controller abstraction**: TBD. Likely: configure RetroArch controllers, let standalones use their own systems. Don't try to unify.

3. **Per-game override identification**: Support both serial/ID (preferred) and filename (fallback). Build/use a lightweight game database.

4. **Settings abstraction**: Shallow. We abstract paths and critical settings. For things like "resolution = 4x", we translate to emulator-native values. We don't try to abstract every setting — users can use emulator UIs for that.

## Success Criteria

A user with zero emulation experience should be able to:

1. Download the app
2. Select "SNES" and "PlayStation"
3. Click Apply
4. Drop ROMs into the created folders
5. Launch games

Total time: under 5 minutes (excluding downloads).
