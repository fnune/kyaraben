# Controller configuration planning document

## Overview

Kyaraben manages emulator configs. Controller support means writing input bindings to those configs so emulators work out-of-the-box with gamepads and keyboard.

## The simplified model

```
Physical controller
       |
SDL GameControllerDB (maps GUID to standard names)
       |
Standard SDL names: a, b, x, y, leftshoulder, leftx, etc.
       |
Emulator config (uses SDL names in its own syntax)
       |
Works with any supported controller
```

We write ONE config per emulator using SDL's standard button names. SDL handles device detection at runtime. No per-controller configs needed for most emulators.

## Emulator investigation

### Group A: Direct SDL standard names

These emulators reference SDL's standard button names. One config works with any controller.

| Emulator | Config file | Section | Example |
|----------|-------------|---------|---------|
| DuckStation | `~/.config/duckstation/settings.ini` | `[Pad1]` | `Cross = SDL-0/A` |
| PCSX2 | `~/.config/PCSX2/inis/PCSX2.ini` | `[Pad1]` | `Cross = SDL-0/A` |
| Dolphin | `~/.config/dolphin-emu/GCPadNew.ini` | `[GCPad1]` | `` Buttons/A = `Button S` `` |

**What Kyaraben does**: Write config entries using SDL standard names. Format varies per emulator but the concept is identical.

**Coverage**: 100% of controllers in SDL GameControllerDB.

#### DuckStation/PCSX2 format

```ini
[Pad1]
Type = AnalogController
Cross = SDL-0/A
Circle = SDL-0/B
Triangle = SDL-0/Y
Square = SDL-0/X
L1 = SDL-0/LeftShoulder
R1 = SDL-0/RightShoulder
L2 = SDL-0/+LeftTrigger
R2 = SDL-0/+RightTrigger
L3 = SDL-0/LeftStick
R3 = SDL-0/RightStick
Up = SDL-0/DPadUp
Down = SDL-0/DPadDown
Left = SDL-0/DPadLeft
Right = SDL-0/DPadRight
LLeft = SDL-0/-LeftX
LRight = SDL-0/+LeftX
LUp = SDL-0/-LeftY
LDown = SDL-0/+LeftY
RLeft = SDL-0/-RightX
RRight = SDL-0/+RightX
RUp = SDL-0/-RightY
RDown = SDL-0/+RightY
Start = SDL-0/Start
Select = SDL-0/Back
SmallMotor = SDL-0/SmallMotor
LargeMotor = SDL-0/LargeMotor
```

#### Dolphin format

```ini
[GCPad1]
Device = SDL/0/Steam Deck Controller
Buttons/A = `Button S`
Buttons/B = `Button W`
Buttons/X = `Button E`
Buttons/Y = `Button N`
Buttons/Z = `Shoulder R`|Back
Buttons/Start = Start
Main Stick/Up = `Axis 1-`
Main Stick/Down = `Axis 1+`
Main Stick/Left = `Axis 0-`
Main Stick/Right = `Axis 0+`
C-Stick/Up = `Axis 4-`
C-Stick/Down = `Axis 4+`
C-Stick/Left = `Axis 3-`
C-Stick/Right = `Axis 3+`
Triggers/L = `Trigger L`
Triggers/R = `Trigger R`
Triggers/L-Analog = `Trigger L`
Triggers/R-Analog = `Trigger R`
D-Pad/Up = `Pad N`
D-Pad/Down = `Pad S`
D-Pad/Left = `Pad W`
D-Pad/Right = `Pad E`
Rumble/Motor = Strong
```

### Group B: Raw button indices

These emulators use numeric button/axis indices rather than SDL standard names.

| Emulator | Config file | Section | Example |
|----------|-------------|---------|---------|
| mGBA | `~/.config/mgba/config.ini` | `[gba.input.SDLB]` | `keyA=0` |
| melonDS | `~/.config/melonDS/melonDS.ini` | root | `Joy_A=0` |
| PPSSPP | `~/.config/ppsspp/PSP/SYSTEM/controls.ini` | `[ControlMapping]` | `Circle = 10-190` |
| Flycast | `~/.config/flycast/mappings/*.cfg` | per-device | `0:btn_a` |

