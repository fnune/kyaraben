# Domain model review

Review of MODEL.md against actual implementation.

## Entity implementation status

### System

MODEL.md specification:
```
System {
    id id
}
```

Actual implementation (`internal/model/system.go`):
```go
type System struct {
    ID          SystemID
    Name        string
    Description string
    Hidden      bool
}
```

Status: matches specification with reasonable additions (Name, Description for UI, Hidden for test systems).

### Emulator

MODEL.md specification:
```
Emulator {
    id id
    PackageSource source
}
```

Actual implementation (`internal/model/emulator.go`):
```go
type Emulator struct {
    ID          EmulatorID
    Name        string
    Systems     []SystemID
    Source      PackageSource
    NixAttr     string
    Provisions  []Provision
    StateKinds  []StateKind
    ConfigPaths []string
}
```

Status: matches with additions. The `NixAttr` field is an implementation detail that arguably leaks into the domain model, but is pragmatic.

### Provision

MODEL.md specification:
```
Provision {
    id id
    ProvisionKind kind
    CheckMethod check
}
```

Actual implementation (`internal/model/provision.go`):
```go
type Provision struct {
    ID          string
    Kind        ProvisionKind
    Filename    string
    Description string
    Required    bool
    MD5Hash     string
    SHA256Hash  string
}
```

Status: matches conceptually. `CheckMethod` became concrete hash fields (MD5Hash, SHA256Hash). This is reasonable since hash verification is the only check method implemented.

Note: SHA256Hash is defined but not yet used. Only MD5 verification is implemented in `internal/store/provision.go:64-75`. MD5 is the common format for BIOS hashes found online, so this is practical for MVP. SHA256 support exists for future use when stronger verification is needed.

### State

MODEL.md specification:
```
State {
    StateKind kind
    Path location
    SyncStrategy sync
}
```

Actual implementation (`internal/model/state.go`):
```go
type State struct {
    Kind StateKind
    Path string
    Sync SyncStrategy
}
```

Status: matches specification exactly.

Issue: State is never instantiated. The type exists, StateKind and SyncStrategy enums exist, but no code creates State values. Emulators define `StateKinds []StateKind` but not actual State instances with paths.

### EmulatorConfig

MODEL.md specification:
```
EmulatorConfig {
    Path path
    ConfigFormat format
}
```

Actual implementation (`internal/model/emulator_config.go`):
```go
type EmulatorConfig struct {
    Path       string
    Format     ConfigFormat
    EmulatorID EmulatorID
}
```

Status: matches with addition of EmulatorID for tracking.

### KyarabenConfig

MODEL.md specification:
```
KyarabenConfig {
    Path userStore
}
```

Actual implementation (`internal/model/config.go`):
```go
type KyarabenConfig struct {
    Global  GlobalConfig
    Systems map[SystemID]SystemConf
}

type GlobalConfig struct {
    UserStore string
}

type SystemConf struct {
    Emulator EmulatorID
}
```

Status: matches. The nested structure is an implementation detail for TOML serialization.

### Synchronizer

MODEL.md specification:
```
Synchronizer {
    SyncBackend backend
}
```

Actual implementation: not implemented.

Status: correctly out of scope for MVP as stated in MVP.md.

### KyarabenState

MODEL.md specification:
```
KyarabenState {
    Path root
}
```

Actual implementation: no unified KyarabenState type exists. XDG paths are computed via `internal/paths/xdg.go` functions.

Status: partially implemented. The concept exists through path functions but there is no unified state object.

### UserStore

MODEL.md specification:
```
UserStore {
    Path root
}
```

Actual implementation (`internal/store/user_store.go`):
```go
type UserStore struct {
    Root string
}
```

Status: matches exactly.

### Manifest

MODEL.md specification:
```
Manifest {
    int version
}
```

Actual implementation (`internal/model/manifest.go`):
```go
type Manifest struct {
    Version            int
    LastApplied        time.Time
    InstalledEmulators map[EmulatorID]InstalledEmulator
    ManagedConfigs     []ManagedConfig
}
```

Status: matches with reasonable extensions.

## Relationship verification

### System can be emulated by many Emulators

Implemented via `Emulator.Systems []SystemID` and `Registry.GetEmulatorsForSystem()`.

### Emulator targets one or more Systems

Implemented via `Emulator.Systems []SystemID`.

### Emulator may need Provisions

Implemented via `Emulator.Provisions []Provision`.

### Emulator produces State

Partially implemented. `Emulator.StateKinds []StateKind` exists but no actual State instances are created or tracked.

### Emulator is configured via EmulatorConfig

Implemented via config generators and `Emulator.ConfigPaths []string`.

### KyarabenConfig enables Systems

Implemented via `KyarabenConfig.Systems map[SystemID]SystemConf`.

### KyarabenConfig specifies emulator per system

Implemented via `SystemConf.Emulator EmulatorID`.

### KyarabenConfig points to UserStore

Implemented via `KyarabenConfig.Global.UserStore`.

### KyarabenConfig configures Synchronizer

Not implemented (Synchronizer out of scope).

### KyarabenState contains KyarabenConfig and Manifest

No unified KyarabenState type. Config and Manifest are loaded independently.

### UserStore holds State

Implemented via directory structure, but State instances are not created.

### Manifest manages EmulatorConfigs with base snapshots

