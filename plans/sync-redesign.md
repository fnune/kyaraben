# Sync feature redesign

Starting from zero. What do we actually need?

## Goals

1. Sync saves and states between devices (play on Steam Deck, continue on desktop)
2. Optionally distribute ROMs from a primary device to secondaries
3. Seamless setup, minimal ongoing maintenance

## Non-goals

- Being a Syncthing wrapper/GUI
- Managing Syncthing as a process
- Supporting non-Syncthing sync backends

## Design options

### Option A: documentation only

kyaraben stays out of sync entirely. We document:
- Recommended Syncthing folder setup
- How to pair devices
- Directory conventions that make sync easy

Pros:
- Zero code to maintain
- Users who know Syncthing are happy
- No coupling to Syncthing version/API changes

Cons:
- Poor UX for users unfamiliar with Syncthing
- No visibility into sync status in kyaraben
- Manual folder setup when enabling new systems

### Option B: configuration only

kyaraben generates Syncthing configuration but doesn't run it.

During apply:
1. Download Syncthing binary (like emulators)
2. Generate `~/.config/kyaraben/syncthing/config.xml` from kyaraben config
3. Generate a systemd user unit that runs Syncthing
4. Enable/start the unit

kyaraben UI:
- Query Syncthing REST API for status display
- Show device ID for pairing
- Provide add/remove device (modifies kyaraben config, regenerates Syncthing config)

Pros:
- Clean separation: kyaraben configures, systemd runs, Syncthing syncs
- Automatic folder setup when systems are enabled
- Status visibility in UI
- Crash recovery via systemd

Cons:
- Need to track Syncthing releases in versions.toml
- Syncthing config format may change between versions
- Two configs to keep in sync (kyaraben's and Syncthing's)

### Option C: API-driven configuration

Like option B, but configure Syncthing via REST API instead of writing config.xml.

During apply:
1. Ensure Syncthing is running (systemd unit)
2. Call Syncthing REST API to add/remove folders, devices
3. Syncthing persists its own config

Pros:
- No config file format coupling
- Can modify running Syncthing without restart
- Syncthing manages its own config consistency

Cons:
- Syncthing must be running during apply
- More complex error handling (API failures)
- Harder to reason about state (where is truth?)

## Recommendation: hybrid B+C

Option B for initial setup (write config.xml, generate systemd unit). Option
C for runtime mutations (use the REST API for pairing, device management, and
status). This gives reliable initial config generation while avoiding
restarts for day-to-day operations.

Rationale:
- Initial config.xml generation is reliable and well-tested
- systemd handles process lifecycle
- REST API enables pairing without restart (`PUT /rest/config/devices/<id>`)
- REST API events SSE stream enables real-time UI updates
- config.toml remains the source of truth (persisted after REST API changes)

## Detailed design

### Syncthing installation

Same as emulators: download via `PackageInstaller`. But triggered by
`kyaraben sync pair`, not by `kyaraben apply`. Sync setup is decoupled from
the emulator apply flow.

1. Add Syncthing to `versions.toml` with URLs for each target (x86_64, aarch64)
2. On first `sync pair`: `installer.InstallEmulator(ctx, "syncthing", ...)`
3. Binary ends up at `~/.local/state/kyaraben/packages/syncthing/{version}/bin/syncthing`

Syncthing releases: https://github.com/syncthing/syncthing/releases (~15MB tarballs)

### Setup flow

```
User runs: kyaraben sync pair (or clicks "Pair a device" in UI)
         ↓
Is Syncthing installed and running?
  No → download binary, generate config.xml, generate systemd unit,
       enable/start service, update config.toml
  Yes → continue
         ↓
Generate pairing code, display on screen
         ↓
Advertise via mDNS: <hostname>._kyaraben._tcp.local
  TXT record: pairing endpoint URL only
  (code and device ID are NOT in the advertisement)
         ↓
Wait for secondary to connect to pairing endpoint
         ↓
Secondary sends: { code, device_id }
         ↓
Primary verifies code, responds with its own device ID
         ↓
Both sides add each other via Syncthing REST API:
  PUT /rest/config/devices/<peer-id>   (takes effect immediately)
  PATCH /rest/config/folders/<id>      (share folders with new device)
         ↓
Persist device to config.toml
         ↓
Stop mDNS advertisement, expire code
```

### Why mDNS is needed (cannot use Syncthing native discovery)

Syncthing's local discovery broadcasts device IDs on the LAN but only
processes broadcasts from devices already in its config. There is no REST
API endpoint for "show me all Syncthing instances on this network." The
bootstrap problem (getting two unknown devices to discover each other) is
what Syncthing does not solve.

