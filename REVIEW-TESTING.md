# Testing review

Review of test coverage and quality against CONTRIBUTING.md and TASKS.md requirements.

## Test files inventory

| File | Type | Tests |
|------|------|-------|
| `test/e2e/cli_test.go` | E2E | 8 tests (CLI, no Nix builds) |
| `internal/model/config_test.go` | Unit | Config parsing |
| `internal/model/manifest_test.go` | Unit | Manifest operations |
| `internal/emulators/registry_test.go` | Unit | Registry operations |
| `internal/store/user_store_test.go` | Unit | UserStore operations |
| `internal/store/provision_test.go` | Unit | Provision checking |
| `internal/doctor/doctor_test.go` | Unit | Doctor logic |
| `internal/status/status_test.go` | Unit | Status logic |
| `ui/e2e/app.spec.ts` | E2E | UI tests (Playwright, includes real Nix build) |

## TASKS.md test requirements

### Unit tests

| Requirement | Status | Notes |
|-------------|--------|-------|
| Model types (serialization, validation) | Partial | `config_test.go`, `manifest_test.go` exist |
| `UserStore` path generation | Implemented | `user_store_test.go` |
| `EmulatorConfig` generation (RetroArch, DuckStation) | Missing | No tests for config generators |
| `Provision` hash verification | Implemented | `provision_test.go` |
| `Manifest` read/write | Implemented | `manifest_test.go` |

### Integration tests

| Requirement | Status | Notes |
|-------------|--------|-------|
| CLI commands with mock Nix | Missing | E2E tests use real binary but skip Nix |
| `KyarabenConfig` parsing and validation | Partial | Basic parsing tested |
| Flake generation | Missing | No tests for `nix.FlakeGenerator` |

### E2E test harness

| Requirement | Status | Notes |
|-------------|--------|-------|
| Research open-licensed emulators | Implemented | TIC-80 used |
| Package test emulator in Nix | Not needed | TIC-80 in nixpkgs |
| Test fixture format | Missing | No fixture system |
| Harness: isolated `UserStore` and `KyarabenState` | Partial | Uses `t.TempDir()` |
| Harness: run commands | Implemented | `kyarabenCmd()` helper |
| Harness: assert on filesystem state | Partial | Basic file existence checks |

### Test infrastructure

| Requirement | Status | Notes |
|-------------|--------|-------|
| CI pipeline | Implemented | GitHub Actions |
| Test helpers for temp directories | Partial | Uses `t.TempDir()` |
| Fake Nix client | Missing | Would enable faster Go unit/integration tests |
| E2E with actual Nix | Implemented | Playwright tests use `e2e-test` system with real Nix |

## E2E test analysis

### `test/e2e/cli_test.go`

Tests cover:
1. `TestCLIInit` - basic init
2. `TestCLIStatus` - status display
3. `TestCLIDoctor` - PSX missing BIOS
4. `TestCLIDoctorTIC80` - TIC-80 no provisions
5. `TestCLIApplyDryRun` - dry run apply
6. `TestCLIHelp` - help output
7. `TestCLIInitForce` - force overwrite
8. `TestCLIUninstall` - uninstall

Missing E2E tests:
1. Actual apply (with Nix build)
2. Config diff display
3. Daemon protocol
4. Multiple systems
5. Invalid config handling
6. Emulator config generation verification

### Test isolation

Tests use `t.TempDir()` for config and user store paths. However, tests do not isolate:
- XDG directories (uses real user directories)
- Nix store paths (would use real store if Nix ran)

The `-c` flag allows custom config path, but other XDG paths are computed from environment variables that tests don't override.

## Unit test analysis

### `internal/model/config_test.go`

Tests:
- `TestLoadConfig` - load from file
- `TestSaveConfig` - save to file
- `TestExpandUserStore` - tilde expansion
- `TestEnabledSystems` - list enabled systems

Missing:
- Invalid TOML handling
- Missing file handling (tested in other places)
- Large config handling
- Unicode paths