**What Kyaraben does**: Write configs using standard SDL button indices. SDL GameController API presents buttons in a consistent order for known controllers.

**Coverage**: ~95% (controllers in SDL GameControllerDB).

#### mGBA format

```ini
[gba.input.SDLB]
device0=03000000de280000ff11000001000000
keyA=0
keyB=1
keySelect=6
keyStart=7
keyRight=-1
keyLeft=-1
keyUp=-1
keyDown=-1
hat0Up=6
hat0Down=7
hat0Left=5
hat0Right=4
keyL=4
keyR=5
```

#### melonDS format

```ini
Joy_A=0
Joy_B=1
Joy_Select=6
Joy_Start=7
Joy_Right=258
Joy_Left=264
Joy_Up=257
Joy_Down=260
Joy_R=5
Joy_L=4
Joy_X=2
Joy_Y=3
JoystickID=0
```

#### PPSSPP format

Uses internal keycodes. Format: `device-keycode` where device 10 is gamepad.

```ini
[ControlMapping]
Up = 10-19
Down = 10-20
Left = 10-21
Right = 10-22
Circle = 10-190
Cross = 10-189
Square = 10-191
Triangle = 10-188
Start = 10-197
Select = 10-196
L = 10-193
R = 10-192
An.Up = 10-4003
An.Down = 10-4002
An.Left = 10-4001
An.Right = 10-4000
```

#### Flycast format

Separate mapping files per device in `~/.config/flycast/mappings/`. Flycast identifies controllers by SDL device name, not GUID. Each physical controller type gets its own mapping file (e.g., `SDL_Xbox 360 Controller.cfg`). Multiple controllers of different types can be used simultaneously, each with their own mapping file. Identical controllers are distinguished by joystick instance index.

```ini
[analog]
bind0 = 0-:btn_analog_left
bind1 = 0+:btn_analog_right
bind2 = 1-:btn_analog_up
bind3 = 1+:btn_analog_down
bind4 = 2+:btn_trigger_left
bind5 = 5+:btn_trigger_right

[digital]
bind0 = 0:btn_a
bind1 = 1:btn_b
bind2 = 2:btn_x
bind3 = 3:btn_y
bind4 = 4:btn_z
bind5 = 5:btn_c
bind6 = 6:btn_menu
bind7 = 7:btn_start

[emulator]
dead_zone = 10
mapping_name = controller_neptune
```

### Group C: GUID-based bindings

These emulators embed the controller's hardware GUID in each binding.

| Emulator | Config file | GUID required? | Format |
|----------|-------------|----------------|--------|
| Eden | `~/.local/share/eden/config/qt-config.ini` | Yes | `engine:sdl,guid:...,button:1` |
| Azahar | `~/.config/azahar-emu/qt-config.ini` | Yes | `button:1,guid:...,engine:sdl` |

Ryujinx is unsupported by Kyaraben; only Eden handles Switch emulation.

#### Key finding: Steam Input virtualizes GUIDs

On Steam Deck in game mode, Steam Input intercepts all controllers and presents them with the same virtual GUID regardless of the physical controller:

```
Steam Deck GUID: 03000000de280000ff11000001000000
```

This means any controller connected to a Steam Deck (Xbox, PlayStation, 8BitDo, etc.) appears to emulators as the Steam Deck controller. EmuDeck exploits this: they ship configs with the Steam Deck GUID and it works for all controllers.

Neither Eden nor Azahar auto-detects controller GUIDs at runtime. Eden loads profiles by name from `~/.config/eden/input/`, not by GUID matching. Azahar stores profiles inline in `qt-config.ini` and uses a static index to select the active one. Neither emulator switches profiles when a new controller is connected. EmuDeck and RetroDECK both rely entirely on Steam Input to virtualize all controllers to the Steam Deck GUID. Kyaraben follows the same approach.

#### Multiplayer behavior

**Eden** (yuzu-based) supports per-player controller profiles. The config uses `player_0_*`, `player_1_*`, etc. keys, each with its own GUID and port. Two players can use different physical controllers simultaneously. Kyaraben generates bindings for two player slots using the Steam Deck GUID.

**Azahar** (Citra-based) is single-player only (the 3DS is a handheld). There is one active profile at a time. Multiplayer requires separate emulator instances. Kyaraben generates a single profile.

