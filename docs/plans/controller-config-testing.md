# Controller configuration testing guide

## Default hotkey bindings (Xbox layout reference)

| Action           | Chord                  |
| ---------------- | ---------------------- |
| Save state       | Back + RB              |
| Load state       | Back + LB              |
| Next slot        | Start + RB             |
| Previous slot    | Start + LB             |
| Fast forward     | Back + RT              |
| Rewind           | Back + LT              |
| Pause            | Back + A               |
| Screenshot       | Back + B               |
| Quit             | Back + Start           |
| Toggle fullscreen| Start + L3             |
| Open menu        | Back + R3              |

For SNES USB controller: Back = Select, Start = Start, RB/LB = R/L.
The SNES controller lacks triggers, sticks, and extra buttons, so only hotkeys
using Select, Start, L, R, and face buttons will work.

---

## Common steps (all emulators)

Run these for every emulator on each platform/controller combination.

### Steam Deck + Xbox controller

- [ ] Launch a game and confirm emulator starts
- [ ] D-pad moves in the correct directions
- [ ] Left analog stick moves in the correct directions
- [ ] Face buttons perform the expected in-game actions
- [ ] Shoulder buttons (LB/RB) register correctly
- [ ] Triggers (LT/RT) register correctly
- [ ] Start and Back/Select buttons register correctly

### Steam Deck + SNES USB controller

- [ ] Launch a game and confirm emulator starts
- [ ] D-pad moves in the correct directions
- [ ] Face buttons (A, B, X, Y) perform the expected in-game actions
- [ ] L and R shoulder buttons register correctly
- [ ] Start and Select buttons register correctly

### Desktop + Xbox controller

- [ ] Launch a game and confirm emulator starts
- [ ] D-pad moves in the correct directions
- [ ] Left analog stick moves in the correct directions
- [ ] Face buttons perform the expected in-game actions
- [ ] Shoulder buttons (LB/RB) register correctly
- [ ] Triggers (LT/RT) register correctly
- [ ] Start and Back/Select buttons register correctly

### Desktop + SNES USB controller

- [ ] Launch a game and confirm emulator starts
- [ ] D-pad moves in the correct directions
- [ ] Face buttons (A, B, X, Y) perform the expected in-game actions
- [ ] L and R shoulder buttons register correctly
- [ ] Start and Select buttons register correctly

---

## Per-emulator testing

### DuckStation (PlayStation)

4-player support, SDL standard names, full hotkey support.

Common steps: refer to the section above.

#### Additional: hotkeys (Xbox controller)

- [ ] Back + RB: saves state
- [ ] Back + LB: loads state
- [ ] Start + RB: advances to the next save slot
- [ ] Start + LB: goes to the previous save slot
- [ ] Back + RT: fast forward engages
- [ ] Back + LT: rewind engages
- [ ] Back + A: pauses emulation
- [ ] Back + B: takes a screenshot
- [ ] Back + R3: opens the pause menu

#### Additional: multiplayer (Xbox controller)

- [ ] Player 2 controller is recognized as Pad2
- [ ] Player 2 inputs map correctly

#### Additional: vibration

- [ ] Rumble works during gameplay (games that support it)

---

### PCSX2 (PlayStation 2)

4-player support, SDL standard names, full hotkey support.

Common steps: refer to the section above.

#### Additional: hotkeys (Xbox controller)

- [ ] Back + RB: saves state
- [ ] Back + LB: loads state
- [ ] Start + RB: advances to the next save slot
- [ ] Start + LB: goes to the previous save slot
- [ ] Back + RT: fast forward (turbo)
- [ ] Back + LT: slow motion
- [ ] Back + A: pauses emulation
- [ ] Back + B: takes a screenshot
- [ ] Back + Start: shuts down VM (quit)
- [ ] Start + L3: toggles fullscreen
- [ ] Back + R3: opens the pause menu

#### Additional: multiplayer (Xbox controller)

- [ ] Player 2 controller is recognized as Pad2
- [ ] Player 2 inputs map correctly

#### Additional: vibration

- [ ] Rumble works during gameplay (games that support it)

#### Additional: analog sticks

- [ ] Right analog stick moves correctly (for dual-analog PS2 games)

---

### Dolphin (GameCube / Wii)

4-player support, Dolphin descriptive names, hotkey support.
Note: GameCube face button layout differs (A=east, B=south, X=north, Y=west).

