## [0.1.5] - 2026-08-27

### Features

- *(sync)* Add per-device ROM deletion protection
- *(sync)* Surface per-device ROM deletion protection and add CLI
- *(savestate)* Split resume into autosave and autoload
- *(nextui)* Add ROM deletion protection to the integration
- *(versions)* Bump emulators and syncthing to latest versions

### Bug fixes

- *(justfile)* Insert blank line between releases in CHANGELOG
- *(shaders)* Disable the CRT shadow mask
- *(sync)* Apply versioning to saves and states folders
- *(sync)* Make UI stop resilient and surface failures
- *(ui)* Rename ROM deletion setting and shorten its help
- *(ui)* Show sync toggles in their real off state
- *(nextui)* Add versioning and the metadata folder to the device

## [0.1.4] - 2026-05-21

This release contains a round of emulator upgrades and a Syncthing minor-version bump,
all tested on my Steam Deck. It also suppresses Eden's recurring "check for updates?"
prompt on every launch.

| Package | Version Change |
|---------|----------------|
| azahar | 2125.0-alpha5 → 2125.1.2 |
| dolphin | 2512 → 2603a |
| duckstation | v0.1-10861 → v0.1-11108 |
| eden | v0.2.0-rc1 → v0.2.0 |
| esde | 3.4.0 → v3.4.1 |
| pcsx2 | v2.7.168 → v2.7.356 |
| ppsspp | v1.20.1 → v1.20.4 |
| syncthing | v2.0.15 → v2.1.0 |
| vita3k | 3935 → 4017 |
| xemu | 0.8.134 → v0.8.135 |
| xenia-edge | 2beb0bf → a71dc5c |

### Features

- *(versions)* Bump emulators and syncthing to latest versions
- *(eden)* Suppress update check prompt

### Bug fixes

- *(sync)* Stop relay polling when pairing is cancelled

### Refactor

- *(versions)* Replace hardcoded version assertions with structural invariants

## [0.1.3] - 2026-03-09

### Features

- *(versions)* Update emulator versions
- *(apply)* Continue on transient download failures

### Bug fixes

- *(ui)* Update documentation URLs to kyaraben.org
- *(versions)* Preserve TOML definition order for upgrade/downgrade detection
- *(duckstation)* Write config to both XDG locations
- *(diff)* Clarify config change display with labels and consistent direction
- *(e2e)* Handle duckstation writing to both XDG locations

## [0.1.2] - 2026-03-09

### Features

- *(versions)* Add Forgejo support for version checking
- *(ui)* Link to release notes in update banner

### Bug fixes

- *(ui)* Fix update download getting stuck on redirects
- *(updater)* Fix self-update relaunch and text file busy error

## [0.1.1] - 2026-03-09

### Features

- Use kyaraben.org as primary relay and documentation URL
# Changelog

## 0.1.0

Initial release.

### Features

- **System and emulator management**: select systems from Atari 2600 through PS3 and Switch, with automatic emulator installation as self-contained AppImages
- **Collection directory**: unified folder structure for ROMs, saves, states, and BIOS files
- **BIOS verification**: provisions panel shows required files, hash verification, and placement instructions
- **Configuration management**: Kyaraben manages specific emulator config keys and shows diffs before applying changes
- **Desktop integration**: creates desktop entries and integrates with ES-DE frontend

### Sync

- **Syncthing-based sync**: automatic Syncthing setup with systemd service management
- **Device pairing**: 6-digit relay-based pairing codes for easy device connection
- **Multi-device support**: sync between desktop, Steam Deck, and headless servers
- **NextUI guest integration**: sync with NextUI handhelds

### Platforms

- Linux x86_64 (AppImage)
- Steam Deck (works in Game Mode)
- Headless server (CLI-only for sync hub use case)
- ARM64 experimental

### CLI

- `kyaraben status` - show current state
- `kyaraben apply` - apply configuration changes
- `kyaraben sync` - manage sync and device pairing
- `kyaraben doctor` - check BIOS and firmware status
