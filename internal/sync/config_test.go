package sync

import (
	"encoding/xml"
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

func TestWriteConfig_CreatesBootstrapWhenNoExisting(t *testing.T) {
	fs := testutil.NewTestFS(t, nil)

	cfg := model.SyncConfig{
		Syncthing: model.SyncthingConfig{
			GUIPort:       8484,
			ListenPort:    22100,
			DiscoveryPort: 21127,
		},
	}

	gen := NewConfigGenerator(fs, cfg, "/collection", nil, nil, nil)
	gen.SetAPIKey("test-key")

	if err := gen.WriteConfig("/config"); err != nil {
		t.Fatalf("WriteConfig() error = %v", err)
	}

	data, err := fs.ReadFile("/config/config.xml")
	if err != nil {
		t.Fatalf("reading config.xml: %v", err)
	}

	var config SyncthingXMLConfig
	if err := xml.Unmarshal(data, &config); err != nil {
		t.Fatalf("parsing config.xml: %v", err)
	}

	if config.Version != 37 {
		t.Errorf("Version = %d, want 37", config.Version)
	}
	if len(config.Devices) != 0 {
		t.Errorf("expected 0 devices in bootstrap, got %d", len(config.Devices))
	}
	if len(config.Folders) != 0 {
		t.Errorf("expected 0 folders in bootstrap, got %d", len(config.Folders))
	}
	if config.GUI.APIKey != "test-key" {
		t.Errorf("APIKey = %s, want test-key", config.GUI.APIKey)
	}
}

func TestWriteConfig_MergesExistingConfig(t *testing.T) {
	existingConfig := `<?xml version="1.0" encoding="UTF-8"?>
<configuration version="35">
  <folder id="existing-folder" label="existing" path="/data/folder" type="sendreceive"></folder>
  <device id="DEVICE-A" name="phone" compression="metadata">
    <address>tcp://192.168.1.100:22000</address>
  </device>
  <device id="DEVICE-B" name="laptop" compression="metadata"></device>
  <gui enabled="true">
    <address>127.0.0.1:8080</address>
    <apikey>old-key</apikey>
  </gui>
  <options>
    <listenAddress>tcp://0.0.0.0:22000</listenAddress>
    <globalAnnounceEnabled>false</globalAnnounceEnabled>
    <relaysEnabled>false</relaysEnabled>
  </options>
</configuration>`

	fs := testutil.NewTestFS(t, map[string]any{
		"/config/config.xml": existingConfig,
	})

	cfg := model.SyncConfig{
		Syncthing: model.SyncthingConfig{
			GUIPort:                9999,
			ListenPort:             22100,
			DiscoveryPort:          21127,
			GlobalDiscoveryEnabled: true,
			RelayEnabled:           true,
		},
	}

	gen := NewConfigGenerator(fs, cfg, "/collection", nil, nil, nil)
	gen.SetAPIKey("new-key")

	if err := gen.WriteConfig("/config"); err != nil {
		t.Fatalf("WriteConfig() error = %v", err)
	}

	data, err := fs.ReadFile("/config/config.xml")
	if err != nil {
		t.Fatalf("reading config.xml: %v", err)
	}

	var merged SyncthingXMLConfig
	if err := xml.Unmarshal(data, &merged); err != nil {
		t.Fatalf("parsing config.xml: %v", err)
	}

	if merged.Version != 35 {
		t.Errorf("Version should be preserved from existing (35), got %d", merged.Version)
	}

	if len(merged.Folders) != 1 || merged.Folders[0].ID != "existing-folder" {
		t.Errorf("existing folder not preserved: %+v", merged.Folders)
	}

	if len(merged.Devices) != 2 {
		t.Errorf("expected 2 devices preserved, got %d", len(merged.Devices))
	}

	deviceIDs := make(map[string]string)
	for _, d := range merged.Devices {
		deviceIDs[d.ID] = d.Name
	}
	if deviceIDs["DEVICE-A"] != "phone" {
		t.Error("DEVICE-A not preserved correctly")
	}
	if deviceIDs["DEVICE-B"] != "laptop" {
		t.Error("DEVICE-B not preserved correctly")
	}

	if merged.GUI.Address != "127.0.0.1:9999" {
		t.Errorf("GUI address should be updated to new port, got %s", merged.GUI.Address)
	}
	if merged.GUI.APIKey != "new-key" {
		t.Errorf("API key should be updated, got %s", merged.GUI.APIKey)
	}

	if !merged.Options.GlobalAnnounceEnabled {
		t.Error("globalAnnounceEnabled should be updated to true")
	}
	if !merged.Options.RelaysEnabled {
		t.Error("relaysEnabled should be updated to true")
	}

	foundNewPort := false
	for _, addr := range merged.Options.ListenAddresses {
		if addr == "tcp://0.0.0.0:22100" {
			foundNewPort = true
		}
	}
	if !foundNewPort {
		t.Errorf("listen port should be updated, got %v", merged.Options.ListenAddresses)
	}
}

func TestWriteConfig_PreservesDeviceAddresses(t *testing.T) {
	existingConfig := `<?xml version="1.0" encoding="UTF-8"?>
<configuration version="37">
  <device id="DEVICE-WITH-ADDRESSES" name="server" compression="metadata">
    <address>tcp://192.168.1.1:22000</address>
    <address>quic://192.168.1.1:22000</address>
    <address>dynamic</address>
  </device>
  <gui enabled="true">
    <address>127.0.0.1:8484</address>
    <apikey>key</apikey>
  </gui>
  <options>
    <listenAddress>tcp://0.0.0.0:22100</listenAddress>
  </options>
</configuration>`

	fs := testutil.NewTestFS(t, map[string]any{
		"/config/config.xml": existingConfig,
	})

	cfg := model.SyncConfig{
		Syncthing: model.SyncthingConfig{
			GUIPort:    8484,
			ListenPort: 22100,
		},
	}

	gen := NewConfigGenerator(fs, cfg, "/collection", nil, nil, nil)
	gen.SetAPIKey("key")

	if err := gen.WriteConfig("/config"); err != nil {
		t.Fatalf("WriteConfig() error = %v", err)
	}

	data, err := fs.ReadFile("/config/config.xml")
	if err != nil {
		t.Fatalf("reading config.xml: %v", err)
	}

	var merged SyncthingXMLConfig
	if err := xml.Unmarshal(data, &merged); err != nil {
		t.Fatalf("parsing config.xml: %v", err)
	}

	if len(merged.Devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(merged.Devices))
	}

	dev := merged.Devices[0]
	if len(dev.Addresses) != 3 {
		t.Errorf("expected 3 addresses preserved, got %d: %v", len(dev.Addresses), dev.Addresses)
	}
}

