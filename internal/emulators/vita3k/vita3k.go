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
			OpaqueContents: "config, ux0 (apps, saves, screenshots)",
		},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

type Config struct{}

func (c *Config) LaunchArgs(store model.StoreReader) []string {
	return []string{"-c", store.EmulatorOpaqueDir(model.EmulatorIDVita3K)}
}

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	opaqueDir := store.EmulatorOpaqueDir(model.EmulatorIDVita3K)

	configTarget := model.ConfigTarget{
		RelPath: filepath.Join(opaqueDir, "config.yml"),
		Format:  model.ConfigFormatYAML,
		BaseDir: model.ConfigBaseDirOpaqueDir,
	}

	return []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"pref-path"}, Value: opaqueDir},
		},
	}}, nil
}
