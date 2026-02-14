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
			label := prov.Filename
			if prov.Description != "" {
				label = fmt.Sprintf("%s (%s)", prov.Filename, prov.Description)
			}

			switch prov.Status {
			case model.ProvisionFound:
				fmt.Printf("    %s %s - found, verified\n", checkMark(), label)
			case model.ProvisionMissing:
				if prov.GroupRequired && !prov.GroupSatisfied {
					fmt.Printf("    %s %s - MISSING (%s)\n", crossMark(), label, prov.GroupMessage)
					if prov.Instructions != "" {
						fmt.Printf("      %s\n", prov.Instructions)
					}
				} else if prov.GroupRequired {
					fmt.Printf("    %s %s - not found (group satisfied)\n", dashMark(), label)
				} else {
					fmt.Printf("    %s %s - not found (optional)\n", dashMark(), label)
				}
			case model.ProvisionInvalid:
				fmt.Printf("    %s %s - INVALID HASH\n", crossMark(), label)
			}
		}

		fmt.Printf("    Expected location: %s\n", sys.BiosDir)
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
