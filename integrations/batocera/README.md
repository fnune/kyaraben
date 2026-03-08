# Kyaraben Batocera integration

This integration enables Kyaraben sync on Batocera, KNULLI, and Koriki.

Status: planned (see `feat/kyaraben-guest` branch)

## What syncs

| Category | Syncs | Notes |
|----------|-------|-------|
| ROMs | Yes | Per-system directories |
| Saves | Yes | Per-system directories |
| States | No | Batocera stores states alongside saves; syncing both would conflict |
| BIOS | No | Batocera uses a flat BIOS directory |
| Screenshots | Yes | Maps to `kyaraben-screenshots-retroarch` |

## Path mappings

| Kyaraben folder | Batocera path |
|-----------------|---------------|
| `kyaraben-roms-{system}` | `/userdata/roms/{system}/` |
| `kyaraben-saves-{system}` | `/userdata/saves/{system}/` |
| `kyaraben-screenshots-retroarch` | `/userdata/screenshots/` |

## System name differences

Kyaraben uses `genesis` while Batocera uses `megadrive`. The integration handles this translation automatically.

## Limitations

### No BIOS sync

Batocera stores all BIOS files in a single flat `/userdata/bios/` directory. Kyaraben expects per-system BIOS folders. Since Syncthing cannot map multiple folder IDs to the same path with different filters, BIOS syncing is not supported.

### No save state sync

Batocera stores both save files (`.srm`) and save states (`.state*`) in the same `/userdata/saves/{system}/` directory. Since we cannot sync the same path with different filters, we sync save files only. Save states are not synced.

This is arguably a feature: save states are often incompatible between different emulator versions, so syncing them between devices can cause issues.
