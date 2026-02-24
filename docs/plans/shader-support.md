# Shader support plan

## Summary

Add shader support to Kyaraben with per-emulator granularity. A global toggle
enables shaders for all systems, and per-emulator toggles allow users to
override the default. The UI exposes both a global "Enable shaders" toggle and
per-emulator shader toggles in a new modal on each emulator subcard.

CRT shader: `crt-mattias` (from libretro/glsl-shaders)
LCD shader: `zfast-lcd` (from libretro/glsl-shaders, EmuDeck default)

## Emulator coverage

Emulators that support programmatic shader configuration:

| Emulator | Systems | Display type | Config mechanism |
|---|---|---|---|
| RetroArch (all cores) | NES, SNES, Genesis, N64, GB/GBC/GBA, NDS, 3DS, Saturn, PCE, NGP, Arcade, NeoGeo, Atari2600, C64 | mixed | Per-core `.slangp` preset files under `config/<CoreName>/` |
| DuckStation | PSX | CRT | `settings.ini` `[PostProcessing]` section |
| PCSX2 | PS2 | CRT | `PCSX2.ini` `[EmuCore/GS]` `TVShader` |
| PPSSPP | PSP | LCD | `ppsspp.ini` `[Graphics]` `PostShaderNames` |

Emulators without programmatic shader support (out of scope): Dolphin,
Flycast, Eden, Cemu, xemu, Xenia Edge, RPCS3, Vita3K.

Dolphin has a `PostProcessingShader` setting in `GFX.ini` but does not ship
CRT shaders in its current release. Only color/novelty filters are included.

## System display type classification

Add a `DisplayType` field to the `System` model.

| System | Display type |
|---|---|
| NES, SNES, Genesis, Master System, N64, Saturn, Dreamcast, PC Engine, PSX, PS2, PS3, GameCube, Wii, Wii U, Arcade, NeoGeo, Xbox, Xbox 360, Atari 2600, C64 | CRT |
| GB, GBC, GBA, NDS, N3DS, NGP, Game Gear, PSP, PS Vita, Switch | LCD |

## Config design

### config.toml

Per-emulator shader state, with a convenience field for the global toggle:

```toml
[shaders]
enabled = false

[shaders.emulators]
# Per-emulator overrides. When absent, inherits from [shaders].enabled.
# When present, overrides the global toggle for this emulator.
# retroarch_mgba = true
# duckstation = false
```

The Go struct:

```go
type ShaderConfig struct {
    Enabled   bool            `toml:"enabled"`
    Emulators map[EmulatorID]bool `toml:"emulators,omitempty"`
}
```

Resolution logic: for a given emulator, if `ShaderConfig.Emulators[emuID]`
exists, use that value. Otherwise fall back to `ShaderConfig.Enabled`.

### Daemon protocol

Extend `SetConfigRequest` and `ConfigResponse` with a `shaders` field:

```go
// In SetConfigRequest
Shaders *ShaderConfigRequest `json:"shaders,omitempty"`

type ShaderConfigRequest struct {
    Enabled   bool                `json:"enabled"`
    Emulators map[string]bool     `json:"emulators,omitempty"`
}

// In ConfigResponse
Shaders ShaderConfigResponse `json:"shaders"`

type ShaderConfigResponse struct {
    Enabled   bool            `json:"enabled"`
    Emulators map[string]bool `json:"emulators,omitempty"`
}
```

The `handleSetConfig` method updates `cfg.Shaders` when the field is present.
The `handleGetConfig` method returns the current shader state.

### UI state

Add to `ConfigState` in `App.tsx`:

```ts
shadersEnabled: boolean
shaderEmulatorOverrides: Map<EmulatorID, boolean>
```

The `hasConfigChanges` check compares these against `savedConfigState`.

The `handleApply` callback passes the shader state in `setConfig`:

```ts
shaders: {
  enabled: configState.shadersEnabled,
  emulators: Object.fromEntries(configState.shaderEmulatorOverrides),
}
```

## UI design

### Global toggle

In the `Settings` component (above the systems list), add a "Shaders" row
with a toggle switch. Toggling it sets `configState.shadersEnabled` and
triggers the Apply bar, just like changing the store path does today.

### Per-emulator toggle

On each `EmulatorSubcard`, add a "Shaders" link in the action row alongside
"Launch", "Paths", and "Provisions". Clicking it opens a `ShadersModal`.

The `ShadersModal` shows:

