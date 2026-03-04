# Feedback

## Technical only

### Proper blocker

### Would be really nice

- Use HTTPS with self-signed certificates for Syncthing's UI
  - This is what Syncthing ships by default
  - Can we do this without sacrificing UX?
- Tight coupling to systemd (`exec.Command("systemctl", ...)`). On non-systemd systems, sync will not work.
  - But where? I thought we had a system manager abstraction
- `setup.Disable()` removes the systemd unit but does not clean up Syncthing config or data directories.
  - Should it? Or should this be left to uninstall? We don't have a "pause sync" feature yet

### Can launch without

- Environment variable security: KYARABEN*\* env vars (KYARABEN_RELEASES_URL, KYARABEN_VERSION, KYARABEN_NIX_PORTABLE_PATH) are useful for testing but could be risky in production if accidentally set. Consider adding a "test mode" flag that must be set to enable these overrides, or prefix them with KYARABEN_TEST* to make intent clear
- What happens when an emulator updates versions and our config setup no longer works? Can we version our strategy for each emulator?
- `config.go` hardcodes Syncthing config `Version: 37`. Low priority since Syncthing is backward compatible, but will eventually need a migration path.
  - But isn't this tied to Syncthing's version? Which we also control?
- Emoji in SyncStatusBanner.tsx (uses ✓, ↻, ○, ⚠, ✕, ●). Should use SVG icons or CSS.
- Inconsistent error wrapping: some use `fmt.Errorf("foo: %w", err)`, others `fmt.Errorf("foo %v", err)`.

## UX gaps

### Proper blocker

- We are not surfacing controller hotkey options in the UI
  - This is hard because the natural way to do this would be using a controller
  - But that would need to happen within Steam because otherwise we don't have Steam Input
  - The alternative is text selection with dropdown options for known keys
- "Shaders" isn't at the right level
  - See https://retrogamecorps.com/2024/09/01/guide-shaders-and-overlays-on-retro-handhelds/comment-page-1/
  - Shaders is only one part of the thing, although this is what people search for
  - There's filters, overlays, bezels, integer scaling...
  - RGC go for three styles, each involving not only shaders but other things as well:
    - Modern pixels
    - Upscaled image
    - Pseudo-authentic
  - I think that level of abstraction is better
  - Although the choice is deeply device-specific, so maybe this is hard
- Documentation lives on the website too much
  - One goal of Kyaraben is to have its app be self-documenting
  - Users are expected to be able to figure things out just from what the app tells them
  - This clashes with having so much information within the docs
  - For example, the emulator support table could ideally be displayed within the app, for each emulator

### Would be really nice

- `DefaultOnly` is uncomfortable if the editor writes out all of its defaults to the config
  - Would it be possible to detect that it was the user editing it and not the emulator process?
- Disruptive default hotkeys
  - RetroArch has default hotkeys that can disrupt gameplay unexpectedly. Example: while playing R-Type on PC Engine, pressing an unknown combo triggered "Switching to 6 button layout" and made the game unplayable. Need to investigate which core-specific hotkeys are enabled by default and consider disabling them
  - There may be others that have it
  - Drawback: users may be used to those hotkeys already?
- Support an option to pause sync during gameplay
  - Problems
    - Hard to detect when a game is running unless we constrain this to e.g. Steam games or other known platforms
    - Need to keep a process running or install a new systemd service

### Can launch without

- Eden/Cemu splash screens: emulator loading screens are ugly but can't be avoided
- ESDE Steam entry does not have an icon (the tiny square one)

## Features

### Proper blocker

### Would be really nice

- Extended directory support
  - Cheats directory layout: decide between per-emulator (`~/Emulation/cheats/{emulator}/`) or per-system (`~/Emulation/cheats/{system}/`). Some emulators support configurable cheat paths (Flycast, PCSX2)
  - DLC, patches and updates directory layout: similar to cheats, figure out folder structure for user-provided DLC and game updates. This could help solve the provision problem where some files must be imported via emulator UI (e.g., Cemu keys.txt). If kyaraben manages these directories, we could check for installed content. Wii U title structure: 00050000 (games), 0005000c (DLC), 0005000e (patches) per WiiUBrew
    - E.g. Eden 0.2.0 (upcoming) supports directories for DLCs and updates
  - Additional emulator paths: configure more optional directories. Flycast (BoxartPath, MappingsPath, TexturePath), PCSX2 (Cheats, Covers, Videos)
- Improve how paths are accessed within the UI
  - One click/tap to open ROMs, saves, etc.?
- Cross-emulator presets: toggle high-level features that cascade to all compatible emulators (widescreen, integer scaling, Discord presence, RetroAchievements, auto-save on exit)
  - With these, emulator feature parity is important to explore first
  - Categories
    - Autosave on exit
    - RetroAchievements integration: global credential storage with per-emulator login. Supported by DuckStation, PCSX2, PPSSPP, RetroArch cores, and Dolphin (experimental)
  - Performance presets (e.g. Steam Deck vs. Beefy Desktop): ship sensible defaults for renderer (Vulkan), resolution scale, recompilers, fast boot. Currently kyaraben focuses on paths but users must manually configure performance settings
- Storage breakdown bar: below the "Emulation folder" input in the UI, show a color-coded bar indicating total size and composition of the directory (ROMs, saves, opaque dirs, etc.)
- Handheld distro sync targets: support SD-based handhelds (Trimui Brick, Miyoo, etc.) running community distros (NextUI, PakUI, MinUI). These distros already bundle emulators, so kyaraben would skip emulator installation and instead generate configs and directory layouts matching the target distro's conventions, then sync ROMs, saves, and configs to the SD card via Syncthing or export. Add a "target" concept to config (e.g. `target = "trimui-brick-nextui"`) so users declare their systems once and kyaraben translates to the right structure per platform
- Display something in the UI when there are Syncthing conflicts
  - We don't need full-blown conflict resolution
  - But if we can show something simple and actionable that would be great

### Can launch without

- Improved security by adding device pairing confirmations
  - 6 characters could be low entropy
  - I think it's fine but adding a confirmation step can paliate this
- "Pause sync" feature that lives for the session
  - Sync continues on e.g. reboot
  - Can probably use systemd (stop the service? it'll restart automatically right?)
- Add a "Disable all systems" button for convenience
- How do we make the user store dir easier to navigate?
  - Can we integrate with file managers to provide icons for each folder?
  - Is it confusing that some things are per-system and others are per-emulator?
  - Is it a good idea to set up symlinks for convenience or is this more confusing?
  - Is it a good idea to add text for users such as README.md files? Or any other more popular format?
  - What else can be improved?
  - Alternative: just surface the path buttons better on the UI
- Compression guidance: tools or documentation for ROM compression (CHD for disc-based, RVZ for GameCube/Wii, ZIP for cartridge-based)
