package sync

import (
	"context"

	"github.com/fnune/kyaraben/internal/syncguest"
)

type Manager interface {
	GUIPort() int
	ConfigureFolders(folders []syncguest.FolderMapping) error
	GetStatus(ctx context.Context) (*syncguest.Status, error)
	CreatePairingSession(ctx context.Context) (*syncguest.PairingSession, error)
	WaitForPeer(ctx context.Context, code string) (string, error)
	JoinPairingSession(ctx context.Context, code string) (string, error)
	AddPeer(ctx context.Context, deviceID string) error
	ShareFoldersWithAllDevices(ctx context.Context) error
}
