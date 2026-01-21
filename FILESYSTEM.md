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

## Open questions

1. Are there emulators where path configuration is insufficient?
2. How do Syncthing users currently handle emulator saves across devices?
3. Does FUSE work reliably inside AppImage?

## Testing needed

Before committing to an approach:

- [ ] Verify RetroArch path config works for saves/states/screenshots
- [ ] Verify DuckStation path config works
- [ ] Test Syncthing sync with configured paths
- [ ] Document any emulators that can't be configured
