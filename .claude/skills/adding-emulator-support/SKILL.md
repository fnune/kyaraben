---
name: adding-emulator-support
description: Documents the complete process for adding support for a new emulator or system to Kyaraben.
---

# Adding emulator support to Kyaraben

This skill documents the complete process for adding support for a new emulator or system to Kyaraben.

## Overview

Kyaraben uses a registry-based architecture where systems and emulators are defined separately and linked together. Adding a new emulator involves:

1. Deciding whether to use a RetroArch core or standalone emulator
2. Adding model constants (SystemID, EmulatorID, Manufacturer)
3. Creating system and emulator definition packages
4. Configuring provisions (BIOS/firmware requirements)
5. Setting up controller mappings
6. Adding UI assets (logos)
7. Registering everything in the registry
8. Running code generation and tests

## Decision: RetroArch core vs standalone emulator

### Use a RetroArch core when

- The core is available in the libretro buildbot bundle
- The core has good compatibility and performance
- You want to leverage RetroArch's unified configuration
- The system does not require complex setup (UI-based imports, etc.)

Note that you may find the correct version of the cores bundle already downloaded in `~/.local/state/kyaraben/downloads`. If not, download it.

### Use a standalone emulator when

- No suitable libretro core exists
- The standalone emulator is significantly better (accuracy, performance, features)
- The emulator requires complex setup that RetroArch cannot handle
- Examples: DuckStation, PCSX2, RPCS3, Dolphin

## Step 1: Add model constants

### Add SystemID

File: `internal/model/system.go`

```go
const (
    // ... existing systems ...
    SystemIDNewSystem SystemID = "newsystem"
)
```

If needed, add a new Manufacturer:

```go
const (
    // ... existing manufacturers ...
    ManufacturerNewCo Manufacturer = "NewCo"
)
```

### Add EmulatorID

File: `internal/model/emulator_id.go`

For RetroArch cores, use the format `retroarch:corename`:

```go
const (
    EmulatorIDRetroArchNewCore EmulatorID = "retroarch:newcore"
)
```

For standalone emulators:

```go
const (
    EmulatorIDNewEmu EmulatorID = "newemu"
)
```

## Step 2: Create system definition

Create a new package: `internal/systems/newsystem/newsystem.go`

```go
package newsystem

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) System() model.System {
    return model.System{
        ID:           model.SystemIDNewSystem,
        Name:         "New System",
        Description:  "Description of the system (year)",
        Manufacturer: model.ManufacturerNewCo,
        Label:        "NEW",  // Short label for UI
        Extensions:   []string{".ext1", ".ext2", ".7z", ".zip"},
    }
}

func (Definition) DefaultEmulatorID() model.EmulatorID {
    return model.EmulatorIDRetroArchNewCore
}
```

### Finding file extensions

Check these sources for supported extensions:

- EmuDeck: `~/Development/EmuDeck/roms/{system}/systeminfo.txt`
- RetroDECK: `~/Development/RetroDECK/config/retrodeck/reference_lists/`
- Libretro docs: `https://docs.libretro.com/library/{corename}/`
- ES-DE: `https://gitlab.com/es-de/emulationstation-de/-/blob/master/resources/systems/linux/es_systems.xml`

## Step 3: Create emulator definition

### For RetroArch cores

#### Add core to coreInfoMap

File: `internal/emulators/retroarch/shared.go`

Every RetroArch core must be added to `coreInfoMap` for `CoreSymlinks()` to work:

```go
var coreInfoMap = map[model.EmulatorID]CoreInfo{
    // ... existing cores ...
    model.EmulatorIDRetroArchNewCore: {
        ShortName:   "newcore",   // Used for directory naming
        LibraryName: "NewCore",   // Name RetroArch uses for save/state subdirs
        SystemID:    model.SystemIDNewSystem,
    },
}
```

Without this entry, `CoreSymlinks()` returns empty and saves/states won't sync properly.

#### Create emulator package

Create: `internal/emulators/retroarchnewcore/retroarchnewcore.go`

