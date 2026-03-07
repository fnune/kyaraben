package guestapp

import (
	"context"

	"github.com/fnune/kyaraben/internal/syncguest"
)

type FakeSyncManager struct {
	GUIPortValue int
	StatusValue  *syncguest.Status
	StatusErr    error
	PairingCode  string
	PairingErr   error
	WaitPeerID   string
	WaitPeerErr  error
	JoinPeerID   string
	JoinErr      error
	AddPeerErr   error
	ConfigErr    error
	ShareErr     error

	ConfigureFoldersCalls     [][]syncguest.FolderMapping
	ShareFoldersWithAllCalls  int
	AddPeerCalls              []string
	CreatePairingSessionCalls int
	WaitForPeerCalls          []string
	JoinPairingSessionCalls   []string
}

func NewFakeSyncManager() *FakeSyncManager {
	return &FakeSyncManager{
		GUIPortValue: 8484,
		StatusValue:  &syncguest.Status{Running: true},
		PairingCode:  "ABC123",
		WaitPeerID:   "DEVICE-ID-1234567",
		JoinPeerID:   "DEVICE-ID-7654321",
	}
}

func (f *FakeSyncManager) GUIPort() int {
	return f.GUIPortValue
}

func (f *FakeSyncManager) ConfigureFolders(folders []syncguest.FolderMapping) error {
	f.ConfigureFoldersCalls = append(f.ConfigureFoldersCalls, folders)
	return f.ConfigErr
}

func (f *FakeSyncManager) GetStatus(ctx context.Context) (*syncguest.Status, error) {
	return f.StatusValue, f.StatusErr
}

func (f *FakeSyncManager) CreatePairingSession(ctx context.Context) (*syncguest.PairingSession, error) {
	f.CreatePairingSessionCalls++
	if f.PairingErr != nil {
		return nil, f.PairingErr
	}
	return &syncguest.PairingSession{
		Code:     f.PairingCode,
		DeviceID: "LOCAL-DEVICE-ID",
	}, nil
}

func (f *FakeSyncManager) WaitForPeer(ctx context.Context, code string) (string, error) {
	f.WaitForPeerCalls = append(f.WaitForPeerCalls, code)
	return f.WaitPeerID, f.WaitPeerErr
}

func (f *FakeSyncManager) JoinPairingSession(ctx context.Context, code string) (string, error) {
	f.JoinPairingSessionCalls = append(f.JoinPairingSessionCalls, code)
	return f.JoinPeerID, f.JoinErr
}

func (f *FakeSyncManager) AddPeer(ctx context.Context, deviceID string) error {
	f.AddPeerCalls = append(f.AddPeerCalls, deviceID)
	return f.AddPeerErr
}

func (f *FakeSyncManager) ShareFoldersWithAllDevices(ctx context.Context) error {
	f.ShareFoldersWithAllCalls++
	return f.ShareErr
}

var _ SyncManager = (*FakeSyncManager)(nil)
