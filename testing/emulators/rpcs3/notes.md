# RPCS3 (PS3 emulator)

## Status: investigating

## Config location

RPCS3 uses `~/.config/rpcs3/` for config and `vfs.yml` for path mappings.

## Key findings

### Firmware requires UI import

Placing PS3UPDAT.PUP in the provision directory is not enough. RPCS3 requires importing via File > Install Firmware. After import, firmware extracts to `dev_flash/` in the VFS.

Kyaraben shows `✓ Firmware (Official firmware) Verified (PS3UPDAT.PUP)` but RPCS3 still needs the UI import step.

### VFS directories created on firmware install

The dev_flash directories don't exist until firmware is installed. RPCS3 creates them during firmware import:
- `dev_flash/` - main firmware files
- `dev_flash2/` - additional system files
- `dev_flash3/` - more system files

### Double slash in game paths

RPCS3 adds double slashes when storing game paths in games.yml:
```
BLES00229: /home/fausto/Emulation/roms/ps3//Grand Theft Auto IV.iso
```

This happens regardless of whether vfs.yml `/games/` has a trailing slash. Appears to be RPCS3 internal behavior. May be cosmetic - RPCS3 still finds and reads the games.

### Game encryption

PS3 disc ISOs are encrypted. RPCS3 can read the PS3_GAME structure but fails on EBOOT.BIN:
```
E SYS: Invalid or unsupported file format: .../PS3_GAME/USRDIR/EBOOT.BIN
```

Games need to be decrypted before RPCS3 can play them. This is expected - not a kyaraben issue.

Decryption tools (Windows-only):
- PS3 Quick Disc Decryptor: https://github.com/ElektroStudios/PS3-Quick-Disc-Decryptor
- PS3Dec (older tool)

No native Linux decryption tool available as of 2026-02.

## ROM formats

RPCS3 cannot play directly from ZIP files. Must extract.

Supported:
- Decrypted ISO files
- Folder format (extracted/decrypted games with EBOOT.BIN)
- pkg files (PSN games, may need rap files for licensing)

## Current kyaraben config

Uses opaque dir model with VFS configuration:
- `$(EmulatorDir)` - opaque directory base
- `/dev_hdd0/` - main storage (saves, game data, DLC)
- `/dev_flash/` - firmware location (created on firmware install)
- `/games/` - configured to ROMs dir (no trailing slash)

## Symlinks-over-opaque assessment

RPCS3 stores everything interconnected in dev_hdd0:
- `dev_hdd0/home/00000001/savedata/` - save data
- `dev_hdd0/game/` - installed games, DLC, patches
- `dev_hdd0/tmp/` - temporary files

Save data location includes user ID subdirectory. May be complex to symlink cleanly.

## TODO

- [ ] Test with decrypted game to verify full functionality
- [ ] Investigate if saves can be symlinked (dev_hdd0/home/00000001/savedata/)
- [ ] Add firmware hash verification
- [ ] Consider keeping opaque dir model due to complex VFS structure
