package store

import (
	"crypto/md5"
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

func (pc *ProvisionChecker) Check(emu model.Emulator, sys model.SystemID) []model.ProvisionGroupResult {
	biosDir := pc.userStore.SystemBiosDir(sys)
	results := make([]model.ProvisionGroupResult, 0, len(emu.ProvisionGroups))

	for _, group := range emu.ProvisionGroups {
		result := pc.checkGroup(group, biosDir)
		results = append(results, result)
	}

	return results
}

func (pc *ProvisionChecker) checkGroup(group model.ProvisionGroup, biosDir string) model.ProvisionGroupResult {
	result := model.ProvisionGroupResult{
		Group:      group,
		Results:    make([]model.ProvisionResult, 0, len(group.Provisions)),
		IsRequired: group.MinRequired > 0,
		BiosDir:    biosDir,
	}

	for _, prov := range group.Provisions {
		provResult := pc.checkProvision(prov, biosDir)
		result.Results = append(result.Results, provResult)
		if provResult.Status == model.ProvisionFound {
			result.Satisfied++
		}
	}

	result.IsSatisfied = result.Satisfied >= group.MinRequired

	return result
}

func (pc *ProvisionChecker) checkProvision(prov model.Provision, biosDir string) model.ProvisionResult {
	result := model.ProvisionResult{
		Provision: prov,
		Status:    model.ProvisionMissing,
	}

	filePath := filepath.Join(biosDir, prov.Filename)

	// Case-insensitive filename matching
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		entries, _ := os.ReadDir(biosDir)
		for _, entry := range entries {
			if strings.EqualFold(entry.Name(), prov.Filename) {
				filePath = filepath.Join(biosDir, entry.Name())
				break
			}
		}
	}

	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		return result
	}

	result.FoundPath = filePath

	if len(prov.Hashes) == 0 {
		result.Status = model.ProvisionFound
		return result
	}

	hash, err := md5File(filePath)
	if err != nil {
		result.Status = model.ProvisionInvalid
		return result
	}

	result.ActualHash = hash

	for _, validHash := range prov.Hashes {
		if strings.EqualFold(hash, validHash) {
			result.Status = model.ProvisionFound
			return result
		}
	}

	result.Status = model.ProvisionInvalid
	return result
}

func HasUnsatisfiedRequired(results []model.ProvisionGroupResult) bool {
	for _, r := range results {
		if r.IsRequired && !r.IsSatisfied {
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
