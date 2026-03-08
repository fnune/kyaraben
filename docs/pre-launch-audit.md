# Pre-launch audit

A comprehensive list of issues, gaps, and improvements to address before launching Kyaraben.

## Critical: infrastructure

### No releases URL or download page

The GitHub releases page is referenced in documentation but there's no:
- Stable download URL (e.g., kyaraben.dev/download)
- Latest release redirect (e.g., /releases/latest/download/Kyaraben-x86_64.AppImage)
- Version API endpoint for programmatic access

The updater in `ui/electron/updater.ts` hardcodes `https://api.github.com/repos/fnune/kyaraben/releases/latest`.

Action: set up stable download URLs, consider a landing page with download buttons.

### E2E tests are disabled in CI

In `.github/workflows/ci.yml:107`, the electron-e2e job has `if: false`. This means end-to-end tests never run.

Action: fix whatever is blocking E2E tests and re-enable them before launch.

## High priority: UX blockers

### No sync lifecycle controls in the app

The main Kyaraben app has no UI for:
- Stop/pause syncing
- Enable/disable sync on boot

Users must manually stop the systemd service. The NextUI guest app already supports these controls.

### Documentation lives on the website too much

From feedback.md: Kyaraben aims to be self-documenting, but critical information (emulator support table, troubleshooting) is only on the docs site. The app should surface this information in-context.

### Synchronization view is hard to use

The Synchronization view tries to do too much: device pairing, sync status, and folder configuration all in one place.

Action:
- Upstream sync information into the Catalog view (folder sync status, size on disk, last synced date)
- Refocus Synchronization view as a setup and device pairing view
- Add notifications for sync events (sync started, completed, conflicts detected)

### Preferences screen graphics

The preferences screen uses placeholder graphics. Need final artwork before launch.

### Import feature needs polish

The Import feature (for migrating from EmuDeck and similar) was rushed. The UI looks rough and its usefulness is unclear. Both the app UI and CLI counterpart need improvement before launch.

## High priority: technical gaps

### Headless server-as-sync-hub is untested

Kyaraben can run on a headless server as a central sync hub, but this use case has not been tested. Need to verify the setup works and document any required configuration.

### No app vs CLI feature parity assessment

The expectation is that the app and CLI have feature parity, but this has not been comprehensively verified.

### NextUI guest cannot unpair devices

The NextUI guest integration allows pairing devices but has no way to unpair them.

### setup.Disable() leaves orphaned data

From feedback.md: disabling sync removes the systemd unit but does not clean up Syncthing config or data directories.

### Unified folder configuration

From feedback.md: Kyaraben configures Syncthing folders two ways:
- `internal/sync/config.go` writes config.xml (before Syncthing starts)
- `internal/syncthing/client.go` uses REST API (after Syncthing running)
- `syncguest` does the config.xml approach

Both need the same input (folder mappings + device list) but produce different output formats.

## Medium priority: UX improvements

### DefaultOnly is uncomfortable

From feedback.md: if an emulator writes all its defaults to config, DefaultOnly does not work well. Consider detecting user edits vs. emulator process writes.

### Disruptive default hotkeys

From feedback.md: RetroArch has core-specific hotkeys that can disrupt gameplay. Example: R-Type on PC Engine triggers "Switching to 6 button layout".

### No Syncthing conflict UI

From feedback.md: when Syncthing creates conflict files, users have no indication in Kyaraben. A simple "X conflict files found" would help.

### Notification system needs polish

The sync notification system has rough edges:
- Multiple rapid sync events create notification spam instead of grouping
- Notifications with CTAs (e.g. "Go to catalog") stay open after the user clicks them

## Medium priority: features

### Extended directory support

From feedback.md:
- Cheats directory layout: per-emulator or per-system?
- DLC, patches, and updates: figure out folder structure
- Additional emulator paths: Flycast (BoxartPath, MappingsPath, TexturePath), PCSX2 (Cheats, Covers, Videos)

Eden 0.2.0 adds directory config support for updates and DLCs. Once Kyaraben tracks these directories, use Eden 0.2.0's directory config for patches/DLCs/updates.

### Cross-emulator presets

From feedback.md: toggle high-level features across all compatible emulators:
- Widescreen
- Integer scaling
- Discord presence
- RetroAchievements
- Auto-save on exit

