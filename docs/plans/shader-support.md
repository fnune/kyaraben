# Shader support plan

## Summary

Add shader support to Kyaraben with per-emulator granularity. Each emulator
that supports shaders gets a three-state control:

- `shaders = true`: Kyaraben writes curated shader config for the emulator
- `shaders = false`: Kyaraben explicitly disables shaders for the emulator
- key absent: Kyaraben does not touch the emulator's shader config

The setting lives on `EmulatorConf` (the existing per-emulator config struct).
There is no global toggle. The default is unmanaged (absent), so existing users
are unaffected.

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

Shader state is per-emulator, stored on the existing `EmulatorConf`. The key
uses a `*bool` in Go so three states are representable in TOML:

```toml
[emulators.retroarch-bsnes]
version = "0.240"
shaders = true

[emulators.ppsspp]
shaders = false

# duckstation has no shaders key: Kyaraben leaves its shader config alone
```

The Go struct:

```go
type EmulatorConf struct {
    Version string `toml:"version,omitempty"`
    Shaders *bool  `toml:"shaders,omitempty"`
}
```

Resolution logic: if `Shaders` is `nil`, the emulator's shader config is
unmanaged. If `*Shaders` is `true`, Kyaraben writes the curated shader. If
`*Shaders` is `false`, Kyaraben explicitly disables shaders.

### Daemon protocol

Extend `EmulatorConfRequest` and `EmulatorConfResponse` with a `shaders`
field:

```go
type EmulatorConfRequest struct {
    Version string `json:"version,omitempty"`
    Shaders *bool  `json:"shaders,omitempty"`
}

type EmulatorConfResponse struct {
    Version string `json:"version,omitempty"`
    Shaders *bool  `json:"shaders,omitempty"`
}
```

The `handleSetConfig` method writes `cfg.Emulators[emuID].Shaders` from the
request. When the request omits an emulator entirely, its shader state is
unchanged. When the request includes an emulator with `shaders: null` (JSON
null), the key is removed from the TOML (unmanaged).

The `handleGetConfig` method returns the current shader state per emulator.

### UI state

Add to the emulator state in `App.tsx`:

```ts
// null = unmanaged (absent in TOML), true = managed on, false = managed off
emulatorShaders: Map<EmulatorID, boolean | null>
```

The `hasConfigChanges` check compares this against `savedConfigState`.

The `handleApply` callback includes shader state in each emulator's conf when
calling `setConfig`.

## UI design

### Per-emulator shader control

On each `EmulatorSubcard`, add a "Shaders" button in the action row alongside
"Launch", "Paths", and "Provisions". Only shown when the emulator's
`supportsShaders` is true.

Clicking the button opens a `ShadersModal` that shows:

