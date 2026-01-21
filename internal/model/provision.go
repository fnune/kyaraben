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
	ID          string
	Kind        ProvisionKind
	Filename    string
	Description string
	Required    bool
	MD5Hash     string // Expected hash for verification
	SHA256Hash  string // Alternative hash
}

// ProvisionResult represents the outcome of checking a provision.
type ProvisionResult struct {
	Provision  Provision
	Status     ProvisionStatus
	FoundPath  string // Actual path where file was found
	ActualHash string // Hash of found file (if any)
}

// IsSatisfied returns true if the provision check passed.
func (pr *ProvisionResult) IsSatisfied() bool {
	return pr.Status == ProvisionFound || pr.Status == ProvisionOptional
}
