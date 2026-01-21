package model

// SystemID uniquely identifies a gaming platform.
type SystemID string

const (
	SystemSNES  SystemID = "snes"
	SystemPSX   SystemID = "psx"
	SystemTIC80 SystemID = "tic80" // Fantasy console for E2E testing
)

// System represents a gaming platform that can be emulated.
type System struct {
	ID          SystemID
	Name        string
	Description string
}
