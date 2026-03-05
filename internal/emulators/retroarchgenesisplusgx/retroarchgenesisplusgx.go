// BIOS hash data compiled from:
// - Libretro documentation (https://docs.libretro.com/library/genesis_plus_gx/)
// - EmuDeck (https://github.com/dragoonDorise/EmuDeck) - GPL-3
package retroarchgenesisplusgx

import (
	"github.com/fnune/kyaraben/internal/emulators/retroarch"
	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDRetroArchGenesisPlusGX,
		Name:    "Genesis Plus GX (RetroArch)",
		Systems: []model.SystemID{model.SystemIDGenesis, model.SystemIDMasterSystem, model.SystemIDGameGear},
		Package: model.AppImageRef("retroarch"),
		ProvisionGroups: []model.ProvisionGroup{{
			MinRequired: 0,
			Message:     "Optional boot ROM for startup animation",
			Provisions:  bootROMProvisions,
			BaseDir: func(store model.StoreReader, _ model.SystemID) string {
				return store.SystemBiosDir(model.SystemIDGenesis)
			},
		}},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher:          retroarch.LauncherWithCore(libretroCoreName),
		PathUsage:         model.StandardPathUsage(),
		SupportedSettings: []string{model.SettingResumeAutosave, model.SettingResumeAutoload},
		SupportedHotkeys:  retroarch.HotkeyMappings.SupportedHotkeys(),
		ResumeRecommended: true,
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

type Config struct{}

func (c *Config) Generate(ctx model.GenerateContext) (model.GenerateResult, error) {
	symlinks, err := retroarch.CoreSymlinks(model.EmulatorIDRetroArchGenesisPlusGX, ctx.Store, ctx.BaseDirResolver)
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
	downloads, err := retroarch.CoreShaderDownloads(model.EmulatorIDRetroArchGenesisPlusGX, ctx.BaseDirResolver, pc)
	if err != nil {
		return model.GenerateResult{}, err
	}
	return model.GenerateResult{
		Patches:          retroarch.CorePatches(model.EmulatorIDRetroArchGenesisPlusGX, ctx.Store, ctx.ControllerConfig, pc, ctx.BaseDirResolver),
		Symlinks:         symlinks,
		InitialDownloads: downloads,
	}, nil
}

const libretroCoreName = "genesis_plus_gx_libretro"

var bootROMProvisions = []model.Provision{
	model.HashedProvision(model.ProvisionBIOS, "bios_E.sms", "Master System EU boot ROM", []string{"840481177270d5642a14ca71ee72844c"}).ForSystems(model.SystemIDMasterSystem),
	model.HashedProvision(model.ProvisionBIOS, "bios_U.sms", "Master System US boot ROM", []string{"840481177270d5642a14ca71ee72844c"}).ForSystems(model.SystemIDMasterSystem),
	model.HashedProvision(model.ProvisionBIOS, "bios_J.sms", "Master System JP boot ROM", []string{"24a519c53f67b00640d0048ef7089105"}).ForSystems(model.SystemIDMasterSystem),
	model.HashedProvision(model.ProvisionBIOS, "bios.gg", "Game Gear boot ROM", []string{"672e104c3be3a238301aceffc3b23fd6"}).ForSystems(model.SystemIDGameGear),
	model.HashedProvision(model.ProvisionBIOS, "bios_MD.bin", "Mega Drive TMSS boot ROM", []string{"45e298905a08f9cfb38fd504cd6dbc84"}).ForSystems(model.SystemIDGenesis),
}
