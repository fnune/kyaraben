package sync

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/twpayne/go-vfs/v5"
)

const unitTemplate = `[Unit]
Description=Kyaraben Syncthing
After=network.target

[Service]
Type=simple
ExecStart={{.BinaryPath}} serve --no-browser --no-default-folder --config={{.ConfigDir}} --data={{.DataDir}} --gui-address=127.0.0.1:{{.GUIPort}} --gui-apikey={{.APIKey}}
Restart=on-failure
RestartSec=10
Environment=STNODEFAULTFOLDER=1
Environment=STNOUPGRADE=1

[Install]
WantedBy=default.target
`

type SystemdUnit struct {
	fs vfs.FS
}

func NewSystemdUnit(fs vfs.FS) *SystemdUnit {
	return &SystemdUnit{fs: fs}
}

func NewDefaultSystemdUnit() *SystemdUnit {
	return NewSystemdUnit(vfs.OSFS)
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
	return filepath.Join(home, ".config", "systemd", "user", "kyaraben-syncthing.service"), nil
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

	if err := exec.Command("systemctl", "--user", "enable", "--now", "kyaraben-syncthing.service").Run(); err != nil {
		return fmt.Errorf("enable service: %w", err)
	}

	log.Info("Enabled and started kyaraben-syncthing.service")
	return nil
}

func (s *SystemdUnit) Disable() error {
	if err := exec.Command("systemctl", "--user", "disable", "--now", "kyaraben-syncthing.service").Run(); err != nil {
		return fmt.Errorf("disable service: %w", err)
	}

	unitPath, err := s.unitPath()
	if err != nil {
		return err
	}

	if err := s.fs.Remove(unitPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing unit file: %w", err)
	}

	log.Info("Disabled and removed kyaraben-syncthing.service")
	return nil
}

func (s *SystemdUnit) IsEnabled() bool {
	err := exec.Command("systemctl", "--user", "is-enabled", "--quiet", "kyaraben-syncthing.service").Run()
	return err == nil
}
