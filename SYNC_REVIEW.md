# Sync feature review

## Status

Review completed. Issues fixed:
- Device removal race condition
- Ignore patterns now written to .stignore files
- FakeClient now supports folders for testing
- Port availability check before starting Syncthing
- Documentation updated and simplified
- Test race condition in TestPrimaryPairingFlowWithFakes fixed

## Summary

The sync feature uses Syncthing to synchronize emulation data (ROMs, saves, states, screenshots) across devices. The implementation includes mDNS-based device discovery, a pairing protocol with 6-character codes, and proper primary/secondary role enforcement.

## Architecture

### Components

```
internal/sync/
├── client.go          # Syncthing REST API wrapper
├── config.go          # XML config generation
├── interface.go       # SyncClient interface
├── pairing.go         # Code generation, types
├── pairingflow.go     # Primary/secondary pairing flows
├── pairingclient.go   # HTTP client for pairing requests
├── pairingserver.go   # HTTP server accepting pairing
├── mdns.go            # mDNS advertiser and browser
├── setup.go           # Syncthing installation and systemd setup
├── status.go          # Status aggregation and OverallState
├── systemd.go         # Systemd unit file management
├── fakeclient.go      # Test fake for SyncClient
└── fakepairing.go     # Test fakes for Advertiser/Browser
```

### Data flow

1. User enables sync in UI -> daemon installs syncthing, writes config, creates systemd unit
2. User starts pairing -> daemon generates code, starts mDNS advertisement and HTTP server
3. Secondary discovers primary via mDNS -> user enters code -> HTTP POST to primary
4. Primary validates code -> both devices add each other to Syncthing config via REST API
5. Syncthing handles all subsequent sync operations

## Test coverage

### Good coverage

| Area | Tests | Notes |
|------|-------|-------|
| Pairing server | 5 tests | Code validation, rate limiting, callback |
| Pairing code generation | 2 tests | Format and uniqueness |
| Config generation | 4 tests | Primary/secondary folder types, devices, versioning |
| Status state machine | 8 tests | All OverallState transitions |
| Full pairing flow | 3 tests | Primary flow, secondary flow, cancellation |
| E2E sync | 3 tests | ROM sync, bidirectional saves, secondary isolation |

### Coverage gaps

| Area | Missing |
|------|---------|
| `client.go` | No unit tests for REST API calls (ShareFoldersWithDevice, IsPaused, etc.) |
| `setup.go` | No tests for Install, UpdateConfig, Disable |
| `systemd.go` | No tests for Write, Enable, Disable |
| `mdns.go` | No tests (relies on network, hard to test) |
| Daemon handlers | No unit tests for handleSyncStatus, handleSyncRemoveDevice, etc. |
| UI components | No tests for SyncView, SyncStatusBanner |

### Test quality

