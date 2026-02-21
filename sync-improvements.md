# Sync improvements

This document captures the current state of syncing in kyaraben, its history, identified problems, and proposed solutions. Covers both UI redesign and pairing flow improvements.

## Current implementation overview

### Pairing flow evolution

1. **mDNS with 6-digit code** (original): User-friendly codes like `ABCD23` but mDNS proved unreliable (blocked by guest WiFi, AP isolation, corporate networks, firewalls)

1. **UDP discovery** (attempted): Same firewall problems as mDNS

1. **Syncthing native discovery** (current): Uses Syncthing's built-in local discovery (UDP 21027) and pending devices API. Primary shows its full device ID (56 chars like `LGFPDIT7-SKNNJVJZ-...`), secondary enters it manually or picks from discovered list

The 6-digit code experience no longer exists in code but is still mentioned in documentation (`sync.mdx`).

### Current pairing UX

**Primary device:**

1. Enable sync as "primary"
2. Click "Start pairing" to enter pairing mode
3. Full device ID displayed with copy button
4. Spinner shows "Waiting for devices to connect..."
5. Polls `/rest/cluster/pending/devices` every 2 seconds
6. Auto-accepts first device that connects

**Secondary device:**

1. Enable sync as "secondary"
2. Sees discovered devices list (if any found via local discovery)
3. Can click "Enter device ID manually" to type the 56-char ID
4. Clicks "Connect" and waits

**Problems identified:**

- Device IDs are unwieldy (56 characters) for manual entry
- Local discovery only works on same network with UDP 21027 open
- No error shown when firewall blocks connection, just spins forever
- No connectivity check before attempting pairing

### Current sync status UI

The `SyncView` component shows:

1. **Status badges**: "Primary/Secondary" + "Running"
2. **Pairing section** (if no devices): Device ID display or discovery list
3. **Paired devices section**: List with connection dots (green/yellow/gray)
4. **Synced folders section**: All folders listed with:
   - Status dot (green=synced, blue pulsing=syncing, yellow=local changes)
   - Folder label
   - Size info (percent when syncing, size diff when mismatched)
   - "Open folder" button
   - Local changes warning box with Show/Revert buttons
5. **Advanced section** (collapsed): Syncthing web UI link + Reset sync

**Problems identified:**

- Too many folders displayed, hard to scan
- Sync state per folder is subtle (2px dots)
- No overall "is anything syncing right now?" indicator
- Local additions in receive-only folders show as warnings even when intentional
- Only option for local changes is "Revert", no "Accept" or "Ignore"
- No prominent sync progress when active
- SyncStatusBanner component exists but is not currently used in SyncView

### Polling strategy

```
Status polling:
- 2 seconds when syncing or service not running
- 10 seconds otherwise

Discovery polling:
- 3 seconds when secondary with no paired devices
```

---

## Problem areas

### Pairing problems

| Problem                                    | Impact                                          | Severity |
| ------------------------------------------ | ----------------------------------------------- | -------- |
| 56-char device IDs are hard to type        | Steam Deck users struggle with virtual keyboard | High     |
| Local discovery requires UDP 21027 open    | Silently fails if blocked                       | High     |
| No connectivity feedback during pairing    | Users wait indefinitely                         | High     |
| Documentation still mentions 6-digit codes | Confuses users                                  | Medium   |

### Sync status UI problems

| Problem                             | Impact                                    | Severity |
| ----------------------------------- | ----------------------------------------- | -------- |
| Too many folders listed             | Visual overload, hard to scan             | High     |
| Sync progress not prominent         | Users don't know if sync is active        | High     |
| Status dots too small               | Hard to see at a glance                   | Medium   |
| Local additions treated as warnings | Creates unnecessary alarm                 | Medium   |
| Only "Revert" option, no "Accept"   | Users must manually copy files to primary | Medium   |
| No "pause sync" per folder          | Can't temporarily stop large syncs        | Low      |
| Missing "sync while gaming" toggle  | Future feature, needs UI home             | Low      |

---

## Proposed solutions

### Pairing flow

**Decision:** Use a Kyaraben relay service with 6-digit codes as the default. Remove local discovery.

The relay server:

- Generates short-lived 6-digit pairing codes (5 minute TTL)
- Maps codes to Syncthing device IDs (stored in memory only)
- Works across any network (no firewall configuration needed)
- Rate limited per IP to prevent abuse
- After pairing, devices connect directly or via Syncthing relays
- Deployed to Koyeb with scale-to-zero for cost efficiency

**User flow:**

1. Primary clicks "Start pairing", gets 6-digit code (e.g., `ABC123`)
2. Code displayed prominently, valid for 5 minutes
3. Secondary enters code
4. Relay server exchanges device IDs
5. Devices connect

**Fallback:** Manual device ID entry available in Advanced settings for:

