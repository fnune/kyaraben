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

#### Multiplayer behavior

**Eden** (yuzu-based) supports per-player controller profiles. The config uses `player_0_*`, `player_1_*`, etc. keys, each with its own GUID and port. Two players can use different physical controllers simultaneously. Kyaraben generates bindings for each known GUID per player slot.

**Azahar** (Citra-based) is single-player only (the 3DS is a handheld). There is one active profile at a time. Multiplayer requires separate emulator instances. Kyaraben generates a single profile.

#### GUID mapping design

Kyaraben ships a built-in GUID-to-profile mapping in Go code. Users extend or override it in `config.toml`. At runtime, user mappings are merged over built-in mappings (user wins on conflicts).

Built-in mapping covers:

| Profile | GUIDs |
|---------|-------|
| Steam Deck | `03000000de280000ff11000001000000` (Neptune) |
| Xbox | `030000005e0400008e02000010010000` (360 wired), `030000005e0400009102000007010000` (360 wireless), `030000005e040000ea02000001030000` (One), `030000005e040000d102000001010000` (One S), `030000005e040000130b000011050000` (Series X/S) |
| PlayStation | `030000004c0500006802000011010000` (DS3), `030000004c050000c405000011010000` (DS4 v1), `030000004c050000cc09000011010000` (DS4 v2), `030000004c050000e60c000011010000` (DualSense) |
| Nintendo | `030000007e0500000920000011010000` (Switch Pro), `03000000c82d00000161000000010000` (8BitDo SN30 Pro) |
| Handheld (Xbox-compatible) | `0300000005b000004c1b000000000000` (ROG Ally X), `030000007eef00008261000000000000` (Legion Go), `03000000861a000010e3000000000000` (Legion Go S), `03000000b00d000001190000000000000` (MSI Claw), `030000006325000058d0000000000000` (OneXPlayer) |

Many handheld devices (ROG Ally original, AYANEO, GPD Win 4/Max 2) use the generic Microsoft Xbox 360 VID/PID (`045e:028e`), so they already match the Xbox profile. Devices like the ROG Ally X, Legion Go, and MSI Claw use their own VID/PID but behave as Xbox 360 controllers, so they map to the Xbox profile.

The user extends this via `config.toml`:

```toml
[controller.guids]
# Map an unsupported controller GUID to a known profile
"030000001234567890abcdef01000000" = "xbox"
```

This means users with uncommon controllers can self-serve without waiting for a Kyaraben update.

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
    GUIDs   map[string]ProfileID  // merged: built-in + user overrides
}

type LayoutID string

const (
    LayoutStandard LayoutID = "standard"  // A=bottom (Xbox, PlayStation)
    LayoutNintendo LayoutID = "nintendo"  // A=right, B=bottom
)

type ProfileID string

const (
    ProfileSteamDeck ProfileID = "steam-deck"
    ProfileXbox      ProfileID = "xbox"
    ProfilePS        ProfileID = "playstation"
    ProfileNintendo  ProfileID = "nintendo"
)

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

### GUID-to-profile mapping

Built-in mapping lives in Go code:

```go
var BuiltinGUIDs = map[string]ProfileID{
    "03000000de280000ff11000001000000": ProfileSteamDeck,
    "030000005e0400008e02000010010000": ProfileXbox,
    // ... all GUIDs from the table above
}
```

At apply time, the apply flow builds `ControllerConfig.GUIDs` by starting with `BuiltinGUIDs` and overlaying the user's `[controller.guids]` from config.toml. User entries win on conflicts.

This means Group C emulators (Eden, Azahar) iterate over the merged GUID map and generate bindings for each GUID that maps to the relevant profile.

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

[controller.guids]
# Users add custom GUID-to-profile mappings here.
# These override built-in mappings for the same GUID.
# Supported profiles: "steam-deck", "xbox", "playstation", "nintendo"
# Example:
# "030000001234567890abcdef01000000" = "xbox"
```

When `[controller]` is absent from an existing config.toml, Kyaraben uses the defaults (standard layout, default hotkeys, no user GUID overrides).

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
- Controllers not in built-in GUID list and not added via `[controller.guids]` (configure through emulator UI)

## Decisions

### GUID mapping: built-in defaults with user overrides

The built-in GUID-to-profile mapping lives in Go code. Users extend or override it via `[controller.guids]` in config.toml. At runtime, user entries merge over built-in entries (user wins on conflicts).

This means:
- Common controllers work out of the box
- Users with uncommon controllers can self-serve by adding their GUID
- No need to wait for a Kyaraben update

Config file size for GUID-based emulators (Eden, Azahar): we generate bindings for each GUID in the merged map. With ~15 built-in GUIDs x ~20 bindings = ~300 config lines per emulator. Acceptable tradeoff.

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
