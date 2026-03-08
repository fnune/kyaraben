# Agent instructions

Go + TypeScript (Electron/React) project. Uses `just` as the task runner.

```bash
just --list          # See all available commands
just ensure          # Install dependencies
just dev             # Run app in development mode
just check           # Lint, typecheck, test
just build           # Build AppImage (run check and build before committing)
just build-cli       # Build CLI binary only (for deployment to other devices)
just fmt             # Format all code
just generate-types  # Regenerate TS types from Go (after changing internal/daemon/types.go)
just site-dev        # Run documentation site locally
```

## Testing

```bash
just test                                    # Go unit tests
cd ui && npm test                            # Vitest unit tests
just build && just ui-e2e                    # Playwright E2E (full suite)
just ui-e2e tests/sync.spec.ts               # Run specific test file
just ui-e2e --grep "shows conflict"          # Filter by test name
just ui-e2e tests/sync.spec.ts --grep "conflict"  # Both
```

## Debugging sync on live devices

SSH aliases: `steamdeck`, `feanor`, `bilbo`

```bash
cat ~/.local/state/kyaraben/kyaraben.log     # Daemon logs
systemctl --user status kyaraben-syncthing   # Syncthing service state
cat ~/.local/state/kyaraben/syncthing/config/config.xml  # Syncthing config
```

Syncthing web UI: `http://<device-ip>:8484`

Key files:
- `config.xml`: devices, folders, options (merged on apply, not overwritten)
- `.apikey`: API key for REST calls
- Ports: TCP 22100 (sync), UDP 21127 (discovery), TCP 8484 (GUI)

## Code style

See `site/src/content/docs/contributing.mdx` for full guidelines. Key points:

- Use clean dependency injection following existing patterns in the codebase
- No "what" comments; make code self-evident instead
- "Why" comments are acceptable when the reasoning isn't obvious

## Skills

Detailed guides for complex tasks live in `.claude/skills/`:

- `adding-emulator-support`: adding support for a new emulator or system
