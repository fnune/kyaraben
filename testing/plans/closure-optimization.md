# Closure optimization

## Status: implementation in progress

## Implementation plan

### Phase 1: Download cores from buildbot (DONE)

Changed from nixpkgs libretro cores to downloading `RetroArch_cores.7z` from buildbot:
- URL: `https://buildbot.libretro.com/stable/{version}/linux/{arch}/RetroArch_cores.7z`
- Same source as RetroArch AppImage (version 1.22.2)
- Extract only the cores we need (bsnes, genesis-plus-gx, mesen, mupen64plus, beetle-saturn)
- Eliminates second nixpkgs input entirely

Files modified:
- `internal/nix/flake.go`: Removed retroarch-cores nixpkgs input, added cores extraction derivation
- `internal/versions/versions.toml`: Added RetroArch_cores.7z URL, hash, and core file mappings
- `internal/versions/versions.go`: Updated RetroArchCoresSpec struct, added GetCoresURL method

Expected savings: ~500-600 MB (eliminates duplicate glibc, gcc-lib, systemd, ffmpeg, retroarch-bare, flite, freepats, wildmidi, one nixpkgs source tarball)

Note: Only x86_64 hash added for now. Need to add aarch64 hash (requires fetching from buildbot).

### Phase 2: Investigate Qt/GTK deps (RESOLVED)

Qt6 (270 MB) and GTK (70 MB) were dependencies of the nixpkgs libretro cores (retroarch-bare).
Eliminated by phase 1 - no longer in closure.

### Phase 3: Investigate gcc/python deps (RESOLVED)

Full gcc (249 MB) and python (113 MB) were build dependencies of nixpkgs libretro cores.
Eliminated by phase 1 - only small gcc-lib and hook scripts remain.

### Phase 4: Avoid source tarball bloat (TODO)

Two nixpkgs source tarballs remain (572 MB total):
- These are cached for flake evaluation
- Could run `nix-collect-garbage` after each build
- Or accept as cost of nix-portable approach

Potential savings: ~570 MB if GC'd, but would slow down subsequent evaluations.

## The problem

Kyaraben's nix closure is larger than necessary, leading to:

1. Slow first installs (hundreds of packages to build/download)
2. Confusing progress indication (user sees 400+ packages when they only requested 3 emulators)
3. Wasted disk space in the nix store

During a typical install, nix outputs multiple waves:
- ~28 packages for evaluation/git repos
- ~400 packages for dependencies (build tools, libraries)
- ~47 packages for the actual emulators/frontends

The user only cares about the last wave.

## Investigation areas

### 1. Ensure RetroArch cores come from cache

libretro-genesis-plus-gx (and possibly other cores) builds from source instead of being fetched from the binary cache. This is slow and unnecessary.

Questions to investigate:
- Which cores are building locally vs fetching from cache?
- Is this a nixpkgs issue (package not in hydra)?
- Can we pin to a nixpkgs commit that has cached binaries?
- Should we report upstream if packages are misconfigured?

Finding: Hydra shows genesis-plus-gx builds from 2023 with note "not a member of the latest evaluation of its jobset". The pinned retroarch-cores commit likely has no cached binaries. Need to find a commit that was actually built by Hydra.

### 2. Reduce the flake's input set

The flake may be fetching inputs we don't need. Investigate:
- What git repos are fetched during evaluation?
- Can we minimize `nix flake update` scope?
- Are there unused inputs in flake.nix?

### 3. Consider prebuilt binaries

Currently kyaraben uses nixpkgs packages. Alternative approaches:
- Use AppImages directly (already available for most emulators)
- Create a custom binary cache with prebuilt kyaraben packages
- Hybrid: use nixpkgs for cores, AppImages for standalone emulators

Trade-offs:
- AppImages: faster downloads, but lose nix's reproducibility
- Custom cache: faster for users, but maintenance burden for us
- Current approach: consistent, but slow first install

### 4. Lazy evaluation

