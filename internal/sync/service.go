package sync

import (
	"os/exec"
	"strconv"
	"strings"
)

type ServiceManager interface {
	DaemonReload() error
	Start(unit string) error
	Stop(unit string) error
	Restart(unit string) error
	Enable(unit string) error
	Disable(unit string) error
	IsEnabled(unit string) bool
	State(unit string) string
	Logs(unit string, lines int) string
}

type SystemctlManager struct{}

func NewDefaultServiceManager() *SystemctlManager {
	return &SystemctlManager{}
}

func (m *SystemctlManager) DaemonReload() error {
	return exec.Command("systemctl", "--user", "daemon-reload").Run()
}

func (m *SystemctlManager) Start(unit string) error {
	return exec.Command("systemctl", "--user", "start", unit).Run()
}

func (m *SystemctlManager) Stop(unit string) error {
	return exec.Command("systemctl", "--user", "stop", unit).Run()
}

func (m *SystemctlManager) Restart(unit string) error {
	return exec.Command("systemctl", "--user", "restart", unit).Run()
}

func (m *SystemctlManager) Enable(unit string) error {
	return exec.Command("systemctl", "--user", "enable", "--now", unit).Run()
}

func (m *SystemctlManager) Disable(unit string) error {
	return exec.Command("systemctl", "--user", "disable", "--now", unit).Run()
}

func (m *SystemctlManager) IsEnabled(unit string) bool {
	err := exec.Command("systemctl", "--user", "is-enabled", "--quiet", unit).Run()
	return err == nil
}

func (m *SystemctlManager) State(unit string) string {
	output, _ := exec.Command("systemctl", "--user", "is-active", unit).Output()
	return strings.TrimSpace(string(output))
}

func (m *SystemctlManager) Logs(unit string, lines int) string {
	output, _ := exec.Command("journalctl", "--user", "-u", unit, "-n", strconv.Itoa(lines), "--no-pager", "-o", "cat").Output()
	return strings.TrimSpace(string(output))
}
