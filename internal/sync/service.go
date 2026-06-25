package sync

import (
	"fmt"
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
	EnableAutostart(unit string) error
	DisableAutostart(unit string) error
	IsEnabled(unit string) bool
	State(unit string) string
	Logs(unit string, lines int) string
}

type SystemctlManager struct{}

func NewDefaultServiceManager() *SystemctlManager {
	return &SystemctlManager{}
}

func runSystemctl(args ...string) error {
	out, err := exec.Command("systemctl", args...).CombinedOutput()
	if err != nil {
		if detail := strings.TrimSpace(string(out)); detail != "" {
			return fmt.Errorf("%w: %s", err, detail)
		}
		return err
	}
	return nil
}

func (m *SystemctlManager) DaemonReload() error {
	return runSystemctl("--user", "daemon-reload")
}

func (m *SystemctlManager) Start(unit string) error {
	return runSystemctl("--user", "start", unit)
}

func (m *SystemctlManager) Stop(unit string) error {
	return runSystemctl("--user", "stop", unit)
}

func (m *SystemctlManager) Restart(unit string) error {
	return runSystemctl("--user", "restart", unit)
}

func (m *SystemctlManager) Enable(unit string) error {
	return runSystemctl("--user", "enable", "--now", unit)
}

func (m *SystemctlManager) Disable(unit string) error {
	return runSystemctl("--user", "disable", "--now", unit)
}

func (m *SystemctlManager) EnableAutostart(unit string) error {
	return runSystemctl("--user", "enable", unit)
}

func (m *SystemctlManager) DisableAutostart(unit string) error {
	return runSystemctl("--user", "disable", unit)
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