```go
package retroarchnewcore

import (
    "github.com/fnune/kyaraben/internal/emulators/retroarch"
    "github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
    return model.Emulator{
        ID:              model.EmulatorIDRetroArchNewCore,
        Name:            "New Core (RetroArch)",
        Systems:         []model.SystemID{model.SystemIDNewSystem},
        Package:         model.AppImageRef("retroarch"),
        ProvisionGroups: nil,  // Or define BIOS requirements
        StateKinds: []model.StateKind{
            model.StateSaves,
            model.StateSavestates,
            model.StateScreenshots,
        },
        Launcher:         retroarch.LauncherWithCore(libretroCoreName),
        PathUsage:        model.StandardPathUsage(),
        SupportedSettings: []string{
            model.SettingPreset,         // Graphics presets (clean/retro)
            model.SettingResumeAutosave, // Auto-save on exit
            model.SettingResumeAutoload, // Auto-load on start
        },
    }
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
    return &Config{}
}

type Config struct{}

func (c *Config) Generate(ctx model.GenerateContext) (model.GenerateResult, error) {
    pc := &retroarch.PresetConfig{
        Preset:             ctx.Preset,
        TargetDevice:       ctx.TargetDevice,
        Resume:             ctx.Resume,
        SystemDisplayTypes: ctx.SystemDisplayTypes,
    }

    symlinks, err := retroarch.CoreSymlinks(model.EmulatorIDRetroArchNewCore, ctx.Store, ctx.BaseDirResolver)
    if err != nil {
        return model.GenerateResult{}, err
    }

    downloads, err := retroarch.CoreShaderDownloads(model.EmulatorIDRetroArchNewCore, ctx.BaseDirResolver, pc)
    if err != nil {
        return model.GenerateResult{}, err
    }

    embedded, err := retroarch.CoreEmbeddedFiles(model.EmulatorIDRetroArchNewCore, []model.SystemID{model.SystemIDNewSystem}, pc, ctx.BaseDirResolver)
    if err != nil {
        return model.GenerateResult{}, err
    }

    return model.GenerateResult{
        Patches:          retroarch.CorePatches(model.EmulatorIDRetroArchNewCore, ctx.Store, ctx.ControllerConfig, pc, ctx.BaseDirResolver),
        Symlinks:         symlinks,
        InitialDownloads: downloads,
        EmbeddedFiles:    embedded,
    }, nil
}

const libretroCoreName = "newcore_libretro"
```

### For standalone emulators

See existing examples:

- Simple: `internal/emulators/flycast/flycast.go`
- With BIOS: `internal/emulators/duckstation/duckstation.go`
- Complex: `internal/emulators/rpcs3/rpcs3.go`

Key differences from RetroArch cores:

- Define `Package` with specific emulator AppImage reference
- Implement full `ConfigGenerator.Generate()` with emulator-specific config patches
- Handle emulator-specific controller mapping
- Configure save/state directory paths (or symlinks as last resort)

## Step 4: Configure provisions (BIOS/firmware)

### Provision types

- `ProvisionBIOS`: System BIOS files
- `ProvisionKeys`: Decryption keys
- `ProvisionFirmware`: Firmware files

### Provision strategies

#### FileProvision (name-only check)

```go
model.FileProvision(model.ProvisionBIOS, "bios.bin", "Description")
```

#### HashedProvision (MD5 verification)

```go
model.HashedProvision(model.ProvisionBIOS, "bios.bin", "Region/variant", []string{
    "md5hash1",
    "md5hash2",  // Multiple valid hashes for variants
})
```

#### PatternProvision (glob matching)

```go
model.PatternProvision(model.ProvisionBIOS, "*.bin", "pattern desc", "Description")
```

### Provision groups

```go
ProvisionGroups: []model.ProvisionGroup{{
    MinRequired: 1,  // 0 = optional, 1+ = minimum required
    Message:     "BIOS required (no HLE available)",
    Provisions:  biosProvisions,
}},
```

### System-specific provisions

```go
provision.ForSystems(model.SystemIDSpecific)
```

### Finding BIOS requirements

Check these sources:

- Libretro docs: `https://docs.libretro.com/library/{corename}/`
- EmuDeck BIOS checker: `~/Development/EmuDeck/functions/checkBIOS.sh`
- RetroDECK manifests: `~/Development/RetroDECK/` (search for MD5 hashes)
- Emulator documentation

## Step 5: Add core to versions.toml

For RetroArch cores from the buildbot bundle:

File: `internal/versions/versions.toml`