- The emulator name
- The shader that will be applied (e.g. "CRT: crt-mattias" or "LCD:
  zfast-lcd") based on the system's display type
- A toggle: "Enable shader for this emulator"
- The toggle's state: resolved from the per-emulator override if set,
  otherwise from the global toggle
- When the user toggles it, it creates a per-emulator override in
  `shaderEmulatorOverrides`

Changing the per-emulator shader toggle sets `hasConfigChanges = true` and
shows the Apply bar, just like toggling an emulator on/off does today.

For emulators that do not support shaders (Dolphin, Flycast, etc.), the
"Shaders" link is not shown.

### Which emulators show the shader link

The daemon needs to communicate which emulators support shaders. Add a
`supportsShaders` boolean to `EmulatorRef` in the protocol. Set it based on
whether the emulator's config generator supports shader entries. The UI uses
this to conditionally render the "Shaders" link.

## Not using DefaultOnly

Shader config entries do not use `DefaultOnly`. The reasoning:

`DefaultOnly` means "write this value only if the key does not already exist
in the emulator's config file." This is designed for settings that Kyaraben
sets once as a default but the user might customize directly in the emulator.

Shaders are different: users change them frequently through the Kyaraben UI.
If we used `DefaultOnly`, the first Apply would write the shader setting, but
subsequent toggles via the UI would not take effect because the key already
exists in the config file.

Instead, shader config entries are always written when shaders are enabled for
that emulator, and explicitly disabled (or removed) when shaders are turned
off. This means:

- Turning shaders on in Kyaraben always writes the shader config
- Turning shaders off in Kyaraben always clears the shader config
- If a user changes shaders directly in the emulator, the next Kyaraben Apply
  will overwrite their change (this is acceptable since shader control is
  meant to go through Kyaraben)

For RetroArch, this means the `.slangp` preset file is written when enabled
and deleted when disabled. For standalone emulators, the config entries are
always written with the current state.

## Implementation steps

### 1. Add `DisplayType` to the system model

File: `internal/model/system.go`

```go
type DisplayType string

const (
    DisplayTypeCRT DisplayType = "crt"
    DisplayTypeLCD DisplayType = "lcd"
)
```

Add `DisplayType DisplayType` field to `System` struct. Update all system
definitions to set the appropriate display type.

### 2. Add `ShaderConfig` to `KyarabenConfig`

File: `internal/model/config.go`

```go
type ShaderConfig struct {
    Enabled   bool                    `toml:"enabled"`
    Emulators map[EmulatorID]bool     `toml:"emulators,omitempty"`
}
```

Add `Shaders ShaderConfig` field to `KyarabenConfig` with TOML key `shaders`.
Default is `Enabled: false` with no per-emulator overrides.

### 3. Thread shader config through `GenerateContext`

File: `internal/model/definitions.go`

Add `ShaderConfig *ShaderConfig` to `GenerateContext`. The daemon passes the
parsed shader config when calling each emulator's `Generate()` method.

Also add a `SystemDisplayTypes map[SystemID]DisplayType` field to
`GenerateContext` so config generators can look up the display type for the
systems they serve.

Each config generator resolves whether shaders are enabled for its emulator:
check `ShaderConfig.Emulators[emuID]` first, fall back to
`ShaderConfig.Enabled`.

### 4. RetroArch shader implementation

File: `internal/emulators/retroarch/shared.go`

When shaders are enabled for the core, `CorePatches` writes an additional
per-core shader preset file. RetroArch loads automatic shader presets from
`config/<CoreName>/<CoreName>.slangp`.

The preset file references the shader from RetroArch's bundled Slang shader
directory. RetroArch ships with both `crt-mattias` and `zfast-lcd` in its
Slang shader packs.

For each RetroArch core, look up the system's display type and write the
appropriate preset:

CRT preset (`<CoreName>.slangp`):
```
shaders = 1
shader0 = shaders/crt-mattias.slang
filter_linear0 = false
```

LCD preset (`<CoreName>.slangp`):
```
shaders = 1
shader0 = shaders/zfast_lcd.slang
scale_type0 = viewport
filter_linear0 = true
```

These are written as `ConfigFormatRaw` patches since they are standalone
preset files, not key-value config entries.

Also set `video_shader_enable = "true"` in the shared RetroArch config when
shaders are enabled for any core.

When shaders are disabled for a core, delete its `.slangp` file (emit a
delete patch or a cleanup action). Set `video_shader_enable = "false"` in the
shared config if no cores have shaders enabled.

### 5. DuckStation shader implementation

File: `internal/emulators/duckstation/duckstation.go`

DuckStation ships with `crt-lottes` as a built-in shader. When shaders are
enabled for DuckStation:

```go
{Path: []string{"PostProcessing", "Enabled"}, Value: "true"},
{Path: []string{"PostProcessing", "StageCount"}, Value: "1"},
{Path: []string{"PostProcessing/Stage1", "ShaderName"}, Value: "crt-lottes"},
```

When shaders are disabled:

```go
{Path: []string{"PostProcessing", "Enabled"}, Value: "false"},
```

No `DefaultOnly` -- always written.

### 6. PCSX2 shader implementation

File: `internal/emulators/pcsx2/pcsx2.go`

PCSX2 has a built-in `TVShader` setting under `[EmuCore/GS]` in `PCSX2.ini`.
Value `5` selects the Lottes CRT shader. When shaders are enabled:

```go
{Path: []string{"EmuCore/GS", "TVShader"}, Value: "5"},
```

When shaders are disabled:

```go
{Path: []string{"EmuCore/GS", "TVShader"}, Value: "0"},
```

TVShader values for reference: 0 = none, 1 = scanline filter, 2 = diagonal
filter, 3 = triangular filter, 4 = wave filter, 5 = Lottes CRT.

### 7. PPSSPP shader implementation

File: `internal/emulators/ppsspp/ppsspp.go`

PPSSPP ships with a built-in `LCDPersistence` shader. When shaders are
enabled:

```go
{Path: []string{"Graphics", "PostShaderNames"}, Value: "LCDPersistence"},
```

When shaders are disabled:

```go
{Path: []string{"Graphics", "PostShaderNames"}, Value: "Off"},
```

### 8. Extend the daemon protocol

File: `internal/daemon/protocol.go`

Add `Shaders` to `SetConfigRequest`, `ConfigResponse`, and their handler
methods. Add `SupportsShaders bool` to `EmulatorRef` and `SystemRef`.

File: `internal/daemon/daemon.go`

In `handleSetConfig`: when `data.Shaders` is non-nil, update
`cfg.Shaders.Enabled` and `cfg.Shaders.Emulators`.

In `handleGetConfig`: return `cfg.Shaders.Enabled` and
`cfg.Shaders.Emulators`.

In `handleGetSystems`: set `SupportsShaders` on each `EmulatorRef` based on
whether the emulator's definition supports shader config.

### 9. UI: global shader toggle

File: `ui/src/components/Settings/Settings.tsx`

Add a "Shaders" row with a `ToggleSwitch`. Wire it to
`configState.shadersEnabled` via a new `onShadersToggle` prop threaded from
`App.tsx`.

File: `ui/src/App.tsx`

Add `shadersEnabled` and `shaderEmulatorOverrides` to `ConfigState`. Update
`parseConfigResponse`, `cloneConfigState`, `hasConfigChanges`, and
`handleApply` to include shader state.

### 10. UI: per-emulator shader modal

File: `ui/src/components/ShadersModal/ShadersModal.tsx` (new)

A modal showing:
- Shader type (CRT or LCD) and name
- Toggle switch for this emulator
- Note about the global default

File: `ui/src/components/EmulatorSubcard/EmulatorSubcard.tsx`

Add a "Shaders" link in the action row (alongside Launch, Paths, Provisions).
Only shown when the emulator's `supportsShaders` is true. Clicking opens
`ShadersModal`.

Add `shadersEnabled`, `shaderOverride`, and `onShaderToggle` props.

### 11. Update the daemon and CLI

The CLI `init` command that generates the default `config.toml` should include
the `[shaders]` section with `enabled = false`.

### 12. Update documentation

Update `site/src/content/docs/using-the-app.mdx` and
`site/src/content/docs/using-the-cli.mdx` to document the shader config
options.

### 13. Tests

- Unit tests for display type classification (all systems have a display type)
- Unit tests for each emulator's config generator: verify shader entries are
  present when enabled, absent when disabled
- Unit tests for RetroArch preset file content
- Registry test: verify all system definitions set a display type
- Unit test for shader resolution logic (per-emulator override vs global)
- UI component tests for ShadersModal

## Resolved decisions

- RetroArch presets use `.slangp` (Slang/Vulkan). Slang is the recommended
  shader format and works with Vulkan, OpenGL Core, and Direct3D renderers.
- Dolphin is out of scope. The version Kyaraben ships does not include CRT
  shaders (only color/novelty filters like grayscale, sepia, etc.).
- PPSSPP uses the built-in `LCDPersistence` shader. It is a frame persistence
  effect rather than an LCD grid shader, but it is the only built-in option.
- No `DefaultOnly` for shader entries. Shaders are always written on Apply so
  the UI toggle works reliably across repeated applies.

## Config example

```toml
[shaders]
enabled = true

[shaders.emulators]
# User disabled shaders for PPSSPP specifically
ppsspp = false
```
