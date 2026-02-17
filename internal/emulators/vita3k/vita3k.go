package vita3k

import (
	"path/filepath"
	"strings"

	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDVita3K,
		Name:    "Vita3K",
		Systems: []model.SystemID{model.SystemIDPSVita},
		Package: model.AppImageRef("vita3k"),
		ProvisionGroups: []model.ProvisionGroup{{
			MinRequired: 1,
			Message:     "Firmware required (provides system libraries and OS)",
			Provisions: []model.Provision{
				model.FileProvision(model.ProvisionFirmware, "PSVUPDAT.PUP", "Official firmware").WithImportViaUI(),
			},
		}},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "vita3k",
			GenericName: "PlayStation Vita Emulator",
			Categories:  []string{"Game", "Emulator"},
			RomCommand: func(opts model.RomLaunchOptions) string {
				cmd := opts.BinaryPath
				if len(opts.LaunchArgs) > 0 {
					cmd += " " + strings.Join(opts.LaunchArgs, " ")
				}
				if opts.Fullscreen {
					cmd += " -F"
				}
				cmd += " -r %ROM%"
				return cmd
			},
		},
		PathUsage: model.PathUsage{
			UsesSavesDir:       true,
			UsesScreenshotsDir: true,
		},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

var configTarget = model.ConfigTarget{
	RelPath: "Vita3K/config.yml",
	Format:  model.ConfigFormatYAML,
	BaseDir: model.ConfigBaseDirUserConfig,
}

var userTarget = model.ConfigTarget{
	RelPath: "Vita3K/Vita3K/ux0/user/00/user.xml",
	Format:  model.ConfigFormatRaw,
	BaseDir: model.ConfigBaseDirUserData,
}

const userXML = `<?xml version="1.0" encoding="utf-8"?>
<user id="00" name="Kyaraben">
	<avatar>default</avatar>
	<sort-apps-list type="4" state="1" />
	<theme use-background="true">
		<content-id>default</content-id>
	</theme>
	<start-screen type="default">
		<path></path>
	</start-screen>
	<backgrounds />
</user>
`

type Config struct{}

func (c *Config) Generate(ctx model.GenerateContext) (model.GenerateResult, error) {
	store := ctx.Store

	patches := []model.ConfigPatch{
		{
			Target: configTarget,
			Entries: []model.ConfigEntry{
				{Path: []string{"show-welcome"}, Value: "false"},
				{Path: []string{"check-for-updates"}, Value: "false"},
				{Path: []string{"user-auto-connect"}, Value: "true"},
				{Path: []string{"bgm-volume"}, Value: "0"},
			},
		},
		{
			Target:  userTarget,
			Entries: []model.ConfigEntry{{Value: userXML, Unmanaged: true}},
		},
	}

	dataDir, err := ctx.BaseDirResolver.UserDataDir()
	if err != nil {
		return model.GenerateResult{}, err
	}

	vita3kDataDir := filepath.Join(dataDir, "Vita3K", "Vita3K")
	vita3kScreenshotsDir := filepath.Join(dataDir, "Vita3K", "screenshots")

	symlinks := []model.SymlinkSpec{
		{
			Source: filepath.Join(vita3kDataDir, "ux0", "user", "00", "savedata"),
			Target: store.SystemSavesDir(model.SystemIDPSVita),
		},
		{
			Source: vita3kScreenshotsDir,
			Target: store.EmulatorScreenshotsDir(model.EmulatorIDVita3K),
		},
		{
			Source: filepath.Join(store.SystemRomsDir(model.SystemIDPSVita), "installed"),
			Target: filepath.Join(vita3kDataDir, "ux0", "app"),
		},
	}

	return model.GenerateResult{
		Patches:  patches,
		Symlinks: symlinks,
	}, nil
}
