# Home-manager module for kyaraben
#
# STATUS: Deferred - not yet implemented
#
# This module is a placeholder. A proper home-manager integration is planned
# but requires design work to resolve the impedance mismatch between kyaraben's
# approach and home-manager's philosophy.
#
# ## The Problem
#
# Kyaraben's CLI uses **merge semantics** for emulator config files:
# - Reads existing config (if any)
# - Patches in kyaraben-managed values
# - Preserves user's manual customizations
#
# Home-manager uses **declarative/overwrite semantics**:
# - Config files are generated from Nix expressions
# - On `home-manager switch`, files are replaced entirely
# - User customizations go in the Nix config, not the files
#
# ## Options Considered
#
# 1. **Call CLI from activation script**: Works but feels non-native to Nix.
#    Home-manager shelling out to a Go binary defeats the purpose.
#
# 2. **Duplicate config logic in Nix**: Maintains two implementations (Go + Nix)
#    that must stay in sync. Error-prone.
#
# 3. **Accept different semantics**: Document that home-manager users get
#    fully declarative configs while CLI users get merge behavior. But this
#    means the tools behave differently, which is confusing.
#
# 4. **Generate partial configs + include**: Some emulators support config
#    includes (RetroArch's #include). Could write kyaraben values to a
#    separate file. But not all emulators support this.
#
# ## Current Recommendation
#
# For NixOS/home-manager users who want kyaraben functionality:
#
# 1. Install emulators directly via home.packages or environment.systemPackages
# 2. Use home-manager's xdg.configFile to manage emulator configs declaratively
# 3. Create the directory structure manually or via home.activation
#
# Example (not using kyaraben):
#
#   home.packages = with pkgs; [
#     (retroarch.override { cores = with libretro; [ bsnes ]; })
#     duckstation
#   ];
#
#   xdg.configFile."retroarch/retroarch.cfg".text = ''
#     system_directory = ~/Emulation/bios
#     savefile_directory = ~/Emulation/saves/snes
#     # ... etc
#   '';
#
# This gives you the same end result with native Nix tooling.
#
# ## Future Plans
#
# We may revisit this once we better understand:
# - What home-manager users actually want from kyaraben
# - Whether there's a clean way to share config generation logic
# - If the merge vs declarative distinction matters in practice
#
# Contributions and design discussions welcome.

{ ... }:

{
  # This module intentionally does nothing.
  # See comments above for rationale.
}
