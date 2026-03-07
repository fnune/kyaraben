package sync

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeFolderSharingDrift(t *testing.T) {
	tests := []struct {
		name          string
		folders       []FolderConfig
		deviceIDs     []string
		localDeviceID string
		wantDrift     []FolderSharingDrift
	}{
		{
			name: "no drift when all devices shared",
			folders: []FolderConfig{
				{ID: "folder-1", Devices: []string{"local", "peer-a", "peer-b"}},
				{ID: "folder-2", Devices: []string{"local", "peer-a", "peer-b"}},
			},
			deviceIDs:     []string{"local", "peer-a", "peer-b"},
			localDeviceID: "local",
			wantDrift:     nil,
		},
		{
			name: "device missing from one folder",
			folders: []FolderConfig{
				{ID: "folder-1", Devices: []string{"local", "peer-a", "peer-b"}},
				{ID: "folder-2", Devices: []string{"local", "peer-a"}},
			},
			deviceIDs:     []string{"local", "peer-a", "peer-b"},
			localDeviceID: "local",
			wantDrift: []FolderSharingDrift{
				{FolderID: "folder-2", MissingDeviceIDs: []string{"peer-b"}},
			},
		},
		{
			name: "device missing from all folders",
			folders: []FolderConfig{
				{ID: "folder-1", Devices: []string{"local", "peer-a"}},
				{ID: "folder-2", Devices: []string{"local", "peer-a"}},
			},
			deviceIDs:     []string{"local", "peer-a", "peer-b"},
			localDeviceID: "local",
			wantDrift: []FolderSharingDrift{
				{FolderID: "folder-1", MissingDeviceIDs: []string{"peer-b"}},
				{FolderID: "folder-2", MissingDeviceIDs: []string{"peer-b"}},
			},
		},
		{
			name: "multiple devices missing",
			folders: []FolderConfig{
				{ID: "folder-1", Devices: []string{"local"}},
			},
			deviceIDs:     []string{"local", "peer-a", "peer-b"},
			localDeviceID: "local",
			wantDrift: []FolderSharingDrift{
				{FolderID: "folder-1", MissingDeviceIDs: []string{"peer-a", "peer-b"}},
			},
		},
		{
			name: "empty device list means no peers to check",
			folders: []FolderConfig{
				{ID: "folder-1", Devices: []string{"local"}},
			},
			deviceIDs:     []string{"local"},
			localDeviceID: "local",
			wantDrift:     nil,
		},
		{
			name:          "empty folders list",
			folders:       []FolderConfig{},
			deviceIDs:     []string{"local", "peer-a"},
			localDeviceID: "local",
			wantDrift:     nil,
		},
		{
			name: "local device not in folder is ignored",
			folders: []FolderConfig{
				{ID: "folder-1", Devices: []string{"peer-a"}},
			},
			deviceIDs:     []string{"local", "peer-a"},
			localDeviceID: "local",
			wantDrift:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeFolderSharingDrift(tt.folders, tt.deviceIDs, tt.localDeviceID)
			assert.Equal(t, tt.wantDrift, got)
		})
	}
}
