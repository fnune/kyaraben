package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/fnune/kyaraben/internal/doctor"
	"github.com/fnune/kyaraben/internal/model"
)

type DoctorCmd struct{}

func (cmd *DoctorCmd) Run(ctx *Context) error {
	cfg, err := ctx.LoadConfig()
	if err != nil {
		return err
	}

	registry := ctx.NewRegistry()
	userStore, err := ctx.NewUserStore(cfg)
	if err != nil {
		return err
	}

	fmt.Println("Checking provisions...")
	fmt.Println()

	result, err := doctor.Run(context.Background(), cfg, registry, userStore)
	if err != nil {
		return err
	}

	for _, sys := range result.Systems {
		fmt.Printf("  %s (%s)\n", sys.EmulatorName, sys.SystemID)

		if len(sys.Provisions) == 0 {
			fmt.Println("    No provisions required.")
			fmt.Println()
			continue
		}

		for _, prov := range sys.Provisions {
			kindLabel := prov.DisplayName
			if prov.Description != "" {
				kindLabel = fmt.Sprintf("%s (%s)", kindLabel, prov.Description)
			}

			switch prov.Status {
			case model.ProvisionFound:
				verifiedAs := prov.DisplayName
				if prov.VerifiedDisplayName != "" {
					verifiedAs = prov.VerifiedDisplayName
				}
				fmt.Printf("    %s %s - verified (%s)\n", checkMark(), kindLabel, verifiedAs)
			case model.ProvisionMissing:
				if prov.GroupRequired && !prov.GroupSatisfied {
					fmt.Printf("    %s %s - MISSING\n", crossMark(), kindLabel)
					if prov.ImportViaUI {
						fmt.Printf("      Import %s via emulator\n", prov.DisplayName)
					} else if prov.Instructions != "" {
						fmt.Printf("      %s\n", prov.Instructions)
					}
				} else if prov.GroupRequired {
					fmt.Printf("    %s %s - not found (group satisfied)\n", dashMark(), kindLabel)
				} else {
					fmt.Printf("    %s %s - not found (optional)\n", dashMark(), kindLabel)
				}
			case model.ProvisionInvalid:
				fmt.Printf("    %s %s - INVALID HASH\n", crossMark(), kindLabel)
			}
		}
		fmt.Println()
	}

	fmt.Println("Summary:")
	if result.UnsatisfiedGroups > 0 {
		fmt.Printf("  %d provision group(s) unsatisfied\n", result.UnsatisfiedGroups)
	}
	if result.OptionalGroupsMissed > 0 {
		fmt.Printf("  %d optional group(s) not found\n", result.OptionalGroupsMissed)
	}
	if !result.HasIssues() && result.OptionalGroupsMissed == 0 {
		fmt.Println("  All provisions satisfied!")
	}

	if result.HasIssues() {
		os.Exit(1)
	}
	return nil
}

func checkMark() string {
	return "\u2713"
}

func crossMark() string {
	return "\u2717"
}

func dashMark() string {
	return "-"
}
