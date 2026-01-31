package sync

import (
	"testing"

	"github.com/fnune/kyaraben/internal/model"
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
				Mode:    model.SyncModePrimary,
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
				Enabled:   true,
				Conflicts: []Conflict{{Path: "saves/game.sav"}},
				Devices:   []DeviceStatus{{ID: "A", Connected: true}},
				Folders: []FolderStatusSummary{
					{ID: "folder1", State: "syncing"},
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
			name: "synced with no devices configured",
			status: Status{
				Enabled: true,
				Devices: []DeviceStatus{},
				Folders: []FolderStatusSummary{},
			},
			want: SyncStateSynced,
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
