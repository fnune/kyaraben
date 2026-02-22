# Feedback

## Philosophy

Kyaraben's approach to emulator config is minimal: only edit what's strictly necessary to make games playable and enjoyable. We configure paths, disable annoying prompts, and set up ES-DE integration, but we don't touch performance tuning, graphics settings, or advanced options.

Future possibilities:

- Controller/hotkey configuration (see plans/controller-support-plan.md)
- Basic shader/overlay presets
- Hardware presets (e.g., "beefy desktop" vs "Steam Deck" vs "low-power handheld") rather than per-setting tweaks

What we won't do: full performance tuning, per-game settings, target-specific optimizations. Users who want that level of control can configure emulators directly. Keeping things simple is a goal.

(This should eventually be documented on the docs site.)

## Low-hanging fruit

- Sync tab shows "Start pairing (no devices connected)" even when devices are paired
  - Observed: opened Sync tab on primary, showed no devices, tried to start pairing, after a while the Steam Deck appeared
  - Root cause: race condition during Syncthing startup
    - `IsRunning()` checks `/rest/system/ping` which succeeds early
    - `GetConfiguredDevices()` calls `/rest/config/devices` which may return empty before config fully loaded
    - Backend returns `SyncStatusResponse` with empty `Devices` field
    - Frontend sees `hasDevices = false` and shows "Start pairing"
  - Fix options:
    1. Backend fallback: Read devices from `{stateDir}/syncthing/config/config.xml` when API returns empty
    2. Backend retry: Retry `GetConfiguredDevices()` with backoff if empty
    3. Frontend: Show loading state instead of "Start pairing" when running but devices unexpectedly empty
  - Recommended: Option 1 (config.xml fallback) - most robust, handles API unavailability gracefully

- Controller config gets written as this:
    [controller]
    ```
      layout = ""
      [controller.hotkeys]
        save_state = ""
        load_state = ""
        next_slot = ""
        prev_slot = ""
        fast_forward = ""
        rewind = ""
        pause = ""
        screenshot = ""
        quit = ""
        toggle_fullscreen = ""
        open_menu = ""
    ```
    - That's not good! It should be showing the default keybindings
    - Root cause: config was created before controller defaults were added to NewDefaultConfig()
    - Empty strings should mean "do not set this hotkey" (intentionally disabled)
    - TODO: need migration logic or different serialization to distinguish "not set" from "disabled"
- DuckStation and PCSX2 both show the emulator before launching the game
    - Can this be prevented?
- ES-DE loads pcengine games (R-Type Complete CD (Japan)) even though Kyaraben has not enabled that system
    - Is this expected?
    - Same for Xbox360 (Mushihimesama)
- Nintendo DS doesn't work (we moved it to RetroArch recently from melonDS)
    - Games don't even launch from ES-DE
    - Root cause: the old `melonds_libretro` core has an executable stack requirement that modern kernels block
    - Error: `cannot enable executable stack as shared object requires: Invalid argument`
    - FIXED: switched to `melondsds_libretro` (melonDS DS) downloaded as standalone core from GitHub releases
- Flycast hotkeys open the menu instead of immediately doing what the hotkey is supposed to do
    - At least for load and save state
- RetroArch save state bindings works
    - Load state binding works, but it fails to load state (at least for GBC) -> Are we configuring the paths correctly?
    - It also fails for SNES which was always using RetroArch, so probably unrelated to the RetroArch change for mGBA
    - Saves to state 0, but tries to load from state auto
    - FIXED: added `savestate_auto_index = true` to shared RetroArch config
- For most retroarch cores we show the retroarch logo
    - But they also have their own logos, I think?
    - We could show the default logo with the Retroarch logo in tiny in the bottom right, overlayed?
- On Steam Deck, clicking on 'Open Syncthing web interface' opens Discover
    - Even though Firefox is installed
    - I suppose this is because I haven't refreshed the session?
    - But I logged out and back in and it still happens
    - Firefox is indeed set as default application
- We never succeed in showing the connected device's name, we always just show 'primary (connected)' -> probably a bug?
- The remove device trashcan button has no confirmation step
    - It should show confirmation and inform the user as to what's going to happen
- RetroArch cores bundle gets cleaned up after install and thus needs to redownload every time
  - Maybe consider keeping it around instead of cleaning it up
- Add a "Disable all systems" button for convenience
- Add a way to apply even if no changes have happened (maybe even via keyboard shortcut for development)
- Cheats directory layout: decide between per-emulator (`~/Emulation/cheats/{emulator}/`) or per-system (`~/Emulation/cheats/{system}/`). Some emulators support configurable cheat paths (melonDS, Flycast, PCSX2)
- DLC, patches and updates directory layout: similar to cheats, figure out folder structure for user-provided DLC and game updates. This could help solve the provision problem where some files must be imported via emulator UI (e.g., Cemu keys.txt, 3DS system files). If kyaraben manages these directories, we could check for installed content. Wii U title structure: 00050000 (games), 0005000c (DLC), 0005000e (patches) per WiiUBrew
## Important

