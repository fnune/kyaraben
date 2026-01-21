package cli

import (
	"fmt"
	"os"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/store"
)

// DoctorCmd checks provision status.
type DoctorCmd struct{}

// Run executes the doctor command.
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

	checker := store.NewProvisionChecker(userStore)

	fmt.Println("Checking provisions...")
	fmt.Println()

	hasIssues := false
	requiredMissing := 0
	optionalMissing := 0

	for sys, sysConf := range cfg.Systems {
		emu, err := registry.GetEmulator(sysConf.Emulator)
		if err != nil {
			fmt.Printf("  Warning: unknown emulator %s\n", sysConf.Emulator)
			continue
		}

		results := checker.Check(emu, sys)

		if len(emu.Provisions) == 0 {
			fmt.Printf("  %s (%s)\n", emu.Name, sys)
			fmt.Println("    No provisions required.")
			fmt.Println()
			continue
		}

		fmt.Printf("  %s (%s)\n", emu.Name, sys)

		for _, result := range results {
			switch result.Status {
			case model.ProvisionFound:
				fmt.Printf("    %s %s - found, verified\n", checkMark(), result.Provision.Filename)
			case model.ProvisionMissing:
				if result.Provision.Required {
					fmt.Printf("    %s %s - MISSING (required)\n", crossMark(), result.Provision.Filename)
					hasIssues = true
					requiredMissing++
				} else {
					fmt.Printf("    %s %s - missing (optional)\n", crossMark(), result.Provision.Filename)
					optionalMissing++
				}
			case model.ProvisionInvalid:
				fmt.Printf("    %s %s - INVALID HASH\n", crossMark(), result.Provision.Filename)
				fmt.Printf("      Expected: %s\n", result.Provision.MD5Hash)
				fmt.Printf("      Got:      %s\n", result.ActualHash)
				if result.Provision.Required {
					hasIssues = true
					requiredMissing++
				}
			case model.ProvisionOptional:
				fmt.Printf("    %s %s - not found (optional)\n", dashMark(), result.Provision.Filename)
				optionalMissing++
			}
		}

		// Show where to place files
		biosDir := userStore.SystemBiosDir(sys)
		fmt.Printf("    Expected location: %s\n", biosDir)
		fmt.Println()
	}

	// Summary
	fmt.Println("Summary:")
	if requiredMissing > 0 {
		fmt.Printf("  %d required file(s) missing\n", requiredMissing)
	}
	if optionalMissing > 0 {
		fmt.Printf("  %d optional file(s) missing\n", optionalMissing)
	}
	if !hasIssues && optionalMissing == 0 {
		fmt.Println("  All provisions satisfied!")
	}

	if hasIssues {
		os.Exit(1)
	}
	return nil
}

func checkMark() string {
	return "\u2713" // ✓
}

func crossMark() string {
	return "\u2717" // ✗
}

func dashMark() string {
	return "-"
}