```toml
[newcore]
releases_url = "github:libretro/RetroArch"
url_template = "https://buildbot.libretro.com/stable/{version}/linux/{variant}/RetroArch_cores.7z"
default = "1.22.2"
binary_path = "newcore_libretro.so"
install_type = "retroarch-core"
bundle_size = 274237400

[newcore."1.22.2".targets.x64]
variant = "x86_64"
sha256 = "sha256-..."  # Same as other cores using the bundle
size = 1234567  # Get from: 7z l RetroArch_cores.7z | grep newcore
```

For standalone emulators, add a new section following existing patterns (duckstation, dolphin, etc.).

## Step 6: Update registry

File: `internal/registry/all.go`

Add imports:

```go
import (
    // ... existing imports ...
    "github.com/fnune/kyaraben/internal/emulators/retroarchnewcore"
    "github.com/fnune/kyaraben/internal/systems/newsystem"
)
```

Add to `NewDefault()`:

```go
func NewDefault() *Registry {
    return New(
        []model.SystemDefinition{
            // ... existing systems ...
            newsystem.Definition{},
        },
        []model.EmulatorDefinition{
            // ... existing emulators ...
            retroarchnewcore.Definition{},
        },
        // ...
    )
}
```

## Step 7: Update ESDE frontend mapping

File: `internal/frontends/esde/systems.go`

```go
var systemMappings = map[model.SystemID]SystemMapping{
    // ... existing mappings ...
    model.SystemIDNewSystem: {
        Name:     "newsystem",      // ES-DE system folder name
        FullName: "New System",     // Display name
        Platform: "newsystem",      // Scraping platform ID
    },
}
```

Find ES-DE system names at:

- ES-DE systems list: `https://gitlab.com/es-de/emulationstation-de/-/blob/master/resources/systems/linux/es_systems.xml`

## Step 8: Add UI assets

### System logos

Two logo mappings need to be updated for each new system:

#### ESDE logos (for system cards)

Add SVG: `ui/src/assets/esde/logos/newsystem.svg`

File: `ui/src/assets/esde/index.ts`

```typescript
import newsystemLogo from "./logos/newsystem.svg";

export const ESDE_LOGOS: Record<SystemID, string> = {
  // ... existing logos ...
  newsystem: newsystemLogo,
};
```

#### System logos (for system picker)

Add SVG: `ui/src/assets/systems/newsystem.svg`

File: `ui/src/components/SystemLogo/SystemLogo.tsx`

```typescript
import newsystem from "@/assets/systems/newsystem.svg";

export const SYSTEM_LOGOS: Record<SystemID, string> = {
  // ... existing logos ...
  newsystem,
};
```

### System year

File: `ui/src/components/SystemCard/SystemCard.tsx`

```typescript
export const SYSTEM_YEARS: Record<SystemID, number | null> = {
  // ... existing years ...
  newsystem: 1999, // Use null for categories like arcade
};
```

### Emulator logo

Add logo file: `ui/src/assets/emulators/newemu.png` (or .svg)

File: `ui/src/components/EmulatorLogo/EmulatorLogo.tsx`

```typescript
import newemu from "@/assets/emulators/newemu.png";

const EMULATOR_LOGOS: Partial<Record<EmulatorID, string>> = {
  // ... existing logos ...
  "retroarch:newcore": newemu, // For RetroArch cores
  newemu: newemu, // For standalone emulators
};
```

Note: `EMULATOR_LOGOS` is `Partial` because not all emulators need custom logos.

### Sources for logos

- ES-DE themes: `https://gitlab.com/es-de/themes/slate-es-de`
- EmuDeck: `https://www.emudeck.com/logos/`
- Emulator project repositories
- Libretro thumbnails: `https://github.com/libretro-thumbnails/`

## Step 9: Update registry tests

File: `internal/registry/registry_test.go`

Update the following test cases to include the new system/emulator:

- `TestAllDefinitions`: Add system and emulator definitions to the slices
- `TestRegistryGetEmulator`: Add emulator ID to the test cases
- `TestRegistryGetEmulatorsForSystem`: Add system with expected emulators
- `TestRegistryGetDefaultEmulator`: Add system with expected default emulator
- `TestAllSystems`: Add system to expected list (if count check needs updating)
- `TestGetConfigGenerator`: Add emulator ID to test cases

## Step 10: Update site documentation

File: `site/src/content/docs/using-the-app.mdx`

Add a row to the emulator support table:

```markdown
| Emulator | System | Syncs | Hotkeys | Controller | Players | Provisions |
| ...existing rows... |
| New Emulator | System Name | saves, states, screenshots | full | configured | 4 | none |
```

