package dolphin

import (
	"path/filepath"

	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDDolphin,
		Name:    "Dolphin",
		Systems: []model.SystemID{model.SystemIDGameCube, model.SystemIDWii},
		Package: model.AppImageRef("dolphin"),
		// Wii system menu optional, GameCube IPL optional
		Provisions: nil,
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "dolphin",
			GenericName: "GameCube/Wii Emulator",
			Categories:  []string{"Game", "Emulator"},
		},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

var configTarget = model.ConfigTarget{
	RelPath: "dolphin-emu/Dolphin.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

type Config struct{}

func (c *Config) Generate(store model.StoreReader) ([]model.ConfigPatch, error) {
	// Dolphin stores saves in GC memory cards and Wii NAND, both within its data directory.
	// We use an opaque directory pattern since this structure is complex and tightly coupled.
	// See: https://dolphin-emu.org/docs/guides/
	opaqueDir := store.EmulatorOpaqueDir(model.EmulatorIDDolphin)

	return []model.ConfigPatch{{
		Target: configTarget,
		Entries: []model.ConfigEntry{
			// ROM paths
			{Path: []string{"General", "ISOPath0"}, Value: store.SystemRomsDir(model.SystemIDGameCube)},
			{Path: []string{"General", "ISOPath1"}, Value: store.SystemRomsDir(model.SystemIDWii)},
			{Path: []string{"General", "ISOPaths"}, Value: "2"},

			// GC memory cards - Slot A for GameCube saves
			// Uses per-game memory cards for better organization
			{Path: []string{"Core", "MemcardAPath"}, Value: filepath.Join(opaqueDir, "GC", "MemoryCardA.USA.raw")},
			{Path: []string{"Core", "MemcardBPath"}, Value: filepath.Join(opaqueDir, "GC", "MemoryCardB.USA.raw")},
			{Path: []string{"Core", "GCIFolderAPathOverride"}, Value: filepath.Join(opaqueDir, "GC", "USA", "Card A")},
			{Path: []string{"Core", "GCIFolderBPathOverride"}, Value: filepath.Join(opaqueDir, "GC", "USA", "Card B")},

			// Wii NAND and SD card paths
			{Path: []string{"General", "NANDRootPath"}, Value: filepath.Join(opaqueDir, "Wii")},
			{Path: []string{"General", "WiiSDCardPath"}, Value: filepath.Join(opaqueDir, "Wii", "sd.raw")},

			// Dump paths (screenshots, etc.)
			{Path: []string{"General", "DumpPath"}, Value: store.SystemScreenshotsDir(model.SystemIDGameCube)},
		},
	}}, nil
}