#### GUID approach

Kyaraben hardcodes the Steam Deck virtual gamepad GUID for Eden and Azahar configs. This covers all controllers on Steam Deck (where Steam Input virtualizes every physical controller to this GUID) and all controllers on desktop Linux when launched through Steam.

Desktop Linux users running emulators without Steam Input must configure their controller through the emulator UI.

#### Eden format

Eden uses per-player keys with GUID embedded in each binding value:

```ini
[Controls]
player_0_button_a="engine:sdl,port:0,guid:03000000de280000ff11000001000000,button:1"
player_0_button_b="engine:sdl,port:0,guid:03000000de280000ff11000001000000,button:0"
player_1_button_a="engine:sdl,port:1,guid:03000000de280000ff11000001000000,button:1"
```

#### Azahar format

Azahar uses a single profile (no player indexing):

```ini
[Controls]
profile=1
profiles\1\name=default
profiles\1\button_a="button:1,engine:sdl,guid:03000000de280000ff11000001000000,port:0"
profiles\1\button_b="button:0,engine:sdl,guid:03000000de280000ff11000001000000,port:0"
profiles\1\button_x="button:3,engine:sdl,guid:03000000de280000ff11000001000000,port:0"
profiles\1\button_y="button:2,engine:sdl,guid:03000000de280000ff11000001000000,port:0"
profiles\1\button_l="axis:2,direction:+,engine:sdl,guid:03000000de280000ff11000001000000,port:0,threshold:0.5"
profiles\1\button_r="axis:5,direction:+,engine:sdl,guid:03000000de280000ff11000001000000,port:0,threshold:0.5"
profiles\1\button_zl="button:4,engine:sdl,guid:03000000de280000ff11000001000000,port:0"
profiles\1\button_zr="button:5,engine:sdl,guid:03000000de280000ff11000001000000,port:0"
profiles\1\button_start="button:7,engine:sdl,guid:03000000de280000ff11000001000000,port:0"
profiles\1\button_select="button:6,engine:sdl,guid:03000000de280000ff11000001000000,port:0"
profiles\1\button_up="direction:up,engine:sdl,guid:03000000de280000ff11000001000000,hat:0,port:0"
profiles\1\button_down="direction:down,engine:sdl,guid:03000000de280000ff11000001000000,hat:0,port:0"
profiles\1\button_left="direction:left,engine:sdl,guid:03000000de280000ff11000001000000,hat:0,port:0"
profiles\1\button_right="direction:right,engine:sdl,guid:03000000de280000ff11000001000000,hat:0,port:0"
profiles\1\circle_pad="axis_x:0,axis_y:1,deadzone:0.100000,engine:sdl,guid:03000000de280000ff11000001000000,port:0"
profiles\1\c_stick="axis_x:3,axis_y:4,deadzone:0.100000,engine:sdl,guid:03000000de280000ff11000001000000,port:0"
```

### Group D: Auto-detection

These emulators handle controller detection automatically.