- Good use of fakes over mocks per contributing.mdx
- E2E test uses real syncthing binary when available
- Test helpers are well-designed (`testInstance`, `kyarabenInstance`)
- `testhelper_test.go` uses `t.TempDir()` - should use vfst per project conventions (though it's an E2E test requiring real fs)

## Documentation - FIXED

Documentation has been rewritten:
- Removed stale "Work in progress" disclaimer
- Fixed ports to match defaults (22100, 21127, 8484)
- Fixed config path to `~/.local/state/kyaraben/syncthing/`
- Removed incorrect "list of primaries" description
- Removed unimplemented conflict resolution section
- Simplified from ~300 lines to ~130 lines

## Potential bugs

### 1. Device removal race condition - FIXED

`handleSyncRemoveDevice` now gets device name from Syncthing before removal, removes from Syncthing (authoritative source), then attempts kyaraben config cleanup with only a warning on failure.

### 2. API key not set before first pairing

In `handleSyncStartPairing`, the API key is loaded and set on the client. If Syncthing was just installed and no API key has been saved yet, `loadSyncAPIKey()` might return empty. However, `setup.go:Install` writes the API key before Syncthing starts, so this should not occur in practice.

### 3. Secondary device discovery

The secondary flow takes the first mDNS offer without user selection. This is intentional for simplicity - most users have one primary. Documentation updated to reflect actual behavior.

### 4. Ignore patterns not applied - FIXED

`config.go:ConfigGenerator.WriteConfig` now calls `writeIgnoreFiles()` to create `.stignore` files in each synced folder with the configured ignore patterns.

### 5. FakeClient GetStatus doesn't include Folders - FIXED

`fakeclient.go` now has `SetFolderStatus` and `SetFolders` methods. `GetStatus` returns folders and `GetFolderStatus` returns configured folder status.

### 6. Port collision detection - FIXED

`setup.go:Install` now calls `CheckPorts()` to verify ports are available before installing Syncthing. Returns clear error if ports are in use.

### 7. Config version hardcoded

`config.go:119` hardcodes `Version: 37` for Syncthing config. Low priority - Syncthing is generally backward compatible with older config versions.

## Style violations

### Per contributing.mdx

| Violation | Location | Fix |
|-----------|----------|-----|
| Emoji in code | SyncStatusBanner.tsx uses ✓, ↻, ○, ⚠, ✕, ● | Use SVG icons or CSS |
| t.TempDir() in test | testhelper_test.go:33 | E2E test, acceptable exception |

### Minor style issues

| Issue | Location |
|-------|----------|
| Inconsistent error wrapping | Some use `fmt.Errorf("foo: %w", err)`, others `fmt.Errorf("foo %v", err)` |
| Unused variable suppression | `_ = code` at daemon.go:1021 - code return value is ignored |

## Maintainability

### Good patterns

- `SyncClient` interface enables testing with fakes
- `Advertiser` and `Browser` interfaces decouple mDNS implementation
- Dependency injection via constructors (NewSetup, NewSystemdUnit)
- Clear separation between config generation and runtime operations

### Concerns

1. **Kyaraben config vs Syncthing config duplication**: Devices are stored in both `kyaraben.yaml` (`cfg.Sync.Devices`) and Syncthing's config. GetStatus now reads from Syncthing, but pairing still persists to both. Consider removing kyaraben config storage entirely.

2. **No config migration**: If Syncthing config format changes, there's no migration path. The hardcoded `Version: 37` will eventually cause issues.

3. **Tight coupling to systemd**: `systemd.go` uses `exec.Command("systemctl", ...)` directly. On non-systemd systems (macOS, BSD, some containers), sync won't work. Consider abstracting service management.

4. **No retry logic**: API calls to Syncthing have no retry logic. Transient failures cause user-facing errors.

5. **Missing cleanup on disable**: When sync is disabled, `setup.Disable()` removes the systemd unit but doesn't clean up Syncthing config or data directories.

## Missing functionality

Per documentation vs implementation:

| Feature | Status |
|---------|--------|
| Pairing (local network) | Implemented |
| Manual device add (cross-network) | CLI only, no UI |
| Conflict resolution | Not implemented |
| Per-folder progress | Partial - UI shows folders but not detailed progress |
| Multiple primary discovery | Not implemented |

## Security considerations

### Good

- Pairing code never transmitted (displayed locally, entered manually)
- Rate limiting on pairing attempts (max 5)
- Code expires after 5 minutes
- Primary rejects other primaries (prevents misconfiguration)
- API key stored with 0600 permissions

### Concerns

1. **Pairing server on 0.0.0.0**: `pairingflow.go:52` binds to all interfaces. An attacker on the local network can try to brute-force the 6-char code. Rate limiting helps but 5 attempts may not be enough.

2. **No TLS for pairing**: Pairing HTTP server uses plain HTTP. Device IDs are exchanged in cleartext. On untrusted networks, an attacker could intercept device IDs (though they still need the code).

3. **mDNS exposes hostname**: The mDNS advertisement includes the hostname, potentially revealing device identity to network observers.

## Recommendations

### Completed

1. ~~Update sync.mdx to remove "Work in progress" disclaimer and fix port/path inaccuracies~~ Done
2. ~~Implement ignore patterns (create .stignore file)~~ Done
3. ~~Add tests for FakeClient folders~~ Done
4. ~~Add port availability checks~~ Done

### Remaining high priority

1. Add retry logic to Syncthing API calls
2. Consider removing duplicate device storage in kyaraben config

### Remaining medium priority

1. Add tests for client.go API methods
2. Add tests for setup.go and systemd.go
3. Implement conflict resolution UI
4. Add pause/resume buttons to UI
5. Replace emoji with proper icons in SyncStatusBanner

### Remaining low priority

1. Add support for non-systemd service managers
2. Allow user to select from multiple discovered primaries
3. Add TLS option for pairing server
4. Add Syncthing config migration support
