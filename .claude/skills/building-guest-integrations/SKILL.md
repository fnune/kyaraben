---
name: building-guest-integrations
description: Documents how to build guest app integrations for devices that sync with a main Kyaraben installation.
---

# Building guest integrations

Guest integrations are lightweight apps that run on gaming devices (handhelds, retro consoles) to sync saves, states, ROMs, and BIOS files with the main Kyaraben installation.

## Decision framework

### Can the target run Electron?

If the target has a full Linux desktop environment (Batocera, RetroDECK, EmuDeck on SteamOS):

- Build an Electron app that reuses the main Kyaraben UI
- Focus on sync features only (no emulator management)
- This shared desktop sync app is not yet built

### Limited device?

If the target is constrained (tiny handheld, no desktop, limited storage):

- Build a custom integration using platform-specific UI tools
- Bundle syncthing as a static binary
- Package for the CFW's expected format (PAK, script, etc.)

## Architecture

### Shared packages

| Package | Purpose |
|---------|---------|
| `internal/guestapp` | Interfaces (`ServiceManager`, `SyncManager`, `UI`), config types, utilities |
| `internal/syncguest` | Syncthing management: start/stop, pairing, folder config, status |

### What each integration implements

- Syncthing bundling (download/include the binary for target architecture)
- Path mappings (folder ID to device path for the CFW's directory structure)
- Service management (process lifecycle, autostart mechanism)
- UI (using platform-specific tools or frameworks)
- Build system (cross-compilation, packaging)

### Directory structure

```
integrations/{name}/
├── cmd/kyaraben-{name}/main.go
├── internal/
│   ├── app/app.go
│   ├── config/config.go
│   ├── mapping/mapping.go
│   ├── service/
│   └── ui/{type}/
├── build/
├── justfile
└── test/e2e/
```

## Key concepts

### Syncthing isolation

Guest integrations bundle their own syncthing binary and run it with a dedicated port (e.g., 8484) and config directory. This avoids interfering with any system-installed syncthing, matching the main Kyaraben app's approach.

### Service management

Choose based on target platform:

- **PID file tracking**: For devices without systemd. Use `guestapp.PIDProcessController`. Autostart writes scripts to CFW-specific locations.
- **systemd**: For Batocera/desktop Linux. Use `systemctl` commands and ship a unit file.
- **Other init systems**: Research the target's boot sequence and adapt.

### Configuration

Guest apps use TOML config following the main app's conventions, with a subset of fields. Types are in `internal/guestapp/config.go`.

Distinguish between persisted preferences (autostart, path mappings) and runtime state (whether sync is currently running). What belongs in config depends on the platform's UX - NextUI has `autostart` in config but tracks "enabled" in memory since users explicitly launch the app each session.

## Implementation checklist

1. Define path mappings in `config.DefaultConfig()` for the CFW's directory structure
2. Implement `ServiceManager` (see service management section above)
3. Implement `UI` using platform tools (MinUI binaries, terminal, etc.)
4. Wire dependencies in `main.go`
5. Create justfile for cross-compilation and packaging

For sync operations, use `syncguest.Manager` which fully implements `SyncManager`.

## Testing

Use go-vfs for filesystem isolation and fake implementations (`guestapp.NewFakeServiceManager()`, `NewFakeSyncManager()`, `NewFakeUI()`) to test without real syncthing.

See `integrations/nextui/test/e2e/app_test.go` for the pattern.

## Reference implementation

NextUI (`integrations/nextui/`) is the canonical example:

| Concern | Path |
|---------|------|
| Entry point | `cmd/kyaraben-nextui/main.go` |
| App logic | `internal/app/app.go` |
| Path defaults | `internal/config/config.go` |
| Folder mapping | `internal/mapping/mapping.go` |
| Service manager | `internal/service/service.go` |
| MinUI UI | `internal/ui/minui/` |
| Build system | `justfile` |
| E2E tests | `test/e2e/app_test.go` |

## Extraction candidates

When building a second integration, consider extracting shared logic from NextUI. See `internal/guestapp/doc.go` for notes on extraction boundaries.
