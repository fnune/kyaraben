# Kyaraben MVP

Tracer bullet MVP that validates the full technical stack with minimal scope.

## Supported systems

| `System`    | `Emulator`        |
| ----------- | ----------------- |
| SNES        | `retroarch:bsnes` |
| PlayStation | `duckstation`     |

This covers both `Emulator` types (RetroArch core vs standalone) and exercises `Provision` validation (PSX needs BIOS).

## Domain entities exercised

| Entity           | How it's exercised                                     |
| ---------------- | ------------------------------------------------------ |
| `System`         | `snes`, `psx`                                          |
| `Emulator`       | `retroarch:bsnes`, `duckstation`                       |
| `Provision`      | PSX BIOS files (checked by doctor)                     |
| `State`          | Saves and savestates for both systems                  |
| `EmulatorConfig` | `retroarch.cfg`, `settings.ini` (two formats)          |
| `KyarabenConfig` | `~/.config/kyaraben/config.toml`                       |
| `KyarabenState`  | XDG locations for manifest, installed emulators        |
| `UserStore`      | `~/Emulation` with roms, bios, saves, states           |
| `Manifest`       | Tracks managed configs and installed emulator versions |

## In scope

### UI

- Welcome screen with `UserStore` location picker
- `System` picker (SNES and PlayStation)
- Apply button with progress view
- Settings screen (`UserStore` path)
- `Provision` status display for PSX

### CLI

- `kyaraben apply` - apply `KyarabenConfig`, install `Emulators`, generate `EmulatorConfigs`
- `kyaraben doctor` - check `Provision` status
- `kyaraben status` - show current state (see below)

### `KyarabenConfig`

```toml
[global]
user_store = "~/Emulation"

[systems.snes]
emulator = "retroarch:bsnes"

[systems.psx]
emulator = "duckstation"
```

### `UserStore` structure

```
~/Emulation/
├── roms/
│   ├── snes/
│   └── psx/
├── bios/
│   └── psx/
├── saves/
│   ├── snes/
│   └── psx/
└── states/
    ├── snes/
    └── psx/
```

### `KyarabenState` structure

```
~/.config/kyaraben/
└── config.toml              # KyarabenConfig

~/.local/share/kyaraben/
└── store/                   # Nix store, emulator binaries

~/.local/state/kyaraben/
├── manifest.json            # Manifest
└── baselines/               # EmulatorConfig snapshots for diffing
```

### Doctor output

```
$ kyaraben doctor

Provisions:

  PlayStation (duckstation)
    ✓ scph5501.bin (USA) - found, verified
    ✗ scph5500.bin (Japan) - missing
    ✗ scph5502.bin (Europe) - missing

  SNES (retroarch:bsnes)
    No provisions required.
```

### Status output

```
$ kyaraben status

User Store: ~/Emulation

Installed Emulators:
  duckstation (latest)
    Store path: /nix/store/abc123-duckstation
    Installed: 2024-01-15 10:30:00

  retroarch:bsnes (latest)
    Store path: /nix/store/def456-retroarch-bsnes
    Installed: 2024-01-15 10:30:00

Managed Configs:
  ~/.config/duckstation/settings.ini
    Last applied: 2024-01-15 10:30:00
    6 managed keys
    ⚠ User modified since last apply (BIOS.SearchDirectory changed)

  ~/.config/retroarch/retroarch.cfg
    Last applied: 2024-01-15 10:30:00
    8 managed keys
    ✓ Unchanged
```

The status command would:

- Show the configured `UserStore` location
- List installed emulators with their Nix store paths and install times
- List managed config files with:
  - Last apply timestamp
  - Number of managed keys
  - Whether the file was modified by the user since last apply (detected via baseline hash comparison)
  - Which specific managed keys were changed by the user

This supports the config tracking improvements:

1. **Baseline hash comparison**: Each managed config stores a hash of the file after kyaraben applies it. If the current file hash differs, the user modified it.
2. **Managed keys tracking**: The manifest stores which keys kyaraben manages and their expected values, allowing detection of which specific settings were changed.
3. **User modification warnings**: Both `apply --show-diff` and `status` can warn when users have modified managed keys.

### `EmulatorConfigs` generated

RetroArch (`~/.config/retroarch/retroarch.cfg`):

- `system_directory` → `UserStore`/bios
- `savefile_directory` → `UserStore`/saves/snes
- `savestate_directory` → `UserStore`/states/snes
- `rgui_browser_directory` → `UserStore`/roms

DuckStation (`~/.config/duckstation/settings.ini`):

- BIOS search directory → `UserStore`/bios/psx
- Memory card directory → `UserStore`/saves/psx
- Save state directory → `UserStore`/states/psx

## Out of scope (for MVP)

- All other `Systems` and `Emulators`
- `Emulator` updates (`kyaraben update`)
- Self-updating AppImage
- Frontend integration (ES-DE, Steam ROM Manager)
- `Synchronizer` (save sync)
- Per-game overrides
- Observability diffs (show changes before apply) - **partially implemented via `apply --show-diff`**
- Uninstall command
- Config conflict resolution UI (warn and skip on conflict)

## Technical milestones

1. Nix flake with RetroArch + bsnes and DuckStation packages
2. Home-manager module that generates `EmulatorConfigs` for both
3. CLI tool that reads `KyarabenConfig` and invokes Nix
4. `Manifest` tracking for installed `Emulators` and managed configs
5. AppImage packaging with bundled Nix
6. Basic GUI wrapping the CLI

## Open questions

1. Which GUI framework? (Tauri, GTK4, iced, egui)
2. How to bundle nix-portable in AppImage?
3. Where to host binary cache? (Cachix vs self-hosted)
4. Config format for `KyarabenConfig`: is the `[systems.X]` nesting right?
