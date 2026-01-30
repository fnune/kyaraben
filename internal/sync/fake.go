package sync

import (
	"context"

	"github.com/fnune/kyaraben/internal/model"
)

type FakeClient struct {
	Running        bool
	DeviceID       string
	DeviceIDError  error
	Connections    map[string]ConnectionInfo
	FolderStatuses map[string]*FolderStatus
	StatusResult   *Status
	StatusError    error
}

func NewFakeClient() *FakeClient {
	return &FakeClient{
		Running:        true,
		DeviceID:       "FAKE-DEVICE-ID",
		Connections:    make(map[string]ConnectionInfo),
		FolderStatuses: make(map[string]*FolderStatus),
	}
}

func (f *FakeClient) IsRunning(ctx context.Context) bool {
	return f.Running
}

func (f *FakeClient) GetDeviceID(ctx context.Context) (string, error) {
	if f.DeviceIDError != nil {
		return "", f.DeviceIDError
	}
	return f.DeviceID, nil
}

func (f *FakeClient) GetConnections(ctx context.Context) (map[string]ConnectionInfo, error) {
	return f.Connections, nil
}

func (f *FakeClient) GetFolderStatus(ctx context.Context, folderID string) (*FolderStatus, error) {
	if status, ok := f.FolderStatuses[folderID]; ok {
		return status, nil
	}
	return &FolderStatus{State: "idle"}, nil
}

func (f *FakeClient) GetStatus(ctx context.Context) (*Status, error) {
	if f.StatusError != nil {
		return nil, f.StatusError
	}
	if f.StatusResult != nil {
		return f.StatusResult, nil
	}
	return &Status{
		Enabled:  true,
		Mode:     model.SyncModePrimary,
		DeviceID: f.DeviceID,
		GUIURL:   "http://127.0.0.1:8385",
	}, nil
}

var _ SyncClient = (*FakeClient)(nil)
