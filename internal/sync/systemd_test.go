package sync

import (
	"testing"
	"time"

	"github.com/fnune/kyaraben/internal/paths"
	"github.com/fnune/kyaraben/internal/testutil"
)

func TestSystemdUnit_Generate(t *testing.T) {
	fs := testutil.NewTestFS(t, map[string]any{})

	p := paths.NewPaths("")
	service := NewFakeServiceManager()
	unit := NewSystemdUnit(fs, p, service)

	params := UnitParams{
		BinaryPath: "/usr/bin/syncthing",
		ConfigDir:  "/home/user/.local/state/kyaraben/syncthing/config",
		DataDir:    "/home/user/.local/state/kyaraben/syncthing/data",
		GUIPort:    8484,
		APIKey:     "test-key",
	}

	content, err := unit.Generate(params)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if content == "" {
		t.Error("Generate() returned empty content")
	}

	expected := []string{
		"ExecStart=/usr/bin/syncthing serve",
		"--config=/home/user/.local/state/kyaraben/syncthing/config",
		"--gui-address=127.0.0.1:8484",
		"--gui-apikey=test-key",
	}
	for _, s := range expected {
		if !containsSubstr(content, s) {
			t.Errorf("Generate() missing %q", s)
		}
	}
}

func TestSystemdUnit_Enable_CallsServiceManager(t *testing.T) {
	fs := testutil.NewTestFS(t, map[string]any{
		"/config/systemd/user/kyaraben-syncthing.service": "[Service]\n",
	})
	t.Setenv("XDG_CONFIG_HOME", "/config")

	p := paths.NewPaths("")
	service := NewFakeServiceManager()
	unit := NewSystemdUnit(fs, p, service)

	if err := unit.Enable(); err != nil {
		t.Fatalf("Enable() error = %v", err)
	}

	if !service.DaemonReloaded {
		t.Error("Enable() should call DaemonReload()")
	}

	if !service.IsEnabled("kyaraben-syncthing.service") {
		t.Error("Enable() should enable the service")
	}

	if service.State("kyaraben-syncthing.service") != "active" {
		t.Error("Enable() should start the service")
	}
}

func TestSystemdUnit_Enable_RestartsExistingService(t *testing.T) {
	fs := testutil.NewTestFS(t, map[string]any{
		"/config/systemd/user/kyaraben-syncthing.service": "[Service]\n",
	})
	t.Setenv("XDG_CONFIG_HOME", "/config")

	p := paths.NewPaths("")
	service := NewFakeServiceManager()
	service.SetEnabled("kyaraben-syncthing.service", true)
	service.SetState("kyaraben-syncthing.service", "active")

	unit := NewSystemdUnit(fs, p, service)

	if err := unit.Enable(); err != nil {
		t.Fatalf("Enable() error = %v", err)
	}

	if service.State("kyaraben-syncthing.service") != "active" {
		t.Error("Enable() on existing service should restart and stay active")
	}
}

func TestSystemdUnit_Disable_CallsServiceManager(t *testing.T) {
	fs := testutil.NewTestFS(t, map[string]any{
		"/config/systemd/user/kyaraben-syncthing.service": "[Service]\n",
	})
	t.Setenv("XDG_CONFIG_HOME", "/config")

	p := paths.NewPaths("")
	service := NewFakeServiceManager()
	service.SetEnabled("kyaraben-syncthing.service", true)
	service.SetState("kyaraben-syncthing.service", "active")

	unit := NewSystemdUnit(fs, p, service)

	if err := unit.Disable(); err != nil {
		t.Fatalf("Disable() error = %v", err)
	}

	if service.IsEnabled("kyaraben-syncthing.service") {
		t.Error("Disable() should disable the service")
	}

	if _, err := fs.Stat("/config/systemd/user/kyaraben-syncthing.service"); err == nil {
		t.Error("Disable() should remove the unit file")
	}
}

func TestSystemdUnit_Status_Failed(t *testing.T) {
	fs := testutil.NewTestFS(t, map[string]any{})

	p := paths.NewPaths("")
	service := NewFakeServiceManager()
	service.SetState("kyaraben-syncthing.service", "failed")
	service.SetLogs("kyaraben-syncthing.service", "Error: port in use")

	unit := NewSystemdUnit(fs, p, service)

	status := unit.Status()

	if status.Active != "failed" {
		t.Errorf("Status().Active = %q, want %q", status.Active, "failed")
	}

	if !status.Failed {
		t.Error("Status().Failed should be true for failed state")
	}

	if status.Message != "Error: port in use" {
		t.Errorf("Status().Message = %q, want %q", status.Message, "Error: port in use")
	}
}

func TestSystemdUnit_Status_Active(t *testing.T) {
	fs := testutil.NewTestFS(t, map[string]any{})

	p := paths.NewPaths("")
	service := NewFakeServiceManager()
	service.SetState("kyaraben-syncthing.service", "active")

	unit := NewSystemdUnit(fs, p, service)

	status := unit.Status()

	if status.Active != "active" {
		t.Errorf("Status().Active = %q, want %q", status.Active, "active")
	}

	if status.Failed {
		t.Error("Status().Failed should be false for active state")
	}

	if status.Message != "" {
		t.Errorf("Status().Message should be empty for active state, got %q", status.Message)
	}
}

func TestSystemdUnit_IsEnabled(t *testing.T) {
	fs := testutil.NewTestFS(t, map[string]any{
		"/config/systemd/user/kyaraben-syncthing.service": "[Service]\n",
	})
	t.Setenv("XDG_CONFIG_HOME", "/config")

	p := paths.NewPaths("")
	service := NewFakeServiceManager()

	unit := NewSystemdUnit(fs, p, service)

	if unit.IsEnabled() {
		t.Error("IsEnabled() should return false when service is not enabled")
	}

	service.SetEnabled("kyaraben-syncthing.service", true)

	if !unit.IsEnabled() {
		t.Error("IsEnabled() should return true when service is enabled")
	}
}

func TestWaitForServiceStop_Inactive(t *testing.T) {
	service := NewFakeServiceManager()
	service.SetState("test.service", "inactive")

	err := waitForServiceStop(service, "test.service", 100*time.Millisecond)
	if err != nil {
		t.Errorf("waitForServiceStop() error = %v for inactive service", err)
	}
}

func TestWaitForServiceStop_NonExistent(t *testing.T) {
	service := NewFakeServiceManager()

	err := waitForServiceStop(service, "nonexistent.service", 100*time.Millisecond)
	if err != nil {
		t.Errorf("waitForServiceStop() error = %v for nonexistent service", err)
	}
}

func TestFakeServiceManager_Enable(t *testing.T) {
	service := NewFakeServiceManager()

	if err := service.Enable("test.service"); err != nil {
		t.Fatalf("Enable() error = %v", err)
	}

	if !service.IsEnabled("test.service") {
		t.Error("Enable() should set enabled to true")
	}

	if service.State("test.service") != "active" {
		t.Error("Enable() should set state to active")
	}
}

func TestFakeServiceManager_Disable(t *testing.T) {
	service := NewFakeServiceManager()
	service.SetEnabled("test.service", true)
	service.SetState("test.service", "active")

	if err := service.Disable("test.service"); err != nil {
		t.Fatalf("Disable() error = %v", err)
	}

	if service.IsEnabled("test.service") {
		t.Error("Disable() should set enabled to false")
	}

	if service.State("test.service") != "inactive" {
		t.Error("Disable() should set state to inactive")
	}
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
