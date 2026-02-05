package rpcs3

import "github.com/fnune/kyaraben/internal/model"

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDRPCS3,
		Name:    "RPCS3",
		Systems: []model.SystemID{model.SystemIDPS3},
		Package: model.AppImageRef("rpcs3"),
		// PS3 firmware is installed through the emulator
		Provisions: nil,
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
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

// RPCS3 uses vfs.yml for Virtual File System path mappings.
// This is separate from config.yml which handles emulator settings.
// See: https://wiki.rpcs3.net/index.php?title=Help:Game_Compatibility
var vfsTarget = model.ConfigTarget{
	RelPath: "rpcs3/vfs.yml",
	Format:  model.ConfigFormatYAML,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	// RPCS3's VFS uses $(EmulatorDir) as a base variable that other paths reference.
	// Setting this redirects all of RPCS3's data (games, saves, firmware) to our opaque directory.
	// The trailing slash is required by RPCS3.
	opaqueDir := store.EmulatorOpaqueDir(model.EmulatorIDRPCS3) + "/"

	return []model.ConfigPatch{{
		Target: vfsTarget,
		Entries: []model.ConfigEntry{
			// Base emulator directory - all other VFS paths are relative to this
			{Path: []string{"$(EmulatorDir)"}, Value: opaqueDir},
			// Standard VFS paths that reference the base directory
			{Path: []string{"/dev_hdd0/"}, Value: "$(EmulatorDir)dev_hdd0/"},
			{Path: []string{"/dev_hdd1/"}, Value: "$(EmulatorDir)dev_hdd1/"},
			{Path: []string{"/dev_flash/"}, Value: "$(EmulatorDir)dev_flash/"},
			{Path: []string{"/dev_flash2/"}, Value: "$(EmulatorDir)dev_flash2/"},
			{Path: []string{"/dev_flash3/"}, Value: "$(EmulatorDir)dev_flash3/"},
			{Path: []string{"/dev_usb000/"}, Value: "$(EmulatorDir)dev_usb000/"},
			{Path: []string{"/dev_bdvd/"}, Value: ""},
			{Path: []string{"/app_home/"}, Value: ""},
			// Games directory for disc-based games
			{Path: []string{"/games/"}, Value: store.SystemRomsDir(model.SystemIDPS3) + "/"},
		},
	}}, nil
}
