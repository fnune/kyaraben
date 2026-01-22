# Architecture review

Review of architectural decisions and their implementation.

## Overall architecture

The architecture follows the design documents well:

```
┌─────────────────────────────────────┐
│   UI (Electron + Vite)              │
│   └── electron/main.ts              │
└────────────────┬────────────────────┘
                 │ JSON over stdio
┌────────────────▼────────────────────┐
│   CLI / Daemon (Go)                 │
│   ├── cmd/kyaraben/main.go          │
│   ├── internal/cli/*                │
│   └── internal/daemon/*             │
└────────────────┬────────────────────┘
                 │
    ┌────────────┴────────────┐
    │                         │
    ▼                         ▼
┌─────────────────┐   ┌─────────────────┐
│ Config/Manifest │   │  nix-portable   │
│ XDG directories │   │  Flake build    │
└─────────────────┘   └─────────────────┘
```

## Daemon protocol

### Design decision

TASKS.md specified:
> Protocol: JSON over stdin/stdout, strictly typed (schema TBD)

Implementation: JSON-RPC-like protocol without request IDs.

### Issue: no request ID correlation

The daemon handles commands sequentially and returns events without request IDs. The UI must process responses in order:

`ui/electron/main.ts:109-114`:
```typescript
const firstPending = daemon?.pending.entries().next().value
if (firstPending) {
    const [id, { resolve }] = firstPending
    daemon?.pending.delete(id)
    resolve(event)
}
```

This works but prevents concurrent requests. If the UI sends two commands before receiving the first response, responses could be mismatched.

### Issue: apply streams events without framing

The `apply` command returns multiple progress events followed by a result/error. The daemon emits these as separate JSON lines:

`internal/daemon/daemon.go:226-234`:
```go
opts := apply.Options{
    OnProgress: func(p apply.Progress) {
        events = append(events, Event{
            Type: EventProgress,
            Data: map[string]interface{}{...},
        })
    },
}
```

The UI handles this specially in `applyCommand()`, but there is no protocol-level indication that more events will follow. The UI just keeps listening until it gets a result or error.

### Recommendation

Add request IDs to commands and events:
```json
{"id": 1, "type": "apply", "data": {...}}
{"id": 1, "type": "progress", "data": {...}}
{"id": 1, "type": "result", "data": {...}}
```

## Nix integration

### Design decision

PITCH.md specified:
> Ship a binary that contains a bundled Nix runtime (portable, no system install required)

Implementation uses nix-portable, searched at runtime:

`internal/nix/client.go:62-72`:
```go
searchPaths := []string{
    filepath.Join(execDir, binaryName),
    filepath.Join(execDir, "..", "binaries", binaryName),
}
```

### Issue: nix-portable not bundled

The binary search suggests nix-portable should be placed alongside the kyaraben binary, but there is no packaging that actually bundles it. The user must obtain nix-portable separately.

### Issue: no binary cache configuration

PITCH.md mentioned:
> Fetches and configures emulators on demand from a binary cache

Implementation generates flakes that reference nixpkgs but does not configure a kyaraben-specific binary cache. First-time users will build from source or use the generic Nix cache.

TASKS.md open question:
> Where to host binary cache? (Cachix vs self-hosted)

This remains unresolved.

## Config generation

### Design decision

FILESYSTEM.md specified:
> Option 1 (configure emulators directly) is the only approach.

Implementation follows this. Config generators write to emulator config files.

### Issue: config generators are tightly coupled

Each emulator has a hardcoded generator:

`internal/emulators/config.go:18-29`:
```go
func GetConfigGenerator(emuID model.EmulatorID) ConfigGenerator {
    switch emuID {
    case model.EmulatorRetroArchBsnes:
        return &RetroArchConfig{}
    case model.EmulatorDuckStation:
        return &DuckStationConfig{}
    case model.EmulatorTIC80:
        return &TIC80Config{}
    default:
        return nil
    }
}
```

Adding a new emulator requires modifying this switch. The registry should associate emulators with their generators.

