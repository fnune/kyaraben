# Sync improvements

Remaining work and future ideas for sync. Historical sections (pairing flow evolution, UI options exploration, completed implementation) have been removed.

## Remaining work

- Remove local discovery (UDP 21027)
- Rebuild the sync CLI to match the relay-based pairing system
- Pause sync during gameplay toggle
- Bandwidth limits
- Sync over metered connections
- Document non-Kyaraben devices connecting to Kyaraben's Syncthing (e.g., NextUI on Trimui Brick with its own Syncthing, different directory structure)

## Config ownership

Kyaraben owns its Syncthing config entirely. Manual edits are overwritten on next apply.

Supported: non-Kyaraben devices connecting to Kyaraben's Syncthing instance.

Not supported: custom folders in Kyaraben's Syncthing (users should run their own instance), manual edits to Kyaraben's Syncthing config.

Future first-class settings (instead of manual Syncthing config editing): bandwidth limits, versioning settings, relay server enable/disable, compression settings.

## Architecture notes

- All Syncthing API usage goes through `SyncClient` interface
- Response types are hand-rolled (Syncthing has no OpenAPI spec)
- Tests use `FakeClient`, not real Syncthing

## Ideas for future exploration

1. **Sync state in emulator cards**: show sync state per-system on each emulator card in the main view (e.g., "syncing" or "2 local saves" on the GameCube card) to make sync status visible without visiting the Sync tab.

2. **Fake Syncthing API for e2e tests**: a fake Syncthing HTTP server that keeps state in memory and exposes the same REST API. Connect via env var (`KYARABEN_SYNCTHING_URL`) to test the full sync UI flow.
