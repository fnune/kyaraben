// BIOS hash data compiled from:
// - EmuDeck (https://github.com/dragoonDorise/EmuDeck) - GPL-3
// - RetroDECK (https://github.com/XargonWan/RetroDECK) - GPL-3
// - Libretro documentation (https://docs.libretro.com)
package retroarchbeetlesaturn

import (
	"github.com/fnune/kyaraben/internal/emulators/retroarch"
	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDRetroArchBeetleSaturn,
		Name:    "RetroArch (Beetle Saturn)",
		Systems: []model.SystemID{model.SystemIDSaturn},
		Package: model.AppImageRef("retroarch"),
		ProvisionGroups: []model.ProvisionGroup{{
			MinRequired: 1,
			Message:     "At least one BIOS required",
			Provisions:  saturnBIOSProvisions,
		}},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher:  retroarch.LauncherWithCore(libretroCoreName),
		PathUsage: model.StandardPathUsage(),
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

type Config struct{}

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	return []model.ConfigPatch{
		retroarch.SharedConfig(store),
		coreOverrideConfig(store),
	}, nil
}

func (c *Config) Symlinks(store model.StoreReader, resolver model.BaseDirResolver) ([]model.SymlinkSpec, error) {
	return retroarch.CoreSymlinks(model.EmulatorIDRetroArchBeetleSaturn, store, resolver)
}

const (
	libretroCoreName = "mednafen_saturn_libretro"
	shortCoreName    = "mednafen_saturn"
)

func coreOverrideConfig(store model.StoreReader) model.ConfigPatch {
	return model.ConfigPatch{
		Target: retroarch.CoreOverrideTarget(shortCoreName),
		Entries: []model.ConfigEntry{
			{Path: []string{"rgui_browser_directory"}, Value: store.SystemRomsDir(model.SystemIDSaturn)},
		},
	}
}

var saturnBIOSProvisions = []model.Provision{
	{Kind: model.ProvisionBIOS, Filename: "sega_101.bin", Description: "Japan", Hashes: []string{"85ec9ca47d8f6807718151cbcca8b964"}},
	{Kind: model.ProvisionBIOS, Filename: "sega_100.bin", Description: "Japan", Hashes: []string{"af5828fdff51384f99b3c4926be27762"}},
	{Kind: model.ProvisionBIOS, Filename: "sega_100a.bin", Description: "Japan alternate", Hashes: []string{"f273555d7d91e8a5a6bfd9bcf066331c"}},
	{Kind: model.ProvisionBIOS, Filename: "sega1003.bin", Description: "Japan v1.003", Hashes: []string{"74570fed4d44b2682b560c8cd44b8b6a"}},
	{Kind: model.ProvisionBIOS, Filename: "mpr-17933.bin", Description: "US/EU", Hashes: []string{"3240872c70984b6cbfda1586cab68dbe"}},
	{Kind: model.ProvisionBIOS, Filename: "mpr-18100.bin", Description: "US/EU v1.01", Hashes: []string{"cb2cebc1b6e573b7c44523d037edcd45"}},
	{Kind: model.ProvisionBIOS, Filename: "saturn_bios.bin", Description: "Generic", Hashes: []string{"af5828fdff51384f99b3c4926be27762"}},
	{Kind: model.ProvisionBIOS, Filename: "hisaturn.bin", Description: "Hi-Saturn Japan", Hashes: []string{"3ea3202e2634cb47cb90f3a05c015010"}},
	{Kind: model.ProvisionBIOS, Filename: "vsaturn.bin", Description: "V-Saturn Japan", Hashes: []string{"ac4e4b6522e200c0d23d371a8cecbfd3"}},
	{Kind: model.ProvisionBIOS, Filename: "mpr-18811-mx.ic1", Description: "KOF95 cartridge", Hashes: []string{"255113ba943c92a54facd25a10fd780c"}},
	{Kind: model.ProvisionBIOS, Filename: "mpr-19367-mx.ic1", Description: "Ultraman cartridge", Hashes: []string{"1cd19988d1d72a3e7caa0b73234c96b4"}},
}
