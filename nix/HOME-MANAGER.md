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

None of these options are satisfactory. Revisit when there's clearer demand.
