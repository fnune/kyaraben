# MVP scope review

Review of MVP.md requirements against actual implementation.

## Supported systems

| Requirement | Status | Notes |
|-------------|--------|-------|
| SNES with `retroarch:bsnes` | Implemented | Defined in `internal/emulators/registry.go:50-65` |
| PlayStation with `duckstation` | Implemented | Defined in `internal/emulators/registry.go:67-107` |

Additional systems were added beyond MVP scope:
- TIC-80: added for testing purposes (reasonable)
- E2E Test: hidden system using `hello` package for CI (reasonable)

## Domain entities

| Entity | Status | Notes |
|--------|--------|-------|
| `System` | Implemented | `internal/model/system.go` |
| `Emulator` | Implemented | `internal/model/emulator.go` |
| `Provision` | Implemented | `internal/model/provision.go` |
| `State` | Partially implemented | Types defined but sync strategies not wired |
| `EmulatorConfig` | Implemented | `internal/model/emulator_config.go` |
| `KyarabenConfig` | Implemented | `internal/model/config.go` |
| `KyarabenState` | Partially implemented | XDG paths work, but no unified state abstraction |
| `UserStore` | Implemented | `internal/store/user_store.go` |
| `Manifest` | Implemented | `internal/model/manifest.go` |

## CLI commands

| Command | Status | Notes |
|---------|--------|-------|
| `kyaraben apply` | Implemented | `internal/cli/apply.go` |
| `kyaraben doctor` | Implemented | `internal/cli/doctor.go` |
| `kyaraben status` | Implemented | `internal/cli/status.go` |
| `kyaraben init` | Implemented | `internal/cli/init.go` (not in original MVP spec) |
| `kyaraben daemon` | Implemented | `internal/cli/daemon.go` |
| `kyaraben uninstall` | Implemented | `internal/cli/uninstall.go` (listed as out of scope) |

The `init` command was added for better UX and `uninstall` was implemented despite being listed as out of scope for MVP.

## UI

| Requirement | Status | Notes |
|-------------|--------|-------|
| Welcome screen with `UserStore` picker | Unknown | UI code exists but views need verification |
| `System` picker (SNES and PlayStation) | Partial | `get_systems` IPC handler exists |
| Apply button with progress view | Partial | `apply` IPC handler streams progress |
| Settings screen (`UserStore` path) | Unknown | `get_config`/`set_config` handlers exist |
| `Provision` status display for PSX | Unknown | `doctor` IPC handler exists |

The Electron main process (`ui/electron/main.ts`) implements all necessary IPC handlers. The actual UI views would need further review in the Vue/React frontend code.

## `KyarabenConfig` format

MVP specification:
```toml
[global]
user_store = "~/Emulation"

[systems.snes]
emulator = "retroarch:bsnes"

[systems.psx]
emulator = "duckstation"
```

Actual implementation uses a slightly different format:
```toml
[global]
user_store = "~/Emulation"

[systems]
snes = { emulator = "retroarch:bsnes" }
psx = { emulator = "duckstation" }
```

This is a minor divergence where `[systems.snes]` became `[systems] snes = {...}`. Both work, but the actual format is less readable for multi-line per-system config if needed later.

## `UserStore` structure

Matches specification exactly:
```
~/Emulation/
├── roms/{system}/
├── bios/{system}/
├── saves/{system}/
├── states/{system}/
└── screenshots/{system}/
```

Implemented in `internal/store/user_store.go:27-47`.

## `KyarabenState` structure

| Path | Status |
|------|--------|
| `~/.config/kyaraben/config.toml` | Implemented |
| `~/.local/share/kyaraben/store/` | Partial - actually `~/.local/share/kyaraben/nix-portable/` and `flake/` |
| `~/.local/state/kyaraben/manifest.json` | Implemented |
| `~/.local/state/kyaraben/baselines/` | Not implemented |

Baselines for config diffing were specified but not implemented. The `ManagedConfig.BaselineHash` field exists in the manifest but is never populated.

## Doctor output

Matches specification. Example from `internal/cli/doctor.go` produces output like:
```
Checking provisions...

  PlayStation (duckstation)
    ✓ scph5501.bin (USA) - found, verified
    ✗ scph5500.bin (Japan) - MISSING
```

## `EmulatorConfigs` generated

| Emulator | Config | Status |
|----------|--------|--------|
| RetroArch | `system_directory`, `savefile_directory`, `savestate_directory`, `screenshot_directory`, `rgui_browser_directory` | Implemented |
| DuckStation | BIOS search, memory card, save state, screenshots, game list paths | Implemented |

See `internal/emulators/retroarch.go` and `internal/emulators/duckstation.go`.

## Technical milestones from MVP.md

| Milestone | Status |
|-----------|--------|
| 1. Nix flake with RetroArch + bsnes and DuckStation | Implemented in `flake.nix` |
| 2. Home-manager module | Implemented in `nix/hm-module.nix` |
| 3. CLI tool that reads config and invokes Nix | Implemented |
| 4. Manifest tracking | Implemented (without baselines) |
| 5. AppImage packaging with bundled Nix | Partial - nix-portable integration exists |
| 6. Basic GUI wrapping CLI | Implemented via Electron |

## Out of scope items that were implemented anyway

1. `kyaraben uninstall` - fully implemented
2. Per-game overrides - not implemented (correctly out of scope)
3. Observability diffs (show changes before apply) - implemented via `--show-diff` flag

## Gaps and issues

### 1. Config baseline tracking not implemented

MVP specified baselines for three-way merge:
> On each `apply`, we maintain three versions of managed config files: base, current, new

The `ManagedConfig.BaselineHash` field exists but is never set. Config merging works by preserving existing keys but cannot detect conflicts.

### 2. TypeScript types not generated from Go

TASKS.md specified:
> Set up codegen for Go and TypeScript

Go types in `internal/daemon/types.go` should be the source of truth. TypeScript types in `ui/electron/main.ts` are manually duplicated and could drift. Tools like `tygo` can generate TypeScript from Go structs.

The `protocol/schema.json` exists but is disconnected. It could be generated from Go for documentation purposes.

### 3. Emulator version tracking is placeholder

Manifest stores `Version: "latest"` hardcoded (`internal/apply/apply.go:125`). No actual version extraction from Nix store paths.

### 4. Home-manager module diverges from CLI behavior

The home-manager module (`nix/hm-module.nix`) generates config files directly via `xdg.configFile`, which will overwrite any user changes on `home-manager switch`. The CLI preserves existing config values. This behavioral difference is undocumented.

### 5. No fake Nix client for unit/integration tests

TASKS.md specified:
> Fake Nix client for unit/integration tests (working impl, no actual builds)

No fake Nix client exists. The Go CLI E2E tests skip Nix builds using `--dry-run`. However, Playwright UI tests do exercise real Nix builds via the `e2e-test` system. A fake would still be valuable for faster Go-level testing.