- The shader that would be applied (e.g. "CRT: crt-mattias" or "LCD:
  zfast-lcd") based on the system's display type, for informational context
- A three-state selector with these options:
  - "On": Kyaraben applies the curated shader (`shaders = true`)
  - "Off": Kyaraben disables shaders (`shaders = false`)
  - "Manual": Kyaraben does not touch shader config (key absent)
- "Manual" is the default for all emulators

The three-state selector is a segmented control (three side-by-side buttons
with the active one highlighted), not a dropdown or toggle. All options are
visible at a glance.

Changing the shader state sets `hasConfigChanges = true` and shows the Apply
bar.

### Visual indicator on the subcard

When an emulator's shader state is "On", show a subtle indicator on the
`EmulatorSubcard` so users can see shader status without opening the modal.
This could be a small label or icon next to the emulator name. The exact
treatment is left to implementation.

### Which emulators show the shader button

The daemon communicates which emulators support shaders. Add a
`supportsShaders` boolean to `EmulatorRef` in the protocol. Set it based on
whether the emulator has a config generator that supports shader entries. The
UI uses this to conditionally render the "Shaders" button.

## Not using DefaultOnly

Shader config entries do not use `DefaultOnly`. The reasoning:

`DefaultOnly` means "write this value only if the key does not already exist
in the emulator's config file." This is designed for settings that Kyaraben
sets once as a default but the user might customize directly in the emulator.

Shaders are different: when the user sets shaders to "On" or "Off" in
Kyaraben, the intent is for Kyaraben to control that setting on every Apply.
If we used `DefaultOnly`, the first Apply would write the shader setting, but
subsequent changes via the UI would not take effect because the key already
exists in the config file.

When shaders are "On":
- Kyaraben always writes the curated shader config

When shaders are "Off":
- Kyaraben always writes config that disables shaders

When shaders are "Manual" (absent):
- Kyaraben does not claim the shader config region at all
- The user's own emulator settings are preserved across applies
- This is the default, so existing users are unaffected

For RetroArch, "On" means the `.slangp` preset file is written, "Off" means
it is deleted, and "Manual" means it is not touched.

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

### 2. Add `Shaders *bool` to `EmulatorConf`

File: `internal/model/config.go`

```go
type EmulatorConf struct {
    Version string `toml:"version,omitempty"`
    Shaders *bool  `toml:"shaders,omitempty"`
}
```

Add a helper method:

```go
func (c *KyarabenConfig) EmulatorShaders(id EmulatorID) *bool {
    if conf, ok := c.Emulators[id]; ok {
        return conf.Shaders
    }
    return nil
}
```

### 3. Thread shader state through `GenerateContext`

File: `internal/model/definitions.go`

Add `Shaders *bool` to `GenerateContext`. The applier sets this per emulator
before calling `Generate()`, reading from `cfg.EmulatorShaders(emuID)`.

Also add `SystemDisplayTypes map[SystemID]DisplayType` to `GenerateContext`
so config generators can look up the display type for the systems they serve.

Each config generator checks `genCtx.Shaders`:
- `nil`: do not emit any shader patches
- `true`: emit patches that enable the curated shader
- `false`: emit patches that disable shaders

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

When shaders are explicitly disabled (`false`), delete the `.slangp` file
(emit a delete patch or a cleanup action). Set `video_shader_enable = "false"`
in the shared config if no cores have shaders enabled.

When shaders are unmanaged (`nil`), do not touch the `.slangp` file or the
`video_shader_enable` setting.

### 5. DuckStation shader implementation

File: `internal/emulators/duckstation/duckstation.go`

DuckStation ships with `crt-lottes` as a built-in shader. When shaders are
enabled (`true`):

```go
{Path: []string{"PostProcessing", "Enabled"}, Value: "true"},
{Path: []string{"PostProcessing", "StageCount"}, Value: "1"},
{Path: []string{"PostProcessing/Stage1", "ShaderName"}, Value: "crt-lottes"},
```

When shaders are disabled (`false`):

```go
{Path: []string{"PostProcessing", "Enabled"}, Value: "false"},
```

When shaders are unmanaged (`nil`): no patches emitted.

### 6. PCSX2 shader implementation

File: `internal/emulators/pcsx2/pcsx2.go`

PCSX2 has a built-in `TVShader` setting under `[EmuCore/GS]` in `PCSX2.ini`.
Value `5` selects the Lottes CRT shader. When shaders are enabled (`true`):

```go
{Path: []string{"EmuCore/GS", "TVShader"}, Value: "5"},
```

When shaders are disabled (`false`):

```go
{Path: []string{"EmuCore/GS", "TVShader"}, Value: "0"},
```

When shaders are unmanaged (`nil`): no patches emitted.

TVShader values for reference: 0 = none, 1 = scanline filter, 2 = diagonal
filter, 3 = triangular filter, 4 = wave filter, 5 = Lottes CRT.

### 7. PPSSPP shader implementation

File: `internal/emulators/ppsspp/ppsspp.go`

PPSSPP ships with a built-in `LCDPersistence` shader. When shaders are
enabled (`true`):

```go
{Path: []string{"Graphics", "PostShaderNames"}, Value: "LCDPersistence"},
```

When shaders are disabled (`false`):

```go
{Path: []string{"Graphics", "PostShaderNames"}, Value: "Off"},
```

When shaders are unmanaged (`nil`): no patches emitted.

### 8. Extend the daemon protocol

File: `internal/daemon/protocol.go`

Add `Shaders *bool` to `EmulatorConfRequest` and `EmulatorConfResponse`. Add
`SupportsShaders bool` to `EmulatorRef`.

File: `internal/daemon/daemon.go`

In `handleSetConfig`: when an emulator conf is present in the request, write
its `Shaders` value to `cfg.Emulators[emuID].Shaders`.

In `handleGetConfig`: return each emulator's `Shaders` value.

In `handleGetSystems`: set `SupportsShaders` on each `EmulatorRef` based on
whether the emulator's definition supports shader config.

### 9. UI: state management

File: `ui/src/App.tsx`

Add `emulatorShaders: Map<EmulatorID, boolean | null>` to `ConfigState`.
Update `parseConfigResponse`, `cloneConfigState`, `hasConfigChanges`, and
`handleApply` to include shader state.

The `handleShaderChange(emulatorId, value)` callback accepts
`boolean | null` and updates the map.

### 10. UI: per-emulator shader modal

File: `ui/src/components/ShadersModal/ShadersModal.tsx` (new)

A modal showing:
- The shader name and type (CRT or LCD) that Kyaraben would apply
- A segmented control with three options: "On", "Off", "Manual"
- Current selection highlighted

File: `ui/src/components/EmulatorSubcard/EmulatorSubcard.tsx`

Add a "Shaders" button in the action row (alongside Launch, Paths,
Provisions). Only shown when the emulator's `supportsShaders` is true.
Clicking opens `ShadersModal`.

Add `shaders: boolean | null`, `onShaderChange`, and `supportsShaders` props.

### 11. Update documentation

Update `site/src/content/docs/using-the-app.mdx` and
`site/src/content/docs/using-the-cli.mdx` to document the shader config
options and the three states.

### 12. Tests

- Unit tests for display type classification (all systems have a display type)
- Unit tests for each emulator's config generator: verify shader entries are
  present when `true`, disable entries when `false`, nothing when `nil`
- Unit tests for RetroArch preset file content
- Registry test: verify all system definitions set a display type
- UI component tests for ShadersModal segmented control

## Resolved decisions

- No global toggle. Shader config is purely per-emulator. This avoids the
  complexity of a two-level override system and keeps the mental model simple:
  each emulator has exactly one shader setting.
- Per-emulator shader state is a `*bool` on `EmulatorConf`, not a separate
  config section. This keeps the config flat and follows the pattern of the
  existing `version` field.
- Default is unmanaged (absent). Existing users are unaffected. New users
  start with Kyaraben not touching shader config for any emulator.
- RetroArch presets use `.slangp` (Slang/Vulkan). Slang is the recommended
  shader format and works with Vulkan, OpenGL Core, and Direct3D renderers.
- Dolphin is out of scope. The version Kyaraben ships does not include CRT
  shaders (only color/novelty filters like grayscale, sepia, etc.).
- PPSSPP uses the built-in `LCDPersistence` shader. It is a frame persistence
  effect rather than an LCD grid shader, but it is the only built-in option.
- No `DefaultOnly` for shader entries. When managed (`true` or `false`),
  shaders are always written on Apply so the UI control works reliably.
- The UI uses a segmented control ("On" / "Off" / "Manual") rather than a
  toggle or dropdown. All three states are visible at a glance.

## Config example

```toml
[emulators.retroarch-bsnes]
version = "0.240"
shaders = true

[emulators.ppsspp]
shaders = false

# retroarch-mgba, duckstation, pcsx2, etc. have no shaders key.
# Kyaraben does not touch their shader config.
```
