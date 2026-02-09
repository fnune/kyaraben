package rpcs3

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDRPCS3,
		Name:    "RPCS3",
		Systems: []model.SystemID{model.SystemIDPS3},
		Package: model.AppImageRef("rpcs3"),
		ProvisionGroups: []model.ProvisionGroup{{
			MinRequired: 1,
			Message:     "Firmware required (provides system libraries and OS)",
			Provisions: []model.Provision{
				model.FileProvision(model.ProvisionFirmware, "PS3UPDAT.PUP", "Official firmware").WithImportViaUI(),
			},
		}},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "rpcs3",
			GenericName: "PlayStation 3 Emulator",
			Categories:  []string{"Game", "Emulator"},
			RomCommand: func(opts model.RomLaunchOptions) string {
				cmd := opts.BinaryPath
				if opts.Fullscreen {
					cmd += " --fullscreen"
				}
				cmd += " %ROM%"
				return cmd
			},
		},
		PathUsage: model.PathUsage{
			OpaqueContents: "dev_hdd0, dev_flash (firmware, saves, game data)",
		},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

var vfsTarget = model.ConfigTarget{
	RelPath: "rpcs3/vfs.yml",
	Format:  model.ConfigFormatYAML,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	opaqueDir := store.EmulatorOpaqueDir(model.EmulatorIDRPCS3) + "/"

	return []model.ConfigPatch{{
		Target: vfsTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"$(EmulatorDir)"}, Value: opaqueDir},
			{Path: []string{"/dev_hdd0/"}, Value: "$(EmulatorDir)dev_hdd0/"},
			{Path: []string{"/dev_hdd1/"}, Value: "$(EmulatorDir)dev_hdd1/"},
			{Path: []string{"/dev_flash/"}, Value: "$(EmulatorDir)dev_flash/"},
			{Path: []string{"/dev_flash2/"}, Value: "$(EmulatorDir)dev_flash2/"},
			{Path: []string{"/dev_flash3/"}, Value: "$(EmulatorDir)dev_flash3/"},
			{Path: []string{"/dev_usb000/"}, Value: "$(EmulatorDir)dev_usb000/"},
			{Path: []string{"/dev_bdvd/"}, Value: ""},
			{Path: []string{"/app_home/"}, Value: ""},
			{Path: []string{"/games/"}, Value: store.SystemRomsDir(model.SystemIDPS3) + "/"},
		},
	}}, nil
}
