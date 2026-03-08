package sync

import (
	"context"
	"encoding/xml"
	"net"
	"testing"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/packages"
	"github.com/fnune/kyaraben/internal/paths"
	"github.com/fnune/kyaraben/internal/testutil"
)

func TestCheckPortsAvailable_AllFree(t *testing.T) {
	cfg := model.SyncthingConfig{
		GUIPort:       18380,
		ListenPort:    22080,
		DiscoveryPort: 21080,
	}

	err := checkPortsAvailable(cfg)
	if err != nil {
		t.Errorf("checkPortsAvailable() error = %v, want nil", err)
	}
}

func TestCheckPortsAvailable_TCPPortInUse(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:18381")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer func() { _ = ln.Close() }()

	cfg := model.SyncthingConfig{
		GUIPort:       18381,
		ListenPort:    22081,
		DiscoveryPort: 21081,
	}

	err = checkPortsAvailable(cfg)
	if err == nil {
		t.Error("checkPortsAvailable() should fail when GUI port in use")
	}
}

func TestCheckPortsAvailable_UDPPortInUse(t *testing.T) {
	conn, err := net.ListenPacket("udp", "127.0.0.1:21082")
	if err != nil {
		t.Fatalf("failed to listen UDP: %v", err)
	}
	defer func() { _ = conn.Close() }()

	cfg := model.SyncthingConfig{
		GUIPort:       18382,
		ListenPort:    22082,
		DiscoveryPort: 21082,
	}

	err = checkPortsAvailable(cfg)
	if err == nil {
		t.Error("checkPortsAvailable() should fail when discovery port in use")
	}
}

func TestCheckPortsAvailable_ListenPortInUse(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:22083")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer func() { _ = ln.Close() }()

	cfg := model.SyncthingConfig{
		GUIPort:       18383,
		ListenPort:    22083,
		DiscoveryPort: 21083,
	}

	err = checkPortsAvailable(cfg)
	if err == nil {
		t.Error("checkPortsAvailable() should fail when listen port in use")
	}
}

func TestCheckPorts_CanBeReplaced(t *testing.T) {
	original := CheckPorts
	defer func() { CheckPorts = original }()

	called := false
	CheckPorts = func(_ model.SyncthingConfig) error {
		called = true
		return nil
	}

	cfg := model.SyncthingConfig{GUIPort: 8385}
	_ = CheckPorts(cfg)

	if !called {
		t.Error("CheckPorts replacement was not called")
	}
}

func TestLoadPairedDevices_ExcludesLocalDevice(t *testing.T) {
	localDeviceID := "SLBSDKR-1234567-1234567-1234567-1234567-1234567-1234567-1234567"
	remoteDeviceID := "GXCWEZS-7654321-7654321-7654321-7654321-7654321-7654321-7654321"

	existingConfig := `<?xml version="1.0" encoding="UTF-8"?>
<configuration version="37">
  <device id="` + localDeviceID + `" name="steamdeck" compression="metadata"></device>
  <device id="` + remoteDeviceID + `" name="desktop" compression="metadata"></device>
</configuration>`

	fs := testutil.NewTestFS(t, map[string]any{
		"/config/config.xml": existingConfig,
	})

	fakeClient := NewFakeClient(model.SyncConfig{})
	fakeClient.SetDeviceID(localDeviceID)

	setup := NewSetup(fs, paths.DefaultPaths(), nil, "/state", nil, nil)

	ctx := context.Background()
	paired, err := setup.loadPairedDevices(ctx, fakeClient, "/config")
	if err != nil {
		t.Fatalf("loadPairedDevices() error = %v", err)
	}

	if len(paired) != 1 {
		t.Fatalf("expected 1 paired device, got %d", len(paired))
	}

	if paired[0].ID != remoteDeviceID {
		t.Errorf("expected paired device ID %s, got %s", remoteDeviceID, paired[0].ID)
	}

	if paired[0].Name != "desktop" {
		t.Errorf("expected paired device name 'desktop', got %s", paired[0].Name)
	}
}

func TestLoadPairedDevices_LocalDeviceNamedThisDevice(t *testing.T) {
	localDeviceID := "LOCAL-ID-12345-12345-12345-12345-12345-12345-12345"
	remoteDeviceID := "REMOTE-ID-67890-67890-67890-67890-67890-67890-67890"

	existingConfig := `<?xml version="1.0" encoding="UTF-8"?>
<configuration version="37">
  <device id="` + localDeviceID + `" name="this-device" compression="metadata"></device>
  <device id="` + remoteDeviceID + `" name="my-phone" compression="metadata"></device>
</configuration>`

	fs := testutil.NewTestFS(t, map[string]any{
		"/config/config.xml": existingConfig,
	})

	fakeClient := NewFakeClient(model.SyncConfig{})
	fakeClient.SetDeviceID(localDeviceID)

	setup := NewSetup(fs, paths.DefaultPaths(), nil, "/state", nil, nil)

	ctx := context.Background()
	paired, err := setup.loadPairedDevices(ctx, fakeClient, "/config")
	if err != nil {
		t.Fatalf("loadPairedDevices() error = %v", err)
	}

	if len(paired) != 1 {
		t.Fatalf("expected 1 paired device, got %d", len(paired))
	}

	if paired[0].ID != remoteDeviceID {
		t.Errorf("expected paired device ID %s, got %s", remoteDeviceID, paired[0].ID)
	}
}

