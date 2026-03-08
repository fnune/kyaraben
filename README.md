<p align="center">
  <img src="assets/kyaraben-logo-with-bg.svg" alt="Kyaraben" width="400">
</p>

<p align="center">
  <a href="https://github.com/fnune/kyaraben/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License: MIT"></a>
</p>

Kyaraben is a declarative emulation manager for Linux. It handles the installation and configuration of emulators for various gaming systems.

<p align="center">
  <img src="assets/screenshots/catalog-idle.png" alt="Kyaraben screenshot" width="800">
</p>

## Installation

Download the latest AppImage from the [releases page](https://github.com/fnune/kyaraben/releases) and run it:

```bash
chmod +x Kyaraben-*.AppImage
./Kyaraben-*.AppImage
```

Works on most Linux distributions, including SteamOS on the Steam Deck.

## How it works

1. Select the systems you want to emulate
2. Click apply to install the emulators and configure them
3. Kyaraben shows which BIOS or firmware files are required for each system
4. Drop your ROMs into the created folders and play

## Requirements

- Linux (x86_64)
- systemd (for the sync feature; emulators work without it)

## Documentation

- [Getting started](https://kyaraben.dev/getting-started/)
- [App reference](https://kyaraben.dev/using-the-app/)
- [CLI reference](https://kyaraben.dev/using-the-cli/)
- [Synchronization](https://kyaraben.dev/sync/)

## Contributing

See the [contributing guide](site/src/content/docs/contributing.mdx) for development setup and conventions.

## License

MIT

---

<sub>System logos from [ES-DE](https://es-de.org)</sub>
