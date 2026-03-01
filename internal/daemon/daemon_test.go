package daemon

import (
	"encoding/json"
	"sync/atomic"
	"testing"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/paths"
	syncpkg "github.com/fnune/kyaraben/internal/sync"
)

func TestSetConfigCommandParsing(t *testing.T) {
	jsonData := `{
		"type": "set_config",
		"data": {
			"userStore": "~/Emulation",
			"systems": {
				"switch": ["eden"],
				"psx": ["duckstation"]
			},
			"emulators": {
				"eden": {"version": "v0.1.0"}
			}
		}
	}`

	var cmd SetConfigCommand
	if err := json.Unmarshal([]byte(jsonData), &cmd); err != nil {
		t.Fatalf("failed to unmarshal SetConfigCommand: %v", err)
	}

	if cmd.Type != CommandTypeSetConfig {
		t.Errorf("expected type %q, got %q", CommandTypeSetConfig, cmd.Type)
	}

	if cmd.Data.UserStore != "~/Emulation" {
		t.Errorf("expected userStore %q, got %q", "~/Emulation", cmd.Data.UserStore)
	}

	if len(cmd.Data.Systems) != 2 {
		t.Errorf("expected 2 systems, got %d", len(cmd.Data.Systems))
	}

	switchEmulators := cmd.Data.Systems["switch"]
	if len(switchEmulators) != 1 || switchEmulators[0] != "eden" {
		t.Errorf("expected switch emulators [eden], got %v", switchEmulators)
	}

	psxEmulators := cmd.Data.Systems["psx"]
	if len(psxEmulators) != 1 || psxEmulators[0] != "duckstation" {
		t.Errorf("expected psx emulators [duckstation], got %v", psxEmulators)
	}

	edenConf := cmd.Data.Emulators["eden"]
	if edenConf.Version != "v0.1.0" {
		t.Errorf("expected eden version v0.1.0, got %s", edenConf.Version)
	}
}

func TestSetConfigCommandParsingWithMultipleEmulators(t *testing.T) {
	jsonData := `{
		"type": "set_config",
		"data": {
			"userStore": "~/Games",
			"systems": {
				"psx": ["duckstation", "retroarch:mednafen_psx"]
			}
		}
	}`

	var cmd SetConfigCommand
	if err := json.Unmarshal([]byte(jsonData), &cmd); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	psxEmulators := cmd.Data.Systems["psx"]
	if len(psxEmulators) != 2 {
		t.Errorf("expected 2 psx emulators, got %d", len(psxEmulators))
	}

	if psxEmulators[0] != "duckstation" {
		t.Errorf("expected first emulator duckstation, got %s", psxEmulators[0])
	}

	if psxEmulators[1] != "retroarch:mednafen_psx" {
		t.Errorf("expected second emulator retroarch:mednafen_psx, got %s", psxEmulators[1])
	}
}

func TestSyncRemoveDeviceCommandParsing(t *testing.T) {
	jsonData := `{
		"type": "sync_remove_device",
		"data": {
			"deviceId": "XYZ789"
		}
	}`

	var cmd SyncRemoveDeviceCommand
	if err := json.Unmarshal([]byte(jsonData), &cmd); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if cmd.Type != CommandTypeSyncRemoveDevice {
		t.Errorf("expected type %q, got %q", CommandTypeSyncRemoveDevice, cmd.Type)
	}

	if cmd.Data.DeviceID != "XYZ789" {
		t.Errorf("expected deviceId %q, got %q", "XYZ789", cmd.Data.DeviceID)
	}
}

func TestBasicCommandDoesNotCaptureData(t *testing.T) {
	jsonData := `{
		"type": "set_config",
		"data": {
			"userStore": "~/Emulation",
			"systems": {"switch": ["eden"]}
		}
	}`

	var cmd Command
	if err := json.Unmarshal([]byte(jsonData), &cmd); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if cmd.Type != CommandTypeSetConfig {
		t.Errorf("expected type %q, got %q", CommandTypeSetConfig, cmd.Type)
	}

	// This test documents the behavior: Command struct only captures Type.
	// The data field is silently discarded. This is why we need to use
	// the specific command types (SetConfigCommand, etc.) for commands with data.
}

func TestEventSerialization(t *testing.T) {
	event := Event{
		Type: EventTypeResult,
		Data: SetConfigResponse{Success: true},
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed["type"] != string(EventTypeResult) {
		t.Errorf("expected type %q, got %q", EventTypeResult, parsed["type"])
	}

	dataMap, ok := parsed["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be a map, got %T", parsed["data"])
	}

	if dataMap["success"] != true {
		t.Errorf("expected success true, got %v", dataMap["success"])
	}
}

func TestErrorEventSerialization(t *testing.T) {
	event := Event{
		Type: EventTypeError,
		Data: ErrorResponse{Error: "something went wrong"},
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed["type"] != string(EventTypeError) {
		t.Errorf("expected type %q, got %q", EventTypeError, parsed["type"])
	}

	dataMap, ok := parsed["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be a map, got %T", parsed["data"])
	}

	if dataMap["error"] != "something went wrong" {
		t.Errorf("expected error message, got %v", dataMap["error"])
	}
}

type fakeServiceManager struct {
	state      string
	stateCalls atomic.Int32
}

func (f *fakeServiceManager) DaemonReload() error            { return nil }
func (f *fakeServiceManager) Start(unit string) error        { return nil }
func (f *fakeServiceManager) Stop(unit string) error         { return nil }
func (f *fakeServiceManager) Restart(unit string) error      { return nil }
func (f *fakeServiceManager) Enable(unit string) error       { return nil }
func (f *fakeServiceManager) Disable(unit string) error      { return nil }
func (f *fakeServiceManager) IsEnabled(unit string) bool     { return f.state == "active" }
func (f *fakeServiceManager) State(unit string) string       { f.stateCalls.Add(1); return f.state }
func (f *fakeServiceManager) Logs(unit string, n int) string { return "" }

var _ syncpkg.ServiceManager = (*fakeServiceManager)(nil)

func TestEnsureSyncthingManagedSkipsWhenActivating(t *testing.T) {
	t.Parallel()

	service := &fakeServiceManager{state: "activating"}
	d := &Daemon{
		deps: Deps{
			Paths:   paths.NewPaths(""),
			Service: service,
		},
	}

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{UserStore: "/tmp/test"},
		Sync:   model.SyncConfig{Enabled: true},
	}

	d.ensureSyncthingManaged(cfg)

	if service.stateCalls.Load() != 1 {
		t.Errorf("expected State to be called once, got %d", service.stateCalls.Load())
	}
}

func TestEnsureSyncthingManagedSkipsWhenActive(t *testing.T) {
	t.Parallel()

	service := &fakeServiceManager{state: "active"}
	d := &Daemon{
		deps: Deps{
			Paths:   paths.NewPaths(""),
			Service: service,
		},
	}

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{UserStore: "/tmp/test"},
		Sync:   model.SyncConfig{Enabled: true},
	}

	d.ensureSyncthingManaged(cfg)

	if service.stateCalls.Load() != 1 {
		t.Errorf("expected State to be called once, got %d", service.stateCalls.Load())
	}
}
