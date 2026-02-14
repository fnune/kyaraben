# Feedback

## Philosophy

Kyaraben's approach to emulator config is minimal: only edit what's strictly necessary to make games playable and enjoyable. We configure paths, disable annoying prompts, and set up ES-DE integration, but we don't touch performance tuning, graphics settings, or advanced options.

Future possibilities:
- Controller/hotkey configuration (see testing/plans/controller-support-plan.md)
- Basic shader/overlay presets
- Hardware presets (e.g., "beefy desktop" vs "Steam Deck" vs "low-power handheld") rather than per-setting tweaks

What we won't do: full performance tuning, per-game settings, target-specific optimizations. Users who want that level of control can configure emulators directly. Keeping things simple is a goal.

(This should eventually be documented on the docs site.)

## Low-hanging fruit

- Refactor Provision struct: Filename, Hashes, and FilePattern are mutually exclusive validation strategies but modeled as optional fields on the same struct. Consider an interface or union type: `FilenameProvision` (just check file exists), `HashedProvision` (check file exists with valid hash), `DirectoryProvision` (check directory contains files matching a glob pattern like `*.nca`). This would make the validation logic cleaner and prevent invalid combinations.
- Cemu says it has required provisions but games launch fine: investigate whether the provision check is wrong or if Cemu has fallback behavior
- Provisions that require importing things via UI will always remain incomplete: need a way to check that they're working. How might we detect this?
- Dolphin autoupdate prompt: needs a default config to disable the built-in autoupdate mechanism
- Backup prompt for opaque-dir emulators: emulators like Vita3K store config inside their opaque directory, which triggers "Create backups before modifying?" prompt on every apply. Need to figure out how to handle config files that live inside opaque dirs
- Audit system extensions against ES-DE bundled config: ensure kyaraben's extension lists are complete for each system
- Flycast CLI hotkey issue: save state hotkeys don't work when launching via CLI (known upstream issue)
- Kyaraben should fetch provision status on focus so that when users add files and come back to Kyaraben to check things update
    - We could even show a toast notification if we find something new
- Cheats directory layout: decide between per-emulator (`~/Emulation/cheats/{emulator}/`) or per-system (`~/Emulation/cheats/{system}/`). Some emulators support configurable cheat paths (melonDS, Flycast, PCSX2)
- DLC, patches and updates directory layout: similar to cheats, figure out folder structure for user-provided DLC and game updates. This could help solve the provision problem where some files must be imported via emulator UI (e.g., Cemu keys.txt, 3DS system files). If kyaraben manages these directories, we could check for installed content. Wii U title structure: 00050000 (games), 0005000c (DLC), 0005000e (patches) per WiiUBrew
- Eden provision summary display: when optional provisions are satisfied (e.g., firmware found), the "(1 optional)" text shows in yellow instead of green. Should satisfied optional provisions show as green?
- Provision summary in the systems list: the clickable area does not span the whole width of the bar. Clickable items within it that don't open the provisions modal should have stoppropagation or similar, but the whole bar should be clickable to open the provisions modal.

## Important

- Expand BIOS hash data: import comprehensive hash alternatives from EmuDeck/RetroDECK into provision Hashes arrays. Current data is minimal; these projects have 30+ PSX hashes, 71+ PS2 hashes, etc.
- Environment variable security: KYARABEN_* env vars (KYARABEN_RELEASES_URL, KYARABEN_VERSION, KYARABEN_NIX_PORTABLE_PATH) are useful for testing but could be risky in production if accidentally set. Consider adding a "test mode" flag that must be set to enable these overrides, or prefix them with KYARABEN_TEST_ to make intent clear
- How do we make the user store dir easier to navigate?
    - Can we integrate with file managers to provide icons for each folder?
    - Is it confusing that some things are per-system and others are per-emulator?
    - Is it a good idea to set up symlinks for convenience or is this more confusing?
    - Is it a good idea to add text for users such as README.md files? Or any other more popular format?
    - What else can be improved?
- Since we moved to the new style, contrast isn't great. It's not easy to tell which emulators are installed and which aren't, and disabled styles leave things with very low contrast.
- We add symlinks for emulator for which we can't configure routes
    - What should we do with those symlinks when Kyaraben uninstalls?
