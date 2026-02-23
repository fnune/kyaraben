# Claude instructions for kyaraben

Kyaraben is a declarative emulation manager for Linux. Users pick gaming
systems (SNES, PSX, GameCube, etc.), and kyaraben installs emulators,
generates config files, creates directory structures, and manages saves. It
runs as an Electron app with a Go backend (daemon). The primary target is
Steam Deck; it works on any Linux distribution.

Read and follow `site/src/content/docs/contributing.mdx` for full detail on
architecture, conventions, testing, and adding systems or emulators.

## Quick orientation

- `site/src/content/docs/index.mdx`: what kyaraben is and how it works
- `site/src/content/docs/getting-started.mdx`: installation and first-run walkthrough
- `site/src/content/docs/using-the-app.mdx`: app reference and guarantees
- `site/src/content/docs/using-the-cli.mdx`: CLI reference

## Repository structure

```
cmd/kyaraben/          CLI entry point (kong CLI framework)
cmd/check-versions/    Version update checker
internal/
  model/               Domain types (System, Emulator, Provision, State, etc.)
  registry/            System and emulator registry (all.go wires definitions)
  emulators/           Per-emulator config generators and definitions
    retroarch/         Shared RetroArch config (retroarch.cfg)
    retroarch*/        Per-core RetroArch wrappers (bsnes, mgba, mesen, etc.)
    duckstation/       DuckStation (PSX)
    pcsx2/             PCSX2 (PS2)
    dolphin/           Dolphin (GC, Wii)
    ppsspp/            PPSSPP (PSP)
    cemu/              Cemu (Wii U)
    eden/              Eden (Switch)
    flycast/           Flycast (Dreamcast)
    rpcs3/             RPCS3 (PS3)
    vita3k/            Vita3K (PS Vita)
    xemu/              xemu (Xbox)
    xeniaedge/         Xenia Edge (Xbox 360)
    symlink/           Symlink creation utilities
  systems/             Per-system definitions (26 systems)
  apply/               Apply orchestration
  daemon/              Daemon protocol (JSON over stdin/stdout) and handlers
  cli/                 CLI command implementations
  store/               UserStore management (~/Emulation)
  sync/                Syncthing integration for save synchronization
  frontends/           Frontend integrations (EmulationStation DE)
  launcher/            Desktop launcher (.desktop entry) management
  doctor/              BIOS/provision checking
  status/              Status reporting
  paths/               XDG path management
  packages/            Emulator package download and installation
  configformat/        Config file format parsers (INI, TOML, XML, YAML)
  hardware/            CPU architecture detection
  fileutil/            File utilities
  logging/             Unified logging
  version/             Build version
  versions/            Emulator version management (versions.toml)
  steam/               Steam integration
  cleanup/             Cleanup utilities
  testutil/            Test utilities
ui/                    Electron + React frontend (TypeScript)
  electron/            Electron main process (main.ts, preload.ts, updater.ts)
  src/                 React app
    components/        UI components (SystemCard, SyncView, Settings, etc.)
    lib/               Shared utilities, hooks, daemon communication
    types/             TypeScript types (generated + hand-written)
    assets/            SVG logos for systems and emulators
  e2e/                 Playwright E2E tests
relay/                 Sync relay server (separate Go module)
site/                  Documentation site (Astro + Starlight)
  src/content/docs/    MDX documentation pages
scripts/               Build and test scripts
test/e2e/              CLI E2E tests
protocol/              JSON schema for daemon protocol
```

## Development commands

Uses `just` as task runner. Key recipes:

```
just check             Run all checks: lint + test + typecheck + site fmt
just test              Run Go tests (main + relay)
just lint              Run golangci-lint
just fmt               Format all code (Go, TypeScript, site markdown)
just dev               Run the Electron app in development mode
just build             Build AppImage (dev version)
just generate-types    Generate TypeScript types from Go via tygo
just e2e               Run CLI E2E tests in container (podman)
just ui-e2e            Run Playwright UI E2E tests
just site-dev          Run documentation site locally
just site-build        Build documentation site
just relay-dev         Run relay server locally
just relay-test        Run relay tests
just clean             Clean build artifacts
```

Run `just --list` for the full list.

## Languages and tooling

Go 1.24+ for backend and CLI. TypeScript for UI (Electron 40, React 19,
Vite 7, Tailwind CSS 4). Astro + Starlight for the documentation site.

