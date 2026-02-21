package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

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
	fs      vfs.FS
	paths   *paths.Paths
	service ServiceManager
}

func NewSystemdUnit(fs vfs.FS, p *paths.Paths, service ServiceManager) *SystemdUnit {
	return &SystemdUnit{fs: fs, paths: p, service: service}
}

func NewDefaultSystemdUnit() *SystemdUnit {
	return NewSystemdUnit(vfs.OSFS, paths.DefaultPaths(), NewDefaultServiceManager())
}

func (s *SystemdUnit) UnitName() string {
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
	configDir, err := paths.ConfigDir()
	if err != nil {
		return "", fmt.Errorf("getting config dir: %w", err)
	}
	return filepath.Join(configDir, "systemd", "user", s.UnitName()), nil
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
	if err := s.service.DaemonReload(); err != nil {
		return fmt.Errorf("daemon-reload: %w", err)
	}

	unitName := s.UnitName()

	if s.IsEnabled() {
		if err := s.service.Restart(unitName); err != nil {
			return fmt.Errorf("restart service: %w", err)
		}
		log.Info("Restarted %s", unitName)
		return nil
	}

	if err := s.service.Enable(unitName); err != nil {
		return fmt.Errorf("enable service: %w", err)
	}

	log.Info("Enabled and started %s", unitName)
	return nil
}

func (s *SystemdUnit) Disable() error {
	unitName := s.UnitName()
	if err := s.service.Disable(unitName); err != nil {
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
	unitPath, err := s.unitPath()
	if err != nil {
		return false
	}
	if _, err := s.fs.Stat(unitPath); err != nil {
		return false
	}
	return s.service.IsEnabled(s.UnitName())
}

func FindKyarabenSyncthingServices() ([]string, error) {
	configDir, err := paths.ConfigDir()
	if err != nil {
		return nil, fmt.Errorf("getting config dir: %w", err)
	}

	systemdDir := filepath.Join(configDir, "systemd", "user")
	entries, err := os.ReadDir(systemdDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading systemd user dir: %w", err)
	}

	var services []string
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, "kyaraben") && strings.Contains(name, "syncthing") && strings.HasSuffix(name, ".service") {
			services = append(services, filepath.Join(systemdDir, name))
		}
	}
	return services, nil
}

func StopAndRemoveService(servicePath string) error {
	return StopAndRemoveServiceWithWait(NewDefaultServiceManager(), servicePath, 10*time.Second, nil)
}

func StopAndRemoveServiceWithWait(service ServiceManager, servicePath string, timeout time.Duration, ports []int) error {
	name := filepath.Base(servicePath)

	if err := service.Stop(name); err != nil {
		log.Debug("Error stopping %s: %v", name, err)
	}

	if err := waitForServiceStop(service, name, timeout); err != nil {
		log.Debug("Service %s did not stop cleanly: %v", name, err)
	}

	for _, port := range ports {
		pid, err := FindPIDByPort(port)
		if err != nil {
			continue
		}
		if pid != 0 && IsKyarabenInstance(pid) {
			log.Info("Killing orphaned syncthing process (PID %d) on port %d", pid, port)
			if err := KillProcess(pid, 5*time.Second); err != nil {
				log.Debug("Error killing process %d: %v", pid, err)
			}
		}
	}

	for _, port := range ports {
		if err := WaitForPortRelease(port, 5*time.Second); err != nil {
			log.Debug("Port %d not released: %v", port, err)
		}
	}

	_ = service.Disable(name)

	if err := os.Remove(servicePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing %s: %w", servicePath, err)
	}

	_ = service.DaemonReload()
	return nil
}

func waitForServiceStop(service ServiceManager, name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		state := service.State(name)
		if state == "inactive" || state == "failed" || state == "" {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("service %s did not stop within %v", name, timeout)
}

type ServiceStatus struct {
	Active  string
	Failed  bool
	Message string
}

func (s *SystemdUnit) Status() ServiceStatus {
	unitName := s.UnitName()
	active := s.service.State(unitName)

	status := ServiceStatus{
		Active: active,
		Failed: active == "failed",
	}

	if active != "active" && active != "activating" {
		logs := s.service.Logs(unitName, 10)
		if logs != "" {
			status.Message = logs
		} else if active == "inactive" {
			status.Message = "Service is not running. It may not have been started yet."
		} else if active == "" {
			status.Message = "Service unit not found. Sync may not be fully configured."
		}
	}

	return status
}
