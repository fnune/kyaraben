package sync

import (
	"strings"
	"testing"

	"github.com/fnune/kyaraben/internal/folders"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/testutil"
)

func TestConfigGenerator_GenerateFolders(t *testing.T) {
	cfg := model.SyncConfig{
		Enabled: true,
		Syncthing: model.SyncthingConfig{
			ListenPort:    22001,
			DiscoveryPort: 21028,
			GUIPort:       8385,
			RelayEnabled:  true,
		},
	}

	fs := testutil.NewTestFS(t, nil)

	systems := []model.SystemID{"snes", "psx"}
	emulators := []folders.EmulatorInfo{
		{ID: "retroarch:bsnes", UsesStatesDir: true},
		{ID: "duckstation", UsesStatesDir: true},
	}
	gen := NewConfigGenerator(fs, cfg, "/home/user/Emulation", systems, emulators, nil)
	gen.SetDeviceID("TEST-DEVICE-ID")
	gen.SetAPIKey("test-api-key")

	xmlCfg, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	foldersByID := make(map[string]XMLFolder)
	for _, f := range xmlCfg.Folders {
		foldersByID[f.ID] = f
	}

	tests := []struct {
		folderID string
		wantType FolderType
		wantPath string
	}{
		{"kyaraben-roms-snes", FolderTypeSendReceive, "/home/user/Emulation/roms/snes"},
		{"kyaraben-roms-psx", FolderTypeSendReceive, "/home/user/Emulation/roms/psx"},
		{"kyaraben-bios-snes", FolderTypeSendReceive, "/home/user/Emulation/bios/snes"},
		{"kyaraben-bios-psx", FolderTypeSendReceive, "/home/user/Emulation/bios/psx"},
		{"kyaraben-saves-snes", FolderTypeSendReceive, "/home/user/Emulation/saves/snes"},
		{"kyaraben-saves-psx", FolderTypeSendReceive, "/home/user/Emulation/saves/psx"},
		{"kyaraben-screenshots", FolderTypeSendReceive, "/home/user/Emulation/screenshots"},
	}

	for _, tt := range tests {
		t.Run(tt.folderID, func(t *testing.T) {
			folder, ok := foldersByID[tt.folderID]
			if !ok {
				t.Fatalf("folder %s not found", tt.folderID)
			}
			if folder.Type != tt.wantType {
				t.Errorf("folder %s type = %v, want %v", tt.folderID, folder.Type, tt.wantType)
			}
			if folder.Path != tt.wantPath {
				t.Errorf("folder %s path = %v, want %v", tt.folderID, folder.Path, tt.wantPath)
			}
		})
	}
}

func TestConfigGenerator_Versioning(t *testing.T) {
	cfg := model.SyncConfig{
		Enabled:   true,
		Syncthing: model.SyncthingConfig{GUIPort: 8385},
	}

	fs := testutil.NewTestFS(t, nil)

	emulators := []folders.EmulatorInfo{
		{ID: "retroarch:bsnes", UsesStatesDir: true},
	}
	gen := NewConfigGenerator(fs, cfg, "/tmp", []model.SystemID{"snes"}, emulators, nil)

	xmlCfg, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	var savesFolder, romsFolder XMLFolder
	for _, f := range xmlCfg.Folders {
		switch f.ID {
		case "kyaraben-saves-snes":
			savesFolder = f
		case "kyaraben-roms-snes":
			romsFolder = f
		}
	}

	if savesFolder.Versioning.Type != "staggered" {
		t.Errorf("saves folder should have staggered versioning, got %s", savesFolder.Versioning.Type)
	}

	if romsFolder.Versioning.Type != "" {
		t.Errorf("roms folder should not have versioning, got %s", romsFolder.Versioning.Type)
	}
}

func TestConfigGenerator_WriteConfig_WritesIgnoreFiles(t *testing.T) {
	cfg := model.SyncConfig{
		Enabled: true,
		Syncthing: model.SyncthingConfig{
			ListenPort:    22001,
			DiscoveryPort: 21028,
			GUIPort:       8385,
		},
		Ignore: model.SyncIgnoreConfig{
			Patterns: []string{
				"**/shader_cache/**",
				"**/cache/**",
				"**/*.tmp",
			},
		},
	}

	fs := testutil.NewTestFS(t, nil)

	emulators := []folders.EmulatorInfo{
		{ID: "retroarch:bsnes", UsesStatesDir: true},
	}
	gen := NewConfigGenerator(fs, cfg, "/emulation", []model.SystemID{"snes"}, emulators, nil)
	gen.SetAPIKey("test-key")

	if err := gen.WriteConfig("/config"); err != nil {
		t.Fatalf("WriteConfig() error = %v", err)
	}

	ignoreFiles := []string{
		"/emulation/roms/snes/.stignore",
		"/emulation/saves/snes/.stignore",
		"/emulation/states/retroarch:bsnes/.stignore",
		"/emulation/bios/snes/.stignore",
		"/emulation/screenshots/.stignore",
	}

	for _, path := range ignoreFiles {
		content, err := fs.ReadFile(path)
		if err != nil {
			t.Errorf("expected %s to exist: %v", path, err)
			continue
		}

		expected := "**/shader_cache/**\n**/cache/**\n**/*.tmp\n"
		if string(content) != expected {
			t.Errorf("%s content = %q, want %q", path, string(content), expected)
		}
	}
}

