# Multi-Emulator Support Analysis

This document analyzes what changes would be needed for kyaraben to support multiple emulators per system, considering user needs, architectural implications, and various implementation options.

## Table of Contents

1. [Current State](#current-state)
2. [Why Multi-Emulator Support?](#why-multi-emulator-support)
3. [Design Options](#design-options)
4. [Recommended Approach](#recommended-approach)
5. [Detailed Change Analysis](#detailed-change-analysis)
6. [Migration Path](#migration-path)
7. [Open Questions](#open-questions)

---

## Current State

### The Single-Emulator Constraint

Kyaraben currently enforces a **one emulator per system** model at multiple layers:

```
┌─────────────────────────────────────────────────────────────────┐
│ config.toml                                                     │
│ ─────────────────────────────────────────────────────────────── │
│ [systems.psx]                                                   │
│ emulator = "duckstation"    ← Single choice                     │
│                                                                 │
│ [systems.switch]                                                │
│ emulator = "eden@v0.1.0"    ← Version pinning, still one        │
└─────────────────────────────────────────────────────────────────┘
```

This constraint flows through:

| Layer | Current Implementation |
|-------|------------------------|
| **Config Model** | `SystemConf.Emulator string` — one emulator |
| **Domain Model** | `SystemDefinition.DefaultEmulatorID()` — one default |
| **Registry** | `defaultEmulators map[SystemID]EmulatorID` — one per system |
| **Protocol** | `SetConfigRequest.Systems map[SystemID]EmulatorID` |
| **Frontend State** | `selections: Map<SystemID, EmulatorID>` |
| **Storage Paths** | Some paths use system ID, some use emulator ID |
| **Manifest** | Tracks emulators, not system-emulator pairs |

### What Works Well

- **Simplicity**: Users make one choice per system
- **Conflict avoidance**: No path collisions between emulators for same system
- **Clear mental model**: "PSX uses DuckStation"

---

## Why Multi-Emulator Support?

### User Perspective

From interviewing emulation communities and considering common workflows:

#### 1. **Compatibility Fallbacks**
> "Game X doesn't work in DuckStation but runs fine in Mednafen"

Users often need different emulators for different games. No emulator has 100% compatibility with all games for a system.

**Example scenarios:**
- PSX: DuckStation (primary) + RetroArch:mednafen (fallback for specific games)
- N64: RetroArch:mupen64plus (HLE) + Ares (LLE for accuracy)
- PS2: PCSX2 (software renderer) + PCSX2 (hardware renderer as separate config?)

#### 2. **Feature Differences**
> "I use Dolphin for netplay and Dolphin-Ishiiruka for texture packs"

Different emulator variants/forks offer different features:
- Netplay support
- Shader/enhancement support
- RetroAchievements integration
- VR support

#### 3. **Performance vs Accuracy Tradeoffs**
> "Use fast emulator for casual play, accurate one for 'authentic' runs"

- Fast: Lower accuracy, higher performance (good for weaker hardware)
- Accurate: Cycle-accurate, slower (good for specific games or verification)

#### 4. **Testing and Development**
> "I'm a ROM hacker and need to verify my patches work on multiple emulators"

Power users working on:
- ROM hacks
- Translations
- Homebrew development

### Non-Goals (Things Users Probably Don't Need)

- Running multiple emulators simultaneously for the same game
- Automatically selecting emulator based on game (too complex, game-specific)
- Per-game configuration storage (massive scope increase)

---

## Design Options

### Option A: Multiple Enabled Emulators, Manual Selection at Launch

**Concept**: Allow enabling multiple emulators per system. At game launch time, user selects which emulator to use.

```toml
# config.toml
[systems.psx]
emulators = ["duckstation", "retroarch:mednafen_psx"]
default = "duckstation"  # Used when launching without explicit choice
```

**User Experience**:
```
┌─────────────────────────────────────────────────────────────┐
│  Launch "Final Fantasy VII"                                 │
├─────────────────────────────────────────────────────────────┤
│  Select emulator:                                           │
│                                                             │
│  ● DuckStation (default)                                    │
│  ○ RetroArch (mednafen_psx)                                 │
│                                                             │
│  [Remember for this game]  [Launch]                         │
└─────────────────────────────────────────────────────────────┘
```

**Pros**:
- Simple to understand
- User stays in control
- Minimal storage path conflicts

**Cons**:
- Extra click to launch games
- Kyaraben doesn't own game launching (relies on external launchers)

---

### Option B: Primary + Secondary Emulators

**Concept**: Each system has one primary emulator and optional secondary emulators installed but not integrated into workflow.

```toml
[systems.psx]
primary = "duckstation"
secondary = ["retroarch:mednafen_psx"]  # Installed but user launches manually
```

**User Experience**:
- Primary emulator fully configured (paths, desktop files, etc.)
- Secondary emulators installed but user manages them separately
- Kyaraben configures paths for secondaries but doesn't create desktop files

**Pros**:
- Keeps the simple "one emulator per system" launch experience
- Power users can still access alternatives
- Lower complexity than full multi-emulator

**Cons**:
- Secondary emulators are second-class citizens
- User needs to know how to launch secondaries manually
- Unclear value proposition vs just installing emulator yourself

---

### Option C: First-Class Multi-Emulator (Recommended)

**Concept**: All enabled emulators are equal citizens. Each system-emulator pair is a distinct configuration unit.

```toml
[systems.psx.duckstation]
enabled = true
version = "v0.1-10655"

[systems.psx."retroarch:mednafen_psx"]
enabled = true
```

**Alternative flat syntax** (simpler TOML):
```toml
[[emulator]]
id = "duckstation"
systems = ["psx"]
version = "v0.1-10655"

[[emulator]]
id = "retroarch:mednafen_psx"
systems = ["psx", "saturn"]  # RetroArch cores often support multiple systems
```

**User Experience**:
- UI shows enabled emulators per system as cards/list
- Each emulator has its own desktop file: "PSX (DuckStation)", "PSX (Mednafen)"
- All emulators share the same UserStore paths (saves, bios, etc.)

**Pros**:
- Full flexibility
- Clear mental model: each emulator is independent
- Desktop integration for all emulators

**Cons**:
- More complex UI
- Potential confusion: "which emulator should I use?"
- Need to handle save compatibility between emulators (see Storage section)

---

### Option D: Per-Game Emulator Assignment

**Concept**: Store per-game preferences mapping games to preferred emulators.

```toml
[systems.psx]
emulators = ["duckstation", "retroarch:mednafen_psx"]
default = "duckstation"

[games."Final Fantasy VII"]
emulator = "duckstation"

[games."Vagrant Story"]
emulator = "retroarch:mednafen_psx"  # Better compatibility
```

**Pros**:
- Most powerful for compatibility workarounds
- "Set and forget" experience

**Cons**:
- **Massive scope increase**: kyaraben would need to understand ROMs/games
- Game identification is hard (checksums? filenames? both?)
- Database maintenance burden
- Outside current project scope

**Recommendation**: Consider this as a future extension, not initial implementation.

---

## Recommended Approach

### Option C: First-Class Multi-Emulator

This provides the best balance of flexibility, clarity, and implementation complexity.

**Key Design Principles**:

1. **Emulator-Centric, Not System-Centric**
   - Current: "What emulator does PSX use?"
   - New: "Which emulators are enabled?" (each emulator declares its systems)

2. **Shared Storage, Isolated Configs**
   - All emulators share: `bios/`, `roms/`, `saves/` (when compatible)
   - Each emulator owns: its config files, states, cache

3. **All Enabled Emulators Are First-Class**
   - Desktop files for each
   - Full path configuration for each
   - Equal treatment in UI

4. **Per-Emulator Versioning**
   - Version pinning applies to emulator, not system
   - Different emulators can be on different versions

---

## Detailed Change Analysis

### 1. Domain Model Changes

#### Current Model

```go
// internal/model/definitions.go
type SystemDefinition interface {
    ID() SystemID
    Name() string
    DefaultEmulatorID() EmulatorID  // ← Single default
}
```

#### Proposed Model

```go
type SystemDefinition interface {
    ID() SystemID
    Name() string
    // Remove DefaultEmulatorID - no longer meaningful
}

// New: Emulator takes center stage
type Emulator struct {
    ID          EmulatorID
    Name        string
    Systems     []SystemID      // Systems this emulator supports
    Package     PackageRef
    Provisions  []Provision     // Per-system provisions
    StateKinds  []StateKind
    Launcher    LauncherInfo
}

// New: Configuration unit is the emulator, not system-emulator pair
type EnabledEmulator struct {
    ID      EmulatorID
    Version string           // Optional version pin
    // Systems derived from Emulator.Systems
}
```

**Impact**: Moderate. Emulator definitions already list supported systems; we just remove the reverse mapping.

---

### 2. Configuration Changes

#### Current Format

```toml
[global]
user_store = "~/Emulation"

[systems.psx]
emulator = "duckstation@v0.1-10655"

[systems.switch]
emulator = "eden"
```

#### Proposed Format

```toml
[global]
user_store = "~/Emulation"

# Emulators are the primary configuration unit
[emulators.duckstation]
enabled = true
version = "v0.1-10655"

[emulators.eden]
enabled = true
# version omitted = use default

[emulators."retroarch:mednafen_psx"]
enabled = true
# Also enables this for PSX (RetroArch core supports it)

[emulators."retroarch:mednafen_saturn"]
enabled = true
# Enables Saturn support via this core
```

**Alternative** (if we want to preserve system-centric view in config):

```toml
[emulators]
duckstation = { enabled = true, version = "v0.1-10655" }
eden = { enabled = true }
"retroarch:mednafen_psx" = { enabled = true }
```

#### Go Model Changes

```go
// Current
type KyarabenConfig struct {
    Global  GlobalConf
    Systems map[SystemID]SystemConf
    Sync    SyncConf
}

type SystemConf struct {
    Emulator string `toml:"emulator"`
}

// Proposed
type KyarabenConfig struct {
    Global    GlobalConf
    Emulators map[EmulatorID]EmulatorConf  // ← New primary structure
    Sync      SyncConf
}

type EmulatorConf struct {
    Enabled bool   `toml:"enabled"`
    Version string `toml:"version,omitempty"`
}
```

**Impact**: Breaking change to config format. Migration needed.

---

### 3. Registry Changes

#### Current Registry

```go
type Registry struct {
    systems          map[SystemID]System
    emulators        map[EmulatorID]Emulator
    defaultEmulators map[SystemID]EmulatorID  // ← Remove
}

func (r *Registry) GetDefaultEmulator(id SystemID) Emulator
func (r *Registry) GetEmulatorsForSystem(id SystemID) []Emulator
```

#### Proposed Registry

```go
type Registry struct {
    systems   map[SystemID]System
    emulators map[EmulatorID]Emulator
    // Remove defaultEmulators - not needed
}

// Keep existing
func (r *Registry) GetEmulatorsForSystem(id SystemID) []Emulator

// Remove or deprecate
// func (r *Registry) GetDefaultEmulator(id SystemID) Emulator

// Add: Given enabled emulators, which systems are covered?
func (r *Registry) GetCoveredSystems(enabled []EmulatorID) []SystemID
```

**Impact**: Low. Registry already stores all relationships.

---

### 4. Protocol/API Changes

#### Current Protocol

```go
// Get all systems and their emulator options
type GetSystemsResponse struct {
    Systems []SystemWithEmulators
}

type SystemWithEmulators struct {
    ID       SystemID
    Name     string
    Emulators []EmulatorRef  // Available options
}

// Set configuration: choose one emulator per system
type SetConfigRequest struct {
    UserStore string
    Systems   map[SystemID]EmulatorID  // ← One per system
}
```

#### Proposed Protocol

```go
// Option 1: Emulator-centric API (matches new config model)
type GetEmulatorsResponse struct {
    Emulators []EmulatorInfo
}

type EmulatorInfo struct {
    ID               EmulatorID
    Name             string
    Systems          []SystemID        // What this emulator supports
    DefaultVersion   string
    AvailableVersions []string
}

type SetConfigRequest struct {
    UserStore string
    Emulators map[EmulatorID]EmulatorConf  // ← Enable/version per emulator
}

// Option 2: Dual view (keep system view for UI convenience)
type GetSystemsResponse struct {
    Systems []SystemInfo
}

type SystemInfo struct {
    ID       SystemID
    Name     string
    Emulators []EmulatorRef          // Available emulators for this system
    Enabled   []EmulatorID           // ← NEW: Currently enabled for this system
}
```

**Recommendation**: Provide both views in API:
- Emulator-centric for configuration
- System-centric for UI display

**Impact**: Breaking API change. Frontend needs updates.

---

### 5. Frontend/UI Changes

#### Current UI Structure

```
┌─────────────────────────────────────────────────────────────────┐
│  Systems                                                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │
│  │ PlayStation │  │   Switch    │  │  GameCube   │              │
│  │             │  │             │  │             │              │
│  │ DuckStation▼│  │   Eden    ▼ │  │  Dolphin  ▼ │              │
│  │  [Enabled]  │  │  [Enabled]  │  │ [Disabled]  │              │
│  └─────────────┘  └─────────────┘  └─────────────┘              │
│                                                                 │
│  ▼ = Dropdown to select ONE emulator                            │
└─────────────────────────────────────────────────────────────────┘
```

#### Proposed UI Structure

**Option A: System-centric with multi-select**

```
┌─────────────────────────────────────────────────────────────────┐
│  Systems                                                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  PlayStation (PSX)                                              │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ ☑ DuckStation          v0.1-10655 ▼    [Configure]      │    │
│  │ ☑ RetroArch (mednafen)  latest    ▼    [Configure]      │    │
│  │ ☐ RetroArch (beetle)    latest    ▼    [Configure]      │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
│  Nintendo Switch                                                │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ ☑ Eden                  v0.1.0    ▼    [Configure]      │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Option B: Emulator-centric view**

```
┌─────────────────────────────────────────────────────────────────┐
│  Emulators                                                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ ☑ DuckStation                                           │    │
│  │   Systems: PlayStation                                  │    │
│  │   Version: v0.1-10655 ▼                                 │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ ☑ RetroArch                                             │    │
│  │   Cores:                                                │    │
│  │     ☑ mednafen_psx (PlayStation)                        │    │
│  │     ☐ beetle_psx   (PlayStation)                        │    │
│  │     ☑ mednafen_saturn (Saturn)                          │    │
│  │   Version: latest ▼                                     │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ ☑ Eden                                                  │    │
│  │   Systems: Nintendo Switch                              │    │
│  │   Version: v0.1.0 ▼                                     │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Recommendation**: Option A (System-centric with multi-select) for familiarity, with ability to expand to emulator details.

#### State Management Changes

```typescript
// Current
interface AppState {
    selections: Map<SystemID, EmulatorID>;  // One per system
}

// Proposed
interface AppState {
    enabledEmulators: Map<EmulatorID, EmulatorConfig>;
}

interface EmulatorConfig {
    enabled: boolean;
    version?: string;
}

// Derived from enabledEmulators + registry
function getEmulatorsForSystem(systemId: SystemID): EmulatorID[];
```

**Impact**: Significant UI refactor. Views need redesign.

---

### 6. Storage/Path Changes

#### Current Storage Model

```
UserStore/
├── roms/
│   └── {system}/          # e.g., psx/, switch/
├── bios/
│   └── {system}/          # System-specific BIOS
├── saves/
│   └── {system}/          # System-specific saves
├── states/
│   └── {emulator}/        # Emulator-specific savestates
├── screenshots/
│   └── {emulator}/
└── opaque/
    └── {emulator}/        # Emulator-managed directories
```

#### Multi-Emulator Storage Considerations

**The Good News**: Current structure already handles this well!

- `roms/{system}/` — Shared across emulators ✓
- `bios/{system}/` — Shared across emulators ✓
- `states/{emulator}/` — Already emulator-specific ✓
- `opaque/{emulator}/` — Already emulator-specific ✓

**The Challenge**: `saves/{system}/`

Different emulators may use different save formats:
- Memory card images (.mcr, .mcd, .srm)
- Raw SRAM dumps
- Proprietary formats

**Options for saves**:

1. **Keep shared saves (optimistic)**
   - Assume emulators for same system use compatible formats
   - Works for most cases (e.g., PSX memory cards are standardized)
   - Risk: format incompatibility corrupts saves

2. **Emulator-specific saves**
   ```
   saves/
   └── {emulator}/
       └── {system}/    # e.g., saves/duckstation/psx/
   ```
   - Safe but loses save sharing between emulators
   - User must manually copy saves to switch emulators

3. **Hybrid: shared by default, option to isolate**
   ```toml
   [emulators.duckstation]
   enabled = true
   isolated_saves = false  # Use saves/{system}/

   [emulators."retroarch:mednafen_psx"]
   enabled = true
   isolated_saves = true   # Use saves/{emulator}/{system}/
   ```

**Recommendation**: Keep current `saves/{system}/` structure. Most emulators for the same system use compatible formats. Document known incompatibilities.

**Impact**: Minimal changes needed. Current structure works.

---

### 7. Manifest Changes

#### Current Manifest

```toml
# ~/.local/state/kyaraben/manifest.toml
[emulators]
duckstation = { version = "v0.1-10655", nix_path = "/nix/store/..." }
eden = { version = "v0.1.0", nix_path = "/nix/store/..." }
```

#### Proposed Manifest

No structural changes needed. Manifest already tracks emulators, not systems.

```toml
[emulators]
# Multiple emulators can be installed
duckstation = { version = "v0.1-10655", nix_path = "..." }
"retroarch:mednafen_psx" = { version = "1.19.1", nix_path = "..." }
eden = { version = "v0.1.0", nix_path = "..." }
```

**Impact**: None. Already supports multi-emulator.

---

### 8. Desktop Integration Changes

#### Current Behavior

- One desktop file per enabled system
- Filename: `kyaraben-{system}.desktop`
- Example: `kyaraben-psx.desktop` → launches DuckStation

#### Proposed Behavior

- One desktop file per enabled emulator
- Filename: `kyaraben-{emulator}-{system}.desktop`
- Examples:
  - `kyaraben-duckstation-psx.desktop`
  - `kyaraben-retroarch_mednafen_psx-psx.desktop`

**Display names**:
- "PlayStation (DuckStation)"
- "PlayStation (Mednafen)"

```go
// Current
func desktopFileName(systemID SystemID) string {
    return fmt.Sprintf("kyaraben-%s.desktop", systemID)
}

// Proposed
func desktopFileName(emulatorID EmulatorID, systemID SystemID) string {
    // Sanitize emulator ID for filename (replace : with _)
    safe := strings.ReplaceAll(string(emulatorID), ":", "_")
    return fmt.Sprintf("kyaraben-%s-%s.desktop", safe, systemID)
}
```

**Impact**: Moderate. Launcher package needs updates.

---

### 9. Apply Process Changes

#### Current Apply Flow

```
1. Read config (system → emulator mapping)
2. For each enabled system:
   a. Get the one configured emulator
   b. Install emulator package
   c. Generate emulator config
   d. Create desktop file
3. Update manifest
```

#### Proposed Apply Flow

```
1. Read config (enabled emulators with versions)
2. For each enabled emulator:
   a. Install emulator package (if not already at version)
   b. For each system the emulator supports:
      i.  Generate emulator config (paths for this system)
      ii. Create desktop file for emulator-system pair
3. Update manifest
```

**Key difference**: Emulator installation is decoupled from system iteration. One emulator install serves multiple systems (e.g., Dolphin serves GameCube AND Wii).

**Impact**: Moderate refactor of apply logic.

---

### 10. Doctor Command Changes

#### Current Doctor

```
$ kyaraben doctor

PlayStation (DuckStation)
  ✓ BIOS: scph5501.bin (USA)
  ✓ BIOS: scph5500.bin (Japan)
  ✗ BIOS: scph5502.bin (Europe) - missing

Switch (Eden)
  ✗ Firmware: prod.keys - missing
```

#### Proposed Doctor

```
$ kyaraben doctor

PlayStation
  DuckStation:
    ✓ BIOS: scph5501.bin (USA)
    ✓ BIOS: scph5500.bin (Japan)
    ✗ BIOS: scph5502.bin (Europe) - missing

  RetroArch (mednafen_psx):
    ✓ BIOS: scph5501.bin (USA)
    ⚠ BIOS: scph5500.bin (Japan) - optional
    ⚠ BIOS: scph5502.bin (Europe) - optional

Switch
  Eden:
    ✗ Firmware: prod.keys - missing
```

**Impact**: Output formatting change. Same underlying provision checking.

---

## Migration Path

### For Existing Users

Since kyaraben is unreleased, breaking changes are acceptable. However, a migration path is still useful for testing:

```go
func migrateConfig(old OldConfig) NewConfig {
    new := NewConfig{
        Global: old.Global,
        Sync:   old.Sync,
        Emulators: make(map[EmulatorID]EmulatorConf),
    }

    for systemID, systemConf := range old.Systems {
        emulatorID, version := parseEmulatorString(systemConf.Emulator)
        new.Emulators[emulatorID] = EmulatorConf{
            Enabled: true,
            Version: version,
        }
    }

    return new
}
```

### Version Detection

```toml
# Add version field to detect config format
version = 2  # New format

[emulators]
# ...
```

---

## Open Questions

### 1. RetroArch Core Granularity

Should RetroArch cores be treated as:

**Option A: Individual emulators**
```toml
[emulators."retroarch:bsnes"]
enabled = true

[emulators."retroarch:mednafen_psx"]
enabled = true
```
- Pro: Fine-grained control
- Con: Must install RetroArch multiple times? Or share?

**Option B: RetroArch as one emulator with core selection**
```toml
[emulators.retroarch]
enabled = true
cores = ["bsnes", "mednafen_psx", "mupen64plus"]
```
- Pro: Single RetroArch install
- Con: Different config model than other emulators

**Current implementation**: Option A (cores are separate emulator IDs). This seems correct.

### 2. Default Emulator Concept

Do we need a "default" emulator per system?

**Use cases for default**:
- External tools that query "what plays PSX?"
- Quick launch without selection
- Migration from old config

**Recommendation**: Remove the concept. With proper desktop integration, users launch the emulator they want directly.

### 3. Provision Sharing

When two emulators for the same system need the same BIOS:
- Currently: Both point to `bios/{system}/`
- Works correctly, no changes needed

When emulators need different provisions:
- Already handled: provisions are per-emulator in model

### 4. Config Conflict Detection

What if user enables two emulators that would conflict?

**Example**: Two emulators wanting to write the same config file (unlikely but possible).

**Recommendation**: Trust emulator definitions. Each emulator writes to its own config paths. If conflicts arise, detect at apply time and error with clear message.

### 5. UI Complexity Concerns

Multi-emulator adds complexity. Mitigations:
- Show only relevant emulators (filter by available provisions)
- Group by system in UI (familiar mental model)
- Clear "why" messaging: "Enable multiple emulators when you need compatibility fallbacks"
- "Quick setup" mode that just enables recommended emulator per system

---

## Summary of Changes

| Component | Change Level | Description |
|-----------|--------------|-------------|
| **Domain Model** | Moderate | Remove `DefaultEmulatorID`, add `EnabledEmulator` |
| **Config Format** | Breaking | System-centric → Emulator-centric |
| **Registry** | Low | Remove default emulator concept |
| **Protocol/API** | Breaking | New request/response structures |
| **Frontend** | High | Redesign selection UI for multi-select |
| **Storage Paths** | None | Already supports multi-emulator |
| **Manifest** | None | Already tracks emulators |
| **Desktop Files** | Moderate | New naming scheme, multiple per system |
| **Apply Process** | Moderate | Iterate emulators, not systems |
| **Doctor** | Low | Output formatting changes |

---

## Appendix: Type Definitions Summary

### Go Types (Proposed)

```go
// Config
type KyarabenConfig struct {
    Version   int                           `toml:"version"`
    Global    GlobalConf                    `toml:"global"`
    Emulators map[EmulatorID]EmulatorConf   `toml:"emulators"`
    Sync      SyncConf                      `toml:"sync,omitempty"`
}

type EmulatorConf struct {
    Enabled bool   `toml:"enabled"`
    Version string `toml:"version,omitempty"`
}

// Protocol
type GetEmulatorsResponse struct {
    Emulators []EmulatorInfo `json:"emulators"`
}

type EmulatorInfo struct {
    ID                EmulatorID   `json:"id"`
    Name              string       `json:"name"`
    Systems           []SystemID   `json:"systems"`
    DefaultVersion    string       `json:"defaultVersion"`
    AvailableVersions []string     `json:"availableVersions"`
    Enabled           bool         `json:"enabled"`
    ConfiguredVersion string       `json:"configuredVersion,omitempty"`
}

type SetConfigRequest struct {
    UserStore string                       `json:"userStore"`
    Emulators map[EmulatorID]EmulatorConf  `json:"emulators"`
}
```

### TypeScript Types (Proposed)

```typescript
interface EmulatorInfo {
    id: EmulatorID;
    name: string;
    systems: SystemID[];
    defaultVersion: string;
    availableVersions: string[];
    enabled: boolean;
    configuredVersion?: string;
}

interface EmulatorConf {
    enabled: boolean;
    version?: string;
}

interface SetConfigRequest {
    userStore: string;
    emulators: Record<EmulatorID, EmulatorConf>;
}
```
