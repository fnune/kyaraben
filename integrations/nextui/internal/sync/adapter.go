package sync

import (
	"context"

	"github.com/fnune/kyaraben/internal/syncguest"
)

type ManagerAdapter struct {
	manager *syncguest.Manager
}

func NewManagerAdapter(m *syncguest.Manager) *ManagerAdapter {
	return &ManagerAdapter{manager: m}
}

func (a *ManagerAdapter) GUIPort() int {
	return a.manager.Client().Config().GUIPort
}

func (a *ManagerAdapter) ConfigureFolders(folders []syncguest.FolderMapping) error {
	return a.manager.ConfigureFolders(folders)
}

func (a *ManagerAdapter) GetStatus(ctx context.Context) (*syncguest.Status, error) {
	return a.manager.GetStatus(ctx)
}

func (a *ManagerAdapter) CreatePairingSession(ctx context.Context) (*syncguest.PairingSession, error) {
	return a.manager.CreatePairingSession(ctx)
}

func (a *ManagerAdapter) WaitForPeer(ctx context.Context, code string) (string, error) {
	return a.manager.WaitForPeer(ctx, code)
}

func (a *ManagerAdapter) JoinPairingSession(ctx context.Context, code string) (string, error) {
	return a.manager.JoinPairingSession(ctx, code)
}

func (a *ManagerAdapter) AddPeer(ctx context.Context, deviceID string) error {
	return a.manager.AddPeer(ctx, deviceID)
}

func (a *ManagerAdapter) ShareFoldersWithAllDevices(ctx context.Context) error {
	return a.manager.ShareFoldersWithAllDevices(ctx)
}

var _ Manager = (*ManagerAdapter)(nil)
