# Sync feature progress

Checklist for tracking sync feature implementation.

## Cleanup

- [x] Remove `packages.syncthing` from flake.nix (shouldn't be there)
- [x] Delete `internal/sync/process.go` (systemd handles this)
- [x] Delete `internal/sync/fake.go` (unused)

## Phase 1: basic sync

- [x] Add Syncthing to versions.toml (x86_64, aarch64 targets)
- [x] Install Syncthing during apply when sync is enabled (Setup.Install)
- [x] Generate Syncthing config.xml during apply (ConfigGenerator.WriteConfig)
- [x] Generate systemd user unit during apply (SystemdUnit.Write)
- [x] Enable/start systemd unit after apply (SystemdUnit.Enable)
- [x] Wire API key from config generation to client (Setup manages this)
- [ ] Decouple sync setup from apply (sync pair triggers install, not apply)
- [x] Add SyncView route to app navigation
- [ ] End-to-end test: two machines syncing a save file

## Phase 2: mDNS pairing

- [ ] Implement mDNS advertisement (`<hostname>._kyaraben._tcp.local`)
  - TXT record: pairing endpoint URL only
  - No device ID or pairing code in advertisement
- [ ] Implement mDNS browser for discovering nearby primaries
- [ ] Implement pairing HTTP endpoint on primary
  - Accepts: code + secondary device ID
  - Validates code, rate-limits attempts (5 per session)
  - Responds with primary device ID
  - Code expires after 5 minutes
- [ ] Implement `kyaraben sync pair` CLI command
  - Without code: become primary, install Syncthing if needed, advertise,
    display code, wait
  - With code: become secondary, install Syncthing if needed, browse,
    exchange IDs
- [ ] Use Syncthing REST API for runtime config changes after pairing
  - `PUT /rest/config/devices/<id>` to add peer without restart
  - `PATCH /rest/config/folders/<id>` to share folders with new device
  - Persist to config.toml as source of truth
- [ ] UI: pairing flow in Settings > Sync
  - Primary: "Pair a device" button, shows code on screen
  - Secondary: "Join a primary" button, shows discovered primaries, code
    entry
- [ ] E2e test for pairing protocol (see testing notes below)

## Phase 3: sync management via REST API

- [ ] Generate .stignore files from config patterns
- [ ] Use `/rest/events` SSE stream for real-time sync status in UI
  - Replace polling in GetFolderStatus with event-driven updates
- [ ] Use `/rest/db/completion` for per-device per-folder progress
- [ ] Use `/rest/stats/device` for last-seen timestamps
- [ ] Add sync connection indicator to sidebar
- [ ] Handle systemd unit updates on config change
- [ ] Log sync errors to kyaraben log
- [ ] Improve manual cross-network pairing with pending devices API
  - User adds primary ID on secondary
  - Primary sees secondary in `/rest/cluster/pending/devices`
  - Primary UI shows accept/reject prompt instead of requiring manual
    add-device in both directions

## Phase 4: conflict resolution and controls

- [ ] Show sync conflicts in UI
  - Use `/rest/db/need` to detect conflict files
- [ ] Add conflict resolution actions
  - `POST /rest/db/override` to keep local version
  - `POST /rest/db/revert` to keep remote version
  - `kyaraben sync resolve` CLI command
- [ ] Show sync progress for large transfers
  - Per-category aggregation (ROMs, saves, etc.)
- [ ] Pause/resume sync
  - `POST /rest/system/pause` / `POST /rest/system/resume`
  - `kyaraben sync pause` / `kyaraben sync resume` CLI commands
  - Pause button in UI

## Deferred

- Handheld distro sync targets (see feedback.md)
- Per-folder sync settings in UI
- Bandwidth limiting controls
- macOS/Windows support (launchd, Windows services)

## Testing notes

### mDNS pairing e2e test strategy

The existing e2e test (`kyaraben_e2e_test.go`) validates file sync with
pre-configured peers. For pairing tests:

1. Test mDNS and HTTP separately behind interfaces:
   - `PairingAdvertiser` / `PairingBrowser` for the discovery layer
   - Unit test the HTTP pairing exchange with `httptest.Server`
2. Integration test uses a fake `PairingBrowser` (returns hardcoded offer)
   wired to the real HTTP exchange and real Syncthing instances
3. Real mDNS test on loopback (skip in CI if multicast unavailable):
   - Force mDNS library to use `lo` interface
   - `t.Skip` if `loopbackSupportsMulticast()` returns false

### Why mDNS cannot be replaced by Syncthing native discovery

Syncthing's local discovery broadcasts device IDs on the LAN but only
processes broadcasts from devices already in its config. There is no REST
API endpoint for "show me all Syncthing instances on this network." The
bootstrap problem (two devices that do not know each other's IDs discovering
each other for the first time) is what Syncthing itself does not solve.

mDNS fills this gap: it lets kyaraben instances announce themselves as
pairable and provides the channel for the initial device ID exchange. After
pairing completes, Syncthing's native discovery handles everything else
(address resolution, reconnection, relay fallback). Kyaraben's mDNS is only
active during the pairing window.
