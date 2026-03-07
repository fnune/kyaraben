package retroarchfbneo

import (
	"github.com/fnune/kyaraben/internal/emulators/retroarch"
	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDRetroArchFBNeo,
		Name:    "FBNeo (RetroArch)",
		Systems: []model.SystemID{model.SystemIDArcade, model.SystemIDNeoGeo},
		Package: model.AppImageRef("retroarch"),
		ProvisionGroups: []model.ProvisionGroup{
			{
				MinRequired: 0,
				Message:     "Neo Geo BIOS",
				Provisions: []model.Provision{
					model.FileProvision(model.ProvisionBIOS, "neogeo.zip", "Neo Geo AES/MVS BIOS").ForSystems(model.SystemIDNeoGeo),
				},
			},
			{
				MinRequired: 0,
				Message:     "PGM BIOS (for IGS PGM games like Knights of Valour)",
				Provisions: []model.Provision{
					model.FileProvision(model.ProvisionBIOS, "pgm.zip", "PGM system BIOS").ForSystems(model.SystemIDArcade),
				},
			},
			{
				MinRequired: 0,
				Message:     "Super Kaneko Nova System BIOS",
				Provisions: []model.Provision{
					model.FileProvision(model.ProvisionBIOS, "skns.zip", "SKNS BIOS").ForSystems(model.SystemIDArcade),
				},
			},
			{
				MinRequired: 0,
				Message:     "DECO Cassette System BIOS",
				Provisions: []model.Provision{
					model.FileProvision(model.ProvisionBIOS, "decocass.zip", "DECO Cassette BIOS").ForSystems(model.SystemIDArcade),
				},
			},
			{
				MinRequired: 0,
				Message:     "Bubble System BIOS (for Konami Bubble System games)",
				Provisions: []model.Provision{
					model.FileProvision(model.ProvisionBIOS, "bubsys.zip", "Bubble System BIOS").ForSystems(model.SystemIDArcade),
				},
			},
			{
				MinRequired: 0,
				Message:     "Namco C69 BIOS",
				Provisions: []model.Provision{
					model.FileProvision(model.ProvisionBIOS, "namcoc69.zip", "Namco C69 BIOS").ForSystems(model.SystemIDArcade),
				},
			},
			{
				MinRequired: 0,
				Message:     "Namco C70 BIOS",
				Provisions: []model.Provision{
					model.FileProvision(model.ProvisionBIOS, "namcoc70.zip", "Namco C70 BIOS").ForSystems(model.SystemIDArcade),
				},
			},
			{
				MinRequired: 0,
				Message:     "Namco C75 BIOS",
				Provisions: []model.Provision{
					model.FileProvision(model.ProvisionBIOS, "namcoc75.zip", "Namco C75 BIOS").ForSystems(model.SystemIDArcade),
				},
			},
			{
				MinRequired: 0,
				Message:     "ISG Selection Master BIOS",
				Provisions: []model.Provision{
					model.FileProvision(model.ProvisionBIOS, "isgsm.zip", "ISG Selection Master Type 2006 BIOS").ForSystems(model.SystemIDArcade),
				},
			},
			{
				MinRequired: 0,
				Message:     "Midway SSIO Sound Board ROM",
				Provisions: []model.Provision{
					model.FileProvision(model.ProvisionBIOS, "midssio.zip", "Midway SSIO Sound Board Internal ROM").ForSystems(model.SystemIDArcade),
				},
			},
			{
				MinRequired: 0,
				Message:     "NMK004 Internal ROM",
				Provisions: []model.Provision{
					model.FileProvision(model.ProvisionBIOS, "nmk004.zip", "NMK004 Internal ROM").ForSystems(model.SystemIDArcade),
				},
			},
			{
				MinRequired: 0,
				Message:     "YM2608 Internal ROM (for FM sound in some games)",
				Provisions: []model.Provision{
					model.FileProvision(model.ProvisionBIOS, "ym2608.zip", "YM2608 Internal ROM").ForSystems(model.SystemIDArcade),
				},
			},
			{
				MinRequired: 0,
				Message:     "C-Chip Internal ROM",
				Provisions: []model.Provision{
					model.FileProvision(model.ProvisionBIOS, "cchip.zip", "C-Chip Internal ROM").ForSystems(model.SystemIDArcade),
				},
			},
		},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher:          retroarch.LauncherWithCore(libretroCoreName),
		PathUsage:         model.StandardPathUsage(),
		SupportedSettings: []string{model.SettingPreset, model.SettingResumeAutosave, model.SettingResumeAutoload},
		SupportedHotkeys:  retroarch.HotkeyMappings.SupportedHotkeys(),
		ResumeRecommended: true,
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

var coreOptionsTarget = model.ConfigTarget{
	RelPath: "retroarch/config/FinalBurn Neo/FinalBurn Neo.opt",
	Format:  model.ConfigFormatCFG,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(ctx model.GenerateContext) (model.GenerateResult, error) {
	symlinks, err := retroarch.CoreSymlinks(model.EmulatorIDRetroArchFBNeo, ctx.Store, ctx.BaseDirResolver)
	if err != nil {
		return model.GenerateResult{}, err
	}
	pc := &retroarch.PresetConfig{
		Preset:             ctx.Preset,
		Bezels:             ctx.Bezels,
		TargetDevice:       ctx.TargetDevice,
		Resume:             ctx.Resume,
		SystemDisplayTypes: ctx.SystemDisplayTypes,
	}
	downloads, err := retroarch.CoreShaderDownloads(model.EmulatorIDRetroArchFBNeo, ctx.BaseDirResolver, pc)
	if err != nil {
		return model.GenerateResult{}, err
	}
	systems := []model.SystemID{model.SystemIDArcade}
	embeddedFiles, err := retroarch.CoreEmbeddedFiles(systems, pc, ctx.BaseDirResolver)
	if err != nil {
		return model.GenerateResult{}, err
	}

	patches := retroarch.CorePatches(model.EmulatorIDRetroArchFBNeo, ctx.Store, ctx.ControllerConfig, pc, ctx.BaseDirResolver)
	patches = append(patches, model.ConfigPatch{
		Target: coreOptionsTarget,
		Entries: []model.ConfigEntry{
			model.Default(model.None, model.Path("fbneo-allow-patched-romsets"), "enabled"),
			model.Default(model.None, model.Path("fbneo-allow-depth-32"), "enabled"),
		},
	})
	patches = append(patches, retroarch.OverlayPatches(model.EmulatorIDRetroArchFBNeo, systems, pc, ctx.BaseDirResolver)...)

	return model.GenerateResult{
		Patches:          patches,
		Symlinks:         symlinks,
		InitialDownloads: downloads,
		EmbeddedFiles:    embeddedFiles,
	}, nil
}

const libretroCoreName = "fbneo_libretro"
