# Filesystem mapping exploration

How do we map emulator-native paths to a unified `UserStore` layout?

## The problem

Each emulator has its own expectations for where state lives:

- DuckStation: `~/.config/duckstation/memcards/`, `~/.config/duckstation/savestates/`
- RetroArch: `~/.config/retroarch/saves/`, `~/.config/retroarch/states/`
- Dolphin: `~/.local/share/dolphin-emu/GC/`, `~/.local/share/dolphin-emu/StateSaves/`

We want a unified `UserStore`:

```
~/Emulation/
├── saves/
│   ├── psx/
│   └── gamecube/
└── states/
    ├── psx/
    └── gamecube/
```

How do we bridge these?

## Options

### Option 1: configure emulators directly

Most emulators have config options to change paths:

- RetroArch: `savefile_directory`, `savestate_directory`
- DuckStation: settings.ini has path options
- Dolphin: `Dolphin.ini` has path options

Pros:

- Simplest approach
- No special filesystem tricks
- Works everywhere

Cons:

- Some emulators may not support full path customization
- Need to verify each emulator's capabilities

This is likely sufficient for MVP. Start here.

### Option 2: symlinks

Create symlinks from emulator-expected locations to `UserStore`:

```
~/.config/duckstation/memcards/ -> ~/Emulation/saves/psx/
```

Pros:

- Simple to implement
- No special permissions

Cons:

- Syncthing behavior with symlinks is confusing (follows by default, but can cause issues)
- User sees symlinks in their home directory
- Some emulators may not handle symlinks well
- Clutters XDG directories

### Option 3: FUSE filesystem

Create a FUSE mount that presents `UserStore` as whatever structure an emulator expects:

```
~/.local/share/kyaraben/mounts/duckstation/  (FUSE)
    memcards/ -> actually ~/Emulation/saves/psx/
```

Configure emulator to use the FUSE mount as its base.

Pros:

- Transparent to emulators
- Clean separation
- Can handle complex remapping

Cons:

- Requires FUSE (usually available but adds dependency)
- More complex to implement
- Mount lifecycle management
- Performance overhead (minimal but present)
- May have issues in Flatpak/AppImage sandbox

### Option 4: bind mounts

Use bind mounts to overlay paths:

```
mount --bind ~/Emulation/saves/psx ~/.config/duckstation/memcards
```

Pros:

- Transparent to emulators
- Native kernel support

Cons:

- Requires root (traditionally)
- User namespaces can do rootless bind mounts but adds complexity
- Mount lifecycle management
- Probably overkill

### Option 5: OverlayFS

Layer `UserStore` over emulator directories.

Cons:

- Requires root or user namespaces
- Overkill for this use case
- Complex semantics

## Recommendation

**Option 1 (configure emulators directly) is the only approach.**

Most modern emulators support path configuration. This is the only mechanism we use.

If an emulator can't be configured to use custom paths:

1. Don't support that emulator, or
2. File upstream feature request, or
3. Support it with documented limitation (state lives outside `UserStore`, won't sync)

No FUSE, no symlinks, no fallbacks. This keeps the architecture simple. Path configurability is a requirement for full `Emulator` support.

**Important constraint:** kyaraben manages emulator configuration completely. Using kyaraben's sync features without its configuration management is not supported. We need full control over emulator paths to ensure state lands in `UserStore` where the `Synchronizer` expects it.

## Opaque emulator directories

Some emulators have their own opinionated internal directory structure that resists granular path configuration. Examples:

- **Eden** (Switch): Uses a complex internal structure for NAND, keys, firmware, shader cache
- **Standalone PPSSPP**: Uses a "memstick" directory mimicking the PSP's memory stick structure

For these emulators, instead of trying to map individual paths (saves → `~/Emulation/saves/switch`), we configure the emulator's entire data directory to live within `UserStore` under an `opaque/` directory:

```
~/Emulation/
├── saves/           # kyaraben-structured (for emulators with granular path config)
├── states/          # kyaraben-structured
├── opaque/          # emulators that manage their own directory structure
│   └── eden/        # Eden manages this directory internally
│       └── (internal structure owned by emulator)
└── ...
```

The pattern:
1. The `opaque/` directory contains emulators that manage their own internal structure
2. Kyaraben syncs the entire `opaque/<emulator>/` directory without understanding its internals
3. The emulator is configured to use `opaque/<emulator>/` as its data root

This is a middle ground between full path control and no sync support. We lose the unified structure but gain sync capability for emulators that otherwise wouldn't fit the model.

**When to use opaque directories:**
- Emulator doesn't support individual path configuration for saves/states/etc.
- Emulator's internal structure is complex or mimics original hardware
- The emulator can at least configure its overall data directory location

**Trade-offs:**
- Sync includes everything (shader cache, etc.) not just saves - may need per-emulator sync config
- User's `UserStore` layout is less uniform
- Users need to look inside the emulator directory to find their saves

## Open questions

1. Are there emulators where path configuration is insufficient?
2. How do Syncthing users currently handle emulator saves across devices?
3. Does FUSE work reliably inside AppImage?
4. For opaque directories, should we support per-emulator sync configuration to exclude regenerable data like shader cache?

## Testing needed

Before committing to an approach:

- [ ] Verify RetroArch path config works for saves/states/screenshots
- [ ] Verify DuckStation path config works
- [ ] Test Syncthing sync with configured paths
- [ ] Document any emulators that can't be configured
