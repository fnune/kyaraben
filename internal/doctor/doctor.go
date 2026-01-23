package doctor

import (
	"github.com/fnune/kyaraben/internal/emulators"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/store"
)

type ProvisionResult struct {
	Filename     string
	Description  string
	Required     bool
	Status       model.ProvisionStatus
	FoundPath    string
	ActualHash   string
	ExpectedHash string
}

type SystemResult struct {
	SystemID     model.SystemID
	EmulatorID   model.EmulatorID
	EmulatorName string
	BiosDir      string
	Provisions   []ProvisionResult
}

type Result struct {
	Systems         []SystemResult
	RequiredMissing int
	OptionalMissing int
}

func Run(cfg *model.KyarabenConfig, registry *emulators.Registry) (*Result, error) {
	userStorePath, err := cfg.ExpandUserStore()
	if err != nil {
		return nil, err
	}

	userStore := store.NewUserStore(userStorePath)
	checker := store.NewProvisionChecker(userStore)

	result := &Result{}

	for sys, sysConf := range cfg.Systems {
		emu, err := registry.GetEmulator(sysConf.Emulator)
		if err != nil {
			continue
		}

		sysResult := SystemResult{
			SystemID:     sys,
			EmulatorID:   emu.ID,
			EmulatorName: emu.Name,
			BiosDir:      userStore.SystemBiosDir(sys),
		}

		provResults := checker.Check(emu, sys)
		for _, r := range provResults {
			pr := ProvisionResult{
				Filename:     r.Provision.Filename,
				Description:  r.Provision.Description,
				Required:     r.Provision.Required,
				Status:       r.Status,
				FoundPath:    r.FoundPath,
				ActualHash:   r.ActualHash,
				ExpectedHash: r.Provision.MD5Hash,
			}
			sysResult.Provisions = append(sysResult.Provisions, pr)

			switch r.Status {
			case model.ProvisionMissing:
				if r.Provision.Required {
					result.RequiredMissing++
				} else {
					result.OptionalMissing++
				}
			case model.ProvisionInvalid:
				if r.Provision.Required {
					result.RequiredMissing++
				}
			case model.ProvisionOptional:
				result.OptionalMissing++
			}
		}

		result.Systems = append(result.Systems, sysResult)
	}

	return result, nil
}

func (r *Result) HasIssues() bool {
	return r.RequiredMissing > 0
}
