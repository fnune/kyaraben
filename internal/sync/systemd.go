package sync

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/paths"
)

const unitTemplate = `[Unit]
Description=Kyaraben Syncthing
After=network.target

[Service]
Type=simple
ExecStart={{.BinaryPath}} serve --no-browser --config={{.ConfigDir}} --data={{.DataDir}} --gui-address=127.0.0.1:{{.GUIPort}} --gui-apikey={{.APIKey}}
Restart=on-failure
RestartSec=10
Environment=STNODEFAULTFOLDER=1
Environment=STNOUPGRADE=1

[Install]
WantedBy=default.target
`

type SystemdUnit struct {
	fs    vfs.FS
	paths *paths.Paths
}

func NewSystemdUnit(fs vfs.FS, p *paths.Paths) *SystemdUnit {
	return &SystemdUnit{fs: fs, paths: p}
}

func NewDefaultSystemdUnit() *SystemdUnit {
	return NewSystemdUnit(vfs.OSFS, paths.DefaultPaths())
}

func (s *SystemdUnit) unitName() string {
	return s.paths.DirName() + "-syncthing.service"
}

type UnitParams struct {
	BinaryPath string
	ConfigDir  string
	DataDir    string
	GUIPort    int
	APIKey     string
}

func (s *SystemdUnit) unitPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home dir: %w", err)
	}
	return filepath.Join(home, ".config", "systemd", "user", s.unitName()), nil
}

func (s *SystemdUnit) Generate(params UnitParams) (string, error) {
	tmpl, err := template.New("unit").Parse(unitTemplate)
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, params); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return buf.String(), nil
}

func (s *SystemdUnit) Write(params UnitParams) error {
	content, err := s.Generate(params)
	if err != nil {
		return err
	}

	unitPath, err := s.unitPath()
	if err != nil {
		return err
	}

	unitDir := filepath.Dir(unitPath)
	if err := vfs.MkdirAll(s.fs, unitDir, 0755); err != nil {
		return fmt.Errorf("creating systemd user dir: %w", err)
	}

	if err := s.fs.WriteFile(unitPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing unit file: %w", err)
	}

	log.Info("Wrote systemd unit to %s", unitPath)
	return nil
}

func (s *SystemdUnit) Enable() error {
	if err := exec.Command("systemctl", "--user", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("daemon-reload: %w", err)
	}

	unitName := s.unitName()

	if s.IsEnabled() {
		if err := exec.Command("systemctl", "--user", "restart", unitName).Run(); err != nil {
			return fmt.Errorf("restart service: %w", err)
		}
		log.Info("Restarted %s", unitName)
		return nil
	}

	if err := exec.Command("systemctl", "--user", "enable", "--now", unitName).Run(); err != nil {
		return fmt.Errorf("enable service: %w", err)
	}

	log.Info("Enabled and started %s", unitName)
	return nil
}

func (s *SystemdUnit) Disable() error {
	unitName := s.unitName()
	if err := exec.Command("systemctl", "--user", "disable", "--now", unitName).Run(); err != nil {
		return fmt.Errorf("disable service: %w", err)
	}

	unitPath, err := s.unitPath()
	if err != nil {
		return err
	}

	if err := s.fs.Remove(unitPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing unit file: %w", err)
	}

	log.Info("Disabled and removed %s", unitName)
	return nil
}

func (s *SystemdUnit) IsEnabled() bool {
	err := exec.Command("systemctl", "--user", "is-enabled", "--quiet", s.unitName()).Run()
	return err == nil
}

type ServiceStatus struct {
	Active  string
	Failed  bool
	Message string
}

func (s *SystemdUnit) Status() ServiceStatus {
	unitName := s.unitName()

	output, err := exec.Command("systemctl", "--user", "is-failed", unitName).Output()
	if err == nil && strings.TrimSpace(string(output)) == "failed" {
		logs, _ := exec.Command("journalctl", "--user", "-u", unitName, "-n", "5", "--no-pager", "-o", "cat").Output()
		return ServiceStatus{
			Active:  "failed",
			Failed:  true,
			Message: strings.TrimSpace(string(logs)),
		}
	}

	output, _ = exec.Command("systemctl", "--user", "is-active", unitName).Output()
	active := strings.TrimSpace(string(output))

	return ServiceStatus{
		Active: active,
		Failed: false,
	}
}
