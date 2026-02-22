# Steam Deck testing round 2

Testing after emulator config fixes.

## Cemu (Wii U)

- [ ] Hotkeys don't work (Cemu limitation - https://github.com/cemu-project/Cemu/issues/721)

## Dolphin (GameCube/Wii)

- [x] Fixed: Speed hotkey now uses toggle() wrapper for toggle behavior
- [x] Fixed: Load state key was wrong section (Save State/ -> Load State/)
- [x] Fixed: ConfirmStop was in wrong section (General -> Interface)

## PCSX2 (PS2)

- [ ] Shows emulator UI while loading game

## RetroArch

- [ ] Assets still missing

## DuckStation (PS1)

- [x] Fixed: ConfirmPowerOff = false disables confirmation dialog
- [x] Fixed: SaveStateOnExit = false disables auto-save
- [x] Fixed: Added -batch flag to exit completely instead of returning to GUI

## Eden (Switch)

- [x] Fixed: use_fast_gpu_time -> fast_gpu_time with proper enum value
- [ ] Hotkeys not working (needs further investigation)

## PPSSPP (PSP)

- [x] Fixed: Added FirstRun = False to suppress startup messages
- [x] Fixed: L/R keycodes updated to match EmuDeck (L=193, R=192)

## Azahar (3DS)

- [x] Controls working
- [ ] Hotkeys don't work at all

## Flycast (Dreamcast)

- [x] Fixed: switched to controller_neptune mapping, no btn_menu binding

## ES-DE

- [ ] Final Fantasy Tactics shows up twice