- Offline/air-gapped setups
- Users who want to use Syncthing's native relay features
- When the relay service is unavailable

The current local discovery (UDP 21027) can be removed since the relay handles discovery across any network.

---

### Sync status UI options

#### Option A: Simplified summary view

Replace folder list with high-level summary:

```
┌────────────────────────────────────────┐
│  ✓ All synced                          │
│  Last sync: 2 minutes ago              │
│                                        │
│  5 folders · 12.4 GB · 1 device        │
│                                        │
│  [Show details]  [Open Syncthing]      │
└────────────────────────────────────────┘
```

When syncing:

```
┌────────────────────────────────────────┐
│  ↻ Syncing...                          │
│  saves/gamecube                        │
│  ████████████░░░░░░░░░  56%           │
│                                        │
│  847 MB remaining · ~2 min left        │
│                                        │
│  [Show details]  [Pause]               │
└────────────────────────────────────────┘
```

Details expand to show folder list (collapsed by default).

**Pros:**

- Much cleaner at a glance
- Syncing state is obvious
- Reduces cognitive load

**Cons:**

- Less visibility into per-folder state
- Users may want folder-level control

#### Option B: Categorized folders with smart hiding

Group folders and hide "boring" ones:

```
Syncing now:
  saves/gamecube  ████████░░  78%  (234 MB left)

Local files:
  roms/gba  3 files only on this device  [Keep] [Remove]

Synced (4 folders):
  [Expand to see]
```

Show folders with local-only files expanded. Synced folders collapse into a single line.

**Pros:**

- Surfaces what matters
- Still shows all info on demand
- Clear actions for each state

**Cons:**

- More complex logic
- Animation/transitions needed for polish

#### Option C: Link to Syncthing UI

Provide a link that opens Syncthing's web UI in the user's browser (not embedded).

**Pros:**

- Full Syncthing functionality for power users
- No iframe security/styling issues
- Clean separation between Kyaraben and Syncthing

**Cons:**

- Users leave the app
- Different visual style

This is preferable to iframe embedding, which has UX and security issues.

#### Option D: Card-based layout (recommended)

Three cards: Status, Activity, and Settings.

```
┌─── Sync status ────────────────────────┐
│  ✓ All synced                          │
│  Connected to: steamdeck-kyaraben      │
└────────────────────────────────────────┘

┌─── Activity ───────────────────────────┐
│  No activity                           │
│  Last synced: saves/gamecube (2m ago)  │
└────────────────────────────────────────┘

OR when syncing:

┌─── Activity ───────────────────────────┐
│  Syncing: saves/gamecube               │
│  ████████████░░░░░░░░░  56%  234 MB    │
│                                        │
│  Queue: 2 more folders                 │
└────────────────────────────────────────┘

┌─── Folders (6) ────────────────────────┐
│  [Expand]                              │
│  · 2 have local-only files             │
└────────────────────────────────────────┘
```

Expanding "Folders" shows full list with actions.

---

### Local changes handling options

Current: Yellow warning + "Revert" button only.

#### Option A: Three-action menu (recommended)

For receive-only folders with local files:

```
roms/gba: 3 files only on this device
[Keep here] [Copy to primary] [Remove]
```

- **Keep here**: These files stay on this device, not synced elsewhere
- **Copy to primary**: Opens the primary's folder so user can add files there (they'll sync back)
- **Remove**: Delete the local files to match the primary

This is informational, not alarming. The UI shows the current state and offers clear paths forward without implying anything is wrong.

#### Option B: Mode selector

Let user change folder behavior:

```
roms/gba: 3 files only on this device

This folder receives from primary. Local files stay local.

○ Keep these files on this device
○ Remove these files
○ Change to two-way sync (files will sync to primary)
```

More complex but gives full control over folder behavior.

---

### Future settings home

The sync tab will need to house additional settings:

- Pause sync during gameplay (Game Mode)
- Sync over metered connections
- Bandwidth limits
- Folder-specific pause controls

**Recommendation:** Add a "Sync settings" card/section (collapsed by default) that exposes these as toggles. Keep the main view clean.

---

## Summary of recommendations

### Pairing

1. Build relay service with 6-digit codes as the default pairing method
2. Remove local discovery (UDP 21027)
3. Keep manual device ID entry as fallback in Advanced settings

### Sync status UI

1. Adopt card-based layout with Status, Activity, and Folders cards
2. Make sync progress prominent (Activity card)
3. Collapse synced folders by default, surface only those with local-only files
4. Create collapsed "Sync settings" section for advanced options

### Actions for local files

1. Three clear options: Keep here, Copy to primary, Remove
2. Language is factual, not alarming ("3 files only on this device")
3. "Copy to primary" guides user to the proper workflow

---

## Decisions

