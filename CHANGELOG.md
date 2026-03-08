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
