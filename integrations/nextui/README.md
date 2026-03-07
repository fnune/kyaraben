# Kyaraben.pak

A NextUI tool pak for syncing ROMs, saves, and BIOS files with [Kyaraben](https://github.com/fnune/kyaraben) devices using Syncthing.

## How it works

This pak wraps Syncthing and adds:
- Pairing with 6-character codes via the Kyaraben relay
- Folder mapping between Kyaraben and NextUI structures
- Symlinks so NextUI sees content at expected paths

Syncthing syncs to a hidden `/.kyaraben/` directory, and symlinks expose content where NextUI expects it:

| Syncs to | Symlinked from |
|----------|----------------|
| `/.kyaraben/roms/gb/` | `/Roms/Game Boy (GB)/` |
| `/.kyaraben/saves/gb/` | `/Saves/GB/` |
| `/.kyaraben/bios/gb/` | `/Bios/GB/` |

## Supported platforms

- `tg5040`: TrimUI Brick, TrimUI Smart Pro
- `miyoomini`: Miyoo Mini Plus
- `my282`: Miyoo A30
- `my355`: Miyoo Flip
- `rg35xxplus`: RG-35XX Plus, RG-34XX, RG-35XX H, RG-35XX SP

## Installation

1. Download the latest `Kyaraben.pak.zip` release
2. Copy to `/Tools/$PLATFORM/` on your SD card (e.g., `/Tools/tg5040/` for TrimUI Brick)
3. Extract in place, then delete the zip
4. Confirm `/Tools/$PLATFORM/Kyaraben.pak/launch.sh` exists

## Usage

1. Open Tools > Kyaraben on your NextUI device
2. Select "Pair with device"
3. Enter the 6-character code shown on your other Kyaraben device
4. Select which systems to sync
5. Enable "Start on boot" for automatic syncing

## Building

From the kyaraben repo root:

```sh
just nextui-build
```

This fetches binaries from upstream and assembles the pak.

To create a release zip:

```sh
just nextui-release
```

## Upstream

This pak is based on [josegonzalez/minui-syncthing-pak](https://github.com/josegonzalez/minui-syncthing-pak), included as a git submodule at `upstream/`.

To update upstream:

```sh
git submodule update --remote integrations/nextui/upstream
just nextui-build
```

## Debug logging

Logs are written to `/.userdata/$PLATFORM/logs/Kyaraben.txt`.

## License

MIT (inherited from upstream)
