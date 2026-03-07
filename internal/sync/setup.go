package sync

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/folders"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/packages"
	"github.com/fnune/kyaraben/internal/paths"
)

type Setup struct {
	fs        vfs.FS
	paths     *paths.Paths
	installer packages.Installer
	stateDir  string
	service   ServiceManager
}

func NewSetup(fs vfs.FS, p *paths.Paths, installer packages.Installer, stateDir string, service ServiceManager) *Setup {
	return &Setup{
		fs:        fs,
		paths:     p,
		installer: installer,
		stateDir:  stateDir,
		service:   service,
	}
}

func NewDefaultSetup(installer packages.Installer, stateDir string) *Setup {
	return NewSetup(vfs.OSFS, paths.DefaultPaths(), installer, stateDir, NewDefaultServiceManager())
}

type SetupResult struct {
	SyncthingBinary string
	ConfigDir       string
	DataDir         string
	APIKey          string
	SystemdUnitPath string
}

func (s *Setup) Install(ctx context.Context, cfg model.SyncConfig, collectionPath string, allSystems []model.SystemID, allEmulators []folders.EmulatorInfo, allFrontends []model.FrontendID, onProgress func(packages.InstallProgress)) (*SetupResult, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	if err := s.killUnmanagedSyncthing(cfg.Syncthing); err != nil {
		return nil, err
	}

	if err := CheckPorts(cfg.Syncthing); err != nil {
		return nil, err
	}

	binary, err := s.installer.InstallEmulator(ctx, "syncthing", onProgress)
	if err != nil {
		return nil, fmt.Errorf("installing syncthing: %w", err)
	}

	configDir := filepath.Join(s.stateDir, "syncthing", "config")
	dataDir := filepath.Join(s.stateDir, "syncthing", "data")

	apiKey, err := s.loadOrGenerateAPIKey(configDir)
	if err != nil {
		return nil, fmt.Errorf("generating API key: %w", err)
	}

	configGen := NewConfigGenerator(s.fs, cfg, collectionPath, allSystems, allEmulators, allFrontends)
	configGen.SetAPIKey(apiKey)

	if err := configGen.WriteConfig(configDir); err != nil {
		return nil, fmt.Errorf("writing syncthing config: %w", err)
	}

	unitGen := NewSystemdUnit(s.fs, s.paths, s.service)
	params := UnitParams{
		BinaryPath: binary.Path,
		ConfigDir:  configDir,
		DataDir:    dataDir,
		GUIPort:    cfg.Syncthing.GUIPort,
	}

	if err := unitGen.Write(params); err != nil {
		return nil, fmt.Errorf("writing systemd unit: %w", err)
	}

	if err := unitGen.Enable(); err != nil {
		return nil, fmt.Errorf("enabling syncthing service: %w", err)
	}

	unitPath, _ := unitGen.unitPath()

	return &SetupResult{
		SyncthingBinary: binary.Path,
		ConfigDir:       configDir,
		DataDir:         dataDir,
		APIKey:          apiKey,
		SystemdUnitPath: unitPath,
	}, nil
}

func (s *Setup) UpdateConfig(cfg model.SyncConfig, collectionPath string, allSystems []model.SystemID, allEmulators []folders.EmulatorInfo, allFrontends []model.FrontendID) error {
	if !cfg.Enabled {
		return nil
	}

	configDir := filepath.Join(s.stateDir, "syncthing", "config")

	apiKey, err := s.loadOrGenerateAPIKey(configDir)
	if err != nil {
		return fmt.Errorf("loading API key: %w", err)
	}

	configGen := NewConfigGenerator(s.fs, cfg, collectionPath, allSystems, allEmulators, allFrontends)
	configGen.SetAPIKey(apiKey)

	if err := configGen.WriteConfig(configDir); err != nil {
		return fmt.Errorf("writing syncthing config: %w", err)
	}

	log.Info("Updated syncthing config with %d systems, %d emulators", len(allSystems), len(allEmulators))
	return nil
}

func (s *Setup) Disable() error {
	unitGen := NewSystemdUnit(s.fs, s.paths, s.service)
	return unitGen.Disable()
}