### Issue: home-manager module duplicates config logic

The home-manager module (`nix/hm-module.nix`) generates the same configs as Go code:

```nix
xdg.configFile."retroarch/retroarch.cfg" = mkIf (retroarchConfig != "") {
    text = retroarchConfig;
};
```

Changes to config format must be made in two places. PITCH.md stated:
> The standalone binary and the home-manager module share the same core logic.

This is not achieved. They have parallel implementations.

## Storage layout

### Design decision

MODEL.md specified XDG-compliant paths:
- Config: `~/.config/kyaraben/`
- Data: `~/.local/share/kyaraben/`
- State: `~/.local/state/kyaraben/`

Implementation follows this via `internal/paths/xdg.go`.

### Issue: flake location is under data, not state

The generated flake is stored in `~/.local/share/kyaraben/flake/`. Flakes are regenerated on each apply, which is state behavior, not data.

XDG spec:
- DATA_HOME: user-specific data files
- STATE_HOME: user-specific state files (logs, history, recently used)

Generated flakes are arguably state since they are transient build artifacts.

## CLI framework

### Design decision

TASKS.md open question:
> CLI framework: `cobra`, `urfave/cli`, or stdlib `flag`?

Implementation uses `kong` (alecthomas/kong). This is a reasonable choice not listed in the options.

### Observation

Kong is used well. The CLI struct in `cmd/kyaraben/main.go` is clean and commands are properly separated.

## UI framework

### Design decision

TASKS.md open question:
> UI framework: Electron, Tauri, or web served locally?

Implementation uses Electron with Vite.

### Issue: Tauri artifacts remain

The UI search paths reference `src-tauri`:

`ui/electron/main.ts:54`:
```typescript
searchPaths.push(path.join(appPath, 'src-tauri', 'binaries', sidecarName))
```

This suggests a migration from Tauri to Electron. The path references should be cleaned up.

## Registry design

### Observation

The registry is a good domain service pattern:

`internal/emulators/registry.go`:
```go
type Registry struct {
    systems   map[model.SystemID]model.System
    emulators map[model.EmulatorID]model.Emulator
}
```

### Issue: hardcoded registration

Systems and emulators are registered in code:

```go
func (r *Registry) registerSystems() {
    r.systems[model.SystemSNES] = model.System{...}
}
```

For the MVP this is fine. For extensibility, consider data-driven registration from configuration files.

### Issue: default emulators hardcoded separately

`Registry.GetDefaultEmulator()` has a separate map:

```go
defaults := map[model.SystemID]model.EmulatorID{
    model.SystemSNES:    model.EmulatorRetroArchBsnes,
    ...
}
```

This duplicates knowledge. The default should be part of system registration or derived from the first emulator registered for a system.

## Manifest design

### Issue: manifest path is computed repeatedly

`model.DefaultManifestPath()` is called in multiple places. The path should be computed once and passed through.

### Issue: no atomic writes

`manifest.Save()` uses `os.WriteFile()` which is not atomic:

```go
if err := os.WriteFile(path, data, 0644); err != nil {
```

A crash during write could corrupt the manifest. Consider write-to-temp-then-rename pattern.

## Concurrency

### Observation

The codebase is single-threaded. The daemon handles one command at a time. This is appropriate for the current scope.

### Future consideration

If multiple apply operations or concurrent provision checks are needed, the current design would need synchronization. The manifest, in particular, would need locking.

## Summary of architectural issues

| Issue | Impact | Recommendation |
|-------|--------|----------------|
| No request IDs in protocol | Medium | Add correlation IDs |
| nix-portable not bundled | High | Implement AppImage packaging |
| No binary cache | Medium | Set up Cachix |
| Duplicate config logic (Go + Nix) | High | Share config generation or generate Nix from Go |
| Flake in data not state dir | Low | Move to state dir |
| Tauri path artifacts | Low | Remove dead paths |
| Non-atomic manifest writes | Low | Use atomic write pattern |
