# Headless server support plan

This document captures the current state of headless server support in Kyaraben and outlines what's needed to properly support the "server as sync hub" use case.

## Use case

A headless Linux server (e.g., a media server like bilbo) acts as a central sync hub:

- Always-on, stores the canonical copy of all game data
- Desktop and Steam Deck sync their saves/states/ROMs through the server
- Devices can sync even when not online at the same time (server bridges the gap)
- No display, no emulators needed on the server itself

## Current state

### What already works

1. **CLI-only release artifact exists**
   - `.goreleaser.yaml` produces `kyaraben-cli-{version}-linux-amd64.tar.gz`
   - Available on GitHub releases alongside the AppImage

2. **CLI sync commands are functional and tested**
   - `kyaraben sync status` - show sync state
   - `kyaraben sync pair [CODE]` - relay-based pairing with 6-digit codes
   - `kyaraben sync pair --device-id` - manual pairing with full device ID
   - `kyaraben sync add-device <ID>` - add device directly
   - `kyaraben sync remove-device <ID>` - remove paired device
   - E2E tests exist in `test/e2e/sync_test.go`

3. **User systemd services work for headless**
   - Syncthing runs as `~/.config/systemd/user/kyaraben-syncthing.service`
   - With `loginctl enable-linger <user>`, services persist without login
   - This is the standard pattern for user services on headless servers

4. **Empty systems config is handled gracefully**
   - `kyaraben status` shows "Enabled systems: none"
   - `kyaraben apply` doesn't fail with no systems

### What's missing

1. **Documentation for headless setup**
   - No guide explaining CLI-only installation
   - No minimal sync-only config example
   - No mention of `loginctl enable-linger` for headless operation

2. **The folder creation gap** (see below)

3. **Manual verification on real hardware**
   - E2E tests use fake Syncthing
   - Real end-to-end headless deployment hasn't been tested

## The folder creation gap

This is the core design issue blocking headless server support.

### How folder creation works today

From `internal/sync/config.go`:

```go
func (g *ConfigGenerator) FolderCreateRequests() []syncthing.FolderCreateRequest {
    specs := folders.GenerateSpecs(folders.HostInput{
        Systems:   g.allSystems,      // from config [systems]
        Emulators: g.allEmulators,    // derived from systems
        Frontends: g.allFrontends,
    })
    // ...
}
```

Folder structure created:
- `roms/{system}` - per enabled system
- `bios/{system}` - per enabled system
- `saves/{system}` - per enabled system
- `states/{emulator}` - per enabled emulator
- `screenshots/{emulator}` - per enabled emulator
- `frontends/{frontend}/*` - per enabled frontend

### The problem

When you enable a system in `config.toml`:

```toml
[systems]
snes = ["retroarch:bsnes"]
```

Kyaraben does two things:
1. Creates the folder structure for that system
2. Installs the emulator (retroarch with bsnes core)

On a headless server:
- We want (1) folder creation
- We don't want (2) emulator installation

With no systems enabled:
- No folders are created
- Syncthing has nothing to sync
- The server is useless as a sync hub

### Chosen approach: `[global] headless = true`

**Simple `headless = true` flag in `[global]`:**

```toml
[global]
headless = true
collection = "/mnt/storage/Emulation"

[sync]
enabled = true
```

In headless mode:
- Ignore `[systems]` for folder creation
- Create folders for ALL known systems using default emulators
- Skip emulator installation, config generation, desktop entries
- Updates adding new systems automatically work (no config changes needed)

For emulator-specific folders (`states/{emulator}`, `screenshots/{emulator}`), use the default emulator for each system (same mapping as `NewDefaultConfig()`).

## CLI self-update

The Electron app has an updater (`ui/electron/updater.ts`) that:
1. Checks GitHub releases API for latest version
2. Compares with current version using semver
3. Downloads the new AppImage
4. User restarts to apply

The CLI needs equivalent functionality for headless servers.

### Proposed: `kyaraben update` command

```
$ kyaraben update

Checking for updates...
Current version: 0.1.0
Latest version:  0.2.0

Downloading kyaraben-cli-0.2.0-linux-amd64.tar.gz...
[████████████████████████] 100%

Update downloaded. Replace current binary? [y/N] y
Updated successfully. Run 'kyaraben version' to verify.
```

Implementation:
1. Check `https://api.github.com/repos/fnune/kyaraben/releases/latest`
2. Find asset matching `kyaraben-cli-*-linux-amd64.tar.gz`
3. Download to temp directory
4. Extract and replace current binary (need write permission to binary location)
5. Optionally: `--check` flag to just check without downloading