Currently we build all enabled emulators in one derivation. Could we:
- Build each emulator separately and combine at the end?
- Use nix's lazy evaluation to avoid building unused dependencies?
- Profile the evaluation phase to find bottlenecks?

### 5. Measure and track

Before optimizing, we should measure:
- Total closure size per emulator combination
- Time to install from scratch
- Time to install with cached store
- Number of packages built vs fetched

Add CI checks to track closure size regressions.

## Measurements

### Before optimization (2026-02-10)

Full install with all emulators + frontends enabled:

| Metric | Value |
|--------|-------|
| Total nix store size | 4.3 GB |
| Packages in store | 1108 |

### After phase 1 (buildbot cores)

| Metric | Value | Change |
|--------|-------|--------|
| Total nix store size | 3.0 GB | -1.3 GB (30%) |
| Packages in store | 425 | -683 packages |

What was eliminated:
- Qt6 (qtbase, qtdeclarative): ~270 MB gone
- GTK3/4: ~70 MB gone
- retroarch-bare: 19 MB gone
- flite, freepats, wildmidi: ~100 MB gone
- Duplicate systemd, ffmpeg: ~100 MB gone
- One nixpkgs source tarball: ~300 MB gone
- genesis-plus-gx no longer builds from source

What remains:
- 3x glibc (90 MB): probably unavoidable with nix-portable
- 2x source tarballs (572 MB): could be GC'd after build
- git-minimal (48 MB): needed for flake fetching
- kyaraben-retroarch-cores: 16 MB (just the .so files we need)

### Locally built packages

The 47 locally built packages break down as:

1. AppImage downloads (expected): cemu, melonds, retroarch, ppsspp, azahar, flycast, dolphin, mgba, eden, rpcs3, duckstation, vita3k, pcsx2, es-de
2. Icon fetches (expected): ~14 icon downloads
3. Profile/symlinkJoin derivations (expected): kyaraben-profile, kyaraben-icons, kyaraben-retroarch-cores
4. RetroArch cores (problem): libretro-genesis-plus-gx builds from source

### Core sizes

| Core | Size |
|------|------|
| libretro-genesis-plus-gx | 13 MB |
| libretro-mupen64plus-next | 8.4 MB |
| libretro-mednafen-saturn | 8.0 MB |
| libretro-bsnes | 6.1 MB |
| libretro-mesen | 3.7 MB |
| retroarch-bare (from nixpkgs) | 19 MB |
| retroarch (AppImage) | 10 MB |

Note: `retroarch-bare-1.21.0` appears in the store as a dependency of the cores even though we use the AppImage. This is wasteful.

### Two nixpkgs issue

The flake uses two nixpkgs inputs:
- `nixpkgs` at `fa83fd837f...` for main packages
- `retroarch-cores` at `d03088749a...` for libretro cores

This causes redundant dependencies since both nixpkgs versions pull in their own glibc, gcc-lib, etc. The wildmidi dependency observed during build comes from the cores input.

### Full closure breakdown (4.3GB)

| Category | Size | Notes |
|----------|------|-------|
| Emulator AppImages | ~820 MB | Unavoidable - this is what we want |
| Libretro cores | ~40 MB | Unavoidable |
| Nixpkgs source tarballs | ~900 MB | 3 copies (two inputs + flake-utils) |
| Qt (qtbase + qtdeclarative) | ~270 MB | Needed by some packages for extraction? |
| gcc | 249 MB | Build tool - shouldn't be in runtime closure |
| glibc (3 versions) | 90 MB | 2.37, 2.39, 2.40 from different inputs |
| python3 | 113 MB | Build dependency |
| systemd (2 versions) | ~80 MB | From different inputs |
| ffmpeg (2 versions) | ~60 MB | From different inputs |
| gtk3 + gtk4 | ~70 MB | Needed by some packages |
| flite (text-to-speech) | 61 MB | Dependency of retroarch-bare |
| retroarch-bare | 19 MB | Waste - we use AppImage instead |
| freepats (MIDI samples) | 33 MB | Dependency of wildmidi for cores |
| git-minimal | 48 MB | Needed for flake fetching |
| nix | 18 MB | Needed for nix-portable |
| Other deps | ~350 MB | icu4c, binutils, pipewire, etc |

