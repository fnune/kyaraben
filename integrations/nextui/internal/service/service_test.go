package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFriendlyDeviceName(t *testing.T) {
	tests := []struct {
		platform string
		want     string
	}{
		{"tg5040", "trimui-brick"},
		{"trimuismart", "trimui-smart-pro"},
		{"rg35xxplus", "rg35xx-plus"},
		{"rg35xx", "rg35xx"},
		{"miyoomini", "miyoo-mini"},
		{"rgb30", "rgb30"},
		{"unknown-platform", "unknown-platform"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.platform, func(t *testing.T) {
			got := friendlyDeviceName(tt.platform)
			if got != tt.want {
				t.Errorf("friendlyDeviceName(%q) = %q, want %q", tt.platform, got, tt.want)
			}
		})
	}
}

func TestLoadAPIKey_ValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configXML := `<?xml version="1.0" encoding="UTF-8"?>
<configuration>
  <gui>
    <apikey>test-api-key-12345</apikey>
  </gui>
</configuration>`

	configPath := filepath.Join(tmpDir, "syncthing", "config.xml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, []byte(configXML), 0644); err != nil {
		t.Fatal(err)
	}

	proc := NewProcessManager(tmpDir)
	mgr := &Manager{process: proc}

	apiKey, err := mgr.loadAPIKey()
	if err != nil {
		t.Fatalf("loadAPIKey() error = %v", err)
	}

	if apiKey != "test-api-key-12345" {
		t.Errorf("loadAPIKey() = %q, want %q", apiKey, "test-api-key-12345")
	}
}

func TestLoadAPIKey_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	proc := NewProcessManager(tmpDir)
	mgr := &Manager{process: proc}

	_, err := mgr.loadAPIKey()
	if err == nil {
		t.Error("loadAPIKey() should fail when config.xml missing")
	}
}

func TestLoadAPIKey_InvalidXML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "syncthing", "config.xml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, []byte("not valid xml <broken"), 0644); err != nil {
		t.Fatal(err)
	}

	proc := NewProcessManager(tmpDir)
	mgr := &Manager{process: proc}

	_, err := mgr.loadAPIKey()
	if err == nil {
		t.Error("loadAPIKey() should fail on invalid XML")
	}
}

func TestLoadAPIKey_EmptyAPIKey(t *testing.T) {
	tmpDir := t.TempDir()
	configXML := `<?xml version="1.0" encoding="UTF-8"?>
<configuration>
  <gui>
    <apikey></apikey>
  </gui>
</configuration>`

	configPath := filepath.Join(tmpDir, "syncthing", "config.xml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, []byte(configXML), 0644); err != nil {
		t.Fatal(err)
	}

	proc := NewProcessManager(tmpDir)
	mgr := &Manager{process: proc}

	apiKey, err := mgr.loadAPIKey()
	if err != nil {
		t.Fatalf("loadAPIKey() error = %v", err)
	}

	if apiKey != "" {
		t.Errorf("loadAPIKey() = %q, want empty string", apiKey)
	}
}

func TestLoadAPIKey_MissingGUIElement(t *testing.T) {
	tmpDir := t.TempDir()
	configXML := `<?xml version="1.0" encoding="UTF-8"?>
<configuration>
  <options>
    <something>value</something>
  </options>
</configuration>`

	configPath := filepath.Join(tmpDir, "syncthing", "config.xml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, []byte(configXML), 0644); err != nil {
		t.Fatal(err)
	}

	proc := NewProcessManager(tmpDir)
	mgr := &Manager{process: proc}

	apiKey, err := mgr.loadAPIKey()
	if err != nil {
		t.Fatalf("loadAPIKey() error = %v", err)
	}

	if apiKey != "" {
		t.Errorf("loadAPIKey() = %q, want empty string when gui element missing", apiKey)
	}
}

type fakeClient struct {
	running           bool
	apiKey            string
	deviceID          string
	addedDevices      []addedDevice
	disabledReporting bool
	allowedInsecure   bool
	runningCallCount  int
}

type addedDevice struct {
	id        string
	name      string
	addresses []string
}

func (f *fakeClient) IsRunning(ctx context.Context) bool {
	f.runningCallCount++
	return f.running
}

func (f *fakeClient) SetAPIKey(key string) {
	f.apiKey = key
}

func (f *fakeClient) GetDeviceID(ctx context.Context) (string, error) {
	return f.deviceID, nil
}

func (f *fakeClient) AddDeviceWithAddresses(ctx context.Context, id, name string, addresses []string) error {
	f.addedDevices = append(f.addedDevices, addedDevice{id, name, addresses})
	return nil
}

func (f *fakeClient) DisableUsageReporting(ctx context.Context) error {
	f.disabledReporting = true
	return nil
}

func (f *fakeClient) AllowInsecureAdmin(ctx context.Context) error {
	f.allowedInsecure = true
	return nil
}

func TestSetDeviceName(t *testing.T) {
	client := &fakeClient{deviceID: "DEVICE-ABC123"}
	mgr := &Manager{client: client}

	err := mgr.setDeviceName(context.Background(), "my-device")
	if err != nil {
		t.Fatalf("setDeviceName() error = %v", err)
	}

	if len(client.addedDevices) != 1 {
		t.Fatalf("expected 1 device added, got %d", len(client.addedDevices))
	}

	dev := client.addedDevices[0]
	if dev.id != "DEVICE-ABC123" {
		t.Errorf("device ID = %q, want %q", dev.id, "DEVICE-ABC123")
	}
	if dev.name != "my-device" {
		t.Errorf("device name = %q, want %q", dev.name, "my-device")
	}
	if len(dev.addresses) != 1 || dev.addresses[0] != "dynamic" {
		t.Errorf("device addresses = %v, want [dynamic]", dev.addresses)
	}
}

func TestStop_DelegatesToProcess(t *testing.T) {
	tmpDir := t.TempDir()
	proc := NewProcessManager(tmpDir)
	mgr := &Manager{process: proc}

	err := mgr.Stop()
	if err != nil {
		t.Errorf("Stop() error = %v", err)
	}
}

func TestIsRunning_NoPID(t *testing.T) {
	tmpDir := t.TempDir()
	proc := NewProcessManager(tmpDir)
	client := &fakeClient{running: true}
	mgr := &Manager{process: proc, client: client}

	if mgr.IsRunning(context.Background()) {
		t.Error("expected IsRunning() = false when no PID file")
	}

	if client.runningCallCount != 0 {
		t.Error("client.IsRunning should not be called when no PID")
	}
}

func TestEnableAutostart(t *testing.T) {
	tmpDir := t.TempDir()
	autostart := NewAutostartManager(tmpDir, "tg5040", "/pak", tmpDir)
	mgr := &Manager{autostart: autostart}

	if err := mgr.EnableAutostart(); err != nil {
		t.Errorf("EnableAutostart() error = %v", err)
	}

	if !mgr.IsAutostartEnabled() {
		t.Error("expected autostart to be enabled")
	}
}

func TestDisableAutostart(t *testing.T) {
	tmpDir := t.TempDir()
	autostart := NewAutostartManager(tmpDir, "tg5040", "/pak", tmpDir)
	mgr := &Manager{autostart: autostart}

	_ = mgr.EnableAutostart()

	if err := mgr.DisableAutostart(); err != nil {
		t.Errorf("DisableAutostart() error = %v", err)
	}

	if mgr.IsAutostartEnabled() {
		t.Error("expected autostart to be disabled")
	}
}
