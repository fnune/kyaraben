# Sync review

Remaining items from the sync feature review. Completed fixes (device removal race condition, ignore patterns, FakeClient folders, port checks, docs rewrite) have been removed.

## Open bugs

- `config.go` hardcodes Syncthing config `Version: 37`. Low priority since Syncthing is backward compatible, but will eventually need a migration path.
- Emoji in SyncStatusBanner.tsx (uses ✓, ↻, ○, ⚠, ✕, ●). Should use SVG icons or CSS.
- Inconsistent error wrapping: some use `fmt.Errorf("foo: %w", err)`, others `fmt.Errorf("foo %v", err)`.

## Maintainability concerns

1. Tight coupling to systemd (`exec.Command("systemctl", ...)`). On non-systemd systems, sync will not work.
2. `setup.Disable()` removes the systemd unit but does not clean up Syncthing config or data directories.

## Missing functionality

| Feature | Status |
|---------|--------|
| Conflict resolution | Not implemented |

## Remaining recommendations

### Medium priority

1. Implement conflict resolution UI
2. Add pause/resume buttons to UI

### Low priority

1. Add support for non-systemd service managers
2. Add Syncthing config migration support
