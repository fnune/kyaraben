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
	ProvisionFound    ProvisionStatus = "found"
	ProvisionMissing  ProvisionStatus = "missing"
	ProvisionInvalid  ProvisionStatus = "invalid" // File exists but hash mismatch
	ProvisionOptional ProvisionStatus = "optional"
)

// Provision represents something the user must provide for an emulator.
type Provision struct {
	ID       string
	Kind     ProvisionKind
	Filename string
	// Description is a short label for the provision variant, typically a region
	// like "USA", "Europe", or "Japan". Shown in the UI as "{Kind} ({Description})"
	// e.g. "BIOS (USA)" or "Firmware (Europe)".
	Description string
	Required    bool
	MD5Hash     string // Expected hash for verification
	SHA256Hash  string // Alternative hash
	// ImportViaUI indicates that this provision must be imported through the
	// emulator's settings UI rather than being placed in a folder. The emulator
	// stores the imported files in an internal location that Kyaraben checks.
	ImportViaUI bool
}

// ProvisionResult represents the outcome of checking a provision.
type ProvisionResult struct {
	Provision  Provision
	Status     ProvisionStatus
	FoundPath  string // Actual path where file was found
	ActualHash string // Hash of found file (if any)
}

func (pr *ProvisionResult) IsSatisfied() bool {
	return pr.Status == ProvisionFound || pr.Status == ProvisionOptional
}
