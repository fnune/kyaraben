// Package retroarch provides shared configuration for RetroArch cores.
// All RetroArch cores use the same retroarch.cfg for base settings.
// Per-core overrides handle system-specific paths like ROM browsers.
// See: https://docs.libretro.com/guides/change-directories/
package retroarch

import "github.com/fnune/kyaraben/internal/model"

var SharedLauncher = model.LauncherInfo{
	Binary:      "retroarch",
	DisplayName: "RetroArch",
	GenericName: "Multi-system Emulator",
	Categories:  []string{"Game", "Emulator"},
}

var MainConfigTarget = model.ConfigTarget{
	RelPath: "retroarch/retroarch.cfg",
	Format:  model.ConfigFormatCFG,
	BaseDir: model.ConfigBaseDirUserConfig,
}

// SharedConfig generates the base RetroArch configuration shared by all cores.
// Only contains system_directory for BIOS files. Path settings are in per-core overrides.
// See: https://docs.libretro.com/guides/change-directories/
func SharedConfig(store model.StoreReader) model.ConfigPatch {
	return model.ConfigPatch{
		Target: MainConfigTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"system_directory"}, Value: store.BiosDir()},
		},
	}
}

func CoreOverrideTarget(coreName string) model.ConfigTarget {
	return model.ConfigTarget{
		RelPath: "retroarch/config/" + coreName + "/" + coreName + ".cfg",
		Format:  model.ConfigFormatCFG,
		BaseDir: model.ConfigBaseDirUserConfig,
	}
}
