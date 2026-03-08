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

## Handling flat directories

Some CFWs use flat directory structures where Kyaraben expects per-system or per-emulator subdirectories. Syncthing does not support multiple folders pointing to the same path with different ignore patterns (they share one `.stignore` file and would pick up each other's files). This limits our mapping options.

When a CFW has a flat directory structure, apply one of these strategies:

| Situation | Strategy | Example |
|-----------|----------|---------|
| No good mapping exists | Don't sync, document the limitation | Batocera's flat `/userdata/bios/` |
| Multiple categories share a directory | Pick one category, skip the others | Batocera `/userdata/saves/{system}/` contains both saves and states; sync saves only |
| Single category, reasonable assumption | Map to the most likely Kyaraben folder ID | Flat screenshots dir → `kyaraben-screenshots-retroarch` |

Document any limitations in the integration's README so users understand what syncs and what doesn't.

Examples for Batocera:

| Category | Batocera path | Kyaraben mapping | Notes |
|----------|---------------|------------------|-------|
| ROMs | `/userdata/roms/{system}/` | `kyaraben-roms-{system}` | Per-system, works normally |
| Saves | `/userdata/saves/{system}/` | `kyaraben-saves-{system}` | Works, but states excluded |
| States | `/userdata/saves/{system}/` | Not synced | Shares directory with saves |
| BIOS | `/userdata/bios/` | Not synced | Flat directory, no per-system mapping |
| Screenshots | `/userdata/screenshots/` | `kyaraben-screenshots-retroarch` | Flat, but reasonable assumption |

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
1. Cross-compile the Go binary for all target architectures
2. Fetch any UI tool dependencies
3. Package everything into the CFW's expected format (PAK, script, etc.)

### 6. Build for multiple architectures

CFWs often run on multiple hardware platforms. Build binaries for all supported architectures and let users download the correct one:

| Platform | Architecture | Go build flags |
|----------|--------------|----------------|
| PC, Steam Deck, x86 SBCs | x86_64 | `GOOS=linux GOARCH=amd64` |
| Raspberry Pi 4/5, most ARM64 SBCs | arm64 | `GOOS=linux GOARCH=arm64` |
| Older Pi, 32-bit ARM devices | arm | `GOOS=linux GOARCH=arm GOARM=7` |
| RISC-V boards | riscv64 | `GOOS=linux GOARCH=riscv64` |

Example justfile recipe:

```just
build-all:
    GOOS=linux GOARCH=amd64 go build -o dist/kyaraben-example-amd64 ./cmd/kyaraben-example
    GOOS=linux GOARCH=arm64 go build -o dist/kyaraben-example-arm64 ./cmd/kyaraben-example
    GOOS=linux GOARCH=arm GOARM=7 go build -o dist/kyaraben-example-arm ./cmd/kyaraben-example
```

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
