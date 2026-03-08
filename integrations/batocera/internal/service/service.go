package service

import (
	"context"
	"os/exec"
	"strings"

	"github.com/fnune/kyaraben/internal/syncthing"
)

const serviceName = "syncthing"

type Manager struct {
	client *syncthing.Client
}

func NewManager(client *syncthing.Client) *Manager {
	return &Manager{
		client: client,
	}
}

func (m *Manager) IsRunning(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "pidof", "syncthing")
	return cmd.Run() == nil
}

func (m *Manager) Start(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "batocera-services", "start", serviceName)
	return cmd.Run()
}

func (m *Manager) Stop() error {
	cmd := exec.Command("batocera-services", "stop", serviceName)
	return cmd.Run()
}

func (m *Manager) EnableAutostart() error {
	cmd := exec.Command("batocera-services", "enable", serviceName)
	return cmd.Run()
}

func (m *Manager) DisableAutostart() error {
	cmd := exec.Command("batocera-services", "disable", serviceName)
	return cmd.Run()
}

func (m *Manager) IsAutostartEnabled() bool {
	cmd := exec.Command("batocera-services", "list")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	for _, line := range strings.Split(string(output), "\n") {
		parts := strings.Split(line, ";")
		if len(parts) >= 2 && parts[0] == serviceName && parts[1] == "*" {
			return true
		}
	}
	return false
}
