# Shader support plan

## Summary

Add a `[shaders]` section to `config.toml` with an `enabled` boolean (default
`false`). When enabled, Kyaraben applies a CRT shader to systems that used CRT
displays and an LCD shader to handheld/portable systems, covering all emulators
that support programmatic shader configuration.

CRT shader: `crt-mattias` (from libretro/glsl-shaders)
LCD shader: `zfast-lcd` (from libretro/glsl-shaders, EmuDeck default)

## Emulator coverage

Emulators that support programmatic shader configuration:

| Emulator | Systems | Display type | Config mechanism |
|---|---|---|---|
| RetroArch (all cores) | NES, SNES, Genesis, N64, GB/GBC/GBA, NDS, 3DS, Saturn, PCE, NGP, Arcade, NeoGeo, Atari2600, C64 | mixed | Per-core `.glslp` preset files under `config/<CoreName>/` |
| DuckStation | PSX | CRT | `settings.ini` `[PostProcessing]` section |
| PCSX2 | PS2 | CRT | `PCSX2.ini` `[EmuCore/GS]` `TVShader` |
| Dolphin | GameCube, Wii | CRT | `GFX.ini` `[Settings]` `PostProcessingShader` |
| PPSSPP | PSP | LCD | `ppsspp.ini` `[Graphics]` `PostShaderNames` |

Emulators without programmatic shader support (out of scope): Flycast, Eden,
Cemu, xemu, Xenia Edge, RPCS3, Vita3K.

## System display type classification

Add a `DisplayType` field to the `System` model.

| System | Display type |
|---|---|
| NES, SNES, Genesis, Master System, N64, Saturn, Dreamcast, PC Engine, PSX, PS2, PS3, GameCube, Wii, Wii U, Arcade, NeoGeo, Xbox, Xbox 360, Atari 2600, C64 | CRT |
| GB, GBC, GBA, NDS, N3DS, NGP, Game Gear, PSP, PS Vita, Switch | LCD |

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
definitions in `internal/systems/*/` to set the appropriate display type.

### 2. Add `ShaderConfig` to `KyarabenConfig`

File: `internal/model/config.go`

```go
type ShaderConfig struct {
    Enabled bool `toml:"enabled"`
}
```

Add `Shaders ShaderConfig` field to `KyarabenConfig` with TOML key `shaders`.
Default is `Enabled: false`.

### 3. Thread shader config through `GenerateContext`

File: `internal/model/definitions.go`

Add `ShaderConfig *ShaderConfig` to `GenerateContext`. The daemon passes the
parsed shader config when calling each emulator's `Generate()` method.

Also add a `SystemDisplayTypes map[SystemID]DisplayType` field to
`GenerateContext` so config generators can look up the display type for the
systems they serve.

### 4. RetroArch shader implementation

File: `internal/emulators/retroarch/shared.go`

When shaders are enabled, `CorePatches` writes an additional per-core shader
preset file. RetroArch loads automatic shader presets from
`config/<CoreName>/<CoreName>.slangp` (or `.glslp`).

The preset file references the shader from RetroArch's bundled shader
directory. RetroArch ships with both `crt-mattias` and `zfast-lcd` in its
shader packs.

For each RetroArch core, look up the system's display type and write the
appropriate preset:

CRT preset (`<CoreName>.glslp`):
```
shaders = 1
shader0 = shaders/crt-mattias.glsl
filter_linear0 = false
```

LCD preset (`<CoreName>.glslp`):
```
shaders = 1
shader0 = shaders/zfast_lcd.glsl
scale_type0 = viewport
filter_linear0 = true
```

These are written as `ConfigFormatRaw` patches since they are standalone
preset files, not key-value config entries.

The shader preset target:
```go
func CoreShaderTarget(shortName string) model.ConfigTarget {
    configDirName := shortName
    if displayName, ok := coreConfigDirNames[shortName]; ok {
        configDirName = displayName
    }
    return model.ConfigTarget{
        RelPath: "retroarch/config/" + configDirName + "/" + configDirName + ".glslp",
        Format:  model.ConfigFormatRaw,
        BaseDir: model.ConfigBaseDirUserConfig,
    }
}
```

Also set `video_shader_enable = "true"` in the shared RetroArch config when
shaders are enabled.

When shaders are disabled, do not write preset files and set
`video_shader_enable = "false"`.

Multi-system cores (like mGBA which serves GB, GBC, GBA): since all three are
LCD, one preset works. The `coreToSystem` map already picks one primary system
per core. All current multi-system RetroArch cores in Kyaraben serve systems
that share the same display type, so a single preset per core is sufficient.

