# Kyaraben NextUI PAK

Build a Tool PAK for NextUI devices that integrates with Kyaraben's sync ecosystem.

Target: all devices supported by NextUI, testing on TrimUI Brick.

## Progress

### Done

- [x] Created `internal/syncthing/` package with extracted, dependency-free code:
  - `client.go`: Syncthing HTTP client
  - `relay.go`: Relay client for pairing codes
  - `types.go`: Shared types (ConnectionInfo, FolderStatus, etc.)
  - `config.go`: Config types with defaults
  - `interface.go`: SyncClient interface
  - `logger.go`: Injectable logger interface (no dependency on internal/logging)
- [x] Updated `internal/sync/` to use `internal/syncthing/` (removed ~900 lines of duplication)
- [x] Created `integrations/nextui/` directory structure:
  - `cmd/kyaraben-nextui/main.go`: Main binary entry point
  - `internal/mapping/`: Kyaraben system <-> NextUI TAG mapping with tests
  - `internal/ui/interface.go`: UI abstraction interfaces (MenuUI, KeyboardUI, PresenterUI)
  - `internal/ui/minui/`: minui-list/keyboard/presenter wrappers
  - `internal/ui/fake/`: Fake UI for testing
  - `internal/config/`: TOML config with tag_overrides support
  - `internal/app/`: Main application loop (uses syncguest)
- [x] Created `internal/syncguest/` package - generic sync management for guest devices:
  - Manages Syncthing process lifecycle (start/stop/wait for ready)
  - Pairing flow (create session, join session, wait for peer)
  - Peer management (add peer, share folders)
  - Status reporting
- [x] Created `cmd/kyaraben-guest/` CLI:
  - `status`: Show sync status and connected peers
  - `pair`: Generate pairing code and wait for peer
  - `join <code>`: Join using a pairing code
  - `start`: Start Syncthing and keep running

### Next steps

1. Add build system for ARM64 targets
2. Create launch.sh and PAK packaging
3. E2E tests with fakes

### Architectural note

The core sync functionality is not NextUI-specific. `kyaraben-guest` is the generic CLI for any device joining a Kyaraben cluster:
- Android, other handhelds (OnionOS, GarlicOS), desktop testing
- Manages Syncthing with Kyaraben ports/relay
- Simple CLI for pairing and status

The NextUI PAK is a thin wrapper:
- Provides minui-based UI
- Handles TAG -> Kyaraben system mapping
- Bundles platform-specific binaries

## Scope

### What we sync

- Native saves (.sav/.srm) - per system
- Screenshots
- ROMs
- BIOS files

### What we do not sync (for now)

- Save states: not portable across core versions or different cores, low priority
  - May add in the future with careful core matching

## Folder mapping

Kyaraben uses folder IDs like `kyaraben-{category}-{system}`. NextUI uses TAGs.

### System ID to TAG mapping (base systems)

| Kyaraben system | NextUI TAG | NextUI core |
|-----------------|------------|-------------|
| nes | FC | fceumm |
| snes | SFC | snes9x |
| gb | GB | gambatte |
| gbc | GBC | gambatte |
| gba | GBA | gpsp |
| psx | PS | pcsx_rearmed |
| genesis | MD | picodrive |

### Extras (common ones)

| Kyaraben system | NextUI TAG | NextUI core |
|-----------------|------------|-------------|
| gamegear | GG | picodrive |
| mastersystem | SMS | picodrive |
| pcengine | PCE | mednafen_pce_fast |
| ngp | NGP | race |
| atari2600 | A2600 | stella |
| c64 | C64 | vice_x64 |
| arcade | FBN | fbneo |

### Path mapping

| Category | Kyaraben folder ID | Kyaraben path | NextUI path |
|----------|-------------------|---------------|-------------|
| ROMs | kyaraben-roms-{system} | ~/Emulation/roms/{system}/ | /Roms/{Display Name} ({TAG})/ |
| Saves | kyaraben-saves-{system} | ~/Emulation/saves/{system}/ | /Saves/{TAG}/ |
| BIOS | kyaraben-bios-{system} | ~/Emulation/bios/{system}/ | /Bios/{TAG}/ |
| Screenshots | kyaraben-screenshots | ~/Emulation/screenshots/ | /Screenshots/ |

### Alternative cores (power user config)

Users with alternative core PAKs (MGBA.pak, SUPA.pak) can configure overrides:

```toml
# /.userdata/{platform}/kyaraben/config.toml
[tag_overrides]
MGBA = "gba"   # /Saves/MGBA/ syncs to kyaraben-saves-gba
SUPA = "snes"  # /Saves/SUPA/ syncs to kyaraben-saves-snes
SGB = "gba"    # Super Game Boy saves to GBA folder
```

Default: only base TAGs sync. Overrides let alternative TAGs map to the same Kyaraben folder.

## Features

### Syncthing management

