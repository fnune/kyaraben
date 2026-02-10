package model

// ProvisionKind categorizes what type of provision this is.
type ProvisionKind string

const (
	ProvisionBIOS     ProvisionKind = "bios"
	ProvisionKeys     ProvisionKind = "keys"
	ProvisionFirmware ProvisionKind = "firmware"
)

// ProvisionStatus represents the verification state of a provision.
type ProvisionStatus string

const (
	ProvisionFound   ProvisionStatus = "found"
	ProvisionMissing ProvisionStatus = "missing"
	ProvisionInvalid ProvisionStatus = "invalid" // File exists but hash mismatch
)

// Provision represents a file the user may need to provide for an emulator.
type Provision struct {
	Kind        ProvisionKind
	Filename    string   // Canonical filename (for display and lookup)
	Description string   // Short label like "USA", "ARM7", "IPL"
	Hashes      []string // Valid MD5 hashes for this file
	// FilePattern is a glob pattern (e.g., "*.nca"). When set, Filename is treated
	// as a directory and the provision is satisfied if any files match the pattern.
	FilePattern string
	// ImportViaUI indicates that this provision must be imported through the
	// emulator's settings UI rather than being placed in a folder.
	ImportViaUI bool
}

// ProvisionGroup represents a set of provisions with shared requirement semantics.
// For regional BIOS, MinRequired=1 means "at least one of these".
// For optional provisions like boot animations, MinRequired=0.
type ProvisionGroup struct {
	Provisions  []Provision
	MinRequired int    // 0 = optional, 1+ = at least N required
	Message     string // Shown when requirement unsatisfied
}

// ProvisionResult represents the outcome of checking a single provision.
type ProvisionResult struct {
	Provision  Provision
	Status     ProvisionStatus
	FoundPath  string // Path where file was found (if any)
	ActualHash string // Hash of found file (if any)
}

// ProvisionGroupResult represents the outcome of checking a provision group.
type ProvisionGroupResult struct {
	Group       ProvisionGroup
	Results     []ProvisionResult
	Satisfied   int  // How many provisions were found
	IsRequired  bool // Whether this group has requirements (MinRequired > 0)
	IsSatisfied bool // Whether MinRequired provisions were found
	BiosDir     string
}