mDNS fills this gap: it announces kyaraben instances as pairable and
provides the initial ID exchange channel. After pairing, Syncthing's native
discovery handles everything else (address resolution, reconnection, relay
fallback). Kyaraben's mDNS is only active during the 5-minute pairing
window.

### Pairing protocol security

The pairing code is never transmitted over the network. It is displayed on
the primary's screen and typed by hand on the secondary.

The mDNS advertisement reveals:
- That a kyaraben instance exists at this hostname/IP
- The pairing endpoint URL

It does NOT reveal:
- The pairing code
- The Syncthing device ID

An attacker on the LAN can see the advertisement but cannot pair without the
code. The pairing endpoint rate-limits to 5 attempts per session. The code
expires after 5 minutes.

### Folder structure

```
~/Emulation/
├── roms/{system}/        # primary→secondary
├── bios/{system}/        # primary→secondary
├── saves/{system}/       # bidirectional, versioned
├── states/{system}/      # bidirectional, versioned
└── screenshots/          # bidirectional
```

Each becomes a Syncthing folder. Folder IDs: `kyaraben-{category}-{system}` or `kyaraben-{category}` for non-system-specific.

### REST API usage

Initial setup (config file generation):
- config.xml written before first Syncthing start
- systemd unit generated and enabled

Runtime mutations via REST API (no restart):
- `PUT /rest/config/devices/<id>` - add/remove devices after pairing
- `PATCH /rest/config/folders/<id>` - update folder device lists
- `GET /rest/events` - SSE stream for real-time sync status
- `GET /rest/db/completion?device=<id>&folder=<id>` - per-device progress
- `GET /rest/stats/device` - device statistics, last seen
- `GET /rest/cluster/pending/devices` - accept/reject incoming devices
- `POST /rest/db/override` / `POST /rest/db/revert` - conflict resolution
- `POST /rest/system/pause` / `POST /rest/system/resume` - flow control

Status display (read-only, existing):
- `GET /rest/system/status` - device ID
- `GET /rest/system/connections` - connected devices
- `GET /rest/db/status?folder=X` - folder sync progress

### Cross-network pairing improvement

For manual pairing (devices not on the same LAN), the pending devices API
eliminates one direction of ID exchange:

1. User adds primary's ID on secondary via `kyaraben sync add-device`
2. Secondary's Syncthing tries to connect to primary
3. Primary sees secondary in `GET /rest/cluster/pending/devices`
4. Primary's UI/CLI shows "steamdeck-kyaraben wants to pair - accept?"
5. User accepts, primary adds secondary via REST API

This reduces the manual flow from "exchange IDs in both directions" to "enter
one ID, accept on the other side."

### Conflict handling

Syncthing creates `.sync-conflict-*` files. kyaraben surfaces these in UI
and CLI with resolution actions:
- Keep local version (`POST /rest/db/override`)
- Keep remote version (`POST /rest/db/revert`)
- `kyaraben sync resolve` for CLI

### Systemd unit

```ini
[Unit]
Description=Kyaraben Syncthing
After=network.target

[Service]
Type=simple
ExecStart=%h/.local/state/kyaraben/packages/syncthing/VERSION/bin/syncthing serve --no-browser --no-default-folder --config=%h/.config/kyaraben/syncthing --data=%h/.local/state/kyaraben/syncthing
Restart=on-failure
RestartSec=10
Environment=STNODEFAULTFOLDER=1
Environment=STNOUPGRADE=1

[Install]
WantedBy=default.target
```

Note: VERSION would be resolved at generation time.

### What to delete

Remove entirely:
- `internal/sync/process.go` - systemd handles this
- `internal/sync/fake.go` - testing via API mock instead

Keep and modify:
- `internal/sync/config.go` - generates config.xml (already works)
- `internal/sync/client.go` - extend with write operations (REST API mutations)
- `internal/sync/status.go` - status types (already works)
- `internal/sync/interface.go` - extend with write methods

New:
- `internal/sync/pairing.go` - mDNS advertisement/browsing, HTTP endpoint
- `internal/sync/events.go` - SSE event stream consumer

## Open questions

1. ~~Should Syncthing config be in `~/.config/kyaraben/syncthing/` or separate?~~ Resolved: `~/.config/kyaraben/syncthing/`
2. What if user already has Syncthing running on default ports? (kyaraben uses non-default ports: 22001, 21028, 8385)
3. ~~How to handle sync disable?~~ Resolved: `SystemdUnit.Disable()` stops and removes service
4. Should the pairing endpoint use TLS? (self-signed cert would prevent passive eavesdropping of the code during HTTP exchange, but adds complexity for a one-time LAN operation)
