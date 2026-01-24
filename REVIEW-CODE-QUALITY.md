# Code quality review

Review of codebase against CONTRIBUTING.md standards.

## Languages and typing

CONTRIBUTING.md:
> We use Go for the backend and CLI... The UI is TypeScript. Both languages offer static typing, which we rely on heavily.

Status: followed. Go code uses proper types. TypeScript code uses interfaces.

CONTRIBUTING.md:
> JSON schemas serve as the source of truth, with types generated for both Go and TypeScript.

Status: not followed, but the stated approach in CONTRIBUTING.md deserves reconsideration.

Go is the primary language and the daemon protocol types are defined in `internal/daemon/types.go`. Using Go types as the source of truth is more practical:
- Go is where the domain model lives
- TypeScript types can be generated from Go (tools like `tygo` or `go-typescript` exist)
- JSON Schema can be generated from Go for documentation or runtime validation

Current state:
- `internal/daemon/types.go` defines protocol types in Go
- `ui/electron/main.ts` has inline TypeScript interfaces (manually duplicated)
- `protocol/schema.json` exists but is disconnected from both

Recommendation: update CONTRIBUTING.md to specify Go as source of truth, then set up codegen from Go to TypeScript. The JSON schema can remain as generated documentation.

## Testing

CONTRIBUTING.md:
> We follow Martin Fowler's distinction between fakes and mocks. Prefer fakes.

Status: partially followed. No mocks are used. However, no fakes exist either. The Nix client has no fake implementation for testing.

CONTRIBUTING.md:
> Test harnesses matter more than individual test cases.

Status: partially followed. E2E tests have a basic harness (`kyarabenCmd`, `projectRoot` helpers in `test/e2e/cli_test.go`). But there is no isolated environment setup beyond `t.TempDir()`.

CONTRIBUTING.md:
> E2E tests invoke the real system, including actual Nix builds.

Status: partially followed. Go CLI E2E tests (`test/e2e/cli_test.go`) avoid Nix builds using `--dry-run`. However, Playwright UI tests (`ui/e2e/app.spec.ts`) include a real Nix build test using the `e2e-test` system.

## Simplicity

CONTRIBUTING.md:
> Start with the simplest solution that works.

Status: generally followed. The codebase is straightforward.

CONTRIBUTING.md:
> When a dependency introduces breaking changes, treat the new version as a new entity.

Status: followed. MODEL.md documents this approach for emulators.

CONTRIBUTING.md:
> Explicit configuration is better than implicit defaults.

Status: partially followed. Systems require explicit emulator choice. However, `GetDefaultEmulator()` exists and is used when creating initial configs.

## Domain modeling

CONTRIBUTING.md:
> Name things precisely using the language of the domain.

Status: followed. Terms like System, Emulator, Provision, UserStore are used consistently.

CONTRIBUTING.md:
> When writing documentation, use backticks when referring to domain concepts.

Status: followed in markdown files.

CONTRIBUTING.md:
> Keep the domain model clean. Implementation details such as serialization formats or database schemas should not leak into the model.

Status: partially followed. Issues:

1. `Emulator.NixAttr` is a Nix implementation detail in the domain model
2. TOML tags on `KyarabenConfig` fields
3. JSON tags on `Manifest` fields

Serialization tags are pragmatic but violate the strict interpretation.

## Dependency management

CONTRIBUTING.md:
> Pass dependencies explicitly. There should be no hidden instantiation deep in the call stack.

Status: followed. Dependencies are passed via constructors:
- `daemon.New(configPath, registry, nixClient, flakeGenerator, configWriter)`
- `apply.Applier` struct with explicit fields

CONTRIBUTING.md:
> Define dependencies as interfaces where substitution is needed.

Status: partially followed. Only `ConfigGenerator` is defined as an interface. `nix.Client` is a concrete type with no interface, making it impossible to substitute for testing.

## Code style

CONTRIBUTING.md:
> Make code self-evident. Do not write comments explaining what code does.

Status: violated in multiple places.

Comments that should be removed or addressed by clearer code:

`internal/nix/client.go`:
```go
// nix-portable has nix-command and flakes enabled by default
func (c *Client) runNix(ctx context.Context, args []string) (*exec.Cmd, error) {
```

`internal/emulators/config_writer.go`:
```go
// Existing values are preserved unless kyaraben sets them.
func (w *ConfigWriter) Apply(patch model.ConfigPatch) error {
```

