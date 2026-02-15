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
- [ ] Integrate Setup.Install into apply.go
- [x] Add SyncView route to app navigation
- [ ] End-to-end test: two machines syncing a save file

## Phase 2: improve reliability

- [ ] Generate .stignore files from config patterns
- [ ] Fetch folder status in GetStatus
- [ ] Add sync connection indicator to sidebar
- [ ] Handle systemd unit updates on config change
- [ ] Log sync errors to kyaraben log

## Phase 3: enhance UX

- [ ] Show sync conflicts in UI
- [ ] Add conflict resolution actions
- [ ] Show sync progress for large transfers
- [ ] Evaluate simpler pairing (QR, local discovery)

## Deferred

- Handheld distro sync targets (see feedback.md)
- Per-folder sync settings in UI
- Bandwidth limiting controls
- macOS/Windows support (launchd, Windows services)
