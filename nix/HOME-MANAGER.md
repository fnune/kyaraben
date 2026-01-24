# Home-manager Integration

A home-manager module for kyaraben is not yet implemented.

## The Problem

Kyaraben's CLI uses **merge semantics** for emulator config files:
- Reads existing config (if any)
- Patches in kyaraben-managed values
- Preserves user's manual customizations

Home-manager uses **declarative/overwrite semantics**:
- Config files are generated from Nix expressions
- On `home-manager switch`, files are replaced entirely
- User customizations go in the Nix config, not the files

## Options Considered

1. **Call CLI from activation script**: Works but feels non-native to Nix.
2. **Duplicate config logic in Nix**: Two implementations that must stay in sync.
3. **Accept different semantics**: Confusing that the tools behave differently.
4. **Partial configs with includes**: Not all emulators support config includes.

## Workaround for NixOS/Home-manager Users

You can achieve similar results using native Nix tooling:

```nix
{ pkgs, ... }:

{
  home.packages = with pkgs; [
    (retroarch.override { cores = with libretro; [ bsnes ]; })
    duckstation
  ];

  # Create directory structure
  home.activation.emulationDirs = lib.hm.dag.entryAfter [ "writeBoundary" ] ''
    mkdir -p ~/Emulation/{roms,bios,saves,states,screenshots}/{snes,psx}
  '';

  # RetroArch config
  xdg.configFile."retroarch/retroarch.cfg".text = ''
    system_directory = ~/Emulation/bios
    savefile_directory = ~/Emulation/saves/snes
    savestate_directory = ~/Emulation/states/snes
    screenshot_directory = ~/Emulation/screenshots/snes
    rgui_browser_directory = ~/Emulation/roms/snes
  '';

  # DuckStation config
  xdg.configFile."duckstation/settings.ini".text = ''
    [BIOS]
    SearchDirectory = ~/Emulation/bios/psx

    [MemoryCards]
    Directory = ~/Emulation/saves/psx

    [Folders]
    SaveStates = ~/Emulation/states/psx
    Screenshots = ~/Emulation/screenshots/psx

    [GameList]
    RecursivePaths = ~/Emulation/roms/psx
  '';
}
```

## Future Plans

We may revisit this once we better understand what home-manager users want
from kyaraben and whether there's a clean way to share config logic.

Contributions and design discussions welcome.