1. **Relay server**: Yes, this is the default pairing method. Local discovery removed.
2. **Syncthing UI access**: Link to browser (not iframe). Keeps Kyaraben clean, gives power users full access.
3. **Pause sync during gameplay**: Just a toggle in the Sync settings section.

### Config ownership

**Philosophy:** You do NOT touch Syncthing config owned by Kyaraben. Kyaraben applies cleanly.

**Supported:**

- **Non-Kyaraben devices connecting to Kyaraben's Syncthing**: Expected use case. Example: user runs NextUI on a Trimui Brick with their own Syncthing, connects to Kyaraben's primary to sync saves/ROMs with a different directory structure. Kyaraben is not installed on the other device. TODO: document this on the doc site.

**Not supported:**

- **Custom folders in Kyaraben's Syncthing**: Users should run their own Syncthing instance for this. Kyaraben leaves default Syncthing ports (22000, 21027) untouched.
- **Manual edits to Kyaraben's Syncthing config**: Will be overwritten on next apply.

**Future Kyaraben options** (instead of manual config editing):

- Bandwidth limits
- Versioning settings
- Relay server enable/disable
- Compression settings

These become first-class Kyaraben settings if users need them, rather than expecting users to manually edit Syncthing config.

---

## Architecture notes

### Current state (better than expected)

- `SyncClient` interface already exists (`internal/sync/interface.go`)
- `FakeClient` implementation exists for testing (`internal/sync/fakeclient.go`)
- Both `Client` and `FakeClient` are verified against the interface

### Types

Response types in `client.go` are hand-rolled (e.g., `FolderStatus`, `localChangedResponse`). Syncthing does not provide an OpenAPI spec - their [REST API](https://docs.syncthing.net/dev/rest.html) is documented per-endpoint in prose.

### Guidelines for new work

- All Syncthing API usage must go through `SyncClient` interface
- Add methods to interface as needed, implement in both `Client` and `FakeClient`
- Hand-roll types as needed (no spec to generate from)
- Tests use `FakeClient`, not real Syncthing

---

## Implementation progress

### Completed

**Sync status UI (card-based layout):**
- [x] StatusCard: mode badge, connection status, pairing UI
- [x] ActivityCard: sync progress with bar, or "all synced" with timestamp
- [x] FoldersCard: collapsible folder list with local changes summary
- [x] SyncSettingsSection: syncthing web UI link and reset button
- [x] LocalFilesActions: "Copy to primary..." and "Revert..." actions
- [x] SyncView refactored from 817 lines to ~210 lines

**Local files handling:**
- [x] Three clear options: "Copy to primary...", "Revert..." (removed "Keep here" as no-op)
- [x] Factual language ("X local changes" not "warning")
- [x] "Copy to primary" modal explains workflow
- [x] "Revert" has confirmation with file list
- [x] Hide "Copy to primary..." when only deletions exist

**Progress and status:**
- [x] Track lastSyncedAt for "last synced X ago" display
- [x] Progress bar uses correct formula: (globalSize - needSize) / globalSize
- [x] Refresh sync status on window focus
- [x] Refresh "Show details" on window focus
- [x] Directories show "(directory)" instead of misleading "0 KB"
- [x] Deleted files don't show size (was showing 0 KB)
- [x] Size labels: "X local / Y remote"

**Polling improvements:**
- [x] 1s when syncing or not running
- [x] 2s when viewing sync tab
- [x] 10s otherwise

**Backend improvements:**
- [x] Derive file action (changed/deleted) via filesystem stat check
- [x] Add /installed to sync ignore patterns (vita3k symlink)
- [x] Add open_url IPC handler for opening syncthing web UI
- [x] Sync config schema versioning (regenerate when defaults change)

**Relay server pairing:**
- [x] Build relay service with 6-digit codes (deployed to Koyeb)
- [x] Integrate relay client into daemon with URL fallback support
- [x] Keep manual device ID entry as fallback (advanced section)
- [ ] Remove local discovery (UDP 21027)
- [ ] Rebuild the sync CLI to match the new system
- [ ] Update documentation site

### Not started

**Future settings:**
- [ ] Pause sync during gameplay toggle
- [ ] Bandwidth limits
- [ ] Sync over metered connections

---

## Ideas for future exploration

1. **Sync state in emulator cards**: Since sync status is per-system (saves/gamecube, saves/psx, etc.), we could show sync state directly on each emulator card in the main view. For example, a small indicator showing "syncing" or "2 local saves" on the GameCube card. This would make sync status visible without visiting the Sync tab.

1. **Fake Syncthing API for e2e tests**: The existing `FakeClient` is for unit tests. For e2e tests, we could build a fake Syncthing HTTP server that keeps state in memory and exposes the same REST API. Kyaraben would connect to it via env var (e.g., `KYARABEN_SYNCTHING_URL`). This would let us test the full sync UI flow: enabling sync, pairing, watching progress, handling local files, etc.
