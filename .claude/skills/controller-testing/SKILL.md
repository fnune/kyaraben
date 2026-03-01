---
name: controller-testing
description: Companion for manually testing controller configuration on new or existing emulators.
---

# Controller testing companion

This skill helps you manually test controller configuration for kyaraben emulators. Use it when adding a new emulator, debugging controller issues, or verifying existing setups.

## Quick reference

See `site/src/content/docs/using-the-app.mdx` under "Emulator support" for the authoritative table of what kyaraben configures per emulator.

### Default hotkeys (Xbox layout)

| Action      | Chord        |
| ----------- | ------------ |
| Save state  | Back + RB    |
| Load state  | Back + LB    |
| Next slot   | Start + RB   |
| Prev slot   | Start + LB   |
| Fast fwd    | Back + RT    |
| Rewind      | Back + LT    |
| Pause       | Back + A     |
| Screenshot  | Back + B     |
| Quit        | Back + Start |
| Fullscreen  | Start + L3   |
| Menu        | Back + R3    |

SNES USB controllers lack triggers, sticks, and extra buttons. Only hotkeys using Select, Start, L, R, and face buttons work.

## Native button layouts by system

This section documents each console's native face button layout and how kyaraben maps it to your controller. Kyaraben assumes an Xbox-layout controller (Steam Deck, Xbox, most third-party pads).

Face button positions:
```
      North
West         East
      South
```

On Xbox-layout controllers: A=south, B=east, X=west, Y=north.

### Confirm button position setting

Kyaraben asks: "Where is your controller's confirm button?" This determines how face buttons map to Nintendo games.

**South (default)**: Your confirm button is at the south position (Xbox A, PlayStation Cross). This is the standard for Xbox and PlayStation controllers.

**East**: Your confirm button is at the east position (Nintendo A). Use this if you have a Nintendo Pro Controller or Switch controller and want button labels to match Nintendo games.

| Setting | Your A button triggers | Your B button triggers |
| ------- | ---------------------- | ---------------------- |
| South | Nintendo B (south action) | Nintendo A (east action) |
| East | Nintendo A (east action) | Nintendo B (south action) |

**Which systems are affected:**

Only Nintendo systems with diamond-layout face buttons: SNES, GameCube, Wii, Wii U, Switch, DS, 3DS, GBA.

**Not affected:**
- N64: Unique controller layout, no clean A/B swap applies
- PlayStation, Sega, Xbox: No ambiguity, positions already match
- Systems using auto-detection (Vita3K, RPCS3, xemu, Xenia)

### Nintendo systems

All Nintendo systems use A=east, B=south (opposite of Xbox).

| System | Native layout | Kyaraben mapping (Xbox controller) |
| ------ | ------------- | ---------------------------------- |
| NES | A (east), B (west) | A→east (B), B→west (X) |
| SNES | A (east), B (south), X (north), Y (west) | A→east (B), B→south (A), X→north (Y), Y→west (X) |
| N64 | A (east), B (south), C-buttons (right side) | A→east (B), B→south (A), C→right stick |
| GameCube | A (east, large), B (south), X (north), Y (west) | A→east (B), B→south (A), X→north (Y), Y→west (X) |
| Wii | Same as GameCube for classic controller | Same as GameCube |
| Wii U | A (east), B (south), X (north), Y (west) | A→east (B), B→south (A), X→north (Y), Y→west (X) |
| Switch | A (east), B (south), X (north), Y (west) | A→east (B), B→south (A), X→north (Y), Y→west (X) |
| GB/GBA | A (east), B (west) | A→east (B), B→west (X) |
| DS/3DS | A (east), B (south), X (north), Y (west) | A→east (B), B→south (A), X→north (Y), Y→west (X) |

Testing tip: On Nintendo systems, pressing your controller's B button (east) should trigger the game's "A" action (usually confirm/jump). Pressing A (south) triggers "B" (usually cancel/run).

### PlayStation systems

PlayStation uses shapes. Position matters, not the shape name.

| System | Native layout | Kyaraben mapping (Xbox controller) |
| ------ | ------------- | ---------------------------------- |
| PSX/PS2/PS3 | Cross (south), Circle (east), Square (west), Triangle (north) | A→Cross, B→Circle, X→Square, Y→Triangle |
| PSP | Same as PlayStation | Same as PlayStation |
| PS Vita | Same as PlayStation | Same as PlayStation |