### Storage breakdown bar

From feedback.md: show a color-coded bar indicating total size and composition of the collection directory.

### Handheld distro sync targets

From feedback.md: support SD-based handhelds running community distros (NextUI, PakUI, MinUI) with automatic structure translation.

## Medium priority: technical debt

### Electron UI needs consistent logging

The Electron UI has no consistent logging to `~/.local/state/kyaraben/kyaraben.log`. The daemon logs are captured, but the electron main process events (IPC, window lifecycle, errors) are not. This makes debugging production issues difficult.

Investigation context: while debugging relay pairing e2e tests, daemon logs confirmed progress events were being emitted to stdout, but the renderer never received them via IPC. Attempts to trace the issue failed because:
- `console.error` in electron/main.ts doesn't appear in e2e test output
- File-based logging (`fs.appendFileSync`) to kyarabenStateDir didn't produce files, suggesting either the directory path was wrong or the event handler code wasn't reached
- No visibility into whether readline receives the daemon's stdout lines or whether `webContents.send()` succeeds

The `pairing:progress` IPC event flow (daemon stdout -> readline -> JSON parse -> webContents.send -> renderer) has no observability. Similar issues likely affect `pendingDevice` events.

Action: implement structured logging in electron/main.ts that writes to the same log file as the daemon, or to a separate electron.log file in the same directory. Ensure the log path is resolved correctly (kyarabenStateDir uses XDG_STATE_HOME which may not be set at module load time).

### Environment variable security

From feedback.md: KYARABEN_* env vars (KYARABEN_RELEASES_URL, KYARABEN_VERSION, KYARABEN_NIX_PORTABLE_PATH) are useful for testing but could be risky if accidentally set in production. Consider a "test mode" flag or KYARABEN_TEST_* prefix.

### Syncthing config version hardcoded

From feedback.md: `config.go` hardcodes Syncthing config `Version: 37`. Low priority since Syncthing is backward compatible, but will eventually need a migration path.

## Low priority: polish

### Eden/Cemu splash screens

From feedback.md: emulator loading screens are ugly but cannot be avoided.

### ES-DE Steam entry has no icon

From feedback.md: the small square icon is missing.

### Compression guidance

From feedback.md: tools or documentation for ROM compression (CHD for disc-based, RVZ for GameCube/Wii, ZIP for cartridge-based).

### User store navigation

From feedback.md:
- Can we integrate with file managers to provide icons for each folder?
- Is it confusing that some things are per-system and others per-emulator?
- Would symlinks for convenience help or confuse?
- README.md files in folders?

## Documentation gaps

### No audit of emulator and system coverage

Before launch, audit which popular emulators and systems are missing. Compare against commonly requested systems and emulators to identify gaps.

### No FAQ

Common questions are not addressed (what systems are supported? does it work on Steam Deck? do I need to install emulators separately?).

## Security considerations

### Relay server security

The relay server stores device IDs temporarily during pairing. Consider:
- What data is logged?
- How long are pairing sessions stored?
- Is there any authentication beyond the pairing code?

## Accessibility

The UI has some accessibility support (34 occurrences of aria-/role= across 13 TSX files) but no comprehensive accessibility audit has been done.

Consider:
- Screen reader testing
- Keyboard navigation for all features
- Color contrast
- Focus management in modals

## Release process

### No release automation for documentation site

The site is built in CI but there's no automatic deployment step visible.

### No rollback guidance

If an update breaks something, users have no documented way to roll back to a previous version.

## Other observations

### plan.md in repository root

A `plan.md` file exists that appears to be development planning notes. Consider moving to docs/ or removing before launch.

## Summary

### Must fix before launch

1. Re-enable E2E tests in CI
2. Add sync lifecycle controls (stop/pause/enable/disable on boot)
3. Preferences screen graphics
4. Rework Synchronization view (upstream info to Catalog, refocus on setup/pairing)

### Should fix before launch

1. FAQ
2. Accessibility audit
3. Audit emulator and system coverage
4. Notification grouping and auto-close on CTA
5. Test headless server-as-sync-hub setup
6. Polish Import feature (UI and CLI)
7. Verify app vs CLI feature parity
8. Add unpair functionality to NextUI guest

### Nice to have for launch

1. Releases URL / download page
2. In-app documentation (emulator support table, troubleshooting)