func TestWriteConfig_PreservesFolderDeviceShares(t *testing.T) {
	existingConfig := `<?xml version="1.0" encoding="UTF-8"?>
<configuration version="37">
  <folder id="shared-folder" label="shared" path="/data" type="sendreceive">
    <device id="DEVICE-A"></device>
    <device id="DEVICE-B"></device>
  </folder>
  <device id="DEVICE-A" name="phone" compression="metadata"></device>
  <device id="DEVICE-B" name="laptop" compression="metadata"></device>
  <gui enabled="true">
    <address>127.0.0.1:8484</address>
    <apikey>key</apikey>
  </gui>
  <options>
    <listenAddress>tcp://0.0.0.0:22100</listenAddress>
  </options>
</configuration>`

	fs := testutil.NewTestFS(t, map[string]any{
		"/config/config.xml": existingConfig,
	})

	cfg := model.SyncConfig{
		Syncthing: model.SyncthingConfig{
			GUIPort:    8484,
			ListenPort: 22100,
		},
	}

	gen := NewConfigGenerator(fs, cfg, "/collection", nil, nil, nil)
	gen.SetAPIKey("key")

	if err := gen.WriteConfig("/config"); err != nil {
		t.Fatalf("WriteConfig() error = %v", err)
	}

	data, err := fs.ReadFile("/config/config.xml")
	if err != nil {
		t.Fatalf("reading config.xml: %v", err)
	}

	var merged SyncthingXMLConfig
	if err := xml.Unmarshal(data, &merged); err != nil {
		t.Fatalf("parsing config.xml: %v", err)
	}

	if len(merged.Folders) != 1 {
		t.Fatalf("expected 1 folder, got %d", len(merged.Folders))
	}

	folder := merged.Folders[0]
	if len(folder.Devices) != 2 {
		t.Errorf("expected 2 device shares preserved, got %d", len(folder.Devices))
	}

	deviceIDs := make(map[string]bool)
	for _, d := range folder.Devices {
		deviceIDs[d.ID] = true
	}
	if !deviceIDs["DEVICE-A"] || !deviceIDs["DEVICE-B"] {
		t.Errorf("device shares not preserved correctly: %v", folder.Devices)
	}
}

