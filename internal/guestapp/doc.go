// Package guestapp provides reusable components for device-side Kyaraben integrations.
//
// It defines interfaces and utilities that are shared across all CFW integrations,
// allowing each integration to focus on CFW-specific concerns like path defaults and UI.
//
// Components:
//   - UI interfaces (MenuUI, KeyboardUI, PresenterUI) for device interaction
//   - ServiceManager interface for controlling the sync service
//   - ProcessController interface with PIDProcessController for non-systemd devices
//   - Config types for service settings and path mappings
//
// The PIDProcessController uses Linux /proc to track processes via PID files.
// For systemd-based systems, use systemctl commands instead.
package guestapp
