// BIOS hash data compiled from:
// - EmuDeck (https://github.com/dragoonDorise/EmuDeck) - GPL-3
// - RetroDECK (https://github.com/XargonWan/RetroDECK) - GPL-3
// - Libretro documentation (https://docs.libretro.com)
package dolphin

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:              model.EmulatorIDDolphin,
		Name:            "Dolphin",
		Systems:         []model.SystemID{model.SystemIDGameCube, model.SystemIDWii},
		Package:         model.AppImageRef("dolphin"),
		ProvisionGroups: buildDolphinProvisionGroups(),
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "dolphin",
			GenericName: "GameCube/Wii Emulator",
			Categories:  []string{"Game", "Emulator"},
			RomCommand: func(opts model.RomLaunchOptions) string {
				cmd := opts.BinaryPath
				if len(opts.LaunchArgs) > 0 {
					cmd += " " + strings.Join(opts.LaunchArgs, " ")
				}
				if opts.Fullscreen {
					cmd += " -C Dolphin.Display.Fullscreen=True"
				}
				cmd += " -e %ROM%"
				return cmd
			},
		},
		PathUsage: model.PathUsage{
			UsesSavesDir:       true,
			UsesStatesDir:      true,
			UsesScreenshotsDir: true,
		},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

type Config struct{}

var configTarget = model.ConfigTarget{
	RelPath: "dolphin-emu/Dolphin.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

func (c *Config) Generate(ctx model.GenerateContext) (model.GenerateResult, error) {
	store := ctx.Store

	patches := []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"General", "ISOPath0"}, Value: store.SystemRomsDir(model.SystemIDGameCube)},
			{Path: []string{"General", "ISOPath1"}, Value: store.SystemRomsDir(model.SystemIDWii)},
			{Path: []string{"General", "ISOPaths"}, Value: "2"},
			{Path: []string{"General", "DumpPath"}, Value: store.EmulatorScreenshotsDir(model.EmulatorIDDolphin)},
			{Path: []string{"GBA", "BIOS"}, Value: store.SystemBiosDir(model.SystemIDGBA) + "/gba_bios.bin"},
			{Path: []string{"GBA", "SavesPath"}, Value: store.SystemSavesDir(model.SystemIDGBA)},
			{Path: []string{"GBA", "SavesInRomPath"}, Value: "0"},
		},
	}}

	dataDir, err := ctx.BaseDirResolver.UserDataDir()
	if err != nil {
		return model.GenerateResult{}, err
	}
	dolphinDir := filepath.Join(dataDir, "dolphin-emu")

	symlinks := []model.SymlinkSpec{
		{Source: filepath.Join(dolphinDir, "GC"), Target: store.SystemSavesDir(model.SystemIDGameCube)},
		{Source: filepath.Join(dolphinDir, "Wii"), Target: store.SystemSavesDir(model.SystemIDWii)},
		{Source: filepath.Join(dolphinDir, "StateSaves"), Target: store.EmulatorStatesDir(model.EmulatorIDDolphin)},
		{Source: filepath.Join(dolphinDir, "ScreenShots"), Target: store.EmulatorScreenshotsDir(model.EmulatorIDDolphin)},
	}

	return model.GenerateResult{
		Patches:  patches,
		Symlinks: symlinks,
	}, nil
}

type iplRegionSpec struct {
	Name   string
	Dir    string
	File   string
	Hashes []string
}

var iplRegionSpecs = []iplRegionSpec{
	{
		Name:   "USA",
		Dir:    "USA",
		File:   "IPL.bin",
		Hashes: []string{"6dac1f2a14f659a1a7fbf749892b4e41", "019e39822a9ca3029124f74dd4d55ac4"},
	},
	{
		Name:   "Europe",
		Dir:    "EUR",
		File:   "IPL.bin",
		Hashes: []string{"db92574caab77a7ec99d4605fd6f2450", "0cdda509e2da83c85bfe423dd87346cc"},
	},
	{
		Name:   "Japan",
		Dir:    "JAP",
		File:   "IPL.bin",
		Hashes: []string{"fc924a7c879b661abc37cec4f018fdf3", "81df278301dc7bdf57bb760d7393ab4d"},
	},
}

func buildDolphinProvisionGroups() []model.ProvisionGroup {
	groups := make([]model.ProvisionGroup, 0, len(iplRegionSpecs)+1)
	for _, region := range iplRegionSpecs {
		region := region
		groups = append(groups, model.ProvisionGroup{
			MinRequired: 0,
			Message:     fmt.Sprintf("GameCube IPL (%s)", region.Name),
			BaseDir: func(store model.StoreReader, sys model.SystemID) string {
				return filepath.Join(store.SystemSavesDir(model.SystemIDGameCube), region.Dir)
			},
			Provisions: []model.Provision{
				model.HashedProvision(model.ProvisionBIOS, region.File, region.Name, region.Hashes).ForSystems(model.SystemIDGameCube),
			},
		})
	}

	groups = append(groups, model.ProvisionGroup{
		MinRequired: 0,
		Message:     "Game Boy Advance BIOS (optional, shared with mGBA)",
		BaseDir: func(store model.StoreReader, sys model.SystemID) string {
			return store.SystemBiosDir(model.SystemIDGBA)
		},
		Provisions: []model.Provision{
			model.HashedProvision(model.ProvisionBIOS, "gba_bios.bin", "Game Boy Advance BIOS", []string{"a860e8c0b6d573d191e4ec7db1b1e4f6"}).ForSystems(model.SystemIDGameCube),
		},
	})

	return groups
}
