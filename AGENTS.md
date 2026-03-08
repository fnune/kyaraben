# Agent instructions

Go + TypeScript (Electron/React) project. Uses `just` as the task runner.

```bash
just --list          # See all available commands
just ensure          # Install dependencies
just dev             # Run app in development mode
just check           # Lint, typecheck, test
just build           # Build AppImage (run check and build before committing)
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

## Skills

Detailed guides for complex tasks live in `.claude/skills/`:

- `adding-emulator-support`: adding support for a new emulator or system
