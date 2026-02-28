package xemu

import (
	"path/filepath"

	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDXemu,
		Name:    "xemu",
		Systems: []model.SystemID{model.SystemIDXbox},
		Package: model.AppImageRef("xemu"),
		ProvisionGroups: []model.ProvisionGroup{
			{
				MinRequired: 1,
				Message:     "Boot ROM required",
				Provisions: []model.Provision{
					model.HashedProvision(model.ProvisionBIOS, "mcpx_1.0.bin", "MCPX Boot ROM",
						[]string{"d49c52a4102f6df7bcf8d0617ac475ed"}),
				},
			},
			{
				MinRequired: 1,
				Message:     "Flash ROM required",
				Provisions: []model.Provision{
					model.HashedProvision(model.ProvisionBIOS, "Complex_4627v1.03.bin", "Complex 4627 v1.03",
						[]string{"21445c6f28fca7285b0f167ea770d1e5"}),
					model.HashedProvision(model.ProvisionBIOS, "Complex_4627.bin", "Complex 4627 Retail",
						[]string{"ec00e31e746de2473acfe7903c5a4cb7"}),
				},
			},
		},
		StateKinds: []model.StateKind{
			model.StatePersistent,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "xemu",
			GenericName: "Xbox Emulator",
			Categories:  []string{"Game", "Emulator"},
			Env:         map[string]string{"GDK_BACKEND": "x11"},
			RomCommand: func(opts model.RomLaunchOptions) string {
				cmd := opts.BinaryPath + " -dvd_path %ROM%"
				if opts.Fullscreen {
					cmd += " -full-screen"
				}
				return cmd
			},
		},
		PathUsage: model.PathUsage{
			UsesBiosDir:        true,
			UsesSavesDir:       false,
			UsesStatesDir:      true,
			UsesScreenshotsDir: true,
		},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

var configTarget = model.ConfigTarget{
	RelPath: "xemu/xemu/xemu.toml",
	Format:  model.ConfigFormatTOML,
	BaseDir: model.ConfigBaseDirUserData,
}

const hddFilename = "xbox_hdd.qcow2"

type Config struct{}

func (c *Config) Generate(ctx model.GenerateContext) (model.GenerateResult, error) {
	store := ctx.Store
	biosDir := store.SystemBiosDir(model.SystemIDXbox)
	statesDir := store.EmulatorStatesDir(model.EmulatorIDXemu)
	screenshotsDir := store.EmulatorScreenshotsDir(model.EmulatorIDXemu)
	romsDir := store.SystemRomsDir(model.SystemIDXbox)
	hddPath := filepath.Join(statesDir, hddFilename)

	entries := []model.ConfigEntry{
		{Path: []string{"sys", "mem_limit"}, Value: "128"},
		{Path: []string{"sys", "files", "bootrom_path"}, Value: biosDir + "/mcpx_1.0.bin"},
		{Path: []string{"sys", "files", "flashrom_path"}, Value: biosDir + "/Complex_4627v1.03.bin"},
		{Path: []string{"sys", "files", "hdd_path"}, Value: hddPath},
		{Path: []string{"sys", "files", "eeprom_path"}, Value: statesDir + "/eeprom.bin"},
		{Path: []string{"general", "show_welcome"}, Value: "false"},
		{Path: []string{"general", "check_for_update"}, Value: "false"},
		{Path: []string{"general", "screenshot_dir"}, Value: screenshotsDir},
		{Path: []string{"general", "games_dir"}, Value: romsDir},
		{Path: []string{"general", "misc", "skip_boot_anim"}, Value: "true", DefaultOnly: true},
		{Path: []string{"display", "renderer"}, Value: "VULKAN", DefaultOnly: true},
		{Path: []string{"display", "quality", "surface_scale"}, Value: "2", DefaultOnly: true},
		{Path: []string{"input", "bindings", "port1"}, Value: model.SteamDeckGUID, DefaultOnly: true},
		{Path: []string{"input", "bindings", "port2"}, Value: model.SteamDeckGUID, DefaultOnly: true},
		{Path: []string{"input", "bindings", "port3"}, Value: model.SteamDeckGUID, DefaultOnly: true},
		{Path: []string{"input", "bindings", "port4"}, Value: model.SteamDeckGUID, DefaultOnly: true},
	}

	patches := []model.ConfigPatch{{
		Target:  configTarget,
		Entries: entries,
	}}

	initialDownloads := []model.InitialDownload{{
		URL:      "https://github.com/xemu-project/xemu-dashboard/releases/download/v20250806-0635/xbox_hdd.qcow2",
		SHA256:   "sha256-xew+D7xPMdNsEyEd7X8HDztfBOmaNn++XYwtjG5Pkl0=",
		DestPath: hddPath,
	}}

	return model.GenerateResult{
		Patches:          patches,
		InitialDownloads: initialDownloads,
	}, nil
}