### 5. DuckStation shader implementation

File: `internal/emulators/duckstation/duckstation.go`

DuckStation ships with `crt-lottes` as a built-in shader. When shaders are
enabled, add these entries to `settings.ini`:

```go
{Path: []string{"PostProcessing", "Enabled"}, Value: "true"},
{Path: []string{"PostProcessing", "StageCount"}, Value: "1"},
{Path: []string{"PostProcessing/Stage1", "ShaderName"}, Value: "crt-lottes"},
```

When shaders are disabled:
```go
{Path: []string{"PostProcessing", "Enabled"}, Value: "false"},
```

Use `DefaultOnly: true` so users who have customized their shader setup are
not overwritten.

### 6. PCSX2 shader implementation

File: `internal/emulators/pcsx2/pcsx2.go`

PCSX2 has a built-in `TVShader` setting under `[EmuCore/GS]` in `PCSX2.ini`.
Value `5` selects the Lottes CRT shader. When shaders are enabled, add:

```go
{Path: []string{"EmuCore/GS", "TVShader"}, Value: "5", DefaultOnly: true},
```

When shaders are disabled, do not add the entry (the PCSX2 default of `0`
means no TV shader).

TVShader values for reference: 0 = none, 1 = scanline filter, 2 = diagonal
filter, 3 = triangular filter, 4 = wave filter, 5 = Lottes CRT.

### 7. Dolphin shader implementation

File: `internal/emulators/dolphin/dolphin.go`

Dolphin includes CRT shaders from the Clownacy PR (crt-pi, crt-lottes-fast).
When shaders are enabled, add this entry to the existing `gfxTarget` patch:

```go
{Path: []string{"Settings", "PostProcessingShader"}, Value: "crt-pi", DefaultOnly: true},
```

When shaders are disabled, do not add the entry (existing user config
preserved). If a user manually set a shader, `DefaultOnly` ensures Kyaraben
does not overwrite it.

### 8. PPSSPP shader implementation

File: `internal/emulators/ppsspp/ppsspp.go`

PPSSPP does not ship with an LCD shader by default. The `zfast_lcd` shader
exists as a community port for PPSSPP (jdgleaver/ppsspp_shaders). Two options:

Option A: Use the built-in `LCDPersistence` shader that PPSSPP ships with:
```go
{Path: []string{"Graphics", "PostShaderNames"}, Value: "LCDPersistence", DefaultOnly: true},
```

Option B: Skip PPSSPP shader support until a proper LCD shader can be bundled.

Recommend option A since `LCDPersistence` is built-in and requires no
additional files.

When shaders are disabled, do not add the entry.

### 9. Update the daemon and CLI

The daemon's `apply` handler already reads `KyarabenConfig` and passes
`GenerateContext` to each emulator. Add the `ShaderConfig` field to the
context construction. No new daemon commands needed.

The CLI `init` command that generates the default `config.toml` should include
the `[shaders]` section with `enabled = false`.

### 10. Update documentation

Update `site/src/content/docs/using-the-app.mdx` and
`site/src/content/docs/using-the-cli.mdx` to document the `[shaders]` config
option. Mention which emulators support it and which do not.

### 11. Tests

- Unit tests for display type classification (all systems have a display type)
- Unit tests for each emulator's config generator: verify shader entries are
  present when enabled, absent when disabled
- Unit tests for RetroArch preset file content
- Registry test: verify all system definitions set a display type

## Open questions

- Should we use `.glslp` (OpenGL) or `.slangp` (Vulkan) for RetroArch shader
  presets? The AppImage RetroArch likely uses OpenGL by default, so `.glslp`
  is safer. But if the user switches to Vulkan, `.glslp` presets would not
  work. We could write both, or detect the video driver from `retroarch.cfg`.
  Recommend: write `.glslp` for now since the AppImage defaults to OpenGL.
  Document the limitation.

- Dolphin's CRT shaders (crt-pi) were merged in PR #12014. We should verify
  they ship in the version Kyaraben installs. If not, the shader file would
  need to be bundled or the feature skipped for Dolphin.

- For PPSSPP, `LCDPersistence` is not a true LCD grid shader but a frame
  persistence effect. It is the only built-in option. A proper zfast-lcd port
  would need to be installed separately.

## Config example

```toml
[shaders]
enabled = true
```

That is the entire user-facing config surface. No per-system, per-emulator,
or per-shader options.