Common steps: refer to the section above.

#### Additional: hotkeys (Xbox controller)

- [ ] Back + RB: saves state
- [ ] Back + LB: loads state
- [ ] Start + RB: next save slot
- [ ] Start + LB: previous save slot
- [ ] Back + RT: disables speed limit (fast forward)
- [ ] Back + A: pauses emulation
- [ ] Back + B: takes a screenshot
- [ ] Back + Start: exits emulator
- [ ] Start + L3: toggles fullscreen

#### Additional: multiplayer (Xbox controller)

- [ ] Player 2 controller is recognized as GCPad2
- [ ] Player 2 inputs map correctly

#### Additional: analog sticks

- [ ] Left analog mapped as GameCube main stick
- [ ] Right analog mapped as GameCube C-stick
- [ ] Triggers are analog (progressive input, not just on/off)

---

### RetroArch: mGBA (Game Boy / Game Boy Advance)

Now a RetroArch core. Full hotkey support via enable_hotkey + action pattern.

Common steps: refer to the section above.

#### Additional: hotkeys (Xbox controller)

Same as RetroArch: bsnes above.

#### Additional: GBA-specific

- [ ] A and B buttons mapped correctly (GBA only has A and B)
- [ ] L and R shoulder buttons register correctly
- [ ] D-pad works in all directions

---

### RetroArch: melonDS (Nintendo DS)

Now a RetroArch core (melondsds_libretro). Full hotkey support via enable_hotkey + action pattern.

Common steps: refer to the section above.

#### Additional: hotkeys (Xbox controller)

Same as RetroArch: bsnes above.

#### Additional: DS-specific

- [ ] Touch screen input works (mouse or touchpad)

---

### PPSSPP (PlayStation Portable)

Single player, internal keycodes (device 10), hotkey support.

Common steps: refer to the section above.

#### Additional: hotkeys (Xbox controller)

- [ ] Back + RB: saves state
- [ ] Back + LB: loads state
- [ ] Start + RB: next save slot
- [ ] Start + LB: previous save slot
- [ ] Back + RT: fast forward
- [ ] Back + LT: rewind
- [ ] Back + A: pauses emulation
- [ ] Back + B: takes a screenshot
- [ ] Back + Start: exits app

Note: PPSSPP hotkey chords use `:` separator internally.

#### Additional: analog stick

- [ ] Left analog stick maps to PSP analog nub

---

### Flycast (Dreamcast)

Single player, digital/analog split config, no hotkey support.

Common steps: refer to the section above.

#### Additional: Dreamcast-specific

- [ ] Triggers mapped as analog (LT, RT)
- [ ] Left analog stick maps to Dreamcast analog stick
- [ ] A, B, X, Y face buttons match Dreamcast layout

---

### Eden (Nintendo Switch)

2-player support, GUID-embedded bindings, no hotkey support.
Note: Switch face button layout differs (A=east, B=south, X=north, Y=west).

Common steps: refer to the section above.

#### Additional: multiplayer (Xbox controller)

- [ ] Player 2 controller maps to player_1 profile
- [ ] Player 2 inputs register correctly

#### Additional: Switch-specific

- [ ] ZL and ZR (triggers) register correctly
- [ ] Left and right analog sticks work
- [ ] Plus and Minus buttons (Start/Back) register correctly
- [ ] Controller GUID is recognized (check config output)

#### Additional: SNES USB controller

- [ ] Verify GUID matching works (the SNES controller may need a GUID entry)

---

### RetroArch: Citra (Nintendo 3DS)

Now a RetroArch core. Full hotkey support via enable_hotkey + action pattern.

Common steps: refer to the section above.

#### Additional: hotkeys (Xbox controller)

Same as RetroArch: bsnes above.

#### Additional: 3DS-specific

- [ ] Circle pad (left analog) works in all directions
- [ ] C-stick (right analog) works in all directions
- [ ] L and R shoulder buttons register correctly
- [ ] Touch screen input works (mouse or touchpad)

---

### RetroArch: bsnes (SNES)

Auto-detection enabled, full hotkey support via enable_hotkey + action pattern.

Common steps: refer to the section above.

#### Additional: hotkeys (Xbox controller)

