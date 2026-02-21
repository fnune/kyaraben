package eden

import (
	"path/filepath"
	"strings"

	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDEden,
		Name:    "Eden",
		Systems: []model.SystemID{model.SystemIDSwitch},
		Package: model.AppImageRef("eden"),
		// Eden copies keys to ~/.local/share/eden/keys/ on import (just a file copy).
		// See: https://gitlab.com/codxjb/eden/-/blob/master/src/yuzu/main.cpp#L4353
		// Firmware is copied to nand/system/Contents/registered/ as *.nca files.
		// See: https://gitlab.com/codxjb/eden/-/blob/master/src/yuzu/main.cpp#L4216
		ProvisionGroups: []model.ProvisionGroup{
			{
				MinRequired: 1,
				Message:     "Decryption keys required",
				Provisions: []model.Provision{
					model.FileProvision(model.ProvisionKeys, "prod.keys", "Production keys"),
				},
			},
			{
				MinRequired: 0,
				Message:     "Title keys (optional, for DLC and updates)",
				Provisions: []model.Provision{
					model.FileProvision(model.ProvisionKeys, "title.keys", "Title keys"),
				},
			},
			{
				MinRequired: 0,
				Message:     "Firmware (optional, enables system applets)",
				Provisions: []model.Provision{
					model.PatternProvision(model.ProvisionFirmware, "*.nca", "firmware NCA", "Switch firmware"),
				},
			},
		},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "eden",
			GenericName: "Nintendo Switch Emulator",
			Categories:  []string{"Game", "Emulator"},
			RomCommand: func(opts model.RomLaunchOptions) string {
				cmd := opts.BinaryPath
				if len(opts.LaunchArgs) > 0 {
					cmd += " " + strings.Join(opts.LaunchArgs, " ")
				}
				if opts.Fullscreen {
					cmd += " -f"
				}
				cmd += " -g %ROM%"
				return cmd
			},
		},
		PathUsage: model.PathUsage{
			UsesBiosDir:        true,
			UsesSavesDir:       true,
			UsesScreenshotsDir: true,
		},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

type Config struct{}

var configTarget = model.ConfigTarget{
	RelPath: "eden/qt-config.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

func (c *Config) Generate(ctx model.GenerateContext) (model.GenerateResult, error) {
	store := ctx.Store

	patches := []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"UI", "Screenshots\\screenshot_path"}, Value: store.EmulatorScreenshotsDir(model.EmulatorIDEden)},
			{Path: []string{"UI", "Paths\\gamedirs\\size"}, Value: "1"},
			{Path: []string{"UI", "Paths\\gamedirs\\1\\deep_scan"}, Value: "false"},
			{Path: []string{"UI", "Paths\\gamedirs\\1\\expanded"}, Value: "true"},
			{Path: []string{"UI", "Paths\\gamedirs\\1\\path"}, Value: store.SystemRomsDir(model.SystemIDSwitch)},
		},
	}}

	dataDir, err := ctx.BaseDirResolver.UserDataDir()
	if err != nil {
		return model.GenerateResult{}, err
	}
	edenDir := filepath.Join(dataDir, "eden")
	biosDir := store.SystemBiosDir(model.SystemIDSwitch)

	symlinks := []model.SymlinkSpec{
		{Source: filepath.Join(edenDir, "keys"), Target: biosDir},
		{Source: filepath.Join(edenDir, "nand", "system", "Contents", "registered"), Target: biosDir},
		{Source: filepath.Join(edenDir, "screenshots"), Target: store.EmulatorScreenshotsDir(model.EmulatorIDEden)},
		{Source: filepath.Join(edenDir, "nand", "user", "save"), Target: store.SystemSavesDir(model.SystemIDSwitch)},
	}

	return model.GenerateResult{
		Patches:  patches,
		Symlinks: symlinks,
	}, nil
}