Considerations:
- Binary might be in `/usr/local/bin` (needs sudo) or `~/.local/bin` (user-writable)
- Could use `KYARABEN_INSTALL_DIR` env var to hint where to install

## `kyaraben init --headless`

Add a `--headless` flag to the init command:

```
$ kyaraben init --headless --collection /mnt/storage/Emulation

Created configuration at ~/.config/kyaraben/config.toml

Headless mode: will sync all systems without installing emulators.
Run 'kyaraben apply' to set up synchronization.
```

Generated config:

```toml
[global]
headless = true
collection = "/mnt/storage/Emulation"

[sync]
enabled = true
```

No `[systems]` section needed - headless mode implies all systems.

## Implementation plan

### Phase 1: Manual pilot on bilbo

Before writing code, test the current CLI on bilbo to understand the gaps:

1. Download CLI tarball to bilbo
2. Create minimal config with sync enabled
3. Enable user lingering
4. Run `kyaraben apply`
5. Attempt to pair with desktop
6. Document what works and what fails

### Phase 2: Implement `headless` mode

1. Add `Headless bool` field to `GlobalConfig` in `internal/model/config.go`
2. Modify `internal/apply/apply.go` to skip emulator installation when headless
3. Modify `internal/sync/config.go` to use all systems (from registry) when headless
4. Update `internal/cli/init.go` to support `--headless` flag
5. Add tests

### Phase 3: Implement CLI self-update

1. Add `internal/cli/update.go` with update command
2. Port logic from `ui/electron/updater.ts` to Go
3. Handle binary replacement (temp download, atomic rename)
4. Add `--check` flag for checking without downloading
5. Add tests

### Phase 4: Documentation

1. Add "Headless server setup" guide to docs site
2. Include:
   - CLI-only installation
   - `kyaraben init --headless` usage
   - `loginctl enable-linger` for persistence
   - Pairing from server side
   - `kyaraben update` for keeping up to date
   - Firewall considerations

### Phase 5: Test and iterate

1. Full end-to-end test on bilbo
2. Pair with desktop and Steam Deck
3. Verify bidirectional sync works
4. Test update mechanism
5. Document any issues found

## Bilbo context

Bilbo is a Linux media server with:

- Storage: `/mnt/downloads-1t`, `/mnt/downloads-2t`, `/mnt/mirrored`
- User: `fausto`
- SSH-only access, no display
- Existing services: Jellyfin, Radarr, Sonarr, NZBGet, Immich

Proposed Kyaraben setup on bilbo:

```sh
# Download CLI
curl -L https://github.com/fnune/kyaraben/releases/latest/download/kyaraben-cli-linux-amd64.tar.gz | tar xz
mv kyaraben ~/.local/bin/

# Initialize headless config
kyaraben init --headless --collection /mnt/mirrored/Emulation

# Enable user lingering for persistent service
loginctl enable-linger fausto

# Apply (sets up Syncthing, creates folders)
kyaraben apply

# Pair with desktop
kyaraben sync pair
# -> displays pairing code
```

Generated config:

```toml
[global]
headless = true
collection = "/mnt/mirrored/Emulation"

[sync]
enabled = true
```

## Open questions

1. **Frontends in headless mode**: Should headless mode also create frontend folders (ES-DE gamelists/media)?
   - Probably yes - they're part of the sync
   - Use default frontend config (ES-DE enabled)

2. **Firewall ports**: Need documentation for:
   - 22100 (Syncthing listen)
   - 21127 (Syncthing discovery)
   - 8484 (Syncthing GUI, optional but useful for debugging)

3. **Device naming**: Should headless servers have a distinct name?
   - Currently uses `<hostname>-kyaraben`
   - Could use `<hostname>-hub` or keep it simple

4. **Binary installation location**:
   - If installed to `/usr/local/bin`, needs sudo to update
   - If installed to `~/.local/bin`, user-writable
   - Could prompt for sudo or suggest manual update

## Related files

- `internal/sync/config.go` - folder generation logic
- `internal/sync/systemd.go` - systemd service management
- `internal/model/sync.go` - SyncConfig struct
- `internal/cli/sync.go` - CLI commands
- `test/e2e/sync_test.go` - E2E tests
- `.goreleaser.yaml` - CLI release artifact
- `site/src/content/docs/using-the-cli.mdx` - CLI documentation
