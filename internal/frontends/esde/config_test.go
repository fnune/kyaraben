package esde

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/fnune/kyaraben/internal/model"
)

type fakeStoreReader struct {
	root string
}

func (f *fakeStoreReader) RomsDir() string {
	return filepath.Join(f.root, "roms")
}

func (f *fakeStoreReader) BiosDir() string {
	return filepath.Join(f.root, "bios")
}

func (f *fakeStoreReader) SystemBiosDir(sys model.SystemID) string {
	return filepath.Join(f.root, "bios", string(sys))
}

func (f *fakeStoreReader) SystemSavesDir(sys model.SystemID) string {
	return filepath.Join(f.root, "saves", string(sys))
}

func (f *fakeStoreReader) EmulatorSavesDir(emu model.EmulatorID) string {
	return filepath.Join(f.root, "saves", string(emu))
}

func (f *fakeStoreReader) EmulatorStatesDir(emu model.EmulatorID) string {
	return filepath.Join(f.root, "states", string(emu))
}

func (f *fakeStoreReader) EmulatorScreenshotsDir(emu model.EmulatorID) string {
	return filepath.Join(f.root, "screenshots", string(emu))
}

func (f *fakeStoreReader) SystemRomsDir(sys model.SystemID) string {
	return filepath.Join(f.root, "roms", string(sys))
}

func (f *fakeStoreReader) EmulatorOpaqueDir(emu model.EmulatorID) string {
	return filepath.Join(f.root, "opaque", string(emu))
}

func TestBuildCommandPassesSavesDir(t *testing.T) {
	store := &fakeStoreReader{root: "/emulation"}

	emulators := map[model.EmulatorID]model.Emulator{
		model.EmulatorIDMGBA: {
			ID: model.EmulatorIDMGBA,
			Launcher: model.LauncherInfo{
				Binary: "mgba",
				RomCommand: func(opts model.RomLaunchOptions) string {
					cmd := opts.BinaryPath
					if opts.Fullscreen {
						cmd += " -f"
					}
					if opts.SavesDir != "" {
						cmd += " -C savegamePath=" + opts.SavesDir
					}
					cmd += " %ROM%"
					return cmd
				},
			},
		},
	}

	ctx := model.FrontendContext{
		BinDir: "/opt/bin",
		Store:  store,
		GetEmulator: func(id model.EmulatorID) (model.Emulator, error) {
			return emulators[id], nil
		},
	}

	c := &Config{}

	tests := []struct {
		sysID       model.SystemID
		wantSavesIn string
	}{
		{model.SystemIDGB, "/emulation/saves/gb"},
		{model.SystemIDGBC, "/emulation/saves/gbc"},
		{model.SystemIDGBA, "/emulation/saves/gba"},
	}

	for _, tt := range tests {
		t.Run(string(tt.sysID), func(t *testing.T) {
			cmd := c.buildCommand(ctx, model.EmulatorIDMGBA, tt.sysID)

			if !strings.Contains(cmd, "-C savegamePath="+tt.wantSavesIn) {
				t.Errorf("buildCommand() = %q, want savegamePath=%s", cmd, tt.wantSavesIn)
			}
		})
	}
}
