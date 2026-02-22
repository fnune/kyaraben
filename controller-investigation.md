# Controller configuration investigation

## Problem statement

Controllers don't work in Eden, Dolphin, and Cemu when running via ES-DE on Steam Deck.

## Root cause (Eden)

Eden checks `key\default` suffix to decide whether to use config values or keyboard defaults:
- If `player_0_button_a\default=true` (or missing) → Eden ignores config, uses keyboard
- If `player_0_button_a\default=false` → Eden reads the actual value

**Fix implemented in commit `cfab7bb`**: Write `\default=false` for each binding.

## GUID issue

Two different GUIDs exist for Steam Deck controller:

| Context | GUID suffix | Description |
|---------|-------------|-------------|
| Desktop mode | `0512` | Raw hardware (product 0x1205) |
| Game Mode + Steam Input | `ff11` | Steam Virtual Gamepad (product 0x11FF) |

Kyaraben currently uses `ff11` (Steam Virtual Gamepad). EmuDeck uses the same.

## Test results

### Test 1: Desktop mode (fresh config)

1. Wiped `~/.config/eden/`
2. Ran `kyaraben apply` - wrote SDL bindings with `ff11` GUID and `\default=false`
3. Launched Eden from desktop
4. **Result**: Config preserved (not overwritten with keyboard) but controller didn't work
5. Had to manually: (a) select Kyaraben profile, (b) select "Steam Deck 0" as input device
6. After manual fix, GUID changed from `ff11` to `0512`

**Conclusion**: `\default=false` fix works to prevent keyboard overwrite, but `ff11` GUID doesn't match hardware in desktop mode.

### Test 2: Game Mode with Steam Input

1. Wiped `~/.config/eden/`
2. Ran `kyaraben apply`
3. Launched ES-DE via Game Mode (Steam shortcut)
4. Launched a Switch game in Eden
5. **Result**: Controller works without any manual configuration

**Conclusion**: `ff11` GUID works correctly when Steam Input is active (Game Mode via ES-DE).

## Summary

Eden controller support works correctly for the primary use case (Game Mode via ES-DE):
- `\default=false` fix prevents Eden from overwriting bindings with keyboard defaults
- `ff11` GUID matches Steam Virtual Gamepad when Steam Input is active

Desktop mode requires manual configuration due to GUID mismatch (`ff11` vs `0512`). This is acceptable since kyaraben's target use case is Game Mode via ES-DE.

## Eden source code analysis

### The `\default` suffix mechanism

Location: `src/frontend_common/config.cpp` (lines 715-821)

The `\default` suffix is a config state tracking mechanism, not related to device matching:
- When reading a setting, Eden checks `key\default` flag (lines 811-812)
- If true (or missing), Eden ignores the stored value and uses the default (keyboard)
- When writing, Eden sets `\default=false` if the value differs from default

### GUID handling in SDL driver

Location: `src/input_common/drivers/sdl_driver.cpp` (lines 15-21)

```cpp
Common::UUID GetGUID(SDL_Joystick* joystick) {
    const SDL_JoystickGUID guid = SDL_JoystickGetGUID(joystick);
    std::array<u8, 16> data{};
    std::memcpy(data.data(), guid.data, sizeof(data));
    std::memset(data.data() + 2, 0, sizeof(u16));  // Clear controller name CRC
    return Common::UUID{data};
}
```

Eden zeroes bytes 2-3 (controller name CRC) to normalize GUIDs from the same model. But the `ff11` vs `0512` difference is in the product ID portion (bytes 8-9), so this normalization doesn't help.

### Controller detection flow

Location: `src/input_common/drivers/sdl_driver.cpp` (lines 346-403)

1. Controller connects → `SDL_JOYDEVICEADDED` triggers `InitJoystick()`
2. GUID computed via `GetGUID()`
3. Joystick stored in `joystick_map[guid]`
4. No fallback when GUID doesn't match stored config

### Device matching in UI

Location: `src/yuzu/configuration/configure_input_player.cpp` (lines 920-975)

```cpp
const auto devices_it = std::find_if(
    input_devices.begin(), input_devices.end(),
    [first_engine, first_guid, first_port, first_pad](const Common::ParamPackage& param) {
        return param.Get("engine", "") == first_engine &&
               param.Get("guid", "") == first_guid &&
               param.Get("port", 0) == first_port;
    });
```

Matching requires exact GUID match. No fuzzy matching, no fallback by device name.

### The "Any" device option

Location: `src/input_common/main.cpp` (lines 133-167)

```cpp
Common::ParamPackage{{"display", "Any"}, {"engine", "any"}},
```

When `engine=any`, `GetInputEngine()` returns nullptr - it accepts input from any source. But button mappings still embed specific GUIDs, so "Any" only affects device selection UI, not binding matching.

## EmuDeck analysis

EmuDeck uses the same approach as kyaraben:
- Same `ff11` GUID: `03000000de280000ff11000001000000`
- Static config copied via rsync during initialization
- No dynamic mode detection
- No multiple GUID fallback
- No special handling for desktop vs game mode

EmuDeck scripts: `functions/EmuScripts/emuDeckEden.sh` (lines 78-110)

