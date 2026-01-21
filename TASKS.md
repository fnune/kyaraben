# Kyaraben MVP tasks

Breakdown of work to build the MVP.

## Technical decisions

- Backend/CLI: Go
- UI: TypeScript
- Protocol: JSON over stdin/stdout, strictly typed (schema TBD)
- Nix integration: generate flake, shell out to `nix` / `nix-portable`
- Testing: unit tests, integration tests, E2E with open-licensed emulator (TIC-80 or similar)

## Project structure

```
kyaraben/
├── PITCH.md
├── MODEL.md
├── MVP.md
├── TASKS.md
├── FILESYSTEM.md
├── CONTRIBUTING.md
│
├── cmd/
│   └── kyaraben/
│       └── main.go              # CLI entrypoint
│
├── internal/
│   ├── model/
│   │   ├── system.go            # `System`
│   │   ├── emulator.go          # `Emulator`
│   │   ├── provision.go         # `Provision`
│   │   ├── state.go             # `State`
│   │   ├── emulator_config.go   # `EmulatorConfig`
│   │   └── config.go            # `KyarabenConfig`
│   │
│   ├── store/
│   │   ├── user_store.go        # `UserStore` management
│   │   └── manifest.go          # `Manifest`
│   │
│   ├── emulators/
│   │   ├── registry.go          # Known `Systems` and `Emulators`
│   │   ├── retroarch.go         # RetroArch `EmulatorConfig` generation
│   │   └── duckstation.go       # DuckStation `EmulatorConfig` generation
│   │
│   ├── nix/
│   │   ├── client.go            # Nix CLI abstraction
│   │   └── flake.go             # Flake generation
│   │
│   ├── daemon/
│   │   ├── daemon.go            # JSON protocol handler
│   │   ├── commands.go          # Command handlers
│   │   └── events.go            # Event types
│   │
│   └── cli/
│       ├── apply.go
│       ├── doctor.go
│       ├── status.go
│       └── daemon.go            # Starts daemon mode
│
├── protocol/
│   ├── schema.json              # Protocol schema (source of truth)
│   └── generated/
│       ├── protocol.go          # Generated Go types
│       └── protocol.ts          # Generated TypeScript types
│
├── nix/
│   ├── devshell.nix             # Dev environment for `use flake`
│   ├── package.nix              # Builds kyaraben Go binary
│   ├── hm-module.nix            # Home-manager module
│   ├── emulators.nix            # Emulator packages/references
│   └── appimage.nix             # AppImage bundling
│
├── ui/                          # TypeScript UI
│   ├── package.json
│   ├── src/
│   │   ├── main.ts              # Spawns daemon, handles IPC
│   │   ├── protocol.ts          # Generated types
│   │   └── ...
│   └── ...
│
├── test/
│   ├── fixtures/                # Test ROMs, configs
│   │   └── tic80/               # TIC-80 carts for E2E
│   ├── e2e/                     # E2E tests
│   │   └── e2e_test.go
│   └── integration/             # Integration tests
│       └── cli_test.go
│
├── go.mod
├── go.sum
└── flake.nix                    # Entry point: inputs, imports from nix/*
```

## Task breakdown

### Phase 1: foundation

#### 1.1 Project scaffolding

- [ ] Initialize Go module
- [ ] Set up directory structure
- [ ] Set up Nix flake for Go build
- [ ] Add basic CLI with `cobra` or similar
- [ ] Set up test infrastructure (Go testing, test helpers)

#### 1.2 Domain model types

- [ ] Define `System`
- [ ] Define `Emulator`
- [ ] Define `Provision`
- [ ] Define `State`
- [ ] Define `EmulatorConfig`
- [ ] Define `KyarabenConfig` (TOML parsing with `BurntSushi/toml`)
- [ ] Define `Manifest` (JSON serialization)

#### 1.3 Registry of `Systems` and `Emulators`

- [ ] Hardcode `snes` and `psx` `Systems`
- [ ] Hardcode `retroarch:bsnes` and `duckstation` `Emulators`
- [ ] Define `Provisions` for PSX (BIOS files with hashes)

Note: each `Emulator` ID represents a stable config/state contract. Breaking emulator changes become new IDs (e.g., `duckstation-legacy-1x`).

### Phase 2: core functionality

#### 2.1 `UserStore` management

- [ ] Create `UserStore` directory structure
- [ ] Create per-system subdirectories (roms, bios, saves, states)

#### 2.2 `EmulatorConfig` generation

- [ ] RetroArch config generator
  - [ ] Set paths (system_directory, savefile_directory, etc.)
  - [ ] Enable bsnes core
- [ ] DuckStation config generator
  - [ ] Set BIOS path
  - [ ] Set memory card path
  - [ ] Set save state path

#### 2.3 `Manifest` tracking

- [ ] Write `Manifest` on apply
- [ ] Track installed `Emulators` with versions
- [ ] Track managed `EmulatorConfigs` with base snapshots
- [ ] Read `Manifest` for status

#### 2.4 `Provision` checking

- [ ] Scan `UserStore`/bios for expected files
- [ ] Verify file hashes
- [ ] Report found/missing status

### Phase 3: Nix integration

#### 3.1 Nix client abstraction

