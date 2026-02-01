# Kyaraben

Kyaraben is a declarative emulation manager for Linux. It handles the installation and configuration of emulators for various gaming systems.

This project is in development.

<p align="center">
  <img src="assets/kyaraben.png" alt="Kyaraben screenshot" width="800">
</p>

## How it works

1. Select the systems you want to emulate
2. Kyaraben shows which BIOS or firmware files are required for each system
3. Click Apply to install the emulators and configure them

Kyaraben uses Nix to install emulators, which means installations are reproducible and isolated from the rest of your system. You do not need to have Nix installed; Kyaraben bundles a portable Nix distribution.

## Supported systems

- Nintendo: SNES, Game Boy Advance, GameCube, DS, Wii, 3DS, Wii U, Switch
- Sony: PlayStation, PlayStation 2, PSP, PlayStation 3, PS Vita
- Sega: Dreamcast

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for technical details and conventions.

---

<sub>System logos from [ES-DE](https://es-de.org)</sub>
