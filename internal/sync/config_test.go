package sync

import (
	"testing"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/testutil"
)

func TestConfigGenerator_GenerateFolders_Primary(t *testing.T) {
	cfg := model.SyncConfig{
		Enabled: true,
		Mode:    model.SyncModePrimary,
		Syncthing: model.SyncthingConfig{
			ListenPort:    22001,
			DiscoveryPort: 21028,
			GUIPort:       8385,
			RelayEnabled:  true,
		},
	}

	fs := testutil.NewTestFS(t, nil)

	systems := []model.SystemID{"snes", "psx"}
	gen := NewConfigGenerator(fs, cfg, "/home/user/Emulation", systems)
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
		{"kyaraben-roms-snes", FolderTypeSendOnly, "/home/user/Emulation/roms/snes"},
		{"kyaraben-roms-psx", FolderTypeSendOnly, "/home/user/Emulation/roms/psx"},
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

func TestConfigGenerator_GenerateFolders_Secondary(t *testing.T) {
	cfg := model.SyncConfig{
		Enabled: true,
		Mode:    model.SyncModeSecondary,
		Syncthing: model.SyncthingConfig{
			ListenPort:    22001,
			DiscoveryPort: 21028,
			GUIPort:       8385,
		},
	}

	fs := testutil.NewTestFS(t, nil)

	systems := []model.SystemID{"snes"}
	gen := NewConfigGenerator(fs, cfg, "/home/user/Emulation", systems)
	gen.SetDeviceID("TEST-DEVICE-ID")

	xmlCfg, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	foldersByID := make(map[string]XMLFolder)
	for _, f := range xmlCfg.Folders {
		foldersByID[f.ID] = f
	}

	romsFolder := foldersByID["kyaraben-roms-snes"]
	if romsFolder.Type != FolderTypeReceiveOnly {
		t.Errorf("secondary roms folder type = %v, want %v", romsFolder.Type, FolderTypeReceiveOnly)
	}

	savesFolder := foldersByID["kyaraben-saves-snes"]
	if savesFolder.Type != FolderTypeSendReceive {
		t.Errorf("secondary saves folder type = %v, want %v", savesFolder.Type, FolderTypeSendReceive)
	}
}

func TestConfigGenerator_Devices(t *testing.T) {
	cfg := model.SyncConfig{
		Enabled: true,
		Mode:    model.SyncModePrimary,
		Devices: []model.SyncDevice{
			{ID: "DEVICE-1", Name: "steamdeck"},
			{ID: "DEVICE-2", Name: "laptop"},
		},
		Syncthing: model.SyncthingConfig{GUIPort: 8385},
	}

	fs := testutil.NewTestFS(t, nil)

	gen := NewConfigGenerator(fs, cfg, "/tmp", nil)
	gen.SetDeviceID("MY-DEVICE-ID")

	xmlCfg, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(xmlCfg.Devices) != 3 {
		t.Errorf("got %d devices, want 3 (self + 2 remotes)", len(xmlCfg.Devices))
	}

	var foundSteamdeck, foundAutoAccept bool
	for _, dev := range xmlCfg.Devices {
		if dev.Name == "steamdeck" {
			foundSteamdeck = true
			if !dev.AutoAcceptFolders {
				t.Error("primary should auto-accept folders from paired devices")
			}
			foundAutoAccept = dev.AutoAcceptFolders
		}
	}

	if !foundSteamdeck {
		t.Error("steamdeck device not found")
	}
	if !foundAutoAccept {
		t.Error("auto-accept not enabled for paired device on primary")
	}
}

func TestConfigGenerator_Versioning(t *testing.T) {
	cfg := model.SyncConfig{
		Enabled:   true,
		Mode:      model.SyncModePrimary,
		Syncthing: model.SyncthingConfig{GUIPort: 8385},
	}

	fs := testutil.NewTestFS(t, nil)

	gen := NewConfigGenerator(fs, cfg, "/tmp", []model.SystemID{"snes"})

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