Column definitions:

- **Syncs**: `saves` (in-game), `states` (savestates), `screenshots`
- **Hotkeys**: `full` (all hotkeys), `partial` (subset), `none`
- **Controller**: `configured` (Kyaraben generates bindings), `auto` (emulator handles)
- **Players**: number of controller slots configured, or `-` if emulator handles
- **Provisions**: BIOS/firmware needed, or `none`. Add `(optional)` if not required

## Step 11: Generate types and test

```bash
# Generate TypeScript types from Go models
just generate-types

# Run checks
just check

# Test specific functionality
go test ./internal/emulators/retroarchnewcore/...
go test ./internal/systems/newsystem/...
```

## Step 12: Manual testing checklist

- [ ] System appears in the system picker UI
- [ ] System logo displays correctly
- [ ] Selecting system triggers emulator/core download
- [ ] ROMs in the system folder are detected
- [ ] Games launch successfully
- [ ] Controller input works
- [ ] Hotkeys work (save state, load state, fast forward, menu)
- [ ] Save states create and load correctly
- [ ] In-game saves work
- [ ] BIOS validation works (if applicable)
- [ ] System appears correctly in ES-DE after sync

## Controller configuration

### RetroArch cores

RetroArch cores use shared controller configuration via `retroarch.CorePatches()`. This handles:

- Button mapping based on layout (Standard vs Nintendo)
- Hotkey configuration
- Analog stick configuration

No additional work needed unless the core has special requirements.

### Standalone emulators

Each standalone emulator needs custom controller mapping in its `ConfigGenerator.Generate()` method. See existing implementations:

- INI format: `internal/emulators/duckstation/duckstation.go`
- Custom format: `internal/emulators/dolphin/dolphin.go`
- YAML format: `internal/emulators/rpcs3/rpcs3.go`

## SupportedSettings introspection

Each emulator declares which user-configurable features it supports via `SupportedSettings`. The UI uses this to show or hide settings per emulator.

Available settings:

| Constant                      | UI effect                                  |
| ----------------------------- | ------------------------------------------ |
| `model.SettingPreset`         | Show graphics preset selector (clean/retro)|
| `model.SettingResumeAutosave` | Show "auto-save on exit" toggle            |
| `model.SettingResumeAutoload` | Show "auto-load on start" toggle           |

Example:

```go
SupportedSettings: []string{
    model.SettingPreset,
    model.SettingResumeAutosave,
    model.SettingResumeAutoload,
},
```

Not all emulators support all features:

- RetroArch cores: typically all three
- DuckStation/PCSX2: preset + autosave (no autoload)
- PPSSPP: preset + autoload (no autosave)
- Flycast: autosave + autoload (no preset - 6th gen system)
- RPCS3/Vita3K/Cemu: none (limited config integration)

## ConfigInput and DependsOn

Config entries must declare their dependencies for proper diff detection. When a dependency changes, entries depending on it are reapplied.

### Dependency constants

```go
model.None     // Static value, never changes
model.Store    // Depends on collection path
model.Nintendo // Depends on controller layout (Standard vs Nintendo)
model.Hotkeys  // Depends on hotkey configuration
model.Preset   // Depends on graphics preset
model.Resume   // Depends on resume settings
```

### Usage

```go
// Static entry (never needs reapply)
model.Entry(model.None, model.Path("video_driver"), "vulkan")

// Depends on collection path
model.Entry(model.Store, model.Path("savefile_directory"), store.SavesDir())

// Depends on graphics preset
model.Entry(model.Preset, model.Path("video_shader_enable"), "true")

// Depends on controller layout
model.Entry(model.Nintendo, model.Path("input_b_btn"), southButton)

// DefaultOnly with dependency
model.Default(model.None, model.Path("menu_driver"), "ozone")
```

The diff system tracks `ConfigInputsWhenWritten` in the manifest. If the input value changes (e.g., user switches from Standard to Nintendo layout), entries with that dependency are flagged for reapply.

## PresetConfig for graphics

RetroArch cores use `PresetConfig` to handle graphics presets and resume settings.

### PresetConfig fields

```go
type PresetConfig struct {
    Preset             string                           // "clean" or "retro"
    TargetDevice       string                           // e.g., "steamdeck-lcd"
    Resume             string                           // "on" or "off"
    SystemDisplayTypes map[model.SystemID]model.DisplayType // CRT vs LCD per system
}
```

