package sync

import (
	"context"
	"net"
	"testing"

	"github.com/fnune/kyaraben/internal/model"
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
