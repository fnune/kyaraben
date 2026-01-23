package store

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fnune/kyaraben/internal/model"
)

type ProvisionChecker struct {
	userStore *UserStore
}

func NewProvisionChecker(userStore *UserStore) *ProvisionChecker {
	return &ProvisionChecker{userStore: userStore}
}

func (pc *ProvisionChecker) Check(emu model.Emulator, sys model.SystemID) []model.ProvisionResult {
	results := make([]model.ProvisionResult, 0, len(emu.Provisions))

	biosDir := pc.userStore.SystemBiosDir(sys)

	for _, prov := range emu.Provisions {
		result := pc.checkProvision(prov, biosDir)
		results = append(results, result)
	}

	return results
}

func (pc *ProvisionChecker) checkProvision(prov model.Provision, biosDir string) model.ProvisionResult {
	result := model.ProvisionResult{
		Provision: prov,
		Status:    model.ProvisionMissing,
	}

	// Look for the file in the bios directory
	filePath := filepath.Join(biosDir, prov.Filename)

	// Also check case-insensitive (common issue with BIOS files)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Try to find case-insensitive match
		entries, _ := os.ReadDir(biosDir)
		for _, entry := range entries {
			if strings.EqualFold(entry.Name(), prov.Filename) {
				filePath = filepath.Join(biosDir, entry.Name())
				break
			}
		}
	}

	// Check if file exists
	info, err := os.Stat(filePath)
	if err != nil {
		if !prov.Required {
			result.Status = model.ProvisionOptional
		}
		return result
	}

	if info.IsDir() {
		return result
	}

	result.FoundPath = filePath

	// Verify hash if we have one (prefer SHA256 over MD5)
	if prov.SHA256Hash != "" {
		hash, err := sha256File(filePath)
		if err != nil {
			result.Status = model.ProvisionInvalid
			return result
		}

		result.ActualHash = hash
		if strings.EqualFold(hash, prov.SHA256Hash) {
			result.Status = model.ProvisionFound
		} else {
			result.Status = model.ProvisionInvalid
		}
		return result
	}

	if prov.MD5Hash != "" {
		hash, err := md5File(filePath)
		if err != nil {
			result.Status = model.ProvisionInvalid
			return result
		}

		result.ActualHash = hash
		if strings.EqualFold(hash, prov.MD5Hash) {
			result.Status = model.ProvisionFound
		} else {
			result.Status = model.ProvisionInvalid
		}
		return result
	}

	// No hash to verify, just check existence
	result.Status = model.ProvisionFound
	return result
}

func (pc *ProvisionChecker) CheckAll(emulators map[model.SystemID]model.Emulator) map[model.EmulatorID][]model.ProvisionResult {
	results := make(map[model.EmulatorID][]model.ProvisionResult)

	for sys, emu := range emulators {
		results[emu.ID] = pc.Check(emu, sys)
	}

	return results
}

func HasMissingRequired(results []model.ProvisionResult) bool {
	for _, r := range results {
		if r.Provision.Required && r.Status == model.ProvisionMissing {
			return true
		}
	}
	return false
}

func md5File(path string) (string, error) {
	f, err := os.Open(path)
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

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("opening file: %w", err)
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("hashing file: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
