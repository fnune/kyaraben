package model

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/twpayne/go-vfs/v5"
)

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

// CheckResult contains the outcome of checking a provision.
type CheckResult struct {
	Status     ProvisionStatus
	FoundPath  string
	ActualHash string
}

// UIHints tells the frontend how to display this provision.
type UIHints struct {
	DisplayName         string
	VerifiedDisplayName string // Optional: shown instead of DisplayName when found
	Instructions        string
}

// ProvisionStrategy defines how a provision is validated and displayed.
type ProvisionStrategy interface {
	Check(fs vfs.FS, biosDir string) CheckResult
	Hints() UIHints
}

// DownloadSource describes how to auto-download a provision.
type DownloadSource struct {
	URL               string // Download URL
	SHA256            string // SRI format "sha256-..." or hex
	ArchiveType       string // "zip", "tar.gz", etc. Empty if not archived
	FilenameInArchive string // Path inside archive; empty to use all files
}

// FileStrategy checks that a file exists by name (no hash verification).
type FileStrategy struct {
	Filename string
}

func (s FileStrategy) Check(fs vfs.FS, biosDir string) CheckResult {
	result := CheckResult{Status: ProvisionMissing}
	filePath := findFile(fs, biosDir, s.Filename)
	if filePath == "" {
		return result
	}
	info, err := fs.Stat(filePath)
	if err != nil || info.IsDir() {
		return result
	}
	result.Status = ProvisionFound
	result.FoundPath = filePath
	return result
}

func (s FileStrategy) Hints() UIHints {
	return UIHints{
		DisplayName:  s.Filename,
		Instructions: fmt.Sprintf("Place %s in this directory", s.Filename),
	}
}

// HashedStrategy checks file exists with valid MD5 hash.
type HashedStrategy struct {
	Filename string
	Hashes   []string
}

func (s HashedStrategy) Check(fs vfs.FS, biosDir string) CheckResult {
	result := CheckResult{Status: ProvisionMissing}
	filePath := findFile(fs, biosDir, s.Filename)
	if filePath == "" {
		return result
	}
	info, err := fs.Stat(filePath)
	if err != nil || info.IsDir() {
		return result
	}
	result.FoundPath = filePath

	hash, err := md5File(fs, filePath)
	if err != nil {
		result.Status = ProvisionInvalid
		return result
	}
	result.ActualHash = hash

	for _, validHash := range s.Hashes {
		if strings.EqualFold(hash, validHash) {
			result.Status = ProvisionFound
			return result
		}
	}
	result.Status = ProvisionInvalid
	return result
}

func (s HashedStrategy) Hints() UIHints {
	return UIHints{
		DisplayName:  s.Filename,
		Instructions: fmt.Sprintf("Place %s in this directory", s.Filename),
	}
}

// PatternStrategy checks directory contains files matching glob pattern.
type PatternStrategy struct {
	Pattern     string
	Description string
}

func (s PatternStrategy) Check(fs vfs.FS, biosDir string) CheckResult {
	result := CheckResult{Status: ProvisionMissing}
	matches := globDir(fs, biosDir, s.Pattern)
	if len(matches) > 0 {
		result.Status = ProvisionFound
		result.FoundPath = biosDir
	}
	return result
}

func (s PatternStrategy) Hints() UIHints {
	return UIHints{
		DisplayName:  s.Pattern,
		Instructions: fmt.Sprintf("Place %s files in this directory", s.Description),
	}
}

// ImportStrategy checks for files created by importing via emulator UI.
// Use this when a provision requires UI import and we verify by checking
// for the extracted/installed result rather than the source file.
type ImportStrategy struct {
	Pattern             string // Glob pattern to check (e.g., "dev_flash/sys/*")
	Filename            string // Source filename user needs to import (e.g., "PS3UPDAT.PUP")
	VerifiedDescription string // Shown when found (e.g., "firmware installed")
	Instructions        string // How to import (e.g., "Import via File > Install Firmware")
}

func (s ImportStrategy) Check(fs vfs.FS, baseDir string) CheckResult {
	result := CheckResult{Status: ProvisionMissing}
	matches := globDir(fs, baseDir, s.Pattern)
	if len(matches) > 0 {
		result.Status = ProvisionFound
		result.FoundPath = baseDir
	}
	return result
}

func (s ImportStrategy) Hints() UIHints {
	return UIHints{
		DisplayName:         s.Filename,
		VerifiedDisplayName: s.VerifiedDescription,
		Instructions:        s.Instructions,
	}
}