### `internal/model/manifest_test.go`

Tests:
- `TestNewManifest` - creation
- `TestManifestSaveLoad` - round-trip
- `TestManifestAddEmulator` - add emulator
- `TestManifestAddManagedConfig` - add config
- `TestManifestAddManagedConfigUpdate` - update existing

Good coverage for manifest operations.

### `internal/emulators/registry_test.go`

Tests:
- `TestGetSystem` - get by ID
- `TestGetEmulator` - get by ID
- `TestGetDefaultEmulator` - defaults
- `TestGetEmulatorsForSystem` - system lookup

Missing:
- Unknown system/emulator handling
- Hidden system filtering

### `internal/store/user_store_test.go`

Tests:
- `TestUserStoreInitialize` - directory creation
- `TestUserStoreInitializeSystem` - system subdirs
- `TestUserStorePaths` - path getters

Good coverage.

### `internal/store/provision_test.go`

Tests:
- `TestProvisionCheckerFound` - file exists with correct hash
- `TestProvisionCheckerMissing` - file missing
- `TestProvisionCheckerInvalid` - wrong hash
- `TestProvisionCheckerCaseInsensitive` - uppercase filenames

Good coverage including edge case of case-insensitive matching.

### `internal/doctor/doctor_test.go`

Tests:
- `TestDoctorRun` - basic doctor run

Minimal coverage. Missing:
- Multiple systems
- Mixed results (some found, some missing)
- No provisions case

### `internal/status/status_test.go`

Tests:
- `TestStatusGet` - basic status

Minimal coverage. Missing:
- No config file
- No manifest
- Multiple emulators

## Critical testing gaps

### 1. No fake Nix client

CONTRIBUTING.md requires:
> Fake Nix client for unit/integration tests (working impl, no actual builds)

Without this, testing apply logic requires either:
- Skipping Nix (current approach with dry-run)
- Running actual Nix builds (slow, requires Nix)

The `nix.Client` struct should implement an interface to allow substitution.

### 2. No config generator tests

The config generators (`RetroArchConfig`, `DuckStationConfig`) have no unit tests. These are critical for correctness since they produce files that emulators read.

### 3. No flake generator tests

`nix.FlakeGenerator.Generate()` is untested. The generated flake syntax could be invalid.

### 4. No daemon tests

The daemon command handlers are untested. Protocol correctness is unverified.

### 5. Go E2E tests skip Nix builds

The Go CLI tests in `test/e2e/cli_test.go` avoid Nix builds (using `--dry-run` or testing TIC-80).

However, the Playwright UI tests (`ui/e2e/app.spec.ts`) do include a real Nix build test using the `e2e-test` system with a 120-second timeout. This validates the full UI → daemon → Nix flow.

### 6. UI E2E tests cover daemon integration

The Playwright tests in `ui/e2e/app.spec.ts` do test the UI → Go daemon integration. Tests call `status`, `doctor`, and `apply` through the UI, which invokes the daemon via IPC. The apply test with `e2e-test` system validates the full flow including Nix builds.

## Test quality observations

### Good practices

1. Tests use `t.TempDir()` for cleanup
2. Tests use `t.Helper()` in helper functions
3. Tests check both positive and negative cases
4. Provision tests include case-insensitivity

### Issues

1. No table-driven tests where appropriate (e.g., registry tests)
2. Minimal error message verification
3. No benchmarks
4. No race detection in CI (should run `go test -race`)

## Recommendations

### High priority

1. Create `NixClient` interface and fake implementation
2. Add config generator unit tests
3. Add flake generator tests

### Medium priority

1. Add daemon protocol tests (unit tests, not just via UI)
2. Improve test isolation for XDG paths
3. Convert to table-driven tests where appropriate
4. Add `go test -race` to CI

### Low priority

1. Add benchmarks for config parsing
2. Add fuzz tests for protocol parsing

