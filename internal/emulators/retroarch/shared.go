// Package retroarch provides shared configuration for RetroArch cores.
// All RetroArch cores use the same retroarch.cfg for base settings.
// Per-core overrides handle system-specific paths like ROM browsers.
// See: https://docs.libretro.com/guides/change-directories/
package retroarch

import (
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/paths"
)

var SharedLauncher = model.LauncherInfo{
	Binary:      "retroarch",
	DisplayName: "RetroArch",
	GenericName: "Multi-system Emulator",
	Categories:  []string{"Game", "Emulator"},
}

// LauncherWithCore returns a copy of SharedLauncher with RomArgs set to load
// the given libretro core via the -L flag when launching games.
func LauncherWithCore(coreName string) model.LauncherInfo {
	l := SharedLauncher
	l.RomArgs = "-L " + coreName + " %ROM%"
	return l
}

var MainConfigTarget = model.ConfigTarget{
	RelPath: "retroarch/retroarch.cfg",
	Format:  model.ConfigFormatCFG,
	BaseDir: model.ConfigBaseDirUserConfig,
}

// SharedConfig generates the base RetroArch configuration shared by all cores.
// Sets up common directories used by all cores. Per-core path settings are in core overrides.
// See: https://docs.libretro.com/guides/change-directories/
func SharedConfig(store model.StoreReader) model.ConfigPatch {
	return model.ConfigPatch{
		Target: MainConfigTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"system_directory"}, Value: store.BiosDir()},
			{Path: []string{"libretro_directory"}, Value: paths.MustRetroArchCoresDir()},
			{Path: []string{"menu_driver"}, Value: "rgui", Unmanaged: true},
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
