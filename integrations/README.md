# Kyaraben device integrations

This directory contains device-side integrations for Kyaraben. Each integration enables a specific CFW (Custom Firmware) or device family to sync with Kyaraben.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    Host (desktop, Steam Deck, server)                   │
│                                                                         │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────────────┐  │
│  │ Kyaraben UI │◄──►│  Syncthing  │◄──►│ ROMs / Saves / BIOS / ...   │  │
│  └─────────────┘    └─────────────┘    └─────────────────────────────┘  │
│                            ▲                                            │
└────────────────────────────┼────────────────────────────────────────────┘
                             │ sync via relay or LAN
                             │
┌────────────────────────────▼────────────────────────────────────────────┐
│                    Guest (retro handheld with CFW)                      │
│                                                                         │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────────────┐  │
│  │ Integration │◄──►│  Syncthing  │◄──►│ CFW-specific paths          │  │
│  │     UI      │    │  (bundled)  │    │ (Roms/, /userdata, etc.)    │  │
│  └─────────────┘    └─────────────┘    └─────────────────────────────┘  │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

An integration runs on the device and:

1. Manages Syncthing (bundled, or using the CFW's built-in instance)
2. Maps Kyaraben folder IDs to CFW-specific paths
3. Provides a UI for pairing, status, and configuration

Some CFWs ship with Syncthing built-in (Batocera v33+, KNULLI, ROCKNIX). Integrations for these can configure the existing service rather than bundling a separate binary.

## Shared code

Integrations use these packages from the main Kyaraben codebase:

| Package | Purpose |
|---------|---------|
| `internal/folders` | Category constants and folder ID generation |
| `internal/syncguest` | Syncthing process management and pairing for embedded devices |
| `internal/syncthing` | Syncthing API client |
| `internal/guestapp` | Reusable app framework (menu flow, status, service management) |

## Integration-specific code

Each integration must provide:

| Component | Purpose |
|-----------|---------|
| Path mapping | Translates Kyaraben folder IDs to device paths |
| UI implementation | Device-specific user interface |
| Build configuration | Platform targets and binary dependencies |

## Folder ID scheme

Kyaraben uses a consistent folder ID scheme across all devices:

```
kyaraben-{category}-{system}    # for ROMs, BIOS, saves
kyaraben-{category}-{emulator}  # for states, screenshots
```

Examples:
- `kyaraben-roms-nes` - NES ROMs
- `kyaraben-saves-gba` - GBA save files
- `kyaraben-bios-psx` - PlayStation BIOS
- `kyaraben-states-retroarch:mgba` - RetroArch mGBA save states
- `kyaraben-screenshots-retroarch` - RetroArch screenshots

The `internal/folders` package generates these IDs. Integrations map them to device-specific paths.

## Path mapping example

Each CFW has different conventions. The integration's job is to translate:

| Kyaraben ID | MinUI/NextUI | Batocera/KNULLI | muOS |
|-------------|--------------|-----------------|------|
| `kyaraben-roms-nes` | `Roms/Nintendo (FC)/` | `/userdata/roms/nes/` | `ROMS/fc/` |
| `kyaraben-saves-nes` | `Saves/FC/` | `/userdata/saves/nes/` | emulator-specific |
| `kyaraben-bios-gba` | `Bios/GBA/` | `/userdata/bios/` | `MUOS/bios/` |

## Creating a new integration

### 1. Determine the CFW family

Some CFWs share path conventions and can be covered by one integration:

| Integration | Covers | Base path | Notes |
|-------------|--------|-----------|-------|
| `nextui` | MinUI, NextUI | SD card root | Uses minui-* UI tools, bundles Syncthing |
| `batocera` | Batocera, KNULLI, Koriki | `/userdata` | EmulationStation-based, has built-in Syncthing |
| `rocknix` | ROCKNIX, UnofficialOS, JELOS | `/storage` | Has built-in Syncthing and rclone |
| `muos` | muOS | SD card root | Unique union filesystem, bundles Syncthing |

### 2. Create the integration structure

```
integrations/
└── {name}/
    ├── cmd/{name}/
    │   └── main.go           # entry point
    ├── internal/
    │   ├── config/
    │   │   └── config.go     # path defaults for this CFW
    │   ├── mapping/
    │   │   └── mapping.go    # folder ID to device path translation
    │   └── ui/
    │       └── {ui-type}/    # CFW-specific UI implementation
    ├── build/
    │   ├── launch.sh         # launcher script
    │   └── config.toml.example
    ├── justfile              # build commands
    └── go.mod                # uses replace directive for main module
```

### 3. Implement the path mapping

Use `internal/folders.Category` constants and implement a mapper:

```go
package mapping

import (
    "path/filepath"
    "github.com/fnune/kyaraben/internal/folders"
)

type Mapper struct {
    basePath string
    cfg      Config
}

func (m *Mapper) DevicePath(category folders.Category, system string) string {
    var relativePath string

    switch category {
    case folders.CategoryROMs:
        relativePath = m.cfg.ROMs[system]
    case folders.CategorySaves:
        relativePath = m.cfg.Saves[system]
    // ...
    }

    return filepath.Join(m.basePath, relativePath)
}
```

### 4. Implement the UI

The UI must satisfy the `guestapp.UI` interface:

```go
type UI interface {
    Menu() MenuUI
    Keyboard() KeyboardUI
    Presenter() PresenterUI
}
```

Options vary by CFW:
- MinUI/NextUI: Use `minui-list`, `minui-keyboard`, `minui-presenter` binaries
- Batocera/KNULLI: Use EmulationStation scripting or dialog tools
- Terminal-based: Use a TUI library

### 5. Configure the build

The justfile should:
1. Cross-compile the Go binary for the target architecture
2. Fetch any UI tool dependencies
3. Package everything into the CFW's expected format (PAK, script, etc.)

## CFW landscape reference

See `landscape.md` for a comprehensive survey of CFW path conventions, user bases, and integration priorities.

## Testing

Each integration should include:
- Unit tests for path mapping
- E2E tests using fake UI implementations
- Manual testing on actual hardware

Run tests with:
```
cd integrations/{name}
just test
just e2e
```
