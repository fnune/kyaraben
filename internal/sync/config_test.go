package sync

import (
	"testing"

	"github.com/fnune/kyaraben/internal/folders"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/syncthing"
	"github.com/fnune/kyaraben/internal/testutil"
)

func TestConfigGenerator_FolderCreateRequests(t *testing.T) {
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
		{ID: "retroarch:bsnes", UsesStatesDir: true, UsesScreenshotsDir: true},
		{ID: "duckstation", UsesStatesDir: true, UsesScreenshotsDir: true},
	}
	gen := NewConfigGenerator(fs, cfg, "/home/user/Emulation", systems, emulators, nil)

	requests := gen.FolderCreateRequests()

	foldersByID := make(map[string]syncthing.FolderCreateRequest)
	for _, f := range requests {
		foldersByID[f.ID] = f
	}

	tests := []struct {
		folderID string
		wantType string
		wantPath string
	}{
		{"kyaraben-roms-snes", "sendreceive", "/home/user/Emulation/roms/snes"},
		{"kyaraben-roms-psx", "sendreceive", "/home/user/Emulation/roms/psx"},
		{"kyaraben-bios-snes", "sendreceive", "/home/user/Emulation/bios/snes"},
		{"kyaraben-bios-psx", "sendreceive", "/home/user/Emulation/bios/psx"},
		{"kyaraben-saves-snes", "sendreceive", "/home/user/Emulation/saves/snes"},
		{"kyaraben-saves-psx", "sendreceive", "/home/user/Emulation/saves/psx"},
		{"kyaraben-screenshots-retroarch", "sendreceive", "/home/user/Emulation/screenshots/retroarch"},
		{"kyaraben-screenshots-duckstation", "sendreceive", "/home/user/Emulation/screenshots/duckstation"},
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

func TestConfigGenerator_WriteIgnoreFiles(t *testing.T) {
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
		{ID: "retroarch:bsnes", UsesStatesDir: true, UsesScreenshotsDir: true},
	}
	gen := NewConfigGenerator(fs, cfg, "/emulation", []model.SystemID{"snes"}, emulators, nil)

	requests := gen.FolderCreateRequests()
	if err := gen.WriteIgnoreFiles(requests); err != nil {
		t.Fatalf("WriteIgnoreFiles() error = %v", err)
	}

	ignoreFiles := []string{
		"/emulation/roms/snes/.stignore",
		"/emulation/saves/snes/.stignore",
		"/emulation/states/retroarch:bsnes/.stignore",
		"/emulation/bios/snes/.stignore",
		"/emulation/screenshots/retroarch/.stignore",
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

func TestConfigGenerator_WriteIgnoreFiles_NoFilesWhenNoPatterns(t *testing.T) {
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
		{ID: "retroarch:bsnes", UsesStatesDir: true, UsesScreenshotsDir: true},
	}
	gen := NewConfigGenerator(fs, cfg, "/emulation", []model.SystemID{"snes"}, emulators, nil)

	requests := gen.FolderCreateRequests()
	if err := gen.WriteIgnoreFiles(requests); err != nil {
		t.Fatalf("WriteIgnoreFiles() error = %v", err)
	}

	ignorePath := "/emulation/roms/snes/.stignore"
	if _, err := fs.ReadFile(ignorePath); err == nil {
		t.Errorf("expected %s to not exist when no patterns configured", ignorePath)
	}
}

func TestConfigGenerator_FiltersEmulatorsByPathUsage(t *testing.T) {
	cfg := model.SyncConfig{
		Enabled:   true,
		Syncthing: model.SyncthingConfig{GUIPort: 8385},
	}

	fs := testutil.NewTestFS(t, nil)

	emulators := []folders.EmulatorInfo{
		{ID: "duckstation", UsesStatesDir: true, UsesScreenshotsDir: true},
		{ID: "eden", UsesStatesDir: false, UsesScreenshotsDir: false},
	}
	gen := NewConfigGenerator(fs, cfg, "/emulation", []model.SystemID{"psx", "switch"}, emulators, nil)

	requests := gen.FolderCreateRequests()

	var hasDuckstationStates, hasEdenStates bool
	var hasDuckstationScreenshots, hasEdenScreenshots bool
	for _, f := range requests {
		switch f.ID {
		case "kyaraben-states-duckstation":
			hasDuckstationStates = true
		case "kyaraben-states-eden":
			hasEdenStates = true
		case "kyaraben-screenshots-duckstation":
			hasDuckstationScreenshots = true
		case "kyaraben-screenshots-eden":
			hasEdenScreenshots = true
		}
	}

	if !hasDuckstationStates {
		t.Error("should have states folder for duckstation (UsesStatesDir=true)")
	}
	if hasEdenStates {
		t.Error("should not have states folder for eden (UsesStatesDir=false)")
	}
	if !hasDuckstationScreenshots {
		t.Error("should have screenshots folder for duckstation (UsesScreenshotsDir=true)")
	}
	if hasEdenScreenshots {
		t.Error("should not have screenshots folder for eden (UsesScreenshotsDir=false)")
	}
}

func TestConfigGenerator_FrontendFolders(t *testing.T) {
	cfg := model.SyncConfig{
		Enabled:   true,
		Syncthing: model.SyncthingConfig{GUIPort: 8385},
	}

	fs := testutil.NewTestFS(t, nil)

	emulators := []folders.EmulatorInfo{
		{ID: "retroarch:bsnes", UsesStatesDir: true, UsesScreenshotsDir: true},
	}
	frontends := []model.FrontendID{"esde"}
	gen := NewConfigGenerator(fs, cfg, "/emulation", []model.SystemID{"snes", "psx"}, emulators, frontends)

	requests := gen.FolderCreateRequests()

	wantFrontendFolders := []string{
		"kyaraben-frontends-esde-gamelists-snes",
		"kyaraben-frontends-esde-gamelists-psx",
		"kyaraben-frontends-esde-media-snes",
		"kyaraben-frontends-esde-media-psx",
	}

	foldersByID := make(map[string]syncthing.FolderCreateRequest)
	for _, f := range requests {
		foldersByID[f.ID] = f
	}

	for _, wantID := range wantFrontendFolders {
		if _, ok := foldersByID[wantID]; !ok {
			t.Errorf("missing expected frontend folder: %s", wantID)
		}
	}
}

func TestConfigGenerator_GenerateBootstrap(t *testing.T) {
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
	gen := NewConfigGenerator(fs, cfg, "/emulation", []model.SystemID{"snes"}, nil, nil)
	gen.SetAPIKey("test-api-key")

	xmlCfg := gen.GenerateBootstrap()

	if xmlCfg.Version != 37 {
		t.Errorf("Version = %d, want 37", xmlCfg.Version)
	}
	if len(xmlCfg.Folders) != 0 {
		t.Errorf("Folders = %d, want 0 (bootstrap config has no folders)", len(xmlCfg.Folders))
	}
	if len(xmlCfg.Devices) != 0 {
		t.Errorf("Devices = %d, want 0 (bootstrap config has no devices)", len(xmlCfg.Devices))
	}
	if xmlCfg.GUI.APIKey != "test-api-key" {
		t.Errorf("APIKey = %s, want test-api-key", xmlCfg.GUI.APIKey)
	}
	if xmlCfg.GUI.Address != "127.0.0.1:8385" {
		t.Errorf("GUI Address = %s, want 127.0.0.1:8385", xmlCfg.GUI.Address)
	}
	if !xmlCfg.Options.RelaysEnabled {
		t.Error("RelaysEnabled should be true")
	}
}
