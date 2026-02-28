package xeniaedge

import (
	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:              model.EmulatorIDXeniaEdge,
		Name:            "Xenia Edge",
		Systems:         []model.SystemID{model.SystemIDXbox360},
		Package:         model.AppImageRef("xenia-edge"),
		ProvisionGroups: nil,
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "xenia-edge",
			GenericName: "Xbox 360 Emulator",
			Categories:  []string{"Game", "Emulator"},
			Env:         map[string]string{"GDK_BACKEND": "x11"},
			RomCommand: func(opts model.RomLaunchOptions) string {
				cmd := opts.BinaryPath
				if opts.Fullscreen {
					cmd += " --fullscreen"
				}
				cmd += " %ROM%"
				return cmd
			},
		},
		PathUsage: model.StandardPathUsage(),
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

var configTarget = model.ConfigTarget{
	RelPath: "Xenia/xenia-edge.config.toml",
	Format:  model.ConfigFormatTOML,
	BaseDir: model.ConfigBaseDirUserData,
}

type Config struct{}

func (c *Config) Generate(ctx model.GenerateContext) (model.GenerateResult, error) {
	store := ctx.Store
	savesDir := store.SystemSavesDir(model.SystemIDXbox360)
	romsDir := store.SystemRomsDir(model.SystemIDXbox360)

	entries := []model.ConfigEntry{
		{Path: []string{"Storage", "content_root"}, Value: romsDir},
		{Path: []string{"Storage", "storage_root"}, Value: savesDir},
		{Path: []string{"HID", "hid"}, Value: "sdl"},
		{Path: []string{"HID", "guide_button"}, Value: "false"},
	}

	patches := []model.ConfigPatch{{
		Target:  configTarget,
		Entries: entries,
	}}

	return model.GenerateResult{Patches: patches}, nil
}
