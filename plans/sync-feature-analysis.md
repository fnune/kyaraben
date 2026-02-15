# Sync feature analysis

**Superseded by `sync-redesign.md`** - this document is kept for reference on what exists in the codebase.

This document analyzes the current state of the syncing feature.

## Current state

The sync feature has significant infrastructure already built, but critical integration pieces are missing. The feature is currently non-functional.

### What exists

#### Model layer (`internal/model/sync.go`)
- `SyncConfig` with `Enabled`, `Mode`, `Syncthing`, `Devices`, `Ignore` fields
- `SyncMode` enum: `primary` and `secondary`
- `SyncDevice` struct with `ID` and `Name`
- `SyncthingConfig` for ports and relay settings
- `SyncIgnoreConfig` for ignore patterns
- `DefaultSyncConfig()` with sensible defaults

#### Config generator (`internal/sync/config.go`)
- Generates Syncthing XML configuration
- Creates per-system folders for roms, bios, saves, states
- Single folder for screenshots
- Primary sends roms/bios, bidirectional for saves/states/screenshots
- Secondary receives roms/bios, bidirectional for saves/states/screenshots
- Staggered versioning for saves and states (30-day retention)
- Auto-accept folders from paired devices on primary

#### HTTP client (`internal/sync/client.go`)
- `Client` implementing `SyncClient` interface
- `GetDeviceID()`: reads from Syncthing REST API
- `GetConnections()`: gets connected device status
- `GetFolderStatus()`: gets sync progress for a folder
- `GetStatus()`: aggregates device and folder status
- `IsRunning()`: pings Syncthing to check if running

#### Process management (`internal/sync/process.go`)
- `Process` struct for managing Syncthing subprocess
- `Start()`: starts Syncthing with correct flags
- `Stop()`: graceful shutdown with SIGTERM, fallback to KILL
- `waitForReady()`: polls until Syncthing responds
- `ensureAPIKey()`: generates and persists API key
- Environment variables: `STNODEFAULTFOLDER=1`, `STNOUPGRADE=1`

#### Status reporting (`internal/sync/status.go`)
- `Status` struct with overall state, devices, folders, conflicts
- `OverallState()`: derives state (synced, syncing, disconnected, conflict, error)
- Tests cover all state transitions

#### CLI (`internal/cli/sync.go`)
- `kyaraben sync status`: shows sync status, device ID, connected devices
- `kyaraben sync add-device <ID>`: adds a device to config
- `kyaraben sync remove-device <ID>`: removes a device from config
- Device ID validation (Syncthing base32 format)

#### Daemon handlers (`internal/daemon/daemon.go`)
- `handleSyncStatus`: returns current sync state
- `handleSyncAddDevice`: adds device to config
- `handleSyncRemoveDevice`: removes device from config
- Protocol types defined for all sync operations

#### UI component (`ui/src/components/SyncView/SyncView.tsx`)
- Shows enabled/disabled state
- Displays device ID with copy button
- Lists paired devices with connection status
- Add device form with ID and name inputs
- Remove device button
- "Open Syncthing UI" button

#### Documentation (`site/src/content/docs/sync.mdx`)
- Explains primary/secondary roles
- Documents sync directions per content type
- Setup instructions for primary and secondary devices
- Bundled Syncthing configuration details
- File versioning and ignore patterns

### What's missing

#### Critical blockers

1. **Syncthing process is never started**
   - `internal/sync/process.go` exists but nothing calls `Start()`
   - The daemon doesn't manage Syncthing lifecycle
   - No integration with apply process

2. **Syncthing binary not bundled**
   - `flake.nix` has `packages.syncthing` defined
   - No mechanism to download or install Syncthing binary
   - Path to binary not resolved at runtime

3. **Config not written before startup**
   - `ConfigGenerator.WriteConfig()` exists but isn't called
   - Syncthing would start with default config, not kyaraben's

4. **API key not shared**
   - `Process.ensureAPIKey()` generates key
   - `Client` has `SetAPIKey()` but never receives the key from Process
   - Would fail to authenticate with Syncthing API

#### Missing functionality

5. ~~**SyncView not wired to navigation**~~ (DONE)
   - Component is wired in App.tsx and Sidebar.tsx
   - Users can navigate to sync view via sidebar

6. **Ignore patterns not applied**
   - `SyncIgnoreConfig.Patterns` exists
   - No code writes `.stignore` files to folders

7. **No opaque folder handling**
   - Documentation mentions `opaque/` directory
   - No code creates or syncs this folder

8. **No folder status polling**
   - `GetFolderStatus()` exists
   - No periodic polling to update UI/CLI

9. **No conflict resolution UI**
   - `Conflict` type exists in status
   - No way to view or resolve conflicts

10. **No on-demand sync trigger**
    - Relies on Syncthing's filesystem watcher
    - No manual "sync now" button

#### Nice to have

11. **Connection status indicator in main UI**
    - No visual indicator in app chrome showing sync state

12. **Sync progress in status bar**
    - No way to see ongoing transfers

13. **Folder-level sync status**
    - `Folders` in Status is always empty (not populated)

## Architecture decisions needed

### When to start Syncthing?

Options:
1. **During apply**: start Syncthing after emulators are configured
2. **On daemon start**: always run Syncthing when daemon runs
3. **On-demand**: start when sync UI is opened, stop when closed

Recommendation: option 2 (on daemon start) if sync is enabled in config. This ensures saves sync immediately when playing games.

### How to bundle Syncthing?

Options:
1. **Include in nix-portable bundle**: increases bundle size but simplifies distribution
2. **Download separately**: smaller initial download but more moving parts
3. **Use system Syncthing**: conflicts with existing installations

Recommendation: option 1 (include in bundle). Syncthing is ~15MB, acceptable overhead for reliable distribution.

### How to handle first pairing?

The current flow requires:
1. User enables sync in config.toml
2. User runs `kyaraben sync status` to get device ID
3. User manually exchanges device IDs between machines

This works but is manual. Consider:
- QR code display/scan for device ID exchange
- Discovery via local network
- Relay-assisted pairing code

## Implementation order

Phase 1: make sync functional
- [ ] Add Syncthing binary to nix bundle
- [ ] Integrate sync.Process into daemon lifecycle
- [ ] Call ConfigGenerator.WriteConfig before startup
- [ ] Share API key between Process and Client
- [ ] Wire SyncView to app navigation
- [ ] Test basic two-device sync

Phase 2: improve reliability
- [ ] Write .stignore files from config
- [ ] Populate folder status in Status
- [ ] Add connection indicator to app chrome
- [ ] Add error handling for common failures
- [ ] Add opaque folder support

Phase 3: enhance UX
- [ ] Add conflict detection and resolution UI
- [ ] Add sync progress indicator
- [ ] Add "sync now" button
- [ ] Consider simpler pairing flow

## References

- `site/src/content/docs/sync.mdx`: user-facing documentation
- `feedback.md`: mentions "Sync UI" as nice-to-have
- `internal/sync/`: all sync implementation code
- `ui/src/components/SyncView/SyncView.tsx`: UI component
