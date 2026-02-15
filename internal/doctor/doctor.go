package doctor

import (
	"context"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/registry"
	"github.com/fnune/kyaraben/internal/store"
)

type ProvisionResult struct {
	Filename            string
	Kind                model.ProvisionKind
	Description         string
	Status              model.ProvisionStatus
	ExpectedPath        string
	FoundPath           string
	ImportViaUI         bool
	GroupMessage        string
	GroupRequired       bool
	GroupSatisfied      bool
	GroupSize           int
	DisplayName         string
	VerifiedDisplayName string
	Instructions        string
}

type SystemResult struct {
	SystemID     model.SystemID
	EmulatorID   model.EmulatorID
	EmulatorName string
	BiosDir      string
	Provisions   []ProvisionResult
}

type Result struct {
	Systems              []SystemResult
	UnsatisfiedGroups    int
	OptionalGroupsMissed int
}

func Run(ctx context.Context, cfg *model.KyarabenConfig, reg *registry.Registry, userStore *store.UserStore) (*Result, error) {
	checker := store.NewProvisionChecker(userStore)

	result := &Result{}

	for sys, emulatorIDs := range cfg.Systems {
		for _, emuID := range emulatorIDs {
			emu, err := reg.GetEmulator(emuID)
			if err != nil {
				continue
			}

			sysResult := SystemResult{
				SystemID:     sys,
				EmulatorID:   emu.ID,
				EmulatorName: emu.Name,
				BiosDir:      userStore.SystemBiosDir(sys),
			}

			groupResults := checker.Check(emu, sys)
			for _, gr := range groupResults {
				for _, pr := range gr.Results {
					hints := pr.Provision.Hints()
					expectedPath := gr.BaseDir
					provResult := ProvisionResult{
						Filename:            hints.DisplayName,
						Kind:                pr.Provision.Kind,
						Description:         pr.Provision.Description,
						Status:              pr.Status,
						ExpectedPath:        expectedPath,
						FoundPath:           pr.FoundPath,
						ImportViaUI:         pr.Provision.ImportViaUI,
						GroupMessage:        gr.Group.Message,
						GroupRequired:       gr.IsRequired,
						GroupSatisfied:      gr.IsSatisfied,
						GroupSize:           len(gr.Group.Provisions),
						DisplayName:         hints.DisplayName,
						VerifiedDisplayName: hints.VerifiedDisplayName,
						Instructions:        hints.Instructions,
					}
					sysResult.Provisions = append(sysResult.Provisions, provResult)
				}

				if gr.IsRequired && !gr.IsSatisfied {
					result.UnsatisfiedGroups++
				} else if !gr.IsRequired && gr.Satisfied == 0 {
					result.OptionalGroupsMissed++
				}
			}

			result.Systems = append(result.Systems, sysResult)
		}
	}

	return result, nil
}

func (r *Result) HasIssues() bool {
	return r.UnsatisfiedGroups > 0
}