func (s *Setup) Reset() error {
	unitGen := NewSystemdUnit(s.fs, s.paths, s.service)

	if unitGen.IsEnabled() {
		if err := unitGen.Disable(); err != nil {
			log.Error("Failed to disable syncthing service during reset: %v", err)
		}
	}

	syncthingDir := filepath.Join(s.stateDir, "syncthing")
	if err := os.RemoveAll(syncthingDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing syncthing directory: %w", err)
	}

	log.Info("Reset syncthing state at %s", syncthingDir)
	return nil
}

func (s *Setup) IsEnabled() bool {
	unitGen := NewSystemdUnit(s.fs, s.paths, s.service)
	return unitGen.IsEnabled()
}

// killUnmanagedSyncthing kills any syncthing process using our ports that is
// not managed by our systemd unit. This handles stale processes left over from
// crashes or incomplete shutdowns.
func (s *Setup) killUnmanagedSyncthing(cfg model.SyncthingConfig) error {
	unit := NewSystemdUnit(s.fs, s.paths, s.service)
	state := s.service.State(unit.UnitName())
	if state == "active" || state == "activating" {
		log.Debug("Syncthing is managed by systemd (state=%s), nothing to clean up", state)
		return nil
	}

	ports := []struct {
		port int
		name string
	}{
		{cfg.GUIPort, "GUI"},
		{cfg.ListenPort, "listen"},
	}

	for _, p := range ports {
		pid, err := FindPIDByPort(p.port)
		if err != nil {
			continue
		}
		if pid == 0 {
			continue
		}

		if IsKyarabenSyncthing(pid, s.stateDir) {
			log.Info("Killing orphaned syncthing on %s port %d (PID %d)", p.name, p.port, pid)
			if err := KillProcess(pid, 5*time.Second); err != nil {
				return fmt.Errorf("killing orphaned syncthing on %s port %d: %w", p.name, p.port, err)
			}
			if err := WaitForPortRelease(p.port, 5*time.Second); err != nil {
				return fmt.Errorf("waiting for %s port %d to be released: %w", p.name, p.port, err)
			}
		} else if IsKyarabenInstance(pid) {
			log.Info("Killing syncthing from another kyaraben instance on %s port %d (PID %d)", p.name, p.port, pid)
			if err := KillProcess(pid, 5*time.Second); err != nil {
				return fmt.Errorf("killing syncthing on %s port %d: %w", p.name, p.port, err)
			}
			if err := WaitForPortRelease(p.port, 5*time.Second); err != nil {
				return fmt.Errorf("waiting for %s port %d to be released: %w", p.name, p.port, err)
			}
		} else {
			return fmt.Errorf("syncthing %s port %d is in use by another application (PID %d)", p.name, p.port, pid)
		}
	}

	return nil
}

func (s *Setup) loadOrGenerateAPIKey(configDir string) (string, error) {
	keyPath := filepath.Join(configDir, ".apikey")

	data, err := s.fs.ReadFile(keyPath)
	if err == nil && len(data) >= 32 {
		return string(data), nil
	}

	apiKey, err := generateAPIKey()
	if err != nil {
		return "", err
	}

	if err := vfs.MkdirAll(s.fs, configDir, 0700); err != nil {
		return "", fmt.Errorf("creating config dir: %w", err)
	}

	if err := s.fs.WriteFile(keyPath, []byte(apiKey), 0600); err != nil {
		return "", fmt.Errorf("saving API key: %w", err)
	}

	return apiKey, nil
}

func generateAPIKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func checkPortsAvailable(cfg model.SyncthingConfig) error {
	ports := []struct {
		port int
		name string
		net  string
	}{
		{cfg.GUIPort, "GUI", "tcp"},
		{cfg.ListenPort, "listen", "tcp"},
		{cfg.DiscoveryPort, "discovery", "udp"},
	}

	for _, p := range ports {
		if err := checkPortAvailable(p.net, p.port); err != nil {
			return fmt.Errorf("syncthing %s port %d is already in use: %w", p.name, p.port, err)
		}
	}
	return nil
}

func checkPortAvailable(network string, port int) error {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	switch network {
	case "tcp":
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return err
		}
		_ = ln.Close()
	case "udp":
		conn, err := net.ListenPacket("udp", addr)
		if err != nil {
			return err
		}
		_ = conn.Close()
	}
	return nil
}

type PortChecker func(cfg model.SyncthingConfig) error

var CheckPorts PortChecker = checkPortsAvailable
