# Home-manager Integration

A home-manager module for kyaraben is not yet implemented.

## Design

For home-manager users, kyaraben provides **preconfigured good defaults** for
emulator setups. Config merging is intentionally NOT supported - home-manager
users expect fully declarative configs managed through their Nix configuration.

This means maintaining two implementations:
- Go (CLI/UI): merge semantics, preserves user customizations
- Nix (home-manager): declarative, configs come from Nix expressions

## Open Question

How do the Go and Nix implementations stay in sync? The default config values
(paths, settings) need to match between both implementations.

Options:
- Manual synchronization with tests that verify parity
- Generate Nix from Go (or vice versa)
- Shared config specification (JSON/TOML) that both read

Revisit when there's clearer demand for home-manager support.
