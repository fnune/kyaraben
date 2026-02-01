# Emulator Installation Strategy

This document describes how Kyaraben chooses installation methods for emulators and the specific decisions made for each supported emulator.

## Decision Algorithm

When adding support for an emulator, follow this priority order:

### 1. Check for Official Versioned Binary Releases

The release MUST support versioned downloads (e.g., `/releases/download/v1.2.3/file.AppImage`). Rolling releases without version tags (e.g., "latest" only) cannot be used because Kyaraben needs reproducible installations.

**Preferred formats (in order):**
1. **AppImage** - Self-contained, uses host graphics drivers, no extraction needed
2. **Tarball (.tar.gz, .tar.xz, .7z)** - Extract and install binary
3. **Plain binary** - Direct download

**Rejected formats:**
- Flatpak - Conflicts with Kyaraben's management model (Flatpak manages its own updates, sandboxing, and paths)
- Snap - Same issues as Flatpak
- Distribution packages (.deb, .rpm) - Requires package manager, not portable

### 2. If No Official Binary, Check for Well-Maintained Unofficial Builds

Only consider unofficial builds if:
- Actively maintained (recent releases within last 3 months)
- Has multiple contributors or established maintainer
- Builds from official source code

### 3. For RetroArch Cores

Use nix-portable to fetch libretro cores from nixpkgs. Cores are shared libraries (`.so` files) without GUI code, so they don't have the graphics driver issues that affect GUI applications.

The RetroArch frontend itself should be installed via AppImage.

### 4. If No Viable Option Exists

Either:
- Skip the system/emulator for now
- Document the blocker and revisit when situation changes

## Why Not nix-portable for GUI Apps?

nix-portable works well for CLI tools and libraries, but GUI applications fail because:

1. **Graphics driver mismatch**: Nix-built apps expect libraries at NixOS-specific paths, but non-NixOS systems have drivers elsewhere
2. **EGL/Vulkan initialization fails**: Applications can't find or initialize the graphics context
3. **Workarounds are fragile**: Tools like nixGL exist but require wrapping every binary

AppImages avoid this because they use the host system's graphics drivers directly.

## Emulator Decisions

### Currently Implemented