- [ ] Back + RB: saves state
- [ ] Back + LB: loads state
- [ ] Start + RB: next save slot
- [ ] Start + LB: previous save slot
- [ ] Back + RT: fast forward
- [ ] Back + LT: rewind
- [ ] Back + A: pauses emulation
- [ ] Back + B: takes a screenshot
- [ ] Back + Start: exits emulator
- [ ] Start + L3: toggles fullscreen
- [ ] Back + R3: opens RetroArch menu

#### Additional: RetroArch-specific

- [ ] Controller auto-detected on launch (no manual binding needed)
- [ ] Saves go to the correct per-core sorted directory (saves/bsnes/)
- [ ] Savestates go to the correct directory (states/bsnes/)
- [ ] Screenshots go to the shared retroarch screenshots directory

---

### RetroArch: Mesen (NES)

Auto-detection enabled, full hotkey support. Same hotkeys as bsnes above.

Common steps: refer to the section above.

#### Additional: hotkeys (Xbox controller)

Same as RetroArch: bsnes above.

#### Additional: RetroArch-specific

- [ ] Saves go to saves/mesen/
- [ ] Savestates go to states/mesen/

---

### RetroArch: Genesis Plus GX (Genesis / Mega Drive)

Auto-detection enabled, full hotkey support. Same hotkeys as bsnes above.

Common steps: refer to the section above.

#### Additional: hotkeys (Xbox controller)

Same as RetroArch: bsnes above.

#### Additional: RetroArch-specific

- [ ] Saves go to saves/genesis_plus_gx/
- [ ] Savestates go to states/genesis_plus_gx/

---

### RetroArch: Mupen64Plus-Next (N64)

Auto-detection enabled, full hotkey support. Same hotkeys as bsnes above.

Common steps: refer to the section above.

#### Additional: hotkeys (Xbox controller)

Same as RetroArch: bsnes above.

#### Additional: RetroArch-specific

- [ ] Saves go to saves/mupen64plus_next/
- [ ] Savestates go to states/mupen64plus_next/

#### Additional: N64-specific

- [ ] Analog stick mapped correctly
- [ ] C-buttons work (mapped from right analog)

---

### RetroArch: Beetle Saturn (Saturn)

Auto-detection enabled, full hotkey support. Same hotkeys as bsnes above.
Requires BIOS (no HLE fallback).

Common steps: refer to the section above.

#### Additional: hotkeys (Xbox controller)

Same as RetroArch: bsnes above.

#### Additional: RetroArch-specific

- [ ] Saves go to saves/mednafen_saturn/
- [ ] Savestates go to states/mednafen_saturn/

#### Additional: Saturn-specific

- [ ] BIOS is detected and loaded
- [ ] L and R triggers map to Saturn shoulder buttons

---

### Vita3K (PlayStation Vita)

Auto-detection, no manual controller config, no hotkey support.

Common steps: refer to the section above.

#### Additional: Vita-specific

- [ ] Controller auto-detected on launch
- [ ] Touch screen emulation works (mouse or touchpad)
- [ ] Left and right analog sticks work
- [ ] L and R triggers register correctly

---

### RPCS3 (PlayStation 3)

Auto-detection, no manual controller config, no hotkey support.

Common steps: refer to the section above.

#### Additional: PS3-specific

- [ ] Controller auto-detected on launch
- [ ] Both analog sticks work
- [ ] L1/R1 and L2/R2 register correctly
- [ ] L3/R3 (stick press) register correctly
- [ ] Sixaxis motion input works if applicable

---

### Cemu (Wii U)

Single player, controller configured by kyaraben. No hotkey support (upstream
limitation, see https://github.com/cemu-project/Cemu/issues/721).

Common steps: refer to the section above.

#### Additional: Wii U-specific

- [ ] Gamepad touch screen emulation works

---

## Notes

- For the SNES USB controller, many emulators will lack triggers, analog sticks,
  and extra buttons. Only test what the controller physically has.
- GUID-based emulators (Eden) may not recognize the SNES USB controller
  unless its GUID is added to the `[controller.guids]` config section.
- RetroArch and Group D emulators (Vita3K, RPCS3) use auto-detection and should
  work with any SDL-compatible controller without extra configuration.
- On Steam Deck, Steam Input may remap the controller. If bindings seem wrong,
  check that Steam Input is set to pass through (no remapping).
- mGBA, melonDS, and 3DS (Citra) are now RetroArch cores, not standalone
  emulators. They use RetroArch's unified hotkey system.