`internal/model/manifest.go`:
```go
// Manifest tracks what kyaraben has installed and configured.
type Manifest struct {
```

`internal/model/config.go`:
```go
// KyarabenConfig represents the user's kyaraben configuration.
type KyarabenConfig struct {
```

Many of these are doc comments which are reasonable for exported types, but some are explanatory comments that the code should make self-evident.

### Debug logging to stderr

The nix client outputs extensive debug information to stderr:
```go
fmt.Fprintf(os.Stderr, "[kyaraben-go] Looking for nix-portable: %s\n", binaryName)
fmt.Fprintf(os.Stderr, "[kyaraben-go] Executable dir: %s\n", execDir)
fmt.Fprintf(os.Stderr, "[kyaraben-go] Checking: %s\n", path)
```

This is appropriate for debugging but should use a proper logging abstraction that can be disabled. The `internal/logging/logger.go` exists but is not used here.

### Config file comments

Generated config files include comments:
```go
_, _ = fmt.Fprintln(f, "# Configuration managed by kyaraben")
_, _ = fmt.Fprintln(f, "# Manual changes will be preserved on next apply")
```

These are comments in output files, not source code, so are acceptable.

## Specific code issues

### 1. Non-deterministic map iteration

`internal/emulators/config_writer.go:73-74`:
```go
for key, value := range existing {
    _, _ = fmt.Fprintf(f, "%s = %s\n", key, value)
}
```

Map iteration order is non-deterministic in Go. This causes config files to have different key orderings on each write, making diffs noisy and version control harder.

Same issue in `config_writer.go:133-139` for INI sections.

### 2. Error handling ignores close errors

`internal/emulators/config_writer.go:54`:
```go
_ = data.Close()
```

`internal/emulators/config_writer.go:113`:
```go
_ = data.Close()
```

While file read close errors are rarely actionable, explicit ignore with `_` is fine. However, write close errors should not be ignored (`config_writer.go:67`).

### 3. Hardcoded version string

`internal/apply/apply.go:125`:
```go
manifest.AddEmulator(model.InstalledEmulator{
    ID:        emuID,
    Version:   "latest",
```

Version should be extracted from Nix store path or queried from the emulator.

### 4. Context timeout hardcoded

`internal/apply/apply.go:95`:
```go
buildCtx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
```

Timeout should be configurable or at least a named constant.

### 5. Missing error context in daemon

`internal/daemon/daemon.go:73-78`:
```go
cfg, err = model.NewDefaultConfig()
if err != nil {
    return []Event{{
        Type: EventError,
        Data: map[string]string{"error": err.Error()},
    }}
}
```

Raw `err.Error()` loses stack context. Consider wrapping errors.

### 6. Type assertions without checks

`ui/electron/main.ts:169`:
```go
reject(new Error((event.data as { error?: string })?.error || 'Unknown error'))
```

Type assertion could fail. Should validate shape before casting.

### 7. Inconsistent error wrapping

Some errors are wrapped with context:
```go
return nil, fmt.Errorf("reading manifest: %w", err)
```

Others are not:
```go
return nil, err
```

## Positive observations

### Clean package structure

The `internal/` layout follows Go conventions with clear separation:
- `model/` for domain types
- `cli/` for commands
- `daemon/` for UI communication
- `emulators/` for emulator-specific logic
- `nix/` for Nix integration
- `store/` for user data management

### Explicit dependency injection

The composition root in `cmd/kyaraben/main.go` and `internal/cli/context.go` follows the prescribed pattern.

### No global state

No package-level mutable variables or singletons.

### Proper error wrapping

Most errors include context via `fmt.Errorf("context: %w", err)`.

## Summary of violations

| Standard | Severity | Location |
|----------|----------|----------|
| TypeScript types manually duplicated | Medium | `ui/electron/main.ts` should be generated from Go |
| No Nix client interface/fake | High | `internal/nix/client.go` |
| Non-deterministic map iteration | Medium | `internal/emulators/config_writer.go` |
| Hardcoded version "latest" | Medium | `internal/apply/apply.go:125` |
| Debug logging without level control | Low | `internal/nix/client.go` |
| Explanatory comments | Low | Multiple files |
| Hardcoded timeout | Low | `internal/apply/apply.go:95` |

## Recommendation for CONTRIBUTING.md

The statement about JSON schemas being the source of truth should be updated. A more practical approach:

> Go types are the source of truth for protocols and data interchange. TypeScript types are generated from Go. JSON schemas may be generated for documentation or runtime validation.