### Graphics presets

- `model.PresetClean`: no shaders, bilinear filtering off
- `model.PresetRetro`: CRT/LCD shaders based on system type

### Display types

5th gen and earlier systems have display type metadata:

- `model.DisplayTypeCRT`: home consoles (NES, SNES, Genesis) - use CRT shader
- `model.DisplayTypeLCD`: handhelds (Game Boy, GBA) - use LCD shader

6th gen and newer (GameCube, PS2, Dreamcast) always use clean display regardless of preset.

### Shader downloads

RetroArch cores that support presets need shader downloads:

```go
downloads, err := retroarch.CoreShaderDownloads(emuID, resolver, pc)
embedded, err := retroarch.CoreEmbeddedFiles(emuID, systems, pc, resolver)
```

These download shaders from libretro's slang-shaders repo and generate preset files.

## Common patterns

### Multi-system emulator

One emulator can support multiple systems:

```go
Systems: []model.SystemID{
    model.SystemIDSystem1,
    model.SystemIDSystem2,
},
```

Example: Genesis Plus GX supports Genesis, Master System, and Game Gear.

### Optional BIOS with HLE fallback

```go
ProvisionGroups: []model.ProvisionGroup{{
    MinRequired: 0,  // Optional
    Message:     "BIOS improves compatibility but is not required",
    Provisions:  biosProvisions,
}},
```

### Region-specific BIOS variants

```go
var biosProvisions = []model.Provision{
    model.HashedProvision(model.ProvisionBIOS, "bios_us.bin", "US", []string{"hash1"}),
    model.HashedProvision(model.ProvisionBIOS, "bios_eu.bin", "Europe", []string{"hash2"}),
    model.HashedProvision(model.ProvisionBIOS, "bios_jp.bin", "Japan", []string{"hash3"}),
}
```

### Custom BIOS directory

```go
ProvisionGroups: []model.ProvisionGroup{{
    MinRequired: 1,
    Message:     "BIOS required",
    Provisions:  biosProvisions,
    BaseDir: func(resolver model.BaseDirResolver, store model.Store) (string, error) {
        return filepath.Join(store.BiosDir, "subdir"), nil
    },
}},
```

## Reference implementations

Study these for patterns:

| Type                  | Example                  | Notes                          |
| --------------------- | ------------------------ | ------------------------------ |
| Simple RetroArch core | `retroarchbsnes`         | No BIOS, basic config          |
| RetroArch with BIOS   | `retroarchbeetlesaturn`  | Multiple BIOS variants         |
| Multi-system core     | `retroarchgenesisplusgx` | 3 systems, optional BIOS       |
| Standalone simple     | `flycast`                | Basic standalone setup         |
| Standalone with BIOS  | `duckstation`            | Many BIOS variants, INI config |
| Complex standalone    | `rpcs3`                  | Import-based BIOS, YAML config |

## Troubleshooting

### Core not downloading

- Check `versions.toml` has correct filename in `[retroarch-cores.files]`
- Verify core exists in buildbot bundle

### System not appearing in UI

- Check `just generate-types` ran successfully
- Verify registry imports and registration
- Check ESDE system mapping exists

### Controller not working

- Verify `retroarch.CorePatches()` is called in Generate()
- Check RetroArch autoconfig profiles exist for the controller

### BIOS not detected

- Verify provision filename matches exactly
- Check MD5 hash if using HashedProvision
- Verify BIOS directory path is correct

---

## Standalone emulator patterns

The following sections document advanced patterns used by standalone emulators. RetroArch cores typically don't need these.

### Config targets and formats

Standalone emulators write to one or more config files. Each target specifies:

```go
model.ConfigTarget{
    Path:    "emulator/settings.ini",
    BaseDir: model.ConfigBaseDirUserConfig,  // ~/.config on Linux
    Format:  model.ConfigFormatINI,
}
```

#### Base directories

- `ConfigBaseDirUserConfig`: `~/.config` (most emulators)
- `ConfigBaseDirUserData`: `~/.local/share` (DuckStation, xemu)

#### Config formats

- `ConfigFormatINI`: Key-value pairs with sections (most common)
- `ConfigFormatYAML`: YAML files (RPCS3, Vita3K)
- `ConfigFormatXML`: XML files (Cemu settings)
- `ConfigFormatTOML`: TOML files (xemu, Xenia)
- `ConfigFormatRaw`: Write exact string without parsing (complex YAML, XML profiles)

