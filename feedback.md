# Feedback for future work

## Consider breaking up large interfaces

`syncthing.SyncClient` has 30+ methods but most consumers only need a few. This makes testing difficult - you have to implement the entire interface even if you only use 3 methods.

Candidates for breaking up:
- `syncthing.SyncClient` - split into smaller role-based interfaces (StatusChecker, DeviceManager, FolderManager, etc.)
- Check other interfaces in `internal/syncthing/interface.go`

The guest app already introduced a local `SyncClient` interface with just 6 methods to work around this.
