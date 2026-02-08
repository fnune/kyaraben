# Controller and hotkey support plan

## The domain

### Two layers of controller mapping

Controller support has two distinct layers:

1. **Game controls**: mapping physical buttons to in-game actions (A = jump, B = attack)
2. **Emulator hotkeys**: mapping button combinations to emulator functions (Select+R1 = save state)

Game controls are largely handled by emulators automatically - they detect controllers via SDL/XInput and let users remap if needed. The real value-add for kyaraben is hotkeys.

### The hotkey problem

Every emulator has its own hotkey system:
- Different config formats (INI, JSON, XML, CFG)
- Different key naming (`F1`, `KEY_F1`, `0x3B`, `button:1`)
- Different modifier handling (`Shift+F1`, `@(SELECT+START)`, chord syntax)
- Different features supported (some have rewind, some don't)

Without coordination, users must learn different hotkey schemes for each emulator. Save state might be F5 in one, Ctrl+S in another, Select+R1 in a third.

### Controller identification

Controllers are identified by:
- **GUID**: hardware fingerprint like `03000000de280000ff11000001000000` (Steam Deck)
- **SDL device name**: "Steam Deck Controller", "Xbox 360 Controller"
- **Device path**: `/dev/input/js0`

Steam Deck's GUID is consistent, making it a good target for pre-configured profiles.

## How EmuDeck and RetroDECK solve this

### EmuDeck: template configs per emulator

EmuDeck ships complete controller configs for each emulator:
- `configs/org.DolphinEmu.dolphin-emu/config/dolphin-emu/GCPadNew.ini`
- `configs/duckstation/settings.ini` (includes controller section)
- `configs/Ryujinx/profiles/controller/Deck.json`

These are pre-made for Steam Deck (GUID hardcoded). On setup, they rsync the whole config directory.

**Hotkey conventions (EmuDeck Steam Deck):**
| Action | Combo |
|--------|-------|
| Save state | Select + R1 |
| Load state | Select + L1 |
| Fast forward | Select + R2 |
| Rewind | Select + L2 |
| Screenshot | Select + Y |
| Pause | Select + A |
| Exit | Select + Start |

These are set per-emulator in each config file. No abstraction layer - just consistent manual configuration across all template files.

**Pros:**
- Works immediately, zero runtime complexity
- Full control over each emulator's quirks

**Cons:**
- Massive maintenance burden (updating 30+ config files when conventions change)
- Hardcoded to Steam Deck GUID
- No user customization without manual editing

### RetroDECK: Steam Input abstraction

RetroDECK uses Valve's Steam Input system as an abstraction layer:
1. Define actions in VDF files ("SAVE STATE", "LOAD STATE", etc.)
2. Map physical buttons to actions in controller profiles
3. Steam Input sends synthetic keyboard/mouse events to emulators
4. Emulators are configured to respond to those keyboard shortcuts

**VDF profile structure:**
```
"controller_mappings"
{
  "title": "RetroDECK: Steam Deck v.1.2"
  "Preset_1000003": "State Menu"  // Save/Load States
  "Preset_1000004": "Speed Menu"  // Fast Forward, Rewind
  ...
}
```

They have 10+ controller profiles (Steam Deck, PS4, PS5, Xbox, Switch Pro, etc.) with consistent action mappings.

**Pros:**
- One abstraction for all controllers
- Steam handles hardware detection
- Radial menus and complex UI possible
- Works for any controller Steam supports

**Cons:**
- Requires Steam (won't work on non-Steam Linux desktops)
- Complex VDF format
- Emulators still need keyboard shortcuts configured to match

## The core insight

Both approaches ultimately configure emulators to respond to keyboard shortcuts, then map controller buttons to those shortcuts.

RetroDECK's abstraction is at the controller level (Steam Input).
EmuDeck's abstraction is at the config template level (rsync pre-made files).

Neither has a semantic abstraction for hotkeys - neither can say "set save-state to Select+R1" and have it apply to all emulators.

## Kyaraben's opportunity

Kyaraben already has an abstraction for emulator configuration (`ConfigGenerator` interface, `ConfigPatch` model). We could extend this to hotkeys:

```
User preference: "save_state = Select+R1"
                        ↓
            Kyaraben translates to:
                        ↓
┌─────────────────────────────────────────────────────────┐
│ DuckStation: Hotkey/SaveState = SDL-0/+Button 10        │
│ Dolphin:     General/SaveStateSlot = @(SELECT+SOUTH)    │
│ RetroArch:   input_save_state = "button:4"              │
│ PCSX2:       Hotkeys/SaveStateToSlot = SDL-0/Select...  │
└─────────────────────────────────────────────────────────┘
```

The key is a translation layer that understands:
1. Semantic actions (save_state, load_state, fast_forward, screenshot, exit)
2. Controller inputs (Select, Start, A, B, X, Y, L1, R1, L2, R2, etc.)
3. Emulator-specific syntax for expressing button combinations

## Proposed approach for kyaraben

### Phase 1: define the semantic model

Define a set of common hotkey actions:
- `save_state` / `load_state`
- `next_state_slot` / `prev_state_slot`
- `fast_forward` (toggle or hold)
- `rewind` (if supported)
- `pause`
- `screenshot`
- `exit_emulator`
- `toggle_fullscreen`

And a set of controller inputs:
- Face buttons: A, B, X, Y (or South, East, West, North for layout-agnostic)
- Shoulders: L1, R1, L2, R2
- Sticks: L3, R3 (stick click)
- Meta: Select, Start, Guide
- D-pad: Up, Down, Left, Right

Combinations expressed as: `Select+R1`, `L3+R3`, `Start+Select`

### Phase 2: implement emulator translators

Each emulator needs a translator that converts:
```
(action: save_state, input: Select+R1) → config entries
```

This is similar to how `ConfigGenerator` works today, but for hotkeys.

**Challenges per emulator:**

| Emulator | Format | Modifier syntax | Notes |
|----------|--------|-----------------|-------|
| DuckStation | INI | `SDL-0/+Button 10` | Uses SDL device prefix |
| Dolphin | INI | `@(SELECT+SOUTH)` | Custom chord syntax |
| RetroArch | CFG | `"button:4"` or `"nul"` | Button numbers vary by controller |
| PCSX2 | INI | `SDL-0/Select & SDL-0/R1` | Ampersand for combos |
| PPSSPP | INI | `10-45` | Numeric codes |
| mGBA | INI | Keyboard only for hotkeys | No controller hotkey support? |
| Cemu | XML | `<button>40</button>` | Numeric in XML |

Some emulators may not support controller hotkeys at all (keyboard only). Kyaraben would skip those or document the limitation.

### Phase 3: controller profiles

Instead of hardcoding Steam Deck GUID, define profiles:
- `steam_deck`: GUID `03000000de280000ff11000001000000`
- `xbox_360`: GUID `...`
- `ps4`: GUID `...`
- `generic`: fallback SDL mappings

Profiles specify button numbering since that varies by controller (Steam Deck's "A" is button 1, Xbox's "A" might be button 0).

### Phase 4: user configuration

Add to kyaraben's store:
```yaml
controller:
  profile: steam_deck  # or auto-detect
  hotkeys:
    save_state: Select+R1
    load_state: Select+L1
    fast_forward: Select+R2
    screenshot: Select+Y
    exit: Select+Start
```

Users configure once, kyaraben applies everywhere.

## What about non-Steam usage?

RetroDECK's Steam Input approach requires Steam. For desktop Linux without Steam:
- Emulators use their native controller support (SDL)
- Kyaraben writes hotkey configs directly to each emulator
- No intermediate abstraction needed

This is actually simpler than RetroDECK's approach - direct config writing vs VDF indirection.

For Steam Deck in Gaming Mode, Steam Input is always active. Kyaraben's direct config approach still works because Steam Input just translates controller → keyboard, and the emulator responds to keyboard.

## Scope considerations

### What to include

1. Hotkeys for common actions (save/load state, fast forward, exit)
2. Steam Deck as primary target (consistent hardware)
3. Generic controller fallback
4. User-configurable hotkey bindings

### What to defer

1. Per-game controller remapping (let emulators handle this)
2. Radial menus / complex UI (RetroDECK territory)
3. Gyro/motion controls (very emulator-specific)
4. Multi-player controller assignment (emulator-specific)

### What to explicitly not do

1. Replace emulator controller UIs (too complex, low value)
2. Steam Input VDF generation (ties us to Steam)
3. Force layouts on users (ABXY vs BAYX is preference)

## Implementation complexity

**Low complexity:**
- Define semantic model (actions + inputs)
- Add hotkey fields to existing emulator `ConfigGenerator`s
- Store user preferences in kyaraben config

**Medium complexity:**
- Translate button combos to emulator-specific syntax
- Handle emulators that only support keyboard hotkeys
- Controller profile detection/selection

**High complexity (avoid):**
- Runtime controller detection
- Dynamic GUID resolution
- Steam Input integration

## Emulator hotkey support investigation

Detailed findings on which emulators support controller hotkeys via text config.

### Fully supported (controller hotkeys via config)

#### DuckStation

- Format: INI (`settings.ini`)
- Syntax: `SDL-0/+Button 10` or `SDL-0/Back & SDL-0/A` for combos
- Supported actions: save state, load state, fast forward, rewind, pause, screenshot, exit, fullscreen
- Example:
  ```ini
  [Hotkeys]
  SaveSelectedSaveState = SDL-0/Back & SDL-0/RightShoulder
  LoadSelectedSaveState = SDL-0/Back & SDL-0/LeftShoulder
  ToggleFastForward = SDL-0/Back & SDL-0/+RightTrigger
  ```

#### Dolphin

- Format: INI (`Hotkeys.ini`)
- Syntax: `@(Back+Button S)` custom chord notation
- Supported actions: save state, load state, pause, screenshot, exit, fullscreen, fast forward
- Example:
  ```ini
  [Hotkeys]
  Device = SDL/0/Steam Deck Controller
  Save State/Save to Selected Slot = @(Back+`Shoulder R`)
  Load State/Load from Selected Slot = @(Back+`Shoulder L`)
  General/Exit = @(Back+Start)
  ```

#### PPSSPP

- Format: INI (`controls.ini`)
- Syntax: numeric codes like `10-196:10-192`
- Supported actions: save state, load state, fast forward, rewind, pause, screenshot, exit
- Example:
  ```ini
  [ControlMapping]
  Save State = 10-196:10-192,1-132
  Load State = 10-196:10-193,1-133
  Fast-forward = 1-193:10-4010,1-135
  ```

#### PCSX2

- Format: INI (`PCSX2.ini`)
- Syntax: `SDL-0/Select & SDL-0/R1` ampersand for combos
- Supported actions: save state, load state, pause, screenshot
- Example:
  ```ini
  [Hotkeys]
  SaveStateToSlot = SDL-0/Back & SDL-0/RightShoulder
  LoadStateFromSlot = SDL-0/Back & SDL-0/LeftShoulder
  ```

#### RetroArch

- Format: CFG (`retroarch.cfg`)
- Syntax: button numbers with enable hotkey modifier
- Supported actions: extensive (save, load, fast forward, rewind, pause, screenshot, exit, menu, cheats, disk swap)
- Example:
  ```cfg
  input_enable_hotkey_btn = "4"
  input_save_state_btn = "8"
  input_load_state_btn = "9"
  input_exit_emulator_btn = "6"
  ```
- Most comprehensive hotkey support of all emulators

### Partially supported

#### melonDS

- Format: INI (`melonDS.ini`)
- Controller hotkeys exist (`HKJoy_*`) but save/load state is hardcoded
- Configurable via config: pause, reset, fast forward, fullscreen, swap screens
- Hardcoded (keyboard only): save state (Shift+F1-F9), load state (F1-F9)
- Example:
  ```ini
  HKJoy_Pause=0
  HKJoy_FastForward=-1
  HKJoy_FullscreenToggle=-1
  ```
- Sources: [GitHub issue #1183](https://github.com/Arisotura/melonDS/issues/1183)

#### Eden (Switch)

- Format: INI (`qt-config.ini`)
- Controller button mapping works
- No save states exist for Switch emulation (N/A)

### Keyboard only (no controller hotkeys in config)

#### mGBA

- Format: INI (`config.ini`)
- Hotkeys are keyboard shortcuts only (F1-F10, etc.)
- No `HKJoy_*` or equivalent for controller hotkeys
- Would need Steam Input for controller hotkeys

#### Azahar (3DS)

- Format: INI (Qt config)
- Keyboard shortcuts via Qt framework (`KeySeq=Ctrl+P`)
- No controller hotkey support in config
- Would need Steam Input for controller hotkeys

#### Cemu

- Format: XML (`settings.xml`)
- Hotkeys are hardcoded (Ctrl+F1-F9 for save states)
- No controller hotkey configuration in any config file
- Controller profiles only map game controls, not emulator hotkeys
- Sources: [GitHub issue #721](https://github.com/cemu-project/Cemu/issues/721)

#### Flycast

- Format: INI (`emu.cfg`)
- Config only handles paths
- Known issue: save state hotkeys don't work when launching via CLI
- Would need Steam Input for controller hotkeys

### UI configuration only (no text config)

#### RPCS3

- Format: YAML (`config.yml`)
- Controller and hotkeys configured via emulator UI only
- Text config primarily handles VFS paths
- Would need Steam Input for controller hotkeys

#### Vita3K

- Format: YAML (`config.yml`)
- Minimal config, primarily paths
- Controller/hotkey configuration via UI only
- Would need Steam Input for controller hotkeys

### Summary table

| Emulator | Controller hotkeys (no Steam) | With Steam Input | Notes |
|----------|------------------------------|------------------|-------|
| DuckStation | Yes | Yes | Full support |
| Dolphin | Yes | Yes | Full support |
| PPSSPP | Yes | Yes | Full support |
| PCSX2 | Yes | Yes | Full support |
| RetroArch | Yes | Yes | Most comprehensive |
| melonDS | Partial | Yes | Save/load hardcoded to keyboard |
| mGBA | No | Yes | Keyboard only |
| Azahar | No | Yes | Keyboard only |
| Cemu | No | Yes | Hardcoded keyboard shortcuts |
| Flycast | No | Yes | CLI hotkey issues |
| RPCS3 | No | Probably | UI config only |
| Vita3K | No | Probably | UI config only |
| Eden | N/A | N/A | No save states for Switch |

### Recommended approach

1. Direct controller hotkey config for: DuckStation, Dolphin, PPSSPP, PCSX2, RetroArch
2. Consistent keyboard shortcuts for all emulators (enables Steam Input)
3. Optional Steam Input VDF generation for full coverage
4. Clear documentation of what works where

Desktop Linux users without Steam get 5 emulators with full controller hotkeys.
Steam users get nearly everything working.

## Summary

The path forward:
1. Semantic hotkey model (actions + inputs)
2. Per-emulator translators (extend ConfigGenerator)
3. Controller profiles (Steam Deck first, then generic)
4. User configuration in store

This gives users a "configure once, apply everywhere" experience for hotkeys while staying within kyaraben's existing architecture.
