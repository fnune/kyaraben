package syncguest

import (
	"context"
	"testing"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/sync"
)

func TestConfigureFoldersViaAPIPassesIgnoreDelete(t *testing.T) {
	enabled := true
	client := sync.NewFakeClient(model.DefaultSyncConfig())
	m := NewWithClient(DefaultConfig("/data"), client)

	err := m.ConfigureFoldersViaAPI(context.Background(), []FolderMapping{
		{ID: "kyaraben-roms-gb", Path: "/mnt/SDCARD/Roms/GB", IgnoreDelete: &enabled},
		{ID: "kyaraben-saves-gb", Path: "/mnt/SDCARD/Saves/GB"},
	})
	if err != nil {
		t.Fatalf("configuring folders: %v", err)
	}

	calls := client.AddFoldersCalls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 AddFolders call, got %d", len(calls))
	}

	for _, req := range calls[0] {
		switch req.ID {
		case "kyaraben-roms-gb":
			if req.IgnoreDelete == nil {
				t.Fatal("ROM folder lost its IgnoreDelete on the way to the client")
			}
			if !*req.IgnoreDelete {
				t.Error("expected ROM folder IgnoreDelete to be true")
			}
		case "kyaraben-saves-gb":
			if req.IgnoreDelete != nil {
				t.Errorf("saves folder should stay unmanaged, got %v", *req.IgnoreDelete)
			}
		default:
			t.Errorf("unexpected folder %s", req.ID)
		}
	}
}
