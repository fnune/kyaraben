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

- RetroArch cores bundle gets cleaned up after install and thus needs to redownload every time
  - Maybe consider keeping it around instead of cleaning it up
- Add a "Disable all systems" button for convenience
- Add a way to apply even if no changes have happened (maybe even via keyboard shortcut for development)
- Cheats directory layout: decide between per-emulator (`~/Emulation/cheats/{emulator}/`) or per-system (`~/Emulation/cheats/{system}/`). Some emulators support configurable cheat paths (melonDS, Flycast, PCSX2)
- DLC, patches and updates directory layout: similar to cheats, figure out folder structure for user-provided DLC and game updates. This could help solve the provision problem where some files must be imported via emulator UI (e.g., Cemu keys.txt, 3DS system files). If kyaraben manages these directories, we could check for installed content. Wii U title structure: 00050000 (games), 0005000c (DLC), 0005000e (patches) per WiiUBrew
- Cemu has a required provision that I think is not actually required. We should set it to optional and investigate/document in the app what it's actually needed for.

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
