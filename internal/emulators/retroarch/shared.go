// Package retroarch provides shared configuration for RetroArch cores.
// All RetroArch cores use the same retroarch.cfg for base settings.
// Per-core overrides handle system-specific paths like ROM browsers.
// Symlinks redirect RetroArch's sorted directories to kyaraben's store.
// See: https://docs.libretro.com/guides/change-directories/
package retroarch

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/paths"
)

var SharedLauncher = model.LauncherInfo{
	Binary:      "retroarch",
	DisplayName: "RetroArch",
	GenericName: "Multi-system Emulator",
	Categories:  []string{"Game", "Emulator"},
}

// LauncherWithCore returns a copy of SharedLauncher with RomCommand set to load
// the given libretro core via the -L flag when launching games.
func LauncherWithCore(coreName string) model.LauncherInfo {
	l := SharedLauncher
	l.RomCommand = func(opts model.RomLaunchOptions) string {
		cmd := opts.BinaryPath
		if opts.Fullscreen {
			cmd += " -f"
		}
		cmd += " -L " + coreName + " %ROM%"
		return cmd
	}
	return l
}

var MainConfigTarget = model.ConfigTarget{
	RelPath: "retroarch/retroarch.cfg",
	Format:  model.ConfigFormatCFG,
	BaseDir: model.ConfigBaseDirUserConfig,
}

// SharedConfig generates the base RetroArch configuration shared by all cores.
// Enables per-core sorting so RetroArch creates subdirectories like saves/bsnes/.
// We symlink these sorted directories to kyaraben's store locations.
// Screenshots go directly to a shared retroarch directory (no per-core sorting).
// See: https://docs.libretro.com/guides/change-directories/
func SharedConfig(store model.StoreReader) model.ConfigPatch {
	return model.ConfigPatch{
		Target: MainConfigTarget,
		Entries: []model.ConfigEntry{
			{Path: []string{"system_directory"}, Value: store.BiosDir()},
			{Path: []string{"libretro_directory"}, Value: paths.MustRetroArchCoresDir()},
			{Path: []string{"screenshot_directory"}, Value: store.EmulatorScreenshotsDir(model.EmulatorIDRetroArchBsnes)},
			{Path: []string{"sort_savefiles_enable"}, Value: "true"},
			{Path: []string{"sort_savestates_enable"}, Value: "true"},
			{Path: []string{"sort_savefiles_by_content_enable"}, Value: "false"},
			{Path: []string{"sort_savestates_by_content_enable"}, Value: "false"},
			{Path: []string{"menu_driver"}, Value: "rgui", Unmanaged: true},
			{Path: []string{"menu_show_load_content_animation"}, Value: "false"},
			{Path: []string{"notification_show_config_override_load"}, Value: "false"},
			{Path: []string{"notification_show_remap_load"}, Value: "false"},
		},
	}
}

func CoreOverrideTarget(shortName string) model.ConfigTarget {
	return model.ConfigTarget{
		RelPath: "retroarch/config/" + shortName + "/" + shortName + ".cfg",
		Format:  model.ConfigFormatCFG,
		BaseDir: model.ConfigBaseDirUserConfig,
	}
}

var coreShortNames = map[model.EmulatorID]string{
	model.EmulatorIDRetroArchBsnes:         "bsnes",
	model.EmulatorIDRetroArchMesen:         "mesen",
	model.EmulatorIDRetroArchGenesisPlusGX: "genesis_plus_gx",
	model.EmulatorIDRetroArchMupen64Plus:   "mupen64plus_next",
	model.EmulatorIDRetroArchBeetleSaturn:  "mednafen_saturn",
}

var coreToSystem = map[model.EmulatorID]model.SystemID{
	model.EmulatorIDRetroArchBsnes:         model.SystemIDSNES,
	model.EmulatorIDRetroArchMesen:         model.SystemIDNES,
	model.EmulatorIDRetroArchGenesisPlusGX: model.SystemIDGenesis,
	model.EmulatorIDRetroArchMupen64Plus:   model.SystemIDN64,
	model.EmulatorIDRetroArchBeetleSaturn:  model.SystemIDSaturn,
}

func CoreShortName(emuID model.EmulatorID) string {
	return coreShortNames[emuID]
}

func CreateCoreSymlinks(emuID model.EmulatorID, store model.StoreReader, resolver model.BaseDirResolver) error {
	shortName := CoreShortName(emuID)
	if shortName == "" {
		return nil
	}

	configDir, err := resolver.UserConfigDir()
	if err != nil {
		return fmt.Errorf("getting config dir: %w", err)
	}

	systemID := coreToSystem[emuID]
	links := []struct {
		source string
		target string
	}{
		{
			source: filepath.Join(configDir, "retroarch", "saves", shortName),
			target: store.SystemSavesDir(systemID),
		},
		{
			source: filepath.Join(configDir, "retroarch", "states", shortName),
			target: store.EmulatorStatesDir(emuID),
		},
	}

	for _, link := range links {
		if err := createSymlink(link.source, link.target); err != nil {
			return err
		}
	}
	return nil
}

func createSymlink(source, target string) error {
	info, err := os.Lstat(source)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			existingTarget, err := os.Readlink(source)
			if err != nil {
				return fmt.Errorf("reading symlink %s: %w", source, err)
			}
			if existingTarget == target {
				return nil
			}
			if err := os.Remove(source); err != nil {
				return fmt.Errorf("removing old symlink %s: %w", source, err)
			}
		} else if info.IsDir() {
			entries, err := os.ReadDir(source)
			if err != nil {
				return fmt.Errorf("reading directory %s: %w", source, err)
			}
			if len(entries) > 0 {
				return fmt.Errorf(
					"cannot create symlink at %s: directory contains files; "+
						"please move or remove your existing saves before running kyaraben apply",
					source,
				)
			}
			if err := os.Remove(source); err != nil {
				return fmt.Errorf("removing empty directory %s: %w", source, err)
			}
		} else {
			return fmt.Errorf("cannot create symlink at %s: file exists", source)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("checking %s: %w", source, err)
	}

	if err := os.MkdirAll(filepath.Dir(source), 0755); err != nil {
		return fmt.Errorf("creating parent directory for %s: %w", source, err)
	}

	if err := os.Symlink(target, source); err != nil {
		return fmt.Errorf("creating symlink %s -> %s: %w", source, target, err)
	}

	return nil
}
