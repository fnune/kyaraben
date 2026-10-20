package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fnune/kyaraben/internal/launcher"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/version"
)

type StatusCmd struct {
	JSON bool `help:"Print machine-readable JSON."`
}

func (cmd *StatusCmd) Run(ctx *Context) error {
	if cmd.JSON {
		return printStatusJSON(ctx)
	}

	configPath, err := ctx.GetConfigPath()
	if err != nil {
		return err
	}

	cfg, err := ctx.LoadConfig()
	if err != nil {
		fmt.Printf("Config: %s (not found or invalid)\n", configPath)
		fmt.Println()
		fmt.Println("Run 'kyaraben init' to create a configuration.")
		return nil
	}

	registry := ctx.NewRegistry()
	collection, err := ctx.NewCollection(cfg)
	if err != nil {
		return err
	}
	manifestPath, err := ctx.GetPaths().ManifestPath()
	if err != nil {
		return err
	}

	result, err := ctx.NewStatusGetter().Get(context.Background(), cfg, configPath, registry, collection, manifestPath)
	if err != nil {
		return err
	}

	fmt.Printf("Config: %s\n", result.ConfigPath)
	fmt.Printf("Collection: %s", result.CollectionPath)
	if result.CollectionInitialized {
		fmt.Println(" (initialized)")
	} else {
		fmt.Println(" (not initialized)")
	}
	fmt.Println()

	if len(result.EnabledSystems) == 0 {
		fmt.Println("Enabled systems: none")
	} else {
		names := make([]string, len(result.EnabledSystems))
		for i, sys := range result.EnabledSystems {
			names[i] = sys.Name
		}
		fmt.Printf("Enabled systems: %s\n", strings.Join(names, ", "))
	}
	fmt.Println()

	if len(result.InstalledEmulators) == 0 {
		fmt.Println("Managed emulators: none")
		fmt.Println()
		fmt.Println("Run 'kyaraben apply' to install emulators.")
	} else {
		fmt.Println("Managed emulators:")
		for _, emu := range result.InstalledEmulators {
			versionInfo := emu.Version
			if emu.PinnedVersion != "" {
				versionInfo += " (pinned)"
			} else if emu.DefaultVersion != "" && emu.Version != emu.DefaultVersion {
				versionInfo += fmt.Sprintf(" (update to %s on apply)", emu.DefaultVersion)
			}
			fmt.Printf("  %-20s %s\n", emu.Name, versionInfo)
		}
		fmt.Println()

		fmt.Println("Paths:")
		for _, emu := range result.InstalledEmulators {
			emuDef, err := registry.GetEmulator(emu.ID)
			if err != nil {
				continue
			}
			for sysID, emuIDs := range cfg.Systems {
				for _, id := range emuIDs {
					if id != emu.ID {
						continue
					}
					fmt.Printf("  %s (%s)\n", emu.Name, sysID)
					fmt.Printf("    ROMs:          %s\n", collection.SystemRomsDir(sysID))
					if emuDef.PathUsage.UsesBiosDir {
						fmt.Printf("    BIOS:          %s\n", collection.SystemBiosDir(sysID))
					}
					if emuDef.PathUsage.UsesSavesDir {
						fmt.Printf("    Saves:         %s\n", collection.SystemSavesDir(sysID))
					}
					if emuDef.PathUsage.UsesStatesDir {
						fmt.Printf("    Savestates:    %s\n", collection.EmulatorStatesDir(emu.ID))
					}
					if emuDef.PathUsage.UsesScreenshotsDir {
						fmt.Printf("    Screenshots:   %s\n", collection.EmulatorScreenshotsDir(emu.ID))
					}
					break
				}
			}
		}
		fmt.Println()

		if !result.LastApplied.IsZero() {
			fmt.Printf("Last applied: %s\n", result.LastApplied.Format("2006-01-02 15:04:05"))
		}
	}
	fmt.Println()

	if result.MissingRequiredCount > 0 {
		fmt.Printf("Provision status: %d system(s) missing required files (run 'kyaraben doctor')\n", result.MissingRequiredCount)
	} else if len(result.EnabledSystems) > 0 {
		fmt.Println("Provision status: all required files present")
	}

	return nil
}

