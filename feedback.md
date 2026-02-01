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

---

## RetroArch download size is misleading

The download size shown for RetroArch emulators (e.g., "RetroArch (bsnes)") displays the size of the RetroArch AppImage itself, not:
- RetroArch + the specific core being enabled
- Just the core size (since cores come from nix separately)

This is misleading because:
1. If RetroArch is already installed, enabling a new core only downloads the core (from nix), not RetroArch again
2. Users might think each RetroArch variant is a separate 170MB download
3. The change summary total (e.g., "Downloading 500 MB") is wildly inaccurate when enabling multiple RetroArch cores

The architecture is: RetroArch AppImage is downloaded once if any system needs it, then cores are downloaded individually from nix for each enabled system. The current UI has no way to represent this.

Possible solutions:
- Track RetroArch separately from cores in the download calculation
- Show core size separately (would need to fetch from nix or estimate)
- Show "RetroArch: 170MB + cores from nix" as a note
- Don't show size for RetroArch emulators since cores come from nix
- Have the backend return more granular download info that accounts for shared packages

---

## Garbage collection via nix-portable

We need a way for Kyaraben to run garbage collection using nix-portable to free up disk space. This would:

1. Clean up unused store paths from previous generations
2. Remove old flake generations that are no longer needed
3. Provide a UI affordance (e.g., in settings or installation view) to trigger cleanup
4. Show how much space would be / was freed

This is important because the nix store can grow significantly over time with updates.

---

## Spurious `internal/apply/generations/` directories

During development, `internal/apply/generations/` directories appeared in the repository. This is unexpected since generations should only be created in the user's state directory (`~/.local/state/kyaraben/build/flake/generations/`), not in the source tree.

This needs investigation:
1. Find the root cause (why are generations being created in the working directory?)
2. Clean up any existing spurious directories
3. Add `.gitignore` entries if needed as a safeguard
4. Fix the code to always use the proper state directory path

---

## CRITICAL: manifest.json disappearing, losing track of installed emulators

Sometimes on a fresh launch, kyaraben loses track of all installed emulators and thinks everything needs to be installed again. The manifest.json file appears to be disappearing or getting corrupted.

Possible causes to investigate:
1. Race condition on app close (manifest written while app is shutting down?)
2. Cancelling an installation mid-way corrupts or deletes the manifest
3. Concurrent writes to manifest from multiple goroutines
4. File not being flushed/synced before process exits
5. Error during manifest write silently failing

This is a core failure mode that makes kyaraben unreliable. Users should never lose their installation state. Priority fixes:

1. Add atomic writes for manifest (write to temp file, then rename)
2. Add manifest backup before any modification
3. Log all manifest reads/writes for debugging
4. Verify manifest integrity on load
5. Handle cancellation gracefully (don't modify manifest until operation succeeds)
6. Consider keeping manifest history/versions for recovery

---

## "Discard changes" button shown when config differs from manifest

When the user uninstalls everything via CLI and then opens the UI, the config.toml still expects emulators to be installed. The UI correctly shows the diff (e.g., "1.2GB to download"), but it also shows a "Discard changes" button.

This is confusing because:
1. The user didn't make any changes in the UI
2. The config.toml is the source of truth for what should be installed
3. "Discard changes" implies reverting user actions, but there were none

The real situation is: "config wants X installed, but X is not installed yet." The action isn't "discard my changes" but rather "sync config to match current state" or "I don't want these emulators anymore."

Possible solutions:
1. Rename button to "Reset to installed state" or "Clear pending installs"
2. Only show "Discard changes" when the user has made UI modifications in this session
3. Track whether changes came from config vs UI and show different messaging
4. Show "Config expects: X, Y, Z. Currently installed: none. [Apply] [Edit config]"

---

## Provision links clickable when emulator is disabled

In EmulatorSubcard, the provision items (folder buttons, copy buttons, launch buttons) remain interactive even when the emulator card is disabled (e.g., slated for removal). This is inconsistent - if the emulator is disabled, interacting with its provisions doesn't make sense.

Fix: Disable or hide provision action buttons when the emulator is disabled (`enabled={false}`).

---

## Use kyaraben-specific subdirectories for generated files [DONE]

Kyaraben now uses kyaraben-specific locations for generated files:

- Desktop files: `~/.local/share/applications/kyaraben/*.desktop`
- Icons: `~/.local/share/icons/hicolor/*/apps/kyaraben-*.{svg,png}` (prefixed rather than subdirectory, since hicolor icon theme requires standard structure)

Benefits:
1. Easier to uninstall - desktop files can be removed by deleting the subdirectory
2. Icons are easily identifiable by their `kyaraben-` prefix
3. Avoids filename clashes with other applications
4. Migration from old paths happens automatically on next `kyaraben apply`

---

## Reconsider the manifest

Now that kyaraben uses dedicated directories for most things it installs:
- `~/.local/share/applications/kyaraben/` for desktop files
- `~/.local/share/icons/hicolor/*/apps/kyaraben-*` for icons
- `~/.local/bin/kyaraben*` for binaries

What is the manifest still useful for?

Current uses:
1. Tracking managed configs (scattered in emulator-specific config dirs)
2. Last applied timestamp (for determining pending changes)
3. Installed emulator versions

Potential simplification:
- Scan kyaraben directories at runtime instead of tracking files in manifest
- Only keep managed configs and versions in manifest
- Or eliminate manifest entirely and derive state from filesystem + config

The manifest has been a source of bugs (disappearing, corruption). Reducing its role or eliminating it could improve reliability.

---

## Daemon protocol request/response correlation [DONE]

Added UUID-based request IDs to the daemon protocol for proper request/response matching. Previously, the Electron handler used FIFO matching which broke when commands were sent during apply (their responses got matched to the wrong handler).

---

## Electron user data location [DONE]

Moved Electron's user data from `~/.config/kyaraben-ui/` to `~/.local/state/kyaraben/ui/` following XDG conventions. Cache, cookies, session storage is state, not config.

---

## RetroArch cores building from source [DONE]

Pinned `retroarch-cores` in versions.toml to a hydra-cached nixpkgs commit to avoid building libretro cores from source on every apply.

---

## Vita3K opaque config location

Vita3K config lives in `~/Emulation/opaque/vita3k/config.yml` because the emulator takes a single `-c` path for its entire user directory. We can't separate config from data with Vita3K's current architecture. This is intentional - same pattern as Dolphin and Eden.