| Emulator | Notes |
|----------|-------|
| RetroArch | Extensive [joypad-autoconfig](https://github.com/libretro/retroarch-joypad-autoconfig). Just enable autodetect. |
| Vita3K | No controller config needed. Auto-detects at runtime. |
| RPCS3 | Uses `Handler: Evdev` and `Device: [name]`. Has auto-map feature. |

**What Kyaraben does**: Ensure auto-detection is enabled (usually default). No per-button config needed.

**Coverage**: 100%.

## Coverage summary

| Emulator | Systems | Approach | Coverage |
|----------|---------|----------|----------|
| DuckStation | PSX | SDL standard names | 100% |
| PCSX2 | PS2 | SDL standard names | 100% |
| Dolphin | GameCube, Wii | SDL standard names | 100% |
| RetroArch | NES, SNES, Genesis, Saturn, N64, GB/GBC/GBA | Auto-detect | 100% |
| mGBA | GBA | Raw indices | ~95% |
| melonDS | NDS | Raw indices | ~95% |
| PPSSPP | PSP | Internal keycodes | ~95% |
| Flycast | Dreamcast | Per-device mapping files | Common controllers |
| Eden | Switch | Ship GUIDs | Steam Deck 100%, desktop common controllers |
| Azahar | 3DS | Ship GUIDs | Steam Deck 100%, desktop common controllers |
| Vita3K | PS Vita | Auto-detect | 100% |
| RPCS3 | PS3 | Device name | 100% |

## Implementation

### Interface changes: GenerateContext and GenerateResult

The current `ConfigGenerator.Generate(store StoreReader)` signature cannot pass controller configuration to generators. Separately, `SymlinkProvider` and `LaunchArgsProvider` are discovered via type assertions in the apply flow. This refactor consolidates everything into a single interface with a context struct and result struct, eliminating all type assertions.

This follows the pattern already established by `FrontendContext` for frontend config generators.

```go
type GenerateContext struct {
    Store             StoreReader
    BaseDirResolver   BaseDirResolver
    ControllerConfig  *ControllerConfig  // nil when not configured
}

type GenerateResult struct {
    Patches    []ConfigPatch
    Symlinks   []SymlinkSpec
    LaunchArgs []string
}

type ConfigGenerator interface {
    Generate(ctx GenerateContext) (GenerateResult, error)
}
```

The apply flow changes from three separate discovery paths (Generate, then type-assert SymlinkProvider, then type-assert LaunchArgsProvider) to a single call:

```go
result, err := gen.Generate(ctx)
// result.Patches, result.Symlinks, result.LaunchArgs all available
```

Migration is mechanical: each emulator replaces `store StoreReader` with `ctx GenerateContext` (using `ctx.Store` where `store` was used), and wraps its `[]ConfigPatch` return in `GenerateResult{Patches: ...}`. Emulators that currently implement `SymlinkProvider` move that logic into `Generate` and populate `result.Symlinks`. Emulators that do not need symlinks or launch args leave those fields at zero value.

`SymlinkProvider` and `LaunchArgsProvider` interfaces are removed.

### Controller configuration types

```go
type ControllerConfig struct {
    Layout  LayoutID
    Hotkeys HotkeyConfig
}

type LayoutID string

const (
    LayoutStandard LayoutID = "standard"  // A=bottom (Xbox, PlayStation)
    LayoutNintendo LayoutID = "nintendo"  // A=right, B=bottom
)

const SteamDeckGUID = "03000000de280000ff11000001000000"

type HotkeyConfig struct {
    SaveState        HotkeyBinding
    LoadState        HotkeyBinding
    NextSlot         HotkeyBinding
    PrevSlot         HotkeyBinding
    FastForward      HotkeyBinding
    Rewind           HotkeyBinding
    Pause            HotkeyBinding
    Screenshot       HotkeyBinding
    Quit             HotkeyBinding
    ToggleFullscreen HotkeyBinding
    OpenMenu         HotkeyBinding
}
```

`HotkeyBinding` is a parsed, validated type (not a raw string). It is parsed from the TOML string representation (e.g., `"Back+RightShoulder"`) at config load time. Invalid button names cause an error at load time, not at apply time.

### Keyboard bindings

Kyaraben also generates keyboard bindings for emulators that support them. Keyboard bindings use a fixed, sensible default layout that is not user-configurable in v1 (the controller hotkeys are configurable; keyboard hotkeys use the same semantic actions but with hardcoded key assignments per emulator).

Existing keyboard bindings in emulator configs are overwritten by Kyaraben (they are managed entries). This is consistent with how Kyaraben manages other config entries.

### Runtime generation

Do NOT ship static config files per controller/emulator combination. Instead:

1. Define profiles as data structures (button layout + built-in GUID map)
2. Define per-emulator format generators
3. Generate configs at runtime during `apply`

Each emulator's `Generate` method checks `ctx.ControllerConfig` for nil. When present, it appends controller and hotkey entries to its config patches alongside path entries.

```go
func (c *Config) Generate(ctx model.GenerateContext) (model.GenerateResult, error) {
    entries := []model.ConfigEntry{
        // Path entries (existing)
        {Path: []string{"BIOS", "SearchDirectory"}, Value: ctx.Store.SystemBiosDir(model.SystemIDPSX)},
        // ...
    }

    if cc := ctx.ControllerConfig; cc != nil {
        entries = append(entries, generatePadEntries(cc)...)
        entries = append(entries, generateHotkeyEntries(cc)...)
    }

    return model.GenerateResult{
        Patches: []model.ConfigPatch{{Target: configTarget, Entries: entries}},
    }, nil
}
```

### GUID handling for Group C emulators

Eden and Azahar use the `SteamDeckGUID` constant directly. No GUID-to-profile mapping is needed because Steam Input virtualizes all controllers to a single GUID. The `SteamDeckGUID` is embedded in every binding string for these emulators.

### Multi-player

For emulators that support multiplayer (DuckStation, PCSX2, Dolphin, Eden):

- Generate configs for players 1-4
- Same layout for all players (global preference)
- Hotkeys only on Player 1
- Players 2-4 use SDL indices 1, 2, 3

For single-player emulators (Azahar, mGBA, melonDS, PPSSPP): generate a single player config.

### Apply flow

During `apply`:
1. Load `config.toml` including `[controller]` section
2. Build `ControllerConfig` (merge built-in GUIDs with user overrides, parse hotkeys)
3. Construct `GenerateContext` with `Store`, `BaseDirResolver`, and `ControllerConfig`
4. For each enabled emulator, call `gen.Generate(ctx)`
5. Collect `result.Patches`, `result.Symlinks`, `result.LaunchArgs`
6. Apply patches, create symlinks, build desktop entries (existing flow, but from a single result instead of three separate code paths)

## User configuration

### Source of truth: config.toml

Controller settings live in Kyaraben's `config.toml`. Defaults are written explicitly by `kyaraben init`.

```toml
[controller]
layout = "standard"  # or "nintendo" (swaps A/B and X/Y)

[controller.hotkeys]
save_state = "Back+RightShoulder"
load_state = "Back+LeftShoulder"
next_slot = "Start+RightShoulder"
prev_slot = "Start+LeftShoulder"
fast_forward = "Back+RightTrigger"
rewind = "Back+LeftTrigger"
pause = "Back+A"
screenshot = "Back+B"
quit = "Back+Start"
toggle_fullscreen = "Start+LeftStick"
open_menu = "Back+RightStick"
```

When `[controller]` is absent from an existing config.toml, Kyaraben uses the defaults (standard layout, default hotkeys).

### Layout naming

- `standard`: A=bottom (Xbox, PlayStation, most controllers). The face button in the bottom position triggers the "confirm" action.
- `nintendo`: A=right, B=bottom (Nintendo controllers, some 8BitDo in Nintendo mode). The face button in the right position triggers "confirm".

### Hotkey validation

Hotkey strings are parsed into structured types at config load time. The parser validates:

- Each component is a known SDL button name (A, B, X, Y, Back, Start, LeftShoulder, RightShoulder, LeftTrigger, RightTrigger, LeftStick, RightStick, DPadUp, DPadDown, DPadLeft, DPadRight, Guide)
- The `+` separator produces exactly 1-3 components (single button or chord)
- Unknown button names produce an error with the invalid name highlighted

### Hotkey emulator coverage

Not all emulators support all hotkeys. This table documents what's supported.

| Hotkey | DuckStation | PCSX2 | Dolphin | RetroArch | mGBA | melonDS | PPSSPP | Flycast |
|--------|-------------|-------|---------|-----------|------|---------|--------|---------|
| save_state | SaveSelectedSaveState | SaveStateToSlot | Save to Selected Slot | input_save_state | saveState | - | Save State | - |
| load_state | LoadSelectedSaveState | LoadStateFromSlot | Load from Selected Slot | input_load_state | loadState | - | Load State | - |
| next_slot | SelectNextSaveStateSlot | NextSaveStateSlot | Increase Selected State Slot | input_state_slot_increase | - | - | Next Slot | - |
| prev_slot | SelectPreviousSaveStateSlot | PreviousSaveStateSlot | Decrease Selected State Slot | input_state_slot_decrease | - | - | Previous Slot | - |
| fast_forward | ToggleFastForward | ToggleTurbo | Disable Emulation Speed Limit | input_toggle_fast_forward | fastForward | HKJoy_FastForward | Fast-forward | - |
| rewind | Rewind | ToggleSlowMotion | - | input_rewind | rewind | - | Rewind | - |
| pause | TogglePause | TogglePause | Toggle Pause | input_pause_toggle | pause | HKJoy_Pause | Pause | - |
| screenshot | Screenshot | Screenshot | Take Screenshot | input_screenshot | screenshot | - | Screenshot | - |
| quit | - | ShutdownVM | Exit | input_exit_emulator | quit | - | Exit App | - |
| toggle_fullscreen | - | ToggleFullscreen | Toggle Fullscreen | input_toggle_fullscreen | fullscreen | HKJoy_FullscreenToggle | - | - |
| open_menu | OpenPauseMenu | OpenPauseMenu | - | input_menu_toggle | - | - | - | - |

Key:
- `-` = Not supported or not found in config
- Config key names shown for reference

If an emulator doesn't support a configured hotkey (e.g., rewind on Dolphin), skip silently.

### Emulator-specific hotkeys (not in Kyaraben scope)

These exist but are left for users to configure manually:
- **Dolphin**: Wii Remote connect, controller profile switch, aspect ratio toggle
- **melonDS**: Swap screens, lid close/open, microphone
- **PPSSPP**: Analog limiter, texture dump
- **DuckStation**: Change disc, toggle PGXP

### UI integration

Kyaraben UI exposes:
- Layout toggle: standard / nintendo
- Hotkey configuration with button picker
- Custom GUID mapping
- Changes write to `config.toml`
- Applied during normal `apply` flow

## Steam Input enhancement (optional)

Steam Input operates above SDL:

```
Physical controller -> Steam Input -> Virtual Xbox controller -> SDL -> Emulator
```

**For basic gameplay**: Steam Input is transparent. Our SDL configs work.

**Steam Input adds value for**:
- Steam Deck trackpads
- Back buttons (L4, L5, R4, R5)
- Universal hotkeys (quit, save state, etc.)
- Radial menus

This is an optional enhancement, not required for controllers to work.

## What Kyaraben explicitly leaves to users

- Gyro configuration
- Touchpad configuration
- Per-game profiles
- Turbo/rapid-fire
- Macro recording
- Advanced deadzone curves
- Controllers on desktop Linux without Steam (configure through emulator UI)

## Decisions

### GUID approach: Steam Deck GUID only

Eden and Azahar configs use the Steam Deck virtual gamepad GUID (`03000000de280000ff11000001000000`). Steam Input virtualizes all physical controllers to this GUID on Steam Deck and when launched through Steam on desktop Linux. This matches what EmuDeck and RetroDECK do.

Neither Eden nor Azahar auto-selects profiles based on connected controller GUIDs, so shipping per-controller-model profiles would not improve the out-of-the-box experience. Desktop Linux users running emulators outside Steam configure their controller through the emulator UI.

### Multi-player: context-dependent

- **Emulators with per-player support** (DuckStation, PCSX2, Dolphin, Eden): generate 4 player slots
- **Single-player emulators** (Azahar, mGBA, melonDS, PPSSPP): generate 1 player config
- Same layout for all players (global preference)
- Hotkeys only on Player 1
- Players 2-4 use SDL indices 1, 2, 3

### Emulator-specific hotkeys

Document what exists, leave configuration to users. Kyaraben manages the universal set only.

### Interface consolidation

`ConfigGenerator` gains a rich context (`GenerateContext`) and returns a consolidated result (`GenerateResult`). `SymlinkProvider` and `LaunchArgsProvider` are removed. No type assertions in the apply flow.

## Testing strategy

### Unit tests

Test generation logic produces correct config entries:

```go
func TestDuckStationGenerateController(t *testing.T) {
    gen := duckstation.NewControllerGenerator()
    profile := profiles.Xbox
    hotkeys := config.DefaultHotkeys()

    entries := gen.Generate(profile, hotkeys)

    assertEntry(t, entries, "Pad1", "Cross", "SDL-0/A")
    assertEntry(t, entries, "Pad1", "Circle", "SDL-0/B")
    assertEntry(t, entries, "Pad2", "Cross", "SDL-1/A")

    assertEntry(t, entries, "Hotkeys", "SaveSelectedSaveState", "SDL-0/Back & SDL-0/RightShoulder")
}

func TestNintendoLayoutSwapsButtons(t *testing.T) {
    gen := duckstation.NewControllerGenerator()
    profile := profiles.Nintendo

    entries := gen.Generate(profile, config.DefaultHotkeys())

    assertEntry(t, entries, "Pad1", "Cross", "SDL-0/B")
    assertEntry(t, entries, "Pad1", "Circle", "SDL-0/A")
}
```

### Snapshot tests

Compare generated configs against known-good baselines (bootstrapped from EmuDeck):

```go
func TestDuckStationSnapshot(t *testing.T) {
    gen := duckstation.NewControllerGenerator()
    entries := gen.Generate(profiles.SteamDeck, config.DefaultHotkeys())

    got := renderINI(entries)
    want := readTestData("testdata/duckstation_controller_snapshot.ini")

    if diff := cmp.Diff(want, got); diff != "" {
        t.Errorf("config mismatch (-want +got):\n%s", diff)
    }
}
```

Update snapshots intentionally when changing generation logic.

### Format validation

Verify generated configs parse correctly:

```go
func TestDuckStationConfigParses(t *testing.T) {
    gen := duckstation.NewControllerGenerator()
    entries := gen.Generate(profiles.Xbox, config.DefaultHotkeys())

    path := writeConfig(t, entries, model.ConfigFormatINI)

    handler := configformat.NewHandler(fs, model.ConfigFormatINI)
    parsed, err := handler.Read(path)
    require.NoError(t, err)

    assert.Contains(t, parsed, "Pad1")
    assert.Contains(t, parsed, "Pad2")
    assert.Contains(t, parsed, "Hotkeys")
}
```

### Hotkey validation tests

```go
func TestHotkeyBindingParsing(t *testing.T) {
    _, err := ParseHotkeyBinding("Back+RightShoulder")
    require.NoError(t, err)

    _, err = ParseHotkeyBinding("Bck+RightShoulder")
    require.Error(t, err)
    assert.Contains(t, err.Error(), "Bck")
}
```

## Future plans

### Keyboard bindings

Generate keyboard bindings for emulators alongside gamepad bindings. A fixed, sensible default keyboard layout would allow users without a controller to play immediately. This is lower priority than gamepad support but straightforward to implement since keyboard bindings do not depend on GUIDs or device detection.

### User-selectable controller model

Allow users to select their controller model in kyaraben's config or UI. Kyaraben would then generate emulator configs with the matching GUID for Eden and Azahar. This would improve the desktop-without-Steam experience by eliminating manual emulator configuration. The implementation would involve:

1. A `[controller.device]` setting in config.toml (e.g., `device = "xbox-series-x"`)
2. A mapping from device names to SDL GUIDs
3. Eden and Azahar generators reading the selected device's GUID instead of hardcoding the Steam Deck GUID

This is only useful for Eden and Azahar (Group C emulators). All other emulators use SDL standard names or auto-detection and work with any controller regardless of GUID.

## Sources

- [SDL GameControllerDB](https://github.com/mdqinc/SDL_GameControllerDB)
- [RetroArch joypad-autoconfig](https://github.com/libretro/retroarch-joypad-autoconfig)
- [Eden controller profiles](https://git.eden-emu.dev/eden-emu/eden/src/branch/master/docs/user/ControllerProfiles.md)
- [Citra config.cpp](https://github.com/citra-emu/citra/blob/a0f9c795c820358a825babd06a8697f8c9316b62/src/citra_qt/configuration/config.cpp) (Azahar inherits this format)
- [RPCS3 controller configuration](https://wiki.rpcs3.net/index.php?title=Help:Controller_Configuration)
- [Vita3K quickstart](https://vita3k.org/quickstart.html)
- [Steam Input documentation](https://partner.steamgames.com/doc/features/steam_controller)
- [Flycast gamepad_device.cpp](https://github.com/flyinghead/flycast/blob/master/core/input/gamepad_device.cpp)
- [Linux xpad.c driver](https://github.com/torvalds/linux/blob/master/drivers/input/joystick/xpad.c)
- [SDL handheld device support issue](https://github.com/libsdl-org/SDL/issues/10564)
- EmuDeck configs (GPL-3) for format reference
- RetroDECK configs (GPL-3) for format reference
