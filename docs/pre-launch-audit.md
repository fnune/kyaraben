# Pre-launch audit

A comprehensive list of issues, gaps, and improvements to address before launching Kyaraben.

## Critical: missing files

### ~~No LICENSE file~~ FIXED

MIT license added to repository root.

### ~~No donations page~~ FIXED

Support page created at `site/src/content/docs/support.mdx` with GitHub Sponsors link and upstream project links. App sidebar links to the support page.

## Critical: infrastructure

### ~~Relay URL is hardcoded~~ FIXED

Relay URLs are now configurable via `sync.relays = ["..."]` in config.toml. The list supports multiple URLs for fallback (tried in order). When not set, defaults to the production relay. Empty list disables relay pairing.

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

## High priority: UX blockers from feedback.md

### Documentation lives on the website too much

From feedback.md: Kyaraben aims to be self-documenting, but critical information (emulator support table, troubleshooting) is only on the docs site. The app should surface this information in-context.

## High priority: technical gaps

### Single source of truth for folder IDs

From feedback.md: Kyaraben folder IDs are scattered across:
- `internal/model/system.go` (SystemID constants)
- `internal/sync/config.go` (constructs folder IDs)
- `integrations/nextui/internal/config/config.go` (duplicated)

If Kyaraben adds a new system, guest integrations do not know about it.

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

### No pause sync during gameplay

From feedback.md: users may want to pause sync while playing to avoid I/O contention. Currently requires stopping the systemd service manually.

### No Syncthing conflict UI

From feedback.md: when Syncthing creates conflict files, users have no indication in Kyaraben. A simple "X conflict files found" would help.

## Medium priority: features

### Extended directory support

From feedback.md:
- Cheats directory layout: per-emulator or per-system?
- DLC, patches, and updates: figure out folder structure
- Additional emulator paths: Flycast (BoxartPath, MappingsPath, TexturePath), PCSX2 (Cheats, Covers, Videos)

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

### ~~README.md is minimal~~ FIXED

README now includes installation instructions, link to releases page, system requirements, license badge, and contributing link.

### ~~Documentation site path in README is broken~~ FIXED

Contributing link now correctly points to `site/src/content/docs/contributing.mdx`.

### No FAQ

Common questions are not addressed (what systems are supported? does it work on Steam Deck? do I need to install emulators separately?).

## Security considerations

### Pairing code entropy

From feedback.md: 6 characters could be low entropy. While probably fine for the threat model, adding a confirmation step could help.

The relay server has rate limiting (`relay/internal/server/ratelimit.go`) but the pairing code is still short.

### No HTTPS for Syncthing UI

From feedback.md: Syncthing ships with HTTPS by default using self-signed certificates. Kyaraben uses HTTP.

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

## Privacy

### No privacy policy

The relay server processes device IDs. Users have no visibility into:
- What data is collected
- How long it is retained
- Whether any analytics are present

### No telemetry disclosure

The app checks for updates by fetching from GitHub. This discloses:
- IP address
- User-Agent
- Current version (implicitly)

Users should be informed of this network activity.

## ARM support

### ~~ARM is not officially supported~~ CLARIFIED

Documentation updated to explain partial ARM support. Some emulators have ARM builds (Eden, DuckStation, PPSSPP, Dolphin, Syncthing), but most do not (RetroArch cores, PCSX2, Cemu, Vita3K, RPCS3, Flycast, ES-DE, xemu, xenia-edge). ARM devices can use Kyaraben with limited emulator support.

## Release process

### No release automation for documentation site

The site is built in CI but there's no automatic deployment step visible.

### ~~No version in the UI~~ FIXED

Version is now displayed in settings.

### No rollback guidance

If an update breaks something, users have no documented way to roll back to a previous version.

## Other observations

### ~~.test directory is not gitignored~~ FIXED

The `*.test` pattern is now in .gitignore.

### plan.md in repository root

A `plan.md` file exists that appears to be development planning notes. Consider moving to docs/ or removing before launch.

### ~~No issue/PR templates~~ FIXED

Issue templates (bug_report.md, feature_request.md) and PR template added.

## Summary

### Must fix before launch

1. ~~Add LICENSE file~~ FIXED
2. ~~Add donations page~~ FIXED
3. ~~Fix or document relay URL configurability~~ FIXED
4. Re-enable E2E tests in CI
5. ~~Add basic README content (installation, requirements)~~ FIXED
6. ~~Fix broken documentation link in README~~ FIXED

### Should fix before launch

1. ~~Add privacy disclosure for update checks~~ FIXED
2. Consolidate folder ID definitions
3. ~~Add version display in UI~~ FIXED
4. ~~Clarify ARM support status~~ FIXED

### Nice to have for launch

1. FAQ
2. ~~Issue/PR templates~~ FIXED
3. Accessibility audit