// Provision represents a file the user may need to provide for an emulator.
type Provision struct {
	Kind        ProvisionKind
	Description string
	Strategy    ProvisionStrategy
	ImportViaUI bool
	Systems     []SystemID      // If non-empty, provision only applies to these systems
	Download    *DownloadSource // If set, provision can be auto-downloaded
}

func (p Provision) Check(fs vfs.FS, biosDir string) CheckResult {
	return p.Strategy.Check(fs, biosDir)
}

func (p Provision) Hints() UIHints {
	return p.Strategy.Hints()
}

// FileProvision creates a provision that checks for a file without hash verification.
func FileProvision(kind ProvisionKind, filename, description string) Provision {
	return Provision{
		Kind:        kind,
		Description: description,
		Strategy:    FileStrategy{Filename: filename},
	}
}

// HashedProvision creates a provision that verifies a file's MD5 hash.
func HashedProvision(kind ProvisionKind, filename, description string, hashes []string) Provision {
	return Provision{
		Kind:        kind,
		Description: description,
		Strategy:    HashedStrategy{Filename: filename, Hashes: hashes},
	}
}

// PatternProvision creates a provision that matches files via glob pattern.
func PatternProvision(kind ProvisionKind, pattern, patternDesc, description string) Provision {
	return Provision{
		Kind:        kind,
		Description: description,
		Strategy:    PatternStrategy{Pattern: pattern, Description: patternDesc},
	}
}

// WithImportViaUI returns a copy with ImportViaUI set to true.
func (p Provision) WithImportViaUI() Provision {
	p.ImportViaUI = true
	return p
}

// ForSystems returns a copy that only applies to the specified systems.
func (p Provision) ForSystems(systems ...SystemID) Provision {
	p.Systems = systems
	return p
}

// WithDownload returns a copy with an auto-download source configured.
func (p Provision) WithDownload(src DownloadSource) Provision {
	p.Download = &src
	return p
}

// CanDownload reports whether this provision has an auto-download source.
func (p Provision) CanDownload() bool {
	return p.Download != nil && p.Download.URL != ""
}

// AppliesToSystem returns true if the provision applies to the given system.
func (p Provision) AppliesToSystem(sys SystemID) bool {
	if len(p.Systems) == 0 {
		return true
	}
	for _, s := range p.Systems {
		if s == sys {
			return true
		}
	}
	return false
}

// ProvisionGroup represents a set of provisions with shared requirement semantics.
// For regional BIOS, MinRequired=1 means "at least one of these".
// For optional provisions like boot animations, MinRequired=0.
type ProvisionGroup struct {
	Provisions  []Provision
	MinRequired int    // 0 = optional, 1+ = at least N required
	Message     string // Shown when requirement unsatisfied
	BaseDir     ProvisionBaseDirFunc
}

// BaseDirFor evaluates the configured base directory or defaults to the BIOS dir.
func (g ProvisionGroup) BaseDirFor(store StoreReader, sys SystemID) string {
	if g.BaseDir != nil {
		return g.BaseDir(store, sys)
	}
	return store.SystemBiosDir(sys)
}

// ProvisionBaseDirFunc resolves the base directory that a provision group
// should scan for files. If nil, the default is the system BIOS directory.
type ProvisionBaseDirFunc func(StoreReader, SystemID) string

// ProvisionResult represents the outcome of checking a single provision.
type ProvisionResult struct {
	Provision  Provision
	Status     ProvisionStatus
	FoundPath  string
	ActualHash string
}

// ProvisionGroupResult represents the outcome of checking a provision group.
type ProvisionGroupResult struct {
	Group       ProvisionGroup
	Results     []ProvisionResult
	Satisfied   int
	IsRequired  bool
	IsSatisfied bool
	BaseDir     string
}

func globDir(fs vfs.FS, dir, pattern string) []string {
	entries, err := fs.ReadDir(dir)
	if err != nil {
		return nil
	}
	var matches []string
	for _, entry := range entries {
		if matched, _ := filepath.Match(pattern, entry.Name()); matched {
			matches = append(matches, filepath.Join(dir, entry.Name()))
		}
	}
	return matches
}

func findFile(fs vfs.FS, biosDir, filename string) string {
	filePath := filepath.Join(biosDir, filename)
	if _, err := fs.Stat(filePath); err == nil {
		return filePath
	}
	entries, _ := fs.ReadDir(biosDir)
	for _, entry := range entries {
		if strings.EqualFold(entry.Name(), filename) {
			return filepath.Join(biosDir, entry.Name())
		}
	}
	return ""
}

func md5File(fs vfs.FS, path string) (string, error) {
	f, err := fs.Open(path)
	if err != nil {
		return "", fmt.Errorf("opening file: %w", err)
	}
	defer func() { _ = f.Close() }()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("hashing file: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
