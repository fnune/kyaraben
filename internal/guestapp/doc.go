// Package guestapp provides reusable components for device-side Kyaraben integrations.
//
// It defines interfaces and utilities that are shared across all CFW integrations,
// allowing each integration to focus on CFW-specific concerns like path defaults and UI.
//
// # Components
//
//   - UI interfaces (MenuUI, KeyboardUI, PresenterUI) for device interaction
//   - ServiceManager interface for controlling the sync service
//   - SyncManager interface for sync operations (pairing, status, folder config)
//   - ProcessController interface with PIDProcessController for non-systemd devices
//   - Config types for service settings and path mappings
//   - GetLocalIP utility for displaying device address
//
// The PIDProcessController uses Linux /proc to track processes via PID files.
// For systemd-based systems, use systemctl commands instead.
//
// # Extraction candidates
//
// The following logic from integrations/nextui could be extracted here when
// building a second integration, to avoid duplication:
//
//   - App flow orchestration (~400 lines): main menu loop, toggle handlers,
//     status display. Would need a FolderMapper interface to abstract path mapping.
//
//   - Status calculation: getSyncStatus() logic that interprets SyncManager.GetStatus()
//     and produces UI-friendly status text and colors.
//
//   - Syncthing service lifecycle: start/stop/waitForReady/loadAPIKey logic that
//     manages a bundled syncthing process. Could move to syncguest package.
//
//   - CLI commands: start/stop/status/enable/disable/pair are universal operations
//     that every integration needs. The batocera integration implements these as CLI
//     commands, nextui implements them as menu actions. A shared CLI struct could
//     accept ServiceManager, SyncManager, and a FolderMapper interface, letting
//     integrations wire up their specific implementations. Integrations with graphical
//     UIs could offer both menu and CLI modes.
//
// These are documented here rather than extracted now to avoid premature abstraction.
// When building a second integration, compare with nextui to identify exact boundaries.
package guestapp