## E2E test consolidation

Currently there are two E2E test suites with different capabilities:

| Aspect | Go CLI tests | Playwright UI tests |
|--------|--------------|---------------------|
| Location | `test/e2e/cli_test.go` | `ui/e2e/app.spec.ts` |
| Tests CLI | Yes | No (only via daemon) |
| Tests UI | No | Yes |
| Tests daemon protocol | No | Indirectly |
| Real Nix builds | No (`--dry-run`) | Yes (`e2e-test` system) |

### Options

1. Add real Nix test to Go CLI tests: simplest option, add a test that runs `kyaraben apply` with `e2e-test` system (no `--dry-run`). Keeps test suites separate but both cover Nix.

2. Consolidate into Playwright only: Playwright can spawn CLI processes via Node's `child_process`. This would test both UI and CLI from one suite. Downside: adds Node dependency to CLI testing, Playwright is heavyweight for CLI tests.

3. Consolidate into Go only: Go tests could spawn Electron and interact with it. This is complex and fights against Playwright's strengths.

4. Shared test fixtures: keep both suites but share configuration (temp directories, config files, expected outputs). A `testdata/` directory or shared setup scripts could reduce duplication.

### Recommendation

Option 1 is the pragmatic choice: add one real Nix test to `test/e2e/cli_test.go` using the `e2e-test` system. This gives:
- Go tests: CLI behavior, flags, output formatting, and one real Nix build
- Playwright tests: UI behavior, daemon IPC, and real Nix build

Both suites serve their purpose without forced consolidation. The `e2e-test` system with the `hello` package builds quickly, so the added time is minimal.

## CI considerations

Nix builds are slow on cold caches. When updating tests that involve Nix:

1. Ensure GitHub Actions caches are updated for nix-portable store
2. The `e2e-test` system uses nixpkgs `hello` which is usually cached, but first runs on new cache keys will be slow
3. Consider cache key strategies that balance freshness with hit rate (e.g., weekly rotation vs. per-commit)

Current CI caches (from recent commits):
- nix-portable store is cached
- Cache permissions were fixed in commit 9e56f36

When adding new emulators or changing nixpkgs pins, CI times will spike until caches warm up. Document expected CI times in PR descriptions when making such changes.

### Docker layer caching

Both `Containerfile.electron-e2e` and `Containerfile.electron-build` copy the entire project before installing dependencies:

```dockerfile
COPY . /build
RUN ./scripts/build-sidecar.sh
RUN cd ui && npm ci
```

This invalidates the npm/Go dependency cache on any file change (even README edits). Better structure:

```dockerfile
# 1. Copy dependency manifests first
COPY go.mod go.sum ./
COPY ui/package.json ui/package-lock.json ./ui/

# 2. Install dependencies (cached unless manifests change)
RUN go mod download
RUN cd ui && npm ci

# 3. Copy source code (changes frequently)
COPY . .

# 4. Build
RUN go build ./...
RUN cd ui && npm run build
```

This keeps dependency installation in earlier layers that only rebuild when `go.mod`, `go.sum`, `package.json`, or `package-lock.json` change. Code changes only invalidate the final build layers.

The `scripts/build-sidecar.sh` step also downloads nix-portable on each build. Consider caching that binary or including it in a base image.

## Test coverage estimation

Based on file analysis (no actual coverage tool run):

| Package | Estimated coverage |
|---------|-------------------|
| model | 60% |
| store | 70% |
| emulators | 20% |
| nix | 0% |
| daemon | 0% |
| cli | 40% (via E2E) |
| doctor | 30% |
| status | 30% |
| apply | 10% (via E2E dry-run) |

Overall estimated coverage: 30-40%

Critical untested paths:
- Config file writing (unit tests)
- Flake generation (unit tests)
- Daemon protocol handling (unit tests, though covered via Playwright)
- Error recovery

Note: Actual Nix builds are tested via Playwright UI tests with the `e2e-test` system.