type statusJSON struct {
	Version     string               `json:"version"`
	ConfigPath  string               `json:"configPath"`
	Collection  string               `json:"collection"`
	BinDir      string               `json:"binDir"`
	CoresDir    string               `json:"coresDir"`
	LastApplied *time.Time           `json:"lastApplied,omitempty"`
	Systems     []systemStatusJSON   `json:"systems"`
	Emulators   []emulatorStatusJSON `json:"emulators"`
}

type systemStatusJSON struct {
	ID         model.SystemID     `json:"id"`
	Name       string             `json:"name"`
	Extensions []string           `json:"extensions"`
	Emulators  []model.EmulatorID `json:"emulators"`
	Roms       string             `json:"roms"`
	Bios       string             `json:"bios"`
	Saves      string             `json:"saves"`
}

type emulatorStatusJSON struct {
	ID      model.EmulatorID `json:"id"`
	Name    string           `json:"name"`
	Version string           `json:"version"`
	Binary  string           `json:"binary"`
	Core    string           `json:"core,omitempty"`
	States  string           `json:"states"`
	// RomCommand is a shell command line with a %ROM% placeholder, as ES-DE uses.
	RomCommand string `json:"romCommand,omitempty"`
}

func printStatusJSON(ctx *Context) error {
	configPath, err := ctx.GetConfigPath()
	if err != nil {
		return err
	}
	cfg, err := ctx.LoadConfig()
	if err != nil {
		return err
	}
	registry := ctx.NewRegistry()
	collection, err := ctx.NewCollection(cfg)
	if err != nil {
		return err
	}
	manifestPath, err := ctx.GetPaths().ManifestPath()
	if err != nil {
		return err
	}
	result, err := ctx.NewStatusGetter().Get(context.Background(), cfg, configPath, registry, collection, manifestPath)
	if err != nil {
		return err
	}
	lm, err := launcher.NewManager(ctx.FS, ctx.Paths, ctx.Resolver)
	if err != nil {
		return err
	}

	out := statusJSON{
		Version:    version.Get(),
		ConfigPath: configPath,
		Collection: collection.Root(),
		BinDir:     lm.BinDir(),
		CoresDir:   lm.CoresDir(),
		Systems:    []systemStatusJSON{},
		Emulators:  []emulatorStatusJSON{},
	}
	if !result.LastApplied.IsZero() {
		out.LastApplied = &result.LastApplied
	}

	systemIDs := make([]model.SystemID, 0, len(cfg.Systems))
	for id := range cfg.Systems {
		systemIDs = append(systemIDs, id)
	}
	sort.Slice(systemIDs, func(i, j int) bool { return systemIDs[i] < systemIDs[j] })
	for _, id := range systemIDs {
		sys, err := registry.GetSystem(id)
		if err != nil {
			continue
		}
		out.Systems = append(out.Systems, systemStatusJSON{
			ID:         id,
			Name:       sys.Name,
			Extensions: sys.Extensions,
			Emulators:  cfg.Systems[id],
			Roms:       collection.SystemRomsDir(id),
			Bios:       collection.SystemBiosDir(id),
			Saves:      collection.SystemSavesDir(id),
		})
	}

	for _, installed := range result.InstalledEmulators {
		emu, err := registry.GetEmulator(installed.ID)
		if err != nil {
			continue
		}
		binary := filepath.Join(lm.BinDir(), emu.Launcher.Binary)
		e := emulatorStatusJSON{
			ID:      emu.ID,
			Name:    emu.Name,
			Version: installed.Version,
			Binary:  binary,
			States:  collection.EmulatorStatesDir(emu.ID),
		}
		if emu.Launcher.CoreName != "" {
			e.Core = filepath.Join(lm.CoresDir(), emu.Launcher.CoreName+".so")
		}
		if emu.Launcher.RomCommand != nil {
			e.RomCommand = emu.Launcher.RomCommand(model.RomLaunchOptions{BinaryPath: binary, Fullscreen: true})
		}
		out.Emulators = append(out.Emulators, e)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
