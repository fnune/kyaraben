# Feedback

## Low-hanging fruit

- DuckStation onboarding wizard: needs a default config to prevent the wizard from appearing on first launch
- Dolphin autoupdate prompt: needs a default config to disable the built-in autoupdate mechanism

## Important

- Environment variable security: KYARABEN_* env vars (KYARABEN_RELEASES_URL, KYARABEN_VERSION, KYARABEN_NIX_PORTABLE_PATH) are useful for testing but could be risky in production if accidentally set. Consider adding a "test mode" flag that must be set to enable these overrides, or prefix them with KYARABEN_TEST_ to make intent clear
- Garbage collection: nix store grows unbounded, need a way to trigger cleanup via nix-portable and show space freed
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
- Controller configuration abstraction: every emulator handles this differently, massive scope
- Home-manager module: very niche audience, unclear if AppImages work on NixOS
- Reduce download size: 190MB is acceptable for one-time download, effort better spent elsewhere