| Emulator | System(s) | Method | Source | Notes |
|----------|-----------|--------|--------|-------|
| DuckStation | PSX | VersionedAppImage | [GitHub Releases](https://github.com/stenzek/duckstation/releases) | Flatpak deprecated by upstream |
| Eden | Switch | VersionedAppImage | [GitHub Releases](https://github.com/eden-emulator/Releases/releases) | Multiple hardware-optimized builds |
| PCSX2 | PS2 | VersionedAppImage | [GitHub Releases](https://github.com/PCSX2/pcsx2/releases) | Default for PS2 |
| PPSSPP | PSP | VersionedAppImage | [GitHub Releases](https://github.com/hrydgard/ppsspp/releases) | Default for PSP, replaced RA core |
| mGBA | GBA | VersionedAppImage | [GitHub Releases](https://github.com/mgba-emu/mgba/releases) | Default for GBA, replaced RA core |
| Dolphin | GC/Wii | VersionedAppImage | [pkgforge-dev](https://github.com/pkgforge-dev/Dolphin-emu-AppImage/releases) | Unofficial but well-maintained |
| Cemu | Wii U | VersionedAppImage | [GitHub Releases](https://github.com/cemu-project/Cemu/releases) | Official AppImage since 2.0 |
| Azahar | 3DS | VersionedAppImage | [GitHub Releases](https://github.com/azahar-emu/azahar/releases) | Citra successor |
| RetroArch:bsnes | SNES | VersionedAppImage (7z) | [Buildbot](https://buildbot.libretro.com/stable/) | Shared package with melonDS |
| RetroArch:melonds | NDS | VersionedAppImage (7z) | [Buildbot](https://buildbot.libretro.com/stable/) | Shared package with bsnes |
| TIC-80 | TIC-80 | Nixpkgs | nixpkgs | **BROKEN** - falls back to CLI |

### Planned Changes

| Emulator | System(s) | Method | Source | Versioned? | Notes |
|----------|-----------|--------|--------|------------|-------|
| RPCS3 | PS3 | AppImage | [GitHub Releases](https://github.com/RPCS3/rpcs3-binaries-linux/releases) | ✅ Yes | Rolling releases with build numbers |
| melonDS | NDS | AppImage (in ZIP) | [GitHub Releases](https://github.com/melonDS-emu/melonDS/releases) | ✅ Yes | Standalone, replace RA core as default |
| Vita3K | PS Vita | AppImage | [GitHub Releases](https://github.com/Vita3K/Vita3K-builds/releases) | ⚠️ Continuous | Rolling continuous builds |
| TIC-80 | TIC-80 | Tarball | [GitHub Releases](https://github.com/nesbox/TIC-80/releases) | ✅ Yes | Replace nix version |
| Flycast | Dreamcast | Tarball | [GitHub Releases](https://github.com/flyinghead/flycast/releases) | ✅ Yes | Or use RA core |

### RetroArch Approach

For systems best served by RetroArch cores (SNES, NES, Genesis, Saturn, N64):

1. **Download**: Fetch `RetroArch.7z` from buildbot (contains AppImage + cores)
2. **Extract**: Get both the AppImage frontend and bundled cores
3. **Install**: Place cores in managed directory, configure RetroArch to use them

Using bundled cores ensures compatibility between frontend and cores. The buildbot archive includes a curated set of cores that work with that RetroArch version.

### Systems Without Good Options

| System | Issue | Status |
|--------|-------|--------|
| (none currently) | | |

## Configuration Compatibility

All planned emulators support text-based configuration files:

| Emulator | Config Format | Key Paths Configurable |
|----------|---------------|----------------------|
| DuckStation | INI | BIOS, saves, states, memory cards |
| PCSX2 | INI | BIOS, saves, states, memory cards |
| RPCS3 | YAML | Firmware, games, per-game configs |
| PPSSPP | INI | All paths relative to memstick |
| melonDS | INI | BIOS, firmware, saves |
| mGBA | INI | Saves, states, screenshots |
| Dolphin | INI | All paths |
| Cemu | XML | All paths |
| Azahar | INI (Qt) | All paths |
| RetroArch | CFG | `system_directory`, `savefile_directory`, `libretro_directory` |

## URL Patterns for Versioned Downloads

### Verified Patterns (Ready for Implementation)

| Emulator | URL Pattern | Example Version |
|----------|-------------|-----------------|
| **DuckStation** | `https://github.com/stenzek/duckstation/releases/download/{version}/DuckStation-{target}.AppImage` | `v0.1-10655` |
| **Eden** | `https://github.com/eden-emulator/Releases/releases/download/{version}/Eden-Linux-{version}-{target}-clang-pgo.AppImage` | `v0.1.0` |
| **PCSX2** | `https://github.com/PCSX2/pcsx2/releases/download/{version}/pcsx2-{version}-linux-appimage-x64-Qt.AppImage` | `v2.6.3` |
| **PPSSPP** | `https://github.com/hrydgard/ppsspp/releases/download/{version}/PPSSPP-{version}-anylinux-x86_64.AppImage` | `v1.19.3` |
| **mGBA** | `https://github.com/mgba-emu/mgba/releases/download/{version}/mGBA-{version}-appimage-x64.appimage` | `0.10.5` (no 'v' prefix) |
| **Cemu** | `https://github.com/cemu-project/Cemu/releases/download/v{version}/Cemu-{version}-x86_64.AppImage` | `2.4` |
| **Azahar** | `https://github.com/azahar-emu/azahar/releases/download/{version}/azahar.AppImage` | `2120.1` |
| **RPCS3** | `https://github.com/RPCS3/rpcs3-binaries-linux/releases/download/build-{hash}/rpcs3-v0.0.{minor}-{build}-{hash}_linux64.AppImage` | Rolling |
| **RetroArch** | `https://buildbot.libretro.com/stable/{version}/linux/x86_64/RetroArch.7z` | `1.19.1` |
| **TIC-80** | `https://github.com/nesbox/TIC-80/releases/download/v{version}/tic80-{version}-linux.tar.gz` | `1.1.2837` |

### Patterns Requiring Special Handling

| Emulator | Issue | Workaround |
|----------|-------|------------|
| **melonDS** | AppImage inside ZIP | Download ZIP, extract AppImage |
| **Dolphin (unofficial)** | Tag format includes date/build ID | Use tag like `2512@2026-01-26_1769467304` |
| **Vita3K** | Continuous builds without version tags | Use latest continuous release |
| **RPCS3** | Rolling releases with commit hashes | Track by build number, accept hash changes |

### Patterns That DON'T Work

```
# Rolling "latest" without version in URL
https://example.com/download/latest/app.AppImage

# Version only in redirect, not URL
https://buildbot.example.com/nightly/app.AppImage
```

## Implementation Status

### Ready (AppImage, direct install)
- ✅ DuckStation - implemented
- ✅ Eden - implemented
- ✅ PCSX2 - implemented (PS2 default)
- ✅ PPSSPP - implemented (PSP default, replaces RA core)
- ✅ mGBA - implemented (GBA default, replaces RA core)
- ✅ Cemu - implemented (Wii U default)
- ✅ Azahar - implemented (3DS default)
- ✅ Dolphin - implemented (GameCube/Wii default, uses `release_tag` field)

### Archive Extraction Support
Archive extraction is implemented in flake.go (7z, tar.gz, zip). These emulators use archives instead of direct AppImages:

- ✅ RetroArch - 7z archive from buildbot (shared by bsnes, melonDS cores)
- 🔲 TIC-80 - tar.gz from GitHub (needs emulator definition update)

### Needs Special Handling
- 🔲 melonDS - AppImage inside ZIP file (replace RA core as NDS default)
- 🔲 Vita3K - continuous builds without version tags
- 🔲 RPCS3 - rolling releases with commit hashes in filenames

### All Hashes Are Placeholders
All new entries in `versions.toml` have placeholder hashes. To compute real hashes:
```bash
nix-prefetch-url <url>
nix hash convert --to sri --hash-algo sha256 <hash>
```

## Adding a New Emulator

1. Research official installation recommendations
2. Check if versioned binary releases exist
3. Verify URL pattern supports version substitution
4. Add entry to `internal/versions/versions.toml`
5. Add field to `internal/versions/versions.go` Versions struct
6. Add case to `internal/nix/flake.go` getAppImageVersion function
7. Create emulator definition in `internal/emulators/{name}/`
8. Add to registry in `internal/registry/all.go`
9. Implement `ConfigGenerator` for the emulator
10. Compute SHA256 hash and update versions.toml
11. Test installation and configuration
