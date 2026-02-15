# Controller configuration planning document

## Overview

Kyaraben manages emulator configs. Controller support means writing input bindings to those configs so emulators work out-of-the-box with gamepads.

## The simplified model

```
Physical controller
       ↓
SDL GameControllerDB (maps GUID → standard names)
       ↓
Standard SDL names: a, b, x, y, leftshoulder, leftx, etc.
       ↓
Emulator config (uses SDL names in its own syntax)
       ↓
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

Separate mapping files per device in `~/.config/flycast/mappings/`.

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

### Group C: GUID-based bindings (investigation results)

These emulators embed the controller's hardware GUID in each binding.

| Emulator | Config file | GUID required? | Format |
|----------|-------------|----------------|--------|
| Eden | `~/.config/yuzu/qt-config.ini` | Yes | `button:1,guid:...,engine:sdl` |
| Azahar | `~/.config/azahar-emu/qt-config.ini` | Yes | `button:1,guid:...,engine:sdl` |
| Ryujinx | `~/.config/Ryujinx/profiles/controller/*.json` | Yes | JSON with `id` field |
| Vita3K | `~/.config/Vita3K/config.yml` | No | Auto-detects |
| RPCS3 | `~/.config/rpcs3/config.yml` | No | Uses device name string |

#### Key finding: Steam Input virtualizes GUIDs

On Steam Deck in game mode, Steam Input intercepts all controllers and presents them with the same virtual GUID regardless of the physical controller:

```
Steam Deck GUID: 03000000de280000ff11000001000000
```

This means any controller connected to a Steam Deck (Xbox, PlayStation, 8BitDo, etc.) appears to emulators as the Steam Deck controller. EmuDeck exploits this: they ship configs with the Steam Deck GUID and it works for all controllers.

#### Solution: Ship configs for common GUIDs

For GUID-based emulators, Kyaraben ships configs with known GUIDs:

| Controller | GUID |
|------------|------|
| Steam Deck | `03000000de280000ff11000001000000` |
| Xbox 360 | `030000005e0400008e02000010010000` |
| Xbox One | `030000005e040000ea02000001030000` |
| DualShock 4 | `030000004c050000c405000011010000` |
| DualSense | `030000004c050000e60c000011010000` |
| Switch Pro | `030000007e0500000920000011010000` |

**Coverage**:
- Steam Deck users: 100% (Steam Input virtualizes all controllers)
- Desktop users with common controllers: covered by shipped GUIDs
- Desktop users with other controllers: configure through emulator UI (fallback)

#### Eden/Azahar format

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

#### Ryujinx format (JSON)

```json
{
  "left_joycon_stick": {
    "joystick": "Left",
    "invert_stick_x": false,
    "invert_stick_y": false,
    "rotate90_cw": false,
    "stick_button": "LeftStick"
  },
  "right_joycon_stick": {
    "joystick": "Right",
    "invert_stick_x": false,
    "invert_stick_y": false,
    "rotate90_cw": false,
    "stick_button": "RightStick"
  },
  "deadzone_left": 0.1,
  "deadzone_right": 0.1,
  "range_left": 1,
  "range_right": 1,
  "trigger_threshold": 0.5,
  "left_joycon": {
    "button_minus": "Minus",
    "button_l": "LeftShoulder",
    "button_zl": "LeftTrigger",
    "dpad_up": "DpadUp",
    "dpad_down": "DpadDown",
    "dpad_left": "DpadLeft",
    "dpad_right": "DpadRight"
  },
  "right_joycon": {
    "button_plus": "Plus",
    "button_r": "RightShoulder",
    "button_zr": "RightTrigger",
    "button_a": "B",
    "button_b": "A",
    "button_x": "Y",
    "button_y": "X"
  },
  "backend": "GamepadSDL2",
  "id": "0-00000003-28de-0000-ff11-000001000000",
  "controller_type": "ProController",
  "player_index": "Player1"
}
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
| Ryujinx | Switch | Ship GUIDs | Steam Deck 100%, desktop common controllers |
| Vita3K | PS Vita | Auto-detect | 100% |
| RPCS3 | PS3 | Device name | 100% |

## Implementation phases

### Phase 1: Group A + D (highest value, lowest effort)

| Emulator | Systems | Effort |
|----------|---------|--------|
| DuckStation | PSX | Mapping table |
| PCSX2 | PS2 | Mapping table |
| Dolphin | GameCube, Wii | Mapping table |
| RetroArch | Many | Verify autodetect enabled |
| Vita3K | PS Vita | Nothing needed |
| RPCS3 | PS3 | Verify config |

### Phase 2: Group B (medium effort)

| Emulator | Systems | Effort |
|----------|---------|--------|
| mGBA | GBA | Index mapping |
| melonDS | NDS | Index mapping |
| PPSSPP | PSP | Keycode mapping |
| Flycast | Dreamcast | Ship mapping files |

### Phase 3: Group C (GUID-based)

| Emulator | Systems | Effort |
|----------|---------|--------|
| Eden | Switch | Ship profiles for common GUIDs |
| Azahar | 3DS | Ship profiles for common GUIDs |
| Ryujinx | Switch | Ship JSON profiles for common GUIDs |

## Implementation guidelines

### Code-compact runtime generation

Do NOT ship static config files per controller/emulator combination. Instead:

1. Define profiles as data structures (button layout + GUID list)
2. Define per-emulator format templates
3. Generate configs at runtime during `apply`

```go
type ControllerProfile struct {
    Name    string
    GUIDs   []string
    Layout  ButtonLayout // Xbox or Nintendo
}

type ButtonLayout struct {
    FaceBottom string // A on Xbox, B on Nintendo
    FaceRight  string // B on Xbox, A on Nintendo
    FaceLeft   string // X on Xbox, Y on Nintendo
    FaceTop    string // Y on Xbox, X on Nintendo
    // ... shoulders, triggers, sticks, dpad
}

var XboxLayout = ButtonLayout{
    FaceBottom: "A", FaceRight: "B", FaceLeft: "X", FaceTop: "Y",
}

var NintendoLayout = ButtonLayout{
    FaceBottom: "B", FaceRight: "A", FaceLeft: "Y", FaceTop: "X",
}
```

Each emulator implements a generator that takes a profile and outputs `[]ConfigEntry`:

```go
type ControllerConfigGenerator interface {
    GenerateController(profile ControllerProfile, hotkeys HotkeyConfig) []model.ConfigEntry
}
```

### Profile to GUID mapping

One profile covers multiple hardware GUIDs:

```go
var Profiles = []ControllerProfile{
    {
        Name: "Xbox",
        GUIDs: []string{
            "030000005e0400008e02000010010000", // 360 wired
            "030000005e0400009102000007010000", // 360 wireless
            "030000005e040000ea02000001030000", // One
            "030000005e040000d102000001010000", // One S
            "030000005e040000130b000011050000", // Series X/S
        },
    },
    {
        Name: "PlayStation",
        GUIDs: []string{
            "030000004c0500006802000011010000", // DS3
            "030000004c050000c405000011010000", // DS4 v1
            "030000004c050000cc09000011010000", // DS4 v2
            "030000004c050000e60c000011010000", // DualSense
        },
    },
    {
        Name: "Nintendo",
        GUIDs: []string{
            "030000007e0500000920000011010000", // Switch Pro
            "03000000c82d00000161000000010000", // 8BitDo SN30 Pro
            // ... other 8BitDo variants
        },
    },
    {
        Name: "Steam Deck",
        GUIDs: []string{
            "03000000de280000ff11000001000000", // Neptune
        },
    },
}
```

## User configuration

### Source of truth: config.toml

Controller settings live in Kyaraben's `config.toml`. Uses SDL button names for consistency.

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

### Layout naming

TBD: Investigate what EmuDeck uses. Likely:
- `standard`: A=bottom (Xbox, PlayStation, most controllers)
- `nintendo`: A=right, B=bottom (Nintendo controllers, some 8BitDo)

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

### Emulator-specific hotkeys (not in Kyaraben scope)

These exist but are left for users to configure manually:
- **Dolphin**: Wii Remote connect, controller profile switch, aspect ratio toggle
- **melonDS**: Swap screens, lid close/open, microphone
- **PPSSPP**: Analog limiter, texture dump
- **DuckStation**: Change disc, toggle PGXP

### UI integration

Kyaraben UI exposes:
- Layout toggle: Xbox / Nintendo
- Hotkey configuration with button picker
- Changes write to `config.toml`
- Applied during normal `apply` flow

### Apply flow

During `apply`:
1. Read controller config from `config.toml`
2. Determine target profile (Steam Deck auto-detected, or user-selected, or all common)
3. For each enabled emulator:
   - Call `GenerateController(profile, hotkeys)`
   - Merge resulting entries into emulator config patch
4. Write configs alongside path settings

## Steam Input enhancement (optional)

Steam Input operates above SDL:

```
Physical controller → Steam Input → Virtual Xbox controller → SDL → Emulator
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
- Controllers not in common GUID list (configure through emulator UI)

## Decisions

### Profile selection: Write all common GUIDs

Always write configs for all common controller GUIDs. This results in larger config files but ensures any common controller works with zero friction.

**Config file size**: For GUID-based emulators (Eden, Azahar, Ryujinx), we generate bindings for each GUID. With ~10 common GUIDs × 4 players × ~20 bindings = ~800 config lines per emulator. Acceptable tradeoff for universal compatibility.

For users with unsupported controllers: separate "check controller compatibility" tool (future feature) that detects connected controller and suggests selecting the most similar profile if GUID not found.

### Multi-player: Always write 4 player slots

Always generate configs for players 1-4. No configuration needed.

```go
func (g *Generator) GenerateController(profile Profile, hotkeys HotkeyConfig) []ConfigEntry {
    var entries []ConfigEntry
    for playerIndex := 0; playerIndex < 4; playerIndex++ {
        entries = append(entries, g.generatePad(playerIndex, profile)...)
    }
    entries = append(entries, g.generateHotkeys(0, hotkeys)...) // Hotkeys on P1 only
    return entries
}
```

- Same layout for all players (global preference)
- Hotkeys only on Player 1
- Players 2-4 use SDL indices 1, 2, 3

### Emulator-specific hotkeys

Document what exists, leave configuration to users. Kyaraben manages the universal set only.

### Unsupported hotkeys

If an emulator doesn't support a configured hotkey (e.g., rewind on Dolphin), skip silently. Coverage documented in this file.

## Testing strategy

### Unit tests

Test generation logic produces correct config entries:

```go
func TestDuckStationGenerateController(t *testing.T) {
    gen := duckstation.NewControllerGenerator()
    profile := profiles.Xbox
    hotkeys := config.DefaultHotkeys()

    entries := gen.Generate(profile, hotkeys)

    // Verify button mappings
    assertEntry(t, entries, "Pad1", "Cross", "SDL-0/A")
    assertEntry(t, entries, "Pad1", "Circle", "SDL-0/B")
    assertEntry(t, entries, "Pad2", "Cross", "SDL-1/A")

    // Verify hotkeys
    assertEntry(t, entries, "Hotkeys", "SaveSelectedSaveState", "SDL-0/Back & SDL-0/RightShoulder")
}

func TestNintendoLayoutSwapsButtons(t *testing.T) {
    gen := duckstation.NewControllerGenerator()
    profile := profiles.Nintendo

    entries := gen.Generate(profile, config.DefaultHotkeys())

    // A and B swapped
    assertEntry(t, entries, "Pad1", "Cross", "SDL-0/B")  // Cross = B on Nintendo
    assertEntry(t, entries, "Pad1", "Circle", "SDL-0/A") // Circle = A on Nintendo
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

    // Write to temp file
    path := writeConfig(t, entries, model.ConfigFormatINI)

    // Parse back
    handler := configformat.NewHandler(fs, model.ConfigFormatINI)
    parsed, err := handler.Read(path)
    require.NoError(t, err)

    // Verify key sections exist
    assert.Contains(t, parsed, "Pad1")
    assert.Contains(t, parsed, "Pad2")
    assert.Contains(t, parsed, "Hotkeys")
}
```

### Virtual controller E2E

For CI, test actual input path using Linux uinput:

```go
func TestControllerInputE2E(t *testing.T) {
    if os.Getenv("CI_E2E_CONTROLLER") == "" {
        t.Skip("skipping E2E controller test")
    }

    // Create virtual Xbox controller via uinput
    vctl, err := uinput.CreateGamepad("/dev/uinput", "Test Xbox Controller", 0x045e, 0x028e)
    require.NoError(t, err)
    defer vctl.Close()

    // Apply generated config
    applyControllerConfig(t, profiles.Xbox)

    // Start emulator with test ROM (headless)
    emu := startEmulator(t, "duckstation", "--headless", testROM)
    defer emu.Stop()

    // Send button press
    vctl.ButtonDown(uinput.ButtonSouth)
    time.Sleep(100 * time.Millisecond)
    vctl.ButtonUp(uinput.ButtonSouth)

    // Verify input received (check emulator log or internal state)
    assert.Eventually(t, func() bool {
        return emu.LastInput() == "Cross"
    }, time.Second, 100*time.Millisecond)
}
```

This tests:
- Config generation
- Config format correctness
- Emulator actually receives input via SDL
- GUID matching works

Run in CI on Linux runners with uinput access.

## Open decisions

### Hotkey feature set

Is the proposed set complete?
- save_state, load_state, next_slot, prev_slot
- fast_forward, rewind
- pause, screenshot, quit
- toggle_fullscreen, open_menu

### Layout naming

TBD: Investigate what EmuDeck calls it. Likely:
- `standard`: A=bottom (Xbox, PlayStation, most controllers)
- `nintendo`: A=right, B=bottom (Nintendo controllers)

## Sources

- [SDL GameControllerDB](https://github.com/mdqinc/SDL_GameControllerDB)
- [RetroArch joypad-autoconfig](https://github.com/libretro/retroarch-joypad-autoconfig)
- [Yuzu controller configuration](https://yuzu-emulator.com/wiki/controller-configuration/)
- [RPCS3 controller configuration](https://wiki.rpcs3.net/index.php?title=Help:Controller_Configuration)
- [Vita3K quickstart](https://vita3k.org/quickstart.html)
- [Steam Input documentation](https://partner.steamgames.com/doc/features/steam_controller)
- EmuDeck configs (GPL-3) for format reference
- RetroDECK configs (GPL-3) for format reference
