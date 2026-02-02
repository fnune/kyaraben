package daemon

import (
	"encoding/json"
	"testing"

	"github.com/fnune/kyaraben/internal/model"
)

func TestSetConfigCommandParsing(t *testing.T) {
	jsonData := `{
		"type": "set_config",
		"data": {
			"userStore": "~/Emulation",
			"systems": {
				"switch": "eden@v0.1.0",
				"psx": "duckstation"
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

	if cmd.Data.Systems["switch"] != "eden@v0.1.0" {
		t.Errorf("expected switch emulator %q, got %q", "eden@v0.1.0", cmd.Data.Systems["switch"])
	}

	if cmd.Data.Systems["psx"] != "duckstation" {
		t.Errorf("expected psx emulator %q, got %q", "duckstation", cmd.Data.Systems["psx"])
	}
}

func TestSetConfigCommandParsingPreservesVersionPin(t *testing.T) {
	jsonData := `{
		"type": "set_config",
		"data": {
			"userStore": "~/Games",
			"systems": {
				"switch": "eden@v0.1.0"
			}
		}
	}`

	var cmd SetConfigCommand
	if err := json.Unmarshal([]byte(jsonData), &cmd); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	emulatorStr := cmd.Data.Systems["switch"]
	if emulatorStr != "eden@v0.1.0" {
		t.Errorf("version pin lost: expected %q, got %q", "eden@v0.1.0", emulatorStr)
	}

	// Verify the model can parse this correctly
	sysConf := model.SystemConf{Emulator: emulatorStr}
	if sysConf.EmulatorID() != model.EmulatorIDEden {
		t.Errorf("expected emulator ID %q, got %q", model.EmulatorIDEden, sysConf.EmulatorID())
	}
	if sysConf.EmulatorVersion() != "v0.1.0" {
		t.Errorf("expected version %q, got %q", "v0.1.0", sysConf.EmulatorVersion())
	}
}

func TestSyncAddDeviceCommandParsing(t *testing.T) {
	jsonData := `{
		"type": "sync_add_device",
		"data": {
			"deviceId": "ABC123",
			"name": "My Phone"
		}
	}`

	var cmd SyncAddDeviceCommand
	if err := json.Unmarshal([]byte(jsonData), &cmd); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if cmd.Type != CommandTypeSyncAddDevice {
		t.Errorf("expected type %q, got %q", CommandTypeSyncAddDevice, cmd.Type)
	}

	if cmd.Data.DeviceID != "ABC123" {
		t.Errorf("expected deviceId %q, got %q", "ABC123", cmd.Data.DeviceID)
	}

	if cmd.Data.Name != "My Phone" {
		t.Errorf("expected name %q, got %q", "My Phone", cmd.Data.Name)
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
			"systems": {"switch": "eden"}
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
