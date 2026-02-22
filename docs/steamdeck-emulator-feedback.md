# Steam Deck emulator configuration feedback

Tracking configuration problems discovered during testing. EmuDeck serves as reference for potential solutions.

Testing environment: Steam Deck Game Mode -> EmulationStation Desktop Edition (ESDE) -> launch game from there.

---

## Low-hanging fruit (DONE)

Simple config changes where EmuDeck shows the solution.

### Cemu (Wii U) - DONE

- [x] Lots of shader compilation notifications, gets spammy.
  - Fix: Set `<ShaderCompiling>false</ShaderCompiling>` in settings.xml.

- [x] Very low FPS on Breath of the Wild.
  - Fix: Use Vulkan (`api=1`), `VSync=0`, `AsyncCompile=true`, `GX2DrawdoneSync=true`.

### Dolphin (GameCube/Wii) - DONE

- [x] GameCube looks pixelated (tested on Wind Waker).
  - Fix: Set `InternalResolution = 2` in GFX.ini (2x native).

- [x] Exit hotkey shows confirmation dialog "Do you want to stop the current emulation".
  - Fix: Set `ConfirmStop = False` in Dolphin.ini.

### PCSX2 (PS2) - DONE

- [x] Shutdown asks for confirmation with save state checkbox, requires touchscreen to dismiss.
  - Fix: Set `ConfirmShutdown = false` in PCSX2.ini under [UI].

### RetroArch - DONE

- [x] Analog stick does not translate to D-pad inputs.
  - Fix: Set `input_player1_analog_dpad_mode = "1"`.

### N64 (RetroArch) - DONE

- [x] Could default to higher resolution.
  - Fix: Set `mupen64plus-43screensize = "1280x960"` and `mupen64plus-169screensize = "1920x1080"` in core options.

---

## Needs investigation

Config exists but something isn't working, or requires more complex setup.

### RetroArch (general)

- [x] Quit keybind needs to be pressed twice to exit.
  - Fix: Set `quit_press_twice = "false"`.

- [x] Fonts not loading. Possibly broken by recent asset extraction work (PR #31).
  - Fix: Extract assets/autoconfig subdirs even when config dir already exists.

- [x] Notification: "Valve streaming [controller] not configured, using fallback".
  - Fix: Same as above - autoconfig now extracted from AppImage.

### N64 (RetroArch)

- [x] A and B button placement inconsistent with SNES. On SNES, Steam Deck B = Nintendo A. On N64, Steam Deck A = Nintendo A.
  - Fix: Added explicit RetroPad face button mapping to SharedConfig using FaceButtons(). All RetroArch cores now use consistent button mappings based on the layout setting.

- [x] Hotkeys appear to be in wrong position.
  - Fix: Made hotkey button indices layout-aware via SDLIndex() method. Face button hotkeys now transform through the layout setting.

### Duckstation (PS1)

- Exit hotkey not working (does nothing).
  - EmuDeck: Uses `PowerOff = SDL-0/Back & SDL-0/Start` chord.

### Eden (Switch)

- [x] Button mappings wrong. In Link's Awakening: left trigger opens map, select opens inventory, start does nothing.
  - Fix: Eden uses raw SDL joystick indices (not GameController). Added Steam Deck-specific button indices matching EmuDeck's configuration.

- [x] Most hotkeys not working.
  - Fix: Updated hotkey button names to match Eden's expected format (e.g., Right_Stick instead of Rstick).

### PCSX2 (PS2)

- Exit hotkey goes to emulator UI instead of exiting to ESDE.
  - EmuDeck: Sets `StartFullscreen = true` but `HideMainWindowWhenRunning = false`. Needs investigation.

### PPSSPP (PSP)

- Back + A opens save state menu. Actual save/load state hotkeys don't work.
  - EmuDeck: `Save State = 10-196:10-192,1-132`, `Load State = 10-196:10-193,1-133`.

### Dolphin (GameCube/Wii)

- Hotkeys other than exit don't work.
  - EmuDeck: Sets `Device = SDL/0/Steam Deck Controller` in Hotkeys.ini. Bindings: `Back+Button S`=pause, `Back+Button W`=stop, `Back+Start`=exit, `Back+Shoulder L/R`=load/save state.

### Azahar (3DS)

- Inputs not working at all.
  - EmuDeck: Uses GUID `030079f6de280000ff11000001000000` with SDL engine. Full mapping in qt-config.ini.

### Flycast (Dreamcast)

- Pressing modifier key alone opens Flycast menu. Expected: modifier + key triggers action.
  - EmuDeck: Maps `btn_menu` directly to button 6. Flycast has no "hotkey enable" mechanism. May be emulator limitation.

---

## Nitpicks / future considerations

Low priority or unclear if fixable.

- (General) Select + D-pad hotkey combinations are ergonomically awkward on Steam Deck.

- (General) TODO: Review EmuDeck's performance/resolution defaults for all supported cores and emulators.

- (RetroArch) No autosave on exit. EmuDeck has `RetroArch_autoSaveOn()` but off by default. Unclear if supportable across all emulators.

- (PPSSPP) Fast forward binding is hold, but most other emulators use toggle. EmuDeck has both `Fast-forward` (hold) and `SpeedToggle` (toggle).

- (PPSSPP) Exit shows confirmation "you haven't saved". No disable setting found in EmuDeck.

- (Eden) Emulator loading/splash screen is ugly. Can't be avoided.

- (Cemu) Emulator loading/splash screen is ugly. Can't be avoided.

---

## EmuDeck Eden config reference

Performance settings we may be missing:
- `use_multi_core=true`
- `backend=1` (Vulkan), `gpu_accuracy=0` (fastest)
- `use_asynchronous_gpu_emulation=true`, `use_asynchronous_shaders=true`
- `use_disk_shader_cache=true`, `use_fast_gpu_time=true`
- `resolution_setup=2` (2x scale), `scaling_filter=5` (FSR), `fsr_sharpening_slider=25`
- `use_vsync=2`, `fullscreen_mode=1`, `fps_cap=1000`
- `use_docked_mode=1`
- CPU accuracy=0 (fastest), all JIT optimizations enabled
- Uses GUID `03000000de280000ff11000001000000` for controller