func TestLoadPairedDevices_NoConfigFile(t *testing.T) {
	fs := testutil.NewTestFS(t, nil)

	fakeClient := NewFakeClient(model.SyncConfig{})
	fakeClient.SetDeviceID("LOCAL-DEVICE-ID")

	setup := NewSetup(fs, paths.DefaultPaths(), nil, "/state", nil, nil)

	ctx := context.Background()
	paired, err := setup.loadPairedDevices(ctx, fakeClient, "/config")
	if err != nil {
		t.Fatalf("loadPairedDevices() error = %v", err)
	}

	if paired != nil {
		t.Errorf("expected nil paired devices when no config file, got %v", paired)
	}
}

func TestInstall_PreservesExistingDevices(t *testing.T) {
	localDeviceID := "LOCAL-ID-12345-12345-12345-12345-12345-12345-12345"
	remoteDeviceID := "REMOTE-ID-67890-67890-67890-67890-67890-67890-67890"
	remoteName := "my-paired-device"

	existingConfig := `<?xml version="1.0" encoding="UTF-8"?>
<configuration version="37">
  <device id="` + localDeviceID + `" name="this-device" compression="metadata"></device>
  <device id="` + remoteDeviceID + `" name="` + remoteName + `" compression="metadata">
    <address>dynamic</address>
  </device>
  <gui enabled="true" tls="false" debugging="false">
    <address>127.0.0.1:8484</address>
    <apikey>existing-api-key</apikey>
  </gui>
  <options>
    <listenAddress>tcp://0.0.0.0:22100</listenAddress>
  </options>
</configuration>`

	fs := testutil.NewTestFS(t, map[string]any{
		"/state/syncthing/config/config.xml": existingConfig,
		"/state/syncthing/config/.apikey":    "existing-api-key",
	})

	fakeClient := NewFakeClient(model.SyncConfig{})
	fakeClient.SetDeviceID(localDeviceID)

	installer := packages.NewFakeInstaller(fs, "/packages")
	serviceMgr := NewFakeServiceManager()

	clientFactory := func(config model.SyncConfig) SyncClient {
		return fakeClient
	}

	setup := NewSetup(fs, paths.DefaultPaths(), installer, "/state", serviceMgr, clientFactory)

	cfg := model.SyncConfig{
		Enabled: true,
		Syncthing: model.SyncthingConfig{
			GUIPort:       8484,
			ListenPort:    22100,
			DiscoveryPort: 21127,
		},
	}

	original := CheckPorts
	CheckPorts = func(_ model.SyncthingConfig) error { return nil }
	defer func() { CheckPorts = original }()

	ctx := context.Background()
	_, err := setup.Install(ctx, cfg, "/collection", nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	configData, err := fs.ReadFile("/state/syncthing/config/config.xml")
	if err != nil {
		t.Fatalf("reading config.xml: %v", err)
	}

	var merged SyncthingXMLConfig
	if err := xml.Unmarshal(configData, &merged); err != nil {
		t.Fatalf("parsing merged config: %v", err)
	}

	if len(merged.Devices) != 2 {
		t.Fatalf("expected 2 devices in merged config, got %d", len(merged.Devices))
	}

	var foundRemote bool
	for _, dev := range merged.Devices {
		if dev.ID == remoteDeviceID && dev.Name == remoteName {
			foundRemote = true
			break
		}
	}
	if !foundRemote {
		t.Errorf("remote device %s not preserved in merged config", remoteDeviceID)
	}
}

func TestInstall_PreservesExistingFolders(t *testing.T) {
	existingConfig := `<?xml version="1.0" encoding="UTF-8"?>
<configuration version="37">
  <folder id="kyaraben-saves-psx" label="kyaraben-saves-psx" path="/collection/saves/psx" type="sendreceive"></folder>
  <folder id="kyaraben-roms-snes" label="kyaraben-roms-snes" path="/collection/roms/snes" type="sendreceive"></folder>
  <gui enabled="true" tls="false" debugging="false">
    <address>127.0.0.1:8484</address>
    <apikey>existing-api-key</apikey>
  </gui>
  <options>
    <listenAddress>tcp://0.0.0.0:22100</listenAddress>
  </options>
</configuration>`

	fs := testutil.NewTestFS(t, map[string]any{
		"/state/syncthing/config/config.xml": existingConfig,
		"/state/syncthing/config/.apikey":    "existing-api-key",
	})

	fakeClient := NewFakeClient(model.SyncConfig{})
	fakeClient.SetDeviceID("LOCAL-ID")

	installer := packages.NewFakeInstaller(fs, "/packages")
	serviceMgr := NewFakeServiceManager()

	clientFactory := func(config model.SyncConfig) SyncClient {
		return fakeClient
	}

	setup := NewSetup(fs, paths.DefaultPaths(), installer, "/state", serviceMgr, clientFactory)

	cfg := model.SyncConfig{
		Enabled: true,
		Syncthing: model.SyncthingConfig{
			GUIPort:       8484,
			ListenPort:    22100,
			DiscoveryPort: 21127,
		},
	}

	original := CheckPorts
	CheckPorts = func(_ model.SyncthingConfig) error { return nil }
	defer func() { CheckPorts = original }()

	ctx := context.Background()
	_, err := setup.Install(ctx, cfg, "/collection", nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	configData, err := fs.ReadFile("/state/syncthing/config/config.xml")
	if err != nil {
		t.Fatalf("reading config.xml: %v", err)
	}

	var merged SyncthingXMLConfig
	if err := xml.Unmarshal(configData, &merged); err != nil {
		t.Fatalf("parsing merged config: %v", err)
	}

	if len(merged.Folders) != 2 {
		t.Fatalf("expected 2 folders in merged config, got %d", len(merged.Folders))
	}

	folderIDs := make(map[string]bool)
	for _, f := range merged.Folders {
		folderIDs[f.ID] = true
	}

	if !folderIDs["kyaraben-saves-psx"] {
		t.Error("folder kyaraben-saves-psx not preserved")
	}
	if !folderIDs["kyaraben-roms-snes"] {
		t.Error("folder kyaraben-roms-snes not preserved")
	}
}

func TestInstall_UpdatesOptionsOnMerge(t *testing.T) {
	existingConfig := `<?xml version="1.0" encoding="UTF-8"?>
<configuration version="37">
  <device id="REMOTE-DEVICE" name="peer" compression="metadata"></device>
  <gui enabled="true" tls="false" debugging="false">
    <address>127.0.0.1:8484</address>
    <apikey>existing-api-key</apikey>
  </gui>
  <options>
    <listenAddress>tcp://0.0.0.0:22000</listenAddress>
    <globalAnnounceEnabled>false</globalAnnounceEnabled>
    <relaysEnabled>false</relaysEnabled>
  </options>
</configuration>`

	fs := testutil.NewTestFS(t, map[string]any{
		"/state/syncthing/config/config.xml": existingConfig,
		"/state/syncthing/config/.apikey":    "existing-api-key",
	})

	fakeClient := NewFakeClient(model.SyncConfig{})
	fakeClient.SetDeviceID("LOCAL-ID")

	installer := packages.NewFakeInstaller(fs, "/packages")
	serviceMgr := NewFakeServiceManager()

	clientFactory := func(config model.SyncConfig) SyncClient {
		return fakeClient
	}

	setup := NewSetup(fs, paths.DefaultPaths(), installer, "/state", serviceMgr, clientFactory)

	cfg := model.SyncConfig{
		Enabled: true,
		Syncthing: model.SyncthingConfig{
			GUIPort:                9999,
			ListenPort:             22100,
			DiscoveryPort:          21127,
			GlobalDiscoveryEnabled: true,
			RelayEnabled:           true,
		},
	}

	original := CheckPorts
	CheckPorts = func(_ model.SyncthingConfig) error { return nil }
	defer func() { CheckPorts = original }()

	ctx := context.Background()
	_, err := setup.Install(ctx, cfg, "/collection", nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	configData, err := fs.ReadFile("/state/syncthing/config/config.xml")
	if err != nil {
		t.Fatalf("reading config.xml: %v", err)
	}

	var merged SyncthingXMLConfig
	if err := xml.Unmarshal(configData, &merged); err != nil {
		t.Fatalf("parsing merged config: %v", err)
	}

	if len(merged.Devices) != 1 {
		t.Errorf("expected device to be preserved, got %d devices", len(merged.Devices))
	}

	if merged.GUI.Address != "127.0.0.1:9999" {
		t.Errorf("expected GUI address to be updated to 127.0.0.1:9999, got %s", merged.GUI.Address)
	}

	if !merged.Options.GlobalAnnounceEnabled {
		t.Error("expected globalAnnounceEnabled to be updated to true")
	}

	if !merged.Options.RelaysEnabled {
		t.Error("expected relaysEnabled to be updated to true")
	}

	expectedListen := "tcp://0.0.0.0:22100"
	found := false
	for _, addr := range merged.Options.ListenAddresses {
		if addr == expectedListen {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected listen address %s in options, got %v", expectedListen, merged.Options.ListenAddresses)
	}
}
