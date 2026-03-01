# Feedback

## Philosophy

Kyaraben's approach to emulator config is minimal: only edit what's strictly necessary to make games playable and enjoyable. We configure paths, controller bindings, hotkeys, disable annoying prompts, and set up ES-DE integration, but we don't touch performance tuning, graphics settings, or advanced options.

Future possibilities:

- Basic shader/overlay presets
- Hardware presets (e.g., "beefy desktop" vs "Steam Deck" vs "low-power handheld") rather than per-setting tweaks

What we won't do: full performance tuning, per-game settings, target-specific optimizations. Users who want that level of control can configure emulators directly. Keeping things simple is a goal.

(This should eventually be documented on the docs site.)

## Low-hanging fruit

- For most RetroArch cores we show the RetroArch logo
    - But they also have their own logos, I think?
    - We could show the default logo with the RetroArch logo in tiny in the bottom right, overlayed?
- Add a "Disable all systems" button for convenience
- Add a way to apply even if no changes have happened (maybe even via keyboard shortcut for development)
- RetroArch has default hotkeys that can disrupt gameplay unexpectedly. Example: while playing R-Type on PC Engine, pressing an unknown combo triggered "Switching to 6 button layout" and made the game unplayable. Need to investigate which core-specific hotkeys are enabled by default and consider disabling them
- Final Fantasy Tactics shows up twice in ES-DE (needs investigation: PSX + PSP versions, or actual duplicate?)
- Analog to dpad translation (`input_playerN_analog_dpad_mode`) may not be applying to existing configs due to DefaultOnly flag - users with pre-existing retroarch.cfg won't get this setting
- Cheats directory layout: decide between per-emulator (`~/Emulation/cheats/{emulator}/`) or per-system (`~/Emulation/cheats/{system}/`). Some emulators support configurable cheat paths (Flycast, PCSX2)
- DLC, patches and updates directory layout: similar to cheats, figure out folder structure for user-provided DLC and game updates. This could help solve the provision problem where some files must be imported via emulator UI (e.g., Cemu keys.txt). If kyaraben manages these directories, we could check for installed content. Wii U title structure: 00050000 (games), 0005000c (DLC), 0005000e (patches) per WiiUBrew

## Important

- Environment variable security: KYARABEN*\* env vars (KYARABEN_RELEASES_URL, KYARABEN_VERSION, KYARABEN_NIX_PORTABLE_PATH) are useful for testing but could be risky in production if accidentally set. Consider adding a "test mode" flag that must be set to enable these overrides, or prefix them with KYARABEN_TEST* to make intent clear
- How do we make the user store dir easier to navigate?
  - Can we integrate with file managers to provide icons for each folder?
  - Is it confusing that some things are per-system and others are per-emulator?
  - Is it a good idea to set up symlinks for convenience or is this more confusing?
  - Is it a good idea to add text for users such as README.md files? Or any other more popular format?
  - What else can be improved?
- What happens when an emulator updates versions and our config setup no longer works? Can we version our strategy for each emulator?
- Steam Deck support: the Steam Deck uses an AMD Zen 2 x86_64 APU (not ARM), so all current binaries are compatible. Controller configuration works via Steam Input GUID virtualization. Remaining concerns:
  - Performance profiles: Steam Deck allows per-game TDP/GPU limits. Emulator wrapper scripts could potentially integrate with `gamescope` or Steam's performance overlay, but this may be out of scope
  - Installation path: default ~/.local/share/kyaraben may compete for limited internal storage. Consider detecting Steam Deck and recommending SD card installation, or prompting user during init

## Nice to have

- RetroArch autosave on exit: EmuDeck has `RetroArch_autoSaveOn()` but off by default. Unclear if supportable across all emulators
- RetroAchievements integration: global credential storage with per-emulator login. Supported by DuckStation, PCSX2, PPSSPP, RetroArch cores, and Dolphin (experimental)
- Cross-emulator presets: toggle high-level features that cascade to all compatible emulators (widescreen, integer scaling, Discord presence, RetroAchievements, auto-save on exit)
- Performance defaults: ship sensible defaults for renderer (Vulkan), resolution scale, recompilers, fast boot. Currently kyaraben focuses on paths but users must manually configure performance settings
- Shader and overlay management: CRT shaders, bezel overlays, per-core shader presets, HD texture pack paths
- Compression guidance: tools or documentation for ROM compression (CHD for disc-based, RVZ for GameCube/Wii, ZIP for cartridge-based)
- Additional emulator paths: configure more optional directories. Flycast (BoxartPath, MappingsPath, TexturePath), PCSX2 (Cheats, Covers, Videos)
- Storage breakdown bar: below the "Emulation folder" input in the UI, show a color-coded bar indicating total size and composition of the directory (ROMs, saves, opaque dirs, etc.)
- Handheld distro sync targets: support SD-based handhelds (Trimui Brick, Miyoo, etc.) running community distros (NextUI, PakUI, MinUI). These distros already bundle emulators, so kyaraben would skip emulator installation and instead generate configs and directory layouts matching the target distro's conventions, then sync ROMs, saves, and configs to the SD card via Syncthing or export. Add a "target" concept to config (e.g. `target = "trimui-brick-nextui"`) so users declare their systems once and kyaraben translates to the right structure per platform

## Probably not worth it

- Reconsider the manifest: could derive state from filesystem instead of tracking in manifest, but risky refactor for unclear benefit
- Reduce download size: 190MB is acceptable for one-time download, effort better spent elsewhere
- Eden/Cemu splash screens: emulator loading screens are ugly but can't be avoided
