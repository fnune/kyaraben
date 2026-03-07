package service

import (
	"context"
	"path/filepath"

	"github.com/fnune/kyaraben/internal/guestapp"
)

type ProcessManager struct {
	*guestapp.PIDProcessController
	homePath string
}

func NewProcessManager(dataDir string) *ProcessManager {
	homePath := filepath.Join(dataDir, "syncthing")
	pidFile := filepath.Join(dataDir, "syncthing.pid")

	return &ProcessManager{
		PIDProcessController: guestapp.NewPIDProcessController(guestapp.PIDProcessConfig{
			PIDFile:   pidFile,
			MarkerStr: homePath,
			StartCmd:  nil,
		}),
		homePath: homePath,
	}
}

func (p *ProcessManager) HomePath() string {
	return p.homePath
}

func (p *ProcessManager) StopProcess() error {
	return p.Stop(context.Background())
}