Use Raw format when the config has complex nested structures that the config system can't patch incrementally.

### Multiple config files

Many emulators need multiple config files:

```go
// Main config (user can customize after first run)
mainTarget := model.ConfigTarget{
    Path:    "emulator/settings.ini",
    BaseDir: model.ConfigBaseDirUserConfig,
    Format:  model.ConfigFormatINI,
}

// Controller profile (fully managed by Kyaraben)
profileTarget := model.ConfigTarget{
    Path:    "emulator/profiles/Kyaraben.ini",
    BaseDir: model.ConfigBaseDirUserConfig,
    Format:  model.ConfigFormatINI,
}
```

Examples:

- DuckStation: settings.ini + inputprofiles/kyaraben-steamdeck.ini
- Dolphin: Dolphin.ini + GFX.ini + GCPadNew.ini + Profiles/GCPad/Kyaraben.ini
- RPCS3: vfs.yml + GuiConfigs/CurrentSettings.ini + input_configs/

### DefaultOnly

Apply a config entry only on first run. Users can customize afterward and Kyaraben won't overwrite:

```go
model.ConfigEntry{
    Section:     "Controller",
    Key:         "ButtonA",
    Value:       "Button0",
    DefaultOnly: true,
}
```

Use for settings users might want to customize (controller bindings in main config, graphics options).

### ManagedRegions

Mark a region of the config file as fully managed by Kyaraben. Changes can be reapplied:

```go
model.ConfigEntry{
    Section: "Controller",
    Key:     "ButtonA",
    Value:   "Button0",
    ManagedRegions: []model.ManagedRegion{model.FileRegion{}},
}
```

`FileRegion{}` marks the entire file as managed. Use for separate profile files that Kyaraben generates entirely (e.g., controller profiles that can be regenerated when controller config changes).

### Save and state directories

The preferred approach is to configure the emulator to use Kyaraben's directories directly via config patches:

```go
model.ConfigEntry{
    Section: "Paths",
    Key:     "SaveDirectory",
    Value:   store.SystemSavesDir(model.SystemIDPSX),
},
model.ConfigEntry{
    Section: "Paths",
    Key:     "SaveStateDirectory",
    Value:   store.SystemSavestatesDir(model.SystemIDPSX),
},
```

### Symlinks (last resort)

Symlinks are only needed when:

- The emulator can't be configured to use different paths
- The emulator uses opaque directory structures

```go
symlinks := []model.Symlink{
    {
        // Emulator expects saves here (can't be configured)
        Link: filepath.Join(configDir, "emulator/saves"),
        // Kyaraben manages saves here
        Target: store.SystemSavesDir(model.SystemIDPSX),
    },
}
```

Direction: Link is where the emulator writes, Target is Kyaraben's managed directory.

Multi-system example (Dolphin):

```go
// GameCube and Wii have separate save directories
{Link: "dolphin-emu/GC/", Target: store.SystemSavesDir(model.SystemIDGameCube)},
{Link: "dolphin-emu/Wii/", Target: store.SystemSavesDir(model.SystemIDWii)},
```

### InitialDownloads

Some emulators need large files downloaded on first run (not BIOS):

```go
return model.GenerateResult{
    Patches:  patches,
    Symlinks: symlinks,
    InitialDownloads: []model.InitialDownload{{
        URL:      "https://example.com/required_file.qcow2",
        SHA256:   "sha256-...",
        DestPath: filepath.Join(dataDir, "emulator/file.qcow2"),
    }},
}, nil
```

Example: xemu downloads xbox_hdd.qcow2 virtual hard drive image.

### ImportViaUI provisions (last resort)

Prefer directory-based provisions where users place files in a folder that Kyaraben manages. ImportViaUI is only for emulators that have no directory-based option:

```go
model.Provision{
    Kind:        model.ProvisionFirmware,
    Filename:    "PS3UPDAT.PUP",
    Description: "PS3 firmware",
    ImportViaUI: true,
    ImportStrategy: model.ImportStrategy{
        Instructions: "File > Install Firmware > Select PUP file",
    },
}
```

Example: Eden used to require ImportViaUI for firmware, but a directory-based approach was found and is now preferred (users place files in a symlinked directory).

Used by: RPCS3 (firmware has no directory option)

### Custom provision BaseDir