- Environment variable security: KYARABEN*\* env vars (KYARABEN_RELEASES_URL, KYARABEN_VERSION, KYARABEN_NIX_PORTABLE_PATH) are useful for testing but could be risky in production if accidentally set. Consider adding a "test mode" flag that must be set to enable these overrides, or prefix them with KYARABEN_TEST* to make intent clear
- How do we make the user store dir easier to navigate?
  - Can we integrate with file managers to provide icons for each folder?
  - Is it confusing that some things are per-system and others are per-emulator?
  - Is it a good idea to set up symlinks for convenience or is this more confusing?
  - Is it a good idea to add text for users such as README.md files? Or any other more popular format?
  - What else can be improved?
- Since we moved to the new style, contrast isn't great. It's not easy to tell which emulators are installed and which aren't, and disabled styles leave things with very low contrast.
- Systems view needs rethinking:
  - Needs a way to search/filter emulators faster
  - Hard to quickly visualize what is and isn't installed
  - Fights semantically with Installation view: Systems shows what will/won't be installed, Installation shows raw paths and tools to uninstall/update kyaraben itself
  - Not clear if this scales well or is easy to understand
- What happens when an emulator updates versions and our config setup no longer works? Can we version our strategy for each emulator?
- ES-DE as non-Steam application: add to Steam for Steam Deck game mode launch
- Steam Deck support: the Steam Deck uses an AMD Zen 2 x86_64 APU (not ARM), so all current binaries are compatible. Some emulators already have Steam Deck-specific builds (Eden, ES-DE). Remaining concerns:
  - SteamOS immutable filesystem: nix-portable stores in ~/.nix-portable which should work, but needs testing on a real device to verify permissions and available disk space on the internal SSD vs SD card
  - Gaming Mode integration: users primarily run the Deck in Gaming Mode. Need a way to add ES-DE (or individual emulators) to Steam so they appear in the library and can be launched without switching to Desktop Mode. Could generate .desktop files and use `steam-rom-manager` or `steamos-add-to-steam` to register them
  - Controller input: Steam Input may intercept controller events before they reach emulators. Need to test whether emulators receive correct input, and whether Steam's controller configuration for each emulator is needed
  - Performance profiles: Steam Deck allows per-game TDP/GPU limits. Emulator wrapper scripts could potentially integrate with `gamescope` or Steam's performance overlay, but this may be out of scope
  - Installation path: default ~/.local/share/kyaraben may compete for limited internal storage. Consider detecting Steam Deck and recommending SD card installation, or prompting user during init

## Nice to have

- Hotkey ergonomics: Select + D-pad combinations are awkward on Steam Deck. Consider alternative default bindings
- RetroArch autosave on exit: EmuDeck has `RetroArch_autoSaveOn()` but off by default. Unclear if supportable across all emulators
- RetroAchievements integration: global credential storage with per-emulator login. Supported by DuckStation, PCSX2, PPSSPP, RetroArch cores, and Dolphin (experimental)
- Cross-emulator presets: toggle high-level features that cascade to all compatible emulators (widescreen, integer scaling, Discord presence, RetroAchievements, auto-save on exit)
- Performance defaults: ship sensible defaults for renderer (Vulkan), resolution scale, recompilers, fast boot. Currently kyaraben focuses on paths but users must manually configure performance settings
- Shader and overlay management: CRT shaders, bezel overlays, per-core shader presets, HD texture pack paths
- Compression guidance: tools or documentation for ROM compression (CHD for disc-based, RVZ for GameCube/Wii, ZIP for cartridge-based)
- Additional emulator paths: configure more optional directories. Flycast (BoxartPath, MappingsPath, TexturePath), PCSX2 (Cheats, Covers, Videos)
- Storage breakdown bar: below the "Emulation folder" input in the UI, show a color-coded bar indicating total size and composition of the directory (ROMs, saves, opaque dirs, etc.)
- Version tracking script: programmatic way to check for new emulator versions, compare against versions.toml, create PRs when updates available
- Emulator health check in doctor: verify installed emulators work (check binaries exist, wrapper scripts valid, AppImage integrity)
- Backend preview command: move apply status logic (shared emulators, will-install/update/uninstall) from frontend to backend `CommandTypeApplyPreview`
- Sync UI: dedicated UI for pairing devices, monitoring sync status (currently CLI/config only)
- Handheld distro sync targets: support SD-based handhelds (Trimui Brick, Miyoo, etc.) running community distros (NextUI, PakUI, MinUI). These distros already bundle emulators, so kyaraben would skip nix-based installation and instead generate configs and directory layouts matching the target distro's conventions, then sync ROMs, saves, and configs to the SD card via Syncthing or export. Add a "target" concept to config (e.g. `target = "trimui-brick-nextui"`) so users declare their systems once and kyaraben translates to the right structure per platform

## Probably not worth it

- Reconsider the manifest: could derive state from filesystem instead of tracking in manifest, but risky refactor for unclear benefit
- Home-manager module: very niche audience, unclear if AppImages work on NixOS
- Reduce download size: 190MB is acceptable for one-time download, effort better spent elsewhere
- Eden/Cemu splash screens: emulator loading screens are ugly but can't be avoided
