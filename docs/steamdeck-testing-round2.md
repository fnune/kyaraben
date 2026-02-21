# Steam Deck testing round 2

Testing after emulator config fixes.

## Cemu (Wii U)

- [ ] Hotkeys don't work (Cemu limitation - https://github.com/cemu-project/Cemu/issues/721)

## Dolphin (GameCube/Wii)

- [x] Fixed: Speed hotkey now uses toggle() wrapper for toggle behavior
- [x] Fixed: Load state key was wrong section (Save State/ -> Load State/)
- [x] Fixed: ConfirmStop was in wrong section (General -> Interface)

## PCSX2 (PS2)

- [x] Fixed: Added StartFullscreen = true to hide UI during game load

## RetroArch

- [ ] Assets still missing (requires reinstall to trigger extraction)

## DuckStation (PS1)

- [x] Fixed: ConfirmPowerOff = false disables confirmation dialog
- [x] Fixed: SaveStateOnExit = false disables auto-save
- [x] Fixed: Added -batch flag to exit completely instead of returning to GUI

## Eden (Switch)

- [x] Fixed: use_fast_gpu_time -> fast_gpu_time with proper enum value
- [x] Fixed: Added `\default=false` flags for performance settings (Eden requires this to read custom values)
- [x] Fixed: Added `\default=false` flags for hotkeys (Eden ignores custom values without this)

## PPSSPP (PSP)

- [x] Fixed: Added FirstRun = False to suppress startup messages
- [x] Fixed: L/R keycodes updated to match EmuDeck (L=193, R=192)

## Flycast (Dreamcast)

- [x] Fixed: switched to controller_neptune mapping, no btn_menu binding

## 3DS

- [x] Switched from Azahar (standalone) to RetroArch Citra core for consistent hotkey support

## ES-DE

- [ ] Final Fantasy Tactics shows up twice (needs investigation: is this PSX + PSP versions, or actual duplicate in same system?)
