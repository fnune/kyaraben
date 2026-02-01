# Feedback

## DuckStation onboarding wizard

DuckStation runs an onboarding wizard on first launch that wants to create a config file and set up autoupdates. This is not good for kyaraben's managed experience. We need a better default config that prevents this wizard from appearing.

---

## Version tracking for emulators

It would be useful to have a programmatic, non-LLM way to check if any emulator supported by kyaraben has new versions available that we're not offering. This would:

1. Automatically query GitHub releases (or other sources) for each emulator
2. Compare against versions.toml to identify:
   - New versions available that we don't have
   - Old versions we're tracking that could be removed
3. Provide a quick way to bump versions and remove old versions

This should be a development script (e.g., `scripts/check-versions.go` or a make target) or a CI job that runs periodically and creates PRs when updates are available.

---

## Dolphin prompts for autoupdates on launch

Dolphin tries to enable its built-in autoupdate mechanism when it first launches. Since Kyaraben manages emulator updates, we need to preconfigure Dolphin to disable this prompt/feature.

---

## Reconsider flake generations and lock file approach

Each `kyaraben apply` creates a new flake generation directory and a new lock file:

```
warning: creating lock file '/home/fausto/.local/state/kyaraben/build/flake/generations/2026-02-02T07-55-32/flake.lock'
```

Questions to consider:
- Are we keeping these lock files around? It seems like we're just throwing them away.
- The warning is a bit ugly for users who don't care about nix internals.
- Should we reconsider the generations system for flake versions?

---

## Emulator health check for `kyaraben doctor`

Similar to the version checking script, it would be useful to have a way to quickly verify that installed emulators are actually working. This could be part of `kyaraben doctor`.

Challenges:
- Not all emulators support `--version` or `--help` flags (e.g., flycast, eden just launch the GUI)
- Spawning a bunch of GUI windows on the user's machine is disruptive

Possible approaches:
1. For emulators with `--version`/`--help` flags, use those (retroarch, duckstation, pcsx2, ppsspp, mgba, melonds, azahar, dolphin, cemu, vita3k, rpcs3)
2. For others, check that the binary exists and is executable
3. Could also verify the wrapper script points to a valid store path
4. Run `file` on the binary to ensure it's a valid ELF executable
5. For AppImages, could use `--appimage-extract --appimage-offset` to verify integrity without mounting

The goal is to catch issues like:
- Broken wrapper scripts (the melonds/vita3k `nix shell` issue)
- Missing binaries
- Corrupted downloads
- FUSE/permissions issues

This would complement the existing version check script by verifying runtime health rather than just version currency.

---

## RetroArch missing assets and fonts

RetroArch launches with no icons and uses an ugly bitmap font because the AppImage doesn't include assets. Users need to manually download them via RetroArch's Online Updater (Main Menu > Online Updater > Update Assets).

Options:
1. Document this as a required post-install step
2. Bundle RetroArch assets as a separate nix package
3. Pre-configure RetroArch to use a simpler menu that doesn't need assets (e.g., rgui instead of ozone/xmb)

---

## Consider backend preview command for apply status

The UI now shows appropriate messages for shared emulators ("Already installed for X", "In use by X"), but this logic lives in the frontend and duplicates knowledge about package sharing.

Consider adding a `CommandTypeApplyPreview` that returns per-system actions:

```go
type SystemActionPreview struct {
    SystemID         model.SystemID   `json:"systemId"`
    Action           string           `json:"action"` // will-install, will-update, will-uninstall, already-installed, shared-uninstall, no-change
    SharedWith       []string         `json:"sharedWith"`
    InstalledFor     []string         `json:"installedFor"`
    EffectiveVersion string           `json:"effectiveVersion"`
    InstalledVersion string           `json:"installedVersion"`
}
```

The backend already has everything needed:
- `manifest.InstalledEmulators` for what's installed
- `cfg.Systems` for enabled systems and selected emulators
- `registry.GetEmulator(id).Package.PackageName()` for package sharing (more accurate than the UI's string splitting)

The handler would:
1. Load config and manifest
2. Group enabled systems by package name
3. Check which packages are installed
4. Compute actions for each system

This moves all the logic to one place and makes the UI a simple display layer. The existing `Preflight` function in `apply.go` already does similar work for config patches and could be extended.

---

## CLI philosophy: minimal commands, editable config

The CLI should stay minimal. Rather than adding subcommands for every operation (`kyaraben enable psx`, `kyaraben add-emulator psx retroarch:mednafen`, `kyaraben list-emulators`, etc.), prefer:

1. **Simple, discoverable config format** - TOML is human-readable; users can edit `~/.config/kyaraben/config.toml` directly
2. **Few core commands** - `init`, `apply`, `status`, `doctor` cover the essential workflow
3. **Good defaults** - `kyaraben init` creates a working config; users remove what they don't want

This avoids:
- CLI flag/subcommand explosion
- Duplicating UI functionality in terminal
- Maintaining two interaction paradigms

The config file *is* the interface for advanced users. The UI is for users who don't want to edit files. The CLI is glue: initialize, apply changes, check status.

If users need to discover available emulators for a system, they can:
- Check the UI (shows all options)
- Read EMULATORS.md (reference doc)
- Look at a freshly-generated config from `kyaraben init` (shows defaults)

We should ensure EMULATORS.md stays current and documents all system → emulator mappings.

---

## UI: show all available emulators for enabled systems

Current gap: when a system has one emulator enabled, the UI shows a dropdown that switches between emulators but provides no way to enable a second one. The parent/child row pattern only appears when multiple emulators are already enabled (via config edit).

Proposed fix: always show all available emulators when a system is enabled.

**Current (single emulator):**
```
☑ PSX  [DuckStation ▼]  v0.1 ▼     ← dropdown switches, can't add
```

**Proposed:**
```
☑ PSX
   ☑ DuckStation                    v0.1 ▼
   ☐ RetroArch (Mednafen)           latest
   ☐ RetroArch (Beetle HW)          latest
```

- Enabled system shows all available emulators with checkboxes
- Checked = will be installed, unchecked = available but not enabled
- Removes the dropdown-that-switches pattern
- Consistent: checkboxes everywhere for enable/disable

**Disabled systems** stay as single row:
```
☐ PSX                               ← click to enable with defaults
```

Trade-off: more vertical space, but clearer affordance.