Partially implemented. ManagedConfig exists but `BaselineHash` is never populated.

### Manifest tracks installed Emulators with versions

Partially implemented. InstalledEmulator exists but version is always "latest".

### Synchronizer syncs UserStore

Not implemented (out of scope).

## Model issues

### 1. State type prepared for sync feature

The `State` struct and related types are defined but not yet instantiated:
- `StateKind` enum: StateSaves, StateSavestates, StateScreenshots, StateCache, StatePersistent
- `SyncStrategy` enum: SyncBidirectional, SyncSendOnly, SyncIgnore
- `State` struct with Kind, Path, Sync

Emulators define `StateKinds []StateKind` but this is currently informational. This is expected since the Synchronizer feature is out of scope for MVP. The types are in place for when sync is implemented.

### 2. PackageSource prepared for future systems

`PackageSource` enum defines `PackageSourceNixpkgs` and `PackageSourceGitHub`. Currently only nixpkgs is used, which is expected for the MVP pilot systems (SNES, PSX, TIC-80). The GitHub source exists for future systems like Switch emulators (Eden) mentioned in PITCH.md.

### 3. Missing KyarabenState abstraction

MODEL.md describes KyarabenState as containing both config and manifest, following XDG conventions. Implementation has:
- `paths.KyarabenConfigDir()` for config
- `paths.KyarabenDataDir()` for Nix data
- `paths.KyarabenStateDir()` for manifest

But no unified type that encapsulates all of these.

### 4. Provision model assumes file-based checking

The `Provision` struct has hash fields baked in:

```go
type Provision struct {
    ID          string
    Kind        ProvisionKind
    Filename    string
    MD5Hash     string
    SHA256Hash  string
}
```

This assumes provisions are always files verified by hash. But provisions could be:
- A file with a specific hash (current)
- A file that just needs to exist (no hash check)
- A directory structure
- A running service or daemon
- Hardware (e.g., a controller, GPU capability)
- An environment variable or system configuration

A more flexible design would separate the provision identity from its verification strategy:

```go
type Provision struct {
    ID          string
    Kind        ProvisionKind
    Description string
    Required    bool
    Check       ProvisionCheck  // interface or sum type
}

type FileHashCheck struct {
    Filename string
    MD5      string
    SHA256   string
}

type FileExistsCheck struct {
    Filename string
}

type DirectoryCheck struct {
    Path     string
    Contains []string
}
```

For MVP scope this is over-engineering, but the current design will need refactoring if non-file provisions are added.

### 5. Emulator.NixAttr leaks installation mechanism

The `Emulator` struct includes `NixAttr`:

```go
type Emulator struct {
    ID          EmulatorID
    Name        string
    Systems     []SystemID
    Source      PackageSource
    NixAttr     string  // leaks Nix into domain model
    // ...
}
```

This assumes emulators are installed via Nix. But emulators could be:
- From nixpkgs (current)
- From a GitHub flake
- Downloaded as a binary/AppImage
- Built from source
- Already installed on the system (user-managed)

The `PackageSource` enum hints at this flexibility but `NixAttr` contradicts it. Better design:

```go
type Emulator struct {
    ID       EmulatorID
    Name     string
    Systems  []SystemID
    Install  InstallMethod  // interface or sum type
    // ...
}

type NixpkgsInstall struct {
    Attr string
}

type FlakeInstall struct {
    URL   string
    Attr  string
}

type BinaryDownload struct {
    URL      string
    Checksum string
}

type UserManaged struct {
    BinaryPath string
}
```

This keeps the domain model agnostic to installation mechanism. The `PackageSource` enum could become a discriminator for the `InstallMethod` sum type.

### 6. Provision checking is fragmented

MODEL.md states: "Provision... Can be checked: is it satisfied?"

Implementation has:
- `model.Provision` with hash fields
- `store.ProvisionChecker` that does the checking
- `model.ProvisionResult` for results
- `model.ProvisionStatus` enum

The checking logic is in store package, not model. This is arguably correct (model should be pure), but the relationship is not obvious.

### 5. ConfigPatch is not in MODEL.md

`ConfigPatch` is a key type for config generation but is not documented in MODEL.md:
```go
type ConfigPatch struct {
    Config  EmulatorConfig
    Entries []ConfigEntry
}
```

### 6. Registry is not in MODEL.md

`Registry` is central to the system but not documented:
- Maps SystemID to System
- Maps EmulatorID to Emulator
- Provides GetDefaultEmulator
- Provides GetConfigGenerator

This is an important domain service that should be documented.

## Open questions from MODEL.md status

### 1. Provision kinds taxonomy

> What's the full taxonomy? Is firmware distinct from BIOS?

Implementation has: ProvisionBIOS, ProvisionKeys, ProvisionFirmware

Currently only ProvisionBIOS is used by the MVP pilot systems. ProvisionKeys and ProvisionFirmware are prepared for future systems (e.g., Switch emulators need prod.keys).

### 2. Emulator-native state paths to UserStore

> How does kyaraben map emulator-native state paths to the unified UserStore layout?

Resolved: config generators set emulator paths to point to UserStore subdirectories. See `internal/emulators/retroarch.go` and `duckstation.go`.

### 3. UserStore structure configurability

> Should UserStore structure be configurable or opinionated?

Implementation is opinionated with fixed structure. This matches FILESYSTEM.md recommendation.