They simply copy the same config and expect it to work. If desktop mode doesn't work for them either, they may not care since their target is Game Mode via Steam.

## Solution analysis

### Option 1: Write bindings for both GUIDs

Write `player_0_button_a` with `ff11` GUID and `player_1_button_a` with `0512` GUID. Problem: wastes a player slot, confusing UX.

### Option 2: Use device name instead of GUID

Eden's binding format requires GUID. Unlike Dolphin (which uses `Device = SDL/0/Steam Deck Controller`), Eden embeds GUID in every binding string. No way to use device name.

### Option 3: Don't write controller bindings

Rely on Eden's auto-detection. But Eden defaults to keyboard when no bindings exist (no auto-detect for controllers).

### Option 4: Ship two profiles

Create `Kyaraben-Desktop.ini` (0512) and `Kyaraben-GameMode.ini` (ff11). Users must manually select the appropriate profile.

### Option 5: Accept Game Mode only

Document that controller config works in Game Mode via ES-DE. Desktop mode requires manual configuration. This matches EmuDeck's apparent approach.

### Option 6: Patch Eden

Contribute a PR to Eden that adds GUID fuzzy matching or device-name-based matching. Long-term solution but requires upstream acceptance.

## Decision: Eden

Ship Steam Deck Game Mode support only for now.

### What we're doing

1. Ship `ff11` GUID bindings (Steam Virtual Gamepad) - works in Game Mode via ES-DE
2. Document the limitation in kyaraben docs site
3. Prepare for hardware detection to ship device-specific configs in the future

### Rationale

- Eden's controller support has fundamental limitations (GUID-specific bindings, no fallback, no auto-detection)
- Shipping multiple profiles doesn't help because Eden won't auto-select them
- EmuDeck uses the same approach
- Primary use case (Game Mode via ES-DE) works correctly

### Future work

- Hardware detection already exists for Steam Deck
- Can extend to detect other controllers and ship appropriate GUIDs
- Long-term: consider upstream PR to Eden for device-name-based matching

---

# Dolphin investigation

## Root cause

Two issues prevented Dolphin controller from working:

1. **Missing SIDevice setting**: Dolphin.ini needs `SIDevice0 = 6` to enable GameCube controller port
2. **Invalid device name**: `SDL/0/Gamepad` is a fallback placeholder, not a real device

## Device name approach

Dolphin uses device names instead of GUIDs:
- `SDL/0/Steam Deck Controller` (EmuDeck's approach)
- `SDL/0/Gamepad` (kyaraben's old approach - doesn't work)

The device name "Steam Deck Controller" is the same in both desktop and game mode on Steam Deck. This is better than Eden's GUID approach where the GUID changes between modes.

## Fixes applied

1. Added `SIDevice0 = 6` to Dolphin.ini (enables GC controller port 1)
2. Changed device from `SDL/0/Gamepad` to `SDL/0/Steam Deck Controller`

## Decision: Dolphin

Ship Steam Deck support only for now.

### What we're doing

1. Use `SDL/X/Steam Deck Controller` as device name
2. Document the limitation in kyaraben docs site
3. Prepare for hardware detection to ship device-specific configs in the future

### Rationale

- Device name is Steam Deck specific
- Other controllers would have different device names (e.g., "Xbox 360 Controller")
- EmuDeck uses the same approach
- Primary use case (Steam Deck via ES-DE) works correctly

### Advantage over Eden

Unlike Eden (where GUID changes between desktop/game mode), Dolphin's device name "Steam Deck Controller" stays the same in both modes. This means Dolphin should work in desktop mode too on Steam Deck.

---

# Overall strategy: hardware detection

## The problem with profiles

Neither Eden nor Dolphin auto-selects profiles based on connected controller:

| Emulator | Device identifier | Auto-select profile | Fallback if device missing |
|----------|-------------------|---------------------|---------------------------|
| Eden | GUID (embedded in bindings) | No | No |
| Dolphin | Device name (in Device field) | No | No |

Shipping multiple profiles doesn't help because:
- Eden: profiles are manually selected, and each profile embeds a specific GUID
- Dolphin: profiles are manually selected, and each profile specifies a device name

## The solution: apply-time hardware detection

Instead of runtime profile selection, use hardware detection at `kyaraben apply` time:

1. Detect what hardware/controller is available
2. Write the appropriate device name/GUID to config files
3. Ship profiles as backup for manual override only

Example detection logic:
- Detect Steam Deck → write `Steam Deck Controller` / `ff11` GUID
- Detect Xbox controller → write `Xbox 360 Controller` / appropriate GUID
- Detect unknown → skip controller config or use sensible default

## Current state

Kyaraben already has hardware detection for Steam Deck. Extension needed:
- Detect connected controllers (via SDL or /proc/bus/input/devices)
- Map controller types to appropriate device names/GUIDs
- Allow user override via kyaraben settings

## Why this works

- Config is written once at apply time, not at runtime
- No need for emulators to support auto-selection
- Users can re-run `kyaraben apply` if they change controllers
- Profiles remain available for manual override

---

# Cemu investigation

## Current status

Pending investigation.