- Bundled Syncthing binary for each target architecture
- Uses Kyaraben ports (8484 GUI, 22100 listen, 21127 discovery) to coexist with minui-syncthing-pak (8384)
- Manages startup, shutdown, persistence across reboots
- Stores config in `/.userdata/{platform}/kyaraben/syncthing/`

### Pairing flow

- Uses Kyaraben's relay server for WAN pairing
- Two modes: generate code (initiator) or enter code (joiner)
- After pairing, all folders auto-shared (no accept step)
- Pairing codes are 6 uppercase alphanumeric characters (A-Z, 0-9)
  - Consider UX: minui-keyboard supports caps but requires extra keystrokes
  - May need to display instructions or consider alternative input methods

### UI

Main menu:
- Sync status (synced/syncing/disconnected/error)
- Start/Stop syncing
- Enable/Disable on boot
- Pair new device
- Show Syncthing UI URL
- View paired devices

### Error handling

- Readable, actionable error messages
- Retries for transient failures (network, Syncthing startup)
- Logging to `/.userdata/{platform}/logs/Kyaraben.txt`

## Technical details

### PAK structure

```
Tools/{platform}/Kyaraben.pak/
├── launch.sh           # Entry point, calls the Go binary
├── kyaraben-nextui     # Our Go binary (ARM64)
├── syncthing           # Syncthing binary (ARM64)
├── minui-list          # minui-list binary (platform-specific)
├── minui-keyboard      # minui-keyboard binary (platform-specific)
└── config.toml.example # Example config for power users
```

### Bundled binaries

All binaries are downloaded at build time from official releases (same approach as minui-syncthing-pak):

| Binary | Source | Architecture |
|--------|--------|--------------|
| syncthing | github.com/syncthing/syncthing/releases | linux-arm64 |
| minui-list | github.com/josegonzalez/minui-list/releases | platform-specific (tg5040, tg5050) |
| minui-keyboard | github.com/josegonzalez/minui-keyboard/releases | platform-specific |

Build system will fetch these during `make build` or similar.

### Environment variables (provided by NextUI)

- `SDCARD_PATH`: SD card root (e.g., `/mnt/SDCARD`)
- `SAVES_PATH`: saves directory (e.g., `/mnt/SDCARD/Saves`)
- `BIOS_PATH`: BIOS directory
- `USERDATA_PATH`: user data (e.g., `/.userdata/tg5040/`)
- `LOGS_PATH`: logs directory
- `PLATFORM`: platform ID (tg5040, tg5050)

### Build targets

NextUI maintained platforms:
- tg5040: covers TrimUI Brick and TrimUI Smart Pro (detected at runtime via TRIMUI_MODEL)
- tg5050: TrimUI Smart Pro S

Both are ARM64 (aarch64). Syncthing provides official linux-arm64 builds.

### Testing strategy

- Fake minui-list and minui-keyboard binaries for E2E tests
- Fake Syncthing client (reuse from Kyaraben)
- Real Syncthing in container tests
- Unit tests for folder mapping logic

## Code organization

Location: `integrations/nextui/`

### Architecture

Build abstractions around external dependencies so they can be swapped later:

```
integrations/nextui/
├── cmd/
│   └── kyaraben-nextui/    # Main binary
├── internal/
│   ├── ui/                 # UI abstraction layer
│   │   ├── interface.go    # UI interface (list, keyboard, presenter)
│   │   ├── minui/          # minui-list/keyboard implementation
│   │   └── fake/           # Fake implementation for testing
│   ├── sync/               # Syncthing management (reuse from kyaraben)
│   ├── pairing/            # Pairing flow (reuse relay client)
│   ├── mapping/            # Kyaraben <-> NextUI folder mapping
│   └── config/             # Config file handling
└── ...
```

Key interfaces to define:

```go
type MenuUI interface {
    Show(items []MenuItem, options MenuOptions) (selected int, action Action, err error)
}

type KeyboardUI interface {
    GetInput(options KeyboardOptions) (string, error)
}

type PresenterUI interface {
    ShowMessage(text string) error
    ShowProgress(text string, percent int) error
    Close() error
}
```

### Extracted shared package

`internal/syncthing/` - standalone Syncthing client and relay, no dependencies on internal/logging or internal/model:
- `client.go`: HTTP client for Syncthing REST API
- `relay.go`: Relay client for pairing code exchange
- `config.go`: Config types with Kyaraben defaults (ports 8484/22100/21127)
- `types.go`: Shared types (ConnectionInfo, FolderStatus, etc.)
- `interface.go`: SyncClient interface for mocking
- `logger.go`: Injectable logger interface

Both `internal/sync/` and `integrations/nextui/` can import this package.

## References

- https://github.com/josegonzalez/minui-list - scrollable list UI
- https://github.com/josegonzalez/minui-keyboard - text input UI
- https://github.com/josegonzalez/minui-syncthing-pak - existing Syncthing pak (coexistence reference)
- https://github.com/LoveRetro/nextui-pak-store - eventual publish target
- https://github.com/LoveRetro/NextUI/blob/main/PAKS.md - PAK structure docs
- ~/Development/NextUI - local clone for path reference