Testing tip: PlayStation layouts match Xbox positions (south=confirm in Western games). Your A button triggers Cross, B triggers Circle, etc.

### Sega systems

| System | Native layout | Kyaraben mapping (Xbox controller) |
| ------ | ------------- | ---------------------------------- |
| Master System | 1 (south), 2 (east) | A→1, B→2 |
| Genesis 3-btn | A (west), B (south), C (east) | X→A, A→B, B→C |
| Genesis 6-btn | A (west), B (south), C (east), X (LB), Y (north), Z (RB) | X→A, A→B, B→C, LB→X, Y→Y, RB→Z |
| Saturn | A (west), B (south), C (east), X (LB), Y (north), Z (RB) | Same as Genesis 6-button |
| Dreamcast | A (south), B (east), X (west), Y (north) | Matches Xbox layout exactly |

Testing tip: Dreamcast matches Xbox layout. Genesis/Saturn have a unique 6-button layout where the bottom row (A, B, C) maps to X, A, B and top row (X, Y, Z) maps to LB, Y, RB.

### Xbox systems

| System | Native layout | Kyaraben mapping |
| ------ | ------------- | ---------------- |
| Xbox | A (south), B (east), X (west), Y (north) | Direct 1:1 mapping |
| Xbox 360 | Same as Xbox | Direct 1:1 mapping |

Testing tip: No remapping needed. Your controller is the native layout.

### Other systems

| System | Native layout | Kyaraben mapping (Xbox controller) |
| ------ | ------------- | ---------------------------------- |
| Atari 2600 | 1 button | A |
| PC Engine | I (east), II (west) | B→I, X→II |
| Neo Geo | A (south), B (east), C (north), D (west) | A→A, B→B, Y→C, X→D |
| Arcade | Varies by game | Typically A, B, X, Y, LB, RB for 6-button |

### Quick reference: "confirm" button by system family

When testing menu navigation:

| System family | Confirm button on original | Press this on Xbox controller |
| ------------- | -------------------------- | ----------------------------- |
| Nintendo | A (east) | B |
| PlayStation (Western) | Cross (south) | A |
| PlayStation (Japanese) | Circle (east) | B |
| Sega (Dreamcast) | A (south) | A |
| Sega (Genesis/Saturn) | A or C | X or B |
| Xbox | A (south) | A |

## Emulator tiers

### Tier 1: auto-detection (easiest to test)

These emulators use SDL auto-detection. Controllers work without GUID configuration.

**RetroArch cores** (bsnes, Mesen, Genesis Plus GX, mGBA, melonDS, Citra, Mupen64Plus, Beetle Saturn, FBNeo, Stella, VICE)
- SDL2 joypad driver with `input_autodetect_enable = true`
- Hotkeys use enable button (Back) + action button pattern
- All cores share the same hotkey configuration
- Test any core to verify RetroArch controller setup

**DuckStation / PCSX2**
- SDL button names: `SDL-0/A`, `SDL-0/Back`, `SDL-0/LeftTrigger+`
- Hotkeys use `&` separator: `SDL-0/Back & SDL-0/RightShoulder`
- 4-player support with separate Pad1-Pad4 sections
- Profile files allow user customization after initial setup

**Vita3K / RPCS3**
- Controller handled entirely by emulator auto-detection
- Kyaraben generates path config only, no button bindings
- If controller works in other SDL apps, it works here

**Testing approach:** Launch a game, verify d-pad, sticks, face buttons, shoulders, triggers. Test hotkey chords. No special setup needed.

### Tier 2: standard config generation

These emulators need kyaraben-generated config but use predictable formats.

**Dolphin**
- Custom button names in backticks: `` `Button S` `` (south), `` `Button E` `` (east)
- Hotkeys use toggle wrapper: `` toggle(@(`Back`+`DPad Up`)) ``
- Device line specifies controller: `SDL/0/Steam Deck Controller`

Quirks:
- Stick axes need separate +/- entries
- Profile files in `Dolphin/Config/Profiles/GCPad/`

**PPSSPP**
- Device-keycode format: `10-189` where 10=gamepad, 189=A button
- Hotkeys use colon separator: `10-196:10-193` (Select + L)
- No trigger button support (PSP has no triggers)
- Analog stick mapped as 4 discrete keys (not axis values)