- [ ] Implement `Nix` struct with configurable binary and store paths
- [ ] Implement `Build(flakeRef)` method
- [ ] Implement `Update()` method
- [ ] Handle errors and parse output

#### 3.2 Flake generation

- [ ] Generate `flake.nix` from `KyarabenConfig`
- [ ] Template that imports kyaraben's emulator packages
- [ ] Write to `KyarabenState`

#### 3.3 Emulator packaging

- [ ] Package RetroArch + bsnes in kyaraben flake (or reference nixpkgs)
- [ ] Package DuckStation in kyaraben flake (or reference nixpkgs)
- [ ] Create `mkEmulatorEnv` helper function
- [ ] Verify packages build and run

#### 3.4 Home-manager module

- [ ] Create module with `programs.kyaraben` options
- [ ] Generate same `EmulatorConfigs` as CLI
- [ ] Test with `home-manager switch`

### Phase 4: CLI

#### 4.1 `kyaraben apply`

- [ ] Parse `KyarabenConfig`
- [ ] Create `UserStore` structure
- [ ] Generate flake
- [ ] Invoke Nix to build
- [ ] Generate `EmulatorConfigs`
- [ ] Write `Manifest`
- [ ] Show progress to terminal

#### 4.2 `kyaraben doctor`

- [ ] Check `Provisions` for enabled `Systems`
- [ ] Display found/missing with verification status
- [ ] Exit code reflects status

#### 4.3 `kyaraben status`

- [ ] Read `Manifest`
- [ ] Show enabled `Systems` and `Emulators`
- [ ] Show installed versions
- [ ] Show `Provision` summary

### Phase 5: daemon and protocol

#### 5.1 Protocol design

- [ ] Define JSON schema for commands and events
- [ ] Commands: `apply`, `status`, `doctor`, `config.set`, `config.enable`, etc.
- [ ] Events: `progress`, `result`, `error`
- [ ] Set up codegen for Go and TypeScript

#### 5.2 Daemon implementation

- [ ] `kyaraben daemon` command
- [ ] Read JSON commands from stdin
- [ ] Write JSON events to stdout
- [ ] Route commands to handlers
- [ ] Stream progress events

### Phase 6: AppImage

#### 6.1 Bundle Nix

- [ ] Research nix-portable bundling
- [ ] Create AppImage build process
- [ ] Test on clean system (no Nix installed)

#### 6.2 Self-contained operation

- [ ] Ensure `KyarabenState` paths work from AppImage
- [ ] Ensure Nix store is isolated to `~/.local/share/kyaraben/store`
- [ ] Test full flow: download AppImage → run → apply → play game

### Phase 7: UI

#### 7.1 Project setup

- [ ] Initialize TypeScript project
- [ ] Choose framework (Electron, Tauri, or web + local server?)
- [ ] Set up protocol types from generated schema

#### 7.2 Daemon communication

- [ ] Spawn `kyaraben daemon` subprocess
- [ ] Send commands via stdin
- [ ] Parse events from stdout
- [ ] Handle connection lifecycle

#### 7.3 Basic screens

- [ ] Welcome / `UserStore` picker
- [ ] `System` picker (checkboxes for SNES, PSX)
- [ ] Apply button with progress
- [ ] Settings (`UserStore` path)
- [ ] `Provision` status display

### Testing (cross-cutting)

#### Unit tests

- [ ] Model types (serialization, validation)
- [ ] `UserStore` path generation
- [ ] `EmulatorConfig` generation (RetroArch, DuckStation)
- [ ] `Provision` hash verification
- [ ] `Manifest` read/write

#### Integration tests

- [ ] CLI commands (`apply`, `doctor`, `status`) with mock Nix
- [ ] `KyarabenConfig` parsing and validation
- [ ] Flake generation

#### E2E test harness

- [ ] Research open-licensed emulators for testing (TIC-80, others?)
- [ ] Package test emulator in Nix
- [ ] Create test fixture format (config + expected outcomes)
- [ ] Harness: spin up isolated `UserStore` and `KyarabenState`
- [ ] Harness: run kyaraben commands (invokes actual `nix` binary)
- [ ] Harness: assert on filesystem state, exit codes, output

#### Test infrastructure

- [ ] CI pipeline running tests (unit, integration, E2E)
- [ ] Test helpers for temporary directories
- [ ] Fake Nix client for unit/integration tests (working impl, no actual builds)
- [ ] E2E tests invoke actual `nix` binary (slow, validates full flow)

## Dependencies between phases

```
Phase 1 (foundation)
    │
    ▼
Phase 2 (core functionality)
    │
    ├───────────────────┬──────────────────┐
    ▼                   ▼                  ▼
Phase 3 (Nix)      Phase 4 (CLI)     Phase 5 (daemon/protocol)
    │                   │                  │
    └─────────┬─────────┘                  │
              ▼                            │
       Phase 6 (AppImage)                  │
              │                            │
              └──────────┬─────────────────┘
                         ▼
                   Phase 7 (UI)
```

## Open decisions

1. JSON Schema tooling: which codegen for Go + TypeScript?
2. CLI framework: `cobra`, `urfave/cli`, or stdlib `flag`?
3. UI framework: Electron, Tauri, or web served locally?
4. How to bundle nix-portable in AppImage?
