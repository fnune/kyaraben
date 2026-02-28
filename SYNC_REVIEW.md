# Sync review

Remaining items from the sync feature review. Completed fixes (device removal race condition, ignore patterns, FakeClient folders, port checks, docs rewrite) have been removed.

## Coverage gaps

| Area | Missing |
|------|---------|
| `client.go` | No unit tests for REST API calls (ShareFoldersWithDevice, IsPaused, etc.) |
| `setup.go` | No tests for Install, UpdateConfig, Disable |
| `systemd.go` | No tests for Write, Enable, Disable |
| Daemon handlers | No unit tests for handleSyncStatus, handleSyncRemoveDevice, etc. |

## Open bugs

- `config.go` hardcodes Syncthing config `Version: 37`. Low priority since Syncthing is backward compatible, but will eventually need a migration path.
- Emoji in SyncStatusBanner.tsx (uses ✓, ↻, ○, ⚠, ✕, ●). Should use SVG icons or CSS.
- Inconsistent error wrapping: some use `fmt.Errorf("foo: %w", err)`, others `fmt.Errorf("foo %v", err)`.

## Maintainability concerns

1. Devices are stored in both kyaraben config and Syncthing config. Consider removing kyaraben config storage entirely and using Syncthing as the single source of truth.
2. Tight coupling to systemd (`exec.Command("systemctl", ...)`). On non-systemd systems, sync will not work.
3. No retry logic for Syncthing API calls. Transient failures cause user-facing errors.
4. `setup.Disable()` removes the systemd unit but does not clean up Syncthing config or data directories.

## Missing functionality

| Feature | Status |
|---------|--------|
| Conflict resolution | Not implemented |

## Remaining recommendations

### High priority

1. Add retry logic to Syncthing API calls
2. Remove duplicate device storage in kyaraben config

### Medium priority

1. Add tests for client.go API methods
2. Add tests for setup.go and systemd.go
3. Implement conflict resolution UI
4. Add pause/resume buttons to UI

### Low priority

1. Add support for non-systemd service managers
2. Add Syncthing config migration support