func TestWriteConfig_HandlesCorruptedExistingConfig(t *testing.T) {
	fs := testutil.NewTestFS(t, map[string]any{
		"/config/config.xml": "this is not valid xml <broken",
	})

	cfg := model.SyncConfig{
		Syncthing: model.SyncthingConfig{
			GUIPort:    8484,
			ListenPort: 22100,
		},
	}

	gen := NewConfigGenerator(fs, cfg, "/collection", nil, nil, nil)
	gen.SetAPIKey("key")

	if err := gen.WriteConfig("/config"); err != nil {
		t.Fatalf("WriteConfig() should create fresh config on parse error, got: %v", err)
	}

	data, err := fs.ReadFile("/config/config.xml")
	if err != nil {
		t.Fatalf("reading config.xml: %v", err)
	}

	var config SyncthingXMLConfig
	if err := xml.Unmarshal(data, &config); err != nil {
		t.Fatalf("new config should be valid XML: %v", err)
	}

	if config.Version != 37 {
		t.Errorf("should create bootstrap config with version 37, got %d", config.Version)
	}
}

func TestWriteConfig_PreservesMultipleFolders(t *testing.T) {
	existingConfig := `<?xml version="1.0" encoding="UTF-8"?>
<configuration version="37">
  <folder id="folder-1" label="roms" path="/data/roms" type="sendreceive"></folder>
  <folder id="folder-2" label="saves" path="/data/saves" type="sendreceive">
    <versioning type="simple">
      <param key="keep" val="5"></param>
    </versioning>
  </folder>
  <folder id="folder-3" label="states" path="/data/states" type="sendreceive"></folder>
  <gui enabled="true">
    <address>127.0.0.1:8484</address>
    <apikey>key</apikey>
  </gui>
  <options>
    <listenAddress>tcp://0.0.0.0:22100</listenAddress>
  </options>
</configuration>`

	fs := testutil.NewTestFS(t, map[string]any{
		"/config/config.xml": existingConfig,
	})

	cfg := model.SyncConfig{
		Syncthing: model.SyncthingConfig{
			GUIPort:    8484,
			ListenPort: 22100,
		},
	}

	gen := NewConfigGenerator(fs, cfg, "/collection", nil, nil, nil)
	gen.SetAPIKey("key")

	if err := gen.WriteConfig("/config"); err != nil {
		t.Fatalf("WriteConfig() error = %v", err)
	}

	data, err := fs.ReadFile("/config/config.xml")
	if err != nil {
		t.Fatalf("reading config.xml: %v", err)
	}

	var merged SyncthingXMLConfig
	if err := xml.Unmarshal(data, &merged); err != nil {
		t.Fatalf("parsing config.xml: %v", err)
	}

	if len(merged.Folders) != 3 {
		t.Errorf("expected 3 folders, got %d", len(merged.Folders))
	}

	folderIDs := make(map[string]bool)
	for _, f := range merged.Folders {
		folderIDs[f.ID] = true
	}

	for _, id := range []string{"folder-1", "folder-2", "folder-3"} {
		if !folderIDs[id] {
			t.Errorf("folder %s not preserved", id)
		}
	}
}

func TestWriteConfig_HandlesEmptyExistingConfig(t *testing.T) {
	fs := testutil.NewTestFS(t, map[string]any{
		"/config/config.xml": "",
	})

	cfg := model.SyncConfig{
		Syncthing: model.SyncthingConfig{
			GUIPort:    8484,
			ListenPort: 22100,
		},
	}

	gen := NewConfigGenerator(fs, cfg, "/collection", nil, nil, nil)
	gen.SetAPIKey("key")

	if err := gen.WriteConfig("/config"); err != nil {
		t.Fatalf("WriteConfig() should create fresh config on empty file, got: %v", err)
	}

	data, err := fs.ReadFile("/config/config.xml")
	if err != nil {
		t.Fatalf("reading config.xml: %v", err)
	}

	var config SyncthingXMLConfig
	if err := xml.Unmarshal(data, &config); err != nil {
		t.Fatalf("new config should be valid XML: %v", err)
	}

	if config.Version != 37 {
		t.Errorf("should create bootstrap config, got version %d", config.Version)
	}
}

func TestWriteConfig_HandlesWhitespaceOnlyConfig(t *testing.T) {
	fs := testutil.NewTestFS(t, map[string]any{
		"/config/config.xml": "   \n\t\n   ",
	})

	cfg := model.SyncConfig{
		Syncthing: model.SyncthingConfig{
			GUIPort:    8484,
			ListenPort: 22100,
		},
	}

	gen := NewConfigGenerator(fs, cfg, "/collection", nil, nil, nil)
	gen.SetAPIKey("key")

	if err := gen.WriteConfig("/config"); err != nil {
		t.Fatalf("WriteConfig() should create fresh config on whitespace-only file, got: %v", err)
	}

	data, err := fs.ReadFile("/config/config.xml")
	if err != nil {
		t.Fatalf("reading config.xml: %v", err)
	}

	var config SyncthingXMLConfig
	if err := xml.Unmarshal(data, &config); err != nil {
		t.Fatalf("new config should be valid XML: %v", err)
	}
}

