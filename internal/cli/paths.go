package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/fnune/kyaraben/internal/model"
)

func shortenPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	if strings.HasPrefix(path, home+"/") {
		return "~" + path[len(home):]
	}
	return path
}

type PathsCmd struct{}

func (cmd *PathsCmd) Run(ctx *Context) error {
	cfg, err := ctx.LoadConfig()
	if err != nil {
		return err
	}

	registry := ctx.NewRegistry()
	userStore, err := ctx.NewUserStore(cfg)
	if err != nil {
		return err
	}

	manifestPath, err := model.DefaultManifestPath()
	if err != nil {
		return err
	}

	manifest, err := model.LoadManifest(manifestPath)
	if err != nil {
		return err
	}

	if len(manifest.InstalledEmulators) == 0 {
		fmt.Println("No emulators installed. Run 'kyaraben apply' first.")
		return nil
	}

	fmt.Println("Paths for installed emulators:")
	fmt.Println()

	for emuID := range manifest.InstalledEmulators {
		emu, err := registry.GetEmulator(emuID)
		if err != nil {
			continue
		}

		fmt.Printf("  %s\n", emu.Name)

		for sysID, emuIDs := range cfg.Systems {
			for _, id := range emuIDs {
				if id != emuID {
					continue
				}

				fmt.Printf("    ROMs:        %s\n", shortenPath(userStore.SystemRomsDir(sysID)))

				if emu.PathUsage.UsesBiosDir {
					fmt.Printf("    BIOS:        %s\n", shortenPath(userStore.SystemBiosDir(sysID)))
				}
				if emu.PathUsage.UsesSavesDir {
					fmt.Printf("    Saves:       %s\n", shortenPath(userStore.SystemSavesDir(sysID)))
				}
				if emu.PathUsage.UsesStatesDir {
					fmt.Printf("    Savestates:  %s\n", shortenPath(userStore.EmulatorStatesDir(emuID)))
				}
				if emu.PathUsage.UsesScreenshotsDir {
					fmt.Printf("    Screenshots: %s\n", shortenPath(userStore.EmulatorScreenshotsDir(emuID)))
				}
				if emu.PathUsage.OpaqueContents != "" {
					fmt.Printf("    Emulator data: %s\n", shortenPath(userStore.EmulatorOpaqueDir(emuID)))
					fmt.Printf("      Contains: %s\n", emu.PathUsage.OpaqueContents)
				}

				break
			}
		}

		fmt.Println()
	}

	return nil
}
