package sync

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeMissingDevices(t *testing.T) {
	tests := []struct {
		name          string
		folders       []FolderConfig
		wantDeviceIDs []string
		expected      map[string][]string
	}{
		{
			name:          "empty device list returns nil",
			folders:       []FolderConfig{{ID: "folder1", Devices: []string{"dev1"}}},
			wantDeviceIDs: nil,
			expected:      nil,
		},
		{
			name:          "empty folder list returns empty map",
			folders:       nil,
			wantDeviceIDs: []string{"dev1"},
			expected:      map[string][]string{},
		},
		{
			name: "no missing devices",
			folders: []FolderConfig{
				{ID: "folder1", Devices: []string{"dev1", "dev2"}},
				{ID: "folder2", Devices: []string{"dev1", "dev2"}},
			},
			wantDeviceIDs: []string{"dev1", "dev2"},
			expected:      map[string][]string{},
		},
		{
			name: "one device missing from one folder",
			folders: []FolderConfig{
				{ID: "folder1", Devices: []string{"dev1", "dev2"}},
				{ID: "folder2", Devices: []string{"dev1"}},
			},
			wantDeviceIDs: []string{"dev1", "dev2"},
			expected: map[string][]string{
				"folder2": {"dev2"},
			},
		},
		{
			name: "multiple devices missing from multiple folders",
			folders: []FolderConfig{
				{ID: "folder1", Devices: []string{"dev1"}},
				{ID: "folder2", Devices: []string{}},
				{ID: "folder3", Devices: []string{"dev1", "dev2", "dev3"}},
			},
			wantDeviceIDs: []string{"dev1", "dev2", "dev3"},
			expected: map[string][]string{
				"folder1": {"dev2", "dev3"},
				"folder2": {"dev1", "dev2", "dev3"},
			},
		},
		{
			name: "preserves order of missing devices",
			folders: []FolderConfig{
				{ID: "folder1", Devices: []string{}},
			},
			wantDeviceIDs: []string{"aaa", "bbb", "ccc"},
			expected: map[string][]string{
				"folder1": {"aaa", "bbb", "ccc"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeMissingDevices(tt.folders, tt.wantDeviceIDs)
			assert.Equal(t, tt.expected, got)
		})
	}
}
