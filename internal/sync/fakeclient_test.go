package sync

import (
	"context"
	"testing"

	"github.com/fnune/kyaraben/internal/model"
)

func TestFakeClient_GetStatus_IncludesFolders(t *testing.T) {
	client := NewFakeClient(model.SyncConfig{
		Enabled: true,
		Mode:    model.SyncModePrimary,
	})
	client.SetDeviceID("TEST-DEVICE")

	client.SetFolders([]FolderStatusSummary{
		{
			ID:         "kyaraben-saves-snes",
			Path:       "/emulation/saves/snes",
			State:      "syncing",
			GlobalSize: 1000,
			LocalSize:  500,
			NeedSize:   500,
		},
		{
			ID:         "kyaraben-roms-snes",
			Path:       "/emulation/roms/snes",
			State:      "idle",
			GlobalSize: 5000,
			LocalSize:  5000,
			NeedSize:   0,
		},
	})

	status, err := client.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}

	if len(status.Folders) != 2 {
		t.Errorf("got %d folders, want 2", len(status.Folders))
	}

	foldersByID := make(map[string]FolderStatusSummary)
	for _, f := range status.Folders {
		foldersByID[f.ID] = f
	}

	savesFolder, ok := foldersByID["kyaraben-saves-snes"]
	if !ok {
		t.Fatal("kyaraben-saves-snes not found in status")
	}
	if savesFolder.State != "syncing" {
		t.Errorf("saves folder state = %s, want syncing", savesFolder.State)
	}
	if savesFolder.NeedSize != 500 {
		t.Errorf("saves folder NeedSize = %d, want 500", savesFolder.NeedSize)
	}

	romsFolder, ok := foldersByID["kyaraben-roms-snes"]
	if !ok {
		t.Fatal("kyaraben-roms-snes not found in status")
	}
	if romsFolder.State != "idle" {
		t.Errorf("roms folder state = %s, want idle", romsFolder.State)
	}
}

func TestFakeClient_GetFolderStatus_ReturnsSetStatus(t *testing.T) {
	client := NewFakeClient(model.SyncConfig{Enabled: true})

	client.SetFolderStatus("test-folder", FolderStatusSummary{
		ID:         "test-folder",
		State:      "syncing",
		GlobalSize: 2000,
		LocalSize:  1500,
		NeedSize:   500,
	})

	status, err := client.GetFolderStatus(context.Background(), "test-folder")
	if err != nil {
		t.Fatalf("GetFolderStatus() error = %v", err)
	}

	if status.State != "syncing" {
		t.Errorf("state = %s, want syncing", status.State)
	}
	if status.GlobalBytes != 2000 {
		t.Errorf("GlobalBytes = %d, want 2000", status.GlobalBytes)
	}
	if status.LocalBytes != 1500 {
		t.Errorf("LocalBytes = %d, want 1500", status.LocalBytes)
	}
	if status.NeedBytes != 500 {
		t.Errorf("NeedBytes = %d, want 500", status.NeedBytes)
	}
}

func TestFakeClient_GetFolderStatus_DefaultsToIdle(t *testing.T) {
	client := NewFakeClient(model.SyncConfig{Enabled: true})

	status, err := client.GetFolderStatus(context.Background(), "unknown-folder")
	if err != nil {
		t.Fatalf("GetFolderStatus() error = %v", err)
	}

	if status.State != "idle" {
		t.Errorf("state = %s, want idle", status.State)
	}
}

func TestFakeClient_SetFolders_ReplacesExisting(t *testing.T) {
	client := NewFakeClient(model.SyncConfig{Enabled: true})

	client.SetFolderStatus("folder1", FolderStatusSummary{ID: "folder1", State: "syncing"})
	client.SetFolders([]FolderStatusSummary{
		{ID: "folder2", State: "idle"},
	})

	status, _ := client.GetFolderStatus(context.Background(), "folder1")
	if status.State != "idle" {
		t.Error("SetFolders should have replaced folder1 (now defaults to idle)")
	}

	status, _ = client.GetFolderStatus(context.Background(), "folder2")
	if status.State != "idle" {
		t.Error("folder2 should exist after SetFolders")
	}
}

func TestFakeClient_GetDeviceCompletion_ReturnsSetCompletion(t *testing.T) {
	client := NewFakeClient(model.SyncConfig{Enabled: true})

	client.SetDeviceCompletion("device-123", CompletionResponse{
		Completion:  75.5,
		GlobalBytes: 100_000_000,
		NeedBytes:   24_500_000,
	})

	completion, err := client.GetDeviceCompletion(context.Background(), "device-123")
	if err != nil {
		t.Fatalf("GetDeviceCompletion() error = %v", err)
	}

	if completion.Completion != 75.5 {
		t.Errorf("Completion = %f, want 75.5", completion.Completion)
	}
	if completion.GlobalBytes != 100_000_000 {
		t.Errorf("GlobalBytes = %d, want 100000000", completion.GlobalBytes)
	}
	if completion.NeedBytes != 24_500_000 {
		t.Errorf("NeedBytes = %d, want 24500000", completion.NeedBytes)
	}
}

func TestFakeClient_GetDeviceCompletion_DefaultsTo100(t *testing.T) {
	client := NewFakeClient(model.SyncConfig{Enabled: true})

	completion, err := client.GetDeviceCompletion(context.Background(), "unknown-device")
	if err != nil {
		t.Fatalf("GetDeviceCompletion() error = %v", err)
	}

	if completion.Completion != 100 {
		t.Errorf("Completion = %f, want 100", completion.Completion)
	}
}