- Garbage collection: nix store grows unbounded, need a way to trigger cleanup via nix-portable and show space freed
- What happens when an emulator updates versions and our config setup no longer works? Can we version our strategy for each emulator?
- ES-DE as non-Steam application: add to Steam for Steam Deck game mode launch
- ARM (aarch64) support: needed for Raspberry Pi, Pinebook Pro, Apple Silicon with Asahi Linux, and other ARM Linux devices. Analysis:
  - Code already supports ARM: hardware detection returns aarch64 target, flake generation handles single-arch and multi-arch packages gracefully
  - nix-portable: has aarch64 build ✓
  - Emulators with ARM builds already in versions.toml: Eden, DuckStation, PPSSPP, mGBA, Dolphin, melonDS
  - Emulators with upstream ARM builds to add: Vita3K (AppImage), RPCS3 (AppImage, tested on Asahi), ES-DE (experimental AppImage)
  - Emulators without ARM builds: PCSX2, Cemu, Azahar, Flycast, RetroArch (no AppImage but Flatpak exists)
  - Work needed: (1) add ARM targets to versions.toml for Vita3K/RPCS3/ES-DE, (2) gracefully hide unavailable emulators in UI when on ARM instead of showing them with no download option, (3) consider Flatpak fallback for RetroArch on ARM, (4) test on real ARM hardware
  - Note: some emulators (PCSX2, Cemu) may never have ARM builds due to low-level x86 translation requirements
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
- Hydra cache miss for libretro-genesis-plus-gx: this core builds from source on apply, which is slow. May need to report upstream to nixpkgs or check if the package is misconfigured
- Allow Apply without config changes: users may need to re-apply after updating kyaraben to benefit from new managed config changes. Could store kyaraben version in manifest to detect when user updated but hasn't applied yet
- Version tracking script: programmatic way to check for new emulator versions, compare against versions.toml, create PRs when updates available
- Emulator health check in doctor: verify installed emulators work (check binaries exist, wrapper scripts valid, AppImage integrity)
- Backend preview command: move apply status logic (shared emulators, will-install/update/uninstall) from frontend to backend `CommandTypeApplyPreview`
- UI show all available emulators: when a system is enabled, show all available emulators with checkboxes instead of a dropdown that switches
- Sync UI: dedicated UI for pairing devices, monitoring sync status (currently CLI/config only)
- Handheld distro sync targets: support SD-based handhelds (Trimui Brick, Miyoo, etc.) running community distros (NextUI, PakUI, MinUI). These distros already bundle emulators, so kyaraben would skip nix-based installation and instead generate configs and directory layouts matching the target distro's conventions, then sync ROMs, saves, and configs to the SD card via Syncthing or export. Add a "target" concept to config (e.g. `target = "trimui-brick-nextui"`) so users declare their systems once and kyaraben translates to the right structure per platform

## Probably not worth it

- Longer folder names for systems: short names like `3ds`, `nds`, `psx` could be `nintendo-3ds`, `nintendo-ds`, `playstation` for clarity, or `n3ds` so 3DS sorts after NDS chronologically. However, ES-DE expects these exact short folder names for scraping, theming, and game detection. Changing them would break ES-DE integration. Since users mostly browse via the frontend UI rather than file managers, the alphabetical sorting issue is minor
- Reconsider flake generations: each apply creates new generation directory and lock file with warning. Works fine, unclear benefit from changing
- Reconsider the manifest: could derive state from filesystem instead of tracking in manifest, but risky refactor for unclear benefit
- Controller configuration abstraction: every emulator handles this differently. Detailed investigation in `testing/plans/controller-support-plan.md` shows a path forward (semantic hotkey model, per-emulator translators), but scope is still large. 5 emulators support controller hotkeys via config (DuckStation, Dolphin, PPSSPP, PCSX2, RetroArch); the rest need Steam Input
- Home-manager module: very niche audience, unclear if AppImages work on NixOS
- Reduce download size: 190MB is acceptable for one-time download, effort better spent elsewhere