func TestWriteConfig_CreatesDirectoryIfMissing(t *testing.T) {
	fs := testutil.NewTestFS(t, nil)

	cfg := model.SyncConfig{
		Syncthing: model.SyncthingConfig{
			GUIPort:    8484,
			ListenPort: 22100,
		},
	}

	gen := NewConfigGenerator(fs, cfg, "/collection", nil, nil, nil)
	gen.SetAPIKey("key")

	if err := gen.WriteConfig("/new/nested/config/dir"); err != nil {
		t.Fatalf("WriteConfig() should create directory, got: %v", err)
	}

	data, err := fs.ReadFile("/new/nested/config/dir/config.xml")
	if err != nil {
		t.Fatalf("config.xml should exist: %v", err)
	}

	var config SyncthingXMLConfig
	if err := xml.Unmarshal(data, &config); err != nil {
		t.Fatalf("config should be valid: %v", err)
	}
}

func TestWriteConfig_PreservesVersionFromExisting(t *testing.T) {
	existingConfig := `<?xml version="1.0" encoding="UTF-8"?>
<configuration version="42">
  <gui enabled="true">
    <address>127.0.0.1:8484</address>
    <apikey>key</apikey>
  </gui>
  <options>
    <listenAddress>tcp://0.0.0.0:22100</listenAddress>
  </options>
</configuration>`

	fs := testutil.NewTestFS(t, map[string]any{
		"/config/config.xml": existingConfig,
	})

	cfg := model.SyncConfig{
		Syncthing: model.SyncthingConfig{
			GUIPort:    8484,
			ListenPort: 22100,
		},
	}

	gen := NewConfigGenerator(fs, cfg, "/collection", nil, nil, nil)
	gen.SetAPIKey("key")

	if err := gen.WriteConfig("/config"); err != nil {
		t.Fatalf("WriteConfig() error = %v", err)
	}

	data, err := fs.ReadFile("/config/config.xml")
	if err != nil {
		t.Fatalf("reading config.xml: %v", err)
	}

	var merged SyncthingXMLConfig
	if err := xml.Unmarshal(data, &merged); err != nil {
		t.Fatalf("parsing config.xml: %v", err)
	}

	if merged.Version != 42 {
		t.Errorf("version should be preserved from existing (42), got %d", merged.Version)
	}
}

func TestWriteConfig_DoesNotLoseDataOnRepeatedWrites(t *testing.T) {
	fs := testutil.NewTestFS(t, nil)

	cfg := model.SyncConfig{
		Syncthing: model.SyncthingConfig{
			GUIPort:    8484,
			ListenPort: 22100,
		},
	}

	gen := NewConfigGenerator(fs, cfg, "/collection", nil, nil, nil)
	gen.SetAPIKey("key")

	if err := gen.WriteConfig("/config"); err != nil {
		t.Fatalf("first WriteConfig() error = %v", err)
	}

	firstData, _ := fs.ReadFile("/config/config.xml")
	var first SyncthingXMLConfig
	_ = xml.Unmarshal(firstData, &first)

	manuallyModified := `<?xml version="1.0" encoding="UTF-8"?>
<configuration version="37">
  <folder id="user-added-folder" label="manual" path="/manual" type="sendreceive"></folder>
  <device id="USER-DEVICE" name="manually-added" compression="metadata"></device>
  <gui enabled="true">
    <address>127.0.0.1:8484</address>
    <apikey>key</apikey>
  </gui>
  <options>
    <listenAddress>tcp://0.0.0.0:22100</listenAddress>
  </options>
</configuration>`
	_ = fs.WriteFile("/config/config.xml", []byte(manuallyModified), 0600)

	if err := gen.WriteConfig("/config"); err != nil {
		t.Fatalf("second WriteConfig() error = %v", err)
	}

	secondData, _ := fs.ReadFile("/config/config.xml")
	var second SyncthingXMLConfig
	if err := xml.Unmarshal(secondData, &second); err != nil {
		t.Fatalf("parsing second config: %v", err)
	}

	if len(second.Folders) != 1 || second.Folders[0].ID != "user-added-folder" {
		t.Errorf("user-added folder should be preserved: %+v", second.Folders)
	}
	if len(second.Devices) != 1 || second.Devices[0].ID != "USER-DEVICE" {
		t.Errorf("user-added device should be preserved: %+v", second.Devices)
	}
}