Provisions can go to non-standard directories:

```go
ProvisionGroups: []model.ProvisionGroup{{
    MinRequired: 1,
    Message:     "GameCube IPL required",
    Provisions:  iplProvisions,
    BaseDir: func(store model.StoreReader, sys model.SystemID) string {
        // Goes to saves dir, not BIOS dir
        return filepath.Join(store.SystemSavesDir(model.SystemIDGameCube), "USA")
    },
}},
```

Used by: Dolphin (IPL files per region), RPCS3 (config dir)

### Controller layout awareness

Get face buttons adjusted for controller layout (Standard vs Nintendo):

```go
south, east, west, north := cc.FaceButtons()
// Standard layout: A=south, B=east, X=west, Y=north
// Nintendo layout: A=east, B=south, X=north, Y=west
```

Always use semantic positions, not fixed SDL indices:

```go
// Correct: uses layout-aware buttons
fmt.Sprintf("Cross = %s", south)
fmt.Sprintf("Circle = %s", east)

// Wrong: hardcoded indices break on Nintendo layout
fmt.Sprintf("Cross = Button0")
```

### Hotkey binding formats

Each emulator has its own hotkey format:

| Emulator          | Format                    | Example                     |
| ----------------- | ------------------------- | --------------------------- |
| DuckStation/PCSX2 | `btn & btn`               | `SDL-0/Back & SDL-0/DPadUp` |
| Dolphin           | `@(btn+btn)`              | `@(`Back`+`DPad Up`)`       |
| PPSSPP            | `device-code:device-code` | `10-4001:10-4002`           |
| Flycast           | `btn,btn:action`          | `6,4:0:sequential`          |
| Eden              | `btn+btn`                 | `Plus+Minus+A`              |

### GUID-based controller binding

Some emulators embed the controller GUID in every binding:

```go
// Eden format with GUID
fmt.Sprintf("engine:sdl,port:0,guid:%s,button:%d", model.SteamDeckGUID, buttonIndex)

// xemu/Cemu use similar patterns
```

`model.SteamDeckGUID` provides the Steam Deck controller GUID.

### Environment variables

Some emulators need environment variables:

```go
Launcher: model.LauncherInfo{
    Env: map[string]string{
        "GDK_BACKEND": "x11",  // Required for xemu, Xenia on Steam Deck
    },
    RomCommand: func(opts model.RomLaunchOptions) string {
        return fmt.Sprintf("%s %s", opts.BinaryPath, "%ROM%")
    },
},
```

### Complex launch commands

Some emulators need conditional launch logic:

```go
// Vita3K: handle .psvita manifest files differently
RomCommand: func(opts model.RomLaunchOptions) string {
    return fmt.Sprintf(`sh -c 'case %%ROM%% in *.psvita) %s -r "$(cat %%ROM%%)" ;; *) %s %%ROM%% ;; esac'`,
        opts.BinaryPath, opts.BinaryPath)
},
```

### Custom equality for config values

Some emulators reorder config keys non-deterministically. Use custom equality to prevent spurious conflicts:

```go
model.ConfigEntry{
    Key:          "input_bindings",
    Value:        bindingValue,
    EqualityFunc: configformat.BindingValuesEqual,  // Semantic comparison
}
```

### Standalone emulator reference

| Emulator    | Config format | Multi-file    | Paths via | Special                      |
| ----------- | ------------- | ------------- | --------- | ---------------------------- |
| DuckStation | INI           | Yes (profile) | Config    | -                            |
| PCSX2       | INI           | Yes (profile) | Config    | -                            |
| RPCS3       | YAML + Raw    | Yes (4 files) | Symlinks  | ImportViaUI                  |
| Dolphin     | INI           | Yes (4 files) | Symlinks  | Region BaseDir               |
| PPSSPP      | INI           | No            | Symlinks  | Raw joystick                 |
| Flycast     | INI           | No            | Config    | Combo hotkeys                |
| Cemu        | XML + Raw     | Yes           | Symlinks  | GUID bindings                |
| Vita3K      | YAML + Raw    | Yes           | Symlinks  | ImportViaUI, manifest launch |
| Eden        | INI           | Yes (profile) | Symlinks  | GUID bindings, ImportViaUI   |
| xemu        | TOML          | No            | Config    | InitialDownloads, env vars   |
| Xenia       | TOML          | No            | Config    | env vars                     |