- Go linting: golangci-lint v2 (`.golangci.yml`), linters: errcheck,
  forbidigo, govet, ineffassign, staticcheck, unused
- Go formatting: gofmt + goimports (local prefix: `github.com/fnune/kyaraben`)
- TypeScript linting/formatting: Biome (`ui/biome.json`), single quotes,
  no semicolons, 2-space indent, 100-char line width
- Site formatting: Prettier for MDX/MD (`site/.prettierrc`)
- Pre-commit hooks: `.pre-commit-config.yaml` (gofmt, golangci-lint,
  nixpkgs-fmt, biome, prettier)

## Type generation

Go types in `internal/daemon/types.go` and `internal/daemon/protocol.go`
are the source of truth. TypeScript types are generated via tygo
(`tygo.yaml`) into `ui/src/types/*.gen.ts`. Run `just generate-types`
after modifying Go protocol types.

## CI

GitHub Actions (`.github/workflows/ci.yml`) runs on push/PR to main:

- `go-build`: build + test with race detector
- `go-lint`: golangci-lint
- `relay-build`: relay build + test + lint
- `ui-lint`: Biome check
- `site-build`: prettier check + astro build
- `electron-e2e`: full AppImage build + Playwright tests

## Testing

Prefer fakes over mocks. Fakes are working implementations with shortcuts
(in-memory store instead of real database). Mocks couple tests to
implementation details.

- Go unit tests: `go test ./...`
- Go integration tests: use fakes via dependency injection
- UI unit tests: Vitest with React Testing Library (`cd ui && npm run test`)
- UI E2E tests: Playwright (`just ui-e2e`), use `getByRole`, `getByLabel`,
  `getByText` selectors
- CLI E2E tests: podman container (`just e2e`)

The `forbidigo` linter enforces that tests use injected dependencies (fakes)
instead of production constructors like `vfs.OSFS`, `store.NewDefaultUserStore`,
etc.

## Architecture

### Daemon protocol

The UI (Electron renderer) communicates with the Go backend through a daemon
process. The daemon reads JSON commands from stdin and writes JSON events to
stdout. Command types: `apply`, `status`, `doctor`, `get_config`,
`set_config`, `get_systems`, `preflight`, `sync_*`, `uninstall`, etc.
Event types: `ready`, `result`, `progress`, `error`, `cancelled`.

### Sidecar build

The Go CLI is compiled as an Electron sidecar binary
(`scripts/build-sidecar.sh`). It is placed in `ui/binaries/` with a
target-triple suffix and bundled into the AppImage.

### Config management

Kyaraben manages specific keys in emulator config files (paths, controllers,
hotkeys). User changes to unmanaged keys are preserved. Three-way merge
detects conflicts. Apply is atomic and idempotent.

Config file formats are handled by `internal/configformat/` (INI, TOML, XML,
YAML). Each emulator implements `ConfigGenerator` to produce config patches,
symlink specs, and controller bindings.

### Storage

Two distinct storage areas:

- KyarabenState (XDG): `~/.config/kyaraben/` (config), `~/.local/state/kyaraben/` (binaries, manifest)
- UserStore: `~/Emulation/` (ROMs, BIOS, saves, states, screenshots)

Saves are organized by system (shared across emulators). States and
screenshots are organized by emulator (format-specific).

## Key conventions

- Make code self-evident. Comments explaining "what" are banned. Comments
  explaining "why" are acceptable as a last resort if the code cannot be
  made clearer.
- Use sentence-case in headings. No bold text. No em-dashes. No emoji.
- Pass dependencies explicitly. No hidden instantiation. Follow SOLID
  principles. Expensive instantiation happens at the composition root
  (`cmd/kyaraben/main.go`).
- Define dependencies as interfaces where substitution is needed. Swap
  real implementations for fakes at construction time.
- Use `var log = logging.New("package")` for Go logging.
- TypeScript: no default exports (enforced by Biome), no barrel files,
  no non-null assertions.
- Use backticks for domain concepts: `System`, `Emulator`, `Provision`,
  `UserStore`, `Manifest`.

## Commit messages

Imperative mood title, body explaining what and why, then test plan:

```
Brief, actionable description

What changed and why. Use paragraphs, lists, or both.

## Test plan

Reproducible verification steps.
```

No trailing periods on list items. Use backticks for code references.

## Before committing

Run `just check`. This runs golangci-lint, Go tests, TypeScript typecheck,
Biome lint, and site formatting check. All must pass.