### What's actually unavoidable

~900 MB: the emulator binaries themselves

### What could be eliminated

If we download cores outside nix (option 1 from above):
- Eliminates second nixpkgs input entirely
- Saves: ~300MB source tarball, 30MB glibc, retroarch-bare 19MB, freepats 33MB, flite 61MB, wildmidi, and other core deps
- Estimated savings: ~500-600 MB

If nixpkgs source tarballs could be garbage collected after build:
- Saves: ~600 MB (keeping one input's sources)

If build tools (gcc, python, binutils) aren't kept in runtime closure:
- Saves: ~400 MB
- This may require understanding why they're being retained

## Concrete next steps

1. ~~Run `nix path-info -rsSh` on a typical kyaraben build to measure closure size~~ Done
2. ~~Identify which packages are building locally vs fetching from cache~~ Done
3. Check if libretro cores have hydra builds at our pinned nixpkgs commit
4. Test: pin nixpkgs to a known-good commit with all packages cached
5. Document findings and decide on approach
6. Investigate: download prebuilt core .so files directly instead of using nixpkgs

## Low-hanging fruit: download cores from libretro buildbot

The libretro buildbot provides prebuilt .so files at `https://buildbot.libretro.com/nightly/linux/x86_64/latest/`:

| Core | Compressed size | Current nixpkgs size |
|------|-----------------|---------------------|
| bsnes_libretro.so.zip | 992 KB | 6.1 MB |
| genesis_plus_gx_libretro.so.zip | 1282 KB | 13 MB |
| mesen_libretro.so.zip | 1069 KB | 3.7 MB |
| mupen64plus_next_libretro.so.zip | 2710 KB | 8.4 MB |
| mednafen_saturn_libretro.so.zip | 1529 KB | 8.0 MB |

Total: ~7.5 MB zipped downloads vs pulling in a separate nixpkgs input with all its dependencies (glibc, gcc-lib, wildmidi, etc.).

Benefits:
- Eliminates the second nixpkgs input entirely
- No more building genesis-plus-gx from source
- Downloads are tiny and fast
- Cores update more frequently than nixpkgs packages
- We already do this pattern for RetroArch itself (AppImage from buildbot)

Trade-offs:
- Lose nix's reproducibility guarantees for cores
- Need to track core versions/hashes ourselves
- Nightly builds may be less stable than releases (but the same cores have been stable for years)
- No aarch64 builds available (only x86_64, armv7, armhf) - would need to keep nixpkgs fallback for aarch64

Blocker: nightly URLs are not versioned (`/latest/` always points to current). We can't pin hashes for nix fetchurl. The stable releases bundle all cores in one 274 MB archive, defeating the purpose.

GitHub releases for individual cores (e.g. libretro/bsnes-libretro) also use a rolling "nightly" tag that updates in place. Asset IDs change when updated. No way to pin.

Remaining options:
1. Download cores outside nix entirely (in Go code during apply)
2. Host our own mirror with versioned artifacts
3. Accept the nixpkgs overhead but find a Hydra-cached commit
4. Build cores at kyaraben release time and host the .so files

Option 1 is the cleanest:
- Download .so.zip from buildbot in Go code (like we do for AppImages, but simpler)
- Extract to cores directory directly
- Skip nix entirely for cores
- Pin versions/hashes in versions.toml like we do for emulators
- Nightly URLs are acceptable if we check hash at download time and fail loudly if changed
- Can update hashes via automation (bot that checks buildbot daily)

## Related feedback items

- feedback.md mentions "libretro-genesis-plus-gx builds from source on apply, which is slow"
- Progress indication now filters to only show known emulator names to hide the dependency noise
