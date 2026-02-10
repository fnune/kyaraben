package store

import (
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
		result := pc.checkGroup(group, biosDir, sys)
		if len(result.Results) > 0 {
			results = append(results, result)
		}
	}

	return results
}

func (pc *ProvisionChecker) checkGroup(group model.ProvisionGroup, biosDir string, sys model.SystemID) model.ProvisionGroupResult {
	result := model.ProvisionGroupResult{
		Group:      group,
		Results:    make([]model.ProvisionResult, 0, len(group.Provisions)),
		IsRequired: group.MinRequired > 0,
		BiosDir:    biosDir,
	}

	for _, prov := range group.Provisions {
		if !prov.AppliesToSystem(sys) {
			continue
		}
		checkResult := prov.Check(biosDir)
		provResult := model.ProvisionResult{
			Provision:  prov,
			Status:     checkResult.Status,
			FoundPath:  checkResult.FoundPath,
			ActualHash: checkResult.ActualHash,
		}
		result.Results = append(result.Results, provResult)
		if provResult.Status == model.ProvisionFound {
			result.Satisfied++
		}
	}

	result.IsSatisfied = result.Satisfied >= group.MinRequired

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
