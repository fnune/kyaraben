package sync

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/syncthing"
)

func TestStatus_OverallState(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   OverallSyncState
	}{
		{
			name:   "disabled",
			status: Status{Enabled: false},
			want:   SyncStateDisabled,
		},
		{
			name: "synced with connected devices",
			status: Status{
				Enabled: true,
				Devices: []DeviceStatus{
					{ID: "A", Connected: true},
					{ID: "B", Connected: true},
				},
				Folders: []FolderStatusSummary{
					{ID: "folder1", State: "idle", NeedSize: 0},
				},
			},
			want: SyncStateSynced,
		},
		{
			name: "syncing when folder has need",
			status: Status{
				Enabled: true,
				Devices: []DeviceStatus{{ID: "A", Connected: true}},
				Folders: []FolderStatusSummary{
					{ID: "folder1", State: "idle", NeedSize: 1024},
				},
			},
			want: SyncStateSyncing,
		},
		{
			name: "syncing when folder state is syncing",
			status: Status{
				Enabled: true,
				Devices: []DeviceStatus{{ID: "A", Connected: true}},
				Folders: []FolderStatusSummary{
					{ID: "folder1", State: "syncing", NeedSize: 0},
				},
			},
			want: SyncStateSyncing,
		},
		{
			name: "disconnected when no devices connected",
			status: Status{
				Enabled: true,
				Devices: []DeviceStatus{
					{ID: "A", Connected: false},
					{ID: "B", Connected: false},
				},
			},
			want: SyncStateDisconnected,
		},
		{
			name: "conflict takes precedence",
			status: Status{
				Enabled: true,
				Devices: []DeviceStatus{{ID: "A", Connected: true}},
				Folders: []FolderStatusSummary{
					{ID: "folder1", State: "syncing", ConflictCount: 1},
				},
			},
			want: SyncStateConflict,
		},
		{
			name: "error state",
			status: Status{
				Enabled: true,
				Devices: []DeviceStatus{{ID: "A", Connected: true}},
				Folders: []FolderStatusSummary{
					{ID: "folder1", State: "error"},
				},
			},
			want: SyncStateError,
		},
		{
			name: "disconnected with no devices configured",
			status: Status{
				Enabled: true,
				Devices: []DeviceStatus{},
				Folders: []FolderStatusSummary{},
			},
			want: SyncStateDisconnected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.OverallState()
			if got != tt.want {
				t.Errorf("OverallState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFolderLabel(t *testing.T) {
	tests := []struct {
		id   string
		want string
	}{
		{"kyaraben-saves-dreamcast", "dreamcast (saves)"},
		{"kyaraben-states-psx", "psx (states)"},
		{"kyaraben-screenshots", "screenshots"},
		{"kyaraben-frontends-esde-gamelists-dreamcast", "dreamcast (ES-DE gamelists)"},
		{"kyaraben-frontends-esde-media-snes", "snes (ES-DE media)"},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			got := FolderLabel(tt.id)
			if got != tt.want {
				t.Errorf("FolderLabel(%q) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}

func TestDiagnoseConnectivity_NoAddresses(t *testing.T) {
	recentTime := time.Now()
	result := diagnoseConnectivity("device-1", map[string][]string{}, recentTime)
	if result != "" {
		t.Errorf("diagnoseConnectivity with no addresses = %q, want empty", result)
	}

	result = diagnoseConnectivity("device-1", map[string][]string{"other-device": {"192.168.1.1:22000"}}, recentTime)
	if result != "" {
		t.Errorf("diagnoseConnectivity with wrong device = %q, want empty", result)
	}
}

func TestDiagnoseConnectivity_Reachable(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer func() { _ = listener.Close() }()

	addr := listener.Addr().String()
	discovered := map[string][]string{
		"device-1": {addr},
	}

	result := diagnoseConnectivity("device-1", discovered, time.Now())
	if result != "" {
		t.Errorf("diagnoseConnectivity with reachable port = %q, want empty", result)
	}
}

func TestDiagnoseConnectivity_Unreachable(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	addr := listener.Addr().String()
	_ = listener.Close()

	discovered := map[string][]string{
		"device-1": {addr},
	}

	result := diagnoseConnectivity("device-1", discovered, time.Now())
	if result != "port_unreachable" {
		t.Errorf("diagnoseConnectivity with closed port (recent) = %q, want port_unreachable", result)
	}

	result = diagnoseConnectivity("device-1", discovered, time.Time{})
	if result != "port_unreachable" {
		t.Errorf("diagnoseConnectivity with closed port (never seen) = %q, want port_unreachable", result)
	}

	result = diagnoseConnectivity("device-1", discovered, time.Now().Add(-10*time.Minute))
	if result != "" {
		t.Errorf("diagnoseConnectivity with closed port (old) = %q, want empty", result)
	}
}

func TestDiagnoseConnectivity_MultipleAddresses_OneReachable(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer func() { _ = listener.Close() }()

	reachableAddr := listener.Addr().String()

	closedListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start second listener: %v", err)
	}
	unreachableAddr := closedListener.Addr().String()
	_ = closedListener.Close()

	discovered := map[string][]string{
		"device-1": {unreachableAddr, reachableAddr},
	}

	result := diagnoseConnectivity("device-1", discovered, time.Now())
	if result != "" {
		t.Errorf("diagnoseConnectivity with one reachable = %q, want empty", result)
	}
}

func TestFakeClient_GetStatus_ConnectivityIssue(t *testing.T) {
	client := NewFakeClient(model.SyncConfig{Enabled: true})

	client.SetDeviceID("LOCAL-DEVICE")
	client.SetConfiguredDevice("PEER-1", "Steam Deck")
	client.SetConnectivityIssue("PEER-1", "port_unreachable")

	status, err := client.GetStatus(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if len(status.Devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(status.Devices))
	}

	if status.Devices[0].ConnectivityIssue != "port_unreachable" {
		t.Errorf("ConnectivityIssue = %q, want port_unreachable", status.Devices[0].ConnectivityIssue)
	}
}

func TestFakeClient_GetStatus_NoConnectivityIssue(t *testing.T) {
	client := NewFakeClient(model.SyncConfig{Enabled: true})

	client.SetDeviceID("LOCAL-DEVICE")
	client.SetConfiguredDevice("PEER-1", "Steam Deck")
	client.SetConnection("PEER-1", ConnectionInfo{Connected: true})

	status, err := client.GetStatus(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if len(status.Devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(status.Devices))
	}

	if status.Devices[0].ConnectivityIssue != "" {
		t.Errorf("ConnectivityIssue = %q, want empty", status.Devices[0].ConnectivityIssue)
	}
}

func TestDiagnoseLocalConnectivity_NilStatus(t *testing.T) {
	result := diagnoseLocalConnectivity(nil)
	if result != "" {
		t.Errorf("diagnoseLocalConnectivity(nil) = %q, want empty", result)
	}
}

func TestDiagnoseLocalConnectivity_NilConnectionServiceStatus(t *testing.T) {
	result := diagnoseLocalConnectivity(&syncthing.SystemStatus{})
	if result != "" {
		t.Errorf("diagnoseLocalConnectivity with nil map = %q, want empty", result)
	}
}

func TestDiagnoseLocalConnectivity_ListenError(t *testing.T) {
	status := &syncthing.SystemStatus{
		ConnectionServiceStatus: map[string]syncthing.ConnectionServiceInfo{
			"tcp://0.0.0.0:22000": {
				Error:        "listen tcp 0.0.0.0:22000: bind: address already in use",
				LANAddresses: []string{},
				WANAddresses: []string{},
			},
		},
	}
	result := diagnoseLocalConnectivity(status)
	if result != "listen_error" {
		t.Errorf("diagnoseLocalConnectivity with listen error = %q, want listen_error", result)
	}
}

func TestDiagnoseLocalConnectivity_NoLANAddress(t *testing.T) {
	status := &syncthing.SystemStatus{
		ConnectionServiceStatus: map[string]syncthing.ConnectionServiceInfo{
			"tcp://0.0.0.0:22000": {
				Error:        "",
				LANAddresses: []string{},
				WANAddresses: []string{},
			},
		},
	}
	result := diagnoseLocalConnectivity(status)
	if result != "no_lan_address" {
		t.Errorf("diagnoseLocalConnectivity with no LAN = %q, want no_lan_address", result)
	}
}

func TestDiagnoseLocalConnectivity_Healthy(t *testing.T) {
	status := &syncthing.SystemStatus{
		ConnectionServiceStatus: map[string]syncthing.ConnectionServiceInfo{
			"tcp://0.0.0.0:22000": {
				Error:        "",
				LANAddresses: []string{"tcp://192.168.1.100:22000"},
				WANAddresses: []string{"tcp://203.0.113.1:22000"},
			},
		},
	}
	result := diagnoseLocalConnectivity(status)
	if result != "" {
		t.Errorf("diagnoseLocalConnectivity healthy = %q, want empty", result)
	}
}

func TestDiagnoseLocalConnectivity_IgnoresRelay(t *testing.T) {
	status := &syncthing.SystemStatus{
		ConnectionServiceStatus: map[string]syncthing.ConnectionServiceInfo{
			"relay://relay.syncthing.net": {
				Error:        "some relay error",
				LANAddresses: []string{},
				WANAddresses: []string{},
			},
		},
	}
	result := diagnoseLocalConnectivity(status)
	if result != "" {
		t.Errorf("diagnoseLocalConnectivity with relay error = %q, want empty", result)
	}
}

func TestFakeClient_GetStatus_LocalConnectivityIssue(t *testing.T) {
	client := NewFakeClient(model.SyncConfig{Enabled: true})
	client.SetDeviceID("LOCAL-DEVICE")
	client.SetLocalConnectivityIssue("listen_error")

	status, err := client.GetStatus(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if status.LocalConnectivityIssue != "listen_error" {
		t.Errorf("LocalConnectivityIssue = %q, want listen_error", status.LocalConnectivityIssue)
	}
}
