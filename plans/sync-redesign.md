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

## Recommendation: Option B (configuration only)

Rationale:
- Cleanest separation of concerns
- Syncthing's XML config is stable and well-documented
- systemd handles process lifecycle properly
- kyaraben stays a configuration tool, not a process manager

## Detailed design

### Syncthing installation

Same as emulators: download via `PackageInstaller`.

1. Add Syncthing to `versions.toml` with URLs for each target (x86_64, aarch64)
2. During apply, if sync enabled: `installer.InstallEmulator(ctx, "syncthing", ...)`
3. Binary ends up at `~/.local/state/kyaraben/packages/syncthing/{version}/bin/syncthing`

Syncthing releases: https://github.com/syncthing/syncthing/releases (~15MB tarballs)

### Configuration flow

```
User enables sync in kyaraben config
         ↓
kyaraben apply runs
         ↓
Download Syncthing binary (if not cached)
         ↓
Generate Syncthing config.xml based on:
  - Enabled systems (determines folders)
  - Sync mode (primary/secondary)
  - Paired devices
  - Port settings
         ↓
Generate systemd user unit:
  ~/.config/systemd/user/kyaraben-syncthing.service
         ↓
systemctl --user daemon-reload
systemctl --user enable --now kyaraben-syncthing
```

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

### Device pairing

Current approach requires manual device ID exchange. This is fine for v1.

Future improvement: generate a short pairing code that encodes device ID, share via QR or typed code.

### Status display

UI queries Syncthing REST API:
- `GET /rest/system/status` - device ID
- `GET /rest/system/connections` - connected devices
- `GET /rest/db/status?folder=X` - folder sync progress

Show in UI:
- Connection status (connected to N devices)
- Sync state (synced, syncing X files, error)
- Conflicts (if any)

### Conflict handling

Syncthing creates `.sync-conflict-*` files. kyaraben could:
1. Ignore (user resolves manually)
2. Detect and surface in UI
3. Provide resolution actions

Recommendation: option 2 for v1. Show conflicts, link to folder location.

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
- `internal/sync/client.go` - queries API for status (already works)
- `internal/sync/status.go` - status types (already works)
- `internal/sync/interface.go` - might simplify

### Implementation phases

Phase 1: basic sync
- Add Syncthing to versions.toml
- Install Syncthing during apply (if sync enabled)
- Generate config.xml during apply
- Generate and enable systemd unit
- Wire up API key between config and client
- Status display in UI (device ID, connection state)

Phase 2: polish
- Write .stignore files
- Folder-level sync progress
- Conflict detection and display
- Connection indicator in sidebar

Phase 3: future
- Simpler pairing flow
- Bandwidth controls
- Selective sync (choose which systems to sync)

## Open questions

1. Should Syncthing config be in `~/.config/kyaraben/syncthing/` or separate?
2. What if user already has Syncthing running on default ports?
3. How to handle sync disable (stop service, remove unit, or just stop)?
4. Should we support macOS/Windows eventually? (launchd, Windows services)