Quirks:
- Keycodes are PSP-specific, not SDL indices
- Missing L2/R2 support is intentional (matches PSP hardware)

**Flycast**
- Separate `[digital]` and `[analog]` sections
- Raw joystick indices: A=0, B=1, X=2, Y=3
- D-pad uses HAT indices: 256=up, 257=down, 258=left, 259=right
- Axis notation with sign: `2+:btn_trigger_left`
- Limited hotkeys (screenshot, fast forward, save/load, quit only)

Quirks:
- Uses raw SDL joystick API, not GameController
- Triggers are axes 2 and 5
- Hotkey format: `6,7:action:0` (Back=6, Start=7, 0=simultaneous)

**Cemu**
- XML-based controller profile (not INI)
- VPAD button IDs map to Wii U GamePad
- GUID embedded in profile
- No hotkey support (upstream limitation)

Quirks:
- Single-player only in kyaraben config
- Trigger axes use special SDL indices (42, 43)
- Touch screen emulated via mouse

**Testing approach:** Launch a game, verify all inputs. Check button positions against "Native button layouts by system" above. Test available hotkeys.

### Tier 3: GUID-based and quirky

These emulators require special handling or have significant limitations.

**Eden (Switch)**
- GUID-based bindings: `engine:sdl,port:0,guid:03000000de280000ff11000001000000,button:0`
- Raw joystick indices (not GameController): A=0, B=1, X=2, Y=3
- 2-player support only
- Limited hotkeys (Plus+Minus combinations)

Quirks:
- Requires Steam Deck GUID exactly
- Key ordering in binding strings is non-deterministic (kyaraben uses semantic equality)
- Profile file `Kyaraben.ini` is fully managed (user changes overwritten)
- Axis bindings need threshold: `threshold:0.500000`
- Stick bindings need deadzone: `deadzone:0.100000`

Testing gotchas:
- Must test on Steam Deck or with matching GUID
- SNES USB controller won't work without adding its GUID
- Compare bindings semantically, not as strings

**xemu (Xbox)**
- GUID-only binding: `input.bindings.port1 = "03000000de280000ff11000001000000"`
- No button remapping (controller must match Xbox layout)
- No hotkey support
- 4-player support

Quirks:
- Minimal controller config (just assigns GUID to ports)
- Requires manual Xbox HDD image setup
- If buttons feel wrong, it's the controller, not the config

**Xenia Edge (Xbox 360)**
- Similar to xemu: minimal config, auto-detection
- No hotkey support
- Requires X11 backend on Steam Deck (`GDK_BACKEND=x11`)

## Common test checklist

For any emulator, verify:

1. **D-pad**: All 4 directions work correctly
2. **Sticks**: Left stick moves, right stick moves (if applicable)
3. **Face buttons**: Perform expected in-game actions
4. **Shoulders**: L/R or LB/RB register
5. **Triggers**: LT/RT register (check analog vs digital)
6. **Start/Back**: Menu functions work
7. **Hotkeys**: Save state, load state, fast forward work (if supported)

For multi-player emulators, also verify:
- Player 2 controller detected
- Player 2 inputs map to correct in-game slot

## Debugging tips

**Controller not detected:**
- Check Steam Input settings (pass-through vs remapping)
- Verify controller works in other SDL apps
- For GUID emulators, check if controller GUID is configured

**Wrong button positions:**
- See "Native button layouts by system" above for expected mappings
- Nintendo A=east (your B button), PlayStation Cross=south (your A button)
- Verify kyaraben layout setting matches your controller

**Hotkeys not working:**
- Confirm emulator supports hotkeys (see tier table)
- For RetroArch, verify `input_enable_hotkey_btn` is set to Back
- Hold enable button before pressing action button

**Config not applied:**
- Run kyaraben apply after changes
- Check for preflight conflicts
- Verify config file path matches emulator expectations

## Adding controller support to a new emulator

See `/.claude/skills/adding-emulator-support/SKILL.md` for the full process. Controller-specific steps:

1. Determine if emulator uses SDL GameController or raw joystick API
2. Find button/axis naming convention in emulator docs or existing configs
3. Implement `ConfigGenerator.Generate()` with controller patches
4. Use `ctx.ControllerConfig.FaceButtons()` for layout-aware button mapping
5. Test with Xbox-layout controller and Nintendo-layout controller
6. Document any quirks in this skill under the appropriate tier
