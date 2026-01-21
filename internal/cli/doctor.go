package cli

import (
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

	result, err := doctor.Run(cfg, registry, userStore)
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
			switch prov.Status {
			case model.ProvisionFound:
				fmt.Printf("    %s %s - found, verified\n", checkMark(), prov.Filename)
			case model.ProvisionMissing:
				if prov.Required {
					fmt.Printf("    %s %s - MISSING (required)\n", crossMark(), prov.Filename)
				} else {
					fmt.Printf("    %s %s - missing (optional)\n", crossMark(), prov.Filename)
				}
			case model.ProvisionInvalid:
				fmt.Printf("    %s %s - INVALID HASH\n", crossMark(), prov.Filename)
				fmt.Printf("      Expected: %s\n", prov.ExpectedHash)
				fmt.Printf("      Got:      %s\n", prov.ActualHash)
			case model.ProvisionOptional:
				fmt.Printf("    %s %s - not found (optional)\n", dashMark(), prov.Filename)
			}
		}

		fmt.Printf("    Expected location: %s\n", sys.BiosDir)
		fmt.Println()
	}

	fmt.Println("Summary:")
	if result.RequiredMissing > 0 {
		fmt.Printf("  %d required file(s) missing\n", result.RequiredMissing)
	}
	if result.OptionalMissing > 0 {
		fmt.Printf("  %d optional file(s) missing\n", result.OptionalMissing)
	}
	if !result.HasIssues() && result.OptionalMissing == 0 {
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