func TestConfigGenerator_WriteConfig_NoIgnoreFilesWhenNoPatterns(t *testing.T) {
	cfg := model.SyncConfig{
		Enabled: true,
		Syncthing: model.SyncthingConfig{
			ListenPort:    22001,
			DiscoveryPort: 21028,
			GUIPort:       8385,
		},
	}

	fs := testutil.NewTestFS(t, nil)

	emulators := []folders.EmulatorInfo{
		{ID: "retroarch:bsnes", UsesStatesDir: true},
	}
	gen := NewConfigGenerator(fs, cfg, "/emulation", []model.SystemID{"snes"}, emulators, nil)
	gen.SetAPIKey("test-key")

	if err := gen.WriteConfig("/config"); err != nil {
		t.Fatalf("WriteConfig() error = %v", err)
	}

	ignorePath := "/emulation/roms/snes/.stignore"
	if _, err := fs.ReadFile(ignorePath); err == nil {
		t.Errorf("expected %s to not exist when no patterns configured", ignorePath)
	}
}

func TestConfigGenerator_WriteConfig_PreservesExistingDevices(t *testing.T) {
	existingConfig := `<?xml version="1.0" encoding="UTF-8"?>
<configuration version="37">
  <device id="LOCAL-DEVICE-ID" name="this-device" compression="metadata"></device>
  <device id="PAIRED-DEVICE-ID" name="steamdeck" compression="metadata"></device>
</configuration>`

	fs := testutil.NewTestFS(t, map[string]any{
		"/config/config.xml": existingConfig,
	})

	cfg := model.SyncConfig{
		Enabled: true,
		Syncthing: model.SyncthingConfig{
			ListenPort:    22001,
			DiscoveryPort: 21028,
			GUIPort:       8385,
		},
	}

	emulators := []folders.EmulatorInfo{
		{ID: "retroarch:bsnes", UsesStatesDir: true},
	}
	gen := NewConfigGenerator(fs, cfg, "/emulation", []model.SystemID{"snes"}, emulators, nil)
	gen.SetDeviceID("LOCAL-DEVICE-ID")
	gen.SetAPIKey("test-key")

	if err := gen.WriteConfig("/config"); err != nil {
		t.Fatalf("WriteConfig() error = %v", err)
	}

	content, err := fs.ReadFile("/config/config.xml")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	configStr := string(content)
	if !strings.Contains(configStr, "PAIRED-DEVICE-ID") {
		t.Errorf("config should preserve paired device ID")
	}
	if !strings.Contains(configStr, `name="steamdeck"`) {
		t.Errorf("config should preserve paired device name")
	}
}

func TestConfigGenerator_FiltersEmulatorsByUsesStatesDir(t *testing.T) {
	cfg := model.SyncConfig{
		Enabled:   true,
		Syncthing: model.SyncthingConfig{GUIPort: 8385},
	}

	fs := testutil.NewTestFS(t, nil)

	emulators := []folders.EmulatorInfo{
		{ID: "duckstation", UsesStatesDir: true},
		{ID: "eden", UsesStatesDir: false},
	}
	gen := NewConfigGenerator(fs, cfg, "/emulation", []model.SystemID{"psx", "switch"}, emulators, nil)

	xmlCfg, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	var hasDuckstationStates, hasEdenStates bool
	for _, f := range xmlCfg.Folders {
		if f.ID == "kyaraben-states-duckstation" {
			hasDuckstationStates = true
		}
		if f.ID == "kyaraben-states-eden" {
			hasEdenStates = true
		}
	}

	if !hasDuckstationStates {
		t.Error("should have states folder for duckstation (UsesStatesDir=true)")
	}
	if hasEdenStates {
		t.Error("should not have states folder for eden (UsesStatesDir=false)")
	}
}

func TestConfigGenerator_FrontendFolders(t *testing.T) {
	cfg := model.SyncConfig{
		Enabled:   true,
		Syncthing: model.SyncthingConfig{GUIPort: 8385},
	}

	fs := testutil.NewTestFS(t, nil)

	emulators := []folders.EmulatorInfo{
		{ID: "retroarch:bsnes", UsesStatesDir: true},
	}
	frontends := []model.FrontendID{"esde"}
	gen := NewConfigGenerator(fs, cfg, "/emulation", []model.SystemID{"snes", "psx"}, emulators, frontends)

	xmlCfg, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	wantFrontendFolders := []string{
		"kyaraben-frontends-esde-gamelists-snes",
		"kyaraben-frontends-esde-gamelists-psx",
		"kyaraben-frontends-esde-media-snes",
		"kyaraben-frontends-esde-media-psx",
	}

	foldersByID := make(map[string]XMLFolder)
	for _, f := range xmlCfg.Folders {
		foldersByID[f.ID] = f
	}

	for _, wantID := range wantFrontendFolders {
		if _, ok := foldersByID[wantID]; !ok {
			t.Errorf("missing expected frontend folder: %s", wantID)
		}
	}
}
